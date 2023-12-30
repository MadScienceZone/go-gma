/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __    ______      _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \  / ___  \    / ___   )       #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) \/   \  \   \/   )  |       #
# | |      | || || || (___) | Assistant | (____        | |    ___) /       /   )       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | |   (___ (      _/   /        #
# | | \_  )| |   | || (   ) |                 ) )      | |       ) \    /   _/         #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_/\___/  / _ (   (__/\       #
# (_______)|/     \||/     \| Client    \______/ (_) \____/\______/ (_)\_______/       #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2023 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Coredb is used for maintenance of the core (SRD) database, including loading
local items to the database and saving core data to files.

# SYNOPSIS

(If using the full GMA core tool suite)
   gma go coredb [options (see below) ...]

(Otherwise)
   coredb -h
   coredb -help
   coredb [-debug flags] [-export file] [-filter [!]re] [-ignore-case] [-import file] [-log file] [-preferences file] [-srd] [-type list]
   coredb [-D flags] [-e file] [-f [!]re] [-I] [-i file] [-l file] [-preferences file] [-srd] [-t list]

# OPTIONS

The command-line options described below have a long form (e.g., -log) and a  short form (e.g., -l) which are equivalent.
In either case, the option may be introduced with either one or two hyphens (e.g., -log or --log).

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -log=path or -log path), except for boolean flags which may be given alone (e.g., -I) to indicate that the option is set to ``true'' or may be given an explicit value which must be attached to the option with an equals sign (e.g., -I=true or -I=false).

You may not combine multiple single-letter options into a single composite argument, (e.g., the options -I and -h would need to be entered as two separate options, not as -Ih).

  -D, -debug flags
      Adds debugging messages to coredb's output. The flags
      value is a comma-separated list of debug flag names, which
      may be any of the following:

      all      Enable all debugging messages
      none     Disable all debugging messages
      queries  Database queries
      misc     Miscellaneous debugging messages

  -e, -export file
      Write database entries to the named file.

  -f, -filter [!]regex
      When importing or exporting, only include entries matching the regular expression
      regex. If regex begins with a '!' character, all entries which do NOT match the
      expression are included.

      The regex is matched against the Code and Name fields. If either matches, then
      the entry is included (or excluded). For languages, the Language field is checked.
      For monsters in the bestiary, the Code and Species fields are checked.

  -h, -help
      Print a command summary and exit.

  -I, -ignore-case
    Regex matching (-f/-filter option) is done irrespective of case.

  -i, -import file
      Read the contents of the file into the database.

  -l, -log file
      Write log messages to the named file instead of stdout.
      Use "-" for the file to explicitly send to stdout.

  -preferences file
      Use a custom profile instead of the default.

  -srd
      Normally, all entries imported by coredb into the database are assumed
      to be local entries added for the user's campaign. Entries exported to
      a file will only be local entries, ignoring the SRD entries initially
      loaded from community data by "gma initdb". With this option, the inverse
      is true: only the SRD entries will be exported and anything imported will
      be assumed to be non-local SRD data.

  -t, -type list
      Entries exported will include only database entries of the specified
      type(s). When importing, any records in the import file which are not
      of the specified type(s) will be skipped over.

      The list parameter is a comma-separated list of type names, which
      may be any of the following:

      all         All entry types
      none        Reset list to empty before adding more types
      bestiary    Creatures
      class[es]   Classes
      feat[s]     Feats
      language[s] Languages
      skill[s]    Skills
      spell[s]    Spells
      weapon[s]   Weapons

*/
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MadScienceZone/go-gma/v5/util"
)

var (
	Fdebug       string
	Fexport      string
	Fimport      string
	Flog         string
	Fpreferences string
	Fsrd         bool
	Ftype        string
	Ffilter      string
	Fignorecase  bool
)

func init() {
	const (
		defaultDebug      = "none"
		defaultExport     = ""
		defaultImport     = ""
		defaultLog        = ""
		defaultSRD        = false
		defaultType       = "all"
		defaultFilter     = ""
		defaultIgnoreCase = false
	)
	var defaultPreferences = ""

	homeDir, err := os.UserHomeDir()
	if err == nil {
		defaultPreferences = filepath.Join(homeDir, ".gma", "preferences.json")
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-D list] [-e file] [-f re] [-I] [-i file] [-l file] [-preferences file] [-srd] [-t list]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  An option 'x' with a value may be set by '-x value', '-x=value', '--x value', or '--x=value'.\n")
		fmt.Fprintf(os.Stderr, "  A flag 'x' may be set by '-x', '--x', '-x=true|false' or '--x=true|false'\n")
		fmt.Fprintf(os.Stderr, "  Options may NOT be combined into a single argument (use '-x -y', not '-xy').\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Allowed values for -debug: all, none, %s\n", strings.Join(util.DebugFlagNameSlice(util.DebugAll, false), ", "))
		fmt.Fprintf(os.Stderr, "Allowed values for -type: all, none, %s\n", strings.Join(util.TypeFilterNameSlice(util.AllTypes, false), ", "))
	}

	flag.StringVar(&Fdebug, "debug", defaultDebug, "Comma-separated list of debugging topics to print")
	flag.StringVar(&Fdebug, "D", defaultDebug, "(same as -debug)")

	flag.StringVar(&Fexport, "export", defaultExport, "Export entries to the named file")
	flag.StringVar(&Fexport, "e", defaultExport, "(same as -export)")

	flag.StringVar(&Ffilter, "filter", defaultFilter, "Filter to include (or exclude if re starts with '!') entries matching re")
	flag.StringVar(&Ffilter, "f", defaultFilter, "(same as -filter)")

	flag.BoolVar(&Fignorecase, "ignore-case", defaultIgnoreCase, "-filter should ignore case")
	flag.BoolVar(&Fignorecase, "I", defaultIgnoreCase, "(same as -ignore-case)")

	flag.StringVar(&Fimport, "import", defaultExport, "Export entries to the named file")
	flag.StringVar(&Fimport, "i", defaultExport, "(same as -export)")

	flag.StringVar(&Flog, "log", defaultLog, "Logfile ('-' is standard output)")
	flag.StringVar(&Flog, "l", defaultLog, "(same as -log)")

	flag.BoolVar(&Fsrd, "srd", defaultSRD, "Import/export SRD entries instead of local ones")

	flag.StringVar(&Fpreferences, "preferences", defaultPreferences, "GMA preferences file")

	flag.StringVar(&Ftype, "type", defaultType, "Comma-separated list of type(s) of entries to export or import")
	flag.StringVar(&Ftype, "t", defaultType, "(same as -type)")
}

func main() {
	log.SetPrefix("coredb: ")

	prefs, err := configureApp()
	if err != nil {
		log.Fatalf("unable to set up: %v", err)
	}
	if prefs.LogFile != "" && prefs.LogFile != "-" {
		f, err := os.OpenFile(prefs.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("unable to open log file \"%s\": %v", prefs.LogFile, err)
		}
		log.SetOutput(f)
	}

	prefs.CorePrefs.DebugBits, err = util.NamedDebugFlags(prefs.DebugFlags)
	if err != nil {
		log.Fatalf("-debug: %v", err)
	}

	prefs.CorePrefs.TypeBits, err = util.NamedTypeFilters(prefs.TypeList)
	if err != nil {
		log.Fatalf("-type: %v", err)
	}

	log.Printf("debugging %s, types %s, SRD %v, database path \"%s\"",
		util.DebugFlagNames(prefs.CorePrefs.DebugBits),
		util.TypeFilterNames(prefs.CorePrefs.TypeBits),
		prefs.CorePrefs.SRD,
		prefs.Prefs.CoreDBPath)
	if prefs.CorePrefs.FilterRegexp != nil {
		log.Printf("%s entries with pattern /%s/",
			func(x bool) string {
				if x {
					return "excluding"
				}
				return "including"
			}(prefs.CorePrefs.FilterExclude), prefs.CorePrefs.FilterRegexp.String())
	}

	if _, err = os.Stat(prefs.Prefs.CoreDBPath); os.IsNotExist(err) {
		log.Fatalf("core database does not exist; giving up!")
	}
	db, err := sql.Open("sqlite3", "file:"+prefs.Prefs.CoreDBPath)
	if err != nil {
		log.Fatalf("can't open database: %v", err)
	}
	defer db.Close()

	if prefs.ImportPath != "" {
		if err = importToCoreDB(db, &prefs); err != nil {
			log.Fatalf("error importing data: %v", err)
		}
	}

	if prefs.ExportPath != "" {
		if err = exportFromCoreData(db, &prefs); err != nil {
			log.Fatalf("error exporting data: %v", err)
		}
	}
}

func importToCoreDB(db *sql.DB, prefs *AppPreferences) error {
	log.Printf("importing from \"%s\"", prefs.ImportPath)
	fp, err := os.Open(prefs.ImportPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := fp.Close(); err != nil {
			log.Fatalf("error closing import file: %v", err)
		}
	}()
	if err = util.CoreImport(db, &prefs.CorePrefs, fp); err != nil {
		return err
	}
	return nil
}

func exportFromCoreData(db *sql.DB, prefs *AppPreferences) error {
	log.Printf("exporting to \"%s\"", prefs.ExportPath)
	fp, err := os.Create(prefs.ExportPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := fp.Close(); err != nil {
			log.Fatalf("error closing export file: %v", err)
		}
	}()

	if err = util.CoreExport(db, &prefs.CorePrefs, fp); err != nil {
		return err
	}
	return nil
}

type AppPreferences struct {
	Prefs      util.GMAPreferences
	CorePrefs  util.CorePreferences
	LogFile    string
	DebugFlags string
	ExportPath string
	ImportPath string
	TypeList   string
}

func configureApp() (AppPreferences, error) {
	var prefs AppPreferences
	var err error

	//
	// command-line parameters
	//
	flag.Parse()

	if prefsFile, err := os.Open(Fpreferences); err == nil {
		log.Println("Loading user preferences from", Fpreferences)
		prefs.Prefs, err = util.LoadGMAPreferencesWithDefaults(prefsFile)
		if err != nil {
			return prefs, err
		}
	}

	prefs.LogFile = Flog
	prefs.DebugFlags = Fdebug
	prefs.ExportPath = Fexport
	prefs.ImportPath = Fimport
	prefs.CorePrefs.SRD = Fsrd
	prefs.TypeList = Ftype
	if Ffilter != "" {
		if Ffilter[0:1] == "!" {
			prefs.CorePrefs.FilterExclude = true
			Ffilter = Ffilter[1:]
		}
		if Fignorecase {
			Ffilter = "(?i)" + Ffilter
		}
		if prefs.CorePrefs.FilterRegexp, err = regexp.Compile(Ffilter); err != nil {
			return prefs, err
		}
	}

	return prefs, nil
}

/*
# @[00]@| Go-GMA 5.13.2
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
