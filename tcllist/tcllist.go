/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      _____      _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ___ \    (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (___) )   | (  )  |              #
# | |      | || || || (___) | Assistant | (____      \     /    | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \     / ___ \    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )   ( (   ) )   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ ( (___) ) _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_) \_____/ (_)(_______)              #
#                                                                                      #
########################################################################################
*/
//
////////////////////////////////////////////////////////////////////////////////////////
//                                                                                    //
//                                     TclList                                        //
//                                                                                    //
// Represents a list of values in such a way that we can interact with them as a      //
// list of values in Go, but can translate them to and from TCL list syntax.          //
//                                                                                    //
////////////////////////////////////////////////////////////////////////////////////////

// In the Python GMA library version of this code, we take advantage of the fact that
// we have a Tcl interpreter available in tkinter and use that to process
// Tcl strings.
//
// In our case, we don't have a Tcl interpreter handy,  so we'll implement
// a simple string scanner in Go which will convert these string representations to and from Go slices.

//
// Package tcllist converts between Tcl list strings and Go slices.
//
// Some of the older elements of GMA (which used to be entirely written in
// the Tcl language, after it was ported from the even older C++ code) use
// Tcl  list  objects  as  their  data representation. Notably the biggest
// example is the mapper(6) tool (which is still written in  Tcl  itself),
// whose  map  file  format and TCP/IP communications protocol include marshalled data structures represented as Tcl lists.
//
// While this is obviously convenient for Tcl programs in  that  they  can
// take  such strings and natively use them as lists of values, it is also
// useful generally in that it is a simple string representation of a simple
// data  structure.  The  actual  definition of this string format is
// included below.
//
// The tcllist Go package provides an easy interface to manipulate
// Tcl  lists  as  Go  types.
//
// TCL LIST FORMAT
//
// In  a  nutshell,  a  Tcl  list (as a string representation) is a space-delimited
// list of values. Any value which includes spaces  is  enclosed
// in  curly braces.  An empty string (empty list) as an element in a list
// is represented as “{}”.  (E.g., “1 {} 2” is a list of  three  elements,
// the  middle of which is an empty string.) An entirely empty Tcl list is
// represented as an empty string “”.
//
// A list value must have balanced braces. A balanced pair of braces  that
// happen  to  be  inside a larger string value may be left as-is, since a
// string that happens to contain spaces or braces is  only  distinguished
// from a deeply-nested list value when you attempt to interpret it as one
// or another in the code. Thus, the list
// 	  “a b {this {is a} string}”
// has three elements: “a”, “b”, and “this {is a} string”.   Otherwise,  a
// lone brace that's part of a string value should be escaped with a backslash:
// 	  “a b {this \{ too}”
//
// Literal backslashes may be escaped with a backslash as well.
//
// While extra spaces are ignored when  parsing  lists  into  elements,  a
// properly  formed  string representation of a list will have the minimum
// number of spaces and braces needed to describe the list structure.
package tcllist

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

//--------------------------------------------------------------------------------------------------------------
//
// Rather than imperfectly try to mimic the behavior of Tcl list
// code (and hope we didn't miss some nuance), the following code is
// our port of the code from the Tcl sources, which is itself released
// under the following terms:
//
//  _______________________________________________________________________
// |  The following terms apply to the all versions of the core Tcl/Tk     |
// |  releases, the Tcl/Tk browser plug-in version 2.0, and TclBlend       |
//    and Jacl version 1.0. Please note that the TclPro tools are under
//    a different license agreement. This agreement is part of the
//    standard Tcl/Tk distribution as the file named "license.terms".
//
//    Tcl/Tk License Terms
//
//    This software is copyrighted by the Regents of the University of
//    California, Sun Microsystems, Inc., Scriptics Corporation, and
//    other parties. The following terms apply to all files associated
//    with the software unless explicitly disclaimed in individual files.
//
//    The authors hereby grant permission to use, copy, modify, distribute,
//    and license this software and its documentation for any purpose,
//    provided that existing copyright notices are retained in all copies
//    and that this notice is included verbatim in any distributions. No
//    written agreement, license, or royalty fee is required for any of
//    the authorized uses. Modifications to this software may be
//    copyrighted by their authors and need not follow the licensing
//    terms described here, provided that the new terms are clearly
//    indicated on the first page of each file where they apply.
//
//    IN NO EVENT SHALL THE AUTHORS OR DISTRIBUTORS BE LIABLE TO ANY
//    PARTY FOR DIRECT, INDIRECT, SPECIAL, INCIDENTAL, OR CONSEQUENTIAL
//    DAMAGES ARISING OUT OF THE USE OF THIS SOFTWARE, ITS DOCUMENTATION,
//    OR ANY DERIVATIVES THEREOF, EVEN IF THE AUTHORS HAVE BEEN ADVISED
//    OF THE POSSIBILITY OF SUCH DAMAGE.
//
//    THE AUTHORS AND DISTRIBUTORS SPECIFICALLY DISCLAIM ANY WARRANTIES,
//    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY,
//    FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. THIS SOFTWARE
//    IS PROVIDED ON AN "AS IS" BASIS, AND THE AUTHORS AND DISTRIBUTORS HAVE
//    NO OBLIGATION TO PROVIDE MAINTENANCE, SUPPORT, UPDATES, ENHANCEMENTS,
//    OR MODIFICATIONS.
//
//    GOVERNMENT USE: If you are acquiring this software on behalf of the
//    U.S. government, the Government shall have only "Restricted Rights"
//    in the software and related documentation as defined in the Federal
//    Acquisition Regulations (FARs) in Clause 52.227.19 (c) (2). If you
//    are acquiring the software on behalf of the Department of Defense,
//    the software shall be classified as "Commercial Computer Software"
//    and the Government shall have only "Restricted Rights" as defined
//    in Clause 252.227-7013 (c) (1) of DFARs. Notwithstanding the foregoing,
//    the authors grant the U.S. Government and others acting in its behalf
//    permission to use and distribute the software in accordance with the
//  | terms specified in this license.                                    |
//  |_____________________________________________________________________|
//
//
// The following code was written for GMA by Steven Willoughby, based on the
// original C code distributed in the Tcl core source code files "tclUtil.c",
// "tclParse.c", "tclUtf.c", as a direct port of that original code to Go.
//
const tConvertNone = 0
const tTclDontUseBraces = 1
const tConvertBrace = 2
const tConvertEscape = 4
const tConvertMask = (tConvertBrace | tConvertEscape)
const tTclDontQuoteHash = 8
const tConvertAny = 16

//
// Tcl_ScanElement(string, flagPtr) -> len
// scans the input string, setting flags based on what's in that element
// and returns the string length needed to hold the string representation
// of that element (an overestimation for allocation purposes)
//
func tclScanElement(element string, flags int) (int, int, error) {
	length := len(element)
	nestingLevel := 0
	forbidNone := false
	requireEscape := false
	extra := 0
	bytesNeeded := 0
	preferEscape := false
	preferBrace := false
	braceCount := 0
	afterBackslash := false

	if length == 0 {
		return 2, tConvertBrace, nil
	}

	if element[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
		preferBrace = true
	}

	if element[0] == '{' || element[0] == '"' {
		forbidNone = true
		preferBrace = true
	}

	for i, r := range element {
		if afterBackslash {
			if r == '{' || r == '}' || r == '\\' {
				extra++
			} else if r == '\n' {
				extra++
				requireEscape = true
			}
			afterBackslash = false
			continue
		}

		switch r {
		case '{':
			braceCount++
			extra++
			nestingLevel++
		case '}':
			extra++
			nestingLevel--
			if nestingLevel < 0 {
				requireEscape = true
			}
		case ']', '"':
			forbidNone = true
			extra++
			preferEscape = true
		case '[', '$', ';', ' ', '\f', '\n', '\r', '\t', '\v':
			forbidNone = true
			extra++
			preferBrace = true
		case '\\':
			extra++
			if i+1 >= length {
				requireEscape = true
				break
			}
			afterBackslash = true
			forbidNone = true
			preferBrace = true
		}
	}
	if nestingLevel != 0 {
		requireEscape = true
	}
	bytesNeeded = length
	if requireEscape {
		bytesNeeded += extra
		if element[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
			bytesNeeded++
		}
		flags = tConvertEscape
		goto overflow_check
	}

	if (flags & tConvertAny) != 0 {
		if extra < 2 {
			extra = 2
		}
		flags &= ^tConvertAny
		flags |= tTclDontUseBraces
	}

	if forbidNone {
		if preferEscape && !preferBrace {
			bytesNeeded += (extra - braceCount)
			if element[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
				bytesNeeded++
			}
			if (flags & tTclDontUseBraces) != 0 {
				bytesNeeded += braceCount
			}
			flags = tConvertMask
			goto overflow_check
		}
		if (flags & tTclDontUseBraces) != 0 {
			bytesNeeded += extra
			if element[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
				bytesNeeded++
			}
		} else {
			bytesNeeded += 2
		}
		flags = tConvertBrace
		goto overflow_check
	}

	if element[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
		bytesNeeded += 2
	}
	flags = tConvertNone

overflow_check:
	if bytesNeeded < 0 {
		return 0, 0, fmt.Errorf("string length overflow")
	}
	return bytesNeeded, flags, nil
}

func tclConvertElement(src string, flags int) string {
	conversion := flags & tConvertMask
	var p strings.Builder

	if (flags&tTclDontUseBraces) != 0 && (conversion&tConvertBrace) != 0 {
		conversion = tConvertEscape
	}
	if len(src) == 0 {
		conversion = tConvertBrace
	} else {
		if src[0] == '#' && (flags&tTclDontQuoteHash) == 0 {
			if conversion == tConvertEscape {
				p.WriteRune('\\')
				//p.WriteRune('#') we'll write this later
			} else {
				conversion = tConvertBrace
			}
		}
	}

	if conversion == tConvertNone {
		p.WriteString(src)
		return p.String()
	}

	if conversion == tConvertBrace {
		p.WriteRune('{')
		p.WriteString(src)
		p.WriteRune('}')
		return p.String()
	}

	for _, r := range src {
		switch r {
		case ']', '[', '$', ';', ' ', '\\', '"':
			p.WriteRune('\\')
		case '{', '}':
			if conversion == tConvertEscape {
				p.WriteRune('\\')
			}
		case '\f':
			p.WriteRune('\\')
			p.WriteRune('f')
		case '\n':
			p.WriteRune('\\')
			p.WriteRune('n')
		case '\r':
			p.WriteRune('\\')
			p.WriteRune('r')
		case '\t':
			p.WriteRune('\\')
			p.WriteRune('t')
		case '\v':
			p.WriteRune('\\')
			p.WriteRune('v')
		}
		p.WriteRune(r)
	}
	return p.String()
}

//
// END of ported Tcl core code.____________________________________________________________
//

// By contrast, the code to go the other direction (parsing Tcl strings as slices)
// is all original but seems to work fine.
//
// rules about braces:
//	{ inside a string doesn't count (but still needs to be balanced)
//	} that ends a braced list can't be followed by trailing characters
//
// more formally:
// If the first character of an element is {/", then the element ends with the matching }/"
// The ending } MUST be present and MUST not be followed by any non-space runes.
//
// TCL strings generated here SHOULD also conform to the following:
//  no newlines between elements
//  no unescaped ;,] except in quotes/braces
//  no unescaped $,[,\ except in braces
//  no unescaped # as first character of first element except in quotes/braces
//  don't put \<newline> in the string. Use \\\<newline>.

// ToTclString takes a slice of strings and outputs a single string value
// which represents that slice as a valid Tcl list. This function may return
// an error, but as currently implemented that should rarely happen (it is
// triggered by a string whose length is too long to fit in an integer),
// but there may be other error conditions added in the future, so check it
// anyway).
//
// It returns the Tcl string and the error, if any.
func ToTclString(listval []string) (string, error) {
	var s strings.Builder

	for elementIdx, element := range listval {
		if elementIdx > 0 {
			s.WriteRune(' ')
		}
		flags := 0
		if elementIdx > 0 {
			flags = tTclDontQuoteHash
		}
		_, flags, err := tclScanElement(element, flags)
		if err != nil {
			return "", err
		}
		s.WriteString(tclConvertElement(element, flags))
	}
	return s.String(), nil
}

// ParseTclList takes a properly-formatted Tcl list string
// and returns a slice of the list's elements as string values
// as well as an error (if something went wrong).
//
// Note that this only parses a single nesting level of elements,
// since with Tcl lists it is impossible to distinguish an element
// which happens to contain spaces from a nested list of values. It
// is simply up to the program to use the element as a string or as
// a sublist, so in the latter case you'll need to call ParseTclList
// on that element.
func ParseTclList(tclString string) ([]string, error) {
	l := make([]string, 0, 10)
	level := 0
	var s strings.Builder

	betweenElements := true
	endOfElement := false
	bracedString := false
	literalNext := false
	quotedString := false

	for _, r := range tclString {
		// step through the string representation of the list,
		// handling \{, \}, and \\ escapes as well as multiple
		// spaces between elements
		// We check for this specific set of whitespace characters
		// instead of unicode.IsSpace because the Tcl spec says so.
		if literalNext {
			if r != '{' && r != '}' && r != '\\' && r != '"' && r != ' ' && r != '#' {
				_, _ = s.WriteRune('\\')
			}
			_, _ = s.WriteRune(r)
			literalNext = false
			continue
		}
		if r == '\\' {
			literalNext = true
			continue
		}
		if !bracedString && !quotedString && strings.ContainsRune(" \t\n\v\f\r", r) {
			if betweenElements {
				continue // skip over multiple (superfluous) spaces
			}
			if endOfElement {
				endOfElement = false // past that now
			}
			// not between elements? ship out what we were collecting
			// and start a new one
			l = append(l, s.String())
			s.Reset()
			betweenElements = true
			bracedString = false
			quotedString = false
		} else {
			if endOfElement {
				// we got superfluous text after a closing brace
				return l, fmt.Errorf("list element in braces (\"%s\") followed by '%c' instead of space", s.String(), r)
			}
			if r == '"' {
				if betweenElements {
					// Quotes are just like braces except that they
					// can't really nest.
					betweenElements = false
					quotedString = true
					continue
				} else if quotedString {
					quotedString = false
					endOfElement = true
					continue
				}
			} else if r == '{' {
				level++
				if betweenElements {
					// this is the brace that starts a string which
					// means it may allow leading spaces in the value
					// so let's not retain the brace but stop looking
					// for the next element
					betweenElements = false
					bracedString = true
					continue
				}
			} else if r == '}' {
				level--
				if level == 0 {
					if bracedString {
						// this should be the end of the string
						endOfElement = true
						bracedString = false
						continue // and don't keep the brace
					}
				}
				if level < 0 {
					return l, fmt.Errorf("too many right braces after \"%s\"", s.String())
				}
			}
			if betweenElements {
				betweenElements = false // we're in an element now
			}
			_, _ = s.WriteRune(r)
		}
	}
	if !betweenElements {
		l = append(l, s.String())
	}
	if level != 0 {
		return l, fmt.Errorf("unterminated brace at end of string")
	}
	if quotedString {
		return l, fmt.Errorf("unterminated quote at end of string")
	}
	if literalNext {
		return l, fmt.Errorf("trailing backslash at end of string")
	}
	return l, nil
}

//
// ConvertTypes converts some or all of the elements in a string slice
// such as that returned by ParseTclList to a new slice of values
// which have been converted to other data types as specified by the
// caller.
//
// Since the string representation of a TclList is type-agnostic,
// we can't automatically
// assume its elements should be specific types such as numeric values,
// hence the need for this function to provide the missing type information
// to complete the conversions.
//
// The types string controls this conversion. Each character indicates
// the required type for the corresponding element in the input slice, as follows:
//    "-"  do not convert this element.
//    "s"  copy the element as a string.
//    "b"  copy the element as a []byte slice.
//    "r"  copy the element as a []rune slice.
//    "f"  convert the element to a float value.
//    "i"  convert the element to an int value.
//    "I"  as i, but an empty string is equivalent to 0.
//    "?"  copy the element as a bool value.
//    "*"  stop processing here, ignoring any remaining slice elements.
// If the value cannot be converted as requested, an error is returned.
//
// This provides a simple way to validate the types for all values in
// the slice at once, so you can directly access the slice elements as the
// intended types without individually type-testing each one.
//
// The input slice must have exactly the number of elements as characters
// in the type string unless the * character is used in types, so this
// function also enforces that the expected number of data elements is present.
//
func ConvertTypes(list []string, types string) ([]any, error) {
	converted := make([]any, len(list))
	var err error

	if len(types) > len(list) && types[len(list)] != '*' {
		return nil, fmt.Errorf("too many data elements (%d) for expected values", len(list))
	}
	for i, s := range list {
		if len(types) <= i {
			return nil, fmt.Errorf("Not enough type specifiers for %d-element slice", len(list))
		}
		switch types[i] {
		case '-':
			continue
		case 's':
			converted[i] = s
		case 'b':
			converted[i] = []byte(s)
		case '?':
			converted[i], err = strconv.ParseBool(s)
		case 'r':
			converted[i] = []rune(s)
		case 'f':
			converted[i], err = strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
		case 'I':
			if strings.TrimSpace(s) == "" {
				converted[i] = 0
				continue
			}
			fallthrough
		case 'i':
			converted[i], err = strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
		case '*':
			return converted, nil
		default:
			return nil, fmt.Errorf("Invalid type specifier '%v'", types[i])
		}
	}
	return converted, nil
}

// Parse is a convenience function which combines the operation of the
// ParseTclList and ConvertTypes functions in a single step.
func Parse(tclString, types string) ([]any, error) {
	f, err := ParseTclList(tclString)
	if err != nil {
		return nil, err
	}

	return ConvertTypes(f, types)
}

//
// ToDeepTclString takes a number of arbitrarily-typed values and returns
// a Tcl string which represents them as elements of a list.
// Supports values of type
// bool,
// float64,
// int,
// int16,
// int32,
// int64,
// string,
// uint,
// uint16,
// uint32,
// uint64,
// and slices of any combination of the above.
//
// For example, ToDeepTclString("a", 12, 13.42, []string{"b", "c"})
// returns the string "a 12 13.42 {b c}".
//
func ToDeepTclString(values ...any) (string, error) {
	var list []string

	for _, value := range values {
		switch v := value.(type) {
		case bool:
			if v {
				list = append(list, "1")
			} else {
				list = append(list, "0")
			}
		case string:
			list = append(list, v)
		case float64:
			list = append(list, fmt.Sprintf("%g", v))
		case int:
			list = append(list, fmt.Sprintf("%d", v))
		case int16:
			list = append(list, fmt.Sprintf("%d", v))
		case int32:
			list = append(list, fmt.Sprintf("%d", v))
		case int64:
			list = append(list, fmt.Sprintf("%d", v))
		case uint:
			list = append(list, fmt.Sprintf("%d", v))
		case uint16:
			list = append(list, fmt.Sprintf("%d", v))
		case uint32:
			list = append(list, fmt.Sprintf("%d", v))
		case uint64:
			list = append(list, fmt.Sprintf("%d", v))

		case []string:
			sublist, err := ToTclString(v)
			if err != nil {
				return "", nil
			}
			list = append(list, sublist)

		default:
			vof := reflect.ValueOf(value)
			if vof.Kind() == reflect.Slice {
				subslice := make([]any, vof.Len())
				for i := 0; i < vof.Len(); i++ {
					subslice[i] = vof.Index(i).Interface()
				}
				sublist, err := ToDeepTclString(subslice...)
				if err != nil {
					return "", nil
				}
				list = append(list, sublist)
			} else {
				return "", fmt.Errorf("ToDeepTclString: unsupported value type \"%T\"", value)
			}
		}
	}
	return ToTclString(list)
}

//
// StripLevel strips away the outermost level of {} characters from a string.
// The string must begin and end with { and } characters respectively.
//
func StripLevel(s string) string {
	if len(s) > 1 && s[0] == '{' && s[len(s)-1] == '}' {
		return s[1 : len(s)-1]
	}
	return s
}

// @[00]@| Go-GMA 5.8.0
// @[01]@|
// @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved.
// @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[14]@| License as described in the accompanying LICENSE file distributed
// @[15]@| with GMA.
// @[16]@|
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
