/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   __        __            #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) /  \      /  \           #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  | \/) )     \/) )          #
# | |      | || || || (___) | Assistant | (____         /   )   | |       | |          #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    | |       | |          #
# | | \_  )| |   | || (   ) |                 ) )    /   _/     | |       | |          #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ __) (_ _  __) (_         #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \____/(_) \____/         #
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
Roll provides a command-line utility that rolls dice using the GMA dice library.

It can be used interactively by users or it can be embedded in scripts for web-based tools or other backend services.

# SYNOPSIS

(If using the full GMA core tool suite)
   gma go roll ...

(Otherwise)
   roll -help
   roll -syntax
   roll [-seed value] [-dice spec] [-json]

# OPTIONS

Command-line options may be specified with one or two hyphens (e.g., -json or --json).

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -dice="3d6" or -dice "3d6"), except for boolean flags which may be given alone (e.g., -json) to indicate that the option is set to ``true'' or may be given an explicit value which must be attached to the option with an equals sign (e.g., -json=true or -json=false).

  -dice spec[;...]
      Specify the die-roll expression to be rolled, such as "3d6". If this is not given, roll will interactively prompt for die-roll expressions. Typing a blank line repeats the previous expression. The program will exit on EOF. Multiple die-roll specs may be given here, separated by semicolons. These will be rolled in order after setting the seed (if any).

  -help
      Print a command summary and exit.

  -json
      Print die-roll results in JSON format.

  -seed value
      Instead of using a random seed value, base the die roll results on the given value.
	  Value is a 64-bit integer expressed in decimal digits.

  -syntax
      Print a summary of the die-roll expression syntax and exit. In interactive mode, this help text may be produced by typing "help" as the input line.
*/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/text"
)

const GoVersionNumber="5.21.1" //@@##@@

func main() {
	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-help] [-dice spec] [-json] [-seed value] [-syntax]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  An option 'x' with a value may be set by '-x value', '-x=value', '--x value', or '--x=value'.\n")
		fmt.Fprintf(os.Stderr, "  A flag 'x' may be set by '-x', '--x', '-x=true|false' or '--x=true|false'\n")
		fmt.Fprintf(os.Stderr, "  Options may NOT be combined into a single argument (use '-h -m', not '-hm').\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}
	help := flag.Bool("help", false, "list command-line options and die-roll syntax description")
	rollSpec := flag.String("dice", "", "die-roll expression(s) to be rolled (semicolon-separated) (interactive if this is not given)")
	asJSON := flag.Bool("json", false, "print results in JSON")
	seedValue := flag.Int64("seed", 0, "seed value (0 for random)")
	syntaxHelp := flag.Bool("syntax", false, "print die-roll syntax description and exit")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *syntaxHelp {
		helpText, err := text.Render(dice.DieRollExpressionSyntax)
		if err != nil {
			fmt.Printf("Unable to render syntax help text: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(helpText)
		os.Exit(0)
	}

	var roller *dice.DieRoller
	if *seedValue != 0 {
		roller, err = dice.NewDieRoller(dice.WithSeed(*seedValue))
	} else {
		roller, err = dice.NewDieRoller()
	}
	if err != nil {
		fmt.Printf("Internal error: %v\n", err)
		os.Exit(1)
	}

	if *rollSpec != "" {
		var resultSet []ReportedResultSet
		for i, thisRoll := range strings.Split(*rollSpec, ";") {
			title, results, err := roller.DoRoll(thisRoll)
			if err != nil {
				fmt.Printf("Error in die-roll expression #%d: %v\n", i+1, err)
				os.Exit(1)
			}
			resultSet = append(resultSet, ReportedResultSet{
				Title:   title,
				Results: results,
				Seed:    *seedValue,
			})
		}

		if *asJSON {
			err = ReportJSON(resultSet)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			for _, theResult := range resultSet {
				ReportText(theResult.Title, theResult.Results)
			}
		}
	} else {
		fmt.Println("Enter each die-roll expression below.\nType \"help\" to see a syntax description.\nEOF terminates.")
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			if scanner.Text() == "help" {
				if helpText, err := text.Render(dice.DieRollExpressionSyntax); err == nil {
					fmt.Println(helpText)
				} else {
					fmt.Printf("Unable to print help text: %v\n", err)
				}
			} else {
				title, results, err := roller.DoRoll(scanner.Text())
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				} else {
					ReportText(title, results)
				}
			}
		}
	}
}

func ReportText(title string, results []dice.StructuredResult) {
	if title != "" {
		fmt.Printf("** %s **\n", title)
	}
	for i, res := range results {
		if len(results) > 1 {
			fmt.Printf("Roll #%d: ", i+1)
		}
		if res.InvalidRequest {
			fmt.Println("**INVALID DIE ROLL**")
			continue
		}
		if res.ResultSuppressed {
			fmt.Println("**RESULT HIDDEN**")
			continue
		}
		if description, err := res.Details.Text(); err == nil {
			fmt.Printf("%s\n", description)
		} else {
			fmt.Printf("[%d] **ERROR: %v**\n", res.Result, err)
		}
	}
}

type ReportedResultSet struct {
	Title   string                  `json:"title,omitempty"`
	Results []dice.StructuredResult `json:"results"`
	Seed    int64                   `json:"seed,omitempty"`
}

func ReportJSON(rs []ReportedResultSet) error {
	j, err := json.Marshal(struct {
		ResultSet []ReportedResultSet `json:"result_set"`
	}{
		ResultSet: rs,
	})
	if err != nil {
		return err
	}
	fmt.Print(string(j))
	return nil
}
