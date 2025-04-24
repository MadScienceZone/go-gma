/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  ______       __          #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )/ ___  \     /  \         #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |\/   )  )    \/) )        #
# | |      | || || || (___) | Assistant | (____         /   )    /  /       | |        #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    /  /        | |        #
# | | \_  )| |   | || (   ) |                 ) )    /   _/    /  /         | |        #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ /  /     _  __) (_       #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \_/     (_) \____/       #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the util package
//

package util

import (
	"strings"
	"testing"
)

func TestSimpleConfig(t *testing.T) {
	f := strings.NewReader(`
#test data
a=b

 c  =  d  
x=12
b=off
bb=on
bbb=1
e
`)
	c, err := ParseSimpleConfig(f)
	if err != nil {
		t.Errorf("error unexpected: %v", err)
	}
	s, x := c.Get("a")
	if !x {
		t.Errorf("key not found")
	}
	if s != "b" {
		t.Errorf("%v value %v, expected %v", "a", s, "b")
	}

	s, x = c.GetDefault("a", "xy")
	if !x {
		t.Errorf("key not found")
	}
	if s != "b" {
		t.Errorf("%v value %v, expected %v", "a", s, "b")
	}

	s, x = c.Get("c")
	if !x {
		t.Errorf("key not found")
	}
	if s != "d" {
		t.Errorf("%v value %v, expected %v", "c", s, "d")
	}

	s, x = c.Get("x")
	if !x {
		t.Errorf("key not found")
	}
	if s != "12" {
		t.Errorf("%v value %v, expected %v", "c", s, "d")
	}
	i, err := c.GetInt("x")
	if err != nil {
		t.Errorf("key not found or error %v", err)
	}
	if i != 12 {
		t.Errorf("%v value %v, expected %v", "x", i, 12)
	}
	b, err := c.GetBool("x")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if !b {
		t.Errorf("value not true")
	}

	s, x = c.Get("b")
	if !x {
		t.Errorf("key not found")
	}
	if s != "off" {
		t.Errorf("%v value %v, expected %v", "b", s, "off")
	}
	i, err = c.GetInt("b")
	if err == nil {
		t.Errorf("err expected")
	}
	b, err = c.GetBool("b")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if b {
		t.Errorf("value not false")
	}

	s, x = c.Get("bb")
	if !x {
		t.Errorf("key not found")
	}
	if s != "on" {
		t.Errorf("%v value %v, expected %v", "bb", s, "on")
	}
	i, err = c.GetInt("bb")
	if err == nil {
		t.Errorf("err expected")
	}
	b, err = c.GetBool("bb")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if !b {
		t.Errorf("value not true")
	}

	s, x = c.Get("bbb")
	if !x {
		t.Errorf("key not found")
	}
	if s != "1" {
		t.Errorf("%v value %v, expected %v", "bbb", s, "1")
	}
	i, err = c.GetInt("bbb")
	if err != nil {
		t.Errorf("err %v", err)
	}
	b, err = c.GetBool("bbb")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if !b {
		t.Errorf("value not true")
	}

	s, x = c.Get("e")
	if !x {
		t.Errorf("key not found")
	}
	if s != "1" {
		t.Errorf("%v value %v, expected %v", "e", s, "1")
	}
	i, err = c.GetInt("e")
	if err != nil {
		t.Errorf("err %v", err)
	}
	b, err = c.GetBool("e")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if !b {
		t.Errorf("value not true")
	}

	s, x = c.Get("ee")
	if x {
		t.Errorf("key should not be found")
	}
	if s != "" {
		t.Errorf("%v value %v, expected %v", "ee", s, "")
	}
	i, err = c.GetInt("ee")
	if err == nil {
		t.Errorf("err expected")
	}
	b, err = c.GetBool("ee")
	if err != nil {
		t.Errorf("err %v", err)
	}
	if b {
		t.Errorf("value true")
	}
	i, err = c.GetIntDefault("ee", 32)
	if err != nil {
		t.Errorf("err %v", err)
	}
	if i != 32 {
		t.Errorf("%v value %v, expected %v", "ee", i, 32)
	}
	s, x = c.GetDefault("ee", "ff")
	if x {
		t.Errorf("ee should not exist")
	}
	if s != "ff" {
		t.Errorf("%v value %v, expected %v", "ee", s, "ff")
	}
}

// @[00]@| Go-GMA 5.27.1
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
//
