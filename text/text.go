/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   _____      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) / ___ \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |( (   ) )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )( (___) |   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /  \____  |   | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) |   |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_)(_______)     #
#                                                                                      #
########################################################################################
*/

//
// Package text provides text processing facilities used by GMA.
//
package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MadScienceZone/go-gma/v5/util"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sum(ints ...int) int {
	s := 0
	for _, v := range ints {
		s += v
	}
	return s
}

//  ____                               _   _                                _
// |  _ \ ___  _ __ ___   __ _ _ __   | \ | |_   _ _ __ ___   ___ _ __ __ _| |___
// | |_) / _ \| '_ ` _ \ / _` | '_ \  |  \| | | | | '_ ` _ \ / _ \ '__/ _` | / __|
// |  _ < (_) | | | | | | (_| | | | | | |\  | |_| | | | | | |  __/ | | (_| | \__ \
// |_| \_\___/|_| |_| |_|\__,_|_| |_| |_| \_|\__,_|_| |_| |_|\___|_|  \__,_|_|___/
//
type romanTableEntry struct {
	i int
	r string
}

func romanTable() []romanTableEntry {
	return []romanTableEntry{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}
}

//
// ToRoman converts an integer value to a Roman numeral string.
// This will return "0" for a zero value.
//
func ToRoman(i int) (string, error) {
	var roman strings.Builder

	if i < 0 {
		return "", fmt.Errorf("cannot represent negative values in Roman numerals")
	}

	if i == 0 {
		return "0", nil
	}

	for _, r := range romanTable() {
		for i >= r.i {
			roman.WriteString(r.r)
			i -= r.i
		}
	}

	return roman.String(), nil
}

//
// FromRoman converts a Roman numeral string to integer.
// Accepts "0" as a zero value.
//
func FromRoman(roman string) (int, error) {
	var v int

	roman = strings.ToUpper(strings.TrimFunc(roman, unicode.IsSpace))
	if roman == "0" {
		return 0, nil
	}

	for roman != "" {
		found := false
		for _, r := range romanTable() {
			if strings.HasPrefix(roman, r.r) {
				v += r.i
				roman = roman[len(r.r):]
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("Not a valid Roman numeral (what is %s?)", roman)
		}
	}

	return v, nil
}

//======================================================================================
//  _____ _______  _______   _____ ___  ____  __  __    _  _____ _____ ___ _   _  ____
// |_   _| ____\ \/ /_   _| |  ___/ _ \|  _ \|  \/  |  / \|_   _|_   _|_ _| \ | |/ ___|
//   | | |  _|  \  /  | |   | |_ | | | | |_) | |\/| | / _ \ | |   | |  | ||  \| | |  _
//   | | | |___ /  \  | |   |  _|| |_| |  _ <| |  | |/ ___ \| |   | |  | || |\  | |_| |
//   |_| |_____/_/\_\ |_|   |_|   \___/|_| \_\_|  |_/_/   \_\_|   |_| |___|_| \_|\____|
//

//
// CenterPrefix returns the string of spaces which will need to go before the
// given string so that, when the string is printed, it will end up centered
// inside a field of the given width.  If the width is insufficient for the
// string to be held inside it, the empty string is returned since there
// will therefore not be any padding that goes to the left of the string.
//
func CenterPrefix(s string, width int) string {
	return strings.Repeat(" ", (width-len(s))/2)
}

//
// CenterSuffix is like CenterPrefix but gives the spaces which follow
// the string to complete the centering operation.
//
func CenterSuffix(s string, width int) string {
	return strings.Repeat(" ", (width-len(s)+1)/2)
}

//
// CenterText returns a padded string with the input string s
// centered within a field of the given width.
//
func CenterText(s string, width int) string {
	l := len(s)
	return strings.Repeat(" ", (width-l)/2) + s + strings.Repeat(" ", (width-l+1)/2)
}

//
// CenterTextPadding takes the number of characters in the string
// and the field width, and returns the prefix and suffix padding strings.
//
func CenterTextPadding(strlen, width int) (string, string) {
	return strings.Repeat(" ", (width-strlen)/2), strings.Repeat(" ", (width-strlen+1)/2)
}

//==================================================================
//  _____ _______  _______   __  __    _    ____  _  ___   _ ____
// |_   _| ____\ \/ /_   _| |  \/  |  / \  |  _ \| |/ / | | |  _ \
//   | | |  _|  \  /  | |   | |\/| | / _ \ | |_) | ' /| | | | |_) |
//   | | | |___ /  \  | |   | |  | |/ ___ \|  _ <| . \| |_| |  __/
//   |_| |_____/_/\_\ |_|   |_|  |_/_/   \_\_| \_\_|\_\\___/|_|
//

//
// Options to the Render function are tracked in
// this structure.
//
type renderOptSet struct {
	formatter renderingFormatter
	bulletSet []rune
	compact   bool
}

// AsPlainText may be added as an option to the Render function
// to select plain text output format.
func AsPlainText(o *renderOptSet) {
	o.formatter = &renderPlainTextFormatter{}
}

// AsHTML may be added as an option to the Render function
// to select HTML output format.
func AsHTML(o *renderOptSet) {
	o.formatter = &renderHTMLFormatter{}
}

// AsPostScript may be added as an option to the Render function
// to select PostScript output format.
func AsPostScript(o *renderOptSet) {
	o.formatter = &renderPostScriptFormatter{}
}

//
// WithBullets may be added as an option to the Render function to
// specify a custom set of bullet characters
// to use for bulleted lists. The bullets passed
// to this option are used in order, then the list
// repeats over as necessary for additional levels.
//
// Example:
//  formattedText, err := Render(srcText, AsPlainText, WithBullets('*', '-'))
// This will alternate between '*' and '-' as bullets at each level.
//
// While the default bullet(s) are chosen appropriately for each output format,
// no other processing is made to the runes passed here; they are used as-is
// in each case, but the following special characters are recognized and
// translated to something sensible in each output format:
//    •  U+2022 Standard bullet
//    ‣  U+2023 Triangle bullet
//    ⁃  U+2043 Hyphen bullet
//    ○  U+25CB Unfilled circle bullet
//    ☞  U+261E Pointing index bullet
//    ★  U+2605 Star bullet
//
func WithBullets(bullets ...rune) func(*renderOptSet) {
	return func(o *renderOptSet) {
		o.bulletSet = bullets
	}
}

//
// WithCompactText may be added as an option to the Render function to
// specify that a more compact rendering of text
// blocks in order to conserve paper real estate.
//
// Currently only supported for PostScript output.
//
// Example:
//  ps, err := Render(srcText, AsPostScript, WithCompactText)
//
func WithCompactText(o *renderOptSet) {
	o.compact = true
}

//
// Each output formatter must supply these methods
// which the Render function will invoke as it parses
// the marked up source text.
//
type renderingFormatter interface {
	init(renderOptSet)
	newPar()
	process(text string)
	finalize() string
	setBold(on bool)
	setItal(on bool)
	newLine()
	table(*textTable)
	reference(displayName, linkName string)
	bulletListItem(level int, bullet rune)
	enumListItem(level, counter int)
	title(text string)
	subtitle(text string)
	toString(string) string
}

//
//  ____  _       _     _____         _
// |  _ \| | __ _(_)_ _|_   _|____  _| |_
// | |_) | |/ _` | | '_ \| |/ _ \ \/ / __|
// |  __/| | (_| | | | | | |  __/>  <| |_
// |_|   |_|\__,_|_|_| |_|_|\___/_/\_\\__|
//
// Plain Text output formatter
//
type renderPlainTextFormatter struct {
	buf    strings.Builder
	indent int
	ital   bool
	bold   bool
}

func (f *renderPlainTextFormatter) title(text string) {
	f.buf.WriteString("\n══╣ ")
	f.process(text)
	f.buf.WriteString(" ╠" + strings.Repeat("═", 80-4-2-miniLen(text, f.toString, oneLine)) + "\n")
}

func (f *renderPlainTextFormatter) subtitle(text string) {
	f.buf.WriteString("\n──┤ ")
	f.process(text)
	f.buf.WriteString(" ├" + strings.Repeat("─", 80-4-2-miniLen(text, f.toString, oneLine)) + "\n")
}

func (f *renderPlainTextFormatter) init(o renderOptSet) {}

func (f *renderPlainTextFormatter) setItal(b bool) {
	f.ital = b
}

func (f *renderPlainTextFormatter) setBold(b bool) {
	f.bold = b
}

func (f *renderPlainTextFormatter) process(text string) {
	f.buf.WriteString(f.toString(text))
}

func (f *renderPlainTextFormatter) finalize() string {
	return f.buf.String()
}

func (f *renderPlainTextFormatter) reference(desc, link string) {
	f.buf.WriteString(f.toString(desc))
}

func (f *renderPlainTextFormatter) newLine() {
	fmt.Fprintf(&f.buf, "\n%*s", f.indent*3, "")
}

func (f *renderPlainTextFormatter) newPar() {
	f.buf.WriteString("\n\n")
	f.indent = 0
}

func (f *renderPlainTextFormatter) bulletListItem(level int, bullet rune) {
	if bullet == 0 {
		bullet = '\u2022'
	}
	fmt.Fprintf(&f.buf, "\n%*s%c  ", (level-1)*3, "", bullet)
	f.indent = level
}

func (f *renderPlainTextFormatter) enumListItem(level, counter int) {
	fmt.Fprintf(&f.buf, "\n%*s%s. ", (level-1)*3, "", enumVal(level, counter))
	f.indent = level
}

func miniMaxLen(s string, filter func(string) string, o ...miniFormatterOption) int {
	cf := &countingFormatter{
		filter: filter,
	}
	miniFormatter(s, cf, o...)
	return cf.maxLineWidth()
}

func miniLen(s string, filter func(string) string, o ...miniFormatterOption) int {
	cf := &countingFormatter{
		filter: filter,
	}
	miniFormatter(s, cf, o...)
	return cf.textWidth()
}

func (f *renderPlainTextFormatter) table(t *textTable) {
	//
	// First pass: add up the widths of the non-spanning columns
	//
	colsize := make([]int, 0, 5)
	for _, row := range t.rows {
		for i, col := range row {
			for len(colsize) <= i {
				colsize = append(colsize, 0)
			}
			if col != nil && col.span == 0 {
				colsize[i] = max(colsize[i], miniLen(col.text, f.toString, oneLine))
			}
		}
	}
	//
	// Now that we know how much room the normal column text
	// requires, adjust as needed for spanning text.
	//
	for _, row := range t.rows {
		for i, col := range row {
			if col != nil && col.span > 0 {
				alreadyAllocated := sum(colsize[i : i+col.span+1]...)
				spaceNeeded := miniLen(col.text, f.toString, oneLine) - 3*col.span
				if spaceNeeded > alreadyAllocated {
					add := spaceNeeded - alreadyAllocated
					each := add / (col.span + 1)
					for ci := i; ci <= i+col.span; ci++ {
						colsize[ci] += each
						add -= each
					}
					if add > 0 {
						colsize[i] += add
					}
				}
			}
		}
	}
	//
	// Now lay out the table data in these columns
	//
	f.buf.WriteRune('\n')
	if len(t.captions) > 0 {
		miniFormatter(strings.ToUpper(strings.Join(t.captions, " ")), f)
		f.buf.WriteRune('\n')
	}
	for i, c := range colsize {
		if i == 0 {
			f.buf.WriteRune('┌')
		} else {
			f.buf.WriteRune('┬')
		}
		for xx := 0; xx < c+2; xx++ {
			f.buf.WriteRune('─')
		}
	}
	f.buf.WriteString("┐\n")
	for _, row := range t.rows {
		headerRow := false
		for c := 0; c < len(row); c++ {
			if row[c] != nil {
				colwidth := sum(colsize[c:c+row[c].span+1]...) + 3*row[c].span
				pre, post := CenterTextPadding(miniLen(row[c].text, f.toString, oneLine), colwidth)
				if row[c].header {
					f.process("│ " + pre)
					miniFormatter(strings.ToUpper(row[c].text), f, oneLine)
					f.process(post + " ")
					headerRow = true
				} else {
					switch row[c].align {
					case '>':
						f.process("│ " + pre + post)
						miniFormatter(row[c].text, f, oneLine)
						f.process(" ")
					case '^':
						f.process("│ " + pre)
						miniFormatter(row[c].text, f, oneLine)
						f.process(post + " ")
					default:
						f.process("│ ")
						miniFormatter(row[c].text, f, oneLine)
						f.process(pre + post + " ")
					}
				}
			}
		}
		f.buf.WriteString("│\n")
		if headerRow {
			for i, c := range colsize {
				if i == 0 {
					f.buf.WriteRune('├')
				} else {
					f.buf.WriteRune('┼')
				}
				for xx := 0; xx < c+2; xx++ {
					f.buf.WriteRune('─')
				}
			}
			f.buf.WriteString("┤\n")
		}
	}
	for i, c := range colsize {
		if i == 0 {
			f.buf.WriteRune('└')
		} else {
			f.buf.WriteRune('┴')
		}
		for xx := 0; xx < c+2; xx++ {
			f.buf.WriteRune('─')
		}
	}
	f.buf.WriteString("┘\n")
	for _, footer := range t.footnotes {
		miniFormatter(footer, f)
		f.buf.WriteString("\n")
	}
}

//
//  _   _ _____ __  __ _
// | | | |_   _|  \/  | |
// | |_| | | | | |\/| | |
// |  _  | | | | |  | | |___
// |_| |_| |_| |_|  |_|_____|
//
// HTML output formatter
//
type renderHTMLFormatter struct {
	buf       strings.Builder
	indent    int
	ital      bool
	bold      bool
	listStack []string
}

type replSet struct {
	str  string
	re   *regexp.Regexp
	repl string
}

func (f *renderPlainTextFormatter) toString(s string) string {
	for _, sub := range []replSet{
		{str: "+/-", repl: "±"},
		{str: "---", repl: "\007"},
		{str: "--", repl: "-"},
		{str: "\007", repl: "--"},
		{re: regexp.MustCompile(`(\d)x`), repl: "${1}×"},
		{re: regexp.MustCompile(`x(\d)`), repl: "×${1}"},
		{str: "[x]", repl: "×"},
		{str: "[S]", repl: "§"},
		{str: "[0]", repl: "⁰"},
		{str: "[1]", repl: "¹"},
		{str: "[2]", repl: "²"},
		{str: "[3]", repl: "³"},
		{str: "[4]", repl: "⁴"},
		{str: "[5]", repl: "⁵"},
		{str: "[6]", repl: "⁶"},
		{str: "[7]", repl: "⁷"},
		{str: "[8]", repl: "⁸"},
		{str: "[9]", repl: "⁹"},
		{re: regexp.MustCompile(`\b1/2\b`), repl: "½"},
		{re: regexp.MustCompile(`\b1/4\b`), repl: "¼"},
		{re: regexp.MustCompile(`\b3/4\b`), repl: "¾"},
		{re: regexp.MustCompile(`\b(\d+)_1/2\b`), repl: "${1}½"},
		{re: regexp.MustCompile(`\b(\d+)_1/4\b`), repl: "${1}¼"},
		{re: regexp.MustCompile(`\b(\d+)_3/4\b`), repl: "${1}¾"},
		{str: "^o", repl: "°"},
		{str: "[c]", repl: "©"},
		{str: "[R]", repl: "®"},
		{str: "AE", repl: "Æ"},
		{str: "ae", repl: "æ"},
		{str: "[<<]", repl: "«"},
		{str: "[>>]", repl: "»"},
		{str: "^.", repl: "·"},
		{str: "[/]", repl: "÷"},
		{str: "[+]", repl: "†"},
		{str: "[++]", repl: "‡"},
	} {
		if sub.re != nil {
			s = sub.re.ReplaceAllString(s, sub.repl)
		} else {
			s = strings.ReplaceAll(s, sub.str, sub.repl)
		}
	}
	return s
}

func (f *renderHTMLFormatter) title(text string) {
	f.buf.WriteString("<H1>")
	f.process(text)
	f.buf.WriteString("</H1>")
}

func (f *renderHTMLFormatter) subtitle(text string) {
	f.buf.WriteString("<H2>")
	f.process(text)
	f.buf.WriteString("</H2>")
}

func (f *renderHTMLFormatter) toString(s string) string {
	for _, sub := range []replSet{
		{str: "&", repl: "&amp;"},
		{str: "+/-", repl: "&plusmn;"},
		{str: "---", repl: "&mdash;"},
		{str: "--", repl: "&ndash;"},
		{re: regexp.MustCompile(`(^|\s|\d)-(\d+)`), repl: "${1}&minus;${2}"},
		{re: regexp.MustCompile(`(\d)x`), repl: "${1}×"},
		{re: regexp.MustCompile(`x(\d)`), repl: "×${1}"},
		{str: "[x]", repl: "&times;"},
		{str: "[S]", repl: "&sect;"},
		{str: "[<<]", repl: "&laquo;"},
		{str: "[>>]", repl: "&raquo;"},
		{str: "<", repl: "&lt;"},
		{str: ">", repl: "&gt;"},
		{str: "[0]", repl: "<sup>0</sup>"},
		{str: "[1]", repl: "<sup>1</sup>"},
		{str: "[2]", repl: "<sup>2</sup>"},
		{str: "[3]", repl: "<sup>3</sup>"},
		{str: "[4]", repl: "<sup>4</sup>"},
		{str: "[5]", repl: "<sup>5</sup>"},
		{str: "[6]", repl: "<sup>6</sup>"},
		{str: "[7]", repl: "<sup>7</sup>"},
		{str: "[8]", repl: "<sup>8</sup>"},
		{str: "[9]", repl: "<sup>9</sup>"},
		{re: regexp.MustCompile(`\b1/2\b`), repl: "&frac12;"},
		{re: regexp.MustCompile(`\b1/4\b`), repl: "&frac14;"},
		{re: regexp.MustCompile(`\b3/4\b`), repl: "&frac34;"},
		{re: regexp.MustCompile(`\b(\d+)_1/2\b`), repl: "${1}&frac12;"},
		{re: regexp.MustCompile(`\b(\d+)_1/4\b`), repl: "${1}&frac14;"},
		{re: regexp.MustCompile(`\b(\d+)_3/4\b`), repl: "${1}&frac34;"},
		{str: "^o", repl: "&deg;"},
		{str: "[c]", repl: "&copy;"},
		{str: "[R]", repl: "&reg;"},
		{str: "AE", repl: "&AElig;"},
		{str: "ae", repl: "&aelig;"},
		{str: "^.", repl: "&bull;"},
		{str: "[/]", repl: "&divide;"},
		{str: "[+]", repl: "&dagger;"},
		{str: "[++]", repl: "&Dagger;"},
		{str: `‵`, repl: "&prime;"},
		{str: `′`, repl: "&bprime;"},
	} {
		if sub.re != nil {
			s = sub.re.ReplaceAllString(s, sub.repl)
		} else {
			s = strings.ReplaceAll(s, sub.str, sub.repl)
		}
	}
	return s
}

func (f *renderPostScriptFormatter) toString(s string) string {
	for _, sub := range []replSet{
		{re: regexp.MustCompile(`([()\\])`), repl: "\\${1}"},
		{str: "§", repl: "\\247"},
		{str: "©", repl: "\\345"},
		{str: "«", repl: "\\253"},
		{str: "®", repl: "\\346"},
		{str: "°", repl: "\\347"},
		{str: "±", repl: "\\354"},
		{str: "⁰", repl: "\\330"},
		{str: "¹", repl: "\\331"},
		{str: "²", repl: "\\332"},
		{str: "³", repl: "\\333"},
		{str: "⁴", repl: "\\334"},
		{str: "⁵", repl: "\\335"},
		{str: "⁶", repl: "\\336"},
		{str: "⁷", repl: "\\337"},
		{str: "⁸", repl: "\\340"},
		{str: "⁹", repl: "\\342"},
		{str: "·", repl: "\\267"},
		{str: "»", repl: "\\273"},
		{str: "¼", repl: "\\355"},
		{str: "½", repl: "\\356"},
		{str: "¾", repl: "\\357"},
		{str: "Ä", repl: "\\200"},
		{str: "Æ", repl: "\\341"},
		{str: "×", repl: "\\360"},
		{str: "ä", repl: "\\220"},
		{str: "æ", repl: "\\361"},
		{str: "÷", repl: "\\344"},
		{str: "‒", repl: "\\362"}, // minus
		{str: "–", repl: "\\261"}, // en dash
		{str: "—", repl: "\\320"}, // em dash
		{str: "“", repl: "\\252"},
		{str: "”", repl: "\\272"},
		{str: "‘", repl: "\\140"},
		{str: "’", repl: "\\047"},
		{str: "†", repl: "\\262"},
		{str: "‡", repl: "\\263"},
		{str: "•", repl: "\\267"},
		{str: "ﬀ", repl: "ff"},
		{str: "ﬁ", repl: "\\256"},
		{str: "ﬂ", repl: "\\257"},
		{str: "ﬃ", repl: "f\\256"},
		{str: "ﬄ", repl: "f\\257"},
		{str: "+/-", repl: "\\354"},
		{str: "---", repl: "\\320"},
		{str: "--", repl: "\\261"},
		{re: regexp.MustCompile(`(^|\s|\d)-(\d+)`), repl: "${1}\\362${2}"},
		{re: regexp.MustCompile(`(\d)x`), repl: "${1}\\360"},
		{re: regexp.MustCompile(`x(\d)`), repl: "\\360${1}"},
		{str: "[x]", repl: "\\360"},
		{str: "[S]", repl: "\\247"},
		{str: "[<<]", repl: "\\253"},
		{str: "[>>]", repl: "\\273"},
		{str: "[0]", repl: "\\330"},
		{str: "[1]", repl: "\\331"},
		{str: "[2]", repl: "\\332"},
		{str: "[3]", repl: "\\333"},
		{str: "[4]", repl: "\\334"},
		{str: "[5]", repl: "\\335"},
		{str: "[6]", repl: "\\336"},
		{str: "[7]", repl: "\\337"},
		{str: "[8]", repl: "\\340"},
		{str: "[9]", repl: "\\342"},
		{re: regexp.MustCompile(`\b1/2\b`), repl: "\\356"},
		{re: regexp.MustCompile(`\b1/4\b`), repl: "\\355"},
		{re: regexp.MustCompile(`\b3/4\b`), repl: "\\357"},
		{re: regexp.MustCompile(`\b(\d+)_1/2\b`), repl: "${1}\\356"},
		{re: regexp.MustCompile(`\b(\d+)_1/4\b`), repl: "${1}\\355"},
		{re: regexp.MustCompile(`\b(\d+)_3/4\b`), repl: "${1}\\357"},
		{str: "^o", repl: "\\347"},
		{str: "[c]", repl: "\\345"},
		{str: "[R]", repl: "\\346"},
		{str: "AE", repl: "\\341"},
		{str: "ae", repl: "\\361"},
		{str: "^.", repl: "\\267"},
		{str: "[/]", repl: "\\344"},
		{str: "[+]", repl: "\\262"},
		{str: "[++]", repl: "\\263"},
		{str: `‵`, repl: "'"},
		{str: `′`, repl: "`"},
	} {
		if sub.re != nil {
			s = sub.re.ReplaceAllString(s, sub.repl)
		} else {
			s = strings.ReplaceAll(s, sub.str, sub.repl)
		}
	}
	return s
}

func (f *renderHTMLFormatter) init(o renderOptSet) {
	f.buf.WriteString("<P>")
	f.listStack = make([]string, 0, 4)
}

func (f *renderHTMLFormatter) cancelStyles() {
	if f.ital {
		f.setItal(false)
	}
	if f.bold {
		f.setBold(false)
	}
}

func (f *renderHTMLFormatter) setItal(b bool) {
	if b {
		f.buf.WriteString("<I>")
	} else {
		f.buf.WriteString("</I>")
	}
	f.ital = b
}

func (f *renderHTMLFormatter) setBold(b bool) {
	if b {
		f.buf.WriteString("<B>")
	} else {
		f.buf.WriteString("</B>")
	}
	f.bold = b
}

func (f *renderHTMLFormatter) process(text string) {
	f.buf.WriteString(f.toString(text))
}

func (f *renderHTMLFormatter) finalize() string {
	f.endPar()
	return f.buf.String()
}

func (f *renderHTMLFormatter) reference(desc, link string) {
	fmt.Fprintf(&f.buf, "<A HREF=\"%s\">%s</A>", strings.ToUpper(link), f.toString(desc))
	// XXX toupper??
}

func (f *renderHTMLFormatter) endPar() {
	f.indent = 0
	f.cancelStyles()
	f.levelSet(0, "", "")
	f.buf.WriteString("</P>")
}

func (f *renderHTMLFormatter) levelSet(level int, tag string, extra string) {
	for level > len(f.listStack) {
		if tag == "" {
			tag = "UL"
		}
		if extra == "" {
			fmt.Fprintf(&f.buf, "<%s>", tag)
		} else {
			fmt.Fprintf(&f.buf, "<%s %s>", tag, extra)
		}
		f.listStack = append(f.listStack, fmt.Sprintf("</%s>", tag))
	}
	for level < len(f.listStack) {
		f.buf.WriteString(f.listStack[len(f.listStack)-1])
		f.listStack = f.listStack[:len(f.listStack)-1]
	}
}

func (f *renderHTMLFormatter) newLine() {
	f.buf.WriteString("<BR/>")
}

func (f *renderHTMLFormatter) newPar() {
	f.endPar()
	f.buf.WriteString("<P>")
}

func (f *renderHTMLFormatter) bulletListItem(level int, bullet rune) {
	if bullet == 0 {
		f.levelSet(level, "UL", "")
	} else {
		var style string
		switch bullet {
		case '*', '\u2022':
			style = "disc"
		case '\u2023':
			style = "\"\\2023\""
		case '\u2043', '-':
			style = "-"
		case '\u25cb', 'o':
			style = "circle"
		case '\u261e':
			style = "\"\\261e\""
		case '\u2605':
			style = "\"\\2605\""
		default:
			style = fmt.Sprintf("\"\\%06x\"", bullet)
		}
		f.levelSet(level, "UL", "style='list-style-type:"+style+";'")
	}
	f.buf.WriteString("<LI>")
}

func (f *renderHTMLFormatter) enumListItem(level, counter int) {
	f.levelSet(level, "OL", fmt.Sprintf("style=\"list-style-type: %s;\"", enumType(level)))
	f.buf.WriteString("<LI>")
}

func (f *renderHTMLFormatter) table(t *textTable) {
	f.buf.WriteString("<TABLE BORDER=1>")
	if len(t.captions) > 0 {
		f.buf.WriteString("<CAPTION>")
		miniFormatter(strings.Join(t.captions, " "), f)
		f.buf.WriteString("</CAPTION>\n")
	}
	f.buf.WriteString("<THEAD>")

	inHead := true
	footnoteSpan := 1
	for _, row := range t.rows {
		if len(row) > 0 && !row[0].header && inHead {
			f.buf.WriteString("</THEAD><TBODY>")
			inHead = false
		}
		if len(row) > footnoteSpan {
			footnoteSpan = len(row)
		}
		f.buf.WriteString("\n<TR>")
		for _, col := range row {
			if col != nil {
				td := "TD"
				al := "LEFT"
				cs := ""
				if col.header {
					td = "TH"
				}
				if col.align == '^' {
					al = "CENTER"
				} else if col.align == '>' {
					al = "RIGHT"
				}
				if col.span > 0 {
					cs = fmt.Sprintf(" COLSPAN=%d", col.span+1)
				}

				fmt.Fprintf(&f.buf, "<%s ALIGN=%s%s>", td, al, cs)
				miniFormatter(col.text, f)
				fmt.Fprintf(&f.buf, "</%s>", td)
			}
		}
		f.buf.WriteString("</TR>")
	}
	if inHead {
		// odd, the table only has headers?
		f.buf.WriteString("</THEAD>")
	} else {
		f.buf.WriteString("</TBODY>")
	}
	f.buf.WriteString("<TFOOT>")
	for _, footer := range t.footnotes {
		fmt.Fprintf(&f.buf, "<TR><TD COLSPAN=%d>", footnoteSpan)
		miniFormatter(footer, f)
		f.buf.WriteString("</TD></TR>")
	}
	f.buf.WriteString("</TFOOT></TABLE>")
}

//
//  ____           _   ____            _       _
// |  _ \ ___  ___| |_/ ___|  ___ _ __(_)_ __ | |_
// | |_) / _ \/ __| __\___ \ / __| '__| | '_ \| __|
// |  __/ (_) \__ \ |_ ___) | (__| |  | | |_) | |_
// |_|   \___/|___/\__|____/ \___|_|  |_| .__/ \__|
//                                      |_|
//
// PostScript output formatter
//
type renderPostScriptFormatter struct {
	buf          strings.Builder
	indent       int
	chunks       []psChunk
	curChunk     []string
	lastSetFont  string
	compact      bool
	ital         bool
	bold         bool
	needOutdent  bool
	wasEverBold  bool
	footnoteMode bool
}

type psChunk struct {
	pre      string
	contents []string
	post     string
}

func (f *renderPostScriptFormatter) setFootnoteMode(b bool) {
	f.footnoteMode = b
}

func (f *renderPostScriptFormatter) title(text string) {
	f.newPar()
	f.sendBuffer("{PsFF_section}")
	f.process(text)
	f.sendBuffer("{PsFF_nl PsFF_rm}")
	f.newPar()
}

func (f *renderPostScriptFormatter) subtitle(text string) {
	f.newLine()
	f.sendBuffer("{PsFF_subsection}")
	f.process(text)
	f.sendBuffer("{PsFF_nl PsFF_rm}")
}
func (f *renderPostScriptFormatter) init(o renderOptSet) {
	f.compact = o.compact
}

func psSimpleEscape(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(s, "\\", `\\`),
			"(", `\(`),
		")", `\)`)
}

func (f *renderPostScriptFormatter) fontChange() string {
	return f.fontChangeWithStartCmd("")
}

func (f *renderPostScriptFormatter) fontChangeWithStartCmd(start string) string {
	newFont := "rm"
	if f.bold && f.ital {
		newFont = "bi"
	} else if f.bold {
		newFont = "bf"
	} else if f.ital {
		newFont = "it"
	}

	ft := ""
	if f.footnoteMode {
		ft = "tbl_footer_"
	}

	if newFont != f.lastSetFont {
		f.lastSetFont = newFont
		return fmt.Sprintf("{PsFF_%s%s %s}", ft, newFont, start)
	}
	return fmt.Sprintf("{%s}", start)
}

func (f *renderPostScriptFormatter) setItal(b bool) {
	f.sendBuffer("{}")
	f.ital = b
}

func (f *renderPostScriptFormatter) setBold(b bool) {
	f.sendBuffer("{}")
	f.bold = b
	f.wasEverBold = true
}

func (f *renderPostScriptFormatter) process(text string) {
	sp := regexp.MustCompile(`^(\S*\s+)(.*)$`)
	for sp.MatchString(text) {
		pieces := sp.FindStringSubmatch(text)
		f.curChunk = append(f.curChunk, pieces[1])
		text = pieces[2]
	}
	if text != "" {
		f.curChunk = append(f.curChunk, text)
	}
}

func (f *renderPostScriptFormatter) textContentAsPSLines() (string, int) {
	lineCount := 1

	f.sendBuffer("{}")
	f.setBold(false)
	f.setItal(false)

	var ps strings.Builder
	ps.WriteString("[(")
	for _, chunk := range f.chunks {
		for _, s := range chunk.contents {
			if strings.Contains(chunk.pre, "PsFF_nl") {
				ps.WriteString(")(")
				lineCount++
			}
			ps.WriteString(f.toString(s))
			if strings.Contains(chunk.post, "PsFF_nl") {
				ps.WriteString(")(")
				lineCount++
			}
		}
	}
	ps.WriteString(")]")
	return ps.String(), lineCount
}

func (f *renderPostScriptFormatter) textContentAsPSString() string {
	f.sendBuffer("{}")
	f.setBold(false)
	f.setItal(false)

	var ps strings.Builder
	ps.WriteRune('(')
	for _, chunk := range f.chunks {
		for _, s := range chunk.contents {
			ps.WriteString(f.toString(s))
		}
	}
	ps.WriteRune(')')
	return ps.String()
}

func (f *renderPostScriptFormatter) finalize() string {
	f.sendBuffer("{}")
	f.setBold(false)
	f.setItal(false)

	f.buf.WriteString(" [ ")
	for _, chunk := range f.chunks {
		f.buf.WriteString(" [ ")
		f.buf.WriteString(chunk.post)
		f.buf.WriteString(" [ ")
		for _, s := range chunk.contents {
			f.buf.WriteString("(")
			f.buf.WriteString(f.toString(s))
			f.buf.WriteString(")")
		}
		f.buf.WriteString(" ] ")
		f.buf.WriteString(chunk.pre)
		f.buf.WriteString(" ] ")
	}
	f.buf.WriteString(" ] ")
	f.chunks = nil
	return f.buf.String()
}

func (f *renderPostScriptFormatter) reference(desc, link string) {
	f.toggleItal()
	f.process(desc)
	f.toggleItal()
}

func (f *renderPostScriptFormatter) newLine() {
	if f.compact {
		f.process(" ")
	} else {
		f.sendBuffer("{PsFF_nl}")
	}
}

func (f *renderPostScriptFormatter) newPar() {
	if f.compact {
		f.process(" ")
	} else {
		if f.needOutdent {
			f.sendBuffer("{PsFF_par 0 PsFF_ind}")
			f.needOutdent = false
		} else {
			f.sendBuffer("{PsFF_par}")
		}
	}
}

func (f *renderPostScriptFormatter) bulletListItem(level int, bullet rune) {
	var psb string

	switch bullet {
	case 0, '*', '\u2022':
		psb = "^."
	case '\u2023':
		psb = ">"
	case '\u2043', '-':
		psb = "-"
	case '\u25cb', 'o':
		psb = "o"
	case '\u261e':
		psb = "[>>]"
	case '\u2605':
		psb = "*"
	default:
		psb = string(bullet)
	}

	if f.compact {
		f.process(fmt.Sprintf(" (%c) ", bullet))
	} else {
		f.newLine()
		f.chunks = append(f.chunks, psChunk{
			pre:      fmt.Sprintf("{ %d PsFF_ind }", level-1),
			contents: []string{psb},
			post:     fmt.Sprintf("{ %d PsFF_ind }", level),
		})
		f.needOutdent = true
	}
}

func (f *renderPostScriptFormatter) sendBuffer(end string) {
	if f.curChunk != nil || end != "{}" {
		f.chunks = append(f.chunks, psChunk{
			pre:      f.fontChange(),
			contents: f.curChunk,
			post:     end,
		})
		f.curChunk = nil
	}
}

func (f *renderPostScriptFormatter) sendBufferWithStartCmd(start, end string) {
	if f.curChunk != nil || end != "{}" {
		f.chunks = append(f.chunks, psChunk{
			pre:      f.fontChangeWithStartCmd(start),
			contents: f.curChunk,
			post:     end,
		})
		f.curChunk = nil
	}
}

func (f *renderPostScriptFormatter) boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (f *renderPostScriptFormatter) toggleItal() {
	f.setItal(!f.ital)
}

func (f *renderPostScriptFormatter) enumListItem(level, counter int) {
	if f.compact {
		f.process(fmt.Sprintf(" (%s) ", enumVal(level, counter)))
	} else {
		f.newLine()
		f.chunks = append(f.chunks, psChunk{
			pre:      fmt.Sprintf("{ %d PsFF_ind }", level-1),
			contents: []string{fmt.Sprintf("%s.", enumVal(level, counter))},
			post:     fmt.Sprintf("{ %d PsFF_ind }", level),
		})
		f.needOutdent = true
	}
}

//
//  For PostScript tables, we handle this by
//  emitting a routine up front which estimates
//  the horizontal space required by each column.
//  this way we let the device, which knows its
//  output parameters and font metrics, so all
//  the math the other formatter classes do here
//  will instead by shipped over to the output
//  device and written in PostScript.
//
//  This defines variables called /PsFF_Cw<n>
//  which hold the size in points for column <n>
//  of the table (0-origin).
//
//  The code for this is essentially:
//  [ <col <n> row 0> <col <n> row 1> ... ] {
//    stringwidth pop dup PsFF_Cw<n> gt {
//      /PsFF_Cw<n> exch def
//    } { pop } ifelse
//  } forall
//
//  As the table cells are typeset, they are put
//  into boxes of width PsFF_Cw<n> using the normal
//  boxed text support we use elsewhere, via the
//  PsFF_tXX procedures.
//
//  For spanned columns, we will skip over the
//  spans when doing the initial calculations,
//  and then emit code for each span which adjusts
//  the column widths:
//
//  % span columns 1-3
//  <text> stringwidth pop PsFF_Cw1 PsFF_Cw2 add
//  PsFF_Cw3 add 2 PsFF_TcolSpn mul add gt {
//    /PsFF_Cw1 PsFF_Cw1 <x> add def
//    /PsFF_Cw2 PsFF_Cw2 <x> add def
//    /PsFF_Cw3 PsFF_Cw3 <x> add def
//  } if
//
//  (note that PsFF_TcolSpn is a constant equal
//  to the amount of space added in a table between
//  columns--this needs to be added back into the
//  size of spanned columns)
//
// CAPTIONS
//	Captions are added to the table by joining the caption cell(s), formatting with onestyle=true
//  sending
//		{PsFF_nl PsFF_rm} [ ...words in caption... ] {PsFF_tbl_caption}
//
func (f *renderPostScriptFormatter) table(t *textTable) {
	var linesPerRow []int

	// Emit routine to calculate column widths, then emit
	// code to render the table
	if f.compact {
		f.process(" [table] ")
		return
	}

	f.sendBuffer("{PsFF_nl}")
	f.bold = false
	f.ital = false

	if len(t.captions) > 0 {
		f.newPar()
		miniFormatter(strings.Join(t.captions, " "), f, oneStyle, oneLine)
		f.sendBufferWithStartCmd("PsFF_tbl_caption", "{PsFF_nl PsFF_rm}")
	}

	//func miniFormatter(text string, formatter renderingFormatter, options ...miniFormatterOption) {

	var ps strings.Builder
	ps.WriteString(`{PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
`)
	var deferred strings.Builder
	for c := 0; c < t.numCols(); c++ {
		fmt.Fprintf(&deferred, `
%% Column #%d of %d
/PsFF_Cw%[1]d 0 def
`, c, t.numCols())
		for r, row := range t.rows {
			if row[c] != nil && row[c].span == 0 {
				tempEnv := &renderPostScriptFormatter{}
				miniFormatter(row[c].text, tempEnv)
				if tempEnv.wasEverBold {
					deferred.WriteString("PsFF_bf")
				} else {
					deferred.WriteString("PsFF_rm")
				}
				lines, nLines := tempEnv.textContentAsPSLines()
				fmt.Fprintf(&deferred, " /PsFF_Cwi 0 def %[1]s {stringwidth pop dup dup (PsFF_CLw%[3]dc%[2]dl) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw%[2]d gt {/PsFF_Cw%[2]d exch def} {pop} ifelse} forall\n",
					lines, c, r)
				if len(linesPerRow) <= r {
					linesPerRow = append(linesPerRow, 0)
				}
				if linesPerRow[r] < nLines {
					linesPerRow[r] = nLines
				}
			}
		}
	}

	for r := 0; r < len(t.rows); r++ {
		for c := 0; c < t.numCols(); c++ {
			for l := 0; l < linesPerRow[r]; l++ {
				fmt.Fprintf(&ps, "/PsFF_CLw%dc%dl%d 0 def\n", r, c, l)
			}
		}
	}
	ps.WriteString(deferred.String())
	ps.WriteString(`%
% Now adjust column widths for the spans
%
`)
	for r, row := range t.rows {
		for i, col := range row {
			if col != nil && col.span > 0 {
				fmt.Fprintf(&ps, "%% span row %d, columns %d-%d:\n", r, i, i+col.span)
				fmt.Fprintf(&ps, "/PsFF__t__have PsFF_TcolSpn %d mul ", col.span)
				for j := i; j <= i+col.span; j++ {
					fmt.Fprintf(&ps, "PsFF_Cw%d add ", j)
				}
				ps.WriteString("def\n/PsFF__t__need 0 def\n")

				tempEnv := &renderPostScriptFormatter{}
				miniFormatter(col.text, tempEnv)
				if tempEnv.wasEverBold {
					ps.WriteString("PsFF_bf")
				} else {
					ps.WriteString("PsFF_rm")
				}
				lines, nLines := tempEnv.textContentAsPSLines()
				fmt.Fprintf(&ps, " /PsFF_Cwi 0 def %[1]s {stringwidth pop dup dup (PsFF_CLw%[3]dc%[2]dl) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF__t__need gt {/PsFF__t__need exch def} {pop} ifelse} forall\n", lines, i, r)
				if len(linesPerRow) <= r {
					linesPerRow = append(linesPerRow, 0)
				}
				if linesPerRow[r] < nLines {
					linesPerRow[r] = nLines
				}

				fmt.Fprintf(&ps, `
PsFF__t__need PsFF__t__have gt {
   /PsFF__t__add PsFF__t__need PsFF__t__have sub def
   /PsFF__t__each PsFF__t__add %d 1 add div def
`, col.span)
				for n := i; n <= i+col.span; n++ {
					fmt.Fprintf(&ps, "   /PsFF_Cw%d PsFF_Cw%[1]d PsFF__t__each add def\n", n)
					ps.WriteString("   /PsFF__t__add PsFF__t__add PsFF__t__each sub def\n")
				}
				fmt.Fprintf(&ps, `   PsFF__t__add 0 gt {
      /PsFF_Cw%d PsFF_Cw%[1]d PsFF__t__add add def
   } if
} if
`, i)
			}
		}
	}
	//
	// now typeset the table itself.
	//
	ps.WriteString("%\n% Table contents\n%\n")
	for r, row := range t.rows {
		// We need to print the column contents a line at a time
		fmt.Fprintf(&ps, "%% Row %d (%d %s)\n", r, linesPerRow[r], util.PluralizeString("line", linesPerRow[r]))
		for l := 0; l < linesPerRow[r]; l++ {
			for c, col := range row {
				fmt.Fprintf(&ps, "%% Row %d, Col %d, Line %d\n", r, c, l)
				// textwidth isheader? iscentered? isright? colwidth PsFF_cell{F,T,B,M} (full, top, bottom, middle)
				if col == nil {
					ps.WriteString("% (spanned through)\n")
				} else {
					fmt.Fprintf(&ps, "PsFF_CLw%dc%dl%d %s %s %s PsFF_Cw%[2]d ", r, c, l,
						f.boolString(col.header), f.boolString(col.align == '^'), f.boolString(col.align == '>'))
					for span := c + 1; span <= c+col.span; span++ {
						fmt.Fprintf(&ps, "PsFF_Cw%d add PsFF_TcolSpn add", span)
					}
					ps.WriteString(" PsFF_cell")
					if linesPerRow[r] == 1 {
						ps.WriteRune('F')
					} else if l == 0 {
						ps.WriteRune('T')
					} else if l == linesPerRow[r]-1 {
						ps.WriteRune('B')
					} else {
						ps.WriteRune('M')
					}

					ps.WriteRune(' ')
					ps.WriteString(f.miniCellFormatter(col.text, l))
					ps.WriteString(" PsFF_tEnd\n")
				}
			}
			ps.WriteString("PsFF_nl\n")
		}
	}
	ps.WriteString("/X PsFF_Xsave def}")
	f.sendBuffer(ps.String())
	f.setFootnoteMode(true)
	for _, footer := range t.footnotes {
		miniFormatter(footer, f, footnoteStyle)
		f.sendBufferWithStartCmd("PsFF_tbl_footer", "{PsFF_nl PsFF_rm}")
	}
	f.setFootnoteMode(false)
}

//  ___                   _     ____
// |_ _|_ __  _ __  _   _| |_  |  _ \ __ _ _ __ ___  ___ _ __
//  | || '_ \| '_ \| | | | __| | |_) / _` | '__/ __|/ _ \ '__|
//  | || | | | |_) | |_| | |_  |  __/ (_| | |  \__ \  __/ |
// |___|_| |_| .__/ \__,_|\__| |_|   \__,_|_|  |___/\___|_|
//           |_|

//
// Incoming list items are marked with one of these, to be expanded
// later by the output formatter.
//
type listItem struct {
	bullet rune // '*' for bullet lists or '#' for enumerated
	level  int  // nesting level
}

//
// Representation of a table as a slice of rows, each of which
// is a slice of tableCells.
//
type textTable struct {
	rows      [][]*tableCell
	footnotes []string
	captions  []string
}

//
// Count number of columns
//
func (t *textTable) numCols() int {
	nc := 0
	for _, r := range t.rows {
		nc = max(nc, len(r))
	}
	return nc
}

//
// Add a new caption to a textTable.
//
func (t *textTable) addCaption(caption string) {
	t.captions = append(t.captions, caption)
}

//
// Add a new footer line to a textTable.
//
func (t *textTable) addFooter(footer string) {
	t.footnotes = append(t.footnotes, footer)
}

//
// Add a new row to a textTable.
// Handles spanned columns.
//
func (t *textTable) addRow(cols []string) error {
	if t.rows == nil {
		t.rows = make([][]*tableCell, 0, 32)
	}
	row := make([]*tableCell, 0, 8)
	var lastUsed *tableCell = nil
	for _, col := range cols {
		if strings.HasPrefix(col, "-") {
			if lastUsed == nil {
				return fmt.Errorf("table cell cannot span into the first column")
			}
			lastUsed.span++
			row = append(row, nil)
		} else {
			lastUsed = newTableCell(col)
			row = append(row, lastUsed)
		}
	}
	t.rows = append(t.rows, row)
	return nil
}

//
// Incoming table rows are represented in the input stream
// with one of these, which notes that this should be a table
// row and holds the markup text source from that line of input.
//
type tableRow struct {
	src string
}

//
// This represents a specific cell in a textTable.
//
type tableCell struct {
	text   string // text that belongs in this cell
	span   int    // number of columns to the right to take up
	align  rune   // '<' (left), '^' (center), or '>' (right)
	header bool   // true if this is a header cell
}

//
// Create a new tableCell, figuring out alignment
// from context (based on leading and/or trailing space)
//
// Also identifies cells which start with '=' as headers.
//
func newTableCell(text string) *tableCell {
	c := &tableCell{
		align: '<',
	}
	if text != "" && text[0] == '=' {
		c.header = true
		text = text[1:]
	}
	rt := []rune(text)
	if text != "" && unicode.IsSpace(rt[0]) {
		if unicode.IsSpace(rt[len(rt)-1]) {
			c.align = '^'
		} else {
			c.align = '>'
		}
	}
	c.text = strings.TrimSpace(text)
	return c
}

//
// General-purpose functions to generate enumerated list
// numbering.
//
func enumType(level int) string {
	switch (level - 1) % 5 {
	case 0:
		return "decimal"
	case 1:
		return "lower-alpha"
	case 2:
		return "lower-roman"
	case 3:
		return "upper-alpha"
	case 4:
		return "upper-roman"
	}
	return "decimal"
}

func enumVal(level, value int) string {
	switch (level - 1) % 5 {
	case 0:
		return strconv.Itoa(value)
	case 1:
		v := ""
		for value > 0 {
			vv := (value - 1) % 26
			v = "abcdefghijklmnopqrstuvwxyz"[vv:vv+1] + v
			value /= 26
		}
		return v
	case 2:
		v, err := ToRoman(value)
		if err != nil {
			return "?"
		}
		return strings.ToLower(v)
	case 3:
		v := ""
		for value > 0 {
			vv := (value - 1) % 26
			v = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"[vv:vv+1] + v
			value /= 26
		}
		return v
	case 4:
		v, err := ToRoman(value)
		if err != nil {
			return "?"
		}
		return v
	}
	return "?"
}

type miniFormatterOption byte

const (
	oneStyle miniFormatterOption = 1 << iota
	oneLine
	footnoteStyle
)

//
// Limited style rendering within certain restricted environments
// such as table cells.
//
// We allow \e, \v, \., \\, **, //, [[..]]; we can collapse to a single style
// also if necessary.
//

func miniFormatter(text string, formatter renderingFormatter, options ...miniFormatterOption) {
	linkPattern := regexp.MustCompile(`\[\[(.*?)(?:\|(.*?))?\]\]`)
	splitterPattern := regexp.MustCompile(`\[\[.*?\]\]|\\e|\\v|\\\\|\\\.|\*\*|//`)
	var opts miniFormatterOption

	for _, o := range options {
		opts |= o
	}

	isBold := false
	isItal := false
	hasLinks := 0
	var theLink string
	var theLinkAttrs []string

	if (opts & oneStyle) != 0 {
		// run through the string to see what formatting we'll be using
		for _, frag := range splitterPattern.FindAllString(text, -1) {
			if frag == "**" {
				isBold = true
			} else if frag == "//" {
				isItal = true
			} else if strings.HasPrefix(frag, "[[") {
				hasLinks++
				theLink = frag
			}
		}
		if hasLinks == 1 {
			// if there's a single link, use the whole text as the link.
			// otherwise, don't use any.
			isBold = false
			isItal = false
			theLinkAttrs = linkPattern.FindStringSubmatch(theLink)
			theLinkText := ""
			idx := 0
			for _, ii := range splitterPattern.FindAllStringIndex(text, -1) {
				if ii[0] > idx {
					theLinkText += text[idx:ii[0]]
				}
				switch text[ii[0]:ii[1]] {
				case "\\e":
					theLinkText += "\\"
				case "\\v":
					theLinkText += "|"
				case "\\.", "\\\\", "**", "//":
				default:
					if strings.HasPrefix(text[ii[0]:], "[[") {
						if theLinkAttrs[1] == "" {
							theLinkText += theLinkAttrs[0]
						} else {
							theLinkText += theLinkAttrs[1]
						}
					} else {
						theLinkText += text[ii[0]:ii[1]]
					}
				}
				idx = ii[1]
			}
			if idx < len(text) {
				theLinkText += text[idx:]
			}
			if theLinkAttrs[1] == "" {
				formatter.reference(theLinkAttrs[0], theLinkAttrs[0])
			} else {
				formatter.reference(theLinkAttrs[1], theLinkAttrs[0])
			}
			return
		} else {
			hasLinks = 0
			theLink = ""
			theLinkAttrs = nil
			formatter.setBold(isBold)
			formatter.setItal(isItal)
		}
	}

	idx := 0
	for _, fragmentRange := range splitterPattern.FindAllStringIndex(text, -1) {
		if fragmentRange[0] > idx {
			formatter.process(text[idx:fragmentRange[0]])
		}
		switch text[fragmentRange[0]:fragmentRange[1]] {
		case "\\e":
			formatter.process("\\")
		case "\\v":
			formatter.process("|")
		case "\\.":
		case "\\\\":
			if (opts & oneLine) == 0 {
				formatter.newLine()
			}
		case "**":
			if (opts & oneStyle) == 0 {
				isBold = !isBold
				formatter.setBold(isBold)
			}
		case "//":
			if (opts & oneStyle) == 0 {
				isItal = !isItal
				formatter.setItal(isItal)
			}
		default:
			if (opts & oneStyle) == 0 {
				if linkParts := linkPattern.FindStringSubmatch(text[fragmentRange[0]:fragmentRange[1]]); linkParts != nil {
					if len(linkParts) < 2 || linkParts[1] == "" {
						formatter.reference(linkParts[0], linkParts[0])
					} else {
						formatter.reference(linkParts[1], linkParts[0])
					}
				} else {
					formatter.process(text[fragmentRange[0]:fragmentRange[1]])
				}
			} else {
				formatter.process(text[fragmentRange[0]:fragmentRange[1]])
			}
		}
		idx = fragmentRange[1]
	}
	if idx < len(text) {
		formatter.process(text[idx:])
	}
	if (opts & oneStyle) != 0 {
		formatter.setItal(!isItal)
		formatter.setBold(!isBold)
	}
}

func (f *renderPostScriptFormatter) miniCellFormatter(text string, line int) string {
	var ps strings.Builder
	linkPattern := regexp.MustCompile(`\[\[(.*?)(?:\|(.*?))?\]\]`)
	splitterPattern := regexp.MustCompile(`\[\[.*?\]\]|\\e|\\v|\\\\|\\\.|\*\*|//`)

	isBold := false
	isItal := false

	setCurrentFont := func() {
		if isBold && isItal {
			ps.WriteString("PsFF_bi ")
		} else if isBold {
			ps.WriteString("PsFF_bf ")
		} else if isItal {
			ps.WriteString("PsFF_it ")
		} else {
			ps.WriteString("PsFF_rm ")
		}
	}

	idx := 0
	ps.WriteString("PsFF_rm ")
	for _, fragmentRange := range splitterPattern.FindAllStringIndex(text, -1) {
		if fragmentRange[0] > idx && line == 0 {
			fmt.Fprintf(&ps, "(%s) PsFF_cellfragment ", f.toString(text[idx:fragmentRange[0]]))
		}
		if text[fragmentRange[0]:fragmentRange[1]] == "\\\\" {
			line--
			if line < 0 {
				break
			}
			idx = fragmentRange[1]
			continue
		}
		if line != 0 {
			idx = fragmentRange[1]
			continue
		}

		switch text[fragmentRange[0]:fragmentRange[1]] {
		case "\\e":
			ps.WriteString("(\\\\) PsFF_cellfragment ")
		case "\\v":
			ps.WriteString("(|) PsFF_cellfragment ")
		case "\\.":
		case "**":
			isBold = !isBold
			setCurrentFont()
		case "//":
			isItal = !isItal
			setCurrentFont()
		default:
			if linkParts := linkPattern.FindStringSubmatch(text[fragmentRange[0]:fragmentRange[1]]); linkParts != nil {
				oldItal := isItal
				isItal = true
				setCurrentFont()
				if len(linkParts) < 2 || linkParts[1] == "" {
					fmt.Fprintf(&ps, "(%s) PsFF_cellfragment ", f.toString(linkParts[0]))
				} else {
					fmt.Fprintf(&ps, "(%s) PsFF_cellfragment ", f.toString(linkParts[1]))
				}
				isItal = oldItal
				setCurrentFont()
			} else {
				fmt.Fprintf(&ps, "(%s) PsFF_cellfragment ", f.toString(text[fragmentRange[0]:fragmentRange[1]]))
			}
		}
		idx = fragmentRange[1]
	}
	if idx < len(text) && line == 0 {
		fmt.Fprintf(&ps, "(%s) PsFF_cellfragment ", f.toString(text[idx:]))
	}
	return ps.String()
}

//
// Render converts its input text (in our simple markup notation described
// below) to an output format as specified by the option(s) passed after
// the input text in the parameter list.
//
// The set of options which may follow the string to be formatted
// include these which select the overall output format:
//   AsPlainText  -- render a text-only version of the input
//                   (this is the default)
//   AsHTML       -- render an HTML version of the input
//   AsPostScript -- render a PostScript version of the input
//                   (requires the GMA PostScript preamble and
//                   other supporting code; this merely produces
//                   the formatted text block to the PostScript
//                   data being produced by the application)
//
// and these options to control specific formatting in the selected
// output format:
//   WithBullets(...)  -- use a custom bullet sequence
//   WithCompactText   -- squish verbose text blocks down a bit*
//
//  *(PostScript format only)
//
// The markup syntax is described in gma-markup-syntax(7) and in the
// MarkupSyntax constant string in this package.
//
func Render(text string, opts ...func(*renderOptSet)) (string, error) {
	ops := renderOptSet{
		formatter: &renderPlainTextFormatter{},
		bulletSet: []rune{0},
	}
	for _, o := range opts {
		o(&ops)
	}
	collapseSpaces := regexp.MustCompile(`\s{2,}`)
	newListBullet := regexp.MustCompile(`^[@#]+`)
	formatReqs := regexp.MustCompile(`//|\*\*|\[\[|\]\]|\\\.|\\v|\\e|==\[|\]==|==\(|\)==`)

	paragraphs := make([][][]any, 0, 10)
	thisParagraph := make([][]any, 0, 10)
	thisLine := make([]any, 0, 10)

	for _, sourceLine := range strings.Split(text, "\n") {
		sourceLine = collapseSpaces.ReplaceAllLiteralString(sourceLine, " ")
		//
		// Look for a blank line after a non-blank line. This
		// will be a paragraph break. Other blank lines are
		// discarded.
		//
		trimmedSourceLine := strings.TrimSpace(sourceLine)
		if trimmedSourceLine == "" {
			if len(thisLine) > 0 {
				thisParagraph = append(thisParagraph, thisLine)
				thisLine = make([]any, 0, 10)

				paragraphs = append(paragraphs, thisParagraph)
				thisParagraph = make([][]any, 0, 10)
			}
			continue
		}
		//
		// Look for @ or # at the start of the line. This begins
		// a new list item. We'll just insert a marker to that effect
		// in the current line and keep going, to avoid extra line
		// breaks from sneaking in.
		//
		// Also Look for | at the start of the line. This begins
		// a table row here, which includes the remainder of the
		// input line.
		//
		if bullet := newListBullet.FindString(sourceLine); bullet != "" {
			r := []rune(bullet)
			thisLine = append(thisLine, listItem{
				bullet: r[len(r)-1],
				level:  len(r),
			})
			sourceLine = strings.TrimSpace(sourceLine[len(bullet):])
		} else if strings.HasPrefix(sourceLine, "|") {
			thisLine = append(thisLine, tableRow{
				src: sourceLine,
			})
			continue
		} else {
			sourceLine = trimmedSourceLine
		}
		//
		// add other text to list of lines, breaking on \\
		//
		for strings.Contains(sourceLine, "\\\\") {
			parts := strings.SplitN(sourceLine, "\\\\", 2)
			thisLine = append(thisLine, parts[0])
			thisParagraph = append(thisParagraph, thisLine)
			thisLine = make([]any, 0, 10)
			sourceLine = parts[1]
		}
		thisLine = append(thisLine, sourceLine)
	}
	if len(thisLine) > 0 {
		thisParagraph = append(thisParagraph, thisLine)
	}
	if len(thisParagraph) > 0 {
		paragraphs = append(paragraphs, thisParagraph)
	}
	thisLine = nil
	thisParagraph = nil

	//
	// Now we emit the collected paragraphs of text, handling the other
	// markup tags as we go, using our formatter to render them as desired.
	//
	ital := false
	bold := false
	firstPar := true
	var currentTable *textTable
	ops.formatter.init(ops)

	for _, par := range paragraphs {
		firstLine := true
		enumCounters := []int{0, 0}

		if currentTable != nil {
			ops.formatter.table(currentTable)
			currentTable = nil
		}

		if !firstPar {
			if bold {
				ops.formatter.setBold(false)
				bold = false
			}
			if ital {
				ops.formatter.setItal(false)
				ital = false
			}
			ops.formatter.newPar()
		} else {
			firstPar = false
		}

		pendingReference := ""
		pendingTitle := ""
		pendingSubtitle := ""
		suppressSpace := false
		for _, fragments := range par {
			if pendingReference != "" {
				// false alarm; the [[ we saw earlier were just brackets
				ops.formatter.process(pendingReference)
				pendingReference = ""
			}

			if !firstLine {
				ops.formatter.newLine()
			} else {
				firstLine = false
			}

			for i, f := range fragments {
				switch fragment := f.(type) {
				case string:
					if i+1 < len(fragments) {
						switch fragments[i+1].(type) {
						case string:
							if len(fragment) > 0 {
								fragment += " "
							}
						}
					}
					if currentTable != nil {
						ops.formatter.table(currentTable)
						currentTable = nil
					}

					// break apart string from inline formatting requests
					ixs := formatReqs.FindAllStringIndex(fragment, -1)
					if ixs == nil {
						if len(fragment) > 0 {
							ops.formatter.process(fragment)
						}
					} else {
						// collect the delimeters AND surrounding text together
						pieces := make([]string, 0, 10)
						cur := 0
						for _, ix := range ixs {
							if ix[0] > cur {
								pieces = append(pieces, fragment[cur:ix[0]])
							}
							pieces = append(pieces, fragment[ix[0]:ix[1]])
							cur = ix[1]
						}
						if cur < len(fragment) {
							pieces = append(pieces, fragment[cur:])
						}

						for _, piece := range pieces {
							if piece == " " && suppressSpace {
								suppressSpace = false
								continue
							}
							suppressSpace = false

							if piece == "\\." {
								continue
							}
							if pendingReference != "" {
								if piece == "]]" {
									parts := strings.SplitN(pendingReference[2:], "|", 2)
									if len(parts) == 1 {
										ops.formatter.reference(parts[0], parts[0])
									} else {
										ops.formatter.reference(parts[1], parts[0])
									}
									pendingReference = ""
								} else {
									pendingReference += piece
								}
							} else if pendingTitle != "" {
								if piece == "]==" {
									ops.formatter.title(pendingTitle[3:])
									pendingTitle = ""
									suppressSpace = true
								} else {
									pendingTitle += piece
								}
							} else if pendingSubtitle != "" {
								if piece == ")==" {
									ops.formatter.subtitle(pendingSubtitle[3:])
									pendingSubtitle = ""
									suppressSpace = true
								} else {
									pendingSubtitle += piece
								}
							} else {
								switch piece {
								case "//":
									ital = !ital
									ops.formatter.setItal(ital)

								case "**":
									bold = !bold
									ops.formatter.setBold(bold)

								case "[[":
									pendingReference = "[["

								case "\\.":
									// ignore
								case "\\e":
									ops.formatter.process("\\")
								case "\\v":
									ops.formatter.process("|")
								case "==[":
									pendingTitle = "==["
								case "==(":
									pendingSubtitle = "==("
								default:
									ops.formatter.process(piece)
								}
							}
						}
					}

				case tableRow:
					if currentTable == nil {
						currentTable = &textTable{}
					}
					columns := strings.Split(fragment.src[1:], "|")
					if len(columns) > 1 && columns[len(columns)-1] == "" {
						// trailing | is optional; remove it
						columns = columns[:len(columns)-1]
					}
					if len(columns) > 0 {
						if strings.HasPrefix(columns[0], "::") {
							currentTable.addFooter(columns[0][2:])
						} else if strings.HasPrefix(columns[0], ":") {
							currentTable.addCaption(columns[0][1:])
						} else {
							currentTable.addRow(columns)
						}
					} else {
						currentTable.addRow(nil)
					}

				case listItem:
					if currentTable != nil {
						ops.formatter.table(currentTable)
						currentTable = nil
					}
					if fragment.bullet == '#' {
						for len(enumCounters) < fragment.level {
							enumCounters = append(enumCounters, 0)
						}
						enumCounters[fragment.level-1]++
						ops.formatter.enumListItem(fragment.level, enumCounters[fragment.level-1])
					} else {
						ops.formatter.bulletListItem(fragment.level,
							ops.bulletSet[(fragment.level-1)%len(ops.bulletSet)])
					}

				default:
					return "", fmt.Errorf("internal error in list Rendering: unknown element type detected")
				}
			}
		}
	}

	if currentTable != nil {
		ops.formatter.table(currentTable)
		currentTable = nil
	}

	return ops.formatter.finalize(), nil
}

//
// Character counter formatter. You can point miniFormatter into this
// as a rendering engine but all it does is count the characters that
// were sent to it.
//
// This is only intended for use in restricted environments where
// miniFormatter can work, so not all markup is supported.
//
type countingFormatter struct {
	buf    strings.Builder
	filter func(string) string
}

func (f *countingFormatter) init(options renderOptSet) {}
func (f *countingFormatter) newPar() {
	f.buf.WriteRune('\n')
}
func (f *countingFormatter) process(text string) {
	if f.filter != nil {
		f.buf.WriteString(f.filter(text))
	} else {
		f.buf.WriteString(text)
	}
}
func (f *countingFormatter) finalize() string { return "" }
func (f *countingFormatter) setBold(on bool)  {}
func (f *countingFormatter) setItal(on bool)  {}
func (f *countingFormatter) newLine() {
	f.buf.WriteRune('\n')
}
func (f *countingFormatter) table(t *textTable) {}
func (f *countingFormatter) reference(displayname, linkName string) {
	if f.filter != nil {
		f.buf.WriteString(f.filter(displayname))
	} else {
		f.buf.WriteString(displayname)
	}
}
func (f *countingFormatter) title(text string)                     {}
func (f *countingFormatter) subtitle(text string)                  {}
func (f *countingFormatter) toString(text string) string           { return "" }
func (f *countingFormatter) bulletListItem(level int, bullet rune) {}
func (f *countingFormatter) enumListItem(level, counter int)       {}
func (f *countingFormatter) textWidth() int {
	return utf8.RuneCountInString(f.buf.String())
}
func (f *countingFormatter) maxLineWidth() int {
	maxLen := 0
	for _, s := range strings.Split(f.buf.String(), "\n") {
		l := utf8.RuneCountInString(s)
		if l > maxLen {
			maxLen = l
		}
	}
	return maxLen
}

const MarkupSyntax = `
==[GMA Text Markup Syntax]==
The markup syntax is simple. Lines are collected together into a single
logical line which is then wrapped as appropriate to the output format
(which may rely on whatever is printing the output to break lines as
it prefers).

==(Basic Markup)==
A blank line marks a paragraph break.

**\e\e** marks a line break.

A literal **\e** or **|** character may be entered without being interpreted as part of markup syntax using the codes **\ee** and **\ev** respectively.

**/\./text/\./** sets "text" in Italics*[+]\\
**\.*\.*text\.*\.*** sets "text" in boldface*[+]

**@**//blah//... Starts bulleted list item[++]\\
**@@**//blah//... Starts level-2 bulleted list item[++]\\
**@@@**//blah//... Starts level-3 bulleted list item (and so forth)[++]

**#**//blah//... Starts enumerated list item[++]\\
**##**//blah//... ...and so forth[++]

**[\.[**//name//**]\.]** Creates a hyperlink to "//name//" where this name itself adequately
identifies the linked-to element in GMA (e.g., the name of a spell).\\
**[\.[**//link//**|**//name//**]\.]** Creates a hyperlink called "//name//" which links to GMA element "//link//".

**\e.** does nothing but serves to disambiguate things or prevent otherwise special symbols
from being interpreted as markup syntax.

There is also a special page-break marker **<\.<-\.->\.>** which is not actually processed
by this package, but some output subsystems recognize it when they see it in the output
(e.g., PostScript formatted text blocks).

==(Special Characters)==
Many common Unicode characters are recognized on input, but support for
them in the target output format is not guaranteed.

The following markup symbols may also be used to represent special characters:
|**[\.S]**   |[S] section       |**A\.E**   |AE ligature |
|**[\.c]**   |[c] copyright     |**a\.e**   |ae ligature |
|**[\.<<]**  |[<<] << quotes    |**[\.>>]** |[>>] >> quotes |
|**[\.R]**   |[R] registered    |**1\./4**  |1/4 |
|**^\.o**    |^o  degrees       |**1\./2**  |1/2 |
|**+\.-**    |+/- plusminus     |**3\./4**  |3/4 |
|**^\..**    |^.  bullet        |**[\.x]**  |[x] multiplication |
|**[\.0]**   |[0] superscript 0 |**[\./]**  |[/] division |
|**[\.1]**   |[1] superscript 1 |**-**      |\.- hyphen or minus sign |
|**[\.2]**   |[2] superscript 2 |**-\.-**   |\.-- en dash |
|**[\.3]**   |[3] superscript 3 |**-\.-\.-** |\.--- em dash |
|**[\.4]**   |[4] superscript 4 |**‵** |‘ open single quote |
|**[\.5]**   |[5] superscript 5 |**′** |' close single quote |
|**[\.6]**   |[6] superscript 6 |**‵‵** |“ open double quote |
|**[\.7]**   |[7] superscript 7 |**′′** |'' close double quote|
|**[\.8]**   |[8] superscript 8 |**[\.+]** |[+] dagger|
|**[\.9]**   |[9] superscript 9 |**[\.++]** |[++] double dagger|

The letter **x** immediately next to a digit causes it to be printed as a multiplication sign (e.g., x2 or 3x).\\
A hyphen **-** immediately before a digit causes it to be printed as a minus sign instead of a hyphen (e.g., -1).\\
Separate numbers from fractions with an underscore (e.g., **12\._\.1\./2** prints as **12_1/2**).

==(Titles)==
=\.=[Main (top-level) Heading]==\\
=\.=(Subtitle (2nd-level))==

==(Tables)==
Tables are specified by a set of lines beginning with a **|** character.[++]
Each column in the table is separated from the others with **|** characters
as well. A **|** at the very end of the row is optional.

**|=Size Code|=Area|**\\
**|  S  |  5|**\\
**|  M  |  5|**\\
**|  L  | 10|**

This produces a table like

|=Size Code|=Area|
|  S  |  5|
|  M  |  5|
|  L  | 10|

Table cells beginning with **=** are headers (usually placed in the first row)

Cells are left- or right-justified if there is leading or trailing space between the **|**
separators for that cell, respectively. If there is space before and after the text,
it is centered. In the example above, the size codes will be centered in their column
and the area numbers are right-justified in theirs.

Cells which begin with a hyphen (**-**) indicate that the cell to their left spans into them.
For example:

**|=Column A|=Column B|=Column C**\\
**|stuff    |more stuff|and more**\\
**|a really wide column|- |hello**

produces:

|=Column A|=Column B|=Column C
|stuff    |more stuff|and more
|a really wide column|- |hello

A row in the form

**|: Table Caption |**

places a caption on the table (usually above the table), while each row in the form

**|:: Footnote |**

adds a footnote at the bottom of the table. Footnotes may contain **\e\e** to make explicit line breaks but captions cannot.
Captions and footnotes should be a single cell per line regardless of the number of columns the table has.

So:

**|= Die Roll |= Color |**\\
**|    1    | blue  |**\\
**|   2--4  | green |**\\
**|    5+   | plaid[\.1] |**\\
**|:Random Colors (d8)|**\\
**|::[\.1]Or reroll.|**\\
**|::(Subject to GM discretion.)|**

produces:

|= Die Roll |= Color |
|    1    | blue  |
|   2--4  | green |
|    5+   | plaid[1] |
|:Random Colors (d8)|
|::[1]Or reroll.|
|::(Subject to GM discretion.)|

==(Notes:)==
*May cross line boundaries but not paragraphs.\\
[+]May nest as in **/\./Italic *\.*and*\.* bold/\./.**\\
[++]Must appear at the very beginning of a line.

`

// @[00]@| Go-GMA 5.29.0
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2025 by Steven L. Willoughby (AKA MadScienceZone)
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
