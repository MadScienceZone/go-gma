/*
\
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  _______      __          #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )(  __   )    /  \         #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  || (  )  |    \/) )        #
# | |      | || || || (___) | Assistant | (____         /   )| | /   |      | |        #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   / | (/ /) |      | |        #
# | | \_  )| |   | || (   ) |                 ) )    /   _/  |   / | |      | |        #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\|  (__) | _  __) (_       #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/(_______)(_) \____/       #
#                                                                                      #
########################################################################################
*/
//
///////////////////////////////////////////////////////////////////////////////
//                                                                           //
//                         Dice                                              //
//                                                                           //
// Random number generation for fantasy role-playing games.                  //
// Ported to Go for experimental Go map server project in 2020.              //
// Based on port to Python for new GMA framework in 2006, which in turn was  //
// derived from original dice.tcl module.                                    //
//                                                                           //
///////////////////////////////////////////////////////////////////////////////

// Package dice provides a general facility for generating random numbers
// in fantasy role-playing games.
//
// The preferred usage model is to use the higher-level abstraction provided by
// DieRoller, which rolls dice as described by strings. For example:
//
//	label, results, err := Roll("d20+16 | c")
//	label, result, err := RollOnce("15d6 + 15 fire + 1 acid")
//
// If you need to keep the die roller itself around after the dice are rolled,
// to query its status, or to produce a repeatable string of die rolls given
// a custom seed or number generator, create a new DieRoller value and reuse
// that as needed:
//
//	dr, err := NewDieRoller()
//	label, results, err := dr.DoRoll("d20+16 | c")
//	label, result, err := dr.DoRollOnce("15d6 + 15 fire + 1 acid")
//
// There is also a lower-level abstraction of dice available via the Dice
// type, created by the New function, if for some reason the DieRoller
// interface won't provide what is needed.
//
// NEW in version 5.3: The die-roll expressions now honor the usual algebraic
// order of operations instead of simply evaluating left-to-right. Parentheses
// (round brackets) can be used for grouping in the usual sense for math expressions.
package dice

import (
	"bufio"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/tcllist"
	"github.com/schwarmco/go-cartesian-product"
)

const MinimumSupportedDieRollPresetFileFormat = 1
const MaximumSupportedDieRollPresetFileFormat = 2

// Seed the random number generator with a very random seed.
func init() {
	s, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		panic(fmt.Sprintf("Unable to seed random number generator: %v", err))
	}

	rand.Seed(s.Int64())
}

//////////////////////////////////////////////////////////////////////////////////
//  ____  _
// |  _ \(_) ___ ___
// | | | | |/ __/ _ \
// | |_| | | (_|  __/
// |____/|_|\___\___|
//
//

// Dice is an abstraction of the real-world concept of a set of dice. When constructing
// a new Dice value, specify the number of dice and number of sides.
//
// This is a low-level representation of dice.
//
// See the DieRoller type for a higher-level abstraction which is the
// recommended type to use instead of this one, for most purposes.
type Dice struct {
	// Have we actually rolled the dice yet to get a result?
	Rolled bool

	// Constrained minimum and maximum values for the roll.
	// Specify 0 if there should be no minimum and/or maximum.
	MinValue int
	MaxValue int

	// The result of the last roll of this die. LastValue has a value
	// that can be used if Rolled is true.
	LastValue int

	_natural   int // interim value while confirming critical rolls
	_defthreat int // default threat for confirming critical rolls

	//
	// The parameters from which we construct a Dice
	//
	qty      int
	sides    int
	bonus    int
	diebonus int
	div      int
	factor   int
	desc     string

	// The individual components that make up the overall die-roll
	// operation to be performed.
	multiDice []dieComponent

	// The random number generator to be used with this Dice
	generator *rand.Rand

	_onlydie *dieSpec // for single-die rolls, this is the lone die
}

// ByDescription sets up a Dice value based on the
// text description given. This is the preferred way to make
// a low-level Dice value, since it is a flexible human-readable
// specification which likely came from the user anyway. (Although
// using DieRoller instead of Dice is even better.)
//
// This text description may contain any number of integer constants
// and/or die‐roll expressions separated by the basic math operators “+”,  “−”,
// “*”,  and “//”  which  respectively add, subtract, multiply, and divide the total
// value so far with the value that follows the operator.   Division performed
// with the “//” operator is integer‐only (results are immediately truncated by
// discarding any fractional part). Standard algebraic order of operations is
// followed (unary + and - are performed first, then multiplication and division, and then addition
// and subtraction, reading from left to right, and parentheses (brackets) ‘(’ and ‘)’ are
// used to enclose sub-expressions; strictly speaking, unary '+' is ignored and is actually
// dropped out of the expression by the parser).
// Generally, whitespace  is insignificant in the
// description string.
//
// Additionally, a pair of operators are available to constrain values within given
// limits. The expression x<=y will have the value of x as long as it is less than
// or equal to y; if x is greater than y, then y will be the value taken.  Likewise
// with x>=y but this means to take the value of x as long as it is greater than
// or equal to y. These operators have the highest precedence other than unary minus.
//
// On  fully  Unicode‐aware  implementations (e.g., the Go version of this package), the character “×” (U+00D7) may be used in place of “*”,
// “÷” (U+00F7) in place of “//”, “≤” (U+2264) in place of "<=", and “≥” (U+2265) in place of ">=".
// (Internally, these operators are converted to the non-ASCII Unicode runes, so ‘//‘ will appear as ‘÷’ in any output details about the die roll.)
//
// Each die‐roll expression has the general form
//
//	[>] [<n>[/<div>]] d <sides> [best|worst of <r>] [<label>]
//
// This calls for <n> dice with the given number of <sides> (which  may  be  a
// number  or the character “%” which means percentile dice or d100).  The
// optional <div> part of the expression allows a fractional number of dice:
// the  expression  “1/2d20” rolls half of a d20 (in other words, it rolls
// 1d20 and divides the result by 2, truncating the result).  The optional qualifier
// “best of <r>” will cause the dice to be rolled <r> times, keeping the
// best result. (You may also use the word worst in place of best to  take
// the lowest of the rolls.)
//
// Arbitrary  text  (<label>) may appear at the end of the expression. It is
// simply reported back in the result as a label to  describe  that  value
// (e.g.   “1d10  + 1d6 fire + 2d6 sneak”.)  If the expression begins with
// the character “>”, then the first die in the set is maximized:  in  the
// expression  “>3d6”,  the  first d6 is assumed to have the maximum value
// (6), and the remaining two dice are rolled to produce random values.
//
// The entire die roll expression may be followed by one or  both  of  the
// following  global  modifiers,  separated  from  the expression and each
// other by vertical bars (“|”): “min <a>” or “max <b>”.
//
// These force the final result to be no smaller than <a> and/or  no  larger
// than <b>, where <a> and <b> are integer values. For example:
//
//	2d6 + 1d4 | min 6
//
// which  rolls 2d6 and 1d4 to get a random value between 3 and 16, but if
// the result is less than 6, it will return 6 anyway.
//
// For example:
//
//	d, err := New(ByDescription("3d6 + 12"))
func ByDescription(desc string) func(*Dice) error {
	return func(o *Dice) error {
		o.desc = desc
		return nil
	}
}

// ByDieType sets the Dice up by
// discrete values which determine the number of dice to roll,
// how many sides they each have, and a bonus to add to their
// sum.
//
// For example, to create a Dice value for "3d6+10", you could do:
//
//	d, err := New(ByDieType(3, 6, 10))
func ByDieType(qty, sides, bonus int) func(*Dice) error {
	return func(o *Dice) error {
		o.qty = qty
		o.sides = sides
		o.bonus = bonus
		return nil
	}
}

// WithDieBonus adds a per-die bonus of n which will be
// added to every single die rolled.
func WithDieBonus(n int) func(*Dice) error {
	return func(o *Dice) error {
		o.diebonus = n
		return nil
	}
}

// WithDiv causes the total die-roll to be divided by n
// (as an integer division, truncated toward zero).
//
// Deprecated: use WithDescription("... // n")
// or WithDescription("... ÷ n") instead.
func WithDiv(n int) func(*Dice) error {
	return func(o *Dice) error {
		o.div = n
		return nil
	}
}

// WithFactor causes the total die-roll to be multiplied by
// n.
//
// Deprecated: use WithDescription("... * n")
// or WithDescription("... × n") instead.
func WithFactor(n int) func(*Dice) error {
	return func(o *Dice) error {
		o.factor = n
		return nil
	}
}

// WithSeed sets up the Dice value to use a random
// number generator with the given seed value.
// (Per rand, this generator will not be safe for concurrent
// use by multiple goroutines.)
func WithSeed(s int64) func(*Dice) error {
	return func(o *Dice) error {
		o.generator = rand.New(rand.NewSource(s))
		return nil
	}
}

// WithGenerator sets up the Dice value to use a random
// number generator created by the caller and passed in to this
// option.
func WithGenerator(source rand.Source) func(*Dice) error {
	return func(o *Dice) error {
		o.generator = rand.New(source)
		return nil
	}
}

func withSharedGenerator(generator *rand.Rand) func(*Dice) error {
	return func(o *Dice) error {
		o.generator = generator
		return nil
	}
}

// StructuredDescription values are used to
// report die-roll results as a structured description list.
type StructuredDescription struct {
	// A text label describing what the value means in the context
	// of the die-roll result. Typical type labels are documented
	// in dice(3).
	Type string

	// The value (as a string, since the intent here is just to report
	// these values to a human with various kinds of formatting) for this
	// part of the description.
	Value string
}

// A StructuredDescriptionSet is simply a collection of StructuredDescriptions.
type StructuredDescriptionSet []StructuredDescription

// StructuredResult provides a full, detailed report of the die-roll
// operation.
// See the documentation in dice(3) for full details.
//
// For example, making a die roll such as Roll("1d20+3|min 5|c")
// might result in the following StructuredResult slice, which contains
// a fully-described breakdown of how those results were arrived at:
//
//	 []StructuredResult{{
//	   Result: 23,
//	   Details: StructuredDescriptionSet{
//	     {Type: "result",    Value: "23"},
//	     {Type: "success",   Value: "HIT"},
//	     {Type: "separator", Value: "="},
//	     {Type: "diespec",   Value: "1d20"},
//	     {Type: "roll",      Value: "20"},
//	     {Type: "operator",  Value: "+"},
//	     {Type: "constant",  Value: "3"},
//	     {Type: "moddelim",  Value: "|"},
//	     {Type: "min",       Value: "5"},
//	     {Type: "moddelim",  Value: "|"},
//	     {Type: "critspec",  Value: "c"},
//	   }
//	  }, {
//	   Result: 13,
//	   Details: StructuredDescriptionSet{
//	     {Type: "critlabel", Value: "Confirm:"},
//	     {Type: "result",    Value: "13"},
//	     {Type: "separator", Value: "="},
//	     {Type: "diespec",   Value: "1d20"},
//	     {Type: "roll",      Value: "10"},
//	     {Type: "operator",  Value: "+"},
//	     {Type: "constant",  Value: "3"},
//	     {Type: "moddelim",  Value: "|"},
//	     {Type: "min",       Value: "5"},
//	  }
//	}}
type StructuredResult struct {
	// True if there is no actual result generated (and the Result field should be ignored)
	ResultSuppressed bool `json:",omitempty"`

	// True if the die-roll request was invalid.
	InvalidRequest bool `json:",omitempty"`

	// Total final result of the expression.
	Result int

	// Breakdown of how the result was obtained.
	Details StructuredDescriptionSet
}

// An evalStack is used when parsing the die-roll expression's algebraic
// operators using the standard order of operation and brackets.
type evalStack struct {
	stack   []float64
	opStack []rune
}

func (s *evalStack) isOpEmpty() bool {
	return len(s.opStack) == 0
}

func (s *evalStack) push(v float64) {
	s.stack = append(s.stack, v)
}

func (s *evalStack) pop() (float64, error) {
	stackLen := len(s.stack)
	if stackLen == 0 {
		return 0, fmt.Errorf("stack underflow")
	}
	poppedValue := s.stack[stackLen-1]
	s.stack = s.stack[:stackLen-1]
	return poppedValue, nil
}

func (s *evalStack) pushOp(v rune) {
	s.opStack = append(s.opStack, v)
}

func (s *evalStack) popOp() (rune, error) {
	stackLen := len(s.opStack)
	if stackLen == 0 {
		return 0, fmt.Errorf("operator stack underflow")
	}
	poppedValue := s.opStack[stackLen-1]
	s.opStack = s.opStack[:stackLen-1]
	return poppedValue, nil
}

func (s *evalStack) discardOp() {
	if !s.isOpEmpty() {
		_, _ = s.popOp()
	}
}

func (s *evalStack) applyOp() error {
	var x, y float64

	op, err := s.popOp()
	if err != nil {
		return err
	}
	if op == '(' || op == ')' {
		return nil
	}

	if op == '‾' { // unary - (negation)
		if x, err = s.pop(); err != nil {
			return err
		}
		s.push(-x)
		return nil
	}

	if y, err = s.pop(); err != nil {
		return err
	}
	if x, err = s.pop(); err != nil {
		return err
	}

	switch op {
	case '+':
		s.push(math.Floor(x + y))
	case '-':
		s.push(math.Floor(x - y))
	case '*', '×':
		s.push(math.Floor(x * y))
	case '÷':
		if y == 0 {
			return fmt.Errorf("division by zero is not defined")
		}
		s.push(math.Floor(x / y))
	case '≤':
		if x > y {
			s.push(y)
		} else {
			s.push(x)
		}
	case '≥':
		if x < y {
			s.push(y)
		} else {
			s.push(x)
		}
	default:
		return fmt.Errorf("Unknown operator \"%v\"", op)
	}
	return nil
}

func (s *evalStack) nextOp() rune {
	stackLen := len(s.opStack)
	if stackLen == 0 {
		return 0
	}
	return s.opStack[stackLen-1]
}

// complete the evaluation of the expression by applying all remaining operators
func (s *evalStack) evaluate() (int, error) {
	for !s.isOpEmpty() {
		if s.nextOp() == '(' {
			return 0, fmt.Errorf("'(' without matching ')' in die-roll expression")
		}
		if err := s.applyOp(); err != nil {
			return 0, err
		}
	}
	value, err := s.pop()
	if err != nil {
		return 0, err
	}
	if len(s.stack) > 0 {
		return 0, fmt.Errorf("expression stack not empty at end of evaluation")
	}
	return int(value), nil
}

func (s *evalStack) reset() {
	s.stack = nil
	s.opStack = nil
}

// A dieComponent is something that can be assembled with other dieComponents
// to form a full die-roll spec expression.
//
// E.g., if a die roll specification consists of the components
//
//	    dieOperator +
//	    diespec 1d20
//		   dieOperator -
//	    constant 2
//		   dieOperator +
//	    diespec 2d6
//
// then the evaluation of the overall die-roll spec ("+1d20-2+2d6") is
// performed by evaluating the list of operators and values left to right
// (with algebraic order of operations) resulting in the effective value
// (1d20 - 2) + 2d6, with random values substituted for 1d20 and 2d6.
type dieComponent interface {
	// Feed this value into the expression evaluation in progress.
	compute(s *evalStack) error
	computeMaxValue(s *evalStack) error

	// Return the most recently calculated value. (This can be used to
	// get the random value rolled for diespecs.) This legacy method
	// is not currently used anymore except in a unit test. For non-
	// integer values it returns 0 (including floating-point constants).
	lastValue() int

	// Describe the die-roll component as a string.
	description() string

	// Describe the die-roll component as explicit elements.
	structuredDescribeRoll(bool) []StructuredDescription

	// Returns the natural roll value of the component. This must be a single
	// die; otherwise -1 is returned. If the component isn't a die to be rolled,
	// 0 is returned. The second return value is the number of sides on the die.
	// Thus, a natural 3 on a d20 would be returned as (3, 20).
	naturalRoll() (int, int)
}

// dieLabel represents a bare label appearing outside the normal expression context.
type dieLabel string

func (l dieLabel) compute(s *evalStack) error {
	return nil
}

func (l dieLabel) computeMaxValue(s *evalStack) error {
	return nil
}

func (l dieLabel) lastValue() int {
	return 0
}

func (l dieLabel) description() string {
	return string(l)
}

func (l dieLabel) structuredDescribeRoll(resultSuppressed bool) []StructuredDescription {
	return []StructuredDescription{
		StructuredDescription{Type: "label", Value: string(l)},
	}
}

func (l dieLabel) naturalRoll() (int, int) {
	return 0, 0
}

// dieOperator represents an algebraic operator in our expression.
type dieOperator rune

func precedence(op dieOperator) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '×', '÷':
		return 2
	case '≤', '≥':
		return 3
	case '‾':
		return 4
	}
	return 0
}

func (o dieOperator) compute(s *evalStack) error {
	for !s.isOpEmpty() && s.nextOp() != '(' && precedence(dieOperator(s.nextOp())) >= precedence(o) {
		if err := s.applyOp(); err != nil {
			return err
		}
	}
	s.pushOp(rune(o))
	return nil
}

func (o dieOperator) computeMaxValue(s *evalStack) error {
	return o.compute(s)
}

func (o dieOperator) lastValue() int {
	return 0
}

func (o dieOperator) description() string {
	if o == '‾' {
		return "-"
	}
	return string(o)
}

func (o dieOperator) structuredDescribeRoll(suppressed bool) []StructuredDescription {
	return []StructuredDescription{
		StructuredDescription{Type: "operator", Value: o.description()},
	}
}

func (o dieOperator) naturalRoll() (int, int) {
	return 0, 0
}

// dieBeginGroup and dieEndGroup start and end grouped sub-expressions.
type dieBeginGroup byte
type dieEndGroup byte

func (b dieBeginGroup) compute(s *evalStack) error {
	s.pushOp('(')
	return nil
}

func (b dieBeginGroup) computeMaxValue(s *evalStack) error {
	return b.compute(s)
}

func (b dieBeginGroup) lastValue() int {
	return 0
}

func (b dieBeginGroup) description() string {
	return "("
}

func (b dieBeginGroup) structuredDescribeRoll(bool) []StructuredDescription {
	return []StructuredDescription{
		StructuredDescription{Type: "begingroup", Value: "("},
	}
}

func (b dieBeginGroup) naturalRoll() (int, int) {
	return 0, 0
}

func (b dieEndGroup) compute(s *evalStack) error {
	for s.nextOp() != '(' {
		if err := s.applyOp(); err != nil {
			return err
		}
	}
	if s.nextOp() != '(' {
		return fmt.Errorf("')' with no matching '('")
	}
	s.discardOp()
	return nil
}

func (b dieEndGroup) computeMaxValue(s *evalStack) error {
	return b.compute(s)
}

func (b dieEndGroup) lastValue() int {
	return 0
}

func (b dieEndGroup) description() string {
	return ")"
}

func (b dieEndGroup) structuredDescribeRoll(bool) []StructuredDescription {
	return []StructuredDescription{
		StructuredDescription{Type: "endgroup", Value: ")"},
	}
}

func (b dieEndGroup) naturalRoll() (int, int) {
	return 0, 0
}

// dieConstant is a kind of dieComponent that provides a constant
// value that is part of an expression.
type dieConstant struct {
	// The constant value itself.
	Value float64

	// An optional label to indicate what the constant actually represents.
	Label string
}

func (d *dieConstant) compute(s *evalStack) error {
	s.push(d.Value)
	return nil
}

func (d *dieConstant) computeMaxValue(s *evalStack) error {
	return d.compute(s)
}

func (d *dieConstant) lastValue() int {
	return int(d.Value)
}

func (d *dieConstant) naturalRoll() (int, int) {
	return 0, 0
}

func (d *dieConstant) description() string {
	return strconv.FormatFloat(d.Value, 'g', -1, 64) + d.Label
}

func (d *dieConstant) structuredDescribeRoll(resultSuppressed bool) []StructuredDescription {
	var desc []StructuredDescription
	desc = append(desc, StructuredDescription{Type: "constant", Value: strconv.FormatFloat(d.Value, 'g', -1, 64)})
	if d.Label != "" {
		desc = append(desc, StructuredDescription{Type: "label", Value: d.Label})
	}
	return desc
}

// dieSpec is a part of a die-roll expression that specifies a single
// roll (NdS+B, etc) in a chain of other components.
type dieSpec struct {
	// Various boolean flags
	BestReroll   bool // if making multiple rolls, are we taking the best (vs worst)?
	WasMaximized bool // A boolean that indicates if the result was generated at maximum value.
	InitialMax   bool // Should we maximize the first die (e.g. for ">5d6", the first d6 will be "6").

	// The value of the die after it was rolled.
	Value int

	// The die itself is represented as (numerator)/(denominator)D(sides).
	Numerator   int
	Denominator int
	Sides       int

	// If making multiple rolls, we keep track of them here.
	Rerolls int

	// A bonus applied to the die every time.
	//
	// Deprecated: use die-roll expression strings instead.
	DieBonus int

	// Label string for this component, if any
	Label string

	// A record of the actual die rolls performed, per re-roll attempt.
	History [][]int

	_natural  int
	generator *rand.Rand
}

// Assuming the die (and it must be a single die) for this component
// has already been rolled, return the natural value of that die
// and the number of sides.
func (d *dieSpec) naturalRoll() (int, int) {
	return d._natural, d.Sides
}

func sumOf(a []int) (t int) {
	for _, v := range a {
		t += v
	}
	return
}

func reduceSums(a [][]int) (sums []int) {
	for _, s := range a {
		sums = append(sums, sumOf(s))
	}
	return
}

func maxOf(a []int) (max int, pos int) {
	for i, v := range a {
		if i == 0 || max < v {
			max = v
			pos = i
		}
	}
	return
}

func minOf(a []int) (min int, pos int) {
	for i, v := range a {
		if i == 0 || min > v {
			min = v
			pos = i
		}
	}
	return
}

func intToStrings(a []int) (as []string) {
	for _, v := range a {
		as = append(as, strconv.Itoa(v))
	}
	return
}

func (d *dieSpec) compute(s *evalStack) error {
	d.History = nil
	d.WasMaximized = false
	if d.Sides <= 0 {
		return fmt.Errorf("dice cannot have a nonpositive number of sides")
	}
	for i := 0; i <= d.Rerolls; i++ {
		this := []int{}
		for j := 0; j < d.Numerator; j++ {
			v := 0
			if d.InitialMax && j == 0 {
				v = d.Sides + d.DieBonus
			} else {
				if d.generator == nil {
					v = int(rand.Int31n(int32(d.Sides))) + 1 + d.DieBonus
				} else {
					v = int(d.generator.Int31n(int32(d.Sides))) + 1 + d.DieBonus
				}
			}
			if d.Denominator > 0 {
				v /= d.Denominator
				if v < 1 {
					v = 1
				}
			}
			this = append(this, v)
		}
		d.History = append(d.History, this)
	}

	var pos int
	if d.Rerolls > 0 {
		// select the best or worst roll
		if d.BestReroll {
			d.Value, pos = maxOf(reduceSums(d.History))
		} else {
			d.Value, pos = minOf(reduceSums(d.History))
		}
		if d.Numerator == 1 {
			d._natural = d.History[pos][0] - d.DieBonus
		} else {
			d._natural = -1
		}
	} else {
		// no rerolls, so we just have one set of results
		d.Value = reduceSums(d.History)[0]
		if d.Numerator == 1 {
			d._natural = d.History[0][0] - d.DieBonus
		} else {
			d._natural = -1
		}
	}

	s.push(float64(d.Value))
	return nil
}

func (d *dieSpec) computeMaxValue(s *evalStack) error {
	d.WasMaximized = true
	d.History = nil
	this := []int{}
	for j := 0; j < d.Numerator; j++ {
		v := d.Sides + d.DieBonus
		if d.Denominator > 0 {
			v /= d.Denominator
			if v < 1 {
				v = 1
			}
		}
		this = append(this, v)
	}
	d.History = append(d.History, this)
	d.Value = reduceSums(d.History)[0]
	s.push(float64(d.Value))
	return nil
}

func (d *dieSpec) lastValue() int {
	return d.Value
}

func (d *dieSpec) description() string {
	desc := ""
	if d.InitialMax {
		desc += ">"
	}
	if d.Denominator > 0 {
		desc += fmt.Sprintf("%d/%dd%d", d.Numerator, d.Denominator, d.Sides)
	} else {
		desc += fmt.Sprintf("%dd%d", d.Numerator, d.Sides)
	}
	if d.DieBonus > 0 {
		desc += fmt.Sprintf(" (%+d per die)", d.DieBonus)
	}
	if d.Rerolls > 0 {
		if d.BestReroll {
			desc += fmt.Sprintf(" best of %d", d.Rerolls+1)
		} else {
			desc += fmt.Sprintf(" worst of %d", d.Rerolls+1)
		}
	}
	if d.Label != "" {
		desc += " " + d.Label
	}
	return desc
}

// Returns true if the value rolled for this component was a 1.
func (d *dieSpec) isMinRoll() bool {
	return d.Value == 1
}

// Returns true if the value rolled for this component is the same as
// the number of sides on the die.
func (d *dieSpec) isMaxRoll() bool {
	return d.Value == d.Sides
}

// Given a dieSpec value, the StructuredDescribeRoll method
// returns a detailed description of that component of the roll,
// as a number of StructuredDescription values.
func (d *dieSpec) structuredDescribeRoll(resultSuppressed bool) []StructuredDescription {
	var desc []StructuredDescription
	var rollType string
	if d.WasMaximized {
		rollType = "maxroll"
	} else {
		rollType = "roll"
	}

	if d.InitialMax {
		desc = append(desc, StructuredDescription{Type: "maximized", Value: ">"})
	}
	if d.Denominator > 0 {
		desc = append(desc, StructuredDescription{Type: "diespec", Value: fmt.Sprintf("%d/%dd%d", d.Numerator, d.Denominator, d.Sides)})
	} else {
		desc = append(desc, StructuredDescription{Type: "diespec", Value: fmt.Sprintf("%dd%d", d.Numerator, d.Sides)})
	}
	if d.DieBonus > 0 {
		desc = append(desc, StructuredDescription{Type: "diebonus", Value: fmt.Sprintf("%+d", d.DieBonus)})
	}

	if !resultSuppressed && len(d.History[0]) > 1 {
		desc = append(desc, StructuredDescription{Type: "subtotal", Value: strconv.Itoa(d.Value)})
	}

	if d.Rerolls > 0 {
		if resultSuppressed {
			if d.BestReroll {
				desc = append(desc, StructuredDescription{Type: "best", Value: strconv.Itoa(d.Rerolls + 1)})
			} else {
				desc = append(desc, StructuredDescription{Type: "worst", Value: strconv.Itoa(d.Rerolls + 1)})
			}
		} else {
			if d.BestReroll {
				desc = append(desc, StructuredDescription{Type: "best", Value: strconv.Itoa(d.Rerolls + 1)})
				_, choice := maxOf(reduceSums(d.History))
				for i, roll := range d.History {
					if i == choice {
						desc = append(desc, StructuredDescription{Type: rollType, Value: strings.Join(intToStrings(roll), ",")})
					} else {
						desc = append(desc, StructuredDescription{Type: "discarded", Value: strings.Join(intToStrings(roll), ",")})
					}
				}
			} else {
				desc = append(desc, StructuredDescription{Type: "worst", Value: strconv.Itoa(d.Rerolls + 1)})
				_, choice := minOf(reduceSums(d.History))
				for i, roll := range d.History {
					if i == choice {
						desc = append(desc, StructuredDescription{Type: rollType, Value: strings.Join(intToStrings(roll), ",")})
					} else {
						desc = append(desc, StructuredDescription{Type: "discarded", Value: strings.Join(intToStrings(roll), ",")})
					}
				}
			}
		}
	} else if !resultSuppressed {
		desc = append(desc, StructuredDescription{Type: rollType, Value: strings.Join(intToStrings(d.History[0]), ",")})
	}

	if d.Label != "" {
		desc = append(desc, StructuredDescription{Type: "label", Value: d.Label})
	}
	return desc
}

// New creates a new set of dice (using the low-level representation Dice type;
// for a more user-friendly interface use NewDieRoller instead).
//
// By default, this creates a d20 you can roll. For other kinds of die rolls,
// pass the option(s) ByDescription(description), ByDieType(qty, sides, bonus),
// WithDieBonus(n), WithDiv(n), WithFactor(n), WithGenerator(source), and/or
// WithSeed(s).
//
// For example,
//
//	d, err := New()
//	d, err := New(ByDescription("15d6 + 12"))
//	d, err := New(ByDieType(15, 6, 12))
func New(options ...func(*Dice) error) (*Dice, error) {
	d := new(Dice)

	for _, option := range options {
		err := option(d)
		if err != nil {
			return nil, fmt.Errorf("error setting options in dice.New(): %v", err)
		}
	}

	if d.desc != "" {
		//
		// some up-front error checking
		//
		usedThreat, err := regexp.MatchString(`\bc(\d+)?([-+]\d+)?\b`, d.desc)
		if err != nil {
			return nil, err
		}
		if usedThreat {
			return nil, fmt.Errorf("confirmation specifier (c[threat][±bonus]) not allowed in this location. It must be at the end of a full DieRoller description string only")
		}
		//
		// compile regular expressions
		//
		reMin := regexp.MustCompile(`^\s*min\s*([+-]?\d+)\s*$`)
		reMax := regexp.MustCompile(`^\s*max\s*([+-]?\d+)\s*$`)
		reMinmax := regexp.MustCompile(`\b(min|max)\s*[+-]?\d+`)
		reOpSplit := regexp.MustCompile(`[-+*×÷()≤≥]|[^-+*×÷()≤≥]+`)
		reIsOp := regexp.MustCompile(`^[-+*×÷()≤≥]$`)
		reIsDie := regexp.MustCompile(`\d+\s*[dD]\d*\d+`)
		reIsWS := regexp.MustCompile(`^\s+$`)
		reIsBareLabel := regexp.MustCompile(`^\s*([\p{L}_][\p{L}\p{N}_,.]*\s*)+\s*$`)
		reConstant := regexp.MustCompile(`^\s*(\d+(?:\.\d+)?|\.\d+)\s*(.*?)\s*$`)
		//                                  max?    numerator    denominator       sides          best/worst         rerolls   label
		//                                   _1_    __2__          __3__            __4___       _____5_____         __6__     __7__
		reDieSpec := regexp.MustCompile(`^\s*(>)?\s*(\d*)\s*(?:/\s*(\d+))?\s*[Dd]\s*(%|\d+)\s*(?:(best|worst)\s*of\s*(\d+))?\s*(.*?)\s*$`)

		//
		// break apart the major pieces separated by |
		// here, the first is the basic die spec. The others may be "min" or "max"
		//
		majorPieces := strings.Split(d.desc, "|")
		if len(majorPieces) == 0 {
			return nil, fmt.Errorf("apparently empty dice expression")
		}

		if len(majorPieces) > 1 {
			d.desc = strings.TrimSpace(majorPieces[0])
			for _, modifier := range majorPieces[1:] {
				if m := reMin.FindStringSubmatch(modifier); m != nil {
					d.MinValue, err = strconv.Atoi(m[1])
					if err != nil {
						return nil, err
					}
				} else if m := reMax.FindStringSubmatch(modifier); m != nil {
					d.MaxValue, err = strconv.Atoi(m[1])
					if err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("invalid global modifier %s", modifier)
				}
			}
		}

		//
		// The die spec is a number of basic die rolls or constants separated
		// by math operators +, -, *, or //. A leading constant 0 is assumed if
		// the expression starts with an operator.
		// We'll support the use of the unicode ÷ character in place of the older
		// ASCII "//" operator as well.
		//
		expr := strings.Replace(d.desc, "//", "÷", -1)
		expr = strings.Replace(expr, ">=", "≥", -1)
		expr = strings.Replace(expr, "<=", "≤", -1)
		exprParts := reOpSplit.FindAllString(expr, -1)
		expectedSyntax := "[<n>[/<d>]] d [<sides>|%] [best|worst of <n>] [+|-|*|×|÷|//|<=|>=|≤|≥ ...] ['|'min <n>] ['|'max <n>]"

		if len(exprParts) == 0 {
			return nil, fmt.Errorf("syntax error in die roll description \"%s\"; should be \"%s\"", d.desc, expectedSyntax)
		}

		//
		// exprParts is a list of alternating operators and values.
		// We'll run through the list, building up a sequence of dieComponents
		// to represent the expression we were given.
		//
		opExpected := false
		diceCount := 0
		for _, part := range exprParts {
			if opExpected {
				// we are expecting an operator here; collect it and go to the next part
				if reIsWS.MatchString(part) {
					// ignore white space while expecting an operator (issue #19)
					continue
				}

				if !reIsOp.MatchString(part) {
					// If it's a bare label, allow it here
					if reIsBareLabel.MatchString(part) {
						if reDieSpec.MatchString(part) {
							return nil, fmt.Errorf("\"%v\" looks suspiciously like a die-roll specification but appears as a label; did you forget an operator?", part)
						}
						var bareLabel dieLabel = (dieLabel)(strings.TrimSpace(part))
						d.multiDice = append(d.multiDice, &bareLabel)
						continue
					}
					return nil, fmt.Errorf("expected operator before \"%v\" in die-roll expression", part)
				}

				var thisOp rune
				for _, r := range part {
					if r == '*' {
						thisOp = '×'
					} else {
						thisOp = r
					}
					break
				}

				switch thisOp {
				case '(':
					return nil, fmt.Errorf("expected operator before '(' in die-roll expression")
				case ')':
					d.multiDice = append(d.multiDice, new(dieEndGroup))
					continue
				default:
					var do dieOperator = (dieOperator)(thisOp)
					d.multiDice = append(d.multiDice, &do)
				}
				opExpected = false
				continue
			}
			// we're not expecting to see an operator here, so if we see a + or -, they must
			// be a leading unary + or - to apply to what comes next.
			if reIsWS.MatchString(part) {
				continue
			}
			if part == "+" {
				// and a unary + is essentially a no-op, so we'll just ignore it.
				continue
			}
			if part == "-" {
				var do dieOperator = (dieOperator)('‾')
				d.multiDice = append(d.multiDice, &do)
				continue
			}
			if part == "(" {
				// an open bracket is ok here as well
				d.multiDice = append(d.multiDice, new(dieBeginGroup))
				continue
			}
			// any other operator here is illegal
			if reIsOp.MatchString(part) {
				return nil, fmt.Errorf("unexpected operator \"%v\" in die-roll expression", part)
			}

			// whatever comes next is a value, so switch to looking for an operator after it.
			opExpected = true

			xValues := reDieSpec.FindStringSubmatch(part)
			if xValues == nil {
				//
				// If this is just a constant, this will be easy.
				//
				xValues = reConstant.FindStringSubmatch(part)
				if xValues != nil {
					v, err := strconv.ParseFloat(xValues[1], 64)
					if err != nil {
						return nil, fmt.Errorf("value error in die roll subexpression \"%s\" in \"%s\"; %v", part, d.desc, err)
					}
					labelText := strings.TrimSpace(xValues[2])
					if labelText != "" && !reIsBareLabel.MatchString(labelText) {
						return nil, fmt.Errorf("constant label \"%v\" has illegal characters", labelText)
					}
					d.multiDice = append(d.multiDice, &dieConstant{Value: v, Label: labelText})
					continue
				}
				//
				// Ok, doesn't look valid then.
				//
				return nil, fmt.Errorf("syntax error in die roll subexpression \"%s\" in \"%s\"; should be \"%s\"", part, d.desc, expectedSyntax)
			}

			//
			// Otherwise, we'll look for a complex die-roll specification,
			// but up front we'll make sure they didn't try to sneak in an
			// option that belongs at the end of the expression, not buried
			// here in one of these.
			//
			if reMinmax.MatchString(part) {
				return nil, fmt.Errorf("syntax error in die roll subexpression \"%s\" in \"%s\"; min/max limits must appear after the final operator in the expression, since they apply to the entire set of dice rolls", part, d.desc)
			}

			//
			// Ok, now let's digest the more complex die-roll spec pattern
			// and construct a dieSpec to describe it.
			//
			ds := &dieSpec{generator: d.generator}
			d._onlydie = ds
			diceCount++
			if xValues[1] != "" {
				ds.InitialMax = true
			}
			if xValues[2] != "" {
				ds.Numerator, err = strconv.Atoi(xValues[2])
				if err != nil {
					return nil, fmt.Errorf("value error in die roll subexpression \"%s\": %v", part, err)
				}
			} else {
				ds.Numerator = 1
			}
			if xValues[3] != "" {
				ds.Denominator, err = strconv.Atoi(xValues[3])
				if err != nil {
					return nil, fmt.Errorf("value error in die roll subexpression \"%s\": %v", part, err)
				}
			}
			if xValues[4] == "%" {
				ds.Sides = 100
			} else {
				ds.Sides, err = strconv.Atoi(xValues[4])
				if err != nil {
					return nil, fmt.Errorf("value error in die roll subexpression \"%s\": %v", part, err)
				}
			}
			if xValues[5] != "" {
				ds.Rerolls, err = strconv.Atoi(xValues[6])
				if err != nil {
					return nil, fmt.Errorf("value error in die roll subexpression \"%s\": %v", part, err)
				}
				ds.Rerolls--
				switch xValues[5] {
				case "best":
					ds.BestReroll = true
				case "worst":
					ds.BestReroll = false
				default:
					return nil, fmt.Errorf("value error in die roll subexpression \"%s\": expecting \"best\" or \"worst\"", part)
				}
			}
			if xValues[7] != "" {
				if reIsDie.MatchString(xValues[7]) {
					return nil, fmt.Errorf("label following die roll in \"%s\" looks like another die roll--did you forget an operator?", part)
				}
				if !reIsBareLabel.MatchString(xValues[7]) {
					return nil, fmt.Errorf("label \"%v\" has illegal characters", xValues[7])
				}
				ds.Label = strings.TrimSpace(xValues[7])
			}
			d.multiDice = append(d.multiDice, ds)
		}

		if !opExpected {
			return nil, fmt.Errorf("missing value after last operator in die-roll expression \"%s\"", d.desc)
		}
		if diceCount != 1 {
			d._onlydie = nil
		}
	}

	if d.qty > 0 && d.sides > 0 {
		d.multiDice = append(d.multiDice, &dieSpec{
			Numerator:   d.qty,
			Sides:       d.sides,
			DieBonus:    d.diebonus,
			Denominator: d.div,
			generator:   d.generator,
		})
	}
	if d.bonus < 0 {
		var do dieOperator = (dieOperator)('-')
		d.multiDice = append(d.multiDice, &do, &dieConstant{Value: float64(-d.bonus)})
	} else if d.bonus > 0 {
		var do dieOperator = (dieOperator)('+')
		d.multiDice = append(d.multiDice, &do, &dieConstant{Value: float64(d.bonus)})
	}

	if d.factor != 0 {
		var do dieOperator = (dieOperator)('×')
		var md []dieComponent
		md = append(md, new(dieBeginGroup))
		md = append(md, d.multiDice...)
		d.multiDice = append(md, new(dieEndGroup), &do, &dieConstant{Value: float64(d.factor)})
	}

	return d, nil
}

// Roll rolls the dice which this Dice instance represents. The result is
// returned as an integer value.  Each time  this  is  called,  the
// dice are rerolled to get a new result.  The Dice value’s internal
// state reflects the last call to this method.
func (d *Dice) Roll() (int, error) {
	return d.RollToConfirm(false, 0, 0)
}

// MaxRoll is an alternative to Roll where
// instead of rolling the dice, it just assumes they all came up at their maximum
// possible values. This does NOT set up for subsequent critical rolls.
func (d *Dice) MaxRoll() (int, error) {
	return d.MaxRollToConfirm(0)
}

// MaxRollToConfirm is the analog to RollToConfirm but just assumes all
// dice come up with their maximum values rather than rolling anything.
func (d *Dice) MaxRollToConfirm(bonus int) (int, error) {
	d._natural = 0
	d.Rolled = false
	var err error
	stack := &evalStack{}

	for _, die := range d.multiDice {
		if err := die.computeMaxValue(stack); err != nil {
			return 0, err
		}
	}
	rollSum, err := stack.evaluate()
	if err != nil {
		return 0, err
	}

	rollSum += bonus
	if d.MaxValue > 0 && rollSum > d.MaxValue {
		rollSum = d.MaxValue
	}
	if d.MinValue > 0 && rollSum < d.MinValue {
		rollSum = d.MinValue
	}

	d.LastValue = rollSum
	d.Rolled = true
	return rollSum, nil
}

// RollToConfirm rolls the dice specified in the Dice value, with support
// for making critical confirmation rolls.
//
// If confirm is true, we're rolling to confirm a critical threat.
// In this case, the previous-rolled value is checked against the
// provided threat value. If that previous roll is less than the threat
// value, nothing further is done, and 0 is returned.
//
// If the previous roll was ≥ threat, then we make another roll, adding
// the provided bonus modifier to that roll. This roll's total is returned,
// and becomes the new most-recent roll for this Dice value.
//
// If threat is less than or equal to zero, the default threat of a natural
// maximum roll (e.g., 20 on a d20, or 10 on a d10) is used.
//
// Calling d.RollToConfirm(false, 0, 0) is equivalent to calling d.Roll().
// Confirm a critical roll by rolling again against the normal to-hit target.
func (d *Dice) RollToConfirm(confirm bool, threat int, bonus int) (int, error) {
	if confirm {
		// we're confirming if the previous roll was critical; so first of
		// all, there needs to have been one to confirm.
		// if d._natural was 0, we haven't rolled the other roll yet;
		// if it's < 0, we have but it doesn't qualify for confirmation.
		if d._natural == 0 {
			return 0, fmt.Errorf("you need to roll the dice first before confirming a critical roll")
		}
		if d._natural < 0 {
			return 0, fmt.Errorf("you can't confirm a critical on this roll because it doesn't involve only a single die")
		}

		//
		// default the critical threat range to the maximum face
		// of the die (e.g., a natural 20 on a d20)
		//
		if threat <= 0 {
			threat = d._defthreat
		}
		//
		// Now check if the previous roll was in the threat range
		//
		if d._natural < threat {
			return 0, nil // nothing here to confirm
		}
	}
	//
	// If we get this far, we are either rolling a regular die roll,
	// or trying to confirm a critical that we know needs to be
	// confirmed.
	//
	d._natural = 0
	d.Rolled = false
	stack := &evalStack{}
	var err error

	for _, die := range d.multiDice {
		if err = die.compute(stack); err != nil {
			return 0, err
		}

		// If we happen to be rolling for the first time, leave a
		// note for the confirming roll as to the natural die value
		// here.
		// if naturalRoll() returns -1, then this whole roll is disqualified
		// from confirmation: thus d._natural will be -1 from then on to indicate
		// that. Otherwise if we end up with multiple nonzero values, we are
		// also disqualified due to multiple dice involved, also setting
		// d._natural to -1.
		// Otherwise d._natural will be 0 (no dice involved at all) or the
		// single natural die roll we found.
		theNatural, defThreat := die.naturalRoll()
		if theNatural != 0 {
			if d._natural == 0 {
				d._natural, d._defthreat = theNatural, defThreat
			} else {
				d._natural, d._defthreat = -1, 0
			}
		}
	}
	rollSum, err := stack.evaluate()
	if err != nil {
		return 0, err
	}
	rollSum += bonus
	if d.MaxValue > 0 && rollSum > d.MaxValue {
		rollSum = d.MaxValue
	}
	if d.MinValue > 0 && rollSum < d.MinValue {
		rollSum = d.MinValue
	}

	d.LastValue = rollSum
	d.Rolled = true
	return rollSum, nil
}

// Description produces a human-readable description of the die roll specification represented
// by the Dice object.
func (d *Dice) Description() (desc string) {
	for _, die := range d.multiDice {
		desc += die.description()
	}
	if d.MinValue > 0 {
		desc += fmt.Sprintf(" min %d", d.MinValue)
	}
	if d.MaxValue > 0 {
		desc += fmt.Sprintf(" max %d", d.MaxValue)
	}
	return
}

type sdrOptions struct {
	autoSF           bool
	resultSuppressed bool
	rollBonus        int
	successMessage   string
	failureMessage   string
}

// WithRollBonus augments the operation of a StructuredDescribeRoll call by
// indicating that this roll included an extra bonus (not indicated by the user).
//
// This is not done if the bonus is 0 (zero).
func WithRollBonus(bonus int) func(*sdrOptions) {
	return func(o *sdrOptions) {
		o.rollBonus = bonus
	}
}

// WithAutoSF augments the operation of a StructuredDescribeRoll call by
// indicating that a natural 1 indicates failure regardless of the modified
// result's value, and a natural maximum roll similarly indicates success.
//
// In case an automatic success is indicated, the provided successMessage string
// will be included in the structured description; likewise with failureMessage
// in case of automatic failure.
//
// This behavior is enabled if the enabled parameter is true. Otherwise
// no automatic success or failure interpretation will be made.
func WithAutoSF(enabled bool, successMessage, failureMessage string) func(*sdrOptions) {
	return func(o *sdrOptions) {
		o.autoSF = enabled
		o.successMessage = successMessage
		o.failureMessage = failureMessage
	}
}

// WithNoResults tells StructuredDescribeRoll to assume that no actual die roll was
// made, and report only the request itself, ignoring results.
func WithNoResults() func(*sdrOptions) {
	return func(o *sdrOptions) {
		o.resultSuppressed = true
	}
}

// StructuredDescribeRoll produces a detailed structured description of the result of rolling
// the Dice, in a way that a caller can format as they see fit.
//
// If you wish to interpret the die roll in light of a rule which allows for
// a natural 1 to indicate automatic failure and a natural maximum die face
// (e.g., natural 20 on a d20) to indicate automatic success, then add a
// WithAutoSF option function call to the end of the argument list.
//
// If this die roll included a bonus added to it for some reason (e.g., a
// confirmation bonus on dice rolled to confirm a critical threat),
// then add a WithRollBonus function call to the end of the argument list.
func (d *Dice) StructuredDescribeRoll(options ...func(*sdrOptions)) ([]StructuredDescription, error) {
	var desc []StructuredDescription
	var opts sdrOptions

	for _, o := range options {
		o(&opts)
	}

	if !d.Rolled && !opts.resultSuppressed {
		return nil, nil
	}
	if !opts.resultSuppressed {
		if opts.autoSF {
			if d._onlydie == nil || d._onlydie.Numerator != 1 {
				return nil, fmt.Errorf("you can't indicate auto-success/fail (|sf option) because it involves multiple dice")
			}
			if d._onlydie.isMinRoll() {
				desc = append(desc, StructuredDescription{Type: "fail", Value: opts.failureMessage})
			} else if d._onlydie.isMaxRoll() {
				desc = append(desc, StructuredDescription{Type: "success", Value: opts.successMessage})
			}
		}

		desc = append(desc,
			StructuredDescription{Type: "result", Value: strconv.Itoa(d.LastValue)},
			StructuredDescription{Type: "separator", Value: "="},
		)
	}
	for _, die := range d.multiDice {
		desc = append(desc, die.structuredDescribeRoll(opts.resultSuppressed)...)
	}
	if opts.rollBonus != 0 {
		desc = append(desc, StructuredDescription{Type: "bonus", Value: fmt.Sprintf("%+d", opts.rollBonus)})
	}
	if d.MinValue != 0 {
		desc = append(desc,
			StructuredDescription{Type: "moddelim", Value: "|"},
			StructuredDescription{Type: "min", Value: strconv.Itoa(d.MinValue)},
		)
	}
	if d.MaxValue != 0 {
		desc = append(desc,
			StructuredDescription{Type: "moddelim", Value: "|"},
			StructuredDescription{Type: "max", Value: strconv.Itoa(d.MaxValue)},
		)
	}
	return desc, nil
}

//////////////////////////////////////////////////////////////////////////////
//  ____  _      ____       _ _
// |  _ \(_) ___|  _ \ ___ | | | ___ _ __
// | | | | |/ _ \ |_) / _ \| | |/ _ \ '__|
// | |_| | |  __/  _ < (_) | | |  __/ |
// |____/|_|\___|_| \_\___/|_|_|\___|_|
//

// The DieRoller type provides a higher-level view of die-roll generation,
// and should be generally preferred instead of the lower-level Dice type.
//
// This allows the caller to deal with die rolls in a friendly but expressive
// syntax as a string of characters such as "2d12 + 5 bonus - 1 size | dc 12"
// which game players are more accustomed to using, and then simply roll that
// die-roll expression as needed.
//
// Note that it is not expected for the user to set or query these structures
// directly. Use the provided functions instead.
type DieRoller struct {
	Confirm bool // Are we supposed to confirm potential critical rolls?
	DoMax   bool // Maximize all die rolls?

	// If we need to repeatedly roll dice, we will either do so RepeatFor
	// times (if > 0), or until the result meets or exceeds RepeatUntil
	// (again, if > 0)
	RepeatUntil int
	RepeatFor   int

	// If PctChance ≥ 0 then our target to be "successful" is a score
	// or at least PctChance on a percentile die roll. In that case
	// we can also set a label in PctLabel for that roll.
	PctChance int

	// If DC > 0 we're trying to meet a difficulty class for the
	// roll to be "successful".
	DC int

	critThreat int // --threat threshold (0=default for die type)
	critBonus  int // --added to confirmation rolls

	// User-defined label for this entire die-roll specification, such as
	// "Knowledge Skill Check".
	LabelText string

	// Messages to report if the roll can be categorized as "successful" or "failed".
	SuccessMessage string
	FailMessage    string

	// Template for permuted roll pattern substitution.
	Template string

	// Label to use if PctChance is in effect.
	PctLabel string

	sfOpt string // sf option part of source die-roll spec string or ""

	// Values to be substituted into the Template
	Permutations [][]any

	// Postfix expression(s) generated by the most recent roll
	Postfix []string

	generator *rand.Rand
	d         *Dice // underlying Dice object
}

// RandFloat64 generates a pseudorandom number in the range [0.0, 1.0) using
// the same random number generator as used for die rolls. This means that it
// affects the outcome of subsequent die rolls just as other die rolls do.
func (d *DieRoller) RandFloat64() float64 {
	if d.generator == nil {
		return rand.Float64()
	}
	return d.generator.Float64()
}

// RandIntn generates a pseudorandom integer in the range [0, n) using
// the same random number generator as used for die rolls. This means that it
// affects the outcome of subsequent die rolls just as other die rolls do.
func (d *DieRoller) RandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	if d.generator == nil {
		return rand.Intn(n)
	}
	return d.generator.Intn(n)
}

// NewDieRoller creates a new DieRoller value, which provides the recommended
// higher-level interface for rolling dice. This value can
// then be used for as many die rolls as needed by calling its DoRoll
// or DoRollOnce methods.
//
// You may pass zero or more option specifiers to this function as already
// described for the New constructor, although the only ones which apply
// here are WithSeed(s) and WithGenerator(s).
//
// Initially it is set up to roll a single d20, but this can be changed with
// each DoRoll call.
func NewDieRoller(options ...func(*Dice) error) (*DieRoller, error) {
	var err error
	dr := new(DieRoller)
	opts := new(Dice)

	for _, option := range options {
		err := option(opts)
		if err != nil {
			return nil, fmt.Errorf("error setting options in dice.NewDieRoller(): %v", err)
		}
	}

	if opts.generator != nil {
		dr.generator = opts.generator
	}

	dr.d, err = New(ByDieType(1, 20, 0), withSharedGenerator(dr.generator))
	if err != nil {
		return nil, err
	}

	return dr, nil
}

func (d *DieRoller) setNewSpecification(spec string) error {
	var err error

	d.d = nil
	d.LabelText = ""
	d.Confirm = false
	d.critThreat = 0
	d.critBonus = 0
	d.sfOpt = ""
	d.SuccessMessage = ""
	d.FailMessage = ""
	d.Template = ""
	d.Permutations = nil
	d.RepeatUntil = 0
	d.RepeatFor = 1
	d.DoMax = false
	d.DC = 0
	d.PctChance = -1
	d.PctLabel = ""

	reLabel := regexp.MustCompile(`^\s*(.*?)\s*=\s*(.*?)\s*$`)
	reModMinmax := regexp.MustCompile(`^\s*(min|max)\s*[+-]?\d+`)
	reModConfirm := regexp.MustCompile(`^\s*c(\d+)?([-+]\d+)?\s*$`)
	reModUntil := regexp.MustCompile(`^\s*until\s*(-?\d+)\s*$`)
	reModRepeat := regexp.MustCompile(`^\s*repeat\s*(\d+)\s*$`)
	reModMaximized := regexp.MustCompile(`^\s*(!|maximized)\s*$`)
	reModDC := regexp.MustCompile(`^\s*[Dd][Cc]\s*(-?\d+)\s*$`)
	reModSF := regexp.MustCompile(`^\s*sf(?:\s+(\S.*?)(?:/(\S.*?))?)?\s*$`)
	rePermutations := regexp.MustCompile(`\{(.*?)\}`)
	rePctRoll := regexp.MustCompile(`^\s*(\d+)%(.*)$`)

	//
	// Convert <= and >= so we don't confuse them with the = that indicates a title string
	//
	spec = strings.Replace(spec, ">=", "≥", -1)
	spec = strings.Replace(spec, "<=", "≤", -1)
	//
	// Look for leading "<label>="
	//
	fields := reLabel.FindStringSubmatch(spec)
	if fields != nil {
		spec = fields[2]
		d.LabelText = fields[1]
	}

	//
	// The remainder of the spec is a die-roll string followed by a number
	// of global modifiers, separated by vertical bars.
	//
	majorPieces := strings.Split(spec, "|")
	if len(majorPieces) == 0 {
		return fmt.Errorf("empty dice description")
	}

	//
	// If there are modifiers, process them now.
	//
	if len(majorPieces) > 1 {
		spec = strings.TrimSpace(majorPieces[0])
		for i := 1; i < len(majorPieces); i++ {
			if reModMinmax.MatchString(majorPieces[i]) {
				// min/max options need to be passed down to the Dice parser,
				// so append it to the diespec string
				spec += "|" + majorPieces[i]
			} else {
				if fields := reModConfirm.FindStringSubmatch(majorPieces[i]); fields != nil {
					//
					// MODIFIER
					// 	| c[<threat>][{+|-}<bonus>]
					// critical roll confirmation specifier
					//
					d.Confirm = true
					if fields[1] != "" {
						d.critThreat, err = strconv.Atoi(fields[1])
						if err != nil {
							return fmt.Errorf("value error in die roll confirm expression: %v", err)
						}
					}
					if fields[2] != "" {
						d.critBonus, err = strconv.Atoi(fields[2])
						if err != nil {
							return fmt.Errorf("value error in die roll confirm expression: %v", err)
						}
					}
					//
					// If there wasn't something more explicitly defined,
					// a critical confirmation roll uses HIT/MISS as defaults.
					//
					if d.SuccessMessage == "" {
						d.SuccessMessage = "HIT"
					}
					if d.FailMessage == "" {
						d.FailMessage = "MISS"
					}
				} else if fields := reModUntil.FindStringSubmatch(majorPieces[i]); fields != nil {
					//
					// MODIFIER
					//  | until <n>
					// Repeat rolling until reaching limit <n>
					//
					d.RepeatUntil, err = strconv.Atoi(fields[1])
					if err != nil {
						return fmt.Errorf("value error in die roll until clause: %v", err)
					}
				} else if fields := reModRepeat.FindStringSubmatch(majorPieces[i]); fields != nil {
					//
					// MODIFIER
					//  | repeat <n>
					// Repeat the die roll <n> times
					//
					d.RepeatFor, err = strconv.Atoi(fields[1])
					if err != nil {
						return fmt.Errorf("value error in die roll repeat clause: %v", err)
					}
				} else if reModMaximized.MatchString(majorPieces[i]) {
					//
					// MODIFIER
					//  | !|maximized
					// Maximize all die rolls
					//
					d.DoMax = true
				} else if fields := reModDC.FindStringSubmatch(majorPieces[i]); fields != nil {
					//
					// MODIFIER
					//  | DC <n>
					// Seek a value at least <n>
					//
					d.DC, err = strconv.Atoi(fields[1])
					if err != nil {
						return fmt.Errorf("value error in die roll DC clause: %v", err)
					}
				} else if fields := reModSF.FindStringSubmatch(majorPieces[i]); fields != nil {
					//
					// MODIFIER
					//  | sf [<success>[/<fail>]]
					// Set messages for successful and failed rolls.
					//
					d.sfOpt = fields[0]
					if fields[1] != "" {
						d.SuccessMessage = fields[1]
						if fields[2] != "" {
							d.FailMessage = fields[2]
						} else {
							// Guess the failure message based on the success
							// message.
							switch strings.ToLower(d.SuccessMessage) {
							case "hit":
								d.FailMessage = "MISS"
							case "miss":
								d.FailMessage = "HIT"
							case "success", "succeed":
								d.FailMessage = "FAIL"
							case "fail":
								d.FailMessage = "SUCCESS"
							default:
								d.FailMessage = "NOT " + d.SuccessMessage
							}
						}
					} else {
						d.SuccessMessage = "SUCCESS"
						d.FailMessage = "FAIL"
					}
				} else {
					return fmt.Errorf("global modifier option \"%s\" not understood; must be !, c, dc, min, max, maximized, sf, until, or repeat", majorPieces[i])
				}
			}
		}
	}

	//
	// The global options are all taken care of.
	// What remains is the die roll spec itself, which may include
	// permutations that we'll need to expand here.
	//
	// If there are one or more patterns like {<a>/<b>/.../<z>} in the
	// string, make a copy of the spec for each of <a>, <b>, ... <z> in that
	// position in the string. This will produce the cartesian product
	// of the sets of values, e.g. "d20+{15/10/5}+2d6+{1/2}" will expand
	// to:
	//  "d20+15+2d6+1"
	//  "d20+10+2d6+1"
	//  "d20+5+2d6+1"
	//  "d20+15+2d6+2"
	//  "d20+10+2d6+2"
	//  "d20+5+2d6+2"
	//
	spec = strings.Replace(spec, "//", "÷", -1)
	if permList := rePermutations.FindAllStringSubmatch(spec, -1); permList != nil {
		for _, perm := range permList {
			valueset := strings.Split(perm[1], "/")
			if len(valueset) < 2 {
				return fmt.Errorf("invalid die-roll specification \"%s\": Values in braces must have more than one value separated by slashes", perm[0])
			}
			plist := make([]any, len(valueset))
			for i, p := range valueset {
				plist[i] = p
			}
			d.Permutations = append(d.Permutations, plist)
		}
		//
		// replace the {...} strings with placeholder tokens {0}, {1}, ... {n}
		// to form a template into which we'll substitute all of the permuted values
		// out of d.Permutations.
		//
		pos := -1
		d.Template = rePermutations.ReplaceAllStringFunc(spec, func(_ string) string {
			pos++
			return "{" + strconv.Itoa(pos) + "}"
		})
	}

	if fields := rePctRoll.FindStringSubmatch(spec); fields != nil {
		//
		// Special case: <n>% rolls percentile dice and
		// returns true with a probability of n%.
		//
		if d.Permutations != nil {
			return fmt.Errorf("permutations with percentile die rolls are not supported")
		}
		if strings.Index(spec, "|") >= 0 {
			return fmt.Errorf("invalid global modifier for percentile die rolls: \"%s\"", spec)
		}
		if d.Confirm {
			return fmt.Errorf("you can't confirm critical percentile die rolls")
		}
		if d.DC != 0 {
			return fmt.Errorf("you can't have a percentile die roll with a DC")
		}
		d.d, err = New(ByDieType(1, 100, 0), withSharedGenerator(d.generator))
		if err != nil {
			return err
		}
		d.PctChance, err = strconv.Atoi(fields[1])
		if err != nil {
			return err
		}
		d.PctLabel = fields[2]
	} else if d.Template == "" {
		//
		// Normal case: use the remaining string in spec to define a Dice object
		// that we will subsequently roll using our local modifiers and such.
		//
		d.d, err = New(ByDescription(spec), withSharedGenerator(d.generator))
		if err != nil {
			return err
		}
	}
	return nil
}

// DoRoll rolls dice as described by the specification string. If this string is empty,
// it re-rolls the previously-used specification. Initially, "1d20" is assumed.
//
// Returns the user-specified die-roll label (if any), the result of the roll,
// and an error if one occurred.
//
// In this more comprehensive interface, the spec string is a string of
// the form
//
//	[<title>=] <expression> [|<options>...]
//
// or
//
//	[<title>=] <chance>% [<success>[/<fail>]] [|<options>...]
//
// where []s indicate optional parameters, the given words in angle brackets
// (<title>, etc) represent values to be placed into the string, and the other
// characters are to be taken literally.
//
// The <title> (which, if given, is separated from the rest of the spec with an
// equals sign (“=”)) is optional and will be included as a comment in the result
// list to indicate what the purpose of the die roll was for.
//
// Note that this module does not interpret the <title> value further, but by
// convention two special characters are significant to some clients:
//
//	‖ (U+2016) separates multiple titles in the <title> string
//	≡ (U+2261) separates the title text on the left with a color on the right.
//
// This means that a <title> string of "monster≡blue‖damage≡red" will display
// a title for the die roll as two separate title values, "monster" in blue
// and "damage" in red.
//
// <expression> can be anything that can be given as the description string
// to the New constructor (q.v.). At the end of the spec string there
// may be zero or more options, each beginning with a vertical bar (“|”).
//
// These options may be any of the following:
//
//	| min <n>
//
// The result will be at least <n>.
//
//	| max <n>
//
// The result will be no more than <n>.
//
//	| c[<t>[±<b>]]
//
// This indicates that the roll may need a critical
// confirmation roll to follow it. This will appear
// as an additional result in the list of results returned
// from the DoRoll and DoRollOnce methods.  If the <t> parameter is given,
// a natural roll equal to or greater than <t> is assumed to
// be a critical threat. If a plus or minus sign followed by
// a number <b> is appended to the option, then this value is
// added to the confirmation die roll as a confirmation bonus.
// (The notation "±" here means either a "-" or "+" may appear at that
// position in the string.)
//
//	| dc <n>
//
// This is a roll against a known difficulty class <n>. If the
// result is at least <n>, the roll is "successful".
//
//	| sf [<success>[/<fail>]]
//
// Auto-success/fail: the roll, which must involve only a single
// die, will be considered successful if it's a natural maximum
// value (e.g., 20 on a d20 before modifiers are applied), or
// a failure if a natural 1 was rolled. Optionally, messages
// to report to the user to indicate what success and failure mean
// may be specified. Suitable defaults will be used or derived if
// one or both of those strings is not given.
//
//	| until <n>
//
// Continue making die rolls, adding their results to the returned output,
// until a result of at least <n> is obtained.
//
//	| repeat <n>
//
// Make <n> die rolls, reporting their results.
//
//	| maximized
//
// Assume all dice roll at their maximum possible values. For example,
// the spec "3d6 | maximized" will always return the result 18, as if
// all three dice rolled sixes.
//
// To prevent getting caught in an infinite loop, a maximum of  100  rolls
// will be made regardless of repeat and until options.
//
// Anywhere  in  the  string  you may introduce a combination specifier in
// curly braces as “{<a>/<b>/<c>/...}”.  This will repeat the overall die roll
// expression once for each of the values <a>, <b>, <c>, etc., substituting each
// in turn for the braced list. If multiple specifiers appear, they’ll all
// repeat so you get the Cartesian product of all the sets of values. This
// allows, for example, multiple attack rolls in a single click. For example,
// “Attack=d20+{17/12/7}”  would  roll  three  attack rolls: d20+17,
// d20+12, and d20+7.
//
// In the second form for the spec string,
// <chance> gives the  percentage  chance  of  something
// occurring,  causing  percentile dice to be rolled. The result will be a
// the integer value 1 if the d100 roll was less than or equal to
// <chance>, or 0 otherwise. By default, the result is reported in the detailed description
// as “success” or “fail”. If a success label is  given  in  the  die‐roll
// string,  that  is  reported  in case of a successful roll, and “did not
// <success>” otherwise. If an explicit fail label is also  given,  that  is
// used  for unsuccessful rolls instead.  As a special case, if <success> is
// “miss” then <fail> is assumed to be “hit” and vice versa.
//
// For percentile rolls, only the  until,  repeat,  and  maximized  global
// options  may  be  used. Permutations (“{...}”) are also disallowed with
// percentile rolls.
//
// This method returns values title, resultSet, and error representing
// the  results  of rolling the dice.  The title is the title specified in
// the dice string, or an empty string if one was not given.
// The  resultSet is a slice of StructuredResult structures, one for each
// roll of the dice that was performed.
//
// Example die-roll specifications:
//
//	"d20"             Roll 1d20.
//	"3d6"             Roll 3d6 (3 six-sided dice, adding their values).
//	"15d6+15"         Roll 15d6, add 15 to their sum.
//	"1d10+5*10"       Roll 1d10, add 5, then multiply the result by 10.
//	"1/2 d6"          Roll 1d6, divide result by 2 (truncating toward zero).
//	"2d10+3d6+12"     Roll 2d10, 3d6, add their results and add 12 to the sum.
//	"d20+15|c"        Roll d20+15, automatically rolling to confirm on a natural 20.
//	"d20+15|c19+2"    Roll d20+15, rolling to confirm on natural 19 or 20 with +2 bonus.
//	"d%"              Roll percentile dice, giving result 1-100.
//	"40%"             Roll percentile dice, giving result 1 with 40% probability.
//	"d20+12|max20"    Roll d20+12 but any result > 20 is capped at 20.
//	"d20 best of 2"   Roll d20 twice, discarding the worse result.
//	"d20+4|dc 10"     Roll d20+4, signalling success if the result is 10 or greater.
//	"3d6 fire+1d4 acid+2 bonus"
//	                  Roll 3d6+1d4+2. In the structured results, it will show the values
//	                  rolled for the 3d6 fire, 1d4 acid, and 2 bonus individually.
//	"40% hit"         Reports success ("hit") with a 40% probability; otherwise reports
//	                  failure ("miss").
//	"13% red/blue"    Reports success ("red") with a 13% probability; otherwise reports
//	                  failure ("blue").
//	"2d10+3|until 19" Repeatedly rolls 2d10+3, adding each result to the set of die rolls
//	                  returned, until a roll totals at least 19.
func (d *DieRoller) DoRoll(spec string) (string, []StructuredResult, error) {
	var err error
	//
	// If we're given a new specification, reset our internals to roll according
	// to that spec until we are called with another non-null spec string.
	//

	if spec != "" {
		err = d.setNewSpecification(spec)
		if err != nil {
			return "", nil, err
		}
	}

	var overallResults []StructuredResult
	var results []StructuredResult
	var result int

	repeatIter := 0
	repeatCount := 0
	for repeatIter < d.RepeatFor {
		if d.Template != "" {
			// If we're working with a set of permutations, expand them now
			// into their Cartesian product so we can then substitute each set
			// of those values into the template for each roll of the dice.
			iterlist := cartesian.Iter(d.Permutations...)
			for iteration := range iterlist {
				d.d, err = New(
					ByDescription(substituteTemplateValues(d.Template, iteration)),
					withSharedGenerator(d.generator))
				if err != nil {
					return "", nil, err
				}
				result, results, err = d.rollDice(repeatIter, repeatCount)
				if err != nil {
					return "", nil, err
				}
				overallResults = append(overallResults, results...)
			}
		} else {
			// Otherwise we already have the Dice object set up, just use it.
			result, results, err = d.rollDice(repeatIter, repeatCount)
			if err != nil {
				return "", nil, err
			}
			overallResults = append(overallResults, results...)
		}

		if d.RepeatUntil == 0 || result >= d.RepeatUntil {
			repeatIter++
		}
		repeatCount++
		if repeatCount >= 100 {
			break
		}
	}
	if d.Template != "" {
		d.d = nil
	}

	return d.LabelText, overallResults, nil
}

// ExplainSecretRoll takes a dieroll spec as DoRoll does, along with a string explaining
// why the roll is secret. It returns the same result data as DoRoll, including a broken-out
// description of the dieroll spec, except that no actual die roll is made and no results
// are reported.
//
// The ResultSuppressed field of the returned StructuredResult value is set to true to indicate
// that there is no actual result contained in the returned data.
//
// This is used to send a "receipt"
// back to the requester of a die roll that their request for the roll was sent on to another party
// (typically the GM) who will be the only person to see the actual results (which must be obtained
// by a separate roll to the DoRoll method).
func (d *DieRoller) ExplainSecretRoll(spec, notice string) (string, StructuredResult, error) {
	var thisResult []StructuredDescription

	if spec != "" {
		if err := d.setNewSpecification(spec); err != nil {
			return "", StructuredResult{}, err
		}
	}
	thisResult = append(thisResult, StructuredDescription{Type: "notice", Value: notice})

	//
	// How to report back on the options (aka modifiers) in play for the die roll.
	// This updates the thisResult value in-place.
	//
	reportOptions := func() {
		if d.Confirm {
			thisResult = append(thisResult, StructuredDescription{
				Type: "moddelim", Value: "|",
			})
			c := "c"
			if d.critThreat != 0 {
				c += strconv.Itoa(d.critThreat)
			}
			if d.critBonus != 0 {
				c += fmt.Sprintf("%+d", d.critBonus)
			}
			thisResult = append(thisResult, StructuredDescription{
				Type: "critspec", Value: c,
			})
		}
		if d.RepeatFor > 1 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "repeat", Value: strconv.Itoa(d.RepeatFor)},
			)
		}
		if d.RepeatUntil != 0 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "until", Value: strconv.Itoa(d.RepeatUntil)},
			)
		}
		if d.DC != 0 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "dc", Value: strconv.Itoa(d.DC)},
			)
		}
		if d.sfOpt != "" {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "sf", Value: d.sfOpt},
			)
		}
	}

	//
	// percentile rolls are reported specially.
	// The result will be 0 or 1 and we'll describe the outcome in words
	// like "hit" or "miss"
	//
	reportPctRoll := func(chance int, label string, maximized bool) {
		thisResult = nil

		thisResult = append(thisResult,
			StructuredDescription{Type: "diespec", Value: fmt.Sprintf("%d%%", chance)},
		)
		if label != "" {
			thisResult = append(thisResult,
				StructuredDescription{Type: "label", Value: strings.TrimSpace(label)})
		}
		if maximized {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "fullmax", Value: "maximized"},
			)
		}
	}

	//
	// Enough of the preliminaries, let's get working.
	//
	if d.d == nil {
		// Since this can happen if the die-roll had multiple results, let's fall
		// back to just reporting the raw die-roll spec as-sent.
		thisResult = append(thisResult, StructuredDescription{Type: "diespec", Value: spec})
		return "", StructuredResult{ResultSuppressed: true, Details: thisResult}, nil
	}

	// MAXIMIZED DIE ROLLS_____________________________________________________
	//
	// If we're maximizing rolls, we just assume every die came up at its
	// maximum value instead of bothering to roll them.
	//
	sfo := d.sfOpt
	if sfo == "" && d.Confirm {
		sfo = "c"
	}
	if d.DoMax {
		if d.PctChance >= 0 {
			reportPctRoll(d.PctChance, d.PctLabel, true)
		} else {
			sdesc, err := d.d.StructuredDescribeRoll(WithNoResults(), WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage))
			if err != nil {
				return "", StructuredResult{}, err
			}
			thisResult = append(thisResult, sdesc...)
			reportOptions()
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "fullmax", Value: "maximized"},
			)
		}
	} else {
		// NORMAL DIE ROLLS____________________________________________________
		//
		if d.PctChance >= 0 {
			reportPctRoll(d.PctChance, d.PctLabel, false)
		} else {
			sdesc, err := d.d.StructuredDescribeRoll(
				WithNoResults(),
				WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage))
			if err != nil {
				return "", StructuredResult{}, err
			}
			thisResult = append(thisResult, sdesc...)
			reportOptions()
		}
	}

	return d.LabelText, StructuredResult{ResultSuppressed: true, Details: thisResult}, nil
}

// utility function to replace placeholders {0}, {1}, {2}, ... in an input string
// with corresponding values taken from a list of substitution values, returning
// the resulting string.
func substituteTemplateValues(template string, values []any) string {
	result := template
	for place, value := range values {
		result = strings.Replace(result, fmt.Sprintf("{%d}", place), value.(string), 1)
	}
	return result
}

// This does the work of performing a die roll (possibly two, if we're confirming
// a critical roll) based on the exact specifications already set in place by
// the caller.
func (d *DieRoller) rollDice(repeatIter, repeatCount int) (int, []StructuredResult, error) {
	var results []StructuredResult
	var thisResult []StructuredDescription
	var result int
	var err error

	//
	// how to describe the results of a roll with a DC value.
	// in this case we want to indicate the margin above or below
	// the DC that was rolled.
	//
	describeDCRoll := func(dc, result int) (desc StructuredDescription) {
		if result > dc {
			desc.Type = "exceeded"
			desc.Value = strconv.Itoa(result - dc)
		} else if result == dc {
			desc.Type = "met"
			desc.Value = "successful"
		} else {
			desc.Type = "short"
			desc.Value = strconv.Itoa(dc - result)
		}
		return
	}

	//
	// How to report back on the options (aka modifiers) in play for the die roll.
	// This updates the thisResult value in-place.
	//
	reportOptions := func() {
		if d.Confirm {
			thisResult = append(thisResult, StructuredDescription{
				Type: "moddelim", Value: "|",
			})
			c := "c"
			if d.critThreat != 0 {
				c += strconv.Itoa(d.critThreat)
			}
			if d.critBonus != 0 {
				c += fmt.Sprintf("%+d", d.critBonus)
			}
			thisResult = append(thisResult, StructuredDescription{
				Type: "critspec", Value: c,
			})
		}
		if d.RepeatFor > 1 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "repeat", Value: strconv.Itoa(d.RepeatFor)},
				StructuredDescription{Type: "iteration", Value: strconv.Itoa(repeatCount + 1)},
			)
		}
		if d.RepeatUntil != 0 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "until", Value: strconv.Itoa(d.RepeatUntil)},
				StructuredDescription{Type: "iteration", Value: strconv.Itoa(repeatCount + 1)},
				describeDCRoll(d.RepeatUntil, result),
			)
		}
		if d.DC != 0 {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "dc", Value: strconv.Itoa(d.DC)},
				describeDCRoll(d.DC, result),
			)
		}
		if d.sfOpt != "" {
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "sf", Value: d.sfOpt},
			)
		}
	}

	//
	// percentile rolls are reported specially.
	// The result will be 0 or 1 and we'll describe the outcome in words
	// like "hit" or "miss"
	//
	reSlashDelim := regexp.MustCompile(`\s*/\s*`)
	reportPctRoll := func(chance int, label string, maximized bool) {
		thisResult = nil
		builtInLabels := map[string]string{
			"hit":  "miss",
			"miss": "hit",
		}
		var labels []string

		if label != "" {
			labels = reSlashDelim.Split(strings.TrimSpace(label), 2)
			if len(labels) == 1 {
				// user provided the positive string; we need to make up the other
				neg, ok := builtInLabels[labels[0]]
				if ok {
					labels = append(labels, neg)
				} else {
					labels = append(labels, "did not "+labels[0])
				}
			}
		} else {
			labels = []string{"success", "fail"}
		}
		if result <= chance {
			thisResult = append(thisResult, StructuredDescription{Type: "success", Value: labels[0]})
		} else {
			thisResult = append(thisResult, StructuredDescription{Type: "fail", Value: labels[1]})
		}
		thisResult = append(thisResult,
			StructuredDescription{Type: "separator", Value: "="},
			StructuredDescription{Type: "diespec", Value: fmt.Sprintf("%d%%", chance)},
		)
		if label != "" {
			thisResult = append(thisResult,
				StructuredDescription{Type: "label", Value: strings.TrimSpace(label)})
		}
		if maximized {
			thisResult = append(thisResult,
				StructuredDescription{Type: "maxroll", Value: strconv.Itoa(result)},
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "fullmax", Value: "maximized"},
			)
		} else {
			thisResult = append(thisResult,
				StructuredDescription{Type: "roll", Value: strconv.Itoa(result)})
		}
		if result <= chance {
			results = append(results, StructuredResult{Result: 1, Details: thisResult})
		} else {
			results = append(results, StructuredResult{Result: 0, Details: thisResult})
		}
	}

	//
	// Enough of the preliminaries, let's get working.
	//
	if d.d == nil {
		return 0, nil, fmt.Errorf("no defined Dice object to consume")
	}

	// MAXIMIZED DIE ROLLS_____________________________________________________
	//
	// If we're maximizing rolls, we just assume every die came up at its
	// maximum value instead of bothering to roll them.
	//
	sfo := d.sfOpt
	if sfo == "" && d.Confirm {
		sfo = "c"
	}
	if d.DoMax {
		result, err = d.d.MaxRoll()
		if err != nil {
			return 0, nil, err
		}
		if d.PctChance >= 0 {
			reportPctRoll(d.PctChance, d.PctLabel, true)
		} else {
			sdesc, err := d.d.StructuredDescribeRoll(WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage))
			if err != nil {
				return 0, nil, err
			}
			thisResult = append(thisResult, sdesc...)
			reportOptions()
			thisResult = append(thisResult,
				StructuredDescription{Type: "moddelim", Value: "|"},
				StructuredDescription{Type: "fullmax", Value: "maximized"},
			)
			if d.Confirm {
				results = append(results, StructuredResult{Result: result, Details: thisResult})
				result2, err := d.d.MaxRollToConfirm(d.critBonus)
				if err != nil {
					return 0, nil, err
				}
				thisResult = nil
				sdesc, err := d.d.StructuredDescribeRoll(
					WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage),
					WithRollBonus(d.critBonus))
				if err != nil {
					return 0, nil, err
				}
				thisResult = append(thisResult,
					StructuredDescription{Type: "critlabel", Value: "Confirm:"})
				thisResult = append(thisResult, sdesc...)
				thisResult = append(thisResult,
					StructuredDescription{Type: "moddelim", Value: "|"},
					StructuredDescription{Type: "fullmax", Value: "maximized"},
				)
				results = append(results, StructuredResult{Result: result2, Details: thisResult})
			} else {
				results = append(results, StructuredResult{Result: result, Details: thisResult})
			}
		}
	} else {
		// NORMAL DIE ROLLS____________________________________________________
		//
		result, err = d.d.Roll()
		if err != nil {
			return 0, nil, err
		}
		if d.PctChance >= 0 {
			reportPctRoll(d.PctChance, d.PctLabel, false)
		} else {
			sdesc, err := d.d.StructuredDescribeRoll(
				WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage))
			if err != nil {
				return 0, nil, err
			}
			thisResult = append(thisResult, sdesc...)
			reportOptions()
			if d.Confirm {
				results = append(results, StructuredResult{Result: result, Details: thisResult})
				result2, err := d.d.RollToConfirm(true, d.critThreat, d.critBonus)
				if err != nil {
					return 0, nil, err
				}
				if result2 != 0 {
					thisResult = nil
					sdesc, err := d.d.StructuredDescribeRoll(
						WithAutoSF(sfo != "", d.SuccessMessage, d.FailMessage),
						WithRollBonus(d.critBonus))
					if err != nil {
						return 0, nil, err
					}
					thisResult = append(thisResult,
						StructuredDescription{Type: "critlabel", Value: "Confirm:"})
					thisResult = append(thisResult, sdesc...)
					results = append(results, StructuredResult{Result: result2, Details: thisResult})
				}
			} else {
				results = append(results, StructuredResult{Result: result, Details: thisResult})
			}
		}
	}

	return result, results, nil
}

// Roll rolls the dice specified by the specification string, without
// requiring a separate step to create a DieRoller first.
//
// Calling Roll(spec) is equivalent to the sequence
//
//	dr = NewDieRoller()
//	dr.DoRoll(spec)
func Roll(spec string) (string, []StructuredResult, error) {
	d, err := NewDieRoller()
	if err != nil {
		return "", nil, err
	}
	return d.DoRoll(spec)
}

// RollOnce is just like Roll but adds the constraint that there may only be one result
// returned (no confirmation rolls, no repeated rolls, etc., although multiple
// dice such as "3d6+4d10" or "best of N" kinds of things are allowed). It is an error if the
// die roll spec generates multiple results.
//
// The return value differs from Roll in that only a single StructuredResult
// is returned rather than a slice of them.
func RollOnce(spec string) (string, StructuredResult, error) {
	l, r, err := Roll(spec)
	if err != nil {
		return "", StructuredResult{}, err
	}

	if len(r) != 1 {
		return "", StructuredResult{}, fmt.Errorf("die roll spec calls for more than a single roll")
	}
	return l, r[0], nil
}

// DoRollOnce is just like the DoRoll method but adds the constraint that there may only be one result
// returned, in the same way the RollOnce function differs from the Roll() function.
func (d *DieRoller) DoRollOnce(spec string) (string, StructuredResult, error) {
	l, r, err := d.DoRoll(spec)
	if err != nil {
		return "", StructuredResult{}, err
	}

	if len(r) != 1 {
		return "", StructuredResult{}, fmt.Errorf("die roll spec calls for more than a single roll")
	}
	return l, r[0], nil
}

// IsNaturalMax returns true if the DieRoller has been rolled already, and contains but a single die
// in its specification, and that die was rolled
// to the maximum possible value (i.e., a 20 on a d20).
//
// It returns false if that was not true, even if that was for reasons that it
// could not possibly be true (no die was rolled, the die-roll spec contained multiple
// dice, etc.)
func (d *DieRoller) IsNaturalMax() (result bool) {
	return d.isNatural(true)
}

// IsNatural1 returns true if the DieRoller has been rolled already, and contains but a single die
// in its specification, and that die was rolled
// as a natural 1.
//
// It returns false if that was not true, even if that was for reasons that it
// could not possibly be true (no die was rolled, the die-roll spec contained multiple
// dice, etc.)
func (d *DieRoller) IsNatural1() (result bool) {
	return d.isNatural(false)
}

func (d *DieRoller) isNatural(checkForMax bool) (result bool) {
	if !d.d.Rolled {
		return
	}

	for _, die := range d.d.multiDice {
		naturalRoll, sides := die.naturalRoll()
		if sides > 0 {
			if result {
				// too many dice!
				result = false
				return
			}
			if checkForMax {
				if naturalRoll == sides {
					result = true
				}
			} else if naturalRoll == 1 {
				result = true
			}
		}
	}
	return
}

// Text produces a simple plain-text rendering of the information
// in a StructuredDescription slice. This is useful for showing
// the full details of a die roll to a human, or a log file, etc.,
// if fancier formatting isn't needed.
func (sr StructuredDescriptionSet) Text() (string, error) {
	var t strings.Builder
	var i int
	var err error

	for _, r := range sr {
		switch r.Type {
		case "best":
			fmt.Fprintf(&t, " (best of %s) ", r.Value)

		case "bonus", "constant", "diebonus":
			if i, err = strconv.Atoi(r.Value); err != nil {
				return "", err
			}
			if i < 0 {
				fmt.Fprintf(&t, "(%d)", i)
			} else {
				fmt.Fprintf(&t, "%d", i)
			}

		case "critlabel", "critspec", "fullmax", "moddelim", "separator":
			fmt.Fprintf(&t, "%s ", r.Value)

		case "dc":
			fmt.Fprintf(&t, "DC %s ", r.Value)

		case "diespec", "maximized", "operator":
			fmt.Fprintf(&t, "%s", r.Value)

		case "discarded":
			fmt.Fprintf(&t, "{discarded %s}", r.Value)

		case "exceeded":
			fmt.Fprintf(&t, "(EXCEEDED DC by %s) ", r.Value)

		case "fail", "success":
			fmt.Fprintf(&t, "(%s) ", r.Value)

		case "iteration":
			fmt.Fprintf(&t, "(#%s) ", r.Value)

		case "label":
			fmt.Fprintf(&t, " %s", r.Value)

		case "max":
			fmt.Fprintf(&t, "!%s", r.Value)

		case "maxroll":
			fmt.Fprintf(&t, "{!%s}", r.Value)

		case "met":
			fmt.Fprintf(&t, "MET DC (%s) ", r.Value)

		case "min":
			fmt.Fprintf(&t, " (min %s) ", r.Value)

		case "notice":
			fmt.Fprintf(&t, "[%s] ", r.Value)

		case "repeat":
			fmt.Fprintf(&t, "(x%s) ", r.Value)

		case "result":
			fmt.Fprintf(&t, "[%s] ", r.Value)

		case "roll":
			fmt.Fprintf(&t, "{%s}", r.Value)

		case "short":
			fmt.Fprintf(&t, "(MISSED DC by %s) ", r.Value)

		case "subtotal":
			fmt.Fprintf(&t, "(%s)", r.Value)

		case "until":
			fmt.Fprintf(&t, " (until %s) ", r.Value)

		case "worst":
			fmt.Fprintf(&t, " (worst of %s) ", r.Value)

		default:
			fmt.Fprintf(&t, " <%s: %s> ", r.Type, r.Value)
		}
	}
	return t.String(), nil
}

//  ____  ____  _____ ____  _____ _____ ____
// |  _ \|  _ \| ____/ ___|| ____|_   _/ ___|
// | |_) | |_) |  _| \___ \|  _|   | | \___ \
// |  __/|  _ <| |___ ___) | |___  | |  ___) |
// |_|   |_| \_\_____|____/|_____| |_| |____/
//

// DieRollPreset describes each die-roll specification the user
// has stored on the server or in a file as a ready-to-go preset value which will
// be used often, and needs to be persistent across gaming sessions.
type DieRollPreset struct {
	// The name by which this die-roll preset is identified to the user.
	// This must be unique among that user's presets.
	//
	// Clients typically
	// sort these names before displaying them.
	// Note that if a vertical bar ("|") appears in the name, all text
	// up to and including the bar are suppressed from display. This allows
	// for the displayed names to be forced into a particular order on-screen,
	// and allow a set of presets to appear to have the same name from the user's
	// point of view.
	Name string

	// A text description of the purpose for this die-roll specification.
	Description string `json:",omitempty"`

	// The die-roll specification to send to the server. This must be in a
	// form acceptable to the dice.Roll function.
	DieRollSpec string
}

type DieRollPresetMetaData struct {
	Timestamp   int64  `json:",omitempty"`
	DateTime    string `json:",omitempty"`
	Comment     string `json:",omitempty"`
	FileVersion uint   `json:"-"`
}

// WriteDieRollPresetFile writes a slice of presets to the named file.
func WriteDieRollPresetFile(path string, presets []DieRollPreset, meta DieRollPresetMetaData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: WriteDieRollPresetFile was unable to close the output file: %v\n", err)
		}
	}()

	return SaveDieRollPresetFile(file, presets, meta)
}

// SaveDieRollPresetFile writes a slice of presets to an open stream.
func SaveDieRollPresetFile(output io.Writer, presets []DieRollPreset, meta DieRollPresetMetaData) error {
	writer := bufio.NewWriter(output)
	writer.WriteString("__DICE__:2\n")
	if meta.Timestamp == 0 {
		now := time.Now()
		meta.Timestamp = now.Unix()
		meta.DateTime = now.String()
	}
	data, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return err
	}
	writer.WriteString("«__META__» ")
	writer.WriteString(string(data))
	writer.WriteString("\n")

	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Name < presets[j].Name
	})

	for _, preset := range presets {
		data, err := json.MarshalIndent(preset, "", "    ")
		if err != nil {
			return fmt.Errorf("unable to serialize preset \"%s\": %v", preset.Name, err)
		}

		writer.WriteString("«PRESET» ")
		writer.WriteString(string(data))
		writer.WriteString("\n")
	}
	writer.WriteString("«__EOF__»\n")
	writer.Flush()
	return nil
}

// ReadDieRollPresetFile reads in and returns a slice of die-roll presets from
// the named file.
func ReadDieRollPresetFile(path string) ([]DieRollPreset, DieRollPresetMetaData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, DieRollPresetMetaData{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: ReadDieRollPresetFile was unable to close the file: %v", err)
		}
	}()
	return LoadDieRollPresetFile(file)
}

func loadLegacyDieRollPresetFile(scanner *bufio.Scanner, meta DieRollPresetMetaData, legacyMeta string) ([]DieRollPreset, DieRollPresetMetaData, error) {
	//
	// The older format had a first line of:
	//   __DICE__:1 [<timestamp> <date-string>]
	// Followed by a series of lines of:
	//   <name> <desc> <spec>
	//  Each line is a properly-formatted TCL list.
	//
	var presets []DieRollPreset

	metaList, err := tcllist.ParseTclList(legacyMeta)
	if err != nil {
		return nil, meta, fmt.Errorf("legacy die-roll preset file has invalid metadata: %v", err)
	}
	if len(metaList) > 0 {
		meta.Timestamp, _ = strconv.ParseInt(metaList[0], 10, 64)
		if len(metaList) > 1 {
			meta.DateTime = metaList[1]
		}
	}

	for scanner.Scan() {
		f, err := tcllist.ParseTclList(scanner.Text())
		if err != nil {
			return nil, meta, fmt.Errorf("legacy die-roll preset file has invalid record: %v", err)
		}
		if len(f) != 3 {
			return nil, meta, fmt.Errorf("legacy die-roll preset file has invalid record: field count %d", len(f))
		}
		presets = append(presets, DieRollPreset{
			Name:        f[0],
			Description: f[1],
			DieRollSpec: f[2],
		})
	}

	return presets, meta, nil
}

// LoadDieRollPresetFile reads in and returns a slice of die-roll presets from
// an open stream.
func LoadDieRollPresetFile(input io.Reader) ([]DieRollPreset, DieRollPresetMetaData, error) {
	var meta DieRollPresetMetaData
	var presets []DieRollPreset
	var err error
	var f []string
	var v uint64

	if input == nil {
		return nil, meta, nil
	}

	startPattern := regexp.MustCompile("^__DICE__:(\\d+)\\s*(.*)$")
	recordPattern := regexp.MustCompile("^«(PRESET|__META__)»\\s(.+)$")
	eofPattern := regexp.MustCompile("^«__EOF__»$")
	scanner := bufio.NewScanner(input)

	if !scanner.Scan() {
		return nil, meta, nil
	}

	if f = startPattern.FindStringSubmatch(scanner.Text()); f == nil {
		return nil, meta, fmt.Errorf("invalid die-roll preset file format in initial header")
	}
	if v, err = strconv.ParseUint(f[1], 10, 64); err != nil {
		return nil, meta, fmt.Errorf("invalid die-roll preset file format: can't parse version \"%v\": %v", f[1], err)
	}
	meta.FileVersion = uint(v)
	if v < MinimumSupportedDieRollPresetFileFormat || v > MaximumSupportedDieRollPresetFileFormat {
		if MinimumSupportedDieRollPresetFileFormat == MaximumSupportedDieRollPresetFileFormat {
			return nil, meta, fmt.Errorf("cannot read die-roll preset file format version %d (only version %d is supported)", v, MinimumSupportedDieRollPresetFileFormat)
		}
		return nil, meta, fmt.Errorf("cannot read die-roll preset file format version %d (only versions %d-%d are supported)", v, MinimumSupportedDieRollPresetFileFormat, MaximumSupportedDieRollPresetFileFormat)
	}
	if v < 2 {
		return loadLegacyDieRollPresetFile(scanner, meta, f[2])
	}

	for scanner.Scan() {
	rescan:
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		if eofPattern.MatchString(scanner.Text()) {
			return presets, meta, nil
		}
		if f = recordPattern.FindStringSubmatch(scanner.Text()); f == nil {
			return nil, meta, fmt.Errorf("invalid die-roll preset file format: unexpected data \"%v\"", scanner.Text())
		}

		// Start of record type f[1] with start of JSON string f[2]
		// collect more lines of JSON data...
		var dataPacket strings.Builder
		dataPacket.WriteString(f[2])

		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "«") {
				var err error

				switch f[1] {
				case "__META__":
					err = json.Unmarshal([]byte(dataPacket.String()), &meta)

				case "PRESET":
					var preset DieRollPreset
					if err = json.Unmarshal([]byte(dataPacket.String()), &preset); err == nil {
						presets = append(presets, preset)
					}

				default:
					return nil, meta, fmt.Errorf("invalid die-roll preset file: unexpected record type \"%s\"", f[1])
				}
				if err != nil {
					return nil, meta, fmt.Errorf("invalid die-roll preset file: %v", err)
				}
				goto rescan
			}
			dataPacket.WriteString(scanner.Text())
		}
	}
	return nil, meta, fmt.Errorf("invalid die-roll preset file format: unexpected end of file")
}

/*
NewDieRoller -> dr.d Dice

dr.DoRoll(spec) -> d.setNewSpecification(spec)
							separate out title and trailing options
							find and record permutations if any, set dr.Template = spec but change {...} with {0}, {1}, {2}, etc.
							n% rolls	-> dr.d = d100; dr.PctChance=n
							no permutations -> dr.d = new ByDescription(spec)
					expand permutations
					for each permutation
						dr.d = New with permutation spec
						dr.rollDice()
						null out dr.d
					(no permutations)
						dr.rollDice()

dr.rollDice() -> dr.d.MaxRoll() or dr.d.Roll(), dr.d.RollToConfirm()
	breaks into pieces by option
	breaks into pieces by operator

	multiDice = [constant 0] (op+value) (op+value)...

d representation:
	look for c(\d+)?([-+]\d+)?   ????

*/

// @[00]@| Go-GMA 5.20.1
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2024 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
// @[13]@| points along that historical time line.
// @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[15]@| License as described in the accompanying LICENSE file distributed
// @[16]@| with GMA.
// @[17]@|
// @[20]@| Redistribution and use in source and binary forms, with or without
// @[21]@| modification, are permitted provided that the following conditions
// @[22]@| are met:
// @[23]@| 1. Redistributions of source code must retain the above copyright
// @[24]@|    notice, this list of conditions and the following disclaimer.
// @[25]@| 2. Redistributions in binary form must reproduce the above copy-
// @[26]@|    right notice, this list of conditions and the following dis-
// @[27]@|    claimer in the documentation and/or other materials provided
// @[28]@|    with the distribution.
// @[29]@| 3. Neither the name of the copyright holder nor the names of its
// @[30]@|    contributors may be used to endorse or promote products derived
// @[31]@|    from this software without specific prior written permission.
// @[32]@|
// @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
// @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
// @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
// @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
// @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// @[45]@| SUCH DAMAGE.
// @[46]@|
// @[50]@| This software is not intended for any use or application in which
// @[51]@| the safety of lives or property would be at risk due to failure or
// @[52]@| defect of the software.
