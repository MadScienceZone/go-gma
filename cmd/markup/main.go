/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   _____      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) / ___ \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |( (   ) )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )( (___) |   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /  \____  |   | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) |   |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_)(_______)     #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2024 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2024.
#
########################################################################
*/

/*
Markup provides a command-line utility that applies the GMA text markup formatter to its input.

# SYNOPSIS

(If using the full GMA core tool suite)
   gma go markup ...

(Otherwise)
   markup -help
   markup -syntax
   markup -preamble
   markup [-html] [-ps] <input >output

# OPTIONS

Command-line options may be specified with one or two hyphens (e.g., -html or --html).

  -help
      Print a command summary and exit.

  -html
      Render the markup input in HTML.

  -preamble
      Print GMA PostScript preamble before any other output.

  -ps
      Render the markup input in PostScript.
	  Requires the GMA PostScript preamble, plus PostScript
	  code to format this output appropriately.

  -syntax
      Print a summary of the markup syntax and exit.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MadScienceZone/go-gma/v5/text"
)

const GoVersionNumber="5.29.0" //@@##@@

func main() {
	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-help] [-html] [-preamble] [-ps] [-syntax]\n", os.Args[0])
		flag.PrintDefaults()
	}
	help := flag.Bool("help", false, "list command-line options")
	asHTML := flag.Bool("html", false, "print text in HTML")
	includePreamble := flag.Bool("preamble", false, "print GMA PostScript preamble before any output")
	asPS := flag.Bool("ps", false, "print text in PostScript (requires GMA PostScript preamble)")
	syntaxHelp := flag.Bool("syntax", false, "print markup syntax description and exit")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *asHTML && *asPS {
		fmt.Println("You can't use both -ps and -html at the same time. Pick a lane.")
		os.Exit(1)
	}

	if *includePreamble {
		fmt.Printf("%s", text.CommonPostScriptPreamble)
		fmt.Printf("%s", text.GMAPostScriptPreamble)
	}

	var inputText string
	var outputText string

	if *syntaxHelp {
		inputText = text.MarkupSyntax
	} else {
		var inputBytes []byte

		if inputBytes, err = io.ReadAll(os.Stdin); err != nil {
			fmt.Printf("ERROR reading input: %v\n", err)
			os.Exit(1)
		}
		inputText = string(inputBytes)
	}

	if *asPS {
		outputText, err = text.Render(inputText, text.AsPostScript)
	} else if *asHTML {
		outputText, err = text.Render(inputText, text.AsHTML)
	} else {
		outputText, err = text.Render(inputText)
	}

	if err != nil {
		fmt.Printf("Unable to render syntax help text: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(outputText)
}
