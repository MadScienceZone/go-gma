/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __    ______       __            #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \  / ___  \     /  \           #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) \/   \  \    \/) )          #
# | |      | || || || (___) | Assistant | (____        | |    ___) /      | |          #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | |   (___ (       | |          #
# | | \_  )| |   | || (   ) |                 ) )      | |       ) \      | |          #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_/\___/  / _  __) (_         #
# (_______)|/     \||/     \| Client    \______/ (_) \____/\______/ (_) \____/         #
#                                                                                      #
########################################################################################
*/

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/mapper"
	"github.com/MadScienceZone/go-gma/v5/util"
	"github.com/newrelic/go-agent/v3/newrelic"
	"golang.org/x/exp/slices"
)

type DebugFlags uint64

const (
	DebugAuth DebugFlags = 1 << iota
	DebugDB
	DebugEvents
	DebugIO
	DebugInit
	DebugMessages
	DebugMisc
	DebugState
	DebugAll DebugFlags = 0xffffffff
)

//
// DebugFlagNameSlice returns a list of debug option names
// from the given DebugFlags value.
//
func DebugFlagNameSlice(flags DebugFlags) []string {
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
		{bits: DebugAuth, name: "auth"},
		{bits: DebugDB, name: "db"},
		{bits: DebugEvents, name: "events"},
		{bits: DebugIO, name: "i/o"},
		{bits: DebugInit, name: "init"},
		{bits: DebugMessages, name: "messages"},
		{bits: DebugMisc, name: "misc"},
		{bits: DebugState, name: "state"},
	} {
		if (flags & f.bits) != 0 {
			list = append(list, f.name)
		}
	}
	return list
}

//
// DebugFlagNames returns a single string representation of
// the debugging flags (topics) stored in the DebugFlags
// value passed in.
//
func DebugFlagNames(flags DebugFlags) string {
	list := DebugFlagNameSlice(flags)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

//
// NamedDebugFlags takes a comma-separated list of
// debug flag (topic) names, or a list of individual
// names, or both, and returns the DebugFlags
// value which includes all of them.
//
// If "none" appears in the list, it cancels all previous
// values seen, but subsequent names will add their values
// to the list.
//
func NamedDebugFlags(names ...string) (DebugFlags, error) {
	var d DebugFlags
	var err error
	for _, name := range names {
		for _, flag := range strings.Split(name, ",") {
			switch flag {
			case "none":
				d = 0
			case "all":
				d = DebugAll
			case "auth":
				d |= DebugAuth
			case "db":
				d |= DebugDB
			case "events":
				d |= DebugEvents
			case "I/O", "i/o", "io":
				d |= DebugIO
			case "init":
				d |= DebugInit
			case "messages":
				d |= DebugMessages
			case "misc":
				d |= DebugMisc
			case "state":
				d |= DebugState
			default:
				err = fmt.Errorf("No such -debug flag: \"%s\"", flag)
				// but keep processing the rest
			}
		}
	}
	return d, err
}

//
// Application holds the global settings and other context for the application
// generally.
//
type Application struct {
	// Logger is whatever device or file we're writing logs to.
	Logger *log.Logger

	NrLogFile *os.File
	NrAppName string
	NrApp     *newrelic.Application

	// If DeLugLevel is 0, no extra debugging output will be logged.
	// Otherwise, it gives a set of debugging topics to report.
	DebugLevel DebugFlags

	// Endpoint is the "[host]:port" string which specifies where our
	// incoming socket is listening.
	Endpoint string

	// If not empty, this gives the filename from which we are to read in
	// the initial client command set.
	InitFile string

	// Information given to each connecting client at the start of
	// their session
	clientPreamble struct {
		data   mapper.ClientPreamble
		reload chan byte
		fetch  chan *mapper.ClientPreamble
	}

	// If not empty, we require authentication, with passwords taken
	// from this file.
	PasswordFile string
	clientAuth   struct {
		groupPassword     []byte
		gmPassword        []byte
		personalPasswords map[string][]byte
		lock              sync.RWMutex
	}

	// Pathname for database file.
	DatabaseName string
	sqldb        *sql.DB

	clientData struct {
		add       chan *mapper.ClientConnection
		remove    chan *mapper.ClientConnection
		fetch     chan []*mapper.ClientConnection
		announcer chan byte
	}

	MessageIDGenerator chan int
	MessageIDReset     chan int

	// Current game state
	gameState struct {
		sync   chan *mapper.ClientConnection
		update chan *mapper.MessagePayload
	}

	// Last time we sent out a ping to all clients.
	// If this goes too long, it may indicate that the server
	// has become deadlocked.
	LastPing time.Time

	// Time the server was started.
	ServerStarted time.Time

	CPUProfileFile string

	// The AllowedClients list lets us require minimum versions of various clients.
	AllowedClients []mapper.PackageUpdate
}

func (a *Application) GetClientPreamble() *mapper.ClientPreamble {
	a.Debug(DebugInit, "fetching client preamble from generator channel")
	return <-a.clientPreamble.fetch
}

func (a *Application) GetAllowedClients() []mapper.PackageUpdate {
	return a.AllowedClients
}

//
// AddClient adds the given client connection to our list of active
// connections.
//
func (a *Application) AddClient(c *mapper.ClientConnection) {
	a.clientData.add <- c
	//a.SendPeerListToAll()
}

//
// DropAllClients severs the connection to all clients.
//
func (a *Application) DropAllClients() {
	clients := a.GetClients()
	for i, c := range clients {
		if c.Auth != nil {
			a.Logf("removing client #%d (%s,%s,%s)", i, c.Address, c.Auth.Username, c.Auth.Client)
		} else {
			a.Logf("removing client #%d (%s)", i, c.Address)
		}
		c.Conn.Close()
	}
}

//
// RemoveClients removes the given client from the list of connections.
//
func (a *Application) RemoveClient(c *mapper.ClientConnection) {
	a.clientData.remove <- c
	//a.SendPeerListToAll()
}

//
// GetClients returns a copy of the client list as it existed
// at the time of the call.
//
func (a *Application) GetClients() []*mapper.ClientConnection {
	return <-a.clientData.fetch
}

func (a *Application) announceClients() {
	for {
		<-a.clientData.announcer
		a.SendPeerListToAll()
	}
}

func (a *Application) manageClientList() {
	var clients []*mapper.ClientConnection

	a.Log("client list manager started")
	defer a.Log("client list manager stopped")

	newClientListCopy := func() []*mapper.ClientConnection {
		// make a new copy which a receiver can use with impunity in its own thread
		var cp []*mapper.ClientConnection
		for _, cc := range clients {
			cp = append(cp, cc)
		}
		return cp
	}

	refreshChannel := func() {
		// if we have a list already in the channel, consume it ourselves
		select {
		case <-a.clientData.fetch:
			a.Debug(DebugIO, "removed old client list from channel")
		default:
			a.Debug(DebugIO, "no old client list in channel")
		}
	}

	clientListCopy := newClientListCopy()

	for {
		select {
		case a.clientData.fetch <- clientListCopy:
			clientListCopy = newClientListCopy()
			a.Debug(DebugIO, "pushed client list to channel")

		case c := <-a.clientData.add:
			if c != nil {
				clients = append(clients, c)
				a.Debugf(DebugIO, "added new client to list: %v", c.IdTag())
			} else {
				a.Log("request to add nil client ignored")
			}
			clientListCopy = newClientListCopy()
			refreshChannel()
			a.clientData.announcer <- 0

		case c := <-a.clientData.remove:
			if c == nil {
				a.Log("request to remove nil client ignored")
				continue
			}
			a.Debugf(DebugIO, "removing client %v from list", c.IdTag())
			pos := slices.Index(clients, c)
			if pos < 0 {
				a.Logf("client %v not found in server's client list, so can't delete it more", c.IdTag())
				continue
			}
			clients[pos] = nil
			clients = slices.Delete(clients, pos, pos+1)
			clientListCopy = newClientListCopy()
			refreshChannel()
			a.clientData.announcer <- 0
		}
	}
}

//
// Debug logs messages conditionally based on the currently set
// debug level. It acts just like fmt.Println as far as formatting
// its arguments.
//
func (a *Application) Debug(level DebugFlags, message ...any) {
	if a != nil && a.Logger != nil && (a.DebugLevel&level) != 0 {
		var dmessage []any
		dmessage = append(dmessage, DebugFlagNames(level))
		dmessage = append(dmessage, message...)
		a.Logger.Println(dmessage...)
	}
}

//
// Log logs messages to the application's logger.
// It acts just like fmt.Println as far as formatting
// its arguments.
//
func (a *Application) Log(message ...any) {
	if a != nil && a.Logger != nil {
		a.Logger.Println(message...)
	}
}

//
// Logf logs messages to the application's logger.
// It acts just like fmt.Printf as far as formatting
// its arguments.
//
func (a *Application) Logf(format string, args ...any) {
	if a != nil && a.Logger != nil {
		a.Logger.Printf(format, args...)
	}
}

//
// Debugf works like Debug, but takes a format string and argument
// list just like fmt.Printf does.
//
func (a *Application) Debugf(level DebugFlags, format string, args ...any) {
	if a != nil && a.Logger != nil && (a.DebugLevel&level) != 0 {
		a.Logger.Printf(DebugFlagNames(level)+" "+format, args...)
	}
}

//
// GetAppOptions configures the application by reading command-line options.
//
func (a *Application) GetAppOptions() error {

	var initFile = flag.String("init-file", "", "Load initial client commands from named file path")
	var logFile = flag.String("log-file", "-", "Write log to given pathname (stderr if '-'); special % tokens allowed in path")
	var passFile = flag.String("password-file", "", "Require authentication with named password file")
	var endPoint = flag.String("endpoint", ":2323", "Incoming connection endpoint ([host]:port)")
	//	var saveInterval = flag.String("save-interval", "10m", "Save internal state this often")
	var sqlDbName = flag.String("sqlite", "", "Specify filename for sqlite database to use")
	var debugFlags = flag.String("debug", "", "List the debugging trace types to enable")
	var nrLogger = flag.String("telemetry-log", "", "Debugging log for telemetry collection")
	var nrAppName = flag.String("telemetry-name", "", "Application name for telemetry collection (default: \"gma-server\")")
	var profFile = flag.String("cpuprofile", "", "CPU Profiling output file (default: no profiling)")
	flag.Parse()

	if *debugFlags != "" {
		a.DebugLevel, _ = NamedDebugFlags(*debugFlags)
		a.Debugf(DebugInit, "debugging flags set to %#v%s", a.DebugLevel, DebugFlagNames(a.DebugLevel))
	}

	if *logFile == "" {
		a.Logger = nil
	} else {
		a.Logger = log.Default()
		if *logFile != "-" {
			path, err := util.FancyFileName(*logFile, nil)
			if err != nil {
				return fmt.Errorf("unable to understand log file path \"%s\": %v", *logFile, err)
			}
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("unable to open log file: %v", err)
			} else {
				a.Logger.SetOutput(f)
			}
			a.Debugf(DebugInit, "Logging to %v", path)
		}
	}

	if *profFile != "" {
		a.CPUProfileFile = *profFile
	}

	if *nrLogger == "-" {
		a.NrLogFile = os.Stdout
	} else if *nrLogger != "" {
		var err error
		path, err := util.FancyFileName(*nrLogger, nil)
		if err != nil {
			return fmt.Errorf("unable to understand telemetry log file path \"%s\": %v", *nrLogger, err)
		}
		a.NrLogFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to open telemetry log file: %v", err)
		}
		if !InstrumentCode {
			a.Log("WARNING: -telemetry-log option given but the server is not compiled to enable telemetry!")
		}
	}

	if *nrAppName == "" {
		a.NrAppName = "gma-server"
	} else {
		a.NrAppName = *nrAppName
		if !InstrumentCode {
			a.Log("WARNING: -telemetry-name option given but the server is not compiled to enable telemetry!")
		}
	}

	if *initFile != "" {
		a.InitFile = *initFile
		a.Logf("reading client initial command set from \"%s\"", a.InitFile)
	}

	if *passFile != "" {
		a.PasswordFile = *passFile
		a.Logf("authentication enabled via \"%s\"", a.PasswordFile)
		if err := a.refreshAuthenticator(); err != nil {
			a.Logf("unable to set up authentication: %v", err)
			return err
		}
	} else {
		a.Log("WARNING: authentication not enabled!")
	}

	if *endPoint != "" {
		a.Endpoint = *endPoint
		a.Logf("configured to listen on \"%s\"", a.Endpoint)
	} else {
		return fmt.Errorf("non-empty tcp [host]:port value required")
	}

	/*
		if *saveInterval == "" {
			a.SaveInterval = 10 * time.Minute
			a.Logf("defaulting state save interval to 10 minutes")
		} else {
			d, err := time.ParseDuration(*saveInterval)
			if err != nil {
				return fmt.Errorf("invalid save-time interval: %v", err)
			}
			a.SaveInterval = d
			a.Logf("saving state to disk every %v", a.SaveInterval)
		}
	*/

	if *sqlDbName == "" {
		return fmt.Errorf("database name is required")
	}
	a.DatabaseName = *sqlDbName
	a.Logf("using database \"%s\" to store internal state", a.DatabaseName)

	return nil
}

func (a *Application) GetPersonalCredentials(user string) []byte {
	a.Debug(DebugAuth, "acquiring a read lock on the password data")
	a.clientAuth.lock.RLock()
	defer func() {
		a.Debug(DebugAuth, "releasing read lock on password data")
		a.clientAuth.lock.RUnlock()
	}()
	a.Debug(DebugAuth, "acquired read lock; proceeding")
	secret, ok := a.clientAuth.personalPasswords[user]
	if !ok {
		return nil
	}
	return secret
}

func (a *Application) newClientAuthenticator(user string) (*auth.Authenticator, error) {
	if a.PasswordFile == "" {
		return nil, nil
	}

	a.Debug(DebugAuth, "acquiring a read lock on the password data")
	a.clientAuth.lock.RLock()
	defer func() {
		a.Debug(DebugAuth, "releasing read lock on password data")
		a.clientAuth.lock.RUnlock()
	}()
	a.Debug(DebugAuth, "acquired read lock; proceeding")

	cauth := &auth.Authenticator{
		Secret:   a.clientAuth.groupPassword,
		GmSecret: a.clientAuth.gmPassword,
	}

	if user != "" {
		personalPass, ok := a.clientAuth.personalPasswords[user]
		if ok {
			cauth.SetSecret(personalPass)
			a.Debugf(DebugAuth, "using personal password for %s", user)
		} else {
			a.Debugf(DebugAuth, "no personal password found for %s, using group password", user)
		}
	}

	return cauth, nil
}

func (a *Application) refreshAuthenticator() error {
	if a.PasswordFile == "" {
		return nil
	}

	a.Debug(DebugInit, "acquiring a write lock on the password data")
	a.clientAuth.lock.Lock()
	defer func() {
		a.Debug(DebugInit, "releasing write lock on password data")
		a.clientAuth.lock.Unlock()
	}()
	a.Debug(DebugInit, "acquired write lock; proceeding")

	fp, err := os.Open(a.PasswordFile)
	if err != nil {
		a.Logf("unable to open password file \"%s\": %v", a.PasswordFile, err)
		return err
	}
	defer func() {
		if err := fp.Close(); err != nil {
			a.Logf("error closing %s: %v", a.PasswordFile, err)
		}
	}()

	a.clientAuth.groupPassword = []byte{}
	a.clientAuth.gmPassword = []byte{}
	a.clientAuth.personalPasswords = make(map[string][]byte)

	scanner := bufio.NewScanner(fp)
	if scanner.Scan() {
		// first line is the group password
		a.clientAuth.groupPassword = scanner.Bytes()
		a.Debug(DebugInit, "set group password")
		if scanner.Scan() {
			// next line, if any, is the gm-specific password
			a.clientAuth.gmPassword = scanner.Bytes()
			a.Debug(DebugInit, "set GM password")

			// following lines are <user>:<password> for individual passwords
			line := 3
			for scanner.Scan() {
				pp := strings.SplitN(scanner.Text(), ":", 2)
				if len(pp) != 2 {
					a.Logf("WARNING: %s, line %d: ignoring personal password: missing delimiter", a.PasswordFile, line)
				} else {
					a.clientAuth.personalPasswords[pp[0]] = []byte(pp[1])
					a.Debugf(DebugInit, "set personal password for %s", pp[0])
				}
				line++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		a.Logf("error reading %s: %v", a.PasswordFile, err)
		return err
	}

	return nil
}

func (a *Application) HandleServerMessage(payload mapper.MessagePayload, requester *mapper.ClientConnection) {
	a.Debugf(DebugMessages, "HandleServerMessage received %T %v", payload, payload)
	switch p := payload.(type) {
	case mapper.AddImageMessagePayload:
		for _, instance := range p.Sizes {
			if instance.ImageData != nil && len(instance.ImageData) > 0 {
				a.Logf("not storing image \"%s\"@%v (inline image data not supported)", p.Name, instance.Zoom)
				continue
			}
			if err := a.StoreImageData(p.Name, mapper.ImageInstance{
				Zoom:        instance.Zoom,
				IsLocalFile: instance.IsLocalFile,
				File:        instance.File,
			}, p.Animation); err != nil {
				a.Logf("error storing image data for \"%s\"@%v: %v", p.Name, instance.Zoom, err)
			}
		}
		if err := a.SendToAllExcept(requester, mapper.AddImage, p); err != nil {
			a.Logf("error sending on AddImage to peer systems: %v", err)
		}

	case mapper.QueryImageMessagePayload:
		imgData, err := a.QueryImageData(mapper.ImageDefinition{Name: p.Name})
		if err != nil {
			a.Logf("unable to answer QueryImage (%v)", err)
			if err := a.SendToAllExcept(requester, mapper.QueryImage, p); err != nil {
				a.Logf("error sending QueryImage on to peers, as well: %v", err)
			}
			return
		}

		// Now that we have all the answers we know about, figure out
		// which we can answer directly and which ones we'll need to
		// call in help for.
		var answers mapper.AddImageMessagePayload
		var questions mapper.QueryImageMessagePayload

		answers.Name = p.Name
		questions.Name = p.Name
		answers.Animation = imgData.Animation
		for _, askedFor := range p.Sizes {
			// do we know the answer to this one?
			askOthers := true
			for _, found := range imgData.Sizes {
				if found.Zoom == askedFor.Zoom {
					// yes!
					answers.Sizes = append(answers.Sizes, mapper.ImageInstance{
						Zoom:        found.Zoom,
						IsLocalFile: found.IsLocalFile,
						File:        found.File,
					})
					askOthers = false
					break
				}
			}
			if askOthers {
				// we didn't find it in the database, ask if anyone else knows...
				questions.Sizes = append(questions.Sizes, mapper.ImageInstance{
					Zoom: askedFor.Zoom,
				})
			}
		}

		if len(answers.Sizes) > 0 {
			if err := requester.Conn.Send(mapper.AddImage, answers); err != nil {
				a.Logf("error sending QueryImage answer to requester: %v", err)
			}
		}

		if len(questions.Sizes) > 0 {
			if err := a.SendToAllExcept(requester, mapper.QueryImage, questions); err != nil {
				a.Logf("error asking QueryImage query out to other peers: %v", err)
			}
		}

	case mapper.ClearChatMessagePayload:
		if requester != nil && requester.Auth != nil {
			p.RequestedBy = requester.Auth.Username
		}
		p.MessageID = <-a.MessageIDGenerator
		a.SendToAllExcept(requester, mapper.ClearChat, p)
		if err := a.ClearChatHistory(p.Target); err != nil {
			a.Logf("error clearing chat history (target=%d): %v", p.Target, err)
		}
		if err := a.AddToChatHistory(p.MessageID, mapper.ClearChat, p); err != nil {
			a.Logf("unable to add ClearChat event to chat history: %v", err)
		}

	case mapper.RollDiceMessagePayload:
		if requester.Auth == nil {
			a.Logf("refusing to accept die roll from unauthenticated user")
			requester.Conn.Send(mapper.RollResult, mapper.RollResultMessagePayload{
				ChatCommon: mapper.ChatCommon{
					MessageID: <-a.MessageIDGenerator,
				},
				RequestID: p.RequestID,
				Result: dice.StructuredResult{
					InvalidRequest: true,
					Details: dice.StructuredDescriptionSet{
						{Type: "error", Value: "I can't accept your die roll request. I don't know who you even are."},
					},
				},
			})
			return
		}

		label, results, err := requester.D.DoRoll(p.RollSpec)
		if err != nil {
			// Bad request. Notify the requester
			requester.Conn.Send(mapper.RollResult, mapper.RollResultMessagePayload{
				ChatCommon: mapper.ChatCommon{
					MessageID:  <-a.MessageIDGenerator,
					Recipients: p.Recipients,
					ToAll:      p.ToAll,
					ToGM:       p.ToGM,
					Sender:     requester.Auth.Username,
				},
				RequestID: p.RequestID,
				Result: dice.StructuredResult{
					InvalidRequest: true,
					Details: dice.StructuredDescriptionSet{
						{Type: "error", Value: fmt.Sprintf("Unable to understand your die-roll request: %v", err)},
					},
				},
			})
			return
		}
		var genericParts []string
		for _, part := range strings.Split(label, "‖") {
			if pos := strings.IndexRune(part, '≡'); pos >= 0 {
				genericParts = append(genericParts, part[:pos])
			} else {
				genericParts = append(genericParts, part)
			}
		}
		genericLabel := strings.Join(genericParts, ", ")

		response := mapper.RollResultMessagePayload{
			ChatCommon: mapper.ChatCommon{
				Sender:     requester.Auth.Username,
				Recipients: p.Recipients,
				ToAll:      p.ToAll,
				ToGM:       p.ToGM,
			},
			Title:     label,
			RequestID: p.RequestID,
		}

		if p.ToGM {
			receiptMessageID := <-a.MessageIDGenerator
			receiptGenericLabel := genericLabel
			var receiptLabel string
			var receiptResult dice.StructuredResult
			var err error
			if requester.Auth.GmMode {
				receiptGenericLabel = ""
				receiptResult = dice.StructuredResult{
					ResultSuppressed: true,
					Details: dice.StructuredDescriptionSet{
						{Type: "notice", Value: "rolls behind screen"},
					},
				}
			} else {
				receiptLabel, receiptResult, err = requester.D.ExplainSecretRoll(p.RollSpec, "roll to GM")
				if err != nil {
					receiptResult = dice.StructuredResult{
						ResultSuppressed: true,
						Details: dice.StructuredDescriptionSet{
							{Type: "notice", Value: "roll to GM"},
							{Type: "error", Value: fmt.Sprintf("error preparing receipt message: %v", err)},
						},
					}
				}
			}

			receiptPayload := mapper.RollResultMessagePayload{
				ChatCommon: mapper.ChatCommon{
					MessageID: receiptMessageID,
					Sender:    requester.Auth.Username,
					ToAll:     true,
				},
				RequestID: p.RequestID,
				Title:     receiptLabel,
				Result:    receiptResult,
			}
			if err := a.AddToChatHistory(receiptMessageID, mapper.RollResult, receiptPayload); err != nil {
				a.Logf("unable to add RollResult receipt to chat history: %v", err)
			}
			for _, peer := range a.GetClients() {
				if peer.Auth == nil || !peer.Auth.GmMode {
					if !peer.Features.DiceColorBoxes {
						receiptPayload.Title = receiptGenericLabel
					} else {
						receiptPayload.Title = receiptLabel
					}

					peer.Conn.Send(mapper.RollResult, receiptPayload)
				}
			}
		}

		for seq, r := range results {
			response.MessageID = <-a.MessageIDGenerator
			response.Result = r
			response.MoreResults = seq+1 < len(results)

			if err := a.AddToChatHistory(response.MessageID, mapper.RollResult, response); err != nil {
				a.Logf("unable to add RollResult event to chat history: %v", err)
			}

			for _, peer := range a.GetClients() {
				if p.ToGM {
					if peer.Auth == nil || !peer.Auth.GmMode {
						// we already handled this case above
						continue
					}
				} else if !p.ToAll {
					if peer.Auth == nil || peer.Auth.Username == "" {
						a.Debugf(DebugIO, "sending to explicit list but we don't know who %v is (skipped)", peer.IdTag())
						continue
					}
					if peer.Auth.Username != requester.Auth.Username && slices.Index(p.Recipients, peer.Auth.Username) < 0 {
						a.Debugf(DebugIO, "sending to explicit list but user \"%s\" (from %v) isn't on the list (skipped)", peer.Auth.Username, peer.IdTag())
						continue
					}
				}

				if peer.Features.DiceColorBoxes {
					response.Title = label
				} else {
					response.Title = genericLabel
				}

				if err := peer.Conn.Send(mapper.RollResult, response); err != nil {
					a.Logf("error sending die-roll result %v to %v: %v", response, peer.IdTag(), err)
				}
			}
		}

	case mapper.SyncChatMessagePayload:
		if err := a.QueryChatHistory(p.Target, requester); err != nil {
			a.Logf("error syncing chat history (target=%d): %v", p.Target, err)
		}

	case mapper.DefineDicePresetsMessagePayload:
		if requester.Auth == nil {
			a.Logf("Unable to store die-roll preset for unauthenticated user")
			return
		}

		target := requester.Auth.Username
		if p.For != "" {
			if requester.Auth.GmMode {
				target = p.For
				a.Debugf(DebugIO, "GM requests storage of die-roll presets for %s", target)
			} else {
				a.Logf("non-GM request to change die-roll presets for %s ignored", p.For)
			}
		}

		if err := a.StoreDicePresets(target, p.Presets, true); err != nil {
			a.Logf("error storing die-roll preset: %v", err)
		}
		if err := a.SendDicePresets(target); err != nil {
			a.Logf("error sending die-roll presets after changing them: %v", err)
		}

	case mapper.AddDicePresetsMessagePayload:
		if requester.Auth == nil {
			a.Logf("Unable to store die-roll preset for unauthenticated user")
			return
		}

		target := requester.Auth.Username
		if p.For != "" {
			if requester.Auth.GmMode {
				target = p.For
				a.Debugf(DebugIO, "GM requests add to die-roll presets for %s", target)
			} else {
				a.Logf("non-GM request to add to die-roll presets for %s ignored", p.For)
			}
		}

		if err := a.StoreDicePresets(target, p.Presets, false); err != nil {
			a.Logf("error adding to die-roll preset: %v", err)
		}
		if err := a.SendDicePresets(target); err != nil {
			a.Logf("error sending die-roll presets after changing them: %v", err)
		}

	case mapper.EchoMessagePayload:
		if err := requester.Conn.SendEchoWithTimestamp(mapper.Echo, p); err != nil {
			a.Logf("Error sending ECHO: %v", err)
		}

	case mapper.FilterDicePresetsMessagePayload:
		if requester.Auth == nil {
			a.Logf("Unable to filter die-roll preset for unauthenticated user")
			return
		}

		target := requester.Auth.Username
		if p.For != "" {
			if requester.Auth.GmMode {
				target = p.For
				a.Debugf(DebugIO, "GM requests filter of die-roll presets for %s", target)
			} else {
				a.Logf("non-GM request to filter die-roll presets for %s ignored", p.For)
			}
		}

		if err := a.FilterDicePresets(target, p); err != nil {
			a.Logf("error filtering die-roll preset for %s with /%s/: %v", target, p.Filter, err)
		}
		if err := a.SendDicePresets(target); err != nil {
			a.Logf("error sending die-roll presets after filtering them: %v", err)
		}

	case mapper.FilterImagesMessagePayload:
		if requester.Auth == nil {
			a.Logf("Unable to filter images for unauthenticated user")
			return
		}

		if !requester.Auth.GmMode {
			a.Logf("Rejecting unauthorized AI/ command from user %s", requester.Auth.Username)
			return
		}

		if err := a.FilterImages(p); err != nil {
			a.Logf("error filtering images with /%s/: %v", p.Filter, err)
		}

	case mapper.QueryDicePresetsMessagePayload:
		if requester.Auth == nil {
			a.Logf("Unable to query die-roll preset for unauthenticated user")
			return
		}

		target := requester.Auth.Username
		if p.For != "" {
			if requester.Auth.GmMode {
				target = p.For
				a.Debugf(DebugIO, "GM requests die-roll presets for %s", target)
			} else {
				a.Logf("non-GM request to get die-roll presets for %s ignored", p.For)
			}
		}
		if err := a.SendDicePresets(target); err != nil {
			a.Logf("error sending die-roll presets: %v", err)
		}

	case mapper.ChatMessageMessagePayload:
		if requester.Auth == nil {
			a.Logf("refusing to pass on chat message from unauthenticated user")
			_ = requester.Conn.Send(mapper.ChatMessage, mapper.ChatMessageMessagePayload{
				ChatCommon: mapper.ChatCommon{
					MessageID: <-a.MessageIDGenerator,
				},
				Text: "I can't accept that chat message since I don't know who you even are.",
			})
			return
		}

		p.Sender = requester.Auth.Username
		p.MessageID = <-a.MessageIDGenerator

		if err := a.AddToChatHistory(p.MessageID, mapper.ChatMessage, p); err != nil {
			a.Logf("unable to add ChatMessage event to chat history: %v", err)
		}

		for _, peer := range a.GetClients() {
			if p.ToGM {
				if peer.Auth == nil || (!peer.Auth.GmMode && peer.Auth.Username != requester.Auth.Username) {
					a.Debugf(DebugIO, "sending to GM and %v isn't the GM (skipped)", peer.IdTag())
					continue
				}
			} else if !p.ToAll {
				if peer.Auth == nil || peer.Auth.Username == "" {
					a.Debugf(DebugIO, "sending to explicit list but we don't know who %v is (skipped)", peer.IdTag())
					continue
				}
				if peer.Auth.Username != requester.Auth.Username && slices.Index(p.Recipients, peer.Auth.Username) < 0 {
					a.Debugf(DebugIO, "sending to explicit list but user \"%s\" (from %v) isn't on the list (skipped)", peer.Auth.Username, peer.IdTag())
					continue
				}
			}

			if err := peer.Conn.Send(mapper.ChatMessage, p); err != nil {
				a.Logf("error sending message %v to %v: %v", p, peer.IdTag(), err)
			}
		}

	case mapper.QueryPeersMessagePayload:
		a.SendPeerListTo(requester)

	// These commands are passed on to our peers with no further action required.
	case mapper.MarkMessagePayload,
		mapper.UpdateProgressMessagePayload:
		a.SendToAllExcept(requester, payload.MessageType(), payload)

	// These commands are passed on to our peers and remembered for later sync operations.
	case mapper.AdjustViewMessagePayload, mapper.ClearMessagePayload, mapper.ClearFromMessagePayload,
		mapper.LoadFromMessagePayload,
		mapper.LoadArcObjectMessagePayload,
		mapper.LoadCircleObjectMessagePayload,
		mapper.LoadLineObjectMessagePayload,
		mapper.LoadPolygonObjectMessagePayload,
		mapper.LoadRectangleObjectMessagePayload,
		mapper.LoadSpellAreaOfEffectObjectMessagePayload,
		mapper.LoadTextObjectMessagePayload,
		mapper.LoadTileObjectMessagePayload,
		mapper.AddObjAttributesMessagePayload,
		mapper.RemoveObjAttributesMessagePayload,
		mapper.UpdateObjAttributesMessagePayload,
		mapper.PlaceSomeoneMessagePayload:
		a.SendToAllExcept(requester, payload.MessageType(), payload)
		a.UpdateGameState(&payload)

	// as above but they are privileged
	case mapper.CombatModeMessagePayload, mapper.UpdateStatusMarkerMessagePayload,
		mapper.UpdateTurnMessagePayload, mapper.UpdateInitiativeMessagePayload,
		mapper.UpdateClockMessagePayload, mapper.ToolbarMessagePayload:
		if requester == nil || requester.Auth == nil {
			a.Logf("refusing to execute privileged command %v for unauthenticated user", p.MessageType())
			requester.Conn.Send(mapper.Priv, mapper.PrivMessagePayload{
				Command: p.RawMessage(),
				Reason:  "You are not the GM. You might not even be real.",
			})
			return
		}
		if !requester.Auth.GmMode {
			a.Logf("refusing to execute privileged command %v %v for non-GM user %s", p.MessageType(), p, requester.Auth.Username)
			requester.Conn.Send(mapper.Priv, mapper.PrivMessagePayload{
				Command: p.RawMessage(),
				Reason:  "You are not the GM.",
			})
			return
		}
		a.SendToAllExcept(requester, payload.MessageType(), payload)
		a.UpdateGameState(&payload)

	case mapper.SyncMessagePayload:
		a.SendGameState(requester)

	default:
		a.Logf("received unexpected message (type %T, value %v); ignored", payload, payload)
	}
}

func (a *Application) SendPeerListToAll() {
	allClients := a.GetClients()
	var peers mapper.UpdatePeerListMessagePayload

	for _, peer := range allClients {
		thisPeer := mapper.Peer{
			Addr:     peer.Address,
			LastPolo: time.Since(peer.LastPoloTime).Seconds(),
			IsMe:     false,
		}
		if peer.Auth != nil {
			thisPeer.User = peer.Auth.Username
			thisPeer.Client = peer.Auth.Client
			thisPeer.IsAuthenticated = peer.Auth.Username != ""
		}
		peers.PeerList = append(peers.PeerList, thisPeer)
	}

	for i, peer := range allClients {
		peers.PeerList[i].IsMe = true
		if err := peer.Conn.Send(mapper.UpdatePeerList, peers); err != nil {
			a.Logf("error sending peer list to peer #%v (%v): %v", i, peer.IdTag(), err)
		}
		peers.PeerList[i].IsMe = false
	}
}

func (a *Application) SendPeerListTo(requester *mapper.ClientConnection) {
	var peers mapper.UpdatePeerListMessagePayload
	for _, peer := range a.GetClients() {
		thisPeer := mapper.Peer{
			Addr:     peer.Address,
			LastPolo: time.Since(peer.LastPoloTime).Seconds(),
			IsMe:     peer == requester,
		}
		if peer.Auth != nil {
			thisPeer.User = peer.Auth.Username
			thisPeer.Client = peer.Auth.Client
			thisPeer.IsAuthenticated = peer.Auth.Username != ""
		}
		peers.PeerList = append(peers.PeerList, thisPeer)
	}
	if err := requester.Conn.Send(mapper.UpdatePeerList, peers); err != nil {
		a.Logf("error sending peer list: %v", err)
	}
}

func (a *Application) SendToAllExcept(c *mapper.ClientConnection, cmd mapper.ServerMessage, data any) error {
	if c == nil {
		a.Debugf(DebugIO|DebugMessages, "sending %v %v to all clients", cmd, data)
	} else {
		a.Debugf(DebugIO|DebugMessages, "sending %v %v to all clients except %v", cmd, data, c.IdTag())
	}
	var reportedError error

	for _, peer := range a.GetClients() {
		a.Debugf(DebugIO, "peer %v", peer.IdTag())
		if c == nil || peer != c {
			a.Debugf(DebugIO, "-> %v %v %v", peer.IdTag(), cmd, data)
			if err := peer.Conn.Send(cmd, data); err != nil {
				a.Logf("error sending %v to client %v: %v", data, peer.IdTag(), err)
				reportedError = err
			}
		}
	}
	return reportedError
}

func (a *Application) SendToAll(cmd mapper.ServerMessage, data any) error {
	return a.SendToAllExcept(nil, cmd, data)
}

// NewApplication creates and initializes a new Application value.
func NewApplication() *Application {
	app := Application{
		Logger:             log.Default(),
		MessageIDGenerator: make(chan int),
		MessageIDReset:     make(chan int),
	}
	app.clientPreamble.reload = make(chan byte, 1)
	app.clientPreamble.fetch = make(chan *mapper.ClientPreamble, 1)
	app.gameState.sync = make(chan *mapper.ClientConnection, 1)
	app.gameState.update = make(chan *mapper.MessagePayload, 1)
	app.clientData.add = make(chan *mapper.ClientConnection, 1)
	app.clientData.remove = make(chan *mapper.ClientConnection, 1)
	app.clientData.fetch = make(chan []*mapper.ClientConnection, 1)
	app.clientData.announcer = make(chan byte)
	return &app
}

//
// managePreambleData centralizes access to the common preamble data
// in a single goroutine, providing goroutine-safe access to it via
// channels.
//
func (a *Application) managePreambleData() {
	a.Log("preamble data manager started")
	defer a.Log("preamble data manager stopped")

	commitInitCommand := func(cmd string, src strings.Builder, dst *[]string) error {
		var b []byte
		var err error

		s := []byte(src.String())

		switch cmd {
		case "AI":
			var data mapper.AddImageMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "AI?":
			var data mapper.QueryImageMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "AV":
			var data mapper.AdjustViewMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "CC":
			var data mapper.ClearChatMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "CLR":
			var data mapper.ClearMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "CLR@":
			var data mapper.ClearFromMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "CO":
			var data mapper.CombatModeMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "CS":
			var data mapper.UpdateClockMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "DD=":
			var data mapper.UpdateDicePresetsMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "DSM":
			var data mapper.UpdateStatusMarkerMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "I":
			var data mapper.UpdateTurnMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "IL":
			var data mapper.UpdateInitiativeMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "L":
			var data mapper.LoadFromMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-ARC":
			var data mapper.LoadArcObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-CIRC":
			var data mapper.LoadCircleObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-LINE":
			var data mapper.LoadLineObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-POLY":
			var data mapper.LoadPolygonObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-RECT":
			var data mapper.LoadRectangleObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-SAOE":
			var data mapper.LoadSpellAreaOfEffectObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-TEXT":
			var data mapper.LoadTextObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "LS-TILE":
			var data mapper.LoadTileObjectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "MARK":
			var data mapper.MarkMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "OA":
			var data mapper.UpdateObjAttributesMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "OA+":
			var data mapper.AddObjAttributesMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "OA-":
			var data mapper.RemoveObjAttributesMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "PROGRESS":
			var data mapper.UpdateProgressMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "AC", "PS":
			var data mapper.PlaceSomeoneMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "ROLL":
			var data mapper.RollResultMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "TB":
			var data mapper.ToolbarMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "TO":
			var data mapper.ChatMessageMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "UPDATES":
			var data mapper.UpdateVersionsMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				a.AllowedClients = data.Packages
				if a.AllowedClients != nil {
					for i, aClient := range a.AllowedClients {
						if aClient.VersionPattern != "" {
							a.AllowedClients[i].VersionRegex, err = regexp.Compile(aClient.VersionPattern)
							if err != nil {
								a.Debugf(DebugInit, "ERROR in %s VersionPattern \"%s\": %v; will not limit this client", aClient.Name, aClient.VersionPattern, err)
								a.AllowedClients[i].VersionRegex = nil
							}
							nSubs := len(a.AllowedClients[i].VersionRegex.SubexpNames())
							if nSubs != 2 {
								a.Debugf(DebugInit, "ERROR in %s VersionPattern \"%s\": must have exactly 1 capturing group; this expression has %d; will not limit this client", aClient.Name, aClient.VersionPattern, nSubs-1)
								a.AllowedClients[i].VersionRegex = nil
							}
						}
						if aClient.MinimumVersion != "" && aClient.VersionPattern == "" {
							a.Debugf(DebugInit, "in package %s, you can't have a minimum client version but not version pattern to match it with.", aClient.Name)
						}
					}
				}

				cpkg := make([]mapper.PackageUpdate, len(data.Packages))
				copy(cpkg, data.Packages)
				for i := range cpkg {
					// redact filters from update messages sent to clients
					cpkg[i].VersionPattern = ""
					cpkg[i].VersionRegex = nil
					cpkg[i].MinimumVersion = ""
				}
				b, err = json.Marshal(mapper.UpdateVersionsMessagePayload{
					Packages: cpkg,
				})
				a.Debugf(DebugInit, "allowed client list is now %v", a.AllowedClients)
			}

		case "REDIRECT":
			var data mapper.RedirectMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		case "WORLD":
			var data mapper.WorldMessagePayload
			if err = json.Unmarshal(s, &data); err == nil {
				b, err = json.Marshal(data)
			}

		default:
			return fmt.Errorf("invalid command %v in initialization file", cmd)
		}

		if err == nil {
			*dst = append(*dst, fmt.Sprintf("%s %s", cmd, string(b)))
		}
		return err
	}

	updateClientPreamble := func() {
		if a.InitFile == "" {
			return
		}

		f, err := os.Open(a.InitFile)
		if err != nil {
			a.Logf("error opening initial command file %v: %v", a.InitFile, err)
			return
		}
		defer f.Close()

		recordPattern := regexp.MustCompile("^(\\w+)\\s+({.*)")
		continuationPattern := regexp.MustCompile("^\\s+")
		endOfRecordPattern := regexp.MustCompile("^}")
		commandPattern := regexp.MustCompile("^(\\w+)\\s*$")

		a.clientPreamble.data.Preamble = nil
		a.clientPreamble.data.PostAuth = nil
		a.clientPreamble.data.PostReady = nil
		a.clientPreamble.data.SyncData = false
		currentPreamble := &a.clientPreamble.data.Preamble

		scanner := bufio.NewScanner(f)
	outerScan:
		for scanner.Scan() {
		rescan:
			if strings.TrimSpace(scanner.Text()) == "" {
				continue
			}
			if strings.HasPrefix(scanner.Text(), "//") {
				*currentPreamble = append(*currentPreamble, scanner.Text())
				continue
			}
			if f := commandPattern.FindStringSubmatch(scanner.Text()); f != nil {
				// dataless command f[1]
				switch f[1] {
				case "AUTH":
					currentPreamble = &a.clientPreamble.data.PostAuth
				case "READY":
					currentPreamble = &a.clientPreamble.data.PostReady
				case "SYNC":
					a.clientPreamble.data.SyncData = true
				default:
					a.Logf("invalid command \"%v\" in init file %s", scanner.Text(), a.InitFile)
					return
				}
			} else if f := recordPattern.FindStringSubmatch(scanner.Text()); f != nil {
				// start of record type f[1] with start of JSON string f[2]
				// collect rest of string
				var dataPacket strings.Builder
				dataPacket.WriteString(f[2])

				for scanner.Scan() {
					if continuationPattern.MatchString(scanner.Text()) {
						dataPacket.WriteString(scanner.Text())
					} else {
						if endOfRecordPattern.MatchString(scanner.Text()) {
							dataPacket.WriteString(scanner.Text())
						}
						if err := commitInitCommand(f[1], dataPacket, currentPreamble); err != nil {
							a.Logf("error in initial command file: %v", err)
							return
						}
						if !endOfRecordPattern.MatchString(scanner.Text()) {
							// We already read into next record
							goto rescan
						} else {
							continue outerScan
						}
					}
				}
				// We reached EOF while scanning with a command in progress
				if err := commitInitCommand(f[1], dataPacket, currentPreamble); err != nil {
					a.Logf("error in initial command file: %v", err)
					return
				}
				break
			}
		}

		if err := scanner.Err(); err != nil {
			a.Logf("error in initial command file: %v", err)
			return
		}

		if (a.DebugLevel & DebugInit) != 0 {
			a.Debugf(DebugInit, "client initial commands from %v", a.InitFile)
			a.Debugf(DebugInit, "client sync: %v", a.clientPreamble.data.SyncData)

			for i, p := range a.clientPreamble.data.Preamble {
				a.Debugf(DebugInit, "client preamble #%d: %s", i, p)
			}
			for i, p := range a.clientPreamble.data.PostAuth {
				a.Debugf(DebugInit, "client post-auth #%d: %s", i, p)
			}
			for i, p := range a.clientPreamble.data.PostReady {
				a.Debugf(DebugInit, "client post-ready #%d: %s", i, p)
			}
		}
	}

	copyCurrentPreambleData := func() *mapper.ClientPreamble {
		pp := make([]string, len(a.clientPreamble.data.Preamble))
		pa := make([]string, len(a.clientPreamble.data.PostAuth))
		pr := make([]string, len(a.clientPreamble.data.PostReady))
		copy(pp, a.clientPreamble.data.Preamble)
		copy(pa, a.clientPreamble.data.PostAuth)
		copy(pr, a.clientPreamble.data.PostReady)

		return &mapper.ClientPreamble{
			SyncData:  a.clientPreamble.data.SyncData,
			Preamble:  pp,
			PostAuth:  pa,
			PostReady: pr,
		}
	}

	updateClientPreamble()
	nextValue := copyCurrentPreambleData()
	a.Debugf(DebugInit, "staged preamble data %p, pre=%p, pa=%p, pr=%p", nextValue, &nextValue.Preamble, &nextValue.PostAuth, &nextValue.PostReady)

	for {
		select {
		case <-a.clientPreamble.reload:
			a.Debug(DebugInit, "reloading client preamble data from file")
			updateClientPreamble()

			select {
			case <-a.clientPreamble.fetch:
				a.Debug(DebugInit, "removed stale preamble data from channel")
			default:
				a.Debug(DebugInit, "no stale preamble data in channel to remove")
			}

			a.Debug(DebugInit, "staging fresh preamble copy into channel")
			nextValue = copyCurrentPreambleData()

		case a.clientPreamble.fetch <- nextValue:
			nextValue = copyCurrentPreambleData()
			a.Debugf(DebugInit, "pushed preamble data to channel; next will be %p, pre=%p, pa=%p, pr=%p", nextValue, &nextValue.Preamble, &nextValue.PostAuth, &nextValue.PostReady)
		}
	}
}

//
// manageGameState is a goroutine which tracks the global game state for clients
//
func (a *Application) manageGameState() {
	var isInCombatMode bool
	var toolbarHidden bool
	var viewx, viewy float64
	var viewg string
	var currentTurn *mapper.UpdateTurnMessagePayload
	var currentInitiativeList *mapper.UpdateInitiativeMessagePayload
	var currentTime *mapper.UpdateClockMessagePayload

	newStatusMarkers := make(map[string]mapper.UpdateStatusMarkerMessagePayload)

	// eventHistory maps an event token to a server message we received. The token may be one of these:
	//   new:<id>			creation of new element. supercedes existing *:<id> entries when added.
	//   add:<id>:<attr>	add value(s) to <attr> of object
	//   del:<id>:<attr>	delete value(s) from <attr> of object
	//   mod:<id>			modification of attributes of an object
	//   llf:<name>			load local file
	//   lsf:<name>			load remote file
	//   ulf:<name>			unload local file
	//   usf:<name>			unload remote file
	eventHistory := make(map[string]*mapper.MessagePayload)

	a.Log("game state manager started")
	defer a.Log("game state manager stopped")

	recordElement := func(id string, e *mapper.MessagePayload) {
		if InstrumentCode {
			if a.NrApp != nil {
				defer a.NrApp.StartTransaction("record-element").End()
			}
		}
		for k, _ := range eventHistory {
			if strings.Contains(k, ":"+id) {
				delete(eventHistory, k)
			}
		}
		eventHistory["new:"+id] = e
	}

	for {
		select {
		case event := <-a.gameState.update:
			if event == nil {
				a.Log("received nil event to update game state")
				continue
			}
			a.Debugf(DebugState, "updating game state from event %v", *event)
			switch p := (*event).(type) {
			case mapper.AddObjAttributesMessagePayload:
				func() {
					if InstrumentCode {
						if a.NrApp != nil {
							defer a.NrApp.StartTransaction("track-add-obj-attributes").End()
						}
					}

					// TODO this could be more efficient
					if o, ok := eventHistory["del:"+p.ObjID+":"+p.AttrName]; ok {
						obj, valid := (*o).(mapper.RemoveObjAttributesMessagePayload)
						if !valid {
							a.Logf("value of eventHistory[del:%s:%s] is of type %T (removed)", p.ObjID, p.AttrName, o)
							delete(eventHistory, "del:"+p.ObjID+":"+p.AttrName)
						} else {
							for _, addedValue := range p.Values {
								if pos := slices.Index(obj.Values, addedValue); pos >= 0 {
									// we previously tracked deletion of this, so remove from the delete list now
									slices.Delete(obj.Values, pos, pos+1)
								}
							}
						}
					}
					for _, addedValue := range p.Values {
						if o, ok := eventHistory["add:"+p.ObjID+":"+p.AttrName]; ok {
							obj, valid := (*o).(mapper.AddObjAttributesMessagePayload)
							if !valid {
								a.Logf("value of eventHistory[add:%s:%s] is of type %T (removed)", p.ObjID, p.AttrName, o)
								delete(eventHistory, "add:"+p.ObjID+":"+p.AttrName)
							} else {
								if slices.Contains(obj.Values, addedValue) {
									// we already have a note to add this value, do nothing
								} else {
									// add this to our existing add: record
									obj.Values = append(obj.Values, addedValue)
								}
							}
						} else {
							// we need a new add: record for this attribute
							var pl mapper.MessagePayload
							pl = mapper.AddObjAttributesMessagePayload{
								ObjID:    p.ObjID,
								AttrName: p.AttrName,
								Values: []string{
									addedValue,
								},
							}
							eventHistory["add:"+p.ObjID+":"+p.AttrName] = &pl
						}
					}
				}()

			case mapper.AdjustViewMessagePayload:
				viewx = p.XView
				viewy = p.YView
				viewg = p.Grid
			case mapper.ClearMessagePayload:
				func() {
					if InstrumentCode {
						if a.NrApp != nil {
							defer a.NrApp.StartTransaction("track-clear").End()
						}
					}
					switch p.ObjID {
					case "*":
						viewx = 0.0
						viewy = 0.0
						viewg = ""
						eventHistory = make(map[string]*mapper.MessagePayload)

					case "E*":
						for k, v := range eventHistory {
							if !strings.HasPrefix(k, "new:") {
								delete(eventHistory, k)
							} else if _, isCreature := (*v).(mapper.PlaceSomeoneMessagePayload); !isCreature {
								delete(eventHistory, k)
							}
						}

					case "M*":
						for k, v := range eventHistory {
							if strings.HasPrefix(k, "new:") {
								if creature, ok := (*v).(mapper.PlaceSomeoneMessagePayload); ok {
									if creature.CreatureType != 2 {
										delete(eventHistory, k)
									}
								}
							}
						}

					case "P*":
						for k, v := range eventHistory {
							if strings.HasPrefix(k, "new:") {
								if creature, ok := (*v).(mapper.PlaceSomeoneMessagePayload); ok {
									if creature.CreatureType == 2 {
										delete(eventHistory, k)
									}
								}
							}
						}

					default:
						if pos := strings.IndexRune(p.ObjID, '='); pos > 0 {
							p.ObjID = p.ObjID[pos+1:]
						}

						for k, v := range eventHistory {
							if creature, ok := (*v).(mapper.PlaceSomeoneMessagePayload); ok {
								if creature.Name == p.ObjID {
									delete(eventHistory, k)
									continue
								}
							}
							f := strings.Split(k, ":")
							if len(f) > 1 && f[1] == p.ObjID {
								delete(eventHistory, k)
							}
						}
					}
				}()

			case mapper.ClearFromMessagePayload:
				if p.IsLocalFile {
					delete(eventHistory, "llf:"+p.File)
					eventHistory["ulf:"+p.File] = event
				} else {
					delete(eventHistory, "lsf:"+p.File)
					eventHistory["usf:"+p.File] = event
				}

			case mapper.CombatModeMessagePayload:
				isInCombatMode = p.Enabled

			case mapper.LoadArcObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadCircleObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadLineObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadPolygonObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadRectangleObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadSpellAreaOfEffectObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadTextObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.LoadTileObjectMessagePayload:
				recordElement(p.ID, event)
			case mapper.PlaceSomeoneMessagePayload:
				recordElement(p.ID, event)

			case mapper.LoadFromMessagePayload:
				if p.IsLocalFile {
					delete(eventHistory, "ulf:"+p.File)
					eventHistory["llf:"+p.File] = event
				} else {
					delete(eventHistory, "usf:"+p.File)
					eventHistory["lsf:"+p.File] = event
				}

			case mapper.RemoveObjAttributesMessagePayload:
				func() {
					if InstrumentCode {
						if a.NrApp != nil {
							defer a.NrApp.StartTransaction("track-remove-obj-attributes").End()
						}
					}
					// TODO this could be more efficient
					if o, ok := eventHistory["add:"+p.ObjID+":"+p.AttrName]; ok {
						obj, valid := (*o).(mapper.AddObjAttributesMessagePayload)
						if !valid {
							a.Logf("value of eventHistory[add:%s:%s] is of type %T (removed)", p.ObjID, p.AttrName, o)
							delete(eventHistory, "add:"+p.ObjID+":"+p.AttrName)
						} else {
							for _, addedValue := range p.Values {
								if pos := slices.Index(obj.Values, addedValue); pos >= 0 {
									// we previously tracked addition of this, so remove from the add list now
									slices.Delete(obj.Values, pos, pos+1)
								}
							}
						}
					}
					for _, addedValue := range p.Values {
						if o, ok := eventHistory["del:"+p.ObjID+":"+p.AttrName]; ok {
							obj, valid := (*o).(mapper.RemoveObjAttributesMessagePayload)
							if !valid {
								a.Logf("value of eventHistory[del:%s:%s] is of type %T (removed)", p.ObjID, p.AttrName, o)
								delete(eventHistory, "del:"+p.ObjID+":"+p.AttrName)
							} else {
								if slices.Contains(obj.Values, addedValue) {
									// we already have a note to remove this value, do nothing
								} else {
									// add this to our existing del: record
									obj.Values = append(obj.Values, addedValue)
								}
							}
						} else {
							// we need a new del: record for this attribute
							var pl mapper.MessagePayload
							pl = mapper.RemoveObjAttributesMessagePayload{
								ObjID:    p.ObjID,
								AttrName: p.AttrName,
								Values: []string{
									addedValue,
								},
							}
							eventHistory["del:"+p.ObjID+":"+p.AttrName] = &pl
						}
					}
				}()

			case mapper.ToolbarMessagePayload:
				toolbarHidden = !p.Enabled

			case mapper.UpdateObjAttributesMessagePayload:
				func() {
					if InstrumentCode {
						if a.NrApp != nil {
							defer a.NrApp.StartTransaction("track-update-obj-attributes").End()
						}
					}
					if o, ok := eventHistory["mod:"+p.ObjID]; ok {
						old, valid := (*o).(mapper.UpdateObjAttributesMessagePayload)
						if !valid {
							a.Logf("value of eventHistory[mod:%s] is of type %T (removed)", p.ObjID, o)
							delete(eventHistory, "mod:"+p.ObjID)
						} else {
							// we already have a record for this; edit in place
							for attrName, attrValue := range p.NewAttrs {
								old.NewAttrs[attrName] = attrValue
								// If we have add: or del: events for this object, this supercedes them
								delete(eventHistory, "add:"+p.ObjID+":"+attrName)
								delete(eventHistory, "del:"+p.ObjID+":"+attrName)
							}
						}
					} else {
						eventHistory["mod:"+p.ObjID] = event
						for attrName, _ := range p.NewAttrs {
							// If we have add: or del: events for this object, this supercedes them
							delete(eventHistory, "add:"+p.ObjID+":"+attrName)
							delete(eventHistory, "del:"+p.ObjID+":"+attrName)
						}
					}
				}()

			case mapper.UpdateStatusMarkerMessagePayload:
				newStatusMarkers[p.Condition] = p

			case mapper.UpdateTurnMessagePayload:
				currentTurn = &p

			case mapper.UpdateInitiativeMessagePayload:
				currentInitiativeList = &p

			case mapper.UpdateClockMessagePayload:
				currentTime = &p

			default:
				a.Logf("unknown event %v (can't update game state)", *event)
			}

		case client := <-a.gameState.sync:
			func() {
				if InstrumentCode {
					if a.NrApp != nil {
						defer a.NrApp.StartTransaction("sync").End()
					}
				}
				a.Debugf(DebugState, "client %v requests SYNC", client.IdTag())
				client.Conn.Send(mapper.CombatMode, mapper.CombatModeMessagePayload{Enabled: isInCombatMode})
				client.Conn.Send(mapper.Toolbar, mapper.ToolbarMessagePayload{Enabled: !toolbarHidden})
				client.Conn.Send(mapper.AdjustView, mapper.AdjustViewMessagePayload{Grid: viewg, XView: viewx, YView: viewy})
				if currentTurn == nil {
					client.Conn.Send(mapper.Comment, "no current turn set")
				} else {
					client.Conn.Send(mapper.UpdateTurn, *currentTurn)
				}
				if currentInitiativeList == nil {
					client.Conn.Send(mapper.Comment, "no current initiative list set")
				} else {
					client.Conn.Send(mapper.UpdateInitiative, *currentInitiativeList)
				}
				if currentTime == nil {
					client.Conn.Send(mapper.Comment, "no current time set")
				} else {
					client.Conn.Send(mapper.UpdateClock, *currentTime)
				}
				for _, marker := range newStatusMarkers {
					client.Conn.Send(mapper.UpdateStatusMarker, marker)
				}

				for k, e := range eventHistory {
					if strings.HasPrefix(k, "llf:") || strings.HasPrefix(k, "lsf:") {
						client.Conn.Send((*e).MessageType(), *e)
					}
				}

				for k, e := range eventHistory {
					if strings.HasPrefix(k, "ulf:") || strings.HasPrefix(k, "usf:") {
						client.Conn.Send((*e).MessageType(), *e)
					}
				}

				for k, e := range eventHistory {
					if strings.HasPrefix(k, "new:") {
						client.Conn.Send((*e).MessageType(), *e)
					}
				}

				for k, e := range eventHistory {
					if strings.HasPrefix(k, "add:") || strings.HasPrefix(k, "del:") || strings.HasPrefix(k, "mod:") {
						client.Conn.Send((*e).MessageType(), *e)
					}
				}

				a.Debug(DebugState, "SYNC operation completed")
			}()
		}
	}
}

func (a *Application) UpdateGameState(event *mapper.MessagePayload) {
	a.gameState.update <- event
}

func (a *Application) SendGameState(client *mapper.ClientConnection) {
	a.gameState.sync <- client
}
