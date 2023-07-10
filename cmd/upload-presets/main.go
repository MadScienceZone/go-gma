/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      _____      _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ___ \    (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (___) )   | (  )  |              #
# | |      | || || || (___) | Assistant | (____      \     /    | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \     / ___ \    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )   ( (   ) )   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ ( (___) ) _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_) \_____/ (_)(_______)              #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2022 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Upload-presets connects to a GMA server and uploads the die-roll presets from one or more files, so they are stored in the server for use during game play.

The files must be in the format documented by dice(5), in either format version 1 or 2.

OPTIONS

The following options control the action of upload-presets.

   −endpoint [hostname]: port
      Connect to the server at the specified TCP port.

   −for username
      Store the presets for username instead of the user you are logged in as.

   −pass password
      Log in to the server with the specified password

   −replace
      Replace the user’s entire set of presets with the uploaded data.
      By default the uploaded presets will be added to any that already exist for that user.

   −user username
      Log in to the server with the specified username (default “GM”).
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/mapper"
)

func main() {
	var fEndpoint = flag.String("endpoint", "", "endpoint of server to upload the presets into")
	var fUser = flag.String("user", "GM", "username to log in to server as [default=GM]")
	var fPass = flag.String("pass", "", "password to log in to server")
	var fFor = flag.String("for", "", "who to load the presets for [default is yourself]")
	var fAdd = flag.Bool("replace", false, "replace all existing presets [default is to add to the existing set]")

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("You need to specify at least one preset file to be loaded.\n")
	}

	sync := make(chan mapper.MessagePayload, 1)
	ready := make(chan byte, 1)

	server, err := mapper.NewConnection(*fEndpoint,
		mapper.WithAuthenticator(auth.NewClientAuthenticator(*fUser, []byte(*fPass), "upload-presets")),
		mapper.WithSubscription(sync, mapper.Echo),
		//		mapper.WithDebugging(mapper.DebugAll),
		mapper.WhenReady(ready),
	)
	if err != nil {
		fmt.Printf("can't set up server connection: %v\n", err)
		os.Exit(1)
	}
	go server.Dial()
	fmt.Printf("Waiting for server to be ready\n")
	<-ready

	for _, inputFilename := range flag.Args() {
		presets, metaData, err := dice.ReadDieRollPresetFile(inputFilename)
		if err != nil {
			fmt.Printf("ERROR reading %s: %v\n", inputFilename, err)
			break
		}
		fmt.Printf("Loaded %s (%s) from %s: %d presets\n", inputFilename, metaData.Comment, metaData.DateTime, len(presets))
		if fFor == nil || *fFor == "" {
			if *fAdd {
				err = server.AddDicePresets(presets)
			} else {
				err = server.DefineDicePresets(presets)
			}
		} else {
			if *fAdd {
				err = server.AddDicePresetsFor(*fFor, presets)
			} else {
				err = server.DefineDicePresetsFor(*fFor, presets)
			}
		}
		if err != nil {
			fmt.Printf("ERROR sending dice presets: %v\n", err)
			break
		}
	}
	fmt.Printf("Server sync...\n")
	if err := server.EchoString("xyzzy"); err != nil {
		fmt.Printf("Can't send echo to server: %v\n", err)
		os.Exit(1)
	}
	<-sync
	fmt.Printf("Done.\n")
}
