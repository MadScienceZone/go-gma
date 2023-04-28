/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______       __                   #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \     /  \                  #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \    \/) )                 #
# | |      | || || || (___) | Assistant | (____        ___) /      | |                 #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (       | |                 #
# | | \_  )| |   | || (   ) |                 ) )         ) \      | |                 #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\___/  / _  __) (_                #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ (_) \____/                #
#                                                                                      #
########################################################################################
*/

//
// Package util provides miscellaneous utility functions that don't deserve
// their own package.
//
package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
)

func splitToInts(s string) ([]int, error) {
	al := make([]int, 3)

	for _, v := range strings.Split(s, ".") {
		vv, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		al = append(al, vv)
	}
	return al, nil
}

//
// VersionCompare compares version strings a and b. These strings must consist of
// integers separated with dots, such as "2" or "3.1".
// Any number of version levels are allowed, although generally
// only 2 or 3 are of practical use.
//
// Returns <0 if a is a version before b,
// >0 if a is after b, or zero if they are the same.
//
// As of version 5.0.0, this is simply a wrapper to the hashicorp go-version package.
//
func VersionCompare(a, b string) (int, error) {
	va, err := version.NewSemver(a)
	if err != nil {
		return 0, err
	}

	vb, err := version.NewSemver(b)
	if err != nil {
		return 0, err
	}

	return va.Compare(vb), nil
}

//
// Return a string representation of an ASCII hexdump of the data.
//
type hdopt struct {
	addr  int
	width int
	word  int
	ascii bool
}

//
// Hexdump takes an array of bytes and returns a multi-line string
// representing those bytes in a traditional hexdump format with
// an address field on the left, starting at address 0, showing 16
// bytes per line, and a text bar along the right showing any printable
// ASCII characters found in the hexdump.
//
// For example, calling
//  Hexdump([]byte("\x00\x81\x02\x03Hello, World™<>ABCDEFG"))
// will return the string
//  00000000:  00 81 02 03 48 65 6C 6C 6F 2C 20 57 6F 72 6C 64  |....Hello, World|
//  00000010:  E2 84 A2 3C 3E 41 42 43 44 45 46 47              |...<>ABCDEFG    |
//
// Options may be added after the data slice to control how the hex dump
// will be formatted:
// WithStartingAddress(addr),
// WithWidth(nbytes),
// WithWordSize(nbytes),
// and/or
// WithoutText.
//
func Hexdump(data []byte, opts ...func(*hdopt)) string {
	var result strings.Builder
	result.Grow(64)

	options := hdopt{
		addr:  0,
		width: 16,
		word:  1,
		ascii: true,
	}
	for _, o := range opts {
		o(&options)
	}

	stop := len(data)
	for i := 0; i < stop; i += options.width {
		fmt.Fprintf(&result, "%08X: ", options.addr)
		for j := 0; j < options.width; j += options.word {
			result.WriteByte(' ')
			for k := 0; k < options.word; k++ {
				if i+j+k >= stop {
					result.WriteString("  ")
				} else {
					fmt.Fprintf(&result, "%02X", data[i+j+k])
				}
			}
		}
		if options.ascii {
			result.WriteString("  |")
			for j := 0; j < options.width; j++ {
				if i+j >= stop {
					result.WriteByte(' ')
				} else {
					if data[i+j] >= 32 && data[i+j] < 127 {
						result.WriteByte(data[i+j])
					} else {
						result.WriteByte('.')
					}
				}
			}
			result.WriteByte('|')
		}
		result.WriteByte('\n')
		options.addr += options.width
	}
	return result.String()
}

//
// WithStartingAddress may be added as an option to the Hexdump function
// to change the starting address of the data being shown.
//
// Example:
//  data := []byte("\x00\x81\x02\x03Hello, World™<>ABCDEFG")
//  Hexdump(data, WithStartingAddress(0x4444))
// will return the string
//  00004444:  00 81 02 03 48 65 6C 6C 6F 2C 20 57 6F 72 6C 64  |....Hello, World|
//  00004454:  E2 84 A2 3C 3E 41 42 43 44 45 46 47              |...<>ABCDEFG    |
//
func WithStartingAddress(a int) func(*hdopt) {
	return func(o *hdopt) {
		o.addr = a
	}
}

//
// WithWidth may be added as an option to the Hexdump function
// to change the output width in bytes.
//
// The behavior is undefined if the width is not a multiple
// of the word size.
//
// Example:
//  data := []byte("\x00\x81\x02\x03Hello, World™<>ABCDEFG")
//  Hexdump(data, WithWidth(8), WithStartingAddress(0x4444))
// will return the string
//  00004444:  00 81 02 03 48 65 6C 6C  |....Hell|
//  0000444C:  6F 2C 20 57 6F 72 6C 64  |o, World|
//  00004454:  E2 84 A2 3C 3E 41 42 43  |...<>ABC|
//  0000445C:  44 45 46 47              |DEFG    |
//
func WithWidth(w int) func(*hdopt) {
	return func(o *hdopt) {
		o.width = w
	}
}

//
// WithWordSize may be added as an option to the Hexdump function
// to change the output word size in bytes.
//
// Example:
//  data := []byte("\x00\x81\x02\x03Hello, World™<>ABCDEFG")
//  Hexdump(data, WithWordSize(2))
// will return the string
//  00000000:  0081 0203 4865 6C6C 6F2C 2057 6F72 6C64  |....Hello, World|
//  00000010:  E284 A23C 3E41 4243 4445 4647            |...<>ABCDEFG    |
//
func WithWordSize(w int) func(*hdopt) {
	return func(o *hdopt) {
		o.word = w
	}
}

//
// WithoutText may be added as an option to the Hexdump function
// to suppress the text column from the generated display.
//
// Example:
//  data := []byte("\x00\x81\x02\x03Hello, World™<>ABCDEFG")
//  Hexdump(data, WithWordSize(2), WithoutText)
// will return the string
//  00000000:  0081 0203 4865 6C6C 6F2C 2057 6F72 6C64
//  00000010:  E284 A23C 3E41 4243 4445 4647
//
func WithoutText(o *hdopt) {
	o.ascii = false
}

//
// PluralizeString emits a properly-pluralized version of a string,
// by adding an "s" for quantities other than one.
//
func PluralizeString(base string, qty int) string {
	if qty == 1 {
		return base
	}
	return base + "s"
}

//
// PluralizeCustom emits a properly-pluralized version of a string,
// where that is more complicated than just adding an "s" to the end.
//
func PluralizeCustom(base, singularSuffix, pluralSuffix string, qty int) string {
	if qty == 1 {
		return base + singularSuffix
	}
	return base + pluralSuffix
}

// @[00]@| Go-GMA 5.3.1
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
