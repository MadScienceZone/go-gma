/*
########################################################################################
#  _______  _______  _______                ___       ______      ______               #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___  \              #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   )  )             #
# | |      | || || || (___) |            / (_) (_       ___) /       /  /              #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (       /  /               #
# | | \_  )| |   | || (   ) | Game           ) (           ) \     /  /                #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _  /  /                 #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_) \_/                  #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2021 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

package main

import (
	"fmt"
	"os"

	"github.com/MadScienceZone/go-gma/v4/dice"
	"github.com/MadScienceZone/go-gma/v4/util"
)

const GMAVersionNumber = "4.3.7"     //@@##@@
const GMADieRollPresetFileFormat = 2 //@@##@@

func main() {
	fmt.Printf("GMA map-update tool %s for die-roll preset file format %d\n", GMAVersionNumber, GMADieRollPresetFileFormat)

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

/*
# @[00]@| GMA 4.3.7
*/
