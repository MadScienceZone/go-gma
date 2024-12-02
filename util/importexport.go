/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  _______     _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )(  ____ \   / ___   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  || (    \/   \/   )  |     #
# | |      | || || || (___) | Assistant | (____         /   )| (____         /   )     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   / (_____ \      _/   /      #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) )    /   _/       #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _ (   (__/\     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_)\_______/     #
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

package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

// CorePreferences holds preferences related to how import/export operations
// are filtered and executed.
type CorePreferences struct {
	// If true, the sense of FilterRegexp is reversed: import/export only if pattern NOT matched.
	FilterExclude bool

	// If true, import/export SRD data, otherwise local entries.
	SRD bool

	// These bits indicate what kind of debugging information to send to the log output.
	DebugBits DebugFlags

	// These bits filter the kinds of entries to be imported or exported.
	TypeBits TypeFilter

	// If non-nil, only entries matching this regular expression will be included.
	// For most data types, the regexp is matched against the Code and Name struct fields.
	// The entry is considered a match if either field contains text that matches the regexp.
	// For BaseLanguage, the Language field is matched. For monsters, the Code and Species
	// fields are checked.
	FilterRegexp *regexp.Regexp
}

// CoreImport is the main interface for importing data from JSON files into the GMA core database.
// The format of the input JSON file must be:
//   {
//      "GMA_Core_Database_Export_Version": <v>,
//      <type>: [
//         <objects of that type>, ...
//      ],
//      ...
//      "SRD": <srd_bool>
//   }
// The amount of whitespace between JSON elements is immaterial but GMA_Core_Database_Export_Version
// must appear first, and SRD must appear last.
//   <v>        ::= <integer> (file format version; currently must be 1)
//   <type>     ::= Bestiary | Classes | Feats | Languages | Skills | Spells | Weapons
//   <srd_bool> ::= true | false (true if importing/exporting SRD data; false for local entries)
//
// Given an open database connection db and a file fp open for reading, this will read through the
// JSON-encoded data from fp, calling the appropriate subordinate functions to handle the import of
// each data object found:
//   JSON Field  Go Type      Subordinate Function
//   Bestiary    Monster      ImportMonster
//   Classes     Class        ImportClass
//   Feats       Feat         ImportFeat
//   Languages   BaseLanguage ImportLanguage
//   Skills      Skill        ImportSkill
//   Spells      Spell        ImportSpell
//   Weapons     Weapon       ImportWeapon
//
// The prefs parameter specifies debugging flags which control what information is logged during the
// import operation, as well as filtering options which specify which subset of the file data to
// actually read into the database. The prefs.SRD field indicates whether the imported entries should
// be noted as public SRD data (if true) or as local entries (if false).
//
// Note that the SRD field in the JSON file merely indicates whether the data previously exported into that file
// was public SRD data or local entries; it is ignored by the CoreImport function.
//
func CoreImport(db *sql.DB, prefs *CorePreferences, fp io.Reader) error {
	var token json.Token
	var delim json.Delim
	var version float64
	var valid bool
	var fld string
	var err error
	decoder := json.NewDecoder(fp)

	if token, err = decoder.Token(); err != nil {
		return fmt.Errorf("error decoding import file: %v", err)
	}
	if delim, valid = token.(json.Delim); !valid || delim.String() != "{" {
		return fmt.Errorf("expected '{' not found in import file")
	}

	if token, err = decoder.Token(); err != nil {
		return fmt.Errorf("error decoding import file: %v", err)
	}
	if fld, valid = token.(string); !valid || fld != "GMA_Core_Database_Export_Version" {
		return fmt.Errorf("expected 'GMA_Core_Database_Export_version' field not found at start of import file")
	}

	if token, err = decoder.Token(); err != nil {
		return fmt.Errorf("error decoding import file: %v", err)
	}
	if version, valid = token.(float64); !valid {
		return fmt.Errorf("expected 'GMA_Core_Database_Export_version' value not missing or wrong type")
	}
	if version != 1 {
		return fmt.Errorf("File is version %v, which this version of gma coredb does not support.", version)
	}
	log.Printf("importing from version %v file", version)

	for decoder.More() {
		if token, err = decoder.Token(); err != nil {
			return fmt.Errorf("error decoding next section of import file: %v", err)
		}
		if fld, valid = token.(string); !valid {
			return fmt.Errorf("expected import type block name in import file")
		}

		if fld == "SRD" {
			break
		}

		if token, err = decoder.Token(); err != nil {
			return fmt.Errorf("error decoding next section of import file: %v", err)
		}
		if delim, valid = token.(json.Delim); !valid || delim.String() != "[" {
			return fmt.Errorf("missing '[' delimeter after '%s' in import file", fld)
		}

		if (prefs.DebugBits & DebugMisc) != 0 {
			log.Printf("importing %s records...", fld)
		}

		for decoder.More() {
			switch fld {
			case "Bestiary":
				err = ImportMonster(decoder, db, prefs)
			case "Classes":
				err = ImportClass(decoder, db, prefs)
			case "Feats":
				err = ImportFeat(decoder, db, prefs)
			case "Languages":
				err = ImportLanguage(decoder, db, prefs)
			case "Skills":
				err = ImportSkill(decoder, db, prefs)
			case "Spells":
				err = ImportSpell(decoder, db, prefs)
			case "Weapons":
				err = ImportWeapon(decoder, db, prefs)
			}
			if err != nil {
				return fmt.Errorf("unable to import %v type object: %v", fld, err)
			}

		}

		if token, err = decoder.Token(); err != nil {
			return fmt.Errorf("error decoding next section of import file: %v", err)
		}
		if delim, valid = token.(json.Delim); !valid || delim.String() != "]" {
			return fmt.Errorf("missing ']' delimeter at end of '%s' block in import file", fld)
		}
	}

	if token, err = decoder.Token(); err != nil {
		return fmt.Errorf("error decoding import file: %v", err)
	}
	if _, valid = token.(bool); !valid {
		return fmt.Errorf("expected boolean value for SRD field")
	}

	if token, err = decoder.Token(); err != nil {
		return fmt.Errorf("error decoding import file: %v", err)
	}
	if delim, valid = token.(json.Delim); !valid || delim.String() != "}" {
		return fmt.Errorf("expected final '}' not found in import file")
	}

	if decoder.More() {
		return fmt.Errorf("data after end of expected structure at end of import file")
	}
	return nil
}

// CoreExport is the main interface for exporting data to JSON files from the GMA core database.
// The JSON file will be written in the format documented for the CoreImport function.
//
// Given an open database connection db and a file fp open for writing, this will read through the
// database entries, calling the appropriate subordinate functions to handle the export of
// each data object found:
//   JSON Field  Go Type      Subordinate Function
//   Bestiary    Monster      ExportBestiary
//   Classes     Class        ExportClasses
//   Feats       Feat         ExportFeats
//   Languages   BaseLanguage ExportLanguages
//   Skills      Skill        ExportSkills
//   Spells      Spell        ExportSpells
//   Weapons     Weapon       ExportWeapons
//
// The prefs parameter specifies debugging flags which control what information is logged during the
// export operation, as well as filtering options which specify which subset of the core database to
// actually export. The prefs.SRD field indicates whether the public SRD records should be exported
// (if true) or if the local entries should be exported instead (if false).
//
func CoreExport(db *sql.DB, prefs *CorePreferences, fp *os.File) error {
	var err error

	if _, err = fp.WriteString("{\"GMA_Core_Database_Export_Version\": 1,\n"); err != nil {
		return fmt.Errorf("i/o error: %v", err)
	}
	if (prefs.TypeBits & TypeFeat) != 0 {
		if err = ExportFeats(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting feats: %v", err)
		}
	}
	if (prefs.TypeBits & TypeSpell) != 0 {
		if err = ExportSpells(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting spells: %v", err)
		}
	}
	if (prefs.TypeBits & TypeBestiary) != 0 {
		if err = ExportBestiary(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting bestiary: %v", err)
		}
	}
	if (prefs.TypeBits & TypeClass) != 0 {
		if err = ExportClasses(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting classes: %v", err)
		}
	}
	if (prefs.TypeBits & TypeLanguage) != 0 {
		if err = ExportLanguages(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting languages: %v", err)
		}
	}
	if (prefs.TypeBits & TypeSkill) != 0 {
		if err = ExportSkills(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting skills: %v", err)
		}
	}
	if (prefs.TypeBits & TypeWeapon) != 0 {
		if err = ExportWeapons(fp, db, prefs); err != nil {
			return fmt.Errorf("error exporting weapons: %v", err)
		}
	}
	if _, err = fmt.Fprintf(fp, " \"SRD\": %v\n}\n", prefs.SRD); err != nil {
		return fmt.Errorf("i/o error finishing file: %v", err)
	}

	return nil
}

type DebugFlags uint64

const (
	DebugQuery DebugFlags = 1 << iota
	DebugMisc
	DebugAll DebugFlags = 0xffffffff
)

// DebugFlagNameSlice takes a set of bit-encoded debug flags and returns a slice
// of strings representing the bits that were set. If collapse is true, the single
// string "all" will be the only element in the slice if all bits are set; otherwise
// the individual bit names will be returned.
func DebugFlagNameSlice(flags DebugFlags, collapse bool) []string {
	if flags == 0 {
		return nil
	}
	if collapse && flags == DebugAll {
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

// DebugFlagNames is like DebugFlagNameSlice but returns a single string
// value listing the bit names in angle brackets, separated by commas.
func DebugFlagNames(flags DebugFlags) string {
	list := DebugFlagNameSlice(flags, true)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

// NamedDebugFlags takes any number of string arguments which each name one of
// the DebugFlags bits. It returns the bit-encoded DebugFlags value with the named
// bits set.
func NamedDebugFlags(names ...string) (DebugFlags, error) {
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

// TypeFilterNameSlice takes a set of bit-encoded type filter flags and returns a slice
// of strings representing the bits that were set. If collapse is true, the single
// string "all" will be the only element in the slice if all bits are set; otherwise
// the individual bit names will be returned.
func TypeFilterNameSlice(types TypeFilter, collapse bool) []string {
	if types == 0 {
		return nil
	}
	if collapse && types == AllTypes {
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

// TypeFilterNames is like TypeFilterNameSlice but returns a single string
// value listing the bit names in angle brackets, separated by commas.
func TypeFilterNames(types TypeFilter) string {
	list := TypeFilterNameSlice(types, true)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

// NamedTypeFilters takes any number of string arguments which each name one of
// the TypeFilter bits. It returns the bit-encoded TypeFilter value with the named
// bits set.
func NamedTypeFilters(names ...string) (TypeFilter, error) {
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

//
//  ____    _  _____  _    ____    _    ____  _____
// |  _ \  / \|_   _|/ \  | __ )  / \  / ___|| ____|
// | | | |/ _ \ | | / _ \ |  _ \ / _ \ \___ \|  _|
// | |_| / ___ \| |/ ___ \| |_) / ___ \ ___) | |___
// |____/_/   \_\_/_/   \_\____/_/   \_\____/|_____|
//
// The GMA core database schema looks like this:
// *=primary key; #=bit-encoded integer referencing ID as bit number; >=foreign key
// not all field names are shown
//
//
//         SpellDescriptors(*ID)
//                           :   SpellComponents(*ID)
//                           :                   : CastingTimes(*CastingTime)
//                           :                   :               : Ranges(*Range)
//                           :                   :               :         : SpellResistances(*SpellResistance)
//                           :                   :               :         :      .................:
//                           :                   :               :         :      :  SavingThrows(*Save)
//                           :                   :               :         :      :       ...........:
//                           :....               :               :         :      :       : SaveEffects(*Effect)
//                               :               :               :         :      :       :              : Durations(*Duration)
//                               :               :               :         :      :       :              :             :
//       Spells(*ID, SchoolID, #Descriptors, #Components, >CastingTime, >Range, >SR, >SavingThrow, >SaveEffect, >Duration)
//               |          |
//               |          V
//               V Schools(*ID)
// SpellLevels(*>SpellID, *>ClassID)        [CC]
//               ^             |             :
//               |             +-------+     :
//              [S]                    V     :
//                            Classes(*ID, Code)
//                                     ^
//                                     |<----------------------+<---------------------------------------------+
//                                     |                       |                                              |
// SpellsPrepared(*ID, >MonsterID, >ClassID)   ClassSpells(>ClassID, ClassLevel, SpellLevel)  ClassMagic(>*ClassID, >MagicType, >Ability)
//                  ^       |                                                                                             :          :
//                  |       V                                                                                             :          :
//                  |      [M]                                                                               MagicTypes(*Type)       :
//                  |<------------------------------------------------------+                                         AbilityTypes(*Type)
//                  |                                                       |
// SpellList(*>CollectionID, *>SpellID, *Instance)    SpellSlots(*ID, >CollectionID, >SpellID, Instance)
//                                |                                |                      |
//                                V                                V                      V
//                               [S]             SpellSlotMeta(>SlotID, >MetaID)         [S]
//                                                                          |
//                                                                          V
//                                                            MetaMagic(*>FeatID)
//                                                                          |
//                                                                          V
//                                                                  Feats(*>ID, #Flags)
//                                                                        ^ |      :
//                                                                        | |      :.......
//                                                +-----------------------+ |             :
//                                                |                         V  FeatFlags(*ID)
//                                                |    FeatFeatTypes(*>FeatID, *>FeatType)
//                                                |                                  :
//                                                |                       FeatTypes(*FeatType)
//                                                |
//                    MonsterFeats(>MonsterID, >FeatID)
//                                     |
//                                     V
//                                    [M]
//
// MonsterLanguages(*>MonsterID, *>LanguageID)
//                      |              |
//                      V              V
//                     [M]  Languages(*ID)      [CC]
//                                               :
//                         Skills(*ID, #ClassSkillFor)
//                                 ^
//                                 |
// MonsterSkills(*>MonsterID, *>SkillID)
//                    |
//                    V
//                   [M]
//
//
//        Alignments(*ID)
//                    : Sizes(*Code)
//          [M]       :         : MonsterTypes(*Type)
//           |        :         :      ...........:
//           V        :         :      :
// Monsters(*ID, #Alignment, >Size, >Type, >AgeCategory,
//           ^                                   :
//           |                AgeCategories(*Category)
//           |
//           |           ACTypes(*ID)
//           +-------+            ^
//                   |            |
// ACAdjustments(*>MonsterID, *>ACType)
//                       ACCTypes(*ID)
//                                 ^
//                                 |
// ACComponents(*>MonsterID, *>ACComponent)
//                         MonsterSubtypes(*ID)
//                                          ^
//                             [M]          |
//                              ^           |
//                              |           |
// MonsterMonsterSubtypes(*>MonsterID, *>SubtypeID)
//
// AttackModes(*>MonsterID, *TierGroup, *TierSeq, >BaseWeaponID, >Mode)
//                                                     |           :
//                                                     V           :
//                                            Weapons(*ID)         :
//                                               AttackModeTypes(*Mode)
//
//                   [M]   Domains(*ID)
//                    ^             ^
//                    |             |
// MonsterDomains(*>MonsterID, *>DomainID)

// filterOut returns true if we should skip this entry. It also logs the reason why.
func filterOut(prefs *CorePreferences, entryType TypeFilter, entryTypeName, entryName, entryName2 string, isLocal bool) bool {
	if isLocal == prefs.SRD {
		if (prefs.DebugBits & DebugMisc) != 0 {
			log.Printf("skipping %s %s because -srd=%v", entryTypeName, entryName, prefs.SRD)
			return true
		}
	}

	if (prefs.TypeBits & entryType) == 0 {
		if (prefs.DebugBits & DebugMisc) != 0 {
			log.Printf("skipping %s %s because it's not in the requested type filter", entryTypeName, entryName)
		}
		return true
	}
	if prefs.FilterRegexp != nil {
		if prefs.FilterRegexp.MatchString(entryName) || prefs.FilterRegexp.MatchString(entryName2) {
			if prefs.FilterExclude {
				if (prefs.DebugBits & DebugMisc) != 0 {
					log.Printf("skipping %s %s because it matches filter regexp /%s/", entryTypeName, entryName, prefs.FilterRegexp.String())
				}
				return true // explicitly skip this
			} else {
				return false // explicitly DON'T skip this
			}
		}
		if !prefs.FilterExclude {
			if (prefs.DebugBits & DebugMisc) != 0 {
				log.Printf("skipping %s %s because it does not match filter regexp /%s/", entryTypeName, entryName, prefs.FilterRegexp.String())
			}
			return true // implicitly skip this
		}
	}
	return false
}

// query wraps the sql library's Query function but adds optional debug logging.
func query(db *sql.DB, prefs *CorePreferences, query string, args ...any) (*sql.Rows, error) {
	if (prefs.DebugBits & DebugQuery) != 0 {
		log.Printf("query: \"%s\" with %q", query, args)
	}
	rows, err := db.Query(query, args...)
	if (prefs.DebugBits&DebugQuery) != 0 && err != nil {
		log.Printf("query returned error: %v", err)
	}
	return rows, err
}

// getBitStrings performs the query q on the database, which must return an integer column and a string column.
// It returns a map of the integer values to the corresponding string value.
func getBitStrings(db *sql.DB, prefs *CorePreferences, q string) (map[int]string, error) {
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

// getStringBit performs the query q on the database, which must return an integer column and a string column.
// It returns a map of the string values to the corresponding integer value.
func getStringBits(db *sql.DB, prefs *CorePreferences, q string) (map[string]int, error) {
	flagList := make(map[string]int)

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
		flagList[name] = bit
	}
	return flagList, rows.Err()
}

// getStringList performs the query q on the database, which must return a string column. It
// returns a slice of those strings.
func getStringList(db *sql.DB, prefs *CorePreferences, q string) ([]string, error) {
	var list []string

	rows, err := query(db, prefs, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, name)
	}
	return list, rows.Err()
}

func getCastingTimeList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT CastingTime FROM CastingTimes")
}

func getRangeList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT Range FROM Ranges")
}

func getDurationList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT Duration FROM Durations")
}

func getSpellResistanceList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT SpellResistance FROM SpellResistances")
}

func getSavingThrowList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT Save FROM SavingThrows")
}

func getSaveEffectsList(db *sql.DB, prefs *CorePreferences) ([]string, error) {
	return getStringList(db, prefs, "SELECT Effect FROM SaveEffects")
}

func getAlignmentCodes(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Code FROM Alignments")
}

func getAlignmentCodeIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Code FROM Alignments")
}

func getFeatFlags(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Flag FROM FeatFlags")
}

func getSpellComponents(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Component FROM SpellComponents")
}

func getSpellComponentIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Component FROM SpellComponents")
}

func getSpellDescriptors(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Descriptor FROM SpellDescriptors")
}

func getSpellDescriptorIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Descriptor FROM SpellDescriptors")
}

func getClassCodes(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Code FROM Classes")
}

func getSkillCodes(db *sql.DB, prefs *CorePreferences) (map[int]string, error) {
	return getBitStrings(db, prefs, "SELECT ID, Code FROM Skills")
}

func getSkillIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Code FROM Skills")
}

func getClassIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Code FROM Classes")
}

func getClassNameIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Name FROM Classes")
}

func getSpellSchoolIDs(db *sql.DB, prefs *CorePreferences) (map[string]int, error) {
	return getStringBits(db, prefs, "SELECT ID, Code FROM Schools")
}

// makeRecordExist ensures that there is a row of the given table with the given value,
// inserting that data if necessary.
func makeRecordExist(db *sql.DB, prefs *CorePreferences, table, keyfield string, keyvalue any) error {
	exists, _, err := recordExists(db, prefs, table, "ID", keyfield, keyvalue)
	if exists || err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO "+table+" ("+keyfield+") VALUES (?)", keyvalue)
	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("added \"%v\" to %s", keyvalue, table)
	}
	return err
}

// makeRecordExistWithoutID ensures that there is a row of the given table with the given value,
// inserting that data if necessary. It does not require that there be an ID column.
func makeRecordExistWithoutID(db *sql.DB, prefs *CorePreferences, table, keyfield string, keyvalue any) error {
	exists, err := recordExistsWithoutID(db, prefs, table, keyfield, keyvalue)
	if exists || err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO "+table+" ("+keyfield+") VALUES (?)", keyvalue)
	if (prefs.DebugBits & DebugMisc) != 0 {
		log.Printf("added \"%v\" to %s", keyvalue, table)
	}
	return err
}

// recordExists checks to see if a row exists in the given table with a given column value. It returns a boolean
// indicating whether that row exists, and the ID column from the matching row if one was found.
func recordExists(db *sql.DB, prefs *CorePreferences, table, idfield, keyfield string, keyvalue any) (bool, int64, error) {
	// This shouldn't be a SQL injection risk because we generate and control the values used to build the query
	// and don't get them from an outside source.
	rows, err := query(db, prefs, "SELECT "+idfield+" FROM "+table+" WHERE "+keyfield+"=?", keyvalue)
	if err != nil {
		return false, 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return false, 0, err
		}
		return true, id, nil
	}
	return false, 0, nil
}

// recordExistsWithoutID checks to see if a row exists in the given table with a given column value. It returns a boolean
// indicating whether that row exists.
func recordExistsWithoutID(db *sql.DB, prefs *CorePreferences, table, keyfield string, keyvalue any) (bool, error) {
	// This shouldn't be a SQL injection risk because we generate and control the values used to build the query
	// and don't get them from an outside source.
	rows, err := query(db, prefs, "SELECT 1 FROM "+table+" WHERE "+keyfield+"=?", keyvalue)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

//   ____ _        _    ____ ____  _____ ____
//  / ___| |      / \  / ___/ ___|| ____/ ___|
// | |   | |     / _ \ \___ \___ \|  _| \___ \
// | |___| |___ / ___ \ ___) |__) | |___ ___) |
//  \____|_____/_/   \_\____/____/|_____|____/
//

// Class describes a character class.
type Class struct {
	Code    string
	Name    string
	IsLocal bool `json:",omitempty"`
	// If this class includes spellcasting capability, the details are here.
	Spells struct {
		// Magic type (arcane, etc)
		Type string `json:",omitempty"`
		// Ability score relevant to spells
		Ability string `json:",omitempty"`
		// Does this class allow bonus spells (e.g., domain spells)?
		HasBonusSpells bool `json:",omitempty"`
		// Is this a spontaneous (vs prepared) casting class?
		IsSpontaneous bool `json:",omitempty"`
		// A list of the number of spells which may be cast per day by class and spell level
		CastPerDay []ClassSpellLevel `json:",omitempty"`
		// A list of the number of spells which may be prepared per day by class and spell level (or empty list if that doesn't apply)
		PreparedPerDay []ClassSpellLevel `json:",omitempty"`
		// A list of the number of spells which may be known by class and spell level (or empty list if that doesn't apply)
		SpellsKnown []ClassSpellLevel `json:",omitempty"`
	} `json:",omitempty"`
}

// ClassSpellLevel describes a given spell-casting capability offered at a given class and spell level.
type ClassSpellLevel struct {
	ClassLevel int
	SpellLevel int
	// Is this level of spell even possible at this class level (or is this thing applicable at all)?
	IsProhibited bool `json:",omitempty"`
	// Is this level of spell unlimited (in terms of usages per day) at this class level?
	IsUnlimitedUse bool `json:",omitempty"`
	// The number of spells granted at this spell and class level.
	Number int `json:",omitempty"`
}

// ImportClass reads the next Class object from the open JSON data stream, writing it to
// the database.
func ImportClass(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var class Class
	var err, err2, err3 error
	var id int64
	var exists bool
	var res sql.Result

	if err = decoder.Decode(&class); err != nil {
		return err
	}

	if filterOut(prefs, TypeClass, "class", class.Name, class.Code, class.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Classes", "ID", "Code", class.Code); err != nil {
		return err
	}

	if exists {
		_, err2 = db.Exec(`DELETE FROM ClassSpells WHERE ClassID=?`, id)
		_, err3 = db.Exec(`DELETE FROM ClassMagic WHERE ClassID=?`, id)
		_, err = db.Exec(`UPDATE Classes SET Name=?, Code=?, IsLocal=? WHERE ID=?`, class.Name, class.Code, class.IsLocal, id)
	} else {
		if res, err = db.Exec(`INSERT INTO Classes (Name, Code, IsLocal) VALUES (?,?,?)`, class.Name, class.Code, class.IsLocal); err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing class %s (database id %d)", class.Name, id)
		} else {
			log.Printf("created new class %s (database id %d)", class.Name, id)
		}
	}

	// add spell capabilities
	if class.Spells.Type != "" || class.Spells.Ability != "" {
		var dcpd, dppd, dknown sql.NullInt32

		if _, err = db.Exec(`INSERT INTO ClassMagic (MagicType, Ability, Bonus, IsSpontaneous) VALUES (?,?,?,?)`,
			class.Spells.Type, class.Spells.Ability, class.Spells.HasBonusSpells, class.Spells.IsSpontaneous); err != nil {
			return err
		}
		if len(class.Spells.PreparedPerDay) != 0 && len(class.Spells.PreparedPerDay) != len(class.Spells.CastPerDay) {
			return fmt.Errorf("length of class %s prepared spells per day doesn't match cast per day table length", class.Name)
		}
		if len(class.Spells.SpellsKnown) != 0 && len(class.Spells.SpellsKnown) != len(class.Spells.CastPerDay) {
			return fmt.Errorf("length of class %s spells known table doesn't match cast per day table length", class.Name)
		}
		classLevel := 1
		sl := 0
		for i, cpd := range class.Spells.CastPerDay {
			if classLevel != cpd.ClassLevel || sl != cpd.SpellLevel {
				return fmt.Errorf("class %s cast-per-day table entry for CL %d, SL %d expected, but CL %d, SL %d found instead",
					class.Name, classLevel, sl, cpd.ClassLevel, cpd.SpellLevel)
			}

			dppd.Valid = false
			if len(class.Spells.PreparedPerDay) != 0 {
				if class.Spells.PreparedPerDay[i].ClassLevel != cpd.ClassLevel || class.Spells.PreparedPerDay[i].SpellLevel != cpd.SpellLevel {
					return fmt.Errorf("class %s prepared per day table entry %d (CL %d, SL %d) is out of sync (should be CL %d, SL %d)",
						class.Name, i, class.Spells.PreparedPerDay[i].ClassLevel, class.Spells.PreparedPerDay[i].SpellLevel, classLevel, sl)
				}
				if !class.Spells.PreparedPerDay[i].IsProhibited {
					dppd.Valid = true
					if class.Spells.PreparedPerDay[i].IsUnlimitedUse {
						dppd.Int32 = -1
					} else {
						dppd.Int32 = int32(class.Spells.PreparedPerDay[i].Number)
					}
				}
			}

			dknown.Valid = false
			if len(class.Spells.SpellsKnown) != 0 {
				if class.Spells.SpellsKnown[i].ClassLevel != cpd.ClassLevel || class.Spells.SpellsKnown[i].SpellLevel != cpd.SpellLevel {
					return fmt.Errorf("class %s spells known table entry %d (CL %d, SL %d) is out of sync (should be CL %d, SL %d)",
						class.Name, i, class.Spells.SpellsKnown[i].ClassLevel, class.Spells.SpellsKnown[i].SpellLevel, classLevel, sl)
				}
				if !class.Spells.SpellsKnown[i].IsProhibited {
					dknown.Valid = true
					if class.Spells.SpellsKnown[i].IsUnlimitedUse {
						dknown.Int32 = -1
					} else {
						dknown.Int32 = int32(class.Spells.SpellsKnown[i].Number)
					}
				}
			}

			if cpd.IsProhibited {
				dcpd.Valid = false
			} else {
				dcpd.Valid = true
				if cpd.IsUnlimitedUse {
					dcpd.Int32 = -1
				} else {
					dcpd.Int32 = int32(cpd.Number)
				}
			}

			if _, err = db.Exec(`INSERT INTO ClassSpells (ClassID, ClassLevel, SpellLevel, CastPerDay, PrepPerDay, Known) VALUES (?,?,?,?,?,?)`,
				id, cpd.ClassLevel, cpd.SpellLevel, dcpd, dppd, dknown); err != nil {
				return err
			}

			if sl++; sl > 9 {
				sl = 0
				classLevel++
			}
		}
	}

	return err
}

// ExportClasses exports all classes from the database to the open JSON file,
// subect to any filtering options specified in prefs
func ExportClasses(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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

		if err := rows.Scan(&cls_db_id, &cls.Name, &cls.Code, &mtype, &abil, &bonus, &spon); err != nil {
			return err
		}

		if filterOut(prefs, TypeClass, "class", cls.Name, cls.Code, cls.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
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
				for rows.Next() {
					var cl, sl int
					var cpd, ppd, kn sql.NullInt32
					var c, p, k ClassSpellLevel

					if err := rows.Scan(&cl, &sl, &cpd, &ppd, &kn); err != nil {
						return err
					}
					c.ClassLevel = cl
					c.SpellLevel = sl
					if cpd.Valid {
						if cpd.Int32 < 0 {
							c.IsUnlimitedUse = true
						} else {
							c.Number = int(cpd.Int32)
						}
					} else {
						c.IsProhibited = true
					}

					p.ClassLevel = cl
					p.SpellLevel = sl
					if ppd.Valid {
						if ppd.Int32 < 0 {
							p.IsUnlimitedUse = true
						} else {
							p.Number = int(ppd.Int32)
						}
					} else {
						p.IsProhibited = true
					}

					k.ClassLevel = cl
					k.SpellLevel = sl
					if kn.Valid {
						if kn.Int32 < 0 {
							k.IsUnlimitedUse = true
						} else {
							k.Number = int(kn.Int32)
						}
					} else {
						k.IsProhibited = true
					}

					cls.Spells.CastPerDay = append(cls.Spells.CastPerDay, c)
					cls.Spells.PreparedPerDay = append(cls.Spells.PreparedPerDay, p)
					cls.Spells.SpellsKnown = append(cls.Spells.SpellsKnown, k)
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

//  _____ _____    _  _____ ____
// |  ___| ____|  / \|_   _/ ___|
// | |_  |  _|   / _ \ | | \___ \
// |  _| | |___ / ___ \| |  ___) |
// |_|   |_____/_/   \_\_| |____/
//

// Feat describes each feat that is in play for the game.
type Feat struct {
	Code string
	Name string
	// If there are parameters allowed for this feat, this describes them.
	Parameters        string   `json:",omitempty"`
	IsLocal           bool     `json:",omitempty"`
	Description       string   `json:",omitempty"`
	Flags             uint64   `json:"-"`
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
	// If this is a metamagic feat, the following will apply.
	MetaMagic struct {
		IsMetaMagicFeat     bool   `json:",omitempty"`
		Adjective           string `json:",omitempty"`
		LevelCost           int    `json:",omitempty"`
		IsLevelCostVariable bool   `json:",omitempty"`
		// Symbol to place in checkboxes on character sheets, if any.
		Symbol string `json:",omitempty"`
	} `json:",omitempty"`
}

// ImportFeat reads the next Feat object from the JSON stream, writing it to the database.
func ImportFeat(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var feat Feat
	var err, err2, err3 error
	var id int64
	var exists bool
	var res sql.Result
	var flags int

	if err = decoder.Decode(&feat); err != nil {
		return err
	}

	if filterOut(prefs, TypeFeat, "feat", feat.Name, feat.Code, feat.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Feats", "ID", "Code", feat.Code); err != nil {
		return err
	}

	fflags, err := getFeatFlags(db, prefs)
	if err != nil {
		return err
	}

	for _, f := range feat.FlagNames {
		for bit, fname := range fflags {
			if fname == f {
				flags |= bit
			}
		}
	}

	var params, note, goal, comp, traits sql.NullString
	var prereq, benefit, norm, spec, src, race sql.NullString
	if feat.Prerequisites == "" {
		prereq.Valid = false
	} else {
		prereq.Valid = true
		prereq.String = feat.Prerequisites
	}
	if feat.Benefit == "" {
		benefit.Valid = false
	} else {
		benefit.Valid = true
		benefit.String = feat.Benefit
	}
	if feat.Normal == "" {
		norm.Valid = false
	} else {
		norm.Valid = true
		norm.String = feat.Normal
	}
	if feat.Special == "" {
		spec.Valid = false
	} else {
		spec.Valid = true
		spec.String = feat.Special
	}
	if feat.Source == "" {
		src.Valid = false
	} else {
		src.Valid = true
		src.String = feat.Source
	}
	if feat.Race == "" {
		race.Valid = false
	} else {
		race.Valid = true
		race.String = feat.Race
	}

	if feat.Parameters == "" {
		params.Valid = false
	} else {
		params.Valid = true
		params.String = feat.Parameters
	}
	if feat.Note == "" {
		note.Valid = false
	} else {
		note.Valid = true
		note.String = feat.Note
	}
	if feat.Goal == "" {
		goal.Valid = false
	} else {
		goal.Valid = true
		goal.String = feat.Goal
	}
	if feat.CompletionBenefit == "" {
		comp.Valid = false
	} else {
		comp.Valid = true
		comp.String = feat.CompletionBenefit
	}
	if feat.SuggestedTraits == "" {
		traits.Valid = false
	} else {
		traits.Valid = true
		traits.String = feat.SuggestedTraits
	}

	if exists {
		_, err2 = db.Exec(`DELETE FROM MetaMagic WHERE FeatID=?`, id)
		_, err3 = db.Exec(`DELETE FROM FeatFeatTypes WHERE FeatID=?`, id)
		_, err = db.Exec(`UPDATE Feats SET Code=?, Name=?, Parameters=?, IsLocal=?, Description=?, Flags=?, Prerequisites=?,
			Benefit=?, Normal=?, Special=?, Source=?, Race=?, Note=?, Goal=?, CompletionBenefit=?, SuggestedTraits=?
			WHERE ID=?`,
			feat.Code, feat.Name, params, feat.IsLocal, feat.Description, flags, prereq, benefit, norm, spec, src, race,
			note, goal, comp, traits, id)
	} else {
		if res, err = db.Exec(`INSERT INTO Feats
			(Code, Name, Parameters, IsLocal, Description, Flags, Prerequisites,
			 Benefit, Normal, Special, Source, Race, Note, Goal, CompletionBenefit, SuggestedTraits)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			feat.Code, feat.Name, params, feat.IsLocal, feat.Description, flags, prereq, benefit, norm, spec, src, race,
			note, goal, comp, traits); err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing feat %s (database id %d)", feat.Name, id)
		} else {
			log.Printf("created new feat %s (database id %d)", feat.Name, id)
		}
	}

	// add metamagic
	if feat.MetaMagic.Adjective != "" || feat.MetaMagic.Symbol != "" {
		var cost sql.NullInt32
		var sym sql.NullString

		if feat.MetaMagic.IsLevelCostVariable {
			cost.Valid = false
		} else {
			cost.Valid = true
			cost.Int32 = int32(feat.MetaMagic.LevelCost)
		}

		if feat.MetaMagic.Symbol == "" {
			sym.Valid = false
		} else {
			sym.Valid = true
			sym.String = feat.MetaMagic.Symbol
		}

		if _, err = db.Exec(`INSERT INTO MetaMagic (Adjective, LevelCost, Symbol, FeatID) VALUES (?,?,?,?)`,
			feat.MetaMagic.Adjective, cost, sym, id); err != nil {
			return err
		}
	}

	// add feat types
	for _, ftype := range feat.Types {
		if err = makeRecordExistWithoutID(db, prefs, "FeatTypes", "FeatType", ftype); err != nil {
			return err
		}
		if _, err = db.Exec(`INSERT INTO FeatFeatTypes (FeatID, FeatType) VALUES (?, ?)`, id, ftype); err != nil {
			return err
		}
	}

	return err
}

// ExportFeats reads all Feats from the database, writing them to the open JSON file.
func ExportFeats(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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

		if err := rows.Scan(&feat_db_id, &feat.Code, &feat.Name, &param, &feat.IsLocal, &feat.Description,
			&feat.Flags, &prereq, &benefit, &normal, &special, &source, &race,
			&note, &goal, &comp, &traits,
			&adj, &levelcost, &sym); err != nil {
			return err
		}

		if filterOut(prefs, TypeFeat, "feat", feat.Name, feat.Code, feat.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
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

// __        _______    _    ____   ___  _   _ ____
// \ \      / / ____|  / \  |  _ \ / _ \| \ | / ___|
//  \ \ /\ / /|  _|   / _ \ | |_) | | | |  \| \___ \
//   \ V  V / | |___ / ___ \|  __/| |_| | |\  |___) |
//    \_/\_/  |_____/_/   \_\_|    \___/|_| \_|____/
//

// Weapon describes a weapon in the core data.
type Weapon struct {
	IsLocal bool `json:",omitempty"`
	Code    string
	// Cost is in units of copper pieces.
	Cost int `json:",omitempty"`
	Name string
	// Damage maps a size code to the damage done for a weapon of that size.
	Damage   map[string]string
	Critical struct {
		// If true, no critical information is available / weapon can't inflict critical damage.
		CantCritical bool `json:",omitempty"`
		Multiplier   int  `json:",omitempty"`
		Threat       int  `json:",omitempty"`
	}
	Ranged struct {
		Increment     int  `json:",omitempty"`
		MaxIncrements int  `json:",omitempty"`
		IsRanged      bool `json:",omitempty"`
	} `json:",omitempty"`
	// Weight is in units of grams.
	Weight int `json:",omitempty"`
	// DamageTypes is a bitmap of types of damage done by this weapon
	DamageTypes WeaponDamageType `json:",omitempty"`
	// Qualities is a bitmap of weapon qualities
	Qualities WeaponQuality `json:",omitempty"`
}

type WeaponDamageType byte

// DO NOT add new items in the middle of this list without completely rebuilding
// the core database; adding to the end is fine.
const (
	BludgeoningDamage WeaponDamageType = 1 << iota
	PiercingDamage
	SlashingDamage
)

// MarshalJSON represents the bitmapped set of weapon damage
// types as a list of strings so the JSON version is easily
// understood and managed by other programs.
func (w WeaponDamageType) MarshalJSON() ([]byte, error) {
	var damageList []string
	if w&BludgeoningDamage != 0 {
		damageList = append(damageList, "B")
	}
	if w&PiercingDamage != 0 {
		damageList = append(damageList, "P")
	}
	if w&SlashingDamage != 0 {
		damageList = append(damageList, "S")
	}

	return json.Marshal(damageList)
}

func (w *WeaponDamageType) UnmarshalJSON(input []byte) error {
	var damageList []string
	if err := json.Unmarshal(input, &damageList); err != nil {
		return err
	}
	*w = 0
	for _, t := range damageList {
		switch t {
		case "B":
			*w |= BludgeoningDamage
		case "P":
			*w |= PiercingDamage
		case "S":
			*w |= SlashingDamage
		default:
			return fmt.Errorf("cannot unmarshal value \"%s\" to type WeaponDamageType", t)
		}
	}
	return nil
}

type WeaponQuality uint32

// DO NOT add new items in the middle of this list without completely rebuilding
// the core database; adding to the end is fine.
const (
	BraceWeapon WeaponQuality = 1 << iota
	DeadlyWeapon
	DisarmWeapon
	DoubleWeapon
	DwarvenWeapon
	ElvenWeapon
	ExoticWeapon
	FragileWeapon
	GnomeWeapon
	GrappleWeapon
	HalflingWeapon
	LightWeapon
	MartialWeapon
	MasterworkWeapon
	MonkWeapon
	NonLethalWeapon
	OneHandedWeapon
	OrcWeapon
	RangedWeapon
	ReachWeapon
	SimpleWeapon
	TripWeapon
	TwoHandedWeapon
	UnarmedWeapon
)

// MarshalJSON represents the bitmapped set of weapon qualities
// as a list of strings so the JSON version is easily
// understood and managed by other programs.
func (q WeaponQuality) MarshalJSON() ([]byte, error) {
	var damageList []string
	if q&OneHandedWeapon != 0 {
		damageList = append(damageList, "1")
	}
	if q&TwoHandedWeapon != 0 {
		damageList = append(damageList, "2")
	}
	if q&BraceWeapon != 0 {
		damageList = append(damageList, "b")
	}
	if q&DisarmWeapon != 0 {
		damageList = append(damageList, "D")
	}
	if q&DoubleWeapon != 0 {
		damageList = append(damageList, "d")
	}
	if q&ElvenWeapon != 0 {
		damageList = append(damageList, "E")
	}
	if q&FragileWeapon != 0 {
		damageList = append(damageList, "f")
	}
	if q&GnomeWeapon != 0 {
		damageList = append(damageList, "G")
	}
	if q&GrappleWeapon != 0 {
		damageList = append(damageList, "g")
	}
	if q&HalflingWeapon != 0 {
		damageList = append(damageList, "H")
	}
	if q&LightWeapon != 0 {
		damageList = append(damageList, "L")
	}
	if q&MonkWeapon != 0 {
		damageList = append(damageList, "M")
	}
	if q&MartialWeapon != 0 {
		damageList = append(damageList, "m")
	}
	if q&OrcWeapon != 0 {
		damageList = append(damageList, "O")
	}
	if q&RangedWeapon != 0 {
		damageList = append(damageList, "R")
	}
	if q&ReachWeapon != 0 {
		damageList = append(damageList, "r")
	}
	if q&NonLethalWeapon != 0 {
		damageList = append(damageList, "S")
	}
	if q&SimpleWeapon != 0 {
		damageList = append(damageList, "s")
	}
	if q&TripWeapon != 0 {
		damageList = append(damageList, "t")
	}
	if q&UnarmedWeapon != 0 {
		damageList = append(damageList, "U")
	}
	if q&DwarvenWeapon != 0 {
		damageList = append(damageList, "W")
	}
	if q&ExoticWeapon != 0 {
		damageList = append(damageList, "X")
	}
	if q&DeadlyWeapon != 0 {
		damageList = append(damageList, "x")
	}
	if q&MasterworkWeapon != 0 {
		damageList = append(damageList, "Z")
	}

	return json.Marshal(damageList)
}

func (q *WeaponQuality) UnmarshalJSON(input []byte) error {
	var qualityList []string
	if err := json.Unmarshal(input, &qualityList); err != nil {
		return err
	}
	*q = 0
	for _, t := range qualityList {
		switch t {
		case "1":
			*q |= OneHandedWeapon
		case "2":
			*q |= TwoHandedWeapon
		case "b":
			*q |= BraceWeapon
		case "D":
			*q |= DisarmWeapon
		case "d":
			*q |= DoubleWeapon
		case "E":
			*q |= ElvenWeapon
		case "f":
			*q |= FragileWeapon
		case "G":
			*q |= GnomeWeapon
		case "g":
			*q |= GrappleWeapon
		case "H":
			*q |= HalflingWeapon
		case "L":
			*q |= LightWeapon
		case "M":
			*q |= MonkWeapon
		case "m":
			*q |= MartialWeapon
		case "O":
			*q |= OrcWeapon
		case "R":
			*q |= RangedWeapon
		case "r":
			*q |= ReachWeapon
		case "S":
			*q |= NonLethalWeapon
		case "s":
			*q |= SimpleWeapon
		case "t":
			*q |= TripWeapon
		case "U":
			*q |= UnarmedWeapon
		case "W":
			*q |= DwarvenWeapon
		case "X":
			*q |= ExoticWeapon
		case "x":
			*q |= DeadlyWeapon
		case "Z":
			*q |= MasterworkWeapon
		default:
			return fmt.Errorf("cannot unmarshal value \"%s\" to type WeaponQuality", t)
		}
	}
	return nil
}

// ImportWeapon reads a weapon from the JSON input stream, writing it to the database.
func ImportWeapon(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var weap Weapon
	var err error
	var id int64
	var exists, ok bool
	var res sql.Result
	var cost, ri, rmax, wt, cm, ct, dtype, q sql.NullInt32
	var dt, ds, dm, dl sql.NullString
	var s string

	if err = decoder.Decode(&weap); err != nil {
		return err
	}

	if filterOut(prefs, TypeWeapon, "weapon", weap.Name, weap.Code, weap.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Weapons", "ID", "Code", weap.Code); err != nil {
		return err
	}

	if weap.Cost == 0 {
		cost.Valid = false
	} else {
		cost.Valid = true
		cost.Int32 = int32(weap.Cost)
	}
	if weap.Weight == 0 {
		wt.Valid = false
	} else {
		wt.Valid = true
		wt.Int32 = int32(weap.Weight)
	}
	if weap.Ranged.IsRanged {
		ri.Valid = true
		ri.Int32 = int32(weap.Ranged.Increment)
		rmax.Valid = true
		rmax.Int32 = int32(weap.Ranged.MaxIncrements)
	} else {
		ri.Valid = false
		rmax.Valid = false
	}
	if weap.Critical.CantCritical {
		cm.Valid = false
		ct.Valid = false
	} else {
		cm.Valid = true
		cm.Int32 = int32(weap.Critical.Multiplier)
		ct.Valid = true
		ct.Int32 = int32(weap.Critical.Threat)
	}
	if s, ok = weap.Damage["T"]; !ok || s == "" {
		dt.Valid = false
	} else {
		dt.Valid = true
		dt.String = s
	}
	if s, ok = weap.Damage["S"]; !ok || s == "" {
		ds.Valid = false
	} else {
		ds.Valid = true
		ds.String = s
	}
	if s, ok = weap.Damage["M"]; !ok || s == "" {
		dm.Valid = false
	} else {
		dm.Valid = true
		dm.String = s
	}
	if s, ok = weap.Damage["L"]; !ok || s == "" {
		dl.Valid = false
	} else {
		dl.Valid = true
		dl.String = s
	}
	if weap.DamageTypes == 0 {
		dtype.Valid = false
	} else {
		dtype.Valid = true
		dtype.Int32 = int32(weap.DamageTypes)
	}
	if weap.Qualities == 0 {
		q.Valid = false
	} else {
		q.Valid = true
		q.Int32 = int32(weap.Qualities)
	}

	if exists {
		_, err = db.Exec(`UPDATE Weapons SET
							IsLocal=?, Code=?, Cost=?, Name=?, DmgT=?, DmgS=?, DmgM=?, DmgL=?, CritMultiplier=?,
							CritThreat=?, RangeIncrement=?, RangeMax=?, Weight=?, DmgTypes=?, Qualities=?
						WHERE ID=?`,
			weap.IsLocal, weap.Code, cost, weap.Name, dt, ds, dm, dl, cm, ct, ri, rmax, wt, dtype, q,
			id)
	} else {
		if res, err = db.Exec(`INSERT INTO Weapons
								(IsLocal, Code, Cost, Name, DmgT, DmgS, DmgM, DmgL, CritMultiplier,
								CritThreat, RangeIncrement, RangeMax, Weight, DmgTypes, Qualities)
							VALUES
								(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			weap.IsLocal, weap.Code, cost, weap.Name, dt, ds, dm, dl, cm, ct, ri, rmax, wt, dtype, q); err != nil {
			return err
		}
		id, err = res.LastInsertId()
	}

	if err != nil {
		return err
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing weapon %s (database id %d)", weap.Name, id)
		} else {
			log.Printf("created new weapon %s (database id %d)", weap.Name, id)
		}
	}

	return err
}

// ExportWeapons exports all weapons from the database to the JSON file fp.
func ExportWeapons(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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
		var cost, ri, rmax, wt, cm, ct, dtyp, q sql.NullInt32
		var dt, ds, dm, dl sql.NullString

		if err := rows.Scan(&weap.IsLocal, &weap.Code, &cost, &weap.Name,
			&dt, &ds, &dm, &dl, &cm, &ct, &ri, &rmax, &wt, &dtyp, &q); err != nil {
			return err
		}

		if filterOut(prefs, TypeWeapon, "weapon", weap.Name, weap.Code, weap.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		weap.Damage = make(map[string]string)

		if wt.Valid {
			weap.Weight = int(wt.Int32)
		}
		if dtyp.Valid {
			weap.DamageTypes = WeaponDamageType(dtyp.Int32)
		}
		if q.Valid {
			weap.Qualities = WeaponQuality(q.Int32)
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
				weap.Critical.Threat = int(ct.Int32)
			}
			if cm.Valid {
				weap.Critical.Multiplier = int(cm.Int32)
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

//  ____  _  _____ _     _     ____
// / ___|| |/ /_ _| |   | |   / ___|
// \___ \| ' / | || |   | |   \___ \
//  ___) | . \ | || |___| |___ ___) |
// |____/|_|\_\___|_____|_____|____/
//

// Skill describes a skill that any creature might have.
type Skill struct {
	Name string
	Code string
	// List of the class Code strings for those classes which have this skill as a class skill.
	ClassSkillFor []string `json:",omitempty"`
	// Relevant ability score name
	Ability          string
	HasArmorPenalty  bool   `json:",omitempty"`
	TrainingRequired bool   `json:",omitempty"`
	Source           string `json:",omitempty"`
	Description      string `json:",omitempty"`
	FullText         string `json:",omitempty"`
	// If this is a sub-skill, the skill code for its parent (e.g., "craft" for "craft.alchemy")
	ParentSkill string `json:",omitempty"`
	// If this is a parent skill but ONLY its children should be instantiated, not this skill itself.
	IsVirtual    bool `json:",omitempty"`
	IsBackground bool `json:",omitempty"`
	IsLocal      bool `json:",omitempty"`
}

// ImportSkill reads a skill from the JSON stream and writes it to the database.
func ImportSkill(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var sk Skill
	var err error
	var id int64
	var exists bool

	if err = decoder.Decode(&sk); err != nil {
		return err
	}

	if filterOut(prefs, TypeSkill, "skill", sk.Name, sk.Name, sk.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Skills", "ID", "Code", sk.Code); err != nil {
		return err
	}

	var classes int
	var ps sql.NullInt32
	var src sql.NullString

	classIDs, err := getClassIDs(db, prefs)
	if err != nil {
		return err
	}

	skillIDs, err := getSkillIDs(db, prefs)
	if err != nil {
		return err
	}

	for _, class := range sk.ClassSkillFor {
		if bit, ok := classIDs[class]; ok {
			classes |= 1 << bit
		} else {
			return fmt.Errorf("cannot add unknown class code \"%s\" to skill \"%s\"", class, sk.Name)
		}
	}
	if sk.ParentSkill == "" {
		ps.Valid = false
	} else {
		if pid, ok := skillIDs[sk.ParentSkill]; ok {
			ps.Valid = true
			ps.Int32 = int32(pid)
		} else {
			return fmt.Errorf("cannot add unknown skill code \"%s\" as parent of skill \"%s\"", sk.ParentSkill, sk.Name)
		}
	}

	if sk.Source == "" {
		src.Valid = false
	} else {
		src.Valid = true
		src.String = sk.Source
	}

	if exists {
		_, err = db.Exec(`
			UPDATE Skills 
			SET 
				Code=?, Name=?, Classes=?, Ability=?, ArmorCheck=?, TrainedOnly=?,
				Source=?, Description=?, FullText=?, ParentSkill=?, IsVirtual=?,
				IsLocal=?, IsBackground=?
			WHERE
				ID=?`,
			sk.Code, sk.Name, classes, sk.Ability, sk.HasArmorPenalty, sk.TrainingRequired,
			src, sk.Description, sk.FullText, ps, sk.IsVirtual, sk.IsLocal, sk.IsBackground, id)
	} else {
		_, err = db.Exec(`
			INSERT INTO Skills
				(Code, Name, Classes, Ability, ArmorCheck, TrainedOnly, Source,
				Description, FullText, ParentSkill, IsVirtual, IsLocal, IsBackground)
			VALUES
				(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			sk.Code, sk.Name, classes, sk.Ability, sk.HasArmorPenalty, sk.TrainingRequired,
			src, sk.Description, sk.FullText, ps, sk.IsVirtual, sk.IsLocal, sk.IsBackground)
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing skill %s (database id %d)", sk.Name, id)
		} else {
			log.Printf("created new skill %s", sk.Name)
		}
	}
	return err
}

// ExportSkills exports all the skill data from the database to the open JSON file.
func ExportSkills(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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
	skillCodes, err := getSkillCodes(db, prefs)
	if err != nil {
		return err
	}
	if rows, err = query(db, prefs,
		`SELECT 
			ID, Name, Code, Classes, Ability, ArmorCheck, TrainedOnly, Source,
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

		if err := rows.Scan(&skill_db_id, &sk.Name, &sk.Code, &classbits, &sk.Ability,
			&sk.HasArmorPenalty, &sk.TrainingRequired, &source, &sk.Description,
			&sk.FullText, &parent, &sk.IsVirtual, &sk.IsLocal, &sk.IsBackground); err != nil {
			return err
		}

		if filterOut(prefs, TypeSkill, "skill", sk.Name, sk.Name, sk.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
		}

		if source.Valid {
			sk.Source = source.String
		}
		if parent.Valid {
			if pskill, ok := skillCodes[int(parent.Int32)]; ok {
				sk.ParentSkill = pskill
			} else {
				return fmt.Errorf("skill \"%s\" refers to undefined parent skill ID %d", sk.Name, parent.Int32)
			}
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

//  _        _    _   _  ____ _   _   _    ____ _____ ____
// | |      / \  | \ | |/ ___| | | | / \  / ___| ____/ ___|
// | |     / _ \ |  \| | |  _| | | |/ _ \| |  _|  _| \___ \
// | |___ / ___ \| |\  | |_| | |_| / ___ \ |_| | |___ ___) |
// |_____/_/   \_\_| \_|\____|\___/_/   \_\____|_____|____/
//

// BaseLanguage describes a language in use in the campaign world.
type BaseLanguage struct {
	Language string
	IsLocal  bool `json:",omitempty"`
}

// ImportLanguage reads a BaseLanguage from the input JSON data stream, writing it to the database.
func ImportLanguage(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var lang BaseLanguage
	var err error
	var id int64
	var exists bool

	if err = decoder.Decode(&lang); err != nil {
		return err
	}

	if filterOut(prefs, TypeLanguage, "language", lang.Language, lang.Language, lang.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Languages", "ID", "Language", lang.Language); err != nil {
		return err
	}

	if exists {
		_, err = db.Exec(`UPDATE Languages SET Language=?, IsLocal=? WHERE ID=?`, lang.Language, lang.IsLocal, id)
	} else {
		_, err = db.Exec(`INSERT INTO Languages (Language, IsLocal) VALUES (?,?)`, lang.Language, lang.IsLocal)
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing language %s (database id %d)", lang.Language, id)
		} else {
			log.Printf("created new language %s", lang.Language)
		}
	}
	return err
}

// ExportLanguages reads all languages from the database, writing them to the open JSON file.
func ExportLanguages(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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

		if err := rows.Scan(&lang.Language, &lang.IsLocal); err != nil {
			return err
		}

		if filterOut(prefs, TypeLanguage, "language", lang.Language, lang.Language, lang.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
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

//  __  __  ___  _   _ ____ _____ _____ ____  ____
// |  \/  |/ _ \| \ | / ___|_   _| ____|  _ \/ ___|
// | |\/| | | | |  \| \___ \ | | |  _| | |_) \___ \
// | |  | | |_| | |\  |___) || | | |___|  _ < ___) |
// |_|  |_|\___/|_| \_|____/ |_| |_____|_| \_\____/
//

// Monster describes a creature (could be an individual or descriptive of a species)
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

// SpellBlock describes a set of spells prepared by a creature.
// There may be multiple blocks if it has multiple types of spells.
// There will also be one of these for spell-like abilities
// in which case ClassName will be "SLA".
type SpellBlock struct {
	ClassName            string
	CL                   int  `json:",omitempty"`
	Concentration        int  `json:",omitempty"`
	NoConcentrationValue bool `json:",omitempty"`
	// Add this many domain spells at each level.
	PlusDomain  int    `json:",omitempty"`
	Description string `json:",omitempty"`
	Special     string `json:",omitempty"`
	// List of the spells prepared/known for this block.
	Spells []PreparedSpell
}

// PreparedSpell describes a spell that a creature has prepared to cast,
// or a spell-like ability it can use.
type PreparedSpell struct {
	Name string
	// Some spell-like abilities are based on a core spell (named in Name),
	// with some different effects (listed in Special), and with a new name
	// (named in AlternateName). For example, "coldball" is like "fireball"
	// but deals cold damage.
	AlternateName string `json:",omitempty"`
	// For spell-like abilities, how may times per day the spell can be cast.
	Frequency string `json:",omitempty"`
	Special   string `json:",omitempty"`
	// Each instance of this prepared spell is described in Slots.
	Slots []SpellSlot
}

// SpellSlot describes each prepared instance of a particular spell.
type SpellSlot struct {
	IsCast   bool `json:",omitempty"`
	IsDomain bool `json:",omitempty"`
	// A list of metamagic feats used when preparing this instance.
	MetaMagic []string `json:",omitempty"`
}

// MonsterSkill describes the significant skills for the creature.
type MonsterSkill struct {
	Code     string `json:",omitempty"`
	Modifier int    `json:",omitempty"`
	Notes    string `json:",omitempty"`
}

// MonsterFeat describes each feat known by the creature.
type MonsterFeat struct {
	// Feat code identifies the feat (must exist in the core feat list).
	Code string `json:",omitempty"`
	// Parameters hold any specific parameters (e.g. "sword" in "weapon focus (sword)").
	Parameters string `json:",omitempty"`
	// Is this a bonus feat for this creature?
	IsBonus bool `json:",omitempty"`
}

// Language describes each language known by a creature.
type Language struct {
	Name string `json:",omitempty"`
	// Can the creature understand the language but not speak?
	IsMute bool `json:",omitempty"`
	// Any special considerations for this language for this creature.
	Special string `json:",omitempty"`
}

// AttackMode describes each attack a monster can make.
type AttackMode struct {
	// Each tier is an individual way a monster can attack in a round.
	// If multiple AttackModes have the same Tier, they are all taken
	// together if making a full-round attack.
	Tier int
	// If this attack is based on a weapon, its Code appears here.
	BaseWeaponID string `json:",omitempty"`
	// The number of attacks the creature gets for this. Normally this is 1,
	// but could be, e.g., 2 for an attack listed as "2 Claws".
	Multiple int
	Name     string
	// The name of the attack. If only part of this should be visible to the players
	// (e.g., in the online die-roller), then the player-visible parts should be in
	// square brackets (e.g., "+1 vorpal [shortsword]")
	//
	// If multiple attacks with different attack rolls are applicable, then
	// Multiple will be 1 and the different attacks are listed here with slashes
	// between them (e.g., "+17/+12/+7/+2")
	Attack string `json:",omitempty"`
	// The damage dealt by each blow from this attack. If there are parts of the
	// damage which should be seen by the GM only and not given to the players NOR the
	// die-roller, put the parts which should be given to the players and die roller in
	// square brackets (e.g., "[2d6 acid] plus poison").
	//
	// If there is damage which is not multiplied on critical hits, enclose those
	// portions in angle brackets, including any math operators they use
	// (e.g., "2d6 <+1d6 electricity>"). On a critical hit, any text inside <...> will be
	// omitted entirely from the extra damage.
	Damage   string `json:",omitempty"`
	Critical struct {
		CantCritical bool `json:",omitempty"`
		Threat       int  `json:",omitempty"`
		Multiplier   int  `json:",omitempty"`
	} `json:",omitempty"`
	Ranged struct {
		IsRanged      bool `json:",omitempty"`
		Increment     int  `json:",omitempty"`
		MaxIncrements int  `json:",omitempty"`
	} `json:",omitempty"`
	IsReach bool   `json:",omitempty"`
	Special string `json:",omitempty"`
	// Attack mode (melee or ranged)
	Mode string `json:",omitempty"`
}

// SavingThrow describes each of the creature's saving throw.
type SavingThrow struct {
	Mod           int
	Special       string `json:",omitempty"`
	NoSavingThrow bool   `json:",omitempty"`
}

// AbilityScore describes each of the creature's ability scores.
type AbilityScore struct {
	// Raw score.
	Base    int    `json:",omitempty"`
	Special string `json:",omitempty"`
	// If NullScore is true, the creature simply does not have this ability score at all,
	// which is different than an ability score of zero.
	NullScore bool `json:",omitempty"`
}

// ImportMonster reads a Monster from the JSON input stream, writing it to the database.
func ImportMonster(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var mob Monster
	var err error
	var id int64
	var exists bool
	var res sql.Result

	if err = decoder.Decode(&mob); err != nil {
		return err
	}

	if filterOut(prefs, TypeBestiary, "monster", mob.Code, mob.Species, mob.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Monsters", "ID", "Code", mob.Code); err != nil {
		return err
	}

	var alignments int
	alignmentIDs, err := getAlignmentCodeIDs(db, prefs)
	if err != nil {
		return err
	}
	for _, a := range mob.Alignment.Alignments {
		bit, ok := alignmentIDs[a]
		if !ok {
			return fmt.Errorf("undefined alignment \"%s\" for species \"%s\"", a, mob.Species)
		}
		alignments |= 1 << bit
	}

	// handle the fact that many of these may be NULL in the database
	var alignspec, ispec, senses, aura, hpspec, ft, rt, wt, sm, da sql.NullString
	var drbp, imm, resist, srtext, sa, strtext, dextext, context, inttext, wistext sql.NullString
	var chatext, cmbtext, cmdtext, sq, env, org, treasure, appearance, grp, before sql.NullString
	var during, morale, agecat, gender, bline, patron, altname, varpar, offense, stats sql.NullString
	var gear, other, focused, arch, basestats, rmods, prohibited, opposition, mystery sql.NullString
	var notes sql.NullString

	if mob.Combat.CMBSpecial != "" {
		cmbtext.Valid = true
		cmbtext.String = mob.Combat.CMBSpecial
	}
	if mob.Combat.CMDSpecial != "" {
		cmdtext.Valid = true
		cmdtext.String = mob.Combat.CMDSpecial
	}
	if mob.SQ != "" {
		sq.Valid = true
		sq.String = mob.SQ
	}
	if mob.Environment != "" {
		env.Valid = true
		env.String = mob.Environment
	}
	if mob.Organization != "" {
		org.Valid = true
		org.String = mob.Organization
	}
	if mob.Treasure != "" {
		treasure.Valid = true
		treasure.String = mob.Treasure
	}
	if mob.Appearance != "" {
		appearance.Valid = true
		appearance.String = mob.Appearance
	}
	if mob.Group != "" {
		grp.Valid = true
		grp.String = mob.Group
	}
	if mob.Strategy.BeforeCombat != "" {
		before.Valid = true
		before.String = mob.Strategy.BeforeCombat
	}
	if mob.Strategy.DuringCombat != "" {
		during.Valid = true
		during.String = mob.Strategy.DuringCombat
	}
	if mob.Strategy.Morale != "" {
		morale.Valid = true
		morale.String = mob.Strategy.Morale
	}
	if mob.AgeCategory != "" {
		agecat.Valid = true
		agecat.String = mob.AgeCategory
	}
	if mob.Gender != "" {
		gender.Valid = true
		gender.String = mob.Gender
	}
	if mob.Bloodline != "" {
		bline.Valid = true
		bline.String = mob.Bloodline
	}
	if mob.Patron != "" {
		patron.Valid = true
		patron.String = mob.Patron
	}
	if mob.AlternateNameForm != "" {
		altname.Valid = true
		altname.String = mob.AlternateNameForm
	}
	if mob.VariantParent != "" {
		varpar.Valid = true
		varpar.String = mob.VariantParent
	}
	if mob.OffenseNote != "" {
		offense.Valid = true
		offense.String = mob.OffenseNote
	}
	if mob.StatisticsNote != "" {
		stats.Valid = true
		stats.String = mob.StatisticsNote
	}
	if mob.Gear.Combat != "" {
		gear.Valid = true
		gear.String = mob.Gear.Combat
	}
	if mob.Gear.Other != "" {
		other.Valid = true
		other.String = mob.Gear.Other
	}
	if mob.Schools.Focused != "" {
		focused.Valid = true
		focused.String = mob.Schools.Focused
	}
	if mob.Schools.Prohibited != "" {
		prohibited.Valid = true
		prohibited.String = mob.Schools.Prohibited
	}
	if mob.Schools.Opposition != "" {
		opposition.Valid = true
		opposition.String = mob.Schools.Opposition
	}
	if mob.ClassArchetypes != "" {
		arch.Valid = true
		arch.String = mob.ClassArchetypes
	}
	if mob.BaseStatistics != "" {
		basestats.Valid = true
		basestats.String = mob.BaseStatistics
	}
	if mob.RacialMods != "" {
		rmods.Valid = true
		rmods.String = mob.RacialMods
	}
	if mob.Mystery != "" {
		mystery.Valid = true
		mystery.String = mob.Mystery
	}
	if mob.Notes != "" {
		notes.Valid = true
		notes.String = mob.Notes
	}

	if mob.DR.Bypass != "" {
		drbp.Valid = true
		drbp.String = mob.DR.Bypass
	}
	if mob.Immunities != "" {
		imm.Valid = true
		imm.String = mob.Immunities
	}
	if mob.Resists != "" {
		resist.Valid = true
		resist.String = mob.Resists
	}
	if mob.SR.Special != "" {
		srtext.Valid = true
		srtext.String = mob.SR.Special
	}
	if mob.SpecialAttacks != "" {
		sa.Valid = true
		sa.String = mob.SpecialAttacks
	}
	if mob.Abilities.Str.Special != "" {
		strtext.Valid = true
		strtext.String = mob.Abilities.Str.Special
	}
	if mob.Abilities.Dex.Special != "" {
		dextext.Valid = true
		dextext.String = mob.Abilities.Dex.Special
	}
	if mob.Abilities.Con.Special != "" {
		context.Valid = true
		context.String = mob.Abilities.Con.Special
	}
	if mob.Abilities.Int.Special != "" {
		inttext.Valid = true
		inttext.String = mob.Abilities.Int.Special
	}
	if mob.Abilities.Wis.Special != "" {
		wistext.Valid = true
		wistext.String = mob.Abilities.Wis.Special
	}
	if mob.Abilities.Cha.Special != "" {
		chatext.Valid = true
		chatext.String = mob.Abilities.Cha.Special
	}

	if mob.Alignment.Special != "" {
		alignspec.Valid = true
		alignspec.String = mob.Alignment.Special
	}
	if mob.Initiative.Special != "" {
		ispec.Valid = true
		ispec.String = mob.Initiative.Special
	}
	if mob.Senses != "" {
		senses.Valid = true
		senses.String = mob.Senses
	}
	if mob.Aura != "" {
		aura.Valid = true
		aura.String = mob.Aura
	}
	if mob.HP.Special != "" {
		hpspec.Valid = true
		hpspec.String = mob.HP.Special
	}
	if mob.Save.Fort.Special != "" {
		ft.Valid = true
		ft.String = mob.Save.Fort.Special
	}
	if mob.Save.Refl.Special != "" {
		rt.Valid = true
		rt.String = mob.Save.Refl.Special
	}
	if mob.Save.Will.Special != "" {
		wt.Valid = true
		wt.String = mob.Save.Will.Special
	}
	if mob.Save.Special != "" {
		sm.Valid = true
		sm.String = mob.Save.Special
	}
	if mob.DefensiveAbilities != "" {
		da.Valid = true
		da.String = mob.DefensiveAbilities
	}

	var curhp, fort, refl, will, dr, sr, str, dex, con, int_, wis, cha, bab, cmb, cmd sql.NullInt32
	var mr, mt, acadj, fadj, tadj sql.NullInt32

	if v, ok := mob.AC.Adjustments["AC"]; ok {
		acadj.Valid = true
		acadj.Int32 = int32(v)
	}
	if v, ok := mob.AC.Adjustments["Touch"]; ok {
		tadj.Valid = true
		tadj.Int32 = int32(v)
	}
	if v, ok := mob.AC.Adjustments["Flat"]; ok {
		fadj.Valid = true
		fadj.Int32 = int32(v)
	}

	// TODO is this really right? we could have 0 current HP in some cases which is different than
	// not providing this stat for the monster.
	if mob.HP.Current != 0 {
		curhp.Valid = true
		curhp.Int32 = int32(mob.HP.Current)
	}

	if !mob.Save.Fort.NoSavingThrow {
		fort.Valid = true
		fort.Int32 = int32(mob.Save.Fort.Mod)
	}
	if !mob.Save.Refl.NoSavingThrow {
		refl.Valid = true
		refl.Int32 = int32(mob.Save.Refl.Mod)
	}
	if !mob.Save.Will.NoSavingThrow {
		will.Valid = true
		will.Int32 = int32(mob.Save.Will.Mod)
	}

	if mob.DR.DR != 0 {
		dr.Valid = true
		dr.Int32 = int32(mob.DR.DR)
	}
	if mob.SR.SR != 0 {
		sr.Valid = true
		sr.Int32 = int32(mob.SR.SR)
	}

	if !mob.Abilities.Str.NullScore {
		str.Valid = true
		str.Int32 = int32(mob.Abilities.Str.Base)
	}
	if !mob.Abilities.Dex.NullScore {
		dex.Valid = true
		dex.Int32 = int32(mob.Abilities.Dex.Base)
	}
	if !mob.Abilities.Con.NullScore {
		con.Valid = true
		con.Int32 = int32(mob.Abilities.Con.Base)
	}
	if !mob.Abilities.Int.NullScore {
		int_.Valid = true
		int_.Int32 = int32(mob.Abilities.Int.Base)
	}
	if !mob.Abilities.Wis.NullScore {
		wis.Valid = true
		wis.Int32 = int32(mob.Abilities.Wis.Base)
	}
	if !mob.Abilities.Cha.NullScore {
		cha.Valid = true
		cha.Int32 = int32(mob.Abilities.Cha.Base)
	}

	// TODO we're not going to assume BAB of +0 is NULL. I'm not sure it makes sense for
	// the database to allow BAB to be null in the first place.
	// same for CMB and CMD.
	bab.Valid = true
	bab.Int32 = int32(mob.Combat.BAB)
	cmb.Valid = true
	cmb.Int32 = int32(mob.Combat.CMB)
	cmd.Valid = true
	cmd.Int32 = int32(mob.Combat.CMD)

	if mob.Mythic.IsMythic {
		mr.Valid = true
		mr.Int32 = int32(mob.Mythic.MR)
		mt.Valid = true
		mt.Int32 = int32(mob.Mythic.MT)
	}
	// fk mob.Type mob.Size.Code

	if exists {
		for _, q := range []string{
			`DELETE FROM SpellList 
			WHERE CollectionID IN (
				SELECT ID FROM SpellsPrepared
				WHERE MonsterID=?
			)`, `
			DELETE FROM SpellSlots 
			WHERE CollectionID IN (
				SELECT ID FROM SpellsPrepared
				WHERE MonsterID=?
			)`, `DELETE FROM MonsterDomains WHERE MonsterID=?`,
			`DELETE FROM ACComponents WHERE MonsterID=?`,
			`DELETE FROM AttackModes WHERE MonsterID=?`,
			`DELETE FROM MonsterSubtypes WHERE MonsterID=?`,
			`DELETE FROM MonsterLanguages WHERE MonsterID=?`,
			`DELETE FROM MonsterFeats WHERE MonsterID=?`,
			`DELETE FROM MonsterFeats WHERE MonsterID=?`,
			`DELETE FROM SpellsPrepared WHERE MonsterID=?`,
		} {
			_, err = db.Exec(q, id)
			if err != nil {
				return err
			}
		}

		_, err = db.Exec(`
			UPDATE Monsters 
			SET 
				IsLocal=?, Species=?, Code=?, CR=?, XP=?, Class=?, Alignment=?,
				AlignmentSpecial=?, Source=?, Size=?, SpaceText=?, ReachText=?,
				Type=?, Initiative=?, InitiativeText=?, Senses=?, Aura=?,
				TypicalHP=? CurrentHP=?, HPSpecial=?, HitDice=?, Fort=?, Refl=?,
				Will=?, FortText=?, ReflText=?, WillText=?, SaveMods=?,
				DefensiveAbilities=?, DR=?, DRBypass=?, Immunities=?, Resists=?,
				SR=?, SRText=?, Weaknesses=?, Speed=?, SpeedText=?, SpecialAttacks=?,
				Str=?, Dex=?, Con=?, Int=?, Wis=?, Cha=?,
				StrText=?, DexText=?, ConText=?, IntText=?, WisText=?, ChaText=?,
				BAB=?, CMB=?, CMD=?, CMBText=?, CMDText=?, SQ=?, Environment=?,
				Organization=?, Treasure=?, Appearance=?, Grp=?, IsTemplate=?,
				BeforeCombat=?, DuringCombat=?, Morale=?, CharacterFlag=?,
				CompanionFlag=?, IsUniqueMonster=?, AgeCategory=?, Gender=?,
				Bloodline=?, Patron=?, AlternateNameForm=?, DoneUseRacialHD=?,
				VariantParent=?, MR=?, IsMythic=?, MT=?, OffenseNote=?,
				StatisticsNote=?, Gear=?, OtherGear=?, FocusedSchool=?,
				ClassArchetypes=?, BaseStatistics=?, ACAdj=?, FlatAdj=?, TouchAdj=?,
				RacialMods=?, ProhibitedSchools=?, OppositionSchools=?, Mystery=?,
				Notes=?
			WHERE
				ID=?`,
			mob.IsLocal, mob.Species, mob.Code, mob.CR, mob.XP, mob.Class, alignments,
			alignspec, mob.Source, mob.Size.Code, mob.Size.SpaceText, mob.Size.ReachText,
			mob.Type, mob.Initiative.Mod, ispec, senses, aura, mob.HP.Typical, curhp,
			hpspec, mob.HP.HitDice, fort, refl, will, ft, rt, wt, sm, da, dr, drbp,
			imm, resist, sr, srtext, sa, str, dex, con, int_, wis, cha, strtext,
			dextext, context, inttext, wistext, chatext, bab, cmb, cmd, cmbtext,
			cmdtext, sq, env, org, treasure, appearance, grp, mob.IsTemplate,
			before, during, morale, mob.IsCharacter, mob.IsCompanion, mob.IsUnique,
			agecat, gender, bline, patron, altname, mob.DontUseRacialHD, varpar,
			mr, mob.Mythic.IsMythic, mt, offense, stats, gear, other, focused, arch,
			basestats, acadj, fadj, tadj, rmods, prohibited, opposition, mystery,
			notes,
			id)
	} else {
		if res, err = db.Exec(`
			INSERT INTO Monsters (
				IsLocal, Species, Code, CR, XP, Class, Alignment,
				AlignmentSpecial, Source, Size, SpaceText, ReachText,
				Type, Initiative, InitiativeText, Senses, Aura,
				TypicalHP CurrentHP, HPSpecial, HitDice, Fort, Refl,
				Will, FortText, ReflText, WillText, SaveMods,
				DefensiveAbilities, DR, DRBypass, Immunities, Resists,
				SR, SRText, Weaknesses, Speed, SpeedText, SpecialAttacks,
				Str, Dex, Con, Int, Wis, Cha,
				StrText, DexText, ConText, IntText, WisText, ChaText,
				BAB, CMB, CMD, CMBText, CMDText, SQ, Environment,
				Organization, Treasure, Appearance, Grp, IsTemplate,
				BeforeCombat, DuringCombat, Morale, CharacterFlag,
				CompanionFlag, IsUniqueMonster, AgeCategory, Gender,
				Bloodline, Patron, AlternateNameForm, DoneUseRacialHD,
				VariantParent, MR, IsMythic, MT, OffenseNote,
				StatisticsNote, Gear, OtherGear, FocusedSchool,
				ClassArchetypes, BaseStatistics, ACAdj, FlatAdj, TouchAdj,
				RacialMods, ProhibitedSchools, OppositionSchools, Mystery,
				Notes
			)`,
			mob.IsLocal, mob.Species, mob.Code, mob.CR, mob.XP, mob.Class, alignments,
			alignspec, mob.Source, mob.Size.Code, mob.Size.SpaceText, mob.Size.ReachText,
			mob.Type, mob.Initiative.Mod, ispec, senses, aura, mob.HP.Typical, curhp,
			hpspec, mob.HP.HitDice, fort, refl, will, ft, rt, wt, sm, da, dr, drbp,
			imm, resist, sr, srtext, sa, str, dex, con, int_, wis, cha, strtext,
			dextext, context, inttext, wistext, chatext, bab, cmb, cmd, cmbtext,
			cmdtext, sq, env, org, treasure, appearance, grp, mob.IsTemplate,
			before, during, morale, mob.IsCharacter, mob.IsCompanion, mob.IsUnique,
			agecat, gender, bline, patron, altname, mob.DontUseRacialHD, varpar,
			mr, mob.Mythic.IsMythic, mt, offense, stats, gear, other, focused, arch,
			basestats, acadj, fadj, tadj, rmods, prohibited, opposition, mystery,
			notes); err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing monster %s (database id %d)", mob.Species, id)
		} else {
			log.Printf("created new monster %s (database id %d)", mob.Species, id)
		}
	}

	// add domains
	for _, domainName := range mob.Domains {
		if err = makeRecordExist(db, prefs, "Domains", "Domain", domainName); err != nil {
			return err
		}
		if _, err = db.Exec(`
			INSERT INTO MonsterDomains (MonsterID, DomainID)
			VALUES (?, (SELECT ID From Domains WHERE Domain=?))
			`, id, domainName); err != nil {
			return err
		}
	}

	// armor class
	for actype, acval := range mob.AC.Components {
		if err = makeRecordExist(db, prefs, "ACCTypes", "Code", actype); err != nil {
			return err
		}
		if _, err = db.Exec(`
			INSERT INTO ACComponents (MonsterID, ACComponent, Value)
			VALUES (?, (SELECT ID FROM ACCTypes WHERE Code=?), ?)`,
			id, actype, acval); err != nil {
			return err
		}
	}

	// attack modes
	var seq int
	var lastTier int
	for _, amode := range mob.AttackModes {
		var baseweap sql.NullInt32
		var weapid int64
		var ok bool

		if lastTier != amode.Tier {
			lastTier = amode.Tier
			seq = 0
		}
		if amode.BaseWeaponID != "" {
			if ok, weapid, err = recordExists(db, prefs, "Weapons", "ID", "Code", amode.BaseWeaponID); !ok || err != nil {
				return fmt.Errorf("monster %s attack %s (tier %d, seq %d) references unknown weapon code %s (database error %v)", mob.Species, amode.Name, amode.Tier, seq, amode.BaseWeaponID, err)
			}
			baseweap.Valid = true
			baseweap.Int32 = int32(weapid)
		}
		if ok, err = recordExistsWithoutID(db, prefs, "AttackModeTypes", "Mode", amode.Mode); !ok || err != nil {
			return fmt.Errorf("monster %s attack %s (tier %d, seq %d) references unknown mode %s (database error %v)", mob.Species, amode.Name, amode.Tier, seq, amode.Mode, err)
		}

		var attack, damage, spec sql.NullString
		var ri, rmax, threat, critical sql.NullInt32

		if amode.Attack != "" {
			attack.Valid = true
			attack.String = amode.Attack
		}
		if amode.Damage != "" {
			damage.Valid = true
			damage.String = amode.Damage
		}
		if !amode.Critical.CantCritical {
			threat.Valid = true
			critical.Valid = true
			threat.Int32 = int32(amode.Critical.Threat)
			critical.Int32 = int32(amode.Critical.Multiplier)
		}
		if amode.Ranged.IsRanged {
			ri.Valid = true
			rmax.Valid = true
			ri.Int32 = int32(amode.Ranged.Increment)
			rmax.Int32 = int32(amode.Ranged.MaxIncrements)
		}
		if amode.Special != "" {
			spec.Valid = true
			spec.String = amode.Special
		}

		if _, err = db.Exec(`
			INSERT INTO AttackModes (
				MonsterID, TierGroup, TierSeq, BaseWeaponID,
				Multiple, Name, Attack, Damage, Threat, Critical,
				RangeInc, RangeMax, IsReach, Special, Mode
			)
			VALUES (
		`, id, amode.Tier, seq, baseweap, amode.Multiple, amode.Name,
			attack, damage, threat, critical, ri, rmax, amode.IsReach, spec, amode.Mode,
		); err != nil {
			return err
		}
		seq++
	}

	// monster subtypes
	for _, stype := range mob.Subtypes {
		if err = makeRecordExist(db, prefs, "MonsterSubtypes", "Subtype", stype); err != nil {
			return err
		}
		if _, err = db.Exec(`
			INSERT INTO MonsterMonsterSubtypes (MonsterID, SubtypeID)
			VALUES (?, (SELECT ID FROM MonsterSubtypes WHERE Subtype=?))
		`, id, stype); err != nil {
			return err
		}
	}

	// languages
	for _, lang := range mob.Languages {
		if err = makeRecordExist(db, prefs, "Languages", "Language", lang.Name); err != nil {
			return err
		}

		var spec sql.NullString
		if lang.Special != "" {
			spec.Valid = true
			spec.String = lang.Special
		}

		if _, err = db.Exec(`
			INSERT INTO MonsterLanguages (MonsterID, LanguageID, IsMute, Special)
			VALUES (?, (SELECT ID FROM Languages WHERE Language=?), ?, ?)
		`, id, lang.Name, lang.IsMute, spec); err != nil {
			return err
		}
	}

	// feats
	for _, feat := range mob.Feats {
		var ok bool
		if ok, err = recordExistsWithoutID(db, prefs, "Feats", "Code", feat.Code); !ok || err != nil {
			return fmt.Errorf("monster %s references unknown feat code %s (database error %v)", mob.Species, feat.Code, err)
		}

		var params sql.NullString
		if feat.Parameters != "" {
			params.Valid = true
			params.String = feat.Parameters
		}

		if _, err = db.Exec(`
			INSERT INTO MonsterFeats (MonsterID, FeatID, Parameters, BonusFeat)
			VALUES (?, (SELECT ID FROM Feats WHERE Code=?), ?, ?)
		`, id, feat.Code, params, feat.IsBonus); err != nil {
			return err
		}
	}

	// skills
	for _, skill := range mob.Skills {
		var ok bool
		if ok, err = recordExistsWithoutID(db, prefs, "Skills", "Code", skill.Code); !ok || err != nil {
			return fmt.Errorf("monster %s references unknown skill code %s (database error %v)", mob.Species, skill.Code, err)
		}

		var notes sql.NullString
		if skill.Notes != "" {
			notes.Valid = true
			notes.String = skill.Notes
		}

		if _, err = db.Exec(`
			INSERT INTO MonsterSkills (MonsterID, SkillID, Modifier, Notes)
			VALUES (?, (SELECT ID FROM Skills WHERE Code=?), ?, ?)
		`, id, skill.Code, skill.Modifier, notes); err != nil {
			return err
		}
	}

	// spells
	for _, spell := range mob.Spells {
		var ok bool
		var desc, spec sql.NullString
		var class, conc sql.NullInt32
		var clsid int64
		var spid int64

		if spell.ClassName != "SLA" {
			if ok, clsid, err = recordExists(db, prefs, "Classes", "ID", "Name", spell.ClassName); !ok || err != nil {
				return fmt.Errorf("monster %s references unknown spellcaster class %s (database error %v)", mob.Species, spell.ClassName, err)
			}
			class.Valid = true
			class.Int32 = int32(clsid)
		}
		if spell.Description != "" {
			desc.Valid = true
			desc.String = spell.Description
		}
		if spell.Special != "" {
			spec.Valid = true
			spec.String = spell.Special
		}
		if !spell.NoConcentrationValue {
			conc.Valid = true
			conc.Int32 = int32(spell.Concentration)
		}

		if res, err = db.Exec(`
			INSERT INTO SpellsPrepared (
				MonsterID, ClassID, Description, CL, Concentration,
				PlusDomain, Special
			)
			VALUES (?, (SELECT ID FROM Classes WHERE Name=?), ?)
		`, id, class, desc, spell.CL, conc, spell.PlusDomain, spec); err != nil {
			return err
		}

		if spid, err = res.LastInsertId(); err != nil {
			return err
		}

		for _, eachSpell := range spell.Spells {
			if ok, err = recordExistsWithoutID(db, prefs, "Spells", "Name", eachSpell.Name); !ok || err != nil {
				return fmt.Errorf("monster %s references unknown spell name %s (database error %v)", mob.Species, eachSpell.Name, err)
			}

			var alt, freq, spec sql.NullString
			if eachSpell.AlternateName != "" {
				alt.Valid = true
				alt.String = eachSpell.AlternateName
			}
			if eachSpell.Frequency != "" {
				freq.Valid = true
				freq.String = eachSpell.Frequency
			}
			if eachSpell.Special != "" {
				spec.Valid = true
				spec.String = eachSpell.Special
			}

			if res, err = db.Exec(`
				INSERT INTO SpellList (
					CollectionID, SpellID, AlternateName, Frequency, Special
				)
				VALUES (?, (SELECT ID FROM Spells WHERE Name=?), ?, ?, ?)
			`, spid, eachSpell.Name, alt, freq, spec); err != nil {
				return err
			}

			var spellid, slotid int64
			if spellid, err = res.LastInsertId(); err != nil {
				return err
			}

			for inst, slot := range eachSpell.Slots {
				if res, err = db.Exec(`
					INSERT INTO SpellSlots (
						CollectionID, SpellID, Instance, IsCast, IsDomain
					)
					VALUES (?, ?, ?, ?, ?)
				`, spid, spellid, inst, slot.IsCast, slot.IsDomain); err != nil {
					return nil
				}

				if slotid, err = res.LastInsertId(); err != nil {
					return err
				}

				for _, meta := range slot.MetaMagic {
					if ok, err = recordExistsWithoutID(db, prefs, "Feats", "Code", meta); !ok || err != nil {
						return fmt.Errorf("monster %s references unknown feat code %s (database error %v)", mob.Species, meta, err)
					}

					if _, err = db.Exec(`
						INSERT INTO SpellSlotMeta (SlotID, MetaID)
						VALUES (?, (SELECT ID FROM Feats WHERE Code=?))
					`, slotid, meta); err != nil {
						return err
					}
				}
			}
		}
	}

	return err
}

// ExportBestiary exports all Monster values from the database, writing them to the open JSON file.
func ExportBestiary(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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

		if filterOut(prefs, TypeBestiary, "monster", monster.Species, monster.Code, monster.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
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
		} else {
			monster.Save.Fort.NoSavingThrow = true
		}

		if r.Valid {
			monster.Save.Refl.Mod = int(r.Int32)
		} else {
			monster.Save.Refl.NoSavingThrow = true
		}
		if w.Valid {
			monster.Save.Will.Mod = int(w.Int32)
		} else {
			monster.Save.Will.NoSavingThrow = true
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
					TierGroup, Weapons.Code, Multiple, 
					AttackModes.Name, AttackModes.Attack, AttackModes.Damage,
					AttackModes.Threat, AttackModes.Critical, AttackModes.RangeInc,
					AttackModes.RangeMax, AttackModes.IsReach, AttackModes.Special, AttackModes.Mode
				FROM AttackModes
				LEFT JOIN Weapons
					ON Weapons.ID=BaseWeaponID
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
				var weap, att, dam, spec sql.NullString
				var ri, rmax, threat, mult sql.NullInt32

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
						am.Critical.Threat = int(threat.Int32)
					} else {
						am.Critical.Threat = 20
					}
					if mult.Valid {
						am.Critical.Multiplier = int(mult.Int32)
					} else {
						am.Critical.Multiplier = 2
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

//  ____  ____  _____ _     _     ____
// / ___||  _ \| ____| |   | |   / ___|
// \___ \| |_) |  _| | |   | |   \___ \
//  ___) |  __/| |___| |___| |___ ___) |
// |____/|_|   |_____|_____|_____|____/
//

// Spell describes a spell in the campaign world.
type Spell struct {
	IsLocal     bool `json:",omitempty"`
	Name        string
	Code        string
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
		// Range code
		Range string
		// Specific distance if called for by the range code.
		Distance int `json:",omitempty"`
		// Distance to add per level.
		DistancePerLevel int    `json:",omitempty"`
		DistanceSpecial  string `json:",omitempty"`
	}
	Effect struct {
		Area    string `json:",omitempty"`
		Effect  string `json:",omitempty"`
		Targets string `json:",omitempty"`
	}
	Duration struct {
		// Duration code.
		Duration string
		// Specific amount of time if called for by the duration code.
		Time    string `json:",omitempty"`
		Special string `json:",omitempty"`
		// Does the spell last while concentrated upon?
		Concentration bool `json:",omitempty"`
		// Is the duration specified here multiplied by their level?
		PerLevel bool `json:",omitempty"`
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
	// ClassLevels lists what level of spell this is for each spellcasting class that can cast it.
	// Note that Class is "SLA" for spell-like abilities.
	ClassLevels []struct {
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

// ImportSpell reads the next Spell from the input JSON stream, writing it to the database.
func ImportSpell(decoder *json.Decoder, db *sql.DB, prefs *CorePreferences) error {
	var spell Spell
	var err, err2 error
	var id int64
	var exists bool
	var res sql.Result

	if err = decoder.Decode(&spell); err != nil {
		return err
	}

	if filterOut(prefs, TypeSpell, "spell", spell.Name, spell.Code, spell.IsLocal) {
		return nil
	}

	if exists, id, err = recordExists(db, prefs, "Spells", "ID", "Code", spell.Code); err != nil {
		return err
	}

	var schoolID, descriptors, components int
	var castspec, material, focus, range_, distspec, area, effect sql.NullString
	var targets, durtime, durspec, save, saveeffect, savespec, srspec sql.NullString
	var deity, domain, description, src, bloodline, patron sql.NullString
	var distance, distperlevel, slalevel, materialcosts sql.NullInt32

	if err = makeRecordExistWithoutID(db, prefs, "Schools", "Code", spell.School); err != nil {
		return err
	}
	schoolIDs, err := getSpellSchoolIDs(db, prefs)
	if err != nil {
		return err
	}
	schoolID = schoolIDs[spell.School]
	for _, desc := range spell.Descriptors {
		if err = makeRecordExist(db, prefs, "SpellDescriptors", "Descriptor", desc); err != nil {
			return err
		}
	}
	descriptorIDs, err := getSpellDescriptorIDs(db, prefs)
	if err != nil {
		return err
	}

	for _, desc := range spell.Descriptors {
		if did, ok := descriptorIDs[desc]; ok {
			descriptors |= did
		}
	}

	componentIDs, err := getSpellComponentIDs(db, prefs)
	if err != nil {
		return err
	}

	for _, comp := range spell.Descriptors {
		if cid, ok := componentIDs[comp]; ok {
			components |= cid
		}
	}

	if spell.Casting.Special == "" {
		castspec.Valid = false
	} else {
		castspec.Valid = true
		castspec.String = spell.Casting.Special
	}
	if spell.Components.Material == "" {
		material.Valid = false
	} else {
		material.Valid = true
		material.String = spell.Components.Material
	}
	if spell.Components.Focus == "" {
		focus.Valid = false
	} else {
		focus.Valid = true
		focus.String = spell.Components.Focus
	}
	if spell.Range.Range == "" {
		range_.Valid = false
	} else {
		range_.Valid = true
		range_.String = spell.Range.Range
	}
	if spell.Range.DistanceSpecial == "" {
		distspec.Valid = false
	} else {
		distspec.Valid = true
		range_.String = spell.Range.DistanceSpecial
	}
	if spell.Effect.Area == "" {
		area.Valid = false
	} else {
		area.Valid = true
		area.String = spell.Effect.Area
	}
	if spell.Effect.Effect == "" {
		effect.Valid = false
	} else {
		effect.Valid = true
		effect.String = spell.Effect.Effect
	}
	if spell.Effect.Targets == "" {
		targets.Valid = false
	} else {
		targets.Valid = true
		targets.String = spell.Effect.Targets
	}

	if spell.Duration.Time == "" {
		durtime.Valid = false
	} else {
		durtime.Valid = true
		durtime.String = spell.Duration.Time
	}
	if spell.Duration.Special == "" {
		durspec.Valid = false
	} else {
		durspec.Valid = true
		durspec.String = spell.Duration.Special
	}
	if spell.Save.SavingThrow == "" {
		save.Valid = false
	} else {
		save.Valid = true
		save.String = spell.Save.SavingThrow
	}
	if spell.Save.Effect == "" {
		saveeffect.Valid = false
	} else {
		saveeffect.Valid = true
		saveeffect.String = spell.Save.Effect
	}
	if spell.Save.Special == "" {
		savespec.Valid = false
	} else {
		savespec.Valid = true
		savespec.String = spell.Save.Special
	}
	if spell.SR.Special == "" {
		srspec.Valid = false
	} else {
		srspec.Valid = true
		srspec.String = spell.SR.Special
	}
	if spell.Deity == "" {
		deity.Valid = false
	} else {
		deity.Valid = true
		deity.String = spell.Deity
	}
	if spell.Domain == "" {
		domain.Valid = false
	} else {
		domain.Valid = true
		domain.String = spell.Domain
	}
	if spell.Description == "" {
		description.Valid = false
	} else {
		description.Valid = true
		description.String = spell.Description
	}
	if spell.Source == "" {
		src.Valid = false
	} else {
		src.Valid = true
		src.String = spell.Source
	}
	if spell.Bloodline == "" {
		bloodline.Valid = false
	} else {
		bloodline.Valid = true
		bloodline.String = spell.Bloodline
	}
	if spell.Patron == "" {
		patron.Valid = false
	} else {
		patron.Valid = true
		patron.String = spell.Patron
	}

	slalevel.Valid = false
	for _, clvl := range spell.ClassLevels {
		if clvl.Class == "SLA" {
			slalevel.Valid = true
			slalevel.Int32 = int32(clvl.Level)
			break
		}
	}

	if spell.Components.HasCostlyComponents {
		materialcosts.Valid = false
	} else {
		materialcosts.Valid = true
		materialcosts.Int32 = int32(spell.Components.MaterialCosts)
	}

	if castingTimes, err := getCastingTimeList(db, prefs); err != nil || !slices.Contains(castingTimes, spell.Casting.Time) {
		return fmt.Errorf("invalid casting time \"%s\" (database error %v)", spell.Casting.Time, err)
	}

	if spell.Range.Range != "" {
		if ranges, err := getRangeList(db, prefs); err != nil || !slices.Contains(ranges, spell.Range.Range) {
			return fmt.Errorf("invalid range \"%s\" (database error %v)", spell.Range.Range, err)
		}
		if spell.Range.Range != "feet" && spell.Range.Range != "miles" {
			distance.Valid = false
		} else {
			distance.Valid = true
			distance.Int32 = int32(spell.Range.Distance)
			if spell.Range.DistancePerLevel == 0 {
				distperlevel.Valid = false
			} else {
				distperlevel.Valid = true
				distperlevel.Int32 = int32(spell.Range.DistancePerLevel)
			}
		}
	}

	if durations, err := getDurationList(db, prefs); err != nil || !slices.Contains(durations, spell.Duration.Duration) {
		return fmt.Errorf("invalid duration \"%s\" (database error %v)", spell.Duration.Duration, err)
	}

	spell.SR.SR = strings.ToLower(spell.SR.SR)
	if ress, err := getSpellResistanceList(db, prefs); err != nil || !slices.Contains(ress, spell.SR.SR) {
		return fmt.Errorf("invalid SR \"%s\" (database error %v)", spell.SR.SR, err)
	}

	if spell.Save.SavingThrow != "" {
		if saves, err := getSavingThrowList(db, prefs); err != nil || !slices.Contains(saves, spell.Save.SavingThrow) {
			return fmt.Errorf("invalid saving throw \"%s\" (database error %v)", spell.Save.SavingThrow, err)
		}
	}

	if spell.Save.Effect != "" {
		if se, err := getSaveEffectsList(db, prefs); err != nil || !slices.Contains(se, spell.Save.Effect) {
			return fmt.Errorf("invalid saving throw effect \"%s\" (database error %v)", spell.Save.Effect, err)
		}
	}

	if exists {
		_, err2 = db.Exec(`DELETE FROM SpellLevels WHERE SpellID=?`, id)
		_, err = db.Exec(`
			UPDATE Spells 
			SET 
				IsLocal=?, Code=?, Name=?, SchoolID=?, Descriptors=?, Components=?,
				Material=?, Focus=?, CastingTime=?, CastingSpec=?, Range=?, Distance=?,
				DistPerLevel=?, DistSpec=?, Area=?, Effect=?, Targets=?, Duration=?,
				DurationTime=?, DurationSpec=?, DurationConc=?, DurationPerLvl=?,
				SR=?, SRSpec=?, SRObject=?, SRHarmless=?, SavingThrow=?, SaveEffect=?,
				SaveSpec=?, SaveObject=?, SaveHarmless=?, IsDismissible=?, IsDischarge=?,
				IsShapeable=?, HasCostlyComponents=?, SLALevel=?, Deity=?,
				Domain=?, Description=?, Source=?, MaterialCosts=?, Bloodline=?, 
				Patron=?
			WHERE
				ID=?`,
			spell.IsLocal, spell.Code, spell.Name, schoolID, descriptors, components,
			material, focus, spell.Casting.Time, castspec, range_, distance, distperlevel,
			distspec, area, effect, targets, spell.Duration.Duration, durtime, durspec,
			spell.Duration.Concentration, spell.Duration.PerLevel,
			spell.SR.SR, srspec, spell.SR.Object, spell.SR.Harmless,
			save, saveeffect, savespec, spell.Save.Object, spell.Save.Harmless,
			spell.IsDismissible, spell.IsDischarge, spell.IsShapeable,
			spell.Components.HasCostlyComponents, slalevel, deity, domain,
			description, src, materialcosts, bloodline, patron, id)
	} else {
		if res, err = db.Exec(`
			INSERT INTO Spells (
				IsLocal, Code, Name, SchoolID, Descriptors, Components,
				Material, Focus, CastingTime, CastingSpec, Range, Distance,
				DistPerLevel, DistSpec, Area, Effect, Targets, Duration,
				DurationTime, DurationSpec, DurationConc, DurationPerLvl,
				SR, SRSpec, SRObject, SRHarmless, SavingThrow, SaveEffect,
				SaveSpec, SaveObject, SaveHarmless, IsDismissible, IsDischarge,
				IsShapeable, HasCostlyComponents, SLALevel, Deity,
				Domain, Description, Source, MaterialCosts, Bloodline, 
				Patron
			)`,
			spell.IsLocal, spell.Code, spell.Name, schoolID, descriptors, components,
			material, focus, spell.Casting.Time, castspec, range_, distance, distperlevel,
			distspec, area, effect, targets, spell.Duration.Duration, durtime, durspec,
			spell.Duration.Concentration, spell.Duration.PerLevel,
			spell.SR.SR, srspec, spell.SR.Object, spell.SR.Harmless,
			save, saveeffect, savespec, spell.Save.Object, spell.Save.Harmless,
			spell.IsDismissible, spell.IsDischarge, spell.IsShapeable,
			spell.Components.HasCostlyComponents, slalevel, deity, domain,
			description, src, materialcosts, bloodline, patron); err != nil {
			return err
		}

		if id, err = res.LastInsertId(); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	if (prefs.DebugBits & DebugMisc) != 0 {
		if exists {
			log.Printf("updated existing spell %s (database id %d)", spell.Name, id)
		} else {
			log.Printf("created new spell %s (database id %d)", spell.Name, id)
		}
	}

	// add classes and levels
	classIDs, err := getClassNameIDs(db, prefs)
	if err != nil {
		return err
	}
	for _, scl := range spell.ClassLevels {
		if scl.Class != "SLA" {
			cid, ok := classIDs[scl.Class]
			if !ok {
				return fmt.Errorf("class \"%s\" for spell \"%s\" is not defined", scl.Class, spell.Name)
			}

			if _, err = db.Exec(`
				INSERT INTO SpellLevels (
					SpellID, ClassID, Level
				) 
				VALUES (?,?,?)
			`, id, cid, scl.Level); err != nil {
				return err
			}
		}
	}

	return err
}

// ExportSpells exports the Spell objects from the database, writing them to the open JSON file.
func ExportSpells(fp *os.File, db *sql.DB, prefs *CorePreferences) error {
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
			Spells.ID, IsLocal, Spells.Name, Spells.Code, Schools.Code, Descriptors, Components, Material,
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

		if err := rows.Scan(&spell_db_id, &spell.IsLocal, &spell.Name, &spell.Code, &spell.School, &descriptors, &components,
			&material, &focus, &spell.Casting.Time, &cspec, &rang, &dist, &dpl,
			&distspec, &area, &effect, &targs, &spell.Duration.Duration, &dtime, &dspec, &spell.Duration.Concentration, &spell.Duration.PerLevel,
			&spell.SR.SR, &srspec, &spell.SR.Object, &spell.SR.Harmless, &save, &seffect, &sspec, &spell.Save.Object, &spell.Save.Harmless,
			&spell.IsDismissible, &spell.IsDischarge, &spell.IsShapeable, &spell.Components.HasCostlyComponents, &slalvl,
			&deity, &domain, &desc, &source, &mcosts, &bline, &patron); err != nil {
			return err
		}

		if filterOut(prefs, TypeSpell, "spell", spell.Name, spell.Name, spell.IsLocal) {
			continue
		}

		if !firstLine {
			if _, err = fp.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			firstLine = false
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

/*
# @[00]@| Go-GMA 5.25.2
# @[01]@|
# @[10]@| Overall GMA package Copyright © 1992–2024 by Steven L. Willoughby (AKA MadScienceZone)
# @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
# @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
# @[13]@| points along that historical time line.
# @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
# @[15]@| License as described in the accompanying LICENSE file distributed
# @[16]@| with GMA.
# @[17]@|
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
