/*
########################################################################################
#  _______  _______  _______                ___       ______      ______               #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___  \              #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   )  )             #
# | |      | || || || (___) |            / (_) (_       ___) /       /  /              #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (       /  /               #
# | | \_  )| |   | || (   ) | Game           ) (           ) \     /  /                #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _  /  /                 #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_) \_/                  #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the dice rolling code
//
// TODO: bellcurve and probability tests should use heuristics
//       to check that the results are reasonably distributed

package dice

import (
	"log"
	"sort"
	"strings"
	"testing"
)

const verbose = false // set to true if you want debugging output

func _probtester(t *testing.T, description string, f func() (int, error), minval, maxval int) {
	//
	// Test a die roll by making 10,000 rolls of the die and
	// determining the distribution of values, making certain
	// that in 10,000 rolls no roll fell outside the acceptable
	// range as well.
	//
	hist := make(map[int]int)
	maxvalue := 0
	for i := 0; i < 10000; i++ {
		roll, err := f()
		if err != nil {
			t.Fatalf("%s roll returned error %v", description, err)
		}
		if roll < minval || roll > maxval {
			t.Fatalf("%s roll of %d out of acceptable range [%d,%d]!", description, roll, minval, maxval)
		}
		hist[roll]++
		if maxvalue < hist[roll] {
			maxvalue = hist[roll]
		}
	}
	if verbose {
		log.Printf("Histogram of 10,000 %s die rolls:", description)
		keys := make([]int, 0)
		for k, _ := range hist {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			log.Printf("%3d x%5d %s", k, hist[k], strings.Repeat("*", int((60.0*float64(hist[k]))/float64(maxvalue))))
		}
	}
}

func TestDiceProbabilities(t *testing.T) {
	d4, err1 := New(ByDieType(1, 4, 0))
	d6, err2 := New(ByDieType(1, 6, 0))
	d8, err3 := New(ByDieType(1, 8, 0))
	d10, err4 := New(ByDieType(1, 10, 0))
	d12, err5 := New(ByDieType(1, 12, 0))
	d20, err6 := New(ByDieType(1, 20, 0))
	dpct, err7 := New(ByDieType(1, 100, 0))

	if err1 != nil ||
		err2 != nil ||
		err3 != nil ||
		err4 != nil ||
		err5 != nil ||
		err6 != nil ||
		err7 != nil {
		t.Fatalf("Error constructing Dice object")
	}

	_probtester(t, "d4", d4.Roll, 1, 4)
	_probtester(t, "d6", d6.Roll, 1, 6)
	_probtester(t, "d8", d8.Roll, 1, 8)
	_probtester(t, "d10", d10.Roll, 1, 10)
	_probtester(t, "d12", d12.Roll, 1, 12)
	_probtester(t, "d20", d20.Roll, 1, 20)
	_probtester(t, "d%", dpct.Roll, 1, 100)
}

func TestDiceStrProbabilities(t *testing.T) {
	d4, err1 := New(ByDescription("d4"))
	d6, err2 := New(ByDescription("d6"))
	d8, err3 := New(ByDescription("d4"))
	d10, err4 := New(ByDescription("d10"))
	d12, err5 := New(ByDescription("d12"))
	d20, err6 := New(ByDescription("d20"))
	dpct, err7 := New(ByDescription("d%"))

	if err1 != nil ||
		err2 != nil ||
		err3 != nil ||
		err4 != nil ||
		err5 != nil ||
		err6 != nil ||
		err7 != nil {
		t.Fatalf("Error constructing Dice object")
	}

	_probtester(t, "d4 (str)", d4.Roll, 1, 4)
	_probtester(t, "d6 (str)", d6.Roll, 1, 6)
	_probtester(t, "d8 (str)", d8.Roll, 1, 8)
	_probtester(t, "d10 (str)", d10.Roll, 1, 10)
	_probtester(t, "d12 (str)", d12.Roll, 1, 12)
	_probtester(t, "d20 (str)", d20.Roll, 1, 20)
	_probtester(t, "d% (str)", dpct.Roll, 1, 100)
}

func TestDiceCons(t *testing.T) {
	d, err := New(
		ByDieType(1, 2, 3),
		WithDieBonus(4),
		WithDiv(5),
		WithFactor(0))
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if d.Description() != "1/5d2 (+4 per die)+3" {
		t.Fatalf("Description was %s", d.Description())
	}
}

func TestDiceBellcurves(t *testing.T) {
	_3d6, err := New(ByDieType(3, 6, 0))
	if err != nil {
		t.Fatalf("Construct 3d6: %v", err)
	}

	s3d6, err := New(ByDescription("3d6"))
	if err != nil {
		t.Fatalf("Construct s3d6: %v", err)
	}

	_5d10plus3, err := New(ByDieType(5, 10, 3))
	if err != nil {
		t.Fatalf("Construct 5d10+3: %v", err)
	}

	s5d10plus3, err := New(ByDescription("5d10+3"))
	if err != nil {
		t.Fatalf("Construct s5d10+3: %v", err)
	}

	_probtester(t, "3d6", _3d6.Roll, 3, 18)
	_probtester(t, "3d6 (str)", s3d6.Roll, 3, 18)
	_probtester(t, "5d10+3", _5d10plus3.Roll, 8, 53)
	_probtester(t, "5d10+3 (str)", s5d10plus3.Roll, 8, 53)
}

func compareResults(a, b []StructuredResult) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Result != b[i].Result {
			return false
		}
		if len(a[i].Details) != len(b[i].Details) {
			return false
		}
		for j := range a[i].Details {
			if a[i].Details[j] != b[i].Details[j] {
				return false
			}
		}
	}
	return true
}

func TestDiceHistories(t *testing.T) {
	d, err := New(ByDieType(50, 10, 0))
	if err != nil {
		t.Fatalf("Error creating new Dice: %v", err)
	}

	d.Roll()
	s := 0
	for i, die := range d.multiDice {
		s += die.lastValue()
		if die.getOperator() != "" && die.getOperator() != "+" {
			t.Fatalf("Die #%d operator is %s", i, die.getOperator())
		}
	}
	if s != d.LastValue {
		t.Fatalf("Sum %d, expected %d", s, d.LastValue)
	}
}
func TestDicePercentile(t *testing.T) {
	//	rand.Seed(12345) // static seed so our results will be the same every run
	d, err := NewDieRoller(WithSeed(12345))
	if err != nil {
		t.Fatalf("Error creating new DieRoller: %v", err)
	}

	type testcase struct {
		Roll    string
		Reslist []StructuredResult
	}

	testcases := []testcase{
		// 0
		{Roll: "0%", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "fail"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "0%"},
				{Type: "roll", Value: "84"},
			}},
		}},
		// 1
		{Roll: "100%", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "success"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "100%"},
				{Type: "roll", Value: "44"},
			}},
		}},
		// 2
		{Roll: "52%|!", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "fail"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "maxroll", Value: "100"},
				{Type: "moddelim", Value: "|"},
				{Type: "fullmax", Value: "maximized"},
			}},
		}},
		// 3
		{Roll: "52% blorfl|!", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "did not blorfl"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "blorfl"},
				{Type: "maxroll", Value: "100"},
				{Type: "moddelim", Value: "|"},
				{Type: "fullmax", Value: "maximized"},
			}},
		}},
		// 4
		{Roll: "200% blorfl|!", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "blorfl"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "200%"},
				{Type: "label", Value: "blorfl"},
				{Type: "maxroll", Value: "100"},
				{Type: "moddelim", Value: "|"},
				{Type: "fullmax", Value: "maximized"},
			}},
		}},
		// 5
		{Roll: "200%|!", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "success"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "200%"},
				{Type: "maxroll", Value: "100"},
				{Type: "moddelim", Value: "|"},
				{Type: "fullmax", Value: "maximized"},
			}},
		}},
		// 6
		{Roll: "52% miss", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "hit"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "miss"},
				{Type: "roll", Value: "85"},
			}},
		}},
		// 7
		{Roll: "52% miss", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "miss"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "miss"},
				{Type: "roll", Value: "37"},
			}},
		}},
		// 8
		{Roll: "20% miss", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "hit"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "20%"},
				{Type: "label", Value: "miss"},
				{Type: "roll", Value: "42"},
			}},
		}},
		// 9
		{Roll: "30% blorfl", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "did not blorfl"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "30%"},
				{Type: "label", Value: "blorfl"},
				{Type: "roll", Value: "76"},
			}},
		}},
		// 10
		{Roll: "80% blorfl", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "blorfl"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "80%"},
				{Type: "label", Value: "blorfl"},
				{Type: "roll", Value: "43"},
			}},
		}},
		// 11
		{Roll: "52% xxx/yyy", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "xxx"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "xxx/yyy"},
				{Type: "roll", Value: "7"},
			}},
		}},
		// 12
		{Roll: "52% xxx/yyy", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "success", Value: "xxx"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "xxx/yyy"},
				{Type: "roll", Value: "2"},
			}},
		}},
		// 13
		{Roll: "52% xxx/yyy", Reslist: []StructuredResult{
			{Result: 0, Details: []StructuredDescription{
				{Type: "fail", Value: "yyy"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "52%"},
				{Type: "label", Value: "xxx/yyy"},
				{Type: "roll", Value: "59"},
			}},
		}},
	}

	for i, test := range testcases {
		label, results, err := d.DoRoll(test.Roll)
		if err != nil {
			t.Fatalf("test #%d error %v", i, err)
		}
		if !compareResults(results, test.Reslist) {
			t.Fatalf("test #%d result %v, expected %v", i, results, test.Reslist)
		}
		if label != "" {
			t.Fatalf("test #%d label was %v, expected it to be empty", i, label)
		}
	}
}

func TestDiceStructured(t *testing.T) {
	//rand.Seed(12345) // static seed so our results will be the same every run
	d, err := NewDieRoller(WithSeed(12345))
	if err != nil {
		t.Fatalf("Error creating new DieRoller: %v", err)
	}

	type testcase struct {
		Roll    string
		Reslist []StructuredResult
	}

	testcases := []testcase{
		// 0
		{Roll: "d20", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "4"},
			}},
		}},
		// 1
		{Roll: "", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "4"},
			}},
		}},
		// 2
		{Roll: "", Reslist: []StructuredResult{
			{Result: 5, Details: []StructuredDescription{
				{Type: "result", Value: "5"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "5"},
			}},
		}},
		// 3
		{Roll: "", Reslist: []StructuredResult{
			{Result: 17, Details: []StructuredDescription{
				{Type: "result", Value: "17"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "17"},
			}},
		}},
		// 4
		{Roll: "d20+12 | min 3 | c19+2", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
		}},
		// 5
		{Roll: "d20+12 | min 3 | c19+2", Reslist: []StructuredResult{
			{Result: 28, Details: []StructuredDescription{
				{Type: "result", Value: "28"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "16"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
		}},
		// 6
		{Roll: "", Reslist: []StructuredResult{
			{Result: 15, Details: []StructuredDescription{
				{Type: "result", Value: "15"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "3"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
		}},
		// 7
		{Roll: "", Reslist: []StructuredResult{
			{Result: 19, Details: []StructuredDescription{
				{Type: "result", Value: "19"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "7"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
		}},
		// 8
		{Roll: "", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
		}},
		// 9
		{Roll: "", Reslist: []StructuredResult{
			{Result: 31, Details: []StructuredDescription{
				{Type: "result", Value: "31"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "19"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c19+2"},
			}},
			{Result: 19, Details: []StructuredDescription{
				{Type: "critlabel", Value: "Confirm:"},
				{Type: "result", Value: "19"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "5"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "12"},
				{Type: "bonus", Value: "+2"},
				{Type: "moddelim", Value: "|"},
				{Type: "min", Value: "3"},
			}},
		}},
		// 10
		{Roll: "2d10|until 18", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,6"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "1"},
				{Type: "short", Value: "4"},
			}},
			{Result: 10, Details: []StructuredDescription{
				{Type: "result", Value: "10"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "3,7"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "2"},
				{Type: "short", Value: "8"},
			}},
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "9,4"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "3"},
				{Type: "short", Value: "5"},
			}},
			{Result: 16, Details: []StructuredDescription{
				{Type: "result", Value: "16"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,8"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "4"},
				{Type: "short", Value: "2"},
			}},
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "6,1"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "5"},
				{Type: "short", Value: "11"},
			}},
			{Result: 17, Details: []StructuredDescription{
				{Type: "result", Value: "17"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "10,7"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "6"},
				{Type: "short", Value: "1"},
			}},
			{Result: 11, Details: []StructuredDescription{
				{Type: "result", Value: "11"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "6,5"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "7"},
				{Type: "short", Value: "7"},
			}},
			{Result: 8, Details: []StructuredDescription{
				{Type: "result", Value: "8"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "7,1"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "8"},
				{Type: "short", Value: "10"},
			}},
			{Result: 10, Details: []StructuredDescription{
				{Type: "result", Value: "10"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "7,3"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "9"},
				{Type: "short", Value: "8"},
			}},
			{Result: 10, Details: []StructuredDescription{
				{Type: "result", Value: "10"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "7,3"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "10"},
				{Type: "short", Value: "8"},
			}},
			{Result: 11, Details: []StructuredDescription{
				{Type: "result", Value: "11"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,3"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "11"},
				{Type: "short", Value: "7"},
			}},
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "6,1"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "12"},
				{Type: "short", Value: "11"},
			}},
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,5"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "13"},
				{Type: "short", Value: "5"},
			}},
			{Result: 15, Details: []StructuredDescription{
				{Type: "result", Value: "15"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,7"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "14"},
				{Type: "short", Value: "3"},
			}},
			{Result: 9, Details: []StructuredDescription{
				{Type: "result", Value: "9"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "6,3"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "15"},
				{Type: "short", Value: "9"},
			}},
			{Result: 5, Details: []StructuredDescription{
				{Type: "result", Value: "5"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "1,4"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "16"},
				{Type: "short", Value: "13"},
			}},
			{Result: 10, Details: []StructuredDescription{
				{Type: "result", Value: "10"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "3,7"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "17"},
				{Type: "short", Value: "8"},
			}},
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "4,3"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "18"},
				{Type: "short", Value: "11"},
			}},
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "8,6"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "19"},
				{Type: "short", Value: "4"},
			}},
			{Result: 12, Details: []StructuredDescription{
				{Type: "result", Value: "12"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "3,9"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "20"},
				{Type: "short", Value: "6"},
			}},
			{Result: 11, Details: []StructuredDescription{
				{Type: "result", Value: "11"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "3,8"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "21"},
				{Type: "short", Value: "7"},
			}},
			{Result: 6, Details: []StructuredDescription{
				{Type: "result", Value: "6"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "4,2"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "22"},
				{Type: "short", Value: "12"},
			}},
			{Result: 6, Details: []StructuredDescription{
				{Type: "result", Value: "6"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "1,5"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "23"},
				{Type: "short", Value: "12"},
			}},
			{Result: 11, Details: []StructuredDescription{
				{Type: "result", Value: "11"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "6,5"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "24"},
				{Type: "short", Value: "7"},
			}},
			{Result: 19, Details: []StructuredDescription{
				{Type: "result", Value: "19"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d10"},
				{Type: "roll", Value: "9,10"},
				{Type: "moddelim", Value: "|"},
				{Type: "until", Value: "18"},
				{Type: "iteration", Value: "25"},
				{Type: "exceeded", Value: "1"},
			}},
		}},
		// 11
		{Roll: "1d6+2|repeat 3", Reslist: []StructuredResult{
			{Result: 5, Details: []StructuredDescription{
				{Type: "result", Value: "5"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d6"},
				{Type: "roll", Value: "3"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "repeat", Value: "3"},
				{Type: "iteration", Value: "1"},
			}},
			{Result: 3, Details: []StructuredDescription{
				{Type: "result", Value: "3"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d6"},
				{Type: "roll", Value: "1"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "repeat", Value: "3"},
				{Type: "iteration", Value: "2"},
			}},
			{Result: 3, Details: []StructuredDescription{
				{Type: "result", Value: "3"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d6"},
				{Type: "roll", Value: "1"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "repeat", Value: "3"},
				{Type: "iteration", Value: "3"},
			}},
		}},
		// 12
		{Roll: "3d6|maximized", Reslist: []StructuredResult{
			{Result: 18, Details: []StructuredDescription{
				{Type: "result", Value: "18"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "3d6"},
				{Type: "maxroll", Value: "6,6,6"},
				{Type: "moddelim", Value: "|"},
				{Type: "fullmax", Value: "maximized"},
			}},
		}},
		// 13
		{Roll: "2d8bestof2", Reslist: []StructuredResult{
			{Result: 9, Details: []StructuredDescription{
				{Type: "result", Value: "9"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d8"},
				{Type: "best", Value: "2"},
				{Type: "discarded", Value: "3,5"},
				{Type: "roll", Value: "1,8"},
			}},
		}},
		// 14
		{Roll: "2d8 best of 3", Reslist: []StructuredResult{
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d8"},
				{Type: "best", Value: "3"},
				{Type: "discarded", Value: "6,5"},
				{Type: "discarded", Value: "7,5"},
				{Type: "roll", Value: "7,6"},
			}},
		}},
		// 15
		{Roll: "2d6 worst of 3", Reslist: []StructuredResult{
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "2d6"},
				{Type: "worst", Value: "3"},
				{Type: "roll", Value: "6,1"},
				{Type: "discarded", Value: "4,3"},
				{Type: "discarded", Value: "4,3"},
			}},
		}},
		// 16
		{Roll: "d20+3|dc16", Reslist: []StructuredResult{
			{Result: 21, Details: []StructuredDescription{
				{Type: "result", Value: "21"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "18"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "dc", Value: "16"},
				{Type: "exceeded", Value: "5"},
			}},
		}},
		// 17
		{Roll: "d20+3|dc16", Reslist: []StructuredResult{
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "10"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "dc", Value: "16"},
				{Type: "short", Value: "3"},
			}},
		}},
		// 18
		{Roll: "d20+3|dc21", Reslist: []StructuredResult{
			{Result: 21, Details: []StructuredDescription{
				{Type: "result", Value: "21"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "18"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "dc", Value: "21"},
				{Type: "met", Value: "successful"},
			}},
		}},
		// 19
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 6, Details: []StructuredDescription{
				{Type: "result", Value: "6"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "4"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 20
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 11, Details: []StructuredDescription{
				{Type: "result", Value: "11"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "9"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 21
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 5, Details: []StructuredDescription{
				{Type: "result", Value: "5"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "3"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 22
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 19, Details: []StructuredDescription{
				{Type: "result", Value: "19"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "17"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 23
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 16, Details: []StructuredDescription{
				{Type: "result", Value: "16"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "14"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 24
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 15, Details: []StructuredDescription{
				{Type: "result", Value: "15"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "13"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 25
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 8, Details: []StructuredDescription{
				{Type: "result", Value: "8"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "6"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 26
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 19, Details: []StructuredDescription{
				{Type: "result", Value: "19"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "17"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 27
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 6, Details: []StructuredDescription{
				{Type: "result", Value: "6"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "4"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 28
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 15, Details: []StructuredDescription{
				{Type: "result", Value: "15"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "13"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 29
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 17, Details: []StructuredDescription{
				{Type: "result", Value: "17"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "15"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 30
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 31
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 6, Details: []StructuredDescription{
				{Type: "result", Value: "6"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "4"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 32
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "5"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 33
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 34
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 20, Details: []StructuredDescription{
				{Type: "result", Value: "20"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "18"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 35
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "12"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 36
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 21, Details: []StructuredDescription{
				{Type: "result", Value: "21"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "19"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 37
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 8, Details: []StructuredDescription{
				{Type: "result", Value: "8"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "6"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 38
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "11"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 39
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 5, Details: []StructuredDescription{
				{Type: "result", Value: "5"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "3"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 40
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 21, Details: []StructuredDescription{
				{Type: "result", Value: "21"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "19"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 41
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 3, Details: []StructuredDescription{
				{Type: "fail", Value: "MISS"},
				{Type: "result", Value: "3"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "1"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 42
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 10, Details: []StructuredDescription{
				{Type: "result", Value: "10"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "8"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 43
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 9, Details: []StructuredDescription{
				{Type: "result", Value: "9"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "7"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 44
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 9, Details: []StructuredDescription{
				{Type: "result", Value: "9"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "7"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 45
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 16, Details: []StructuredDescription{
				{Type: "result", Value: "16"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "14"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 46
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 7, Details: []StructuredDescription{
				{Type: "result", Value: "7"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "5"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 47
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "2"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 48
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 8, Details: []StructuredDescription{
				{Type: "result", Value: "8"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "6"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 49
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "12"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 50
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "12"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 51
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "11"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 52
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 14, Details: []StructuredDescription{
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "12"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 53
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 12, Details: []StructuredDescription{
				{Type: "result", Value: "12"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "10"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 54
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 13, Details: []StructuredDescription{
				{Type: "result", Value: "13"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "11"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 55
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 16, Details: []StructuredDescription{
				{Type: "result", Value: "16"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "14"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
		}},
		// 56
		{Roll: "d20+2|c", Reslist: []StructuredResult{
			{Result: 22, Details: []StructuredDescription{
				{Type: "success", Value: "HIT"},
				{Type: "result", Value: "22"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "20"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "critspec", Value: "c"},
			}},
			{Result: 14, Details: []StructuredDescription{
				{Type: "critlabel", Value: "Confirm:"},
				{Type: "result", Value: "14"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d20"},
				{Type: "roll", Value: "12"},
				{Type: "operator", Value: "+"},
				{Type: "constant", Value: "2"},
			}},
		}},
		// 57
		{Roll: "d10|sf", Reslist: []StructuredResult{
			{Result: 9, Details: []StructuredDescription{
				{Type: "result", Value: "9"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d10"},
				{Type: "roll", Value: "9"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf"},
			}},
		}},
		// 58
		{Roll: "d4|sf", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "success", Value: "SUCCESS"},
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "4"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf"},
			}},
		}},
		// 59
		{Roll: "d4|sf", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "fail", Value: "FAIL"},
				{Type: "result", Value: "1"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "1"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf"},
			}},
		}},
		// 60
		{Roll: "d4|sf", Reslist: []StructuredResult{
			{Result: 2, Details: []StructuredDescription{
				{Type: "result", Value: "2"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf"},
			}},
		}},
		// 61
		{Roll: "d4|sf foo/bar", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "success", Value: "foo"},
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "4"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf foo/bar"},
			}},
		}},
		// 62
		{Roll: "d4|sf foo/bar", Reslist: []StructuredResult{
			{Result: 4, Details: []StructuredDescription{
				{Type: "success", Value: "foo"},
				{Type: "result", Value: "4"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "4"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf foo/bar"},
			}},
		}},
		// 63
		{Roll: "d4|sf foo/bar", Reslist: []StructuredResult{
			{Result: 2, Details: []StructuredDescription{
				{Type: "result", Value: "2"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "2"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf foo/bar"},
			}},
		}},
		// 64
		{Roll: "d4|sf foo/bar", Reslist: []StructuredResult{
			{Result: 3, Details: []StructuredDescription{
				{Type: "result", Value: "3"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "3"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf foo/bar"},
			}},
		}},
		// 65
		{Roll: "d4|sf foo/bar", Reslist: []StructuredResult{
			{Result: 1, Details: []StructuredDescription{
				{Type: "fail", Value: "bar"},
				{Type: "result", Value: "1"},
				{Type: "separator", Value: "="},
				{Type: "diespec", Value: "1d4"},
				{Type: "roll", Value: "1"},
				{Type: "moddelim", Value: "|"},
				{Type: "sf", Value: "sf foo/bar"},
			}},
		}},
		// 66
		/* test removed because the cartesian product routine doesn't return
		   the same ordering of values consistently
			{Roll: "d20{+15/+10/+5}+4|c18", Reslist: []StructuredResult{
				{Result: 27, Details: []StructuredDescription{
					{Type: "result", Value: "27"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "8"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "15"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "4"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 30, Details: []StructuredDescription{
					{Type: "result", Value: "30"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "16"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "10"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "4"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 23, Details: []StructuredDescription{
					{Type: "result", Value: "23"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "14"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "5"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "4"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
			}},
		*/
		// 67
		/* test removed because the cartesian product routine doesn't return
		   the same ordering of values consistently

			{Roll: "d20{+15/+10/+5}+{2/3}", Reslist: []StructuredResult{
				{Result: 24, Details: []StructuredDescription{	// 15 3
					{Type: "result", Value: "24"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "6"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "15"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "3"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 25, Details: []StructuredDescription{	// 5 3
					{Type: "result", Value: "25"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "17"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "5"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "3"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 24, Details: []StructuredDescription{	// 5 2
					{Type: "result", Value: "24"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "17"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "5"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "2"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 31, Details: []StructuredDescription{	// 15 2
					{Type: "result", Value: "31"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "14"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "15"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "2"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 24, Details: []StructuredDescription{	// 10 3
					{Type: "result", Value: "24"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "11"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "10"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "3"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
				{Result: 26, Details: []StructuredDescription{	// 10 2
					{Type: "result", Value: "26"},
					{Type: "separator", Value: "="},
					{Type: "diespec", Value: "1d20"},
					{Type: "roll", Value: "14"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "10"},
					{Type: "operator", Value: "+"},
					{Type: "constant", Value: "2"},
					{Type: "moddelim", Value: "|"},
					{Type: "critspec", Value: "c18"},
				}},
			}},
		*/
	}

	for i, test := range testcases {
		label, results, err := d.DoRoll(test.Roll)
		if err != nil {
			t.Fatalf("test #%d error %v", i, err)
		}
		if !compareResults(results, test.Reslist) {
			t.Fatalf("test #%d result %v, expected %v", i, results, test.Reslist)
		}
		if label != "" {
			t.Fatalf("test #%d label was %v, expected it to be empty", i, label)
		}
	}
}

// @[00]@| GMA 4.3.7
// @[01]@|
// @[10]@| Copyright © 1992–2021 by Steven L. Willoughby
// @[11]@| (AKA Software Alchemy), Aloha, Oregon, USA. All Rights Reserved.
// @[12]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[13]@| License as described in the accompanying LICENSE file distributed
// @[14]@| with GMA.
// @[15]@|
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
//
