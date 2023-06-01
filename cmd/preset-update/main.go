/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      ______     _______         _____ #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ____ \   (  __   )       (  ___ #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (    \/   | (  )  |       | (    #
# | |      | || || || (___) | Assistant | (____     | (____     | | /   | _____ | (__/ #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \    |  ___ \    | (/ /) |(_____)|  __  #
# | | \_  )| |   | || (   ) |                 ) )   | (   ) )   |   / | |       | (  \ #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ ( (___) ) _ |  (__) |       | )___ #
# (_______)|/     \||/     \| Client    \______/ (_) \_____/ (_)(_______)       |/ \__ #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2022 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Preset-update reads each file named on the command line and rewrites the file in the current die-roll preset file format.
A copy of the original is kept with the same name plus a .bak suffix.

If no files are named, preset-update reads from its standard input and writes to its standard output.

The file format changed significantly between format versions 1 and 2.
The preset-update program can read format 1 files, so this provides a way to update existing map files to the newer format.
*/
package main

import (
	"fmt"
	"os"

	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/util"
)

const GoVersionNumber="5.6.0-beta.1"      //@@##@@
const GMADieRollPresetFileFormat = 2 //@@##@@

func main() {
	if len(os.Args) < 2 {
		// no files named; filter stdin to stdout
		presets, meta, err := dice.LoadDieRollPresetFile(os.Stdin)
		if err != nil {
			panic(err)
		}
		err = dice.SaveDieRollPresetFile(os.Stdout, presets, meta)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("GMA map-update tool %s for die-roll preset file format %d\n", GoVersionNumber, GMADieRollPresetFileFormat)

		for _, filename := range os.Args[1:] {
			fmt.Printf("Converting %s ", filename)
			presets, meta, err := dice.ReadDieRollPresetFile(filename)
			if err != nil {
				fmt.Printf("FAILED: %v\n", err)
				continue
			}
			fmt.Printf("format %d -> %d ", meta.FileVersion, GMADieRollPresetFileFormat)
			fmt.Printf("(%d %s) ", len(presets), util.PluralizeString("preset", len(presets)))
			if err := os.Rename(filename, filename+".bak"); err != nil {
				fmt.Printf("FAILED: %v\n", err)
				continue
			}
			if err := dice.WriteDieRollPresetFile(filename, presets, meta); err != nil {
				fmt.Printf("FAILED: %v\n", err)
				continue
			}
			fmt.Printf("OK\n")
		}
	}
}

/*
# @[00]@| Go-GMA 5.6.0-beta.1
# @[01]@|
# @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
# @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
# @[12]@| Aloha, Oregon, USA. All Rights Reserved.
# @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
# @[14]@| License as described in the accompanying LICENSE file distributed
# @[15]@| with GMA.
# @[16]@|
# @[20]@| Redistribution and use in source and binary forms, with or without
# @[21]@| modification, are permitted provided that the following conditions
# @[22]@| are met:
# @[23]@| 1. Redistributions of source code must retain the above copyright
# @[24]@|    notice, this list of conditions and the following disclaimer.
# @[25]@| 2. Redistributions in binary form must reproduce the above copy-
# @[26]@|    right notice, this list of conditions and the following dis-
# @[27]@|    claimer in the documentation and/or other materials provided
# @[28]@|    with the distribution.
# @[29]@| 3. Neither the name of the copyright holder nor the names of its
# @[30]@|    contributors may be used to endorse or promote products derived
# @[31]@|    from this software without specific prior written permission.
# @[32]@|
# @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
# @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
# @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
# @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
# @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
# @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
# @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
# @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
# @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
# @[45]@| SUCH DAMAGE.
# @[46]@|
# @[50]@| This software is not intended for any use or application in which
# @[51]@| the safety of lives or property would be at risk due to failure or
# @[52]@| defect of the software.
*/
