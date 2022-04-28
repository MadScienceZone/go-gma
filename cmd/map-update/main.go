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

	"github.com/MadScienceZone/go-gma/v4/mapper"
	"github.com/MadScienceZone/go-gma/v4/util"
)

const GMAVersionNumber = "4.3.7" //@@##@@
const GMAMapperFileFormat = 20   //@@##@@

func main() {
	fmt.Printf("GMA map-update tool %s for map file format %d\n", GMAVersionNumber, GMAMapperFileFormat)

	for _, filename := range os.Args[1:] {
		fmt.Printf("Converting %s ", filename)
		objects, meta, err := mapper.ReadMapFile(filename)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		fmt.Printf("format %d -> %d ", meta.FileVersion, GMAMapperFileFormat)
		fmt.Printf("(%d %s) ", len(objects), util.PluralizeString("object", len(objects)))
		if err := os.Rename(filename, filename+".bak"); err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		if err := mapper.WriteMapFile(filename, objects, meta); err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		fmt.Printf("OK\n")
	}
}

/*
# @[00]@| GMA 4.3.7
*/
