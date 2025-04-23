/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  ______      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )/ ___  \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |\/   )  )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )    /  /    | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    /  /     | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/    /  /      |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ /  /     _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \_/     (_)(_______)     #
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

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -dice="3d6" or -dice "3d6"), except for boolean flags which may be given alone (e.g., -json) to indicate that the option is set to “true” or may be given an explicit value which must be attached to the option with an equals sign (e.g., -json=true or -json=false).

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
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/text"
)

const GoVersionNumber="5.27.0" //@@##@@

func main() {
	var err error
	var seedUsed int64

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-help] [-dice spec] [-json] [-seed value] [-syntax]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  An option 'x' with a value may be set by '-x value', '-x=value', '--x value', or '--x=value'.\n")
		fmt.Fprintf(os.Stderr, "  A flag 'x' may be set by '-x', '--x', '-x=true|false' or '--x=true|false'\n")
		fmt.Fprintf(os.Stderr, "  Options may NOT be combined into a single argument (use '-h -m', not '-hm').\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}
	help := flag.Bool("help", false, "list command-line options and exit")
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
		seedUsed = *seedValue
	} else {
		roller, err = dice.NewDieRoller()
		seedUsed = dice.DefaultSeed
	}
	if err != nil {
		fmt.Printf("Internal error: %v\n", err)
		os.Exit(1)
	}

	report := ReportedData{
		Seed: seedUsed,
	}

	if *rollSpec != "" {
		for i, thisRoll := range strings.Split(*rollSpec, ";") {
			title, results, err := roller.DoRoll(thisRoll)
			if err != nil {
				fmt.Printf("Error in die-roll expression #%d: %v\n", i+1, err)
				os.Exit(1)
			}
			r := ReportedResultSet{
				Title:   title,
				Results: results,
			}
			r.CalculateStats()
			report.AddResult(r)
		}

		if *asJSON {
			if err := report.WriteJSON(os.Stdout); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			report.WriteText(os.Stdout)
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
					r := ReportedResultSet{
						Title:   title,
						Results: results,
					}
					r.CalculateStats()
					r.WriteText(os.Stdout)
				}
			}
		}
	}
}

//
// ReportedData describes the full output of an invocation of the die roller.
// This includes overall metadata such as the PRNG seed value and a slice of
// the result set from each discrete die-roll request.
//
type ReportedData struct {
	ResultSet []ReportedResultSet
	Seed      int64 `json:",omitempty"`
}

//
// ResultStats provides statistics about the data set.
//
type ResultStats struct {
	N      int     // population size
	Mean   float64 // population mean (μ=Σ/N)
	Median float64 // median value
	Mode   []int   // population mode(s)
	StdDev float64 // standard deviation (σ=(sqrt(Σ((Xi-μ)**2))/(N-1)))
	Sum    int64   // sum of all values (Σ)
}

// ReportedResultSet describes the result of a single die-roll request.
// This may involve multiple die rolls, depending on the options included, and each of those
// may involve multiple dice being rolled.
//
type ReportedResultSet struct {
	Title   string `json:",omitempty"`
	Results []dice.StructuredResult
	Stats   *ResultStats `json:",omitempty"`
}

//
// AddResult adds a new result set to the output data.
//
func (rd *ReportedData) AddResult(r ReportedResultSet) {
	rd.ResultSet = append(rd.ResultSet, r)
}

//
// CalculateStats generates the ResultStats for a given result set,
// assign it to the receiver's Stats struct member. If there is only
// a single value, there's not much point so we leave Stats with a nil
// value in that case.
//
func (rs *ReportedResultSet) CalculateStats() {
	if len(rs.Results) < 2 {
		rs.Stats = nil
		return
	}

	// We only get this far if we have 2 or more values and thus
	// the stats become non-trivial. Note that the code below
	// this point ASSUMES there are more than 1 element in the
	// Results slice.

	rs.Stats = &ResultStats{
		N: len(rs.Results),
	}
	data := make([]int, 0, rs.Stats.N)
	for _, res := range rs.Results {
		if res.InvalidRequest || res.ResultSuppressed {
			rs.Stats.N--
			continue
		}

		data = append(data, res.Result)
		rs.Stats.Sum += int64(res.Result)
	}

	// we may have started with 2+ elements but if enough were
	// invalid/hidden we may still not have enough to do anything with.
	if rs.Stats.N < 2 {
		rs.Stats = nil
		return
	}

	slices.Sort(data)
	rs.Stats.Mean = float64(rs.Stats.Sum) / float64(rs.Stats.N)

	var v float64
	var cur, count, largest_count int

	setMode := func() {
		if count == largest_count {
			rs.Stats.Mode = append(rs.Stats.Mode, cur)
		} else if count > largest_count {
			rs.Stats.Mode = append(rs.Stats.Mode[:0], cur)
			largest_count = count
		}
	}

	for i, x := range data {
		v += math.Pow((float64(x) - rs.Stats.Mean), 2)
		if i == 0 {
			// nothing collected yet
			cur = x
			count = 1
		} else if cur != x {
			// hit a new value in the list. See what to do about the value we were tracking
			setMode()
			cur = x
			count = 1
		} else {
			count++
		}
	}
	setMode()
	rs.Stats.StdDev = math.Sqrt(v / float64(rs.Stats.N-1))

	if rs.Stats.N%2 == 0 {
		rs.Stats.Median = float64(data[rs.Stats.N/2]+data[rs.Stats.N/2-1]) / 2.0
	} else {
		rs.Stats.Median = float64(data[rs.Stats.N/2])
	}
}

//
// WriteJSON outputs reported data in JSON format to the designated output device.
//
func (rd ReportedData) WriteJSON(o io.Writer) error {
	j, err := json.Marshal(rd)
	if err != nil {
		return err
	}
	o.Write(j)
	return nil
}

//
// WriteText outputs reported data in plain text format.
//
func (rd ReportedData) WriteText(o io.Writer) {
	for i, set := range rd.ResultSet {
		if i > 0 {
			o.Write([]byte(strings.Repeat("\u2550", 80) + "\n"))
		}
		set.WriteText(o)
	}
}

//
// WriteText outputs reported data in plain text format.
//
func (rs ReportedResultSet) WriteText(o io.Writer) {

	if rs.Title != "" {
		o.Write([]byte("\033[1m\"" + rs.Title + "\":\033[0m\n"))
	}

	for i, res := range rs.Results {
		if len(rs.Results) > 1 {
			o.Write([]byte(fmt.Sprintf("\033[1;34mRoll #%d: \033[0m", i+1)))
		}
		if res.InvalidRequest {
			o.Write([]byte("\033[1;31m**INVALID DIE ROLL**\033[0m\n"))
			continue
		}
		if res.ResultSuppressed {
			o.Write([]byte("\033[1;33m**RESULT HIDDEN**\033[0m\n"))
			continue
		}

		if description, err := res.Details.Text(); err == nil {
			o.Write([]byte(description + "\n"))
		} else {
			o.Write([]byte(fmt.Sprintf("[%d] \033[1;31m**ERROR: %v**\033[0m\n", res.Result, err)))
		}
	}

	// If there are 3 or more data elements, report the stats. This keeps us
	// from being overly eager when interacting with the user. The JSON output
	// may choose to report the stats more often, but we won't.
	if rs.Stats != nil && rs.Stats.N >= 3 {
		o.Write([]byte(fmt.Sprintf("\033[36mN=%d, μ=%v, σ=%v, Md=%v, Mo=%v, Σ=%v\033[0m\n",
			rs.Stats.N,
			rs.Stats.Mean,
			rs.Stats.StdDev,
			rs.Stats.Median,
			rs.Stats.Mode,
			rs.Stats.Sum)))
	}
}
