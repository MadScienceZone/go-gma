//go:build instrumentation
// +build instrumentation

/*
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
*/

package main

import _ "github.com/newrelic/go-agent/v3/integrations/nrsqlite3"

const InstrumentCode = true
const DatabaseDriver = "nrsqlite3"
