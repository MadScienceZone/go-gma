/*
} else {
	weap.Critical.CantCritical = true
}
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      ______     _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ____ \   (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (    \/   | (  )  |              #
# | |      | || || || (___) | Assistant | (____     | (____     | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \    |  ___ \    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )   | (   ) )   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ ( (___) ) _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_) \_____/ (_)(_______)              #
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
   gma go coredb ...

(Otherwise)
   coredb -h
   coredb -help
   coredb [-debug flags] [-export file] [-import file] [-log file] [-preferences file] [-srd] [-type list]
   coredb [-D flags] [-e file] [-i file] [-l file] [-preferences file] [-srd] [-t list]

# OPTIONS

The command-line options described below have a long form (e.g., -log) and a  short form (e.g., -l) which are equivalent.
In either case, the option may be introduced with either one or two hyphens (e.g., -log or --log).

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -log=path or -log path), except for boolean flags which may be given alone (e.g., -L) to indicate that the option is set to ``true'' or may be given an explicit value which must be attached to the option with an equals sign (e.g., -L=true or -L=false).

You may not combine multiple single-letter options into a single composite argument, (e.g., the options -L and -h would need to be entered as two separate options, not as -Lh).

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

  -h, -help
      Print a command summary and exit.

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
	  type(s). This is not necessary for import operations since the data type
	  is indicated in the file being read.

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
	 *weapon[s]   Weapons

*/
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
)

func init() {
	const (
		defaultDebug  = "none"
		defaultExport = ""
		defaultImport = ""
		defaultLog    = ""
		defaultSRD    = false
		defaultType   = "all"
	)
	var defaultPreferences = ""

	homeDir, err := os.UserHomeDir()
	if err == nil {
		defaultPreferences = filepath.Join(homeDir, ".gma", "preferences.json")
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-D list] [-e file] [-i file] [-l file] [-preferences file] [-srd] [-t list]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  An option 'x' with a value may be set by '-x value', '-x=value', '--x value', or '--x=value'.\n")
		fmt.Fprintf(os.Stderr, "  A flag 'x' may be set by '-x', '--x', '-x=true|false' or '--x=true|false'\n")
		fmt.Fprintf(os.Stderr, "  Options may NOT be combined into a single argument (use '-x -y', not '-xy').\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&Fdebug, "debug", defaultDebug, "Comma-separated list of debugging topics to print")
	flag.StringVar(&Fdebug, "D", defaultDebug, "(same as -debug)")

	flag.StringVar(&Fexport, "export", defaultExport, "Export entries to the named file")
	flag.StringVar(&Fexport, "e", defaultExport, "(same as -export)")

	flag.StringVar(&Fimport, "import", defaultExport, "Export entries to the named file")
	flag.StringVar(&Fimport, "i", defaultExport, "(same as -export)")

	flag.StringVar(&Flog, "log", defaultLog, "Logfile ('-' is standard output)")
	flag.StringVar(&Flog, "l", defaultLog, "(same as -log)")

	flag.BoolVar(&Fsrd, "srd", defaultSRD, "Import/export SRD entries instead of local ones")

	flag.StringVar(&Fpreferences, "preferences", defaultPreferences, "GMA preferences file")

	flag.StringVar(&Ftype, "type", defaultType, "Type(s) of entries to export")
	flag.StringVar(&Ftype, "t", defaultType, "(same as -type)")
}

type DebugFlags uint64

const (
	DebugQuery DebugFlags = 1 << iota
	DebugMisc
	DebugAll DebugFlags = 0xffffffff
)

func debugFlagNameSlice(flags DebugFlags) []string {
	if flags == 0 {
		return nil
	}
	if flags == DebugAll {
		return []string{"all"}
	}
	var list []string
	for _, f := range []struct {
		bits DebugFlags
		name string
	}{
		{bits: DebugQuery, name: "queries"},
		{bits: DebugMisc, name: "misc"},
	} {
		if (flags & f.bits) != 0 {
			list = append(list, f.name)
		}
	}
	return list
}

func debugFlagNames(flags DebugFlags) string {
	list := debugFlagNameSlice(flags)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

func namedDebugFlags(names ...string) (DebugFlags, error) {
	var d DebugFlags
	var err error
	for _, name := range names {
		for _, flag := range strings.Split(name, ",") {
			switch flag {
			case "":
			case "none":
				d = 0
			case "all":
				d = DebugAll
			case "query", "queries":
				d |= DebugQuery
			case "misc":
				d |= DebugMisc
			default:
				err = fmt.Errorf("invalid debug flag name")
			}
		}
	}
	return d, err
}

type TypeFilter uint64

const (
	TypeBestiary TypeFilter = 1 << iota
	TypeClass
	TypeFeat
	TypeLanguage
	TypeSkill
	TypeSpell
	TypeWeapon
	AllTypes TypeFilter = 0xffffffff
)

func typeFilterNameSlice(types TypeFilter) []string {
	if types == 0 {
		return nil
	}
	if types == AllTypes {
		return []string{"all"}
	}
	var list []string
	for _, f := range []struct {
		bits TypeFilter
		name string
	}{
		{bits: TypeBestiary, name: "bestiary"},
		{bits: TypeClass, name: "class"},
		{bits: TypeFeat, name: "feat"},
		{bits: TypeLanguage, name: "language"},
		{bits: TypeSkill, name: "skill"},
		{bits: TypeSpell, name: "spell"},
		{bits: TypeWeapon, name: "weapon"},
	} {
		if (types & f.bits) != 0 {
			list = append(list, f.name)
		}
	}
	return list
}

func typeFilterNames(types TypeFilter) string {
	list := typeFilterNameSlice(types)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

func namedTypeFilters(names ...string) (TypeFilter, error) {
	var f TypeFilter
	var err error
	for _, name := range names {
		for _, t := range strings.Split(name, ",") {
			switch t {
			case "":
			case "none":
				f = 0
			case "all":
				f = AllTypes
			case "bestiary", "monster", "monsters", "creature", "creatures":
				f |= TypeBestiary
			case "class", "classes":
				f |= TypeClass
			case "feat", "feats":
				f |= TypeFeat
			case "language", "languages":
				f |= TypeLanguage
			case "spell", "spells":
				f |= TypeSpell
			case "skill", "skills":
				f |= TypeSkill
			case "weapon", "weapons":
				f |= TypeWeapon
			default:
				err = fmt.Errorf("invalid data type filter name")
			}
		}
	}
	return f, err
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

	prefs.DebugBits, err = namedDebugFlags(prefs.DebugFlags)
	if err != nil {
		log.Fatalf("-debug: %v", err)
	}

	prefs.TypeBits, err = namedTypeFilters(prefs.TypeList)
	if err != nil {
		log.Fatalf("-type: %v", err)
	}

	log.Printf("debugging %s, types %s, SRD %v, database path \"%s\"",
		debugFlagNames(prefs.DebugBits),
		typeFilterNames(prefs.TypeBits),
		prefs.SRD,
		prefs.Prefs.CoreDBPath)

	if _, err = os.Stat(prefs.Prefs.CoreDBPath); os.IsNotExist(err) {
		log.Fatalf("core database does not exist; giving up!")
	}
	db, err := sql.Open("sqlite3", "file:"+prefs.Prefs.CoreDBPath)
	if err != nil {
		log.Fatalf("can't open database: %v", err)
	}
	defer db.Close()

	if prefs.ImportPath != "" {
		log.Printf("importing from \"%s\"", prefs.ImportPath)
	}

	if prefs.ExportPath != "" {
		log.Printf("exporting to \"%s\"", prefs.ExportPath)
		fp, err := os.Create(prefs.ExportPath)
		if err != nil {
			log.Fatalf("can't open export file: %v", err)
		}
		defer func() {
			if err := fp.Close(); err != nil {
				log.Fatalf("error closing export file: %v", err)
			}
		}()

		if _, err = fp.WriteString("{\"GMA_Core_Database_Export_Version\": 1,\n"); err != nil {
			log.Fatalf("i/o error: %v", err)
		}
		if (prefs.TypeBits & TypeFeat) != 0 {
			if err = exportFeats(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting feats: %v", err)
			}
		}
		if (prefs.TypeBits & TypeSpell) != 0 {
			if err = exportSpells(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting spells: %v", err)
			}
		}
		if (prefs.TypeBits & TypeBestiary) != 0 {
			if err = exportBestiary(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting bestiary: %v", err)
			}
		}
		if (prefs.TypeBits & TypeClass) != 0 {
			if err = exportClasses(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting classes: %v", err)
			}
		}
		if (prefs.TypeBits & TypeLanguage) != 0 {
			if err = exportLanguages(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting languages: %v", err)
			}
		}
		if (prefs.TypeBits & TypeSkill) != 0 {
			if err = exportSkills(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting skills: %v", err)
			}
		}
		if (prefs.TypeBits & TypeWeapon) != 0 {
			if err = exportWeapons(fp, db, &prefs); err != nil {
				log.Fatalf("error exporting weapons: %v", err)
			}
		}
		if _, err = fmt.Fprintf(fp, " \"SRD\": %v\n}\n", prefs.SRD); err != nil {
			log.Fatalf("i/o error finishing file: %v", err)
		}
	}
}

func query(db *sql.DB, prefs *AppPreferences, query string, args ...any) (*sql.Rows, error) {
	if (prefs.DebugBits & DebugQuery) != 0 {
		log.Printf("query: \"%s\" with %q", query, args)
	}
	rows, err := db.Query(query, args...)
	if (prefs.DebugBits&DebugQuery) != 0 && err != nil {
		log.Printf("query returned error: %v", err)
	}
	return rows, err
}

func getBitStrings(db *sql.DB, prefs *AppPreferences, q string) (map[int]string, error) {
	flagList := make(map[int]string)

	rows, err := query(db, prefs, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bit int
		var name string

		if err := rows.Scan(&bit, &name); err != nil {
			return nil, err
		}
		flagList[bit] = name
	}
	return flagList, rows.Err()
}

func getAlignmentCodes(db *sql.DB, prefs *AppPreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Code FROM Alignments")
}

func getFeatFlags(db *sql.DB, prefs *AppPreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Flag FROM FeatFlags")
}

func getSpellComponents(db *sql.DB, prefs *AppPreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Component FROM SpellComponents")
}

func getSpellDescriptors(db *sql.DB, prefs *AppPreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Descriptor FROM SpellDescriptors")
}

func getClassCodes(db *sql.DB, prefs *AppPreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Code FROM Classes")
}

func exportFeats(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting feat data")
	}
	if _, err = fp.WriteString(" \"Feats\": [\n"); err != nil {
		return err
	}

	featFlags, err := getFeatFlags(db, prefs)
	if err != nil {
		return err
	}
	if rows, err = query(db, prefs,
		`SELECT 
			ID, Code, Name, Parameters, Feats.IsLocal, Description, Flags, Prerequisites, Benefit,
			Normal, Special, Source, Race, Note, Goal, CompletionBenefit, SuggestedTraits,
			Adjective, LevelCost, Symbol
		FROM Feats
		LEFT JOIN MetaMagic
			ON MetaMagic.FeatID = ID`,
	); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var feat Feat
		var adj, sym, param, prereq, benefit, normal, special, source, race, note, goal, comp, traits sql.NullString
		var levelcost sql.NullInt32
		var feat_db_id int

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&feat_db_id, &feat.Code, &feat.Name, &param, &feat.IsLocal, &feat.Description,
			&feat.Flags, &prereq, &benefit, &normal, &special, &source, &race,
			&note, &goal, &comp, &traits,
			&adj, &levelcost, &sym); err != nil {
			return err
		}
		if param.Valid {
			feat.Parameters = param.String
		}
		if prereq.Valid {
			feat.Prerequisites = prereq.String
		}
		if benefit.Valid {
			feat.Benefit = benefit.String
		}
		if normal.Valid {
			feat.Normal = normal.String
		}
		if special.Valid {
			feat.Special = special.String
		}
		if source.Valid {
			feat.Source = source.String
		}
		if race.Valid {
			feat.Race = race.String
		}
		if note.Valid {
			feat.Note = note.String
		}
		if goal.Valid {
			feat.Goal = goal.String
		}
		if comp.Valid {
			feat.CompletionBenefit = comp.String
		}
		if traits.Valid {
			feat.SuggestedTraits = traits.String
		}

		if prefs.SRD == feat.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping feat %s because SRD=%v", feat.Code, prefs.SRD)
			}
			continue
		}

		if err = func() error {
			var rows *sql.Rows
			var err error
			var t string

			if rows, err = query(db, prefs, "SELECT FeatType FROM FeatFeatTypes WHERE FeatID=?", feat_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				if err = rows.Scan(&t); err != nil {
					return err
				}
				feat.Types = append(feat.Types, t)
			}
			return nil
		}(); err != nil {
			return err
		}

		for bit, name := range featFlags {
			if (feat.Flags & (1 << bit)) != 0 {
				feat.FlagNames = append(feat.FlagNames, name)
			}
		}

		if adj.Valid || levelcost.Valid || sym.Valid {
			feat.MetaMagic.IsMetaMagicFeat = true
			if adj.Valid {
				feat.MetaMagic.Adjective = adj.String
			}
			if levelcost.Valid {
				feat.MetaMagic.LevelCost = int(levelcost.Int32)
			} else {
				feat.MetaMagic.IsLevelCostVariable = true
			}
			if sym.Valid {
				feat.MetaMagic.Symbol = sym.String
			}
		}
		bytes, err := json.MarshalIndent(feat, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

func exportWeapons(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting weapon data")
	}
	if _, err = fp.WriteString(" \"Weapons\": [\n"); err != nil {
		return err
	}

	if rows, err = query(db, prefs,
		`SELECT 
			IsLocal, Code, Cost, Name, DmgT, DmgS, DmgM, DmgL,
			CritMultiplier, CritThreat, RangeIncrement, RangeMax,
			Weight, DmgTypes, Qualities
		FROM Weapons`,
	); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var weap Weapon
		var cost, ri, rmax, wt sql.NullInt32
		var dt, ct, cm, ds, dm, dl, dtyp, q sql.NullString

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&weap.IsLocal, &weap.Code, &cost, &weap.Name,
			&dt, &ds, &dm, &dl, &cm, &ct, &ri, &rmax, &wt, &dtyp, &q); err != nil {
			return err
		}

		if prefs.SRD == weap.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping weapon %s because SRD=%v", weap.Code, prefs.SRD)
			}
			continue
		}

		weap.Damage = make(map[string]string)

		if wt.Valid {
			weap.Weight = int(wt.Int32)
		}
		if dtyp.Valid {
			weap.DamageTypes = dtyp.String
		}
		if q.Valid {
			weap.Qualities = q.String
		}
		if cost.Valid {
			weap.Cost = int(cost.Int32)
		}
		if dt.Valid {
			weap.Damage["T"] = dt.String
		}
		if ds.Valid {
			weap.Damage["S"] = ds.String
		}
		if dm.Valid {
			weap.Damage["M"] = dm.String
		}
		if dl.Valid {
			weap.Damage["L"] = dl.String
		}
		if ct.Valid || cm.Valid {
			if ct.Valid {
				weap.Critical.Threat = ct.String
			}
			if cm.Valid {
				weap.Critical.Multiplier = cm.String
			}
		} else {
			weap.Critical.CantCritical = true
		}
		if ri.Valid || rmax.Valid {
			if ri.Valid {
				weap.Ranged.Increment = int(ri.Int32)
			}
			if rmax.Valid {
				weap.Ranged.MaxIncrements = int(rmax.Int32)
			}
			weap.Ranged.IsRanged = true
		}

		bytes, err := json.MarshalIndent(weap, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

func exportSkills(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting skill data")
	}
	if _, err = fp.WriteString(" \"Skills\": [\n"); err != nil {
		return err
	}

	classCodes, err := getClassCodes(db, prefs)
	if err != nil {
		return err
	}
	if rows, err = query(db, prefs,
		`SELECT 
			ID, Name, Classes, Ability, ArmorCheck, TrainedOnly, Source,
			Description, FullText, ParentSkill, IsVirtual, IsLocal, IsBackground
		FROM Skills
		`); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var skill_db_id int
		var classbits int
		var sk Skill
		var source sql.NullString
		var parent sql.NullInt32

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&skill_db_id, &sk.Name, &classbits, &sk.Ability,
			&sk.HasArmorPenalty, &sk.TrainingRequired, &source, &sk.Description,
			&sk.FullText, &parent, &sk.IsVirtual, &sk.IsLocal, &sk.IsBackground); err != nil {
			return err
		}

		if prefs.SRD == sk.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping skill %s because SRD=%v", sk.Name, prefs.SRD)
			}
			continue
		}

		for bit, code := range classCodes {
			if (classbits & (1 << bit)) != 0 {
				sk.ClassSkillFor = append(sk.ClassSkillFor, code)
			}
		}

		bytes, err := json.MarshalIndent(sk, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

func exportLanguages(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting language data")
	}
	if _, err = fp.WriteString(" \"Languages\": [\n"); err != nil {
		return err
	}

	if rows, err = query(db, prefs, `SELECT Language, IsLocal FROM Languages`); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var lang BaseLanguage

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&lang.Language, &lang.IsLocal); err != nil {
			return err
		}

		if prefs.SRD == lang.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping language %s because SRD=%v", lang.Language, prefs.SRD)
			}
			continue
		}

		bytes, err := json.MarshalIndent(lang, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

func exportClasses(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting class data")
	}
	if _, err = fp.WriteString(" \"Classes\": [\n"); err != nil {
		return err
	}

	if rows, err = query(db, prefs,
		`SELECT 
				ID, Name, Code,
				MagicType, Ability, Bonus, IsSpontaneous
			FROM Classes
			LEFT JOIN ClassMagic
				ON ClassID = ID
		`); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var cls Class
		var cls_db_id int
		var mtype, abil sql.NullString
		var bonus, spon sql.NullBool

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&cls_db_id, &cls.Name, &cls.Code, &mtype, &abil, &bonus, &spon); err != nil {
			return err
		}

		if prefs.SRD == cls.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping class %s because SRD=%v", cls.Code, prefs.SRD)
			}
			continue
		}

		if mtype.Valid || abil.Valid || bonus.Valid || spon.Valid {
			if mtype.Valid {
				cls.Spells.Type = mtype.String
			}
			if abil.Valid {
				cls.Spells.Ability = abil.String
			}
			if bonus.Valid {
				cls.Spells.HasBonusSpells = bonus.Bool
			}
			if spon.Valid {
				cls.Spells.IsSpontaneous = spon.Bool
			}

			if err = func() error {
				var rows *sql.Rows
				var err error

				firstResult := true
				lastLevel := 0

				if rows, err = query(db, prefs,
					`SELECT
						ClassLevel, SpellLevel, CastPerDay, PrepPerDay, Known
					FROM ClassSpells
					WHERE
						ClassID=?
					ORDER BY ClassLevel, SpellLevel
				`, cls_db_id); err != nil {
					return err
				}
				defer rows.Close()
				var cast, prep, known []ClassSpellLevel
				for rows.Next() {
					var cl, sl int
					var cpd, ppd, kn sql.NullInt32
					var c, p, k ClassSpellLevel

					if err := rows.Scan(&cl, &sl, &cpd, &ppd, &kn); err != nil {
						return err
					}
					if firstResult || cl != lastLevel {
						if !firstResult {
							cls.Spells.CastPerDay = append(cls.Spells.CastPerDay, cast)
							cls.Spells.PreparedPerDay = append(cls.Spells.PreparedPerDay, prep)
							cls.Spells.SpellsKnown = append(cls.Spells.SpellsKnown, known)
						}
						cast = []ClassSpellLevel{}
						prep = []ClassSpellLevel{}
						known = []ClassSpellLevel{}
						firstResult = false
						lastLevel = cl
					}
					if cpd.Valid {
						c.ClassLevel = cl
						c.SpellLevel = sl
						if cpd.Int32 < 0 {
							c.IsUnlimitedUse = true
						} else {
							c.Number = int(cpd.Int32)
						}
					} else {
						c.IsProhibited = true
					}

					if ppd.Valid {
						p.ClassLevel = cl
						p.SpellLevel = sl
						if ppd.Int32 < 0 {
							p.IsUnlimitedUse = true
						} else {
							p.Number = int(ppd.Int32)
						}
					} else {
						p.IsProhibited = true
					}

					if kn.Valid {
						k.ClassLevel = cl
						k.SpellLevel = sl
						if kn.Int32 < 0 {
							k.IsUnlimitedUse = true
						} else {
							k.Number = int(kn.Int32)
						}
					} else {
						k.IsProhibited = true
					}

					cast = append(cast, c)
					prep = append(prep, p)
					known = append(known, k)
				}
				if !firstResult {
					cls.Spells.CastPerDay = append(cls.Spells.CastPerDay, cast)
					cls.Spells.PreparedPerDay = append(cls.Spells.PreparedPerDay, prep)
					cls.Spells.SpellsKnown = append(cls.Spells.SpellsKnown, known)
				}
				return rows.Err()
			}(); err != nil {
				return err
			}
		}

		bytes, err := json.MarshalIndent(cls, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

func exportBestiary(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting bestiary data")
	}
	if _, err = fp.WriteString(" \"Bestiary\": [\n"); err != nil {
		return err
	}

	alignmentList, err := getAlignmentCodes(db, prefs)
	if err != nil {
		return err
	}
	if rows, err = query(db, prefs,
		`SELECT 
			ID, IsLocal, Species, Code, CR, XP, Class, Alignment, AlignmentSpecial, Source, Size, SpaceText,
			ReachText, Type, Initiative, InitiativeText, Senses, Aura, TypicalHP, CurrentHP, HPSpecial, HitDice,
			Fort, Refl, Will, FortText, ReflText, WillText, SaveMods, DefensiveAbilities, DR, DRBypass,
			Immunities, Resists, SR, SRText, Weaknesses, Speed, SpeedText, SpecialAttacks,
			Str, Dex, Con, Int, Wis, Cha, StrText, DexText, ConText, IntText, WisText, ChaText,
			BAB, CMB, CMD, CMBText, CMDText, SQ, Environment, Organization, Treasure, Appearance, Grp,
			IsTemplate, BeforeCombat, DuringCombat, Morale, CharacterFlag, CompanionFlag,
			IsUniqueMonster, AgeCategory, Gender, Bloodline, Patron, AlternateNameForm, DontUseRacialHD,
			VariantParent, MR, IsMythic, MT, OffenseNote, StatisticsNote, Gear, OtherGear, FocusedSchool,
ClassArchetypes, BaseStatistics, ACAdj, FlatAdj, TouchAdj, RacialMods, ProhibitedSchools,
			OppositionSchools, Mystery, Notes
FROM
			Monsters
		`,
	); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var monster Monster
		var mob_db_id, aligns int
		var curhp, f, r, w, dr, sr, str, dex, con, int_, wis, cha sql.NullInt32
		var bab, cmb, cmd, mr, mt, acadj, flatadj, touchadj sql.NullInt32
		var cls, alspec, src, spctext, rchtext, itext, senses, aura sql.NullString
		var hpspec, ft, rt, wt, savemods, defs, drbypass, immun, srtext, resist sql.NullString
		var weak, sptext, spatt, strt, dext, cont, intt, wist, chat, cmbt, cmdt sql.NullString
		var sq, env, org, treas, appear, grp, before, during, morale, age, gender, bline, patron, altname sql.NullString
		var varpar, offense, stats, gear, other, focused, arch, basestats, racial, prohibit, oppos, mystery, notes sql.NullString

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&mob_db_id, &monster.IsLocal, &monster.Species, &monster.Code, &monster.CR, &monster.XP,
			&cls, &aligns, &alspec, &src, &monster.Size.Code, &spctext, &rchtext, &monster.Type, &monster.Initiative.Mod,
			&itext, &senses, &aura, &monster.HP.Typical, &curhp, &hpspec, &monster.HP.HitDice, &f, &r, &w, &ft, &rt, &wt,
			&savemods, &defs, &dr, &drbypass, &immun, &resist, &sr, &srtext, &weak, &monster.Speed.Code, &sptext, &spatt,
			&str, &dex, &con, &int_, &wis, &cha, &strt, &dext, &cont, &intt, &wist, &chat, &bab, &cmb, &cmd,
			&cmbt, &cmdt, &sq, &env, &org, &treas, &appear, &grp, &monster.IsTemplate, &before, &during, &morale,
			&monster.IsCharacter, &monster.IsCompanion, &monster.IsUnique, &age, &gender, &bline, &patron, &altname,
			&monster.DontUseRacialHD, &varpar, &mr, &monster.Mythic.IsMythic, &mt, &offense, &stats, &gear, &other, &focused,
			&arch, &basestats, &acadj, &flatadj, &touchadj, &racial, &prohibit, &oppos, &mystery, &notes,
		); err != nil {
			return err
		}

		if before.Valid {
			monster.Strategy.BeforeCombat = before.String
		}
		if during.Valid {
			monster.Strategy.DuringCombat = during.String
		}
		if morale.Valid {
			monster.Strategy.Morale = morale.String
		}
		if age.Valid {
			monster.AgeCategory = age.String
		}
		if gender.Valid {
			monster.Gender = gender.String
		}
		if bline.Valid {
			monster.Bloodline = bline.String
		}
		if patron.Valid {
			monster.Patron = patron.String
		}
		if altname.Valid {
			monster.AlternateNameForm = altname.String
		}
		if varpar.Valid {
			monster.VariantParent = varpar.String
		}
		if mr.Valid {
			monster.Mythic.MR = int(mr.Int32)
		}
		if mt.Valid {
			monster.Mythic.MT = int(mt.Int32)
		}
		if offense.Valid {
			monster.OffenseNote = offense.String
		}
		if stats.Valid {
			monster.StatisticsNote = stats.String
		}
		if gear.Valid {
			monster.Gear.Combat = gear.String
		}
		if other.Valid {
			monster.Gear.Other = other.String
		}
		if focused.Valid {
			monster.Schools.Focused = focused.String
		}
		if arch.Valid {
			monster.ClassArchetypes = arch.String
		}
		if basestats.Valid {
			monster.BaseStatistics = basestats.String
		}
		if acadj.Valid || flatadj.Valid || touchadj.Valid {
			monster.AC.Adjustments = make(map[string]int)
			if acadj.Valid {
				monster.AC.Adjustments["AC"] = int(acadj.Int32)
			}
			if flatadj.Valid {
				monster.AC.Adjustments["Flat"] = int(flatadj.Int32)
			}
			if touchadj.Valid {
				monster.AC.Adjustments["Touch"] = int(touchadj.Int32)
			}
		}
		if racial.Valid {
			monster.RacialMods = racial.String
		}
		if prohibit.Valid {
			monster.Schools.Prohibited = prohibit.String
		}
		if oppos.Valid {
			monster.Schools.Opposition = oppos.String
		}
		if mystery.Valid {
			monster.Mystery = mystery.String
		}
		if notes.Valid {
			monster.Notes = notes.String
		}
		if cls.Valid {
			monster.Class = cls.String
		}
		if alspec.Valid {
			monster.Alignment.Special = alspec.String
		}
		if src.Valid {
			monster.Source = src.String
		}
		if spctext.Valid {
			monster.Size.SpaceText = spctext.String
		}
		if rchtext.Valid {
			monster.Size.ReachText = rchtext.String
		}
		if itext.Valid {
			monster.Initiative.Special = itext.String
		}
		if senses.Valid {
			monster.Senses = senses.String
		}
		if aura.Valid {
			monster.Aura = aura.String
		}
		if curhp.Valid {
			monster.HP.Current = int(curhp.Int32)
		}
		if hpspec.Valid {
			monster.HP.Special = hpspec.String
		}
		if f.Valid {
			monster.Save.Fort.Mod = int(f.Int32)
		}
		if r.Valid {
			monster.Save.Refl.Mod = int(r.Int32)
		}
		if w.Valid {
			monster.Save.Will.Mod = int(w.Int32)
		}
		if ft.Valid {
			monster.Save.Fort.Special = ft.String
		}
		if rt.Valid {
			monster.Save.Refl.Special = rt.String
		}
		if wt.Valid {
			monster.Save.Will.Special = wt.String
		}
		if savemods.Valid {
			monster.Save.Special = savemods.String
		}
		if defs.Valid {
			monster.DefensiveAbilities = defs.String
		}
		if dr.Valid {
			monster.DR.DR = int(dr.Int32)
		}
		if drbypass.Valid {
			monster.DR.Bypass = drbypass.String
		}
		if immun.Valid {
			monster.Immunities = immun.String
		}
		if resist.Valid {
			monster.Resists = resist.String
		}
		if sr.Valid {
			monster.SR.SR = int(sr.Int32)
		}
		if srtext.Valid {
			monster.SR.Special = srtext.String
		}
		if weak.Valid {
			monster.Weaknesses = weak.String
		}
		if sptext.Valid {
			monster.Speed.Special = sptext.String
		}
		if spatt.Valid {
			monster.SpecialAttacks = spatt.String
		}
		if str.Valid {
			monster.Abilities.Str.Base = int(str.Int32)
		} else {
			monster.Abilities.Str.NullScore = true
		}
		if strt.Valid {
			monster.Abilities.Str.Special = strt.String
		}
		if dex.Valid {
			monster.Abilities.Dex.Base = int(dex.Int32)
		} else {
			monster.Abilities.Dex.NullScore = true
		}
		if dext.Valid {
			monster.Abilities.Dex.Special = dext.String
		}
		if con.Valid {
			monster.Abilities.Con.Base = int(con.Int32)
		} else {
			monster.Abilities.Con.NullScore = true
		}
		if cont.Valid {
			monster.Abilities.Con.Special = cont.String
		}
		if int_.Valid {
			monster.Abilities.Int.Base = int(int_.Int32)
		} else {
			monster.Abilities.Int.NullScore = true
		}
		if intt.Valid {
			monster.Abilities.Int.Special = intt.String
		}
		if wis.Valid {
			monster.Abilities.Wis.Base = int(wis.Int32)
		} else {
			monster.Abilities.Wis.NullScore = true
		}
		if wist.Valid {
			monster.Abilities.Wis.Special = wist.String
		}
		if cha.Valid {
			monster.Abilities.Cha.Base = int(cha.Int32)
		} else {
			monster.Abilities.Cha.NullScore = true
		}
		if chat.Valid {
			monster.Abilities.Cha.Special = chat.String
		}
		if bab.Valid {
			monster.Combat.BAB = int(bab.Int32)
		}
		if cmb.Valid {
			monster.Combat.CMB = int(cmb.Int32)
		}
		if cmbt.Valid {
			monster.Combat.CMBSpecial = cmbt.String
		}
		if cmd.Valid {
			monster.Combat.CMD = int(cmd.Int32)
		}
		if cmdt.Valid {
			monster.Combat.CMDSpecial = cmdt.String
		}
		if sq.Valid {
			monster.SQ = sq.String
		}
		if env.Valid {
			monster.Environment = env.String
		}
		if org.Valid {
			monster.Organization = org.String
		}
		if treas.Valid {
			monster.Treasure = treas.String
		}
		if appear.Valid {
			monster.Appearance = appear.String
		}
		if grp.Valid {
			monster.Group = grp.String
		}

		for bit, align := range alignmentList {
			if (aligns & (1 << bit)) != 0 {
				monster.Alignment.Alignments = append(monster.Alignment.Alignments, align)
			}
		}

		if prefs.SRD == monster.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping monster %s because SRD=%v", monster.Species, prefs.SRD)
			}
			continue
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Domain
				FROM MonsterDomains
				LEFT JOIN Domains
					ON MonsterDomains.DomainID=Domains.ID
				WHERE
					MonsterDomains.MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var dn string
				if err := rows.Scan(&dn); err != nil {
					return err
				}
				monster.Domains = append(monster.Domains, dn)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Code, Value
				FROM ACComponents
				LEFT JOIN ACCTypes
					ON ACCTypes.ID=ACComponents.ACComponent
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var ac string
				var v int
				if monster.AC.Components == nil {
					monster.AC.Components = make(map[string]int)
				}
				if err := rows.Scan(&ac, &v); err != nil {
					return err
				}
				monster.AC.Components[ac] = v
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT 
					TierGroup, BaseWeaponID, Multiple, Name, Attack, Damage,
					Threat, Critical, RangeInc, RangeMax, IsReach, Special, Mode
				FROM AttackModes
				WHERE
					MonsterID=?
				ORDER BY
					TierGroup, TierSeq
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var am AttackMode
				var weap, att, dam, spec, threat, mult sql.NullString
				var ri, rmax sql.NullInt32

				if err := rows.Scan(&am.Tier, &weap, &am.Multiple, &am.Name, &att, &dam,
					&threat, &mult, &ri, &rmax, &am.IsReach, &spec, &am.Mode); err != nil {
					return err
				}
				if weap.Valid {
					am.BaseWeaponID = weap.String
				}
				if att.Valid {
					am.Attack = att.String
				}
				if dam.Valid {
					am.Damage = dam.String
				}
				if spec.Valid {
					am.Special = spec.String
				}
				if threat.Valid || mult.Valid {
					if threat.Valid {
						am.Critical.Threat = threat.String
					} else {
						am.Critical.Threat = "20"
					}
					if mult.Valid {
						am.Critical.Multiplier = mult.String
					} else {
						am.Critical.Multiplier = "2"
					}
				} else {
					am.Critical.CantCritical = true
				}
				if ri.Valid || rmax.Valid {
					am.Ranged.IsRanged = true
					if ri.Valid {
						am.Ranged.Increment = int(ri.Int32)
					}
					if rmax.Valid {
						am.Ranged.MaxIncrements = int(rmax.Int32)
					}
				}
				monster.AttackModes = append(monster.AttackModes, am)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Subtype
				FROM MonsterMonsterSubtypes
				LEFT JOIN MonsterSubtypes
					ON MonsterSubtypes.ID = MonsterMonsterSubtypes.SubtypeID
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var st string
				if err := rows.Scan(&st); err != nil {
					return err
				}
				monster.Subtypes = append(monster.Subtypes, st)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Language, IsMute, Special
				FROM MonsterLanguages
				LEFT JOIN Languages
					ON Languages.ID = MonsterLanguages.LanguageID
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var spec sql.NullString
				var lang Language
				if err := rows.Scan(&lang.Name, &lang.IsMute, &spec); err != nil {
					return err
				}
				if spec.Valid {
					lang.Special = spec.String
				}
				monster.Languages = append(monster.Languages, lang)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Code, MonsterFeats.Parameters, BonusFeat
				FROM MonsterFeats
				LEFT JOIN Feats
					ON Feats.ID=MonsterFeats.FeatID
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var feat MonsterFeat
				var params sql.NullString
				if err := rows.Scan(&feat.Code, &params, &feat.IsBonus); err != nil {
					return err
				}
				if params.Valid {
					feat.Parameters = params.String
				}
				monster.Feats = append(monster.Feats, feat)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT Code, Modifier, Notes
				FROM MonsterSkills
				LEFT JOIN Skills
					ON Skills.ID=MonsterSkills.SkillID
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var sk MonsterSkill
				var notes sql.NullString

				if err := rows.Scan(&sk.Code, &sk.Modifier, &notes); err != nil {
					return err
				}
				if notes.Valid {
					sk.Notes = notes.String
				}
				monster.Skills = append(monster.Skills, sk)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		if err = func() error {
			var rows *sql.Rows
			var err error

			if rows, err = query(db, prefs,
				`SELECT 
					SpellsPrepared.ID, Classes.Name, Description, CL, Concentration, PlusDomain, Special
				FROM SpellsPrepared
				LEFT JOIN Classes
					ON Classes.ID = SpellsPrepared.ClassID
				WHERE
					MonsterID=?
			`, mob_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var conc sql.NullInt32
				var desc, spec, cls sql.NullString
				var sb SpellBlock
				var sb_db_id int

				if monster.AC.Components == nil {
					monster.AC.Components = make(map[string]int)
				}
				if err := rows.Scan(&sb_db_id, &cls, &desc, &sb.CL, &conc, &sb.PlusDomain, &spec); err != nil {
					return err
				}
				if cls.Valid {
					sb.ClassName = cls.String
				} else {
					sb.ClassName = "SLA"
				}
				if desc.Valid {
					sb.Description = desc.String
				}
				if spec.Valid {
					sb.Special = spec.String
				}
				if conc.Valid {
					sb.Concentration = int(conc.Int32)
				} else {
					sb.NoConcentrationValue = true
				}

				if err = func() error {
					var rows *sql.Rows
					var err error

					if rows, err = query(db, prefs,
						`SELECT 
							SpellID, AlternateName, Frequency, Special, Spells.Name
						FROM SpellList
						LEFT JOIN Spells
							ON SpellList.SpellID = Spells.ID
						WHERE
							CollectionID=?
					`, sb_db_id); err != nil {
						return err
					}
					defer rows.Close()
					for rows.Next() {
						var alt, freq, spec sql.NullString
						var sp PreparedSpell
						var sp_db_id int

						if err := rows.Scan(&sp_db_id, &alt, &freq, &spec, &sp.Name); err != nil {
							return err
						}
						if alt.Valid {
							sp.AlternateName = alt.String
						}
						if freq.Valid {
							sp.Frequency = freq.String
						}
						if spec.Valid {
							sp.Special = spec.String
						}

						if err = func() error {
							var rows *sql.Rows
							var err error

							if rows, err = query(db, prefs,
								`SELECT
									SpellSlots.ID, IsCast, IsDomain
								FROM SpellSlots
								WHERE
									CollectionID = ? AND SpellID = ?
								ORDER BY
									Instance
							`, sb_db_id, sp_db_id); err != nil {
								return err
							}
							defer rows.Close()
							for rows.Next() {
								var slot SpellSlot
								var slot_db_id int

								if err := rows.Scan(&slot_db_id, &slot.IsCast, &slot.IsDomain); err != nil {
									return err
								}

								if err = func() error {
									var rows *sql.Rows
									var err error

									if rows, err = query(db, prefs,
										`SELECT Feats.Code
										FROM SpellSlotMeta
										LEFT JOIN Feats
											ON Feats.ID = MetaID
										WHERE
											SpellSlotMeta.SlotID=?
									`, slot_db_id); err != nil {
										return err
									}
									defer rows.Close()
									for rows.Next() {
										var f string
										if err := rows.Scan(&f); err != nil {
											return err
										}
										slot.MetaMagic = append(slot.MetaMagic, f)
									}
									return rows.Err()
								}(); err != nil {
									return err
								}

								sp.Slots = append(sp.Slots, slot)

							}
							return rows.Err()
						}(); err != nil {
							return err
						}

						sb.Spells = append(sb.Spells, sp)
					}
					return rows.Err()
				}(); err != nil {
					return err
				}

				monster.Spells = append(monster.Spells, sb)
			}
			return rows.Err()
		}(); err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(monster, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

type BaseLanguage struct {
	Language string
	IsLocal  bool `json:",omitempty"`
}

type Skill struct {
	Name             string
	ClassSkillFor    []string `json:",omitempty"`
	Ability          string
	HasArmorPenalty  bool   `json:",omitempty"`
	TrainingRequired bool   `json:",omitempty"`
	Source           string `json:",omitempty"`
	Description      string `json:",omitempty"`
	FullText         string `json:",omitempty"`
	ParentSkill      string `json:",omitempty"`
	IsVirtual        bool   `json:",omitempty"`
	IsBackground     bool   `json:",omitempty"`
	IsLocal          bool   `json:",omitempty"`
}

type Class struct {
	Code    string
	Name    string
	IsLocal bool `json:",omitempty"`
	Spells  struct {
		Type           string              `json:",omitempty"`
		Ability        string              `json:",omitempty"`
		HasBonusSpells bool                `json:",omitempty"`
		IsSpontaneous  bool                `json:",omitempty"`
		CastPerDay     [][]ClassSpellLevel `json:",omitempty"`
		PreparedPerDay [][]ClassSpellLevel `json:",omitempty"`
		SpellsKnown    [][]ClassSpellLevel `json:",omitempty"`
	} `json:",omitempty"`
}
type ClassSpellLevel struct {
	ClassLevel     int
	SpellLevel     int
	IsProhibited   bool `json:",omitempty"`
	IsUnlimitedUse bool `json:",omitempty"`
	Number         int  `json:",omitempty"`
}
type Feat struct {
	Code              string
	Name              string
	Parameters        string   `json:",omitempty"`
	IsLocal           bool     `json:",omitempty"`
	Description       string   `json:",omitempty"`
	Flags             uint64   `json:""`
	Prerequisites     string   `json:",omitempty"`
	Benefit           string   `json:",omitempty"`
	Normal            string   `json:",omitempty"`
	Special           string   `json:",omitempty"`
	Source            string   `json:",omitempty"`
	Race              string   `json:",omitempty"`
	Note              string   `json:",omitempty"`
	Goal              string   `json:",omitempty"`
	CompletionBenefit string   `json:",omitempty"`
	SuggestedTraits   string   `json:",omitempty"`
	Types             []string `json:",omitempty"`
	FlagNames         []string `json:"Flags,omitempty"`
	MetaMagic         struct {
		IsMetaMagicFeat     bool   `json:",omitempty"`
		Adjective           string `json:",omitempty"`
		LevelCost           int    `json:",omitempty"`
		IsLevelCostVariable bool   `json:",omitempty"`
		Symbol              string `json:",omitempty"`
	} `json:",omitempty"`
}

func exportSpells(fp *os.File, db *sql.DB, prefs *AppPreferences) error {
	var rows *sql.Rows
	var err error

	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("Exporting spell data")
	}
	if _, err = fp.WriteString(" \"Spells\": [\n"); err != nil {
		return err
	}

	componentList, err := getSpellComponents(db, prefs)
	if err != nil {
		return err
	}

	descList, err := getSpellDescriptors(db, prefs)
	if err != nil {
		return err
	}

	if rows, err = query(db, prefs,
		`SELECT 
			Spells.ID, IsLocal, Spells.Name, Schools.Code, Descriptors, Components, Material,
			Focus, CastingTime, CastingSpec, Range, Distance, DistPerLevel,
			DistSpec, Area, Effect, Targets, Duration, DurationTime,
			DurationSpec, DurationConc, DurationPerLvl, SR, SRSpec, SRObject,
			SRHarmless, SavingThrow, SaveEffect, SaveSpec, SaveObject, SaveHarmless,
			IsDismissible, IsDischarge, IsShapeable, HasCostlyComponents, SLALevel,
			Deity, Domain, Description, Source, MaterialCosts, Bloodline, Patron
		FROM Spells
		LEFT JOIN Schools
			ON Schools.ID=Spells.SchoolID
		`,
	); err != nil {
		return err
	}
	defer rows.Close()

	firstLine := true
	for rows.Next() {
		var spell Spell
		var material, focus, cspec, rang, distspec, area, effect, targs, dtime, dspec, srspec sql.NullString
		var save, seffect, sspec sql.NullString
		var deity, domain, desc, source, bline, patron sql.NullString
		var dist, dpl, slalvl, mcosts sql.NullInt32
		var spell_db_id, descriptors, components int

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if err := rows.Scan(&spell_db_id, &spell.IsLocal, &spell.Name, &spell.School, &descriptors, &components,
			&material, &focus, &spell.Casting.Time, &cspec, &rang, &dist, &dpl,
			&distspec, &area, &effect, &targs, &spell.Duration.Duration, &dtime, &dspec, &spell.Duration.Concentration, &spell.Duration.PerLevel,
			&spell.SR.SR, &srspec, &spell.SR.Object, &spell.SR.Harmless, &save, &seffect, &sspec, &spell.Save.Object, &spell.Save.Harmless,
			&spell.IsDismissible, &spell.IsDischarge, &spell.IsShapeable, &spell.Components.HasCostlyComponents, &slalvl,
			&deity, &domain, &desc, &source, &mcosts, &bline, &patron); err != nil {
			return err
		}
		if material.Valid {
			spell.Components.Material = material.String
		}
		if focus.Valid {
			spell.Components.Focus = focus.String
		}
		if mcosts.Valid {
			spell.Components.MaterialCosts = int(mcosts.Int32)
		}
		if cspec.Valid {
			spell.Casting.Special = cspec.String
		}
		if rang.Valid {
			spell.Range.Range = rang.String
		}
		if dist.Valid {
			spell.Range.Distance = int(dist.Int32)
		}
		if dpl.Valid {
			spell.Range.DistancePerLevel = int(dpl.Int32)
		}
		if distspec.Valid {
			spell.Range.DistanceSpecial = distspec.String
		}
		if area.Valid {
			spell.Effect.Area = area.String
		}
		if effect.Valid {
			spell.Effect.Effect = effect.String
		}
		if targs.Valid {
			spell.Effect.Targets = targs.String
		}
		if dtime.Valid {
			spell.Duration.Time = dtime.String
		}
		if dspec.Valid {
			spell.Duration.Special = dspec.String
		}
		if srspec.Valid {
			spell.SR.Special = srspec.String
		}
		if save.Valid {
			spell.Save.SavingThrow = save.String
		}
		if seffect.Valid {
			spell.Save.Effect = seffect.String
		}
		if sspec.Valid {
			spell.Save.Special = sspec.String
		}
		if deity.Valid {
			spell.Deity = deity.String
		}
		if domain.Valid {
			spell.Domain = domain.String
		}
		if desc.Valid {
			spell.Description = desc.String
		}
		if source.Valid {
			spell.Source = source.String
		}
		if bline.Valid {
			spell.Bloodline = bline.String
		}
		if patron.Valid {
			spell.Patron = patron.String
		}

		for bit, comp := range componentList {
			if (components & (1 << bit)) != 0 {
				spell.Components.Components = append(spell.Components.Components, comp)
			}
		}

		for bit, desc := range descList {
			if (descriptors & (1 << bit)) != 0 {
				spell.Descriptors = append(spell.Descriptors, desc)
			}
		}

		if prefs.SRD == spell.IsLocal {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("Skipping spell %s because SRD=%v", spell.Name, prefs.SRD)
			}
			continue
		}

		if err = func() error {
			var rows *sql.Rows
			var err error
			var cls string
			var lvl int

			if rows, err = query(db, prefs, `
				SELECT Name, Level 
				FROM SPellLevels 
				LEFT JOIN Classes
					ON Classes.ID = SpellLevels.ClassID
				WHERE SpellID=?
			`, spell_db_id); err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				if err = rows.Scan(&cls, &lvl); err != nil {
					return err
				}
				spell.ClassLevels = append(spell.ClassLevels, struct {
					Class string
					Level int
				}{Class: cls, Level: lvl})
			}
			return nil
		}(); err != nil {
			return err
		}

		if slalvl.Valid {
			spell.ClassLevels = append(spell.ClassLevels, struct {
				Class string
				Level int
			}{Class: "SLA", Level: int(slalvl.Int32)})
		}

		bytes, err := json.MarshalIndent(spell, "", "    ")
		if err != nil {
			return err
		}
		if _, err = fp.Write(bytes); err != nil {
			return err
		}
	}
	if _, err = fp.WriteString(" ],\n"); err != nil {
		return err
	}
	return rows.Err()
}

type Spell struct {
	IsLocal     bool `json:",omitempty"`
	Name        string
	School      string
	Descriptors []string `json:",omitempty"`
	Components  struct {
		Components          []string
		Material            string `json:",omitempty"`
		Focus               string `json:",omitempty"`
		HasCostlyComponents bool   `json:",omitempty"`
		MaterialCosts       int    `json:",omitempty"`
	}
	Casting struct {
		Time    string
		Special string `json:",omitempty"`
	}
	Range struct {
		Range            string
		Distance         int    `json:",omitempty"`
		DistancePerLevel int    `json:",omitempty"`
		DistanceSpecial  string `json:",omitempty"`
	}
	Effect struct {
		Area    string `json:",omitempty"`
		Effect  string `json:",omitempty"`
		Targets string `json:",omitempty"`
	}
	Duration struct {
		Duration      string
		Time          string `json:",omitempty"`
		Special       string `json:",omitempty"`
		Concentration bool   `json:",omitempty"`
		PerLevel      bool   `json:",omitempty"`
	}
	SR struct {
		SR       string `json:",omitempty"`
		Special  string `json:",omitempty"`
		Object   bool   `json:",omitempty"`
		Harmless bool   `json:",omitempty"`
	} `json:",omitempty"`
	Save struct {
		SavingThrow string `json:",omitempty"`
		Effect      string `json:",omitempty"`
		Special     string `json:",omitempty"`
		Object      bool   `json:",omitempty"`
		Harmless    bool   `json:",omitempty"`
	} `json:",omitempty"`
	IsDismissible bool `json:",omitempty"`
	IsDischarge   bool `json:",omitempty"`
	IsShapeable   bool `json:",omitempty"`
	ClassLevels   []struct {
		Class string
		Level int
	}
	Deity       string `json:",omitempty"`
	Domain      string `json:",omitempty"`
	Description string `json:",omitempty"`
	Source      string `json:",omitempty"`
	Bloodline   string `json:",omitempty"`
	Patron      string `json:",omitempty"`
}

type Monster struct {
	IsLocal   bool
	Species   string
	Code      string
	CR        string
	XP        int
	Class     string `json:",omitempty"`
	Alignment struct {
		Alignments []string
		Special    string `json:",omitempty"`
	}
	Source string `json:",omitempty"`
	Size   struct {
		Code      string
		SpaceText string `json:",omitempty"`
		ReachText string `json:",omitempty"`
	}
	Type       string
	Subtypes   []string `json:",omitempty"`
	Initiative struct {
		Mod     int
		Special string `json:",omitempty"`
	}
	Senses string `json:",omitempty"`
	Aura   string `json:",omitempty"`
	HP     struct {
		Typical int    `json:",omitempty"`
		Current int    `json:",omitempty"`
		Special string `json:",omitempty"`
		HitDice string
	}
	Save struct {
		Fort    SavingThrow
		Refl    SavingThrow
		Will    SavingThrow
		Special string `json:",omitempty"`
	}
	DefensiveAbilities string `json:",omitempty"`
	DR                 struct {
		DR     int    `json:",omitempty"`
		Bypass string `json:",omitempty"`
	} `json:",omitempty"`
	Immunities string `json:",omitempty"`
	Resists    string `json:",omitempty"`
	SR         struct {
		SR      int    `json:",omitempty"`
		Special string `json:",omitempty"`
	} `json:",omitempty"`
	Weaknesses string `json:",omitempty"`
	Speed      struct {
		Code    string
		Special string `json:",omitempty"`
	}
	SpecialAttacks string `json:",omitempty"`
	Abilities      struct {
		Str AbilityScore
		Dex AbilityScore
		Con AbilityScore
		Int AbilityScore
		Wis AbilityScore
		Cha AbilityScore
	}
	Combat struct {
		BAB        int
		CMB        int
		CMD        int
		CMBSpecial string `json:",omitempty"`
		CMDSpecial string `json:",omitempty"`
	}
	SQ           string `json:",omitempty"`
	Environment  string `json:",omitempty"`
	Organization string `json:",omitempty"`
	Treasure     string `json:",omitempty"`
	Appearance   string `json:",omitempty"`
	Group        string `json:",omitempty"`
	IsTemplate   bool   `json:",omitempty"`
	Strategy     struct {
		BeforeCombat string `json:",omitempty"`
		DuringCombat string `json:",omitempty"`
		Morale       string `json:",omitempty"`
	} `json:",omitempty"`
	IsCharacter       bool   `json:",omitempty"`
	IsCompanion       bool   `json:",omitempty"`
	IsUnique          bool   `json:",omitempty"`
	AgeCategory       string `json:",omitempty"`
	Gender            string `json:",omitempty"`
	Bloodline         string `json:",omitempty"`
	Patron            string `json:",omitempty"`
	AlternateNameForm string `json:",omitempty"`
	DontUseRacialHD   bool   `json:",omitempty"`
	VariantParent     string `json:",omitempty"`
	Mythic            struct {
		IsMythic bool `json:",omitempty"`
		MR       int  `json:",omitempty"`
		MT       int  `json:",omitempty"`
	} `json:",omitempty"`
	OffenseNote    string `json:",omitempty"`
	StatisticsNote string `json:",omitempty"`
	Gear           struct {
		Combat string `json:",omitempty"`
		Other  string `json:",omitempty"`
	} `json:",omitempty"`
	Schools struct {
		Focused    string `json:",omitempty"`
		Prohibited string `json:",omitempty"`
		Opposition string `json:",omitempty"`
	} `json:",omitempty"`
	ClassArchetypes string   `json:",omitempty"`
	BaseStatistics  string   `json:",omitempty"`
	RacialMods      string   `json:",omitempty"`
	Mystery         string   `json:",omitempty"`
	Notes           string   `json:",omitempty"`
	Domains         []string `json:",omitempty"`
	AC              struct {
		Components  map[string]int `json:",omitempty"`
		Adjustments map[string]int `json:",omitempty"`
	} `json:",omitempty"`
	AttackModes []AttackMode   `json:",omitempty"`
	Languages   []Language     `json:",omitempty"`
	Feats       []MonsterFeat  `json:",omitempty"`
	Skills      []MonsterSkill `json:",omitempty"`
	Spells      []SpellBlock   `json:",omitempty"`
}

type SpellBlock struct {
	ClassName            string
	CL                   int    `json:",omitempty"`
	Concentration        int    `json:",omitempty"`
	NoConcentrationValue bool   `json:",omitempty"`
	PlusDomain           int    `json:",omitempty"`
	Description          string `json:",omitempty"`
	Special              string `json:",omitempty"`
	Spells               []PreparedSpell
}

type PreparedSpell struct {
	Name          string
	AlternateName string `json:",omitempty"`
	Frequency     string `json:",omitempty"`
	Special       string `json:",omitempty"`
	Slots         []SpellSlot
}

type SpellSlot struct {
	IsCast    bool     `json:",omitempty"`
	IsDomain  bool     `json:",omitempty"`
	MetaMagic []string `json:",omitempty"`
}

type MonsterSkill struct {
	Code     string `json:",omitempty"`
	Modifier int    `json:",omitempty"`
	Notes    string `json:",omitempty"`
}

type MonsterFeat struct {
	Code       string `json:",omitempty"`
	Parameters string `json:",omitempty"`
	IsBonus    bool   `json:",omitempty"`
}
type Language struct {
	Name    string `json:",omitempty"`
	IsMute  bool   `json:",omitempty"`
	Special string `json:",omitempty"`
}

type Weapon struct {
	IsLocal  bool
	Code     string
	Cost     int
	Name     string
	Damage   map[string]string
	Critical struct {
		CantCritical bool `json:",omitempty"`
		Multiplier   string
		Threat       string
	}
	Ranged struct {
		Increment     int
		MaxIncrements int
		IsRanged      bool
	}
	Weight      int
	DamageTypes string
	Qualities   string
}

type AttackMode struct {
	Tier         int
	BaseWeaponID string `json:",omitempty"`
	Multiple     int
	Name         string
	Attack       string `json:",omitempty"`
	Damage       string `json:",omitempty"`
	Critical     struct {
		CantCritical bool   `json:",omitempty"`
		Threat       string `json:",omitempty"`
		Multiplier   string `json:",omitempty"`
	} `json:",omitempty"`
	Ranged struct {
		IsRanged      bool `json:",omitempty"`
		Increment     int  `json:",omitempty"`
		MaxIncrements int  `json:",omitempty"`
	} `json:",omitempty"`
	IsReach bool   `json:",omitempty"`
	Special string `json:",omitempty"`
	Mode    string `json:",omitempty"`
}

type SavingThrow struct {
	Mod     int
	Special string `json:",omitempty"`
}

type AbilityScore struct {
	Base      int    `json:",omitempty"`
	Special   string `json:",omitempty"`
	NullScore bool   `json:",omitempty"`
}

type AppPreferences struct {
	Prefs      util.GMAPreferences
	LogFile    string
	DebugFlags string
	ExportPath string
	ImportPath string
	SRD        bool
	TypeList   string
	DebugBits  DebugFlags
	TypeBits   TypeFilter
}

func configureApp() (AppPreferences, error) {
	var prefs AppPreferences

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
	prefs.SRD = Fsrd
	prefs.TypeList = Ftype

	return prefs, nil
}

/*
# @[00]@| Go-GMA 5.6.0
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
