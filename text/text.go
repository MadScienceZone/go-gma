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

type renderOptSet struct {
	formatter renderingFormatter
}

type renderingFormatter interface {
	//enum_val(level, value, style)
	newPar()
	process(text string)
	finalize() string
	//toggleItal()
	setBold(on bool)
	setItal(on bool)
	newLine()
	table(*textTable)
	reference(displayName, linkName string)
	bulletListItem(level int)
	enumListItem(level, counter int)
}

//
// Plain Text output formatter
//
type renderPlainTextFormatter struct {
	buf    strings.Builder
	indent int
	ital   bool
	bold   bool
}

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

func (f *renderPlainTextFormatter) bulletListItem(level int) {
	fmt.Fprintf(&f.buf, "\n%*s*  ", level-1, "")
	f.indent = level
}

func (f *renderPlainTextFormatter) enumListItem(level, counter int) {
	fmt.Fprintf(&f.buf, "\n%*s%s. ", level-1, "", enumVal(level, counter))
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

type renderOpts func(*renderOptSet)

func AsPlainText(o *renderOptSet) {
	o.formatter = &renderPlainTextFormatter{}
}

type listItem struct {
	bullet rune
	level  int
}

type textTable struct {
	rows [][]*tableCell
}

func (t *textTable) addRow(cols []string) error {
	if t.rows == nil {
		t.rows = make([][]*tableCell, 0, 32)
	}
	row := make([]*tableCell, 0, 8)
	var lastUsed *tableCell = nil
	for _, col := range cols {
		if strings.HasSuffix(col, "-") {
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

type tableRow struct {
	src string
}

type tableCell struct {
	text   string
	span   int
	align  rune
	header bool
}

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
// Markup text rendering is done via the Render function.
// The set of options which may follow the string to be formatted
// include:
//   AsPlainText  -- render a text-only version of the input
//                   (this is the default)
//
func Render(text string, opts ...renderOpts) (string, error) {
	ops := renderOptSet{
		formatter: &renderPlainTextFormatter{},
	}
	for _, o := range opts {
		o(&ops)
	}
	collapseSpaces := regexp.MustCompile("\\s{2,}")
	newListBullet := regexp.MustCompile("^[*#]+")
	formatReqs := regexp.MustCompile("//|\\*\\*|\\[\\[|\\]\\]|\\\\\\.")

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
						ops.formatter.bulletListItem(fragment.level)
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
