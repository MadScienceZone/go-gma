/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______   ______      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \ / ___  \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \\/   \  \   | (  )  |     #
# | |      | || || || (___) | Assistant | (____        ___) /   ___) /   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (   (___ (    | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )         ) \      ) \   |   / | |     #
# | (___) || )   ( || )   ( |           /\____) ) _ /\___/  //\___/  / _ |  (__) |     #
# (_______)|/     \||/     \|           \______/ (_)\______/ \______/ (_)(_______)     #
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
Download-presets connects to a GMA server and retrieves the die-roll presets stored there for a user, saving them to a local disk file.

# OPTIONS

The following options control the action of upload-presets.

	−endpoint [hostname]: port
	   Connect to the server at the specified TCP port.

	−for username
	   Retrieve the presets for username instead of the user you are logged in as.

	−pass password
	   Log in to the server with the specified password

	−user username
	   Log in to the server with the specified username (default “GM”).

	-global
	   Retrieve the set of system-wide presets rather than those belonging to a particular user.

	-output filename
	   Save the data to the specified filename.
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
	var fGlobal = flag.Bool("global", false, "Retrieve system-wide global presets instead of a user's personal set")
	var fOutput = flag.String("output", "", "Filename to which to save the retrieved data.")
	var err error

	flag.Parse()
	if *fOutput == "" {
		fmt.Printf("You need to specify a destination filename with -output.\n")
		os.Exit(1)
	}

	if *fGlobal && *fFor != "" {
		fmt.Printf("You cannot specify both -global and -for at the same time.\n")
		os.Exit(1)
	}

	presets := make(chan mapper.MessagePayload, 1)
	sync := make(chan mapper.MessagePayload, 1)
	ready := make(chan byte, 1)

	server, err := mapper.NewConnection(*fEndpoint,
		mapper.WithAuthenticator(auth.NewClientAuthenticator(*fUser, []byte(*fPass), "upload-presets")),
		mapper.WithSubscription(sync, mapper.Echo),
		mapper.WithSubscription(presets, mapper.UpdateDicePresets),
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

	// absorb any die-roll preset info the server may send us as part of the game state info
	// it wants to send as part of the client sign-on or sync info
	fmt.Println("Synchronizing with server.")
	if err := server.EchoString("xyzzy"); err != nil {
		fmt.Printf("Can't send echo to server: %v\n", err)
		os.Exit(1)
	}
serverSync:
	for {
		select {
		case e := <-sync:
			if syncVal, ok := e.(mapper.EchoMessagePayload); ok {
				if syncVal.S == "xyzzy" {
					break serverSync
				}
				fmt.Println("...%s\n", syncVal.S)
			} else {
				fmt.Println("Unexpected sync message type %T received\n", e)
				fmt.Println("Not sure we can recover from this.")
			}
		case <-presets:
			// discard quietly
		}
	}
	fmt.Println("Ready.")

	if *fGlobal {
		err = server.QueryGlobalDicePresets()
	} else if *fFor != "" {
		err = server.QueryDicePresetsFor(*fFor)
	} else {
		err = server.QueryDicePresets()
	}
	if err != nil {
		fmt.Printf("Error querying server: %v\n", err)
		os.Exit(2)
	}

	fmt.Println("Fetching preset data from server...")
	data := <-presets
	if pdata, ok := data.(mapper.UpdateDicePresetsMessagePayload); ok {
		outputFile, err := os.Create(*fOutput)
		if err != nil {
			fmt.Printf("Unable to write to %s: %v\n", *fOutput, err)
			os.Exit(3)
		}
		err = dice.SaveDieRollPresetFile(outputFile, pdata.Presets, dice.DieRollPresetMetaData{
			Comment: fmt.Sprintf("%s retrieved from server %s", func() string {
				if *fGlobal {
					return "global system-wide presets"
				} else if *fFor != "" {
					return "die-roll presets for user " + *fFor
				} else {
					return "personal die-roll presets"
				}
			}(), *fEndpoint),
		})
		if err != nil {
			fmt.Printf("Error saving data to file %s: %v\n", *fOutput, err)
			os.Exit(3)
		}
		err = outputFile.Close()
		if err != nil {
			fmt.Printf("Error saving data to file %s: %v\n", *fOutput, err)
			os.Exit(3)
		}
	}
	fmt.Printf("Done.\n")
	server.Close()
}
