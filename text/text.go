/*
########################################################################################
#  _______  _______  _______                ___       ______      _______              #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___   )             #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   )  |             #
# | |      | || || || (___) |            / (_) (_       ___) /       /   )             #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (      _/   /              #
# | | \_  )| |   | || (   ) | Game           ) (           ) \    /   _/               #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _ (   (__/\             #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_)\_______/             #
#                                                                                      #
########################################################################################
*/

//
// Text processing facilities used by GMA.
//
package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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
// Converts an integer value to a Roman numeral string.
// This will return "0" for a zero value.
//
func ToRoman(i int) (string, error) {
	var roman strings.Builder

	if i < 0 {
		return "", fmt.Errorf("Cannot represent negative values in Roman numerals.")
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
// Converts a Roman numeral string to integer.
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

//==================================================================
//  _____ _______  _______   __  __    _    ____  _  ___   _ ____
// |_   _| ____\ \/ /_   _| |  \/  |  / \  |  _ \| |/ / | | |  _ \
//   | | |  _|  \  /  | |   | |\/| | / _ \ | |_) | ' /| | | | |_) |
//   | | | |___ /  \  | |   | |  | |/ ___ \|  _ <| . \| |_| |  __/
//   |_| |_____/_/\_\ |_|   |_|  |_/_/   \_\_| \_\_|\_\\___/|_|
//

//
// Options to the Render() function are tracked in
// this structure.
//
type renderOptSet struct {
	formatter renderingFormatter
	bulletSet []rune
}

type renderOpts func(*renderOptSet)

// Select plain text output format
func AsPlainText(o *renderOptSet) {
	o.formatter = &renderPlainTextFormatter{}
}

// Select HTML text output format
func AsHTML(o *renderOptSet) {
	o.formatter = &renderHTMLFormatter{}
}

//
// Specify a custom set of bullet characters
// to use for bulleted lists. The bullets passed
// to this option are used in order, then the list
// repeats over as necessary for additional levels.
//
// Example:
//  formattedText, err := Render(srcText, AsPlainText, WithBullets('*', '-'))
// This will alternate between '*' and '-' as bullets at each level.
//
func WithBullets(bullets ...rune) renderOpts {
	return func(o *renderOptSet) {
		o.bulletSet = bullets
	}
}

//
// Each output formatter must supply these methods
// which the Render() function will invoke as it parses
// the marked up source text.
//
type renderingFormatter interface {
	init()
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

func (f *renderPlainTextFormatter) init() {}

func (f *renderPlainTextFormatter) setItal(b bool) {
	f.ital = b
}

func (f *renderPlainTextFormatter) setBold(b bool) {
	f.bold = b
}

func (f *renderPlainTextFormatter) process(text string) {
	f.buf.WriteString(text)
}

func (f *renderPlainTextFormatter) finalize() string {
	return f.buf.String()
}

func (f *renderPlainTextFormatter) reference(desc, link string) {
	f.buf.WriteString(desc)
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
		bullet = '*'
	}
	fmt.Fprintf(&f.buf, "\n%*s%c  ", (level-1)*3, "", bullet)
	f.indent = level
}

func (f *renderPlainTextFormatter) enumListItem(level, counter int) {
	fmt.Fprintf(&f.buf, "\n%*s%s. ", (level-1)*3, "", enumVal(level, counter))
	f.indent = level
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
				colsize[i] = max(colsize[i], len(col.text))
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
				spaceNeeded := len(col.text) - 3*col.span
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
	for _, c := range colsize {
		f.buf.WriteRune('+')
		for xx := 0; xx < c+2; xx++ {
			f.buf.WriteRune('-')
		}
	}
	f.buf.WriteString("+\n")
	for _, row := range t.rows {
		headerRow := false
		for c := 0; c < len(row); c++ {
			if row[c] != nil {
				colwidth := sum(colsize[c:c+row[c].span+1]...) + 3*row[c].span
				if row[c].header {
					fmt.Fprintf(&f.buf, "| %-*s ", colwidth,
						fmt.Sprintf("%-*s%s", (colwidth-len(row[c].text))/2, "", strings.ToUpper(row[c].text)))
					headerRow = true
				} else {
					switch row[c].align {
					case '>':
						fmt.Fprintf(&f.buf, "| %*s ", colwidth, row[c].text)
					case '^':
						fmt.Fprintf(&f.buf, "| %-*s ", colwidth,
							fmt.Sprintf("%-*s%s", (colwidth-len(row[c].text))/2, "", row[c].text))
					default:
						fmt.Fprintf(&f.buf, "| %-*s ", colwidth, row[c].text)
					}
				}
			}
		}
		f.buf.WriteString("|\n")
		if headerRow {
			for _, c := range colsize {
				f.buf.WriteRune('+')
				for xx := 0; xx < c+2; xx++ {
					f.buf.WriteRune('-')
				}
			}
			f.buf.WriteString("+\n")
		}
	}
	for _, c := range colsize {
		f.buf.WriteRune('+')
		for xx := 0; xx < c+2; xx++ {
			f.buf.WriteRune('-')
		}
	}
	f.buf.WriteString("+\n")
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

func (f *renderHTMLFormatter) init() {
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
	f.buf.WriteString(text)
}

func (f *renderHTMLFormatter) finalize() string {
	f.endPar()
	return f.buf.String()
}

func (f *renderHTMLFormatter) reference(desc, link string) {
	fmt.Fprintf(&f.buf, "<A HREF=\"%s\">%s</A>", strings.ToUpper(link), desc)
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
		f.levelSet(level, "UL", fmt.Sprintf("style='list-style-type:\"\\%X\";'", bullet))
	}
	f.buf.WriteString("<LI>")
}

func (f *renderHTMLFormatter) enumListItem(level, counter int) {
	f.levelSet(level, "OL", fmt.Sprintf("style=\"list-style-type: %s;\"", enumType(level)))
	f.buf.WriteString("<LI>")
}

func (f *renderHTMLFormatter) table(t *textTable) {
	f.buf.WriteString("<TABLE BORDER=1>")
	for _, row := range t.rows {
		f.buf.WriteString("<TR>")
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

				fmt.Fprintf(&f.buf, "<%s ALIGN=%s%s>%s</%s>",
					td, al, cs, col.text, td)
			}
		}
		f.buf.WriteString("</TR>")
	}
	f.buf.WriteString("</TABLE>")
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
	rows [][]*tableCell
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
	if text != "" && unicode.IsSpace(rune(text[0])) {
		if unicode.IsSpace(rune(text[len(text)-1])) {
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

//
// Markup text rendering is done via the Render function.
// The set of options which may follow the string to be formatted
// include these which select the overall output format:
//   AsPlainText  -- render a text-only version of the input
//                   (this is the default)
//   AsHTML       -- render an HTML version of the input
//
// and these options to control specific formatting in the selected
// output format:
//   WithBullets(...)  -- use a custom bullet sequence
//
// The markup syntax is simple. Lines are collected together into a single
// logical line which is then wrapped as appropriate to the output format
// (which may rely on whatever is printing the output to break lines as
// it prefers).
//
// A blank line marks a paragraph break.
//
// \\ marks a line break.
//
// //text// sets "text" in Italics*†
//
// **text** sets "text" in boldface*†
//
// *blah... Starts bulleted list item‡
//
// **blah... Starts level-2 bulleted list item‡
//
// ***blah... Starts level-3 bulleted list item (and so forth)‡
//
// #blah... Starts enumerated list item‡
//
// ##blah... ...and so forth‡
//
// [[name]] Creates a hyperlink to "name" where this name itself adequately
// identifies the linked-to element in GMA (e.g., the name of a spell).
//
// [[link|name]] Creates a hyperlink called "name" which links to GMA element "link".
//
// \. does nothing but serves to disambiguate things such as ** to begin
// a section of bolded text from ** to begin a 2nd-level bulleted item
// since the latter must be at the very start of the line.
//
// There is also a special page-break marker <<-->> which is not actually processed
// by this package, but some output subsystems recognize it when they see it in the output
// (e.g., PostScript formatted text blocks).
//
// Tables are specified by a set of lines beginning with a | character.‡
// Each column in the table is separated from the others with | characters
// as well. A | at the very end of the row is optional.
//   |=Size Code|=Area|
//   |  S  |  5|
//   |  M  |  5|
//   |  L  | 10|
// This produces a table like
//   +-----------+------+
//   | SIZE CODE | AREA |
//   +-----------+------+
//   |     S     |    5 |
//   |     M     |    5 |
//   |     L     |   10 |
//   +-----------+------+
//
// Table cells beginning with = are headers (usually placed in the first row)
//
// Cells are left- or right-justified if there is leading or trailing space between the |
// separators for that cell, respectively. If there is space before and after the text,
// it is centered. In the example above, the size codes will be centered in their column
// and the area numbers are right-justified in theirs.
//
// Cells which begin with a hyphen (-) indicate that the cell to their left spans into them.
// For example:
//   |=Column A|=Column B|=Column C
//   |stuff    |more stuff|and more
//   |a really wide column|- |hello
// produces:
//   +----------+------------+----------+
//   | COLUMN A |  COLUMN B  | COLUMN C |
//   +----------+------------+----------+
//   | stuff    | more stuff | and more |
//   | a really wide column  | hello    |
//   +----------+------------+----------+
//
// *May cross line boundaries but not paragraphs.
//
// †May nest as in //Italic **and** bold//.
//
// ‡Must appear at the very beginning of a line.
//
func Render(text string, opts ...renderOpts) (string, error) {
	ops := renderOptSet{
		formatter: &renderPlainTextFormatter{},
		bulletSet: []rune{0},
	}
	for _, o := range opts {
		o(&ops)
	}
	collapseSpaces := regexp.MustCompile(`\s{2,}`)
	newListBullet := regexp.MustCompile(`^[*#]+`)
	formatReqs := regexp.MustCompile(`//|\*\*|\[\[|\]\]|\\\.`)

	paragraphs := make([][][]interface{}, 0, 10)
	thisParagraph := make([][]interface{}, 0, 10)
	thisLine := make([]interface{}, 0, 10)

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
				thisLine = make([]interface{}, 0, 10)

				paragraphs = append(paragraphs, thisParagraph)
				thisParagraph = make([][]interface{}, 0, 10)
			}
			continue
		}
		//
		// Look for * or # at the start of the line. This begins
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
			thisLine = make([]interface{}, 0, 10)
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
	ops.formatter.init()

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
							fragment += " "
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
					currentTable.addRow(columns)

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

// @[00]@| GMA 4.3.2
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
