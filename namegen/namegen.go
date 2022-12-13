/*
########################################################################################
#  _______  _______  _______                ___       ______      _______              #
# (  ____ \(       )(  ___  )              /   )     / ___  \    (  __   )             #
# | (    \/| () () || (   ) |             / /) |     \/   )  )   | (  )  |             #
# | |      | || || || (___) |            / (_) (_        /  /    | | /   |             #
# | | ____ | |(_)| ||  ___  |           (____   _)      /  /     | (/ /) |             #
# | | \_  )| |   | || (   ) | Game           ) (       /  /      |   / | |             #
# | (___) || )   ( || )   ( | Master's       | |   _  /  /     _ |  (__) |             #
# (_______)|/     \||/     \| Assistant      (_)  (_) \_/     (_)(_______)             #
#                                                                                      #
########################################################################################
*/

//
// Package namegen implements random name generation.
//
// It builds names according to letter and phrase patterns representative of naming  conventions
// used by various Golarion cultures.
//
// This code is a port of John Mechalas's JavaScript name generator from https://dungeonetics.com/pfnames/
// and includes a starter set of his Golarion cultural name data.
//
// To quote John's own documentation: "The name generators presented here are based on Markov chains,
// and construct names that tend to follow the same letter/syllable combinations and distributions as
// the source names from which they were seeded. Note that these algorithms are not perfect, and you
// might need to generate several names before settling on one that you like: it is the nature of
// the Markov chain to occasionally produce silly gibberish, names that do not 'fit', or even
// 'real' names. These accidents are part of the fun."
//
// Note that the names are "gender"-based. This term is used simply to differentiate cultural
// naming patterns, which often follow the gender expression of the individual, but many cultures
// have other naming variations based on social constructs such as family/clan traditions, region,
// economic and political status, religion, etc. All of these are "genders" as far as this module
// is concerned, and an arbitrary number of them may be defined for any given culture.
//
// The default "genders" included for the supplied set of cultures includes 'F' for female names,
// 'M' for male names, and 'S' for surnames. Not all cultures implement all of these.
//
// The limits on name length are not hard rules but are rather goals for the generator
// to try for. Also note that there are times when the generator may give up before
// generating the quantity of names requested in order to avoid getting into a loop
// that takes too much time to complete.
//
// This code, and specifically the individual Culture definition source files, uses
// trademarks and/or copyrights owned by Paizo Inc., used under Paizo's Community Use
// Policy (paizo.com/communityuse). We are expressly prohibited from charging you to
// use or access this content. GMA is not published, endorsed, or specifically approved
// by Paizo. For more information about Paizo Inc. and Paizo products, visit paizo.com.
//
package namegen

import (
	"fmt"
	"strings"

	"github.com/MadScienceZone/go-gma/v4/dice"
)

//
// Cultures lists the cultures that are defined in this package as distributed.
// Programs may use this to offer a choice of cultures to the user or to get a
// value which can be passed to Generate or GenerateWithSurnames.
//
// The keys of this map are culture names. Their associated values are values
// of the corresponding Culture type.
//
var Cultures = map[string]Culture{
	"Azlanti":    Azlanti{},
	"Bekyar":     Bekyar{},
	"Bonuwat":    Bonuwat{},
	"Chelaxian":  Chelaxian{},
	"Dwarf":      Dwarf{},
	"Elf":        Elf{},
	"Erutaki":    Erutaki{},
	"Garundi":    Garundi{},
	"Gnome":      Gnome{},
	"Half-orc":   HalfOrc{},
	"Halfling":   Halfling{},
	"Keleshite":  Keleshite{},
	"Kellid":     Kellid{},
	"Kitsune":    Kitsune{},
	"Shoanti":    Shoanti{},
	"Taldan":     Taldan{},
	"Tian-dan":   TianDan{},
	"Tian-dtang": TianDtang{},
	"Tian-hwan":  TianHwan{},
	"Tian-la":    TianLa{},
	"Tian-min":   TianMin{},
	"Tian-shu":   TianShu{},
	"Tian-sing":  TianSing{},
	"Ulfen":      Ulfen{},
	"Varisian":   Varisian{},
	"Vudrani":    Vudrani{},
	"Zenj":       Zenj{},
}

//
// Culture is any specific cultural group with a distinctive naming convention.
//
type Culture interface {
	Name() string
	Genders() []rune
	HasSurnames() bool
	HasGender(rune) bool
	nameWords(rune) int
	prefix(rune) string
	defaultMinMax(rune) (int, int)
	optPfx(rune) []string
	maxCount(rune, rune) int
	db(rune) map[string][]nameFragment
}

//
// BaseCulture provides a baseline common to all cultures.
// Each individual culture should include BaseCulture.
//
type BaseCulture struct {
}

func (c BaseCulture) nameWords(gender rune) int {
	return 1
}

func (c BaseCulture) prefix(gender rune) string {
	return "___"
}

func (c BaseCulture) defaultMinMax(gender rune) (int, int) {
	return 1, 1
}

//
// Genders returns a list of gender codes supported by this Culture.
//
func (c BaseCulture) Genders() []rune {
	return nil
}

//
// HasSurnames returns true if this Culture implements surnames.
//
func (c BaseCulture) HasSurnames() bool {
	return false
}

//
// Name returns a human-readable name of the Culture.
//
func (c BaseCulture) Name() string {
	return "(base culture)"
}

func (c BaseCulture) optPfx(gender rune) []string {
	return nil
}

func (c BaseCulture) maxCount(gender, char rune) int {
	return 0
}

func (c BaseCulture) db(gender rune) map[string][]nameFragment {
	return nil
}

//
// generateOptions holds the configuration data to control
// how we generate names for a particular run.
//
type generateOptions struct {
	dieRoller      *dice.DieRoller
	startingLetter rune
	minLength      int
	maxLength      int
}

//
// WithStartingLetter modifies a Generate or GenerateWithSurnames function call by
// specifying the initial letter for the names to be generated. If this is 0
// or this option is not present, names starting with any letter may be created.
//
func WithStartingLetter(start rune) func(*generateOptions) {
	return func(o *generateOptions) {
		o.startingLetter = start
	}
}

//
// WithMinLength modifies a Generate or GenerateWithSurnames function call by
// specifying the minimum length for the names to be generated. If the value
// given is 0 or this option is not present, the minimum length defined by
// the culture will be used.
//
func WithMinLength(minlen int) func(*generateOptions) {
	return func(o *generateOptions) {
		o.minLength = minlen
	}
}

//
// WithMaxLength modifies a Generate or GenerateWithSurnames function call by
// specifying the maximum length for the names to be generated. If the value
// given is 0 or this option is not present, the maximum length defined by
// the culture will be used.
//
func WithMaxLength(maxlen int) func(*generateOptions) {
	return func(o *generateOptions) {
		o.maxLength = maxlen
	}
}

//
// WithDieRoller modifies a Generate or GenerateWithSurnames function call by
// specifying the DieRoller to use for random number generation. If this is
// not specified, or if a nil value is passed, a standard pseudorandom number
// generator will be employed. However, this option is provided in case you
// are generating names along with other character data that must remain
// consistent from one run of the program to the next, and as such you already
// have a seeded DieRoller in use, that you wish to use for the names as well.
//
func WithDieRoller(dr *dice.DieRoller) func(*generateOptions) {
	return func(o *generateOptions) {
		o.dieRoller = dr
	}
}

//
// HasGender returns true if the specified gender code is defined
// for this culture.
//
func (c BaseCulture) HasGender(gender rune) bool {
	for _, g := range c.Genders() {
		if g == gender {
			return true
		}
	}
	return false
}

//
// initGenerate provides common startup code for Generate and GenerateWithSurnames
//
func initGenerate(c Culture, gender rune, options []func(*generateOptions)) (rune, generateOptions, error) {
	var err error

	if gender == 0 {
		if c.HasGender('F') {
			gender = 'F'
		} else {
			gender = c.Genders()[0]
		}
	}
	if !c.HasGender(gender) {
		return 0, generateOptions{}, fmt.Errorf("culture %s does not have gender %c", c.Name(), gender)
	}

	lmin, lmax := c.defaultMinMax(gender)

	opts := generateOptions{
		minLength: lmin,
		maxLength: lmax,
	}

	for _, o := range options {
		o(&opts)
	}

	if opts.dieRoller == nil {
		opts.dieRoller, err = dice.NewDieRoller()
		if err != nil {
			return 0, opts, err
		}
	}
	return gender, opts, nil
}

//
// Generate creates the requested quantity of random names for the given culture and gender code.
// These are returned as a slice of name strings.
//
func Generate(c Culture, gender rune, qty int, options ...func(*generateOptions)) ([]string, error) {
	var err error
	gender, opts, err := initGenerate(c, gender, options)
	if err != nil {
		return nil, err
	}

	var names []string
	for i := 0; i < qty; i++ {
		retries := 256

	generateName:
		for retries > 0 {
			name, err := doGenerate(c, gender, opts.startingLetter, opts.minLength, opts.maxLength, opts.dieRoller)
			if err != nil {
				return nil, err
			}
			for _, n := range names {
				if n == name {
					retries-- // duplicate name, try again (but not forever)
					continue generateName
				}
			}
			prefixes := c.optPfx(gender)
			if len(prefixes) > 0 {
				if opts.dieRoller.RandIntn(100) < 5 {
					choice := 0
					if len(prefixes) > 1 {
						choice = opts.dieRoller.RandIntn(len(prefixes))
					}
					name = prefixes[choice] + name
				}
			}
			names = append(names, name)
			break
		}
	}
	return names, nil
}

//
// GenerateWithSurnames is just like Generate, but for every given name generated it also
// creates a surname. Names are returned as a slice of names, where each name is a slice of two
// strings (given name and surname).
//
func GenerateWithSurnames(c Culture, gender rune, qty int, options ...func(*generateOptions)) ([][]string, error) {
	var err error
	gender, opts, err := initGenerate(c, gender, options)
	if err != nil {
		return nil, err
	}

	if !c.HasSurnames() {
		return nil, fmt.Errorf("%s culture does not define surnames", c.Name())
	}

	names, err := Generate(c, gender, qty,
		WithMinLength(opts.minLength),
		WithMaxLength(opts.maxLength),
		WithDieRoller(opts.dieRoller))
	if err != nil {
		return nil, err
	}

	surnameList, err := Generate(c, 'S', len(names),
		WithMinLength(opts.minLength),
		WithMaxLength(opts.maxLength),
		WithDieRoller(opts.dieRoller),
	)
	if err != nil {
		return nil, fmt.Errorf("error while generating surnames: %v", err)
	}
	if len(surnameList) == 0 {
		return nil, fmt.Errorf("no surnames generated")
	}

	fullNames := make([][]string, 0, len(names))
	for i, first := range names {
		fullName := make([]string, 2, 2)
		fullName[0] = first
		if i >= len(surnameList) {
			fullName[1] = surnameList[len(surnameList)-1]
		} else {
			fullName[1] = surnameList[i]
		}
		fullNames = append(fullNames, fullName)
	}

	return fullNames, nil
}

//
// nameFragment describes a suffix rune that may be added to a name,
// with the probability in the range [0.0, 1.0] of that suffix being
// chosen.
//
// Suffix may be 0 to indicate the chance of no suffix at all being
// added.
//
type nameFragment struct {
	Suffix      rune
	Probability float64
}

func doGenerate(c Culture, gender, start rune, lmin, lmax int, dr *dice.DieRoller) (string, error) {
	genderData := c.db(gender)
	if genderData == nil {
		return "", fmt.Errorf("culture %s does not have any name data for gender %c", c.Name(), gender)
	}

	repeat := c.nameWords(gender)
	if repeat <= 1 {
		for retries := 0; retries < 256; retries++ {
			name := generatePart(c, gender, genderData, start, lmin, lmax, dr)
			if name != "" {
				return name, nil
			}
		}
		return "", fmt.Errorf("name generator gave up without generating a name")
	}

	retries := 256
	names := make([]string, 0, repeat)
	for len(names) < repeat {
		var thisStart rune
		if len(names) == 0 {
			thisStart = start
		}
		part := generatePart(c, gender, genderData, thisStart, lmin, lmax, dr)
		if part == "" {
			if retries <= 0 {
				return "", fmt.Errorf("name generator gave up without generating a name")
			}
			retries--
		} else {
			names = append(names, part)
		}
	}

	return strings.Join(names, " "), nil
}

func generatePart(c Culture, gender rune, genderData map[string][]nameFragment, start rune, lmin, lmax int, dr *dice.DieRoller) string {
	sPrefix := c.prefix(gender)
	repeat := 10
	chcount := make(map[rune]int)

	if start != 0 {
		// replace last character of sPrefix with start
		// (allows for sPrefix to be empty, in which case it will just be start)
		sPrefix = (sPrefix + string(start))
		if len(sPrefix) > 1 {
			sPrefix = sPrefix[1:]
		}
	}

	for repeat > 0 {
		prefix := sPrefix
		name := ""
		if start != 0 {
			name = string(start)
		}

		for {
			suffix := pickSuffix(genderData, prefix, dr)
			if suffix == 0 {
				if len(name) < lmin {
					repeat--
				} else {
					return name
				}
			} else {
				// is this character one that has a max count?
				maxc := c.maxCount(gender, suffix)
				if maxc > 0 {
					_, ok := chcount[suffix]
					if ok {
						chcount[suffix]++
					} else {
						chcount[suffix] = 1
					}

					if chcount[suffix] > maxc {
						repeat--
						suffix = 0
					}
				}

				if suffix != 0 {
					name += string(suffix)
					prefix += string(suffix)
					if len(prefix) > 1 {
						prefix = prefix[1:]
					}
					if len(name) > lmax {
						repeat--
						suffix = 0
					}
				}
			}

			if suffix == 0 {
				break
			}
		}
		repeat--
	}
	return ""
}

func pickSuffix(genderData map[string][]nameFragment, prefix string, dr *dice.DieRoller) rune {
	choices, ok := genderData[prefix]
	if !ok || len(choices) == 0 {
		return 0
	}

	r := dr.RandFloat64()
	for _, choice := range choices {
		if choice.Probability >= r {
			return choice.Suffix
		}
	}
	return choices[len(choices)-1].Suffix
}

<<<<<<< HEAD
// @[00]@| GMA 4.4.1
||||||| 52d71f1
// @[00]@| GMA 4.3.10
=======
// @[00]@| GMA 4.7.0
>>>>>>> json
// @[01]@|
// @[10]@| Copyright © 1992–2022 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved.
// @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[14]@| License as described in the accompanying LICENSE file distributed
// @[15]@| with GMA.
// @[16]@|
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
