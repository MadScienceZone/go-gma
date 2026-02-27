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

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -sqlite=game.db or -sqlite game.db), except for boolean flags which may be given alone (e.g., -delete) to indicate that the option is set to “true” or may be given an explicit value which must be attached to the option with an equals sign (e.g., -delete=true or -delete=false).

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
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const GoVersionNumber="5.33.0" //@@##@@

type FileType byte

const (
	ImageFile FileType = iota
	MapFile
	UnknownFile
)

type MapFileDescription struct {
	Type       FileType
	InDatabase bool
	OnDisk     bool
	OnClient   bool
	DiskPath   string
	MapperID   string
	Error      error
	Frames     []int
	Formats    []string
	MapperName string
	MapperZoom float64
}

var (
	mapFilePattern = regexp.MustCompile(`^(.)/(..?)/(?::(\d+):)?(.*?)\.(\w+)$`)
)

// parsePath takes a disk path of the form "a/ab/abc.ext" and returns
// the base server ID ("abc") by which the server will refer to it.
//
// A valid file called abcdef.ext will be stored in the path
// a/ab/abcdef.ext and will have serverID abcdef. If ext is "map"
// then it's a map file otherwise a type of image file.
func parsePath(diskPath string) (serverID string, d MapFileDescription, err error) {
	d.DiskPath = diskPath
	if matches := mapFilePattern.FindStringSubmatch(diskPath); matches != nil {
		if len(matches) != 6 {
			err = fmt.Errorf("unexpected number of matches %d", len(matches))
			return
		}

		serverID = matches[4]
		d.Formats = []string{matches[5]}
		d.OnDisk = true
		d.Type = ImageFile
		if matches[3] != "" {
			var f int
			if f, err = strconv.Atoi(matches[3]); err != nil {
				return
			}
			d.Frames = []int{f}
		}

		if matches[5] == "map" {
			d.Type = MapFile
		}

		if matches[5] == "" {
			err = fmt.Errorf("missing file type suffix")
			return
		}

		if len(serverID) < 2 && (matches[1] != serverID[:1] || matches[2] != serverID[:1]) {
			err = fmt.Errorf("file path %s invalid for GMA map file (id=%s, format=%s)", diskPath, serverID, matches[5])
			return
		}

		if matches[1] != serverID[:1] || matches[2] != serverID[:2] {
			err = fmt.Errorf("file path %s invalid for GMA map file (id=%s, format=%s)", diskPath, serverID, matches[5])
			return
		}
	} else {
		err = fmt.Errorf("unable to understand pathname pattern")
	}
	return
}

func getServedFiles(webRootDir string) (map[string]MapFileDescription, []MapFileDescription, error) {
	servedFiles := make(map[string]MapFileDescription)
	badFiles := []MapFileDescription{}

	err := fs.WalkDir(os.DirFS(webRootDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if serverID, thisFile, err := parsePath(path); err != nil {
			thisFile.Error = err
			thisFile.Type = UnknownFile
			badFiles = append(badFiles, thisFile)
		} else {
			entry, exists := servedFiles[serverID]
			if !exists {
				servedFiles[serverID] = thisFile
			} else {
				if entry.Type != thisFile.Type {
					thisFile.Error = fmt.Errorf("type mismatch with other file(s) of the same server ID")
					badFiles = append(badFiles, thisFile)
					return nil
				}

				if len(thisFile.Frames) > 0 && !slices.Contains(entry.Frames, thisFile.Frames[0]) {
					entry.Frames = append(entry.Frames, thisFile.Frames[0])
				}
				if len(thisFile.Formats) > 0 && !slices.Contains(entry.Formats, thisFile.Formats[0]) {
					entry.Formats = append(entry.Formats, thisFile.Formats[0])
				}
				servedFiles[serverID] = entry
			}
		}
		return nil
	})
	return servedFiles, badFiles, err
}

func searchForImages(servedFiles map[string]MapFileDescription, badFiles []MapFileDescription, dbPath string, remove bool) error {
	db, err := sql.Open("sqlite3", "file:"+dbPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	rows, err := db.Query("SELECT name, zoom, location, islocal, frames FROM images")
	if err != nil {
		return err
	}
	defer rows.Close()
	cullThese := []MapFileDescription{}

	for rows.Next() {
		var serverID string
		var frames int
		thisEntry := MapFileDescription{}
		if err := rows.Scan(&thisEntry.MapperName, &thisEntry.MapperZoom, &serverID, &thisEntry.OnClient, &frames); err != nil {
			return err
		}
		thisEntry.InDatabase = true
		thisEntry.MapperID = fmt.Sprintf("%s@%.2f", thisEntry.MapperName, thisEntry.MapperZoom)
		thisEntry.Type = ImageFile
		if existingEntry, exists := servedFiles[serverID]; exists {
			if existingEntry.Type != thisEntry.Type {
				thisEntry.Error = fmt.Errorf("type mismatch with other file(s) of the same server ID")
				badFiles = append(badFiles, thisEntry)
				continue
			}
			existingEntry.InDatabase = true
			existingEntry.MapperID = thisEntry.MapperID
			existingEntry.OnClient = thisEntry.OnClient
			servedFiles[serverID] = existingEntry
		} else {
			servedFiles[serverID] = thisEntry
			if remove && !thisEntry.OnClient {
				cullThese = append(cullThese, thisEntry)
			}
		}
	}

	for _, thisEntry := range cullThese {
		if _, err := db.Exec("DELETE FROM images WHERE name = ? AND zoom = ?", thisEntry.MapperName, thisEntry.MapperZoom); err != nil {
			return err
		}
		log.Printf("*** REMOVED SERVER RECORD FOR IMAGE %s AT %.3f ZOOM ***", thisEntry.MapperName, thisEntry.MapperZoom)
	}
	return nil
}

func main() {
	var sqlDbName = flag.String("sqlite", "", "Specify filename for sqlite database to use")
	var webRootDir = flag.String("webroot", "", "Specify image base directory for served image files")
	var listFiles = flag.Bool("list", false, "List all the files in database and disk")
	var delImages = flag.Bool("delete", false, "Delete database images without files")
	flag.Parse()

	if webRootDir == nil || *webRootDir == "" {
		log.Fatal("-webroot option is required")
	}
	if sqlDbName == nil || *sqlDbName == "" {
		log.Fatal("-sqlite option is required")
	}
	if *delImages {
		log.Print("WARNING: will delete database entries with missing files!")
	}

	servedFiles, unknownFiles, err := getServedFiles(*webRootDir)
	if err != nil {
		log.Fatal("fatal error:", err)
	}

	if err := searchForImages(servedFiles, unknownFiles, *sqlDbName, *delImages); err != nil {
		log.Fatal("fatal error:", err)
	}

	if len(unknownFiles) > 0 {
		log.Printf("Invalid filenames found:     %6d", len(unknownFiles))
		if *listFiles {
			for _, f := range unknownFiles {
				log.Printf("%s: %v", f.DiskPath, f.Error)
			}
		}
	}

	log.Printf("Served files discovered:     %6d", len(servedFiles))
	if *listFiles {
		log.Print("                                                                 type (I=image, M=map)")
		log.Print("                                                                /database entry present (d) or client-stored (c)")
		log.Print("                                                               //web disk file present")
		log.Print("                                                              ///")
		log.Print("SERVER-ID--------------------------------------------------- tdw FRM FORMATS")
	}
	missingDiskFiles := 0
	missingDatabase := 0
	clientFiles := 0
	for serverID, d := range servedFiles {
		if !d.OnDisk && !d.OnClient {
			missingDiskFiles++
			log.Printf("*** IMAGE %s *** MISSING FROM DISK! Should be %s", d.MapperID, serverID)
		}
		if !d.InDatabase {
			missingDatabase++
		}
		if d.OnClient {
			clientFiles++
		}
		if *listFiles {
			flags := "-"
			switch d.Type {
			case UnknownFile:
				flags = "?"
			case ImageFile:
				flags = "I"
			case MapFile:
				flags = "M"
			}
			if d.InDatabase {
				if d.OnClient {
					flags += "c"
				} else {
					flags += "d"
				}
			} else {
				flags += "-"
			}
			if d.OnDisk {
				flags += "w"
			} else {
				flags += "-"
			}
			log.Printf("%-60s %s %3d %v", serverID, flags, len(d.Frames), d.Formats)
		}
	}
	log.Printf("Files missing from database: %6d", missingDatabase)
	log.Printf("Files missing from web dirs: %6d", missingDiskFiles)
}
