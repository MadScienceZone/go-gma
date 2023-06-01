/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      ______     _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ____ \   (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (    \/   | (  )  |              #
# | |      | || || || (___) | Assistant | (____     | (____     | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \    |  ___ \    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )   | (   ) )   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ ( (___) ) _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_) \_____/ (_)(_______)              #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the util package
//

package util

import (
	"testing"
)

func TestHexDump(t *testing.T) {
	type testcase struct {
		in    []byte
		out   string
		out32 string
		out42 string
		outn  string
		outnt string
	}
	for i, test := range []testcase{
		{[]byte{}, "", "", "", "", ""},
		{[]byte{0},
			`00000000:  00                                               |.               |
`,
			`00000000:  00                                       |.               |
`,
			`0000002A:  00                                               |.               |
`,
			`00000000:  00              |.    |
`,
			`00000000:  00                                             
`},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			`00000000:  00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F  |................|
`,
			`00000000:  0001 0203 0405 0607 0809 0A0B 0C0D 0E0F  |................|
`,
			`0000002A:  00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F  |................|
`,
			`00000000:  00 01 02 03 04  |.....|
00000005:  05 06 07 08 09  |.....|
0000000A:  0A 0B 0C 0D 0E  |.....|
0000000F:  0F              |.    |
`,
			`00000000:  00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F
`},
		{[]byte("Hello, World"),
			`00000000:  48 65 6C 6C 6F 2C 20 57 6F 72 6C 64              |Hello, World    |
`,
			`00000000:  4865 6C6C 6F2C 2057 6F72 6C64            |Hello, World    |
`,
			`0000002A:  48 65 6C 6C 6F 2C 20 57 6F 72 6C 64              |Hello, World    |
`,
			`00000000:  48 65 6C 6C 6F  |Hello|
00000005:  2C 20 57 6F 72  |, Wor|
0000000A:  6C 64           |ld   |
`,
			`00000000:  48 65 6C 6C 6F 2C 20 57 6F 72 6C 64            
`},
	} {
		hx := Hexdump(test.in)
		hx32 := Hexdump(test.in, WithWordSize(2))
		hx42 := Hexdump(test.in, WithStartingAddress(42))
		hxn := Hexdump(test.in, WithWidth(5))
		hxnt := Hexdump(test.in, WithoutText)

		if hx != test.out {
			t.Errorf("Case %d: expected:\n%sactual:\n%s", i, test.out, hx)
		}
		if hx32 != test.out32 {
			t.Errorf("Case %d/32: expected:\n%sactual:\n%s", i, test.out32, hx32)
		}
		if hx42 != test.out42 {
			t.Errorf("Case %d/42: expected:\n%sactual:\n%s", i, test.out42, hx42)
		}
		if hxn != test.outn {
			t.Errorf("Case %d/n: expected:\n%sactual:\n%s", i, test.outn, hxn)
		}
		if hxnt != test.outnt {
			t.Errorf("Case %d/nt: expected:\n%sactual:\n%s", i, test.outnt, hxnt)
		}
	}
}

func TestVersions(t *testing.T) {
	const (
		lessThan = iota
		moreThan
		equalTo
	)

	type testcase struct {
		a string
		b string
		c int
		e bool
	}

	for i, test := range []testcase{
		{"0", "0", equalTo, false},
		{"0", "0.0", equalTo, false},
		{"0.0", "0.0", equalTo, false},
		{"0.0.0", "0.0", equalTo, false},
		{"0.0.0", "0.0.0", equalTo, false},
		{"0", "0.1", lessThan, false},
		{"0.0", "0.1", lessThan, false},
		{"0.0.0", "0.1", lessThan, false},
		{"0.0.0", "0.0.1", lessThan, false},
		{"1.2.3.4.5", "1.2.3.4.5", equalTo, false},
		{"1.2.3.4.5", "1.2.3.4", moreThan, false},
		{"x.y", "1,0", equalTo, true},
		{"1.2", "1.2.0", equalTo, false},
		{"1.21", "1.3", moreThan, false},
		{"1.21", "1.21-alpha", moreThan, false},
		{"1.21-beta.1", "1.21-beta.2", lessThan, false},
		{"1.21+ABCDE-beta.1", "1.21+F8432-beta.1", equalTo, false},
	} {
		cmp, err := VersionCompare(test.a, test.b)
		if err != nil && !test.e {
			t.Errorf("case %d, \"%s\" vs \"%s\": error %v unexpected", i, test.a, test.b, err)
		}
		if err == nil && test.e {
			t.Errorf("case %d, \"%s\" vs \"%s\": error unexpected but not found", i, test.a, test.b)
		}
		if test.c == lessThan && cmp >= 0 {
			t.Errorf("case %d, \"%s\" vs \"%s\": should be <0 but got %d", i, test.a, test.b, cmp)
		} else if test.c == moreThan && cmp <= 0 {
			t.Errorf("case %d, \"%s\" vs \"%s\": should be >0 but got %d", i, test.a, test.b, cmp)
		} else if test.c == equalTo && cmp != 0 {
			t.Errorf("case %d, \"%s\" vs \"%s\": should be 0 but got %d", i, test.a, test.b, cmp)
		}
	}
}

// @[00]@| Go-GMA 5.6.0
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
