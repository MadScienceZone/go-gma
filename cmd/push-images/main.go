/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______   _______      __          #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \ / ___   )    /  \         #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \\/   )  |    \/) )        #
# | |      | || || || (___) | Assistant | (____        ___) /    /   )      | |        #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (   _/   /       | |        #
# | | \_  )| |   | || (   ) |                 ) )         ) \ /   _/        | |        #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\___/  /(   (__/\ _  __) (_       #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ \_______/(_) \____/       #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2021-2023 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Push-images pushes the definition of all the images from map files named on the command line to the server so it and all currently-connected clients are aware of it and load the image files into their disk caches.

# SYNOPSIS

(If using the full GMA core tool suite)

	gma go push-images

(Otherwise)

	push-images -h
	push-images -help
	push-images [-D] [-H host] [-l logfile] [-P password] [-p port] [-S profile] [-u user] mapfile ...
	push-images [-debug] [-help] [-host host] [-log logfile] [-password password] [-port port] [-preferences configfile] [-select profile] [-username user] mapfile ...

# OPTIONS

The command-line options described below have a long form (e.g., -port) and a  short form (e.g., -p) which are equivalent.
In either case, the option may be introduced with either one or two hyphens (e.g., -port or --port).

Options which take parameter values may have the value separated from the option name by a space or an equals sign (e.g., -port=2323 or -port 2323), except for boolean flags which may be given alone (e.g., -D) to indicate that the option is set to “true” or may be given an explicit value which must be attached to the option with an equals sign (e.g., -D=true or -D=false).

You may not combine multiple single-letter options into a single composite argument, (e.g., the options -D and -m would need to be entered as two separate options, not as -Dm).

	  -D, -debug flags
	      Adds debugging messages to push-images's output. The flags
	      value is a comma-separated list of debug flag names, which
	      may be any of the following:

	      all      Enable all debugging messages
	      none     Disable all debugging messages
	      auth     Authentication operations
	      binary   Add hexdump output of network data
	      events   Show background events such as expiring timers and signals
	      i/o      Input/output operations used to get data in and out of the client
	      messages Server messages sent and received
	      misc     Miscellaneous debugging messages

	  -H, -host host
	      Specifies the server's hostname.

	  -h, -help
	      Print a command summary and exit.

	  -l, -log file
	      Write log messages to the named file instead of stdout.
	      Use "-" for the file to explicitly send to stdout.

	  -list-profiles
	      Write a list of profiles that are defined in the mapper preferences file
		  and exit.

	  -P, -password password
	      Authenticate to the server using the specified password.

	  -p, -port port
	      Specifies the server's TCP port number.

	  -preferences file
	      Load configuration info from the named file instead
		  of the default ~/.gma/mapper/preferences.json

	  -S, -select profile
	      Selects a server profile to use from the user's saved mapper preferences.

	  -u, -username user
	      Authenticate to the server using the specified user name.
*/
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/mapper"
	"github.com/MadScienceZone/go-gma/v5/tcllist"
	"github.com/MadScienceZone/go-gma/v5/util"
)

const GoVersionNumber = "5.32.1" //@@##@@

var Fhost string
var Fport uint
var Fpass string
var Fuser string
var Fconf string
var Fdebug string
var Flog string
var Fselect string
var Flist bool

func init() {
	const (
		defaultHost     = ""
		defaultPassword = ""
		defaultPort     = 0
		defaultRaw      = false
		defaultUser     = ""
		defaultConfig   = ""
		defaultVerbose  = false
		defaultDebug    = ""
		defaultLog      = ""
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-h] [-D list] [-H host] [-l logfile] [-P password] [-p port] [-S profile] [-u user] [-list-profiles] [-preferences configfile] mapfile ...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  An option 'x' with a value may be set by '-x value', '-x=value', '--x value', or '--x=value'.\n")
		fmt.Fprintf(os.Stderr, "  A flag 'x' may be set by '-x', '--x', '-x=true|false' or '--x=true|false'\n")
		fmt.Fprintf(os.Stderr, "  Options may NOT be combined into a single argument (use '-h -m', not '-hm').\n")
		fmt.Fprintf(os.Stderr, "\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&Flist, "list-profiles", false, "list all defined profile names and exit")

	flag.StringVar(&Fselect, "select", "", "profile to select from mapper preferences")
	flag.StringVar(&Fselect, "S", "", "(same as -select)")

	flag.StringVar(&Fhost, "host", defaultHost, "hostname of mapper service")
	flag.StringVar(&Fhost, "H", defaultHost, "(same as -host)")

	flag.UintVar(&Fport, "port", defaultPort, "TCP port of mapper service (default 2323)")
	flag.UintVar(&Fport, "p", defaultPort, "(same as -port)")

	flag.StringVar(&Fpass, "password", defaultPassword, "Server password (if required)")
	flag.StringVar(&Fpass, "P", defaultPassword, "(same as -password)")

	flag.StringVar(&Fuser, "username", defaultUser, "Username on server or \"GM\" (default is local username)")
	flag.StringVar(&Fuser, "u", defaultUser, "(same as -username)")

	flag.StringVar(&Fconf, "preferences", defaultConfig, "Configuration file")

	flag.StringVar(&Fdebug, "debug", defaultDebug, "Comma-separated list of debugging topics to print")
	flag.StringVar(&Fdebug, "D", defaultDebug, "(same as -debug)")

	flag.StringVar(&Flog, "log", defaultLog, "Logfile ('-' is standard output)")
	flag.StringVar(&Flog, "l", defaultLog, "(same as -log)")
}

func main() {
	log.SetPrefix("push-images: ")

	prefs, err := configureApp()
	if err != nil {
		log.Fatalf("unable to set up: %v", err)
	}

	if Flist {
		fmt.Printf("Profiles defined (for use with -select options)\n")
		for _, prof := range prefs.Prefs.Profiles {
			if prof.Host == "" {
				fmt.Printf("  %s  (no host)\n", prof.Name)
			} else {
				fmt.Printf("  %s  (%s)\n", prof.Name, prof.Host)
			}
		}
		os.Exit(0)
	}
	if prefs.LogFile != "" && prefs.LogFile != "-" {
		f, err := os.OpenFile(prefs.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("unable to open log file \"%s\": %v", prefs.LogFile, err)
		}
		log.SetOutput(f)
	}

	if prefs.Prefs.Profiles[prefs.SelectedIdx].Host == "" {
		log.Fatalf("-host is required")
	}
	if prefs.Prefs.Profiles[prefs.SelectedIdx].Port <= 0 {
		log.Fatalf("-port is required")
	}
	user := prefs.Prefs.Profiles[prefs.SelectedIdx].UserName
	pass := prefs.Prefs.Profiles[prefs.SelectedIdx].Password

	problems := make(chan mapper.MessagePayload, 10)
	messages := make(chan mapper.MessagePayload, 10)
	done := make(chan int, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	debugFlagList := prefs.DebugFlags
	debugFlags, err := mapper.NamedDebugFlags(debugFlagList)
	if err != nil {
		log.Fatalf("-debug: %v", err)
	}

	var images []mapper.ImageDefinition

	for _, mapfile := range flag.Args() {
		elements, metaData, err := mapper.ReadMapFile(mapfile)
		if err != nil {
			log.Fatalf("unable to read %s: %v", mapfile, err)
		}
		log.Printf("Pushing images from %s (%s, %s, %s)...", mapfile, metaData.DateTime, metaData.Comment, metaData.Location)
		for eNumber, element := range elements {
			if img, isImageDef := element.(mapper.ImageDefinition); isImageDef {
				log.Printf("%d Image definition %s\n", eNumber, img.Name)
				images = append(images, img)
			}
		}
	}
	log.Printf("Total of %d images to send.\n", len(images))

	conOpts := []mapper.ConnectionOption{
		mapper.WithContext(ctx),
		mapper.WithSubscription(problems, mapper.ERROR, mapper.UNKNOWN),
		//		mapper.WithSubscription(messages,
		//			mapper.Marco,
		//		),
		mapper.WithDebugging(debugFlags),
		mapper.WithLogger(log.Default()),
	}

	if pass != "" {
		a := auth.NewClientAuthenticator(user, []byte(pass),
			fmt.Sprintf("push-images %s", GoVersionNumber))
		conOpts = append(conOpts, mapper.WithAuthenticator(a))
	}
	server, conerr := mapper.NewConnection(fmt.Sprintf("%s:%d",
		prefs.Prefs.Profiles[prefs.SelectedIdx].Host,
		prefs.Prefs.Profiles[prefs.SelectedIdx].Port),
		conOpts...)
	if conerr != nil {
		log.Fatalf("unable to contact mapper server: %v", conerr)
	}
	go func(done chan int) {
		server.Dial()
		done <- 1
	}(done)

	waitCounter := 0
	for !server.IsReady() {
		if waitCounter++; waitCounter > 30 {
			fmt.Println("Waiting for server to be ready...")
			waitCounter = 0
		}
		time.Sleep(100 * time.Millisecond)
	}
	if server.ServerStats.ServerVersion == "" {
		fmt.Println("Connected to server.")
	} else {
		fmt.Printf("Connected to server version %s.\n", server.ServerStats.ServerVersion)
	}

	if server.ServerStats.Started.IsZero() {
		fmt.Println("Server did not send uptime data.")
	} else if server.ServerStats.ConnectTime.IsZero() {
		fmt.Println("Server did not send it's local time data.")
	} else {
		fmt.Printf("Server up since %v (%v ago)\n",
			server.ServerStats.Started, server.ServerStats.ConnectTime.Sub(server.ServerStats.Started))
		if server.ServerStats.Active.IsZero() {
			fmt.Println("Server did not send activity timing data.")
		} else {
			activeSince := server.ServerStats.ConnectTime.Sub(server.ServerStats.Active)
			if activeSince >= time.Minute*5 {
				fmt.Printf("Server may be deadlocked! Last ping event was %s ago.\n", activeSince)
			} else {
				fmt.Printf("Server active; last ping event was %s ago.\n", activeSince)
			}
		}
	}

eventloop:
	for i, image := range images {
		log.Printf("Sending #%d: %s\n", i, image.Name)
		if err := server.AddImage(image); err != nil {
			log.Fatalf("Error sending image: %v", err)
		}
		select {
		case msg := <-problems:
			switch message := msg.(type) {
			case mapper.ErrorMessagePayload:
				fmt.Printf("ERROR: %v", message.Error)
			case mapper.UnknownMessagePayload:
				fmt.Printf("WARNING: Unknown type of message received from server: %q", msg.RawMessage())
			}
		case msg := <-messages:
			fmt.Printf("Unexpected incoming message: %v\n", msg.RawMessage())
		case <-done:
			log.Printf("Server connection ended.")
			break eventloop

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	log.Printf("Completed image pushes. Disconnecting in 5 seconds.\n")
	time.Sleep(5 * time.Second)
	server.Close()
}

type AppPreferences struct {
	Prefs       util.UserPreferences
	LogFile     string
	DebugFlags  string
	SelectedIdx int
}

func configureApp() (AppPreferences, error) {
	var defUserName string
	var prefs AppPreferences
	var err error

	defUser, err := user.Current()
	if err != nil {
		defUserName = "unknown"
	} else {
		defUserName = defUser.Username
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return prefs, fmt.Errorf("unable to determine user's home directory: %v", err)
	}

	//
	// command-line parameters
	//
	preferencesPath := filepath.Join(homeDir, ".gma", "mapper", "preferences.json")
	defConfigPath := filepath.Join(homeDir, ".gma", "mapper", "mapper.conf")
	flag.Parse()

	if prefsFile, err := os.Open(preferencesPath); err == nil {
		log.Println("Loading user preferences from", preferencesPath)
		prefs.Prefs, err = util.LoadPreferencesWithDefaults(prefsFile)
		if err != nil {
			var uv util.UnsupportedPreferencesVersionError
			if errors.As(err, &uv) {
				log.Printf("(warning) %v", err)
				fmt.Printf("WARNING: The preferences file is version %d, which I don't support.\nHowever, I'll try to continue anyway with what I can understand of that file.\n", uv.DataVersion)
			} else {
				return prefs, err
			}
		}
	} else if Fconf == "" {
		log.Println("No user preferences found; trying old-style config file")
		Fconf = defConfigPath
	}
	if Fselect == "" {
		// Pick the first profile if the user didn't specify one.
		if prefs.Prefs.CurrentProfile != "" {
			Fselect = prefs.Prefs.CurrentProfile
		} else if len(prefs.Prefs.Profiles) > 0 {
			Fselect = prefs.Prefs.Profiles[0].Name
		} else {
			log.Printf("preferences data contain no server profiles")
		}
		if Fselect != "" {
			log.Printf("defaulting to profile \"%s\"\n", Fselect)
		}
	}

	//
	// read in configuration
	//
	if Fconf != "" {
		configFile, err := os.Open(Fconf)
		if err != nil {
			if Fconf == defConfigPath && errors.Is(err, fs.ErrNotExist) {
				log.Printf("warning: configuration file \"%s\" does not exist", Fconf)
			} else {
				return prefs, fmt.Errorf("%s: %v", Fconf, err)
			}
		} else {
			defer configFile.Close()
			conf, err := util.ParseSimpleConfig(configFile)
			if err != nil {
				return prefs, err
			}
			if err = prefs.Prefs.UpdateFromSimpleConfig(Fselect, conf); err != nil {
				return prefs, err
			}
		}
	}

	// Override configuration file settings from command-line
	// options
	if Fselect != "" {
		for idx, pro := range prefs.Prefs.Profiles {
			if pro.Name == Fselect {
				prefs.SelectedIdx = idx
				break
			}
		}
		log.Printf("using profile #%d, \"%s\"\n", prefs.SelectedIdx, prefs.Prefs.Profiles[prefs.SelectedIdx].Name)
	} else {
		prefs.Prefs.Profiles = make([]util.ServerProfile, 1)
		prefs.SelectedIdx = 0
	}

	if Fhost != "" {
		prefs.Prefs.Profiles[prefs.SelectedIdx].Host = Fhost
	}
	if Fport != 0 {
		prefs.Prefs.Profiles[prefs.SelectedIdx].Port = int(Fport)
	}
	if Fpass != "" {
		prefs.Prefs.Profiles[prefs.SelectedIdx].Password = Fpass
	}
	if Fuser != "" {
		prefs.Prefs.Profiles[prefs.SelectedIdx].UserName = Fuser
	}
	if Flog != "" {
		prefs.LogFile = Flog
	}
	if Fdebug != "" {
		prefs.DebugFlags = Fdebug
	}

	// Sanity check and defaults
	if prefs.Prefs.Profiles[prefs.SelectedIdx].UserName == "" {
		prefs.Prefs.Profiles[prefs.SelectedIdx].UserName = defUserName
	}

	return prefs, nil
}

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

// readUserInput takes lines of input from the user which correspond to
// map server requests. While the Python version of the map-console
// just sent the input as-is, trusting the user to form a protocol-conforming
// string, we will parse out the user's request, make sure it is correct,
// and then make that request to the server.
//
// It would be nice to accept a higher-level abstraction of the commands
// but for now, like the earlier Python map-console, we expect the user
// to type the command in the form it would be sent to the server.
// See the mapper protocol as described in mapper(6) for details.
func readUserInput(mono bool, cancel context.CancelFunc, server mapper.Connection) {
	fmt.Printf("map-console> ")
	scanner := bufio.NewScanner(os.Stdin)
	simpleArg := regexp.MustCompile(`^(\w+)([#=:])(.*)$`)

inputloop:
	for scanner.Scan() {
		inputLine := scanner.Text()
		if len(inputLine) == 0 {
			continue
		}
		if inputLine[0] == '`' {
			if err := server.UNSAFEsendRaw(inputLine[1:]); err != nil {
				fmt.Println(colorize(fmt.Sprintf("ERROR sending raw string: %v", err), "Red", mono))
			}
			continue
		}

		fields, err := tcllist.ParseTclList(inputLine)
		if err != nil {
			fmt.Println(colorize(fmt.Sprintf("ERROR: Unrecognized input: %v", err), "Red", mono))
		} else if len(fields) > 0 {
			if len(fields[0]) > 1 && fields[0][0] == '!' {
				var paramList strings.Builder
				fmt.Fprintf(&paramList, "%s {", strings.ToUpper(fields[0][1:]))

				for i, arg := range fields[1:] {
					if i > 0 {
						fmt.Fprint(&paramList, ",")
					}

					f := simpleArg.FindStringSubmatch(arg)
					if f == nil || len(f) != 4 {
						fmt.Println(colorize(fmt.Sprintf("ERROR: parameter #%d can't be parsed", i), "Red", mono))
						continue inputloop
					}
					rawName, err := json.Marshal(f[1])
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("ERROR: parameter #%d, marshalling parameter name: %v", i, err), "Red", mono))
						continue inputloop
					}
					fmt.Fprint(&paramList, string(rawName), ":")
					switch f[2] {
					case "#":
						// name#value	place raw value in parameter list
						fmt.Fprint(&paramList, f[3])
					case "=":
						// name=value	place string value in parameter list
						rawVal, err := json.Marshal(f[3])
						if err != nil {
							fmt.Println(colorize(fmt.Sprintf("ERROR: parameter #%d, marshalling value: %v", i, err), "Red", mono))
							continue inputloop
						}
						fmt.Fprint(&paramList, string(rawVal))
					case ":":
						// name:value	place string value in parameter list with _ standing for space in value
						rawVal, err := json.Marshal(strings.ReplaceAll(f[3], "_", " "))
						if err != nil {
							fmt.Println(colorize(fmt.Sprintf("ERROR: parameter #%d, marshalling value: %v", i, err), "Red", mono))
							continue inputloop
						}
						fmt.Fprint(&paramList, string(rawVal))
					default:
						fmt.Println(colorize(fmt.Sprintf("ERROR: parameter #%d has invalid separator", i), "Red", mono))
						continue inputloop
					}
				}
				fmt.Fprint(&paramList, "}")
				fmt.Println(colorize(fmt.Sprintf("=> %s", paramList.String()), "Green", mono))
				if err := server.UNSAFEsendRaw(paramList.String()); err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR sending raw string: %v", err), "Red", mono))
				}
				continue
			}

		handle_input:
			switch strings.ToUpper(fields[0]) {
			case "HELP", "?":
				fmt.Println(`Command summary:
` + "`" + `<text>                                 Send <text> AS-IS to the server (must conform to protocol)
!<cmd> k1=v1 k2=v2 k3#v3 k4:v4 ...      Send <cmd> with parameters; = quotes value, # does not, : allows _ for space; <cmd> is uppercased
AI <name> <size> <filename>             Upload image from local file
AI? <name> <size>                       Ask for definition of image
AI@ <name> <size> <serverid>            Advertise image stored on server
AI/ <regex> [-keep]                     Filter out/keep some server-side image definitions
AV <grid> <xfrac> <yfrac>               Adjust view to grid label or fraction of each axis
CC <silent?> <target>                   Clear chat history
CLR <id>|*|E*|M*|P*|[<image>=]<name>    Clear specified element(s) from canvas
CLR@ <serverid>                         Remove all elements from server file
CO <bool>                               Enable/disable combat mode
CS <abs> <rel>                          Set game clock
D {<recip>|*|% ...} <dice>              Roll dice to users (*=all, %=GM)
DD {{<name> <desc> <dice>} ...}         Replace dice preset list
DD+ {{<name> <desc> <dice>} ...}        Add to dice preset list
DD/ <regex>                             Delete all presets whose names match RE
DR                                      Request die roll preset
L <filename>                            Tell clients to load local file to mapper
L@ <filename>                           Tell clients to load server file
LS <filename>                           Upload contents of local file to all
M <filename>                            Tell clients to merge local file to mapper
M? <serverid>                           Ensure local cache of server file
M@ <serverid>                           Tell clients to merge contents of server file to canvas
MARK <x> <y>                            Show visual marker at coordinates
OA <id> {<k0> <v0> ... <kN> <vN>}       Set object attribute(s)
OA+ <id> <key> {<v0> <v1> ... <vN>}     Add to list-type object attribute
OA- <id> <key> {<v0> <v1> ... <vN>}     Remove from list-type object attribute
POLO                                    Answer server ping request
PS <id> <color> <name> <area> <size> player|monster <x> <y> <reach>  
QUIT|EXIT                               Exit the client
SYNC [CHAT [<target>]]                  Sync server content / chat history
TO {<recip>|@|*|% ...} <message>        Send chat message
/CONN`)
			case "//":
				// ignore
			case "AC", "ACCEPT", "DENIED", "DSM", "GRANTED", "I", "IL", "MARCO", "OK", "PRIV", "ROLL", "CONN", "CONN:", "CONN.":
				// server messages
				fmt.Println(colorize(fmt.Sprintf("ERROR: %s is not for clients to send.", fields[0]), "Red", mono))

			case "AI":
				// AI name size file
				v, err := tcllist.ConvertTypes(fields, "ssfs")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				data, err := os.ReadFile(v[3].(string))
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("I/O ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AddImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64), ImageData: data},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AI:", "AI.", "AUTH", "DD=", "DD:", "DD.", "LS:", "LS.":
				// AI: AI. DD= DD: DD. LS: LS. internal protocol commands
				// AUTH response [user client]
				fmt.Println(colorize(fmt.Sprintf("ERROR: %s should not be typed directly here.", fields[0]), "Red", mono))

			case "AI?":
				// AI? name size
				v, err := tcllist.ConvertTypes(fields, "ssf")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.QueryImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64)},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AI@":
				// AI@ name size id
				v, err := tcllist.ConvertTypes(fields, "ssfs")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AddImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64), File: v[3].(string), IsLocalFile: false},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AI/":
				// AI/ regex [-keep]
				v, err := tcllist.ConvertTypes(fields, "sss")
				if err != nil {
					v, err = tcllist.ConvertTypes(fields, "ss*")
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
						break
					}
				}
				if len(v) == 3 {
					if v[2].(string) != "-keep" {
						fmt.Println(colorize("Argument #2 must be nothing or \"-keep\".", "Red", mono))
						break
					}
					if err := server.FilterImagesExcept(v[1].(string)); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.FilterImages(v[1].(string)); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "AV":
				// AV label x y
				v, err := tcllist.ConvertTypes(fields, "ssff")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AdjustViewToGridLabel(v[2].(float64), v[3].(float64), v[1].(string)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CC":
				// CC silent? target
				v, err := tcllist.ConvertTypes(fields, "s?i")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.ClearChat(v[2].(int), v[1].(bool)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CLR":
				// CLR id|*|E*|M*|P*|[imagename=]name
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: CLR <id>", "Red", mono))
					break
				}
				if err := server.Clear(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CLR@":
				// CLR@ id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: CLR@ <serverid>", "Red", mono))
					break
				}
				if err := server.ClearFrom(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CO":
				// CO state
				v, err := tcllist.ConvertTypes(fields, "s?")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.CombatMode(v[1].(bool)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CS":
				// CS abs rel
				fmt.Println(colorize("Sorry, CS is not yet implemented for the console.", "Red", mono))

			case "D":
				// D reciplist dice
				var requestID string

				if len(fields) == 4 {
					requestID = fields[3]
				} else if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: D <recip>|*|% <dice> [<id>]", "Red", mono))
					break
				}
				recips, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in recipient list: %v", err), "Red", mono))
					break
				}
				if len(recips) == 1 && recips[0] == "%" {
					if err := server.RollDiceToGMWithID(fields[2], requestID); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else if len(recips) == 1 && recips[0] == "*" {
					if err := server.RollDiceToAllWithID(fields[2], requestID); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.RollDiceWithID(recips, fields[2], requestID); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "DDD":
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DDD {[<name> ...]}", "Red", mono))
					break
				}
				p, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in delegate list: %v", err), "Red", mono))
					break
				}
				if err := server.DefineDicePresetDelegates(p); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "DD", "DD+":
				// DD {{name desc dice} ...}
				// DD+ {{name desc dice} ...}
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DD[+] {{<name> <desc> <dice>} ...}", "Red", mono))
					break
				}
				p, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in preset list: %v", err), "Red", mono))
					break
				}
				var presetList []dice.DieRollPreset
				for i, ps := range p {
					pl, err := tcllist.Parse(ps, "sss")
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("ERROR in preset list, #%d: %v", i+1, err), "Red", mono))
						break handle_input
					}
					presetList = append(presetList, dice.DieRollPreset{
						Name:        pl[0].(string),
						Description: pl[1].(string),
						DieRollSpec: pl[2].(string),
					})
				}
				if fields[0] == "DD" {
					if err := server.DefineDicePresets(presetList); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.AddDicePresets(presetList); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "DD/":
				// DD/ regex
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DD/ <regex>", "Red", mono))
					break
				}
				if err := server.FilterDicePresets(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "DR":
				// DR
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DR", "Red", mono))
					break
				}
				if err := server.QueryDicePresets(); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "L":
				// L filename
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: L <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], true, false); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "L@":
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: L@ <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], false, false); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "LS":
				// LS
				fmt.Println(colorize("LS not supported", "Red", mono))
				/*
					if len(fields) != 2 {
						fmt.Println(colorize("usage ERROR: wrong number of fields: LS <file>", "Red", mono))
						break
					}
					data, err := readLines(fields[1])
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("I/O ERROR: %v", err), "Red", mono))
						break
					}
					log.Printf("Reading objects from local mapper file %s", fields[1])
					objects, images, files, err := mapper.ParseObjects(data)
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("parser ERROR: %v", err), "Red", mono))
						break
					}
					log.Printf("Found %d object%s, %d image%s, and %d file%s in %s",
						len(objects), plural(len(objects)),
						len(images), plural(len(images)),
						len(files), plural(len(files)),
						fields[1],
					)
					if len(objects) > 0 {
						fmt.Print("Sending objects...")
						for i, o := range objects {
							if i%10 == 9 {
								fmt.Print(".")
							}
							if err := server.LoadObject(o); err != nil {
								fmt.Println(colorize(fmt.Sprintf("server ERROR sending object #%d (%s): %v",
									i+1, o.ObjID(), err), "Red", mono))
								break handle_input
							}
						}
						fmt.Println("done")
					}
					if len(images) > 0 {
						fmt.Print("Sending images...")
						for id, image := range images {
							if err := server.AddImage(image); err != nil {
								fmt.Println(colorize(fmt.Sprintf("server ERROR sending image %s: %v",
									id, err), "Red", mono))
								break handle_input
							}
						}
						fmt.Println("done")
					}
					if len(files) > 0 {
						fmt.Print("Sending files...")
						for i, f := range files {
							if i%10 == 9 {
								fmt.Print(".")
							}
							if f.IsLocalFile {
								log.Printf("%s include local file definition %s which doesn't make sense. Ignoring this.", fields[1], f.File)
							} else {
								if err := server.CacheFile(f.File); err != nil {
									fmt.Println(colorize(fmt.Sprintf("server ERROR sending file #%d: %v",
										i+1, err), "Red", mono))
									break handle_input
								}
							}
						}
						fmt.Println("done")
					}
				*/

			case "M":
				// M filenames
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], true, true); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "M?":
				// M" id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M: <serverID>", "Red", mono))
					break
				}
				if err := server.CacheFile(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "M@":
				// M@ id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M@ <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], false, true); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "MARK":
				// MARK x y
				v, err := tcllist.ConvertTypes(fields, "sff")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.Mark(v[1].(float64), v[2].(float64)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "OA":
				// OA id kvlist
				// It just happens to be a side effect of the protocol definition that we
				// can accept string values for the attributes here regardless of their
				// actual types, so we need not worry about converting them now.
				//
				// TODO: it might still be prudent to check anyway, but the point of the
				// console isn't to have all safety belts engaged for the user.

				if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: OA <id> {<k> <v> ...}", "Red", mono))
					break
				}
				alist, err := tcllist.ParseTclList(fields[2])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: can't parse kv list: %v", err), "Red", mono))
					break
				}
				if (len(alist) % 2) != 0 {
					fmt.Println(colorize("usage ERROR: kv list must have an even number of elements", "Red", mono))
					break
				}
				attrs := make(map[string]any)
				for i := 0; i < len(alist); i += 2 {
					attrs[alist[i]] = alist[i+1]
				}
				if err := server.UpdateObjAttributes(fields[1], attrs); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "OA+", "OA-":
				// OA+ id key vlist
				// OA- id key vlist
				if len(fields) != 4 {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: wrong number of fields: %s <id> <k> {<v> ...}", fields[0]), "Red", mono))
					break
				}
				vlist, err := tcllist.ParseTclList(fields[3])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: can't parse value list: %v", err), "Red", mono))
					break
				}
				if fields[0] == "OA+" {
					if err := server.AddObjAttributes(fields[1], fields[2], vlist); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.RemoveObjAttributes(fields[1], fields[2], vlist); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "POLO":
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: POLO", "Red", mono))
					break
				}
				server.Polo()

			case "PS":
				// PS id color name area size player|monster x y reach
				// 0  1  2     3    4    5    6              7 8 9
				v, err := tcllist.ConvertTypes(fields, "sssssssffi")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				// Size (now SkinSize) can now be a list of sizes
				ss, err := tcllist.ParseTclList(v[5].(string))
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v reading size list value \"%s\"", err, v[5].(string)), "Red", mono))
					break
				}

				c := mapper.CreatureToken{
					CreatureType: mapper.CreatureTypeUnknown,
				}
				c.ID = v[1].(string)
				c.Color = v[2].(string)
				c.Name = v[3].(string)
				c.SkinSize = ss
				c.Gx = v[7].(float64)
				c.Gy = v[8].(float64)
				c.Reach = v[9].(int)
				if err := c.SetSizes(c.SkinSize, 0, ""); err != nil {
					fmt.Println(colorize(fmt.Sprintf("size code error: %v", err), "Red", mono))
					break handle_input
				}

				switch v[6].(string) {
				case "player":
					c.CreatureType = mapper.CreatureTypePlayer
					if err := server.PlaceSomeone(mapper.PlayerToken{
						CreatureToken: c,
					}); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case "monster":
					c.CreatureType = mapper.CreatureTypeMonster
					if err := server.PlaceSomeone(mapper.MonsterToken{
						CreatureToken: c,
					}); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				default:
					fmt.Println(colorize("usage ERROR: creature type must be \"monster\" or \"player\"", "Red", mono))
				}

			case "SYNC":
				// SYNC [CHAT [target]]
				switch len(fields) {
				case 1:
					if err := server.Sync(); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case 2:
					if err := server.SyncChat(0); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case 3:
					v, err := tcllist.ConvertTypes(fields, "ssi")
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
						break handle_input
					}
					if err := server.SyncChat(v[2].(int)); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				default:
					fmt.Println(colorize("usage ERROR: SYNC [CHAT [target]]", "Red", mono))
				}

			case "TO":
				// TO recips message
				if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: TO <recip>|*|% <message>", "Red", mono))
					break
				}
				recips, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in recipient list: %v", err), "Red", mono))
					break
				}
				if err := server.ChatMessage(recips, fields[2]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "/CONN":
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: /CONN", "Red", mono))
					break
				}
				server.QueryPeers()

			case "EXIT", "QUIT":
				// stop this client
				break inputloop

			default:
				fmt.Println(colorize(fmt.Sprintf("ERROR: Unrecognized command: %s", fields[0]), "Red", mono))
			}
		}
		fmt.Printf("map-console> ")
	}
	fmt.Println(colorize("Shutting down...", "Yellow", mono))
	cancel()
}

func colorize(text, color string, mono bool) string {
	if mono {
		return text
	}
	var prefix string
	switch color {
	case "blue":
		prefix = "34"
	case "Blue":
		prefix = "1;34"
	case "cyan":
		prefix = "36"
	case "Cyan":
		prefix = "1;36"
	case "green":
		prefix = "32"
	case "Green":
		prefix = "1;32"
	case "magenta":
		prefix = "35"
	case "Magenta":
		prefix = "1;35"
	case "red":
		prefix = "31"
	case "Red":
		prefix = "1;31"
	case "yellow":
		prefix = "33"
	case "Yellow":
		prefix = "1;33"
	}
	return "\x1b[" + prefix + "m" + text + "\x1b[0m"
}

/*
# @[00]@| Go-GMA 5.32.1
# @[01]@|
# @[10]@| Overall GMA package Copyright © 1992–2025 by Steven L. Willoughby (AKA MadScienceZone)
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
