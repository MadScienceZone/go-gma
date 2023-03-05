/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______     _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )   (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |   | (  )  |              #
# | |      | || || || (___) | Assistant | (____         /   )   | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )    /   _/     |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/(_)(_______)              #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the tcllist type
//

package tcllist

import (
	"reflect"
	"testing"
)

// TCL LIST FORMAT
//        In  a  nutshell,  a  Tcl  list (as a string representation) is a space-
//        delimited list of values. Any value which includes spaces  is  enclosed
//        in  curly braces.  An empty string (empty list) as an element in a list
//        is represented as “{}”.  (E.g., “1 {} 2” is a list of  three  elements,
//        the  middle of which is an empty string.) An entirely empty Tcl list is
//        represented as an empty string “”.
//
//        A list value must have balanced braces. A balanced pair of braces  that
//        happen  to  be  inside a larger string value may be left as-is, since a
//        string that happens to contain spaces or braces is  only  distinguished
//        from a deeply-nested list value when you attempt to interpret it as one
//        or another in the code. Thus, the list
//               “a b {this {is a} string}”
//        has three elements: “a”, “b”, and “this {is a} string”.   Otherwise,  a
//        lone brace that's part of a string value should be escaped with a back‐
//        slash:
//               “a b {this \{ too}”
//
//        Literal backslashes may be escaped with a backslash as well.
//
//        While extra spaces are ignored when  parsing  lists  into  elements,  a
//        properly  formed  string representation of a list will have the miminum
//        number of spaces and braces needed to describe the list structure.
//
// 		More examples
//			a b c d		    [a, b, c, d]
//			a {b c} d	    [a, "b c", d]			) depending on how you interpret
//			a {b c} d	    [a, [b, c], d]			) the answer
//			a b {{c d} e f} [a, b, "{c d} e f"]
//			a b {{c d} e f} [a, b, ["c d", e, f]]
//			a b {{c d} e f} [a, b, [[c, d], e, f]]

func TestTclList_Str2list(t *testing.T) {
	type testcase struct {
		tcl     string
		list    []string
		isError bool
	}

	tests := []testcase{
		{tcl: "a b c d", list: []string{"a", "b", "c", "d"}, isError: false},
		{tcl: "a  b c d", list: []string{"a", "b", "c", "d"}, isError: false},
		{tcl: "   a  b  c  d    ", list: []string{"a", "b", "c", "d"}, isError: false},
		{tcl: "a {b  c} d", list: []string{"a", "b  c", "d"}, isError: false},
		{tcl: "a {b  c}x d", list: []string{"a", "b  c", "d"}, isError: true},
		{tcl: "a {b c}{def} x d", list: []string{"a", "b  c", "d"}, isError: true},
		{tcl: "a b {{c d} e f}", list: []string{"a", "b", "{c d} e f"}, isError: false},
		{tcl: "a b {c d} e f}", list: []string{"a", "b", "{c d} e f"}, isError: true},
		{tcl: "a b{cd d} e f}", list: []string{"a", "b", "{c d} e f"}, isError: true},
		{tcl: "a b{cd d} e f", list: []string{"a", "b{cd", "d}", "e", "f"}, isError: false},
		{tcl: "a b{cd d}e e f", list: []string{"a", "b{cd", "d}e", "e", "f"}, isError: false},
		{tcl: "a b{cd d}}e e f", list: []string{"a", "b{cd", "d}e", "e", "f"}, isError: true},
		{tcl: "a b{cd d}{e e f", list: []string{"a", "b{cd", "d}e", "e", "f"}, isError: true},
		{tcl: "               ", list: []string{}, isError: false},
		{tcl: "", list: []string{}, isError: false},
		{tcl: "1 2 \"\" 5", list: []string{"1", "2", "", "5"}, isError: false},
		{tcl: "a \"b  c\" d", list: []string{"a", "b  c", "d"}, isError: false},
		{tcl: "a \"b  c\"x d", list: []string{"a", "b  c", "d"}, isError: true},
		{tcl: "a \"b c\"\"def\" x d", list: []string{"a", "b  c", "d"}, isError: true},
		{tcl: "a b \"{c d} e f\"", list: []string{"a", "b", "{c d} e f"}, isError: false},
		{tcl: "a b \"c d\" e f}", list: []string{"a", "b", "{c d} e f"}, isError: true},
		{tcl: "a b\"cd d\" e f}", list: []string{"a", "b", "{c d} e f"}, isError: true},
		{tcl: "a b\"cd d\" e f", list: []string{"a", "b\"cd", "d\"", "e", "f"}, isError: false},
		{tcl: "a b\"cd d\"e e f", list: []string{"a", "b\"cd", "d\"e", "e", "f"}, isError: false},
		{tcl: "1 2 {} 5", list: []string{"1", "2", "", "5"}, isError: false},
		{tcl: "spam eggs", list: []string{"spam", "eggs"}, isError: false},
		{tcl: "penguin {spam spam}", list: []string{"penguin", "spam spam"}, isError: false},
		{tcl: "penguin \\{spam spam}", list: []string{"penguin", "{spam", "spam}"}, isError: true},
		{tcl: "penguin \\{spam spam\\}", list: []string{"penguin", "{spam", "spam}"}, isError: false},
		{tcl: "aa \\{\\\"bb\\}cc dd", list: []string{"aa", "{\"bb}cc", "dd"}, isError: false},
		{tcl: "\\#aa bb dd", list: []string{"#aa", "bb", "dd"}, isError: false},
		{tcl: "\\#aa bb dd\\", list: []string{"#aa", "bb", "dd"}, isError: true},
		{tcl: "a b {this {is a} string}", list: []string{"a", "b", "this {is a} string"}, isError: false},
		{tcl: "a b {this \\{ too}", list: []string{"a", "b", "this { too"}, isError: false},
		{tcl: "a b this\\ \\{\\ too", list: []string{"a", "b", "this { too"}, isError: false},
		{tcl: "^\\$\\[.*\\]", list: []string{"^\\$\\[.*\\]"}, isError: false},
	}

	for _, test := range tests {
		t.Run("parse tests", func(t *testing.T) {
			l, err := ParseTclList(test.tcl)
			if test.isError {
				if err == nil {
					t.Fatalf("TCL \"%s\" was supposed to return an error but didn't.", test.tcl)
				}
			} else {
				if err != nil {
					t.Fatalf("TCL \"%s\" caused error \"%v\"", test.tcl, err)
				}
				if !reflect.DeepEqual(l, test.list) {
					t.Fatalf("TCL \"%s\" -> %v, expected %v",
						test.tcl, l, test.list)
				}
			}
		})
	}
}

func TestTclList_List2str(t *testing.T) {
	type testcase struct {
		tcl     string
		list    []string
		isError bool
	}

	tests := []testcase{
		{tcl: "a b c d", list: []string{"a", "b", "c", "d"}, isError: false},
		{tcl: "a {b  c} d", list: []string{"a", "b  c", "d"}, isError: false},
		{tcl: "a b {{c d} e f}", list: []string{"a", "b", "{c d} e f"}, isError: false},
		{tcl: "a b\\{cd d\\} e f", list: []string{"a", "b{cd", "d}", "e", "f"}, isError: false},
		{tcl: "a b\\{cd d\\}e e f", list: []string{"a", "b{cd", "d}e", "e", "f"}, isError: false},
		{tcl: "", list: []string{}, isError: false},
		{tcl: "{}", list: []string{""}, isError: false},
		{tcl: "1 2 {} 5", list: []string{"1", "2", "", "5"}, isError: false},
		{tcl: "spam eggs", list: []string{"spam", "eggs"}, isError: false},
		{tcl: "penguin {spam spam}", list: []string{"penguin", "spam spam"}, isError: false},
		{tcl: "penguin \\{spam spam\\}", list: []string{"penguin", "{spam", "spam}"}, isError: false},
		{tcl: "aa {{\"bb}cc} dd", list: []string{"aa", "{\"bb}cc", "dd"}, isError: false},
		{tcl: "{#aa} bb dd", list: []string{"#aa", "bb", "dd"}, isError: false},
		{tcl: "a b {this {is a} string}", list: []string{"a", "b", "this {is a} string"}, isError: false},
		{tcl: "a b this\\ \\{\\ too", list: []string{"a", "b", "this { too"}, isError: false},
	}

	for _, test := range tests {
		t.Run("emit tests", func(t *testing.T) {
			s, err := ToTclString(test.list)
			if test.isError {
				if err == nil {
					t.Fatalf("List %v was supposed to return an error but didn't.", test.list)
				}
			} else {
				if err != nil {
					t.Fatalf("list %v caused error \"%v\"", test.list, err)
				}
				if s != test.tcl {
					t.Errorf("List %v -> \"%s\", expected \"%s\"",
						test.list, s, test.tcl)
				}
				l, err := ParseTclList(test.tcl)
				if err != nil {
					t.Fatalf("List %v -> TCL \"%s\" -> List caused error \"%v\"", test.list, s, err)
				}
				if !reflect.DeepEqual(l, test.list) {
					t.Errorf("List %v -> TCL \"%s\" -> %v",
						test.list, s, l)
				}
			}
		})
	}
}

func TestTclList_ConvertTypes(t *testing.T) {
	type testcase struct {
		types   string
		isError bool
	}

	src := []string{"abc", "def", "1", "12.32", "-2a"}

	b, err := ConvertTypes(src, "")
	if err == nil {
		t.Fatalf("Expected error from empty types")
	}

	b, err = ConvertTypes(src, "ssss")
	if err == nil {
		t.Fatalf("Expected error from too few types")
	}

	b, err = ConvertTypes(src, "ssssssss")
	if err == nil {
		t.Fatalf("Expected error from too many types")
	}

	b, err = ConvertTypes(src, "sisss")
	if err == nil {
		t.Fatalf("Expected error from wrong type")
	}

	b, err = ConvertTypes(src, "brsss")
	if err != nil {
		t.Fatalf("Error converting brsss: %v", err)
	}
	_, ok := b[0].([]byte)
	if !ok {
		t.Fatalf("Error converting brsss: [0] not bytes")
	}
	_, ok = b[1].([]rune)
	if !ok {
		t.Fatalf("Error converting brsss: [1] not runes")
	}
	_, ok = b[2].(string)
	if !ok {
		t.Fatalf("Error converting brsss: [2] not string")
	}
	_, ok = b[3].(string)
	if !ok {
		t.Fatalf("Error converting brsss: [3] not string")
	}
	_, ok = b[4].(string)
	if !ok {
		t.Fatalf("Error converting brsss: [4] not string")
	}

	b, err = ConvertTypes(src, "ssiii")
	if err == nil {
		t.Fatalf("Error converting ssiii: expected error")
	}

	b, err = ConvertTypes(src, "ssffi")
	if err == nil {
		t.Fatalf("Error converting ssffi: expected error")
	}

	b, err = ConvertTypes(src, "ssff*")
	if err != nil {
		t.Fatalf("Error converting ssff*: %v", err)
	}
	_, ok = b[0].(string)
	if !ok {
		t.Fatalf("Error converting ssff*: [0] not string")
	}
	_, ok = b[1].(string)
	if !ok {
		t.Fatalf("Error converting ssff*: [1] not string")
	}
	_, ok = b[2].(float64)
	if !ok {
		t.Fatalf("Error converting ssff*: [2] not float64")
	}
	_, ok = b[3].(float64)
	if !ok {
		t.Fatalf("Error converting ssff*: [3] not float64")
	}
	if b[4] != nil {
		t.Fatalf("Error converting ssff*: expected [4] to be nil")
	}

	b, err = ConvertTypes([]string{"foo", "bar"}, "ss*")
	if err != nil {
		t.Fatalf("Error when trailing * in conversion set: %v", err)
	}

	b, err = ConvertTypes(src, "ssifs")
	if err != nil {
		t.Fatalf("Error converting ssifs: %v", err)
	}
	_, ok = b[0].(string)
	if !ok {
		t.Fatalf("Error converting ssifs: [0] not string")
	}
	_, ok = b[1].(string)
	if !ok {
		t.Fatalf("Error converting ssifs: [1] not string")
	}
	_, ok = b[2].(int)
	if !ok {
		t.Fatalf("Error converting ssifs: [2] not int")
	}
	_, ok = b[3].(float64)
	if !ok {
		t.Fatalf("Error converting ssifs: [3] not float64")
	}
	_, ok = b[4].(string)
	if !ok {
		t.Fatalf("Error converting ssifs: [4] not string")
	}

}

func TestTclList_Deep(t *testing.T) {
	s, err := ToDeepTclString("hello", "world", 123, 45.67, []any{"a", "", "b", []string{"xyz", "333"}, 1, true}, false)

	if err != nil {
		t.Errorf("%v", err)
	} else if s != "hello world 123 45.67 {a {} b {xyz 333} 1 1} 0" {
		t.Errorf("deep string error, returned \"%v\"", s)
	}
}

// @[00]@| GMA 5.2.0
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
//
