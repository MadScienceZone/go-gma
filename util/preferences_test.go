/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  _______      __          #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )(  ____ \    /  \         #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  || (    \/    \/) )        #
# | |      | || || || (___) | Assistant | (____         /   )| (____        | |        #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   / (_____ \       | |        #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) )      | |        #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _  __) (_       #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_) \____/       #
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

func TestOverwriteMissingFields(t *testing.T) {
	j := strings.NewReader(`{
	"GMA_Mapper_preferences_version": 1,
	"animate": true
}`)
	p, err := LoadPreferencesWithDefaults(j)
	if err != nil {
		t.Errorf("LoadPreferencesWithDefaults: %v", err)
	}
	if p.Animate != true {
		t.Errorf("Animate %v", p.Animate)
	}
	if p.GuideLines.Major.Interval != 0 || p.GuideLines.Major.Offsets.X != 0 || p.GuideLines.Major.Offsets.Y != 0 {
		t.Errorf("Major %v %v %v", p.GuideLines.Major.Interval, p.GuideLines.Major.Offsets.X, p.GuideLines.Major.Offsets.Y)
	}
}

// @[00]@| Go-GMA 5.25.1
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
//
