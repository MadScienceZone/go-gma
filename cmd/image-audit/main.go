/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __     _____      _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \   / ___ \    (  __   )       #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) ( (   ) )   | (  )  |       #
# | |      | || || || (___) | Assistant | (____        | | ( (___) |   | | /   |       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | |  \____  |   | (/ /) |       #
# | | \_  )| |   | || (   ) |                 ) )      | |       ) |   |   / | |       #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_/\____) ) _ |  (__) |       #
# (_______)|/     \||/     \| Client    \______/ (_) \____/\______/ (_)(_______)       #
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
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Image-audit is run on a system which hosts both a GMA server and the web server
which offers the map image tiles to mapper clients (or at least has filesystem
access to those files).

It checks for any files in the server's database (which means they will be offered
by the server to any clients asking about those images) which do not have actual
image files being served for them.

# SYNOPSIS

(If using the full GMA core tool suite)
   gma go image-audit ...

(Otherwise)
   image-audit -help
   image-audit [-delete] [-list] -sqlite dbfile -webroot gma_web_dir

# OPTIONS

The command-line options described below may be introduced with either one or two hyphens (e.g., -delete or --delete).

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -sqlite=game.db or -sqlite game.db), except for boolean flags which may be given alone (e.g., -delete) to indicate that the option is set to ``true'' or may be given an explicit value which must be attached to the option with an equals sign (e.g., -delete=true or -delete=false).

   -delete
     Delete images from the server's database which do not actually appear in the database

   -help
      Print a command summary and exit.

   -list
      Print a list of images listed in the database and whether or not they
	  correspond to web server files. Also lists files in the web directory
	  which are not mentioned in the database.

   -sqlite dbfile
      Read dbfile as the game server's database file (as specified to the server program's -sqlite option)
	  Image-audit may be run while the server is also accessing this file, but note that updating the database
	  by either program may briefly lock the other out from accessing it. Therefore, it is best to run
	  image-audit when the server is shut down or at least quiescent.

   -webroot dir
      Consider all the files from dir down to be the directory structure behind what the GMA mapper client
	  knows as the "image base URL". Thus, a server image ID of "abcdef" corresponds to disk file
	  <dir>/a/ab/abcdef.png, et al.
*/
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
)

const GoVersionNumber = "5.19.0" //@@##@@

type ImageDescription struct {
	InDatabase bool
	OnDisk     bool
	DiskPath   string
}

// pathToID takes a disk path of the form "a/ab/abc.ext" and returns
// the base server ID ("abc") by which the server will refer to it.
func pathToID(diskPath string) (string, error) {
	return "", nil
}

func getServedFiles(webRootDir string) (map[string]ImageDescription, error) {
	err := fs.WalkDir(os.DirFS(webRootDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println("Found path", path, "d:", d)
		return nil
	})
	return nil, err
}

func main() {
	//var sqlDbName = flag.String("sqlite", "", "Specify filename for sqlite database to use")
	var webRootDir = flag.String("webroot", "", "Specify image base directory for served image files")
	//var listFiles = flag.Bool("list", false, "List all the files in database and disk")
	//var delImages = flag.Bool("delete", false, "Delete database images without files")
	flag.Parse()

	if webRootDir == nil || *webRootDir == "" {
		log.Fatal("-webroot option is required")
	}

	if _, err := getServedFiles(*webRootDir); err != nil {
		log.Fatal("fatal error:", err)
	}
}
