/*
########################################################################################
#  _______  _______  _______                ___          ___        __                 #
# (  ____ \(       )(  ___  )              /   )        /   )      /  \                #
# | (    \/| () () || (   ) |             / /) |       / /) |      \/) )               #
# | |      | || || || (___) |            / (_) (_     / (_) (_       | |               #
# | | ____ | |(_)| ||  ___  |           (____   _)   (____   _)      | |               #
# | | \_  )| |   | || (   ) | Game           ) (          ) (        | |               #
# | (___) || )   ( || )   ( | Master's       | |   _      | |   _  __) (_              #
# (_______)|/     \||/     \| Assistant      (_)  (_)     (_)  (_) \____/              #
#                                                                                      #
########################################################################################
*/

package util

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type SimpleConfigurationData map[string]string

//
// NewSimpleConfigurationData creates a ready-to-use SampleConfigurationData
// value which you can call Set, et al. directly without having read in a
// configuration from a file first.
//
func NewSimpleConfigurationData() SimpleConfigurationData {
	return make(SimpleConfigurationData)
}

//
// ParseSimpleConfig parses a minimal configuration file format
// used by the mapper that isn't a full INI file. Rather, it's
// a simple "key=value" collection with one entry per line in the file.
// The key must be alphanumeric (including underscores and hyphens),
// while the value may include any characters. Spaces before or after
// the key are ignored, as are spaces before or after the value.
//
// A key alone on a line (without an = sign) indicates a boolean true
// value for that key.
//
// Lines starting with a # sign (allowing for leading spaces before that)
// are ignored as comments.
//
func ParseSimpleConfig(inputFile io.Reader) (SimpleConfigurationData, error) {
	data := make(SimpleConfigurationData)
	commentPattern := regexp.MustCompile(`^\s*#`)
	kvPattern := regexp.MustCompile(`^\s*([a-zA-Z0-9_-]+)\s*(=(.*))?$`)
	lines := bufio.NewScanner(inputFile)
	for lines.Scan() {
		line := lines.Text()
		if commentPattern.MatchString(line) {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		kv := kvPattern.FindStringSubmatch(line)
		if kv != nil {
			_, exists := data[kv[1]]
			if exists {
				return nil, fmt.Errorf("duplicate configuration \"%v\"", line)
			}
			if kv[2] != "" {
				data[kv[1]] = strings.TrimSpace(kv[3])
			} else {
				data[kv[1]] = "1"
			}
		} else {
			return nil, fmt.Errorf("unable to parse configuration line \"%v\"", line)
		}
	}
	if err := lines.Err(); err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}
	return data, nil
}

// Get retrieves a string value from the configuration data.
// Returns the string value, or "" if the key does not exist, and
// a boolean indicating whether the value existed in the data.
func (c SimpleConfigurationData) Get(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	v, b := c[key]
	return v, b
}

// GetDefault retrieves a string value from the configuration data,
// or the supplied default value if no such key exists.
func (c SimpleConfigurationData) GetDefault(key, def string) (string, bool) {
	if c == nil {
		return def, false
	}
	v, exists := c[key]
	if !exists {
		return def, false
	}
	return v, true
}

// GetInt retrieves an integer value from the configuration data.
// Returns an error if the value does not exist or could not be
// converted to an integer.
func (c SimpleConfigurationData) GetInt(key string) (int, error) {
	if c == nil {
		return 0, fmt.Errorf("nil SimpleConfigurationData, thus no key \"%s\"", key)
	}
	v, exists := c[key]
	if !exists {
		return 0, fmt.Errorf("no such key in configuration data: \"%s\"", key)
	}
	return strconv.Atoi(v)
}

// GetIntDefault retrieves an integer value from the configuration data.
// Returns an error if the value could not be converted to an integer,
// or the given default value if the key could not be found.
func (c SimpleConfigurationData) GetIntDefault(key string, def int) (int, error) {
	if c == nil {
		return def, nil
	}
	v, exists := c[key]
	if !exists {
		return def, nil
	}
	return strconv.Atoi(v)
}

// GetBool retrieves a boolean value from the configuration data.
// Returns an error if the value does not exist or could not be
// converted to a boolean.
//
// This considers values "0", "false", "no", or "off" to be false,
// and non-zero integers, "true", "yes", or "on" to be true.
// Non-existent keys are considered to be false.
func (c SimpleConfigurationData) GetBool(key string) (bool, error) {
	if c == nil {
		return false, nil
	}
	v, exists := c[key]
	if !exists {
		return false, nil
	}
	switch strings.ToLower(v) {
	case "false", "no", "off", "0":
		return false, nil
	case "true", "yes", "on", "1":
		return true, nil
	default:
		i, err := strconv.Atoi(v)
		if err != nil {
			return false, err
		}
		return i != 0, nil
	}
}

// Set adds a key/value pair to the SimpleConfigurationData receiver.
// If key already exists, it will be replaced with this new value.
func (c SimpleConfigurationData) Set(key, value string) {
	c[key] = value
}

func (c SimpleConfigurationData) SetInt(key string, value int) {
	c[key] = strconv.Itoa(value)
}

// @[00]@| GMA 4.4.1
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
