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
*/

// Package mapper implements a standard client interface for the mapper service.
//
// This package handles the details of communicating with the
// GMA mapper service communication channel used to keep the mapper
// clients in sync with each other and with the other GMA tools.
//
// A client should establish a connection to the game server by
// calling the Dial method in this package. This function will
// sign on to the server and then enter a loop, sending incoming
// server messages back on the channel(s) established via the
// Subscribe method. Dial returns when the session with the
// server has terminated.
//
// Typically, an application will invoke the Dial method in a
// goroutine. Calling the associated context's cancel function
// will signal that we want to stop talking to the server, resulting
// in the termination of the running Dial method.
package mapper

//
// Since there's a fair amount of code below which is logically
// divided up by server message type (sending or receiving), we
// will use large banners to make it easy to scroll quickly
// and visually distinguish each section with ease.
//

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/util"
)

// ErrAuthenticationRequired is the error returned when the server requires authentication but we didn't provide any.
var ErrAuthenticationRequired = errors.New("authenticator required for connection")

// ErrAuthenticationFailed is the error returned when our authentication was rejected by the server.
var ErrAuthenticationFailed = errors.New("access denied to server")

// ErrServerProtocolError is the error returned when something fundamental about the server's conversation
// with us is so wrong we can't even deal with the conversation any further.
var ErrServerProtocolError = errors.New("server protocol error; unable to continue")

// ErrRetryConnection is returned when the attempt to establish a connection to the server
// fails in such a way that simply trying again would be the right thing to do.
var ErrRetryConnection = errors.New("please retry the server connection")

// Debugging information is enabled by selecting a nummber
// of discrete topics which you want logged as the application
// runs (previous versions used a "level" of verbosity which
// doesn't provide the better granularity this version provides
// to just get the info you want.
type DebugFlags uint64

const (
	DebugAuth DebugFlags = 1 << iota
	DebugBinary
	DebugEvents
	DebugIO
	DebugMessages
	DebugMisc
	DebugQoS
	DebugAll DebugFlags = 0xffffffff
)

// DebugFlagNameSlice returns a slice of debug flat names
// corresponding to the bit-encoded flags parameter.
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
		{bits: DebugBinary, name: "binary"},
		{bits: DebugEvents, name: "events"},
		{bits: DebugIO, name: "i/o"},
		{bits: DebugMessages, name: "messages"},
		{bits: DebugMisc, name: "misc"},
		{bits: DebugQoS, name: "qos"},
	} {
		if (flags & f.bits) != 0 {
			list = append(list, f.name)
		}
	}
	return list
}

// DebugFlagNames returns a string representation of
// the debugging flags (topics) stored in the DebugFlags
// value passed in.
func DebugFlagNames(flags DebugFlags) string {
	list := DebugFlagNameSlice(flags)
	if list == nil {
		return "<none>"
	}
	return "<" + strings.Join(list, ",") + ">"
}

// NamedDebugFlags takes a comma-separated list of
// debug flag (topic) names, or a list of individual
// names, or both, and returns the DebugFlags
// value which includes all of them.
//
// If "none" appears in the list, it cancels all previous
// values seen, but subsequent names will add their values
// to the list.
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
			case "auth":
				d |= DebugAuth
			case "binary":
				d |= DebugBinary
			case "events":
				d |= DebugEvents
			case "i/o", "io":
				d |= DebugIO
			case "messages":
				d |= DebugMessages
			case "misc":
				d |= DebugMisc
			case "qos":
				d |= DebugQoS
			default:
				err = fmt.Errorf("invalid debug flag name")
				// but keep processing the rest
			}
		}
	}
	return d, err
}

// Connection describes a connection to the server. These are
// created with NewConnection and then send methods such as
// Subscribe and Dial.
type Connection struct {
	// If true, we will always try to reconnect to the server if we
	// lose our connection.
	StayConnected bool

	// Do we have an active session now?
	signedOn bool

	// If nonzero, we will re-try a failing connection this many
	// times before giving up on the server. Otherwise we will keep
	// trying forever.
	Retries uint

	// The server's protocol version number.
	Protocol int

	// The verbosity level of debugging log messages.
	DebuggingLevel DebugFlags

	// If nonzero, our connection attempts will timeout after the
	// specified time interval. Otherwise they will wait indefinitely.
	Timeout time.Duration

	// The server endpoint, in any form acceptable to the net.Dial
	// function.
	Endpoint string

	// Characters received from the server.
	Characters map[string]PlayerToken

	// Conditions and their token markings received from the server.
	Conditions map[string]StatusMarkerDefinition

	// If we received pre-authentication data from the server other than
	// definition of characters and condition codes, they are technically
	// too early to be valid (the server shouldn't do anything else before
	// authenticating), so we'll merely collect them here in case they are
	// of interest forensically.
	Preamble []string

	// Any progress gauges sent by the server will be tracked here as well
	// as passed to a channel subscribed to UpdateProgress.
	Gauges map[string]*UpdateProgressMessagePayload

	// The list of advertised updates to our software
	PackageUpdatesAvailable map[string][]PackageVersion

	// The last error encountered while communicating with the server.
	LastError error

	serverConn MapConnection

	// The calendar system the server indicated as preferred, if any
	CalendarSystem string

	// Server overrides to client settings (if non-nil)
	ClientSettings *ClientSettingsOverrides

	// The context for our session, either one we created in the
	// NewConnection function or one we received from the caller.
	Context context.Context

	// If this is non-nil, we will use this to identify the user
	// to the server.
	Authenticator *auth.Authenticator

	// We will log informational messages here as we work.
	Logger *log.Logger

	// Server message subscriptions currently in effect.
	Subscriptions map[ServerMessage]chan MessagePayload

	// Our signal that we're ready for the client to talk.
	ReadySignal chan byte

	// Some statistics we know about the server
	ServerStats struct {
		Started       time.Time // server startup time
		Active        time.Time // time of last ping (at connect-time, this is time of last ping sent by server)
		ConnectTime   time.Time // time server connected (time on the server, for comparison with other server times)
		ServerVersion string    // the server's version number
	}
}

// Log writes data to our log destination.
func (c *Connection) Log(message ...any) {
	if c != nil && c.Logger != nil {
		message = append([]any{"[client] "}, message...)
		c.Logger.Print(message...)
	}
}

// Logf writes data to our log destination.
func (c *Connection) Logf(format string, data ...any) {
	if c != nil && c.Logger != nil {
		c.Logger.Printf("[client] "+format, data...)
	}
}

// IsReady returns true if the connection to the server
// has completed and authentication was successful, so
// the connection is ready for interactive use.
func (c *Connection) IsReady() bool {
	return c != nil && c.serverConn.IsReady() && c.signedOn
}

// WithContext modifies the behavior of the NewConnection function
// by supplying a context for this connection, which may be used to
// signal the Dial method that the connection to the server should
// be terminated.
//
// N.B.: When making the initial TCP connection to the server,
// if there is a timeout value specified via WithTimeout, then
// a hanging connection will terminate when that timer expires,
// regardless of the context. Otherwise, the connection will wait
// indefinitely to complete OR until the context is cancelled.
func WithContext(ctx context.Context) func(*Connection) error {
	return func(c *Connection) error {
		c.Context = ctx
		return nil
	}
}

// WhenReady specifies a channel on which to send a single byte
// when the server login process is complete and the server
// is ready to receive our commands.
func WhenReady(ch chan byte) func(*Connection) error {
	return func(c *Connection) error {
		c.ReadySignal = ch
		return nil
	}
}

// ConnectionOption is an option to be passed to the NewConnection
// function.
type ConnectionOption func(*Connection) error

// WithSubscription modifies the behavior of the NewConnection function
// by adding a server message subscription to the connection just as if
// the Subscribe method had been called on the connection value.
//
// For example, this:
//
//	server, err := NewConnection(endpoint,
//	                 WithSubscription(chats, ChatMessage, RollResult),
//	                 WithSubscription(oops, ERROR, UNKNOWN))
//	go server.Dial()
//
// is equivalent to this:
//
//	server, err := NewConnection(endpoint)
//	err = server.Subscribe(chats, ChatMessage, RollResult)
//	err = server.Subscribe(oops, ERROR, UNKNOWN)
//	go server.Dial()
//
// (Of course, real production code should check the returned error values.)
func WithSubscription(ch chan MessagePayload, messages ...ServerMessage) ConnectionOption {
	return func(c *Connection) error {
		return c.Subscribe(ch, messages...)
	}
}

// WithAuthenticator modifies the behavior of the NewConnection function
// by adding an authenticator which will be used to identify the client
// to the server. If this option is not given, no attempt will be made
// to authenticate, which is only appropriate for servers which do not
// require authentication. (Which, hopefully, won't be the case anyway.)
func WithAuthenticator(a *auth.Authenticator) ConnectionOption {
	return func(c *Connection) error {
		c.Authenticator = a
		return nil
	}
}

// WithLogger modifies the behavior of the NewConnection function
// by specifying a custom logger instead of the default one for
// the Connection to use during its operations.
func WithLogger(l *log.Logger) ConnectionOption {
	return func(c *Connection) error {
		c.Logger = l
		return nil
	}
}

// WithTimeout modifies the behavior of the NewConnection function
// by specifying the time to allow the Dial method to make the TCP
// connection to the server. After this time expires, the attempt
// is abandoned (but may be retried based on the value of
// WithRetries, if any).
//
// N.B.: When making the initial TCP connection to the server,
// if there is a timeout value specified via WithTimeout, then
// a hanging connection will terminate when that timer expires,
// regardless of the context (although a canceled context will
// stop retry attempts). Otherwise, the connection will wait
// indefinitely to complete OR until the context is cancelled.
func WithTimeout(t time.Duration) ConnectionOption {
	return func(c *Connection) error {
		c.Timeout = t
		return nil
	}
}

// WithRetries modifies the behavior of the NewConnection function
// to indicate how many times the Dial method should try to
// establish a connection to the server before giving up.
//
// Setting this to 0 means to retry infinitely many times.
// The default is to make a single attempt to connect to the
// server.
func WithRetries(n uint) ConnectionOption {
	return func(c *Connection) error {
		c.Retries = n
		return nil
	}
}

// StayConnected modifies the behavior of the NewConnection call so that
// when Dial is called on the new Connection, it will
// continue to try to re-establish connections to the server
// (if enable is true) until it utterly fails in the attempt.
// This is useful in case connections to the server tend to
// get inadvertently dropped, since this will allow the client
// to automatically reconnect and resume operations.
//
// If enable is false (the default), Dial will return as soon
// as the server connection is dropped for any reason.
func StayConnected(enable bool) ConnectionOption {
	return func(c *Connection) error {
		c.StayConnected = enable
		return nil
	}
}

// WithDebugging modifies the behavior of the NewConnection function
// so that the operations of the Connection's interaction with the
// server are logged to varying levels of verbosity.
func WithDebugging(flags DebugFlags) ConnectionOption {
	return func(c *Connection) error {
		c.DebuggingLevel = flags
		return nil
	}
}

// NewConnection creates a new server connection value which can then be used to
// manage our communication with the server.
//
// After the endpoint, you may specify any of the following options
// to define the behavior desired for this connection:
//
//	StayConnected(bool)
//	WithAuthenticator(a)
//	WithDebugging(level)
//	WithContext(ctx)
//	WithLogger(l)
//	WithRetries(n)
//	WithSubscription(ch, msgs...)
//	WithTimeout(t)
//
// Example:
//
//	a := NewClientAuthenticator("fred", []byte("sekret"), "some random client")
//	ctx, cancel := context.Background()
//	defer cancel()
//
//	messages := make(chan MessagePayload, 10)
//	problems := make(chan MessagePayload, 10)
//
//	server, err := NewConnection("mygame.example.org:2323",
//	                  WithAuthenticator(a),
//	                  WithContext(ctx),
//	                  StayConnected(true),
//	                  WithSubscription(messages, ChatMessage, RollResult),
//	                  WithSubscription(problems, ERROR, UNKNOWN))
//	if err != nil {
//	   log.Fatalf("can't reach the server: %v", err)
//	}
//	go server.Dial()
func NewConnection(endpoint string, opts ...ConnectionOption) (Connection, error) {
	newCon := Connection{
		Context:  context.Background(),
		Endpoint: endpoint,
		Retries:  1,
		Logger:   log.Default(),
	}
	newCon.Reset()
	newCon.serverConn.debug = newCon.debug
	newCon.serverConn.debugf = newCon.debugf
	newCon.serverConn.bLock = new(sync.Mutex)

	for _, o := range opts {
		if err := o(&newCon); err != nil {
			return newCon, err
		}
	}

	return newCon, nil
}

// Reset returns an existing Connection object to an appropriate pre-connect state.
func (c *Connection) Reset() {
	if c == nil {
		return
	}
	c.PartialReset()
	c.Subscriptions = make(map[ServerMessage]chan MessagePayload)
}

// PartialReset returns an existing Connection object with the connection-related values
// reset to their pre-connect state, but leaving other things like the subscription list
// intact.
func (c *Connection) PartialReset() {
	if c == nil {
		return
	}
	c.signedOn = false
	c.Characters = make(map[string]PlayerToken)
	c.Conditions = make(map[string]StatusMarkerDefinition)
	c.Gauges = make(map[string]*UpdateProgressMessagePayload)
	c.PackageUpdatesAvailable = make(map[string][]PackageVersion)
	c.serverConn.sendChan = make(chan string, 16)
	c.Preamble = nil
	c.ClientSettings = nil
}

// Log debugging info at the given level.
func (c *Connection) debug(level DebugFlags, msg string) {
	if c != nil && (c.DebuggingLevel&level) != 0 {
		for i, line := range strings.Split(msg, "\n") {
			if line != "" {
				c.Logf("DEBUG%s%02d: %s", DebugFlagNames(level), i, line)
			}
		}
	}
}

func (c *Connection) debugf(level DebugFlags, format string, args ...any) {
	if c != nil && (c.DebuggingLevel&level) != 0 {
		args = append([]any{DebugFlagNames(level)}, args...)
		c.Logf("DEBUG%s: "+format, args...)
	}
}

// Close terminates the connection to the server.
// Note that the Dial function normally closes the connection
// before it returns, so calling this explicitly should not
// normally be necessary.
//
// Calling Close will result in the Dial function stopping
// due to the connection disappearing, but it is better to cancel
// the context being watched by Dial instead.
func (c *Connection) Close() {
	if c != nil {
		c.debug(DebugIO, "Close()")
		c.serverConn.Close()
	}
}

// Subscribe arranges for server messages to be sent to the specified channel
// when they arrive.
//
// If multiple messages are specified, they are all directed to send their payloads
// to the channel, which may use the MessageType method to differentiate what
// kind of payload was sent.
//
// This method may be called multiple times for the same channel, in which case
// the specified message(s) are added to the set which sends to that channel.
//
// If another Subscribe method is called with the same ServerMessage that a
// previous Subscribe mentioned, that will change the subscription for that
// message to go to the new channel instead of the previous one.
//
// Unless subscribed, the following default behaviors are assumed:
//
//	Marco:   Auto-reply with Polo
//	ERROR:   Log a message
//	UNKNOWN: Log a message
//
// If any of these are subscribed to, then the default behavior is NOT taken,
// on the assumption that the code consuming the subscribed events will fully
// handle an appropriate response.
//
// Further, if AddCharacter or UpdateStatusMarker messages are received from
// the server, the Connection struct's Characters and Conditions maps are
// automatically updated (respectively) regardless of whether they are
// subscribed to.
//
// The default behavior for all other incoming server messages is to ignore
// them completely. The client will ask the server not to send any non-subscribed
// messages.
//
// This method may be called on an established connection to change the subscription
// list on the fly.
//
// If the channel is nil, the message(s) are unsubscribed and will not be
// received by the client until subscribed to again.
//
// Example: (error checking not shown for the sake of brevity)
//
//	cm := make(chan MessagePayload, 1)
//	service, err := NewConnection(endpoint)
//	err = service.Subscribe(cm, ChatMessage)
func (c *Connection) Subscribe(ch chan MessagePayload, messages ...ServerMessage) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}

	for _, m := range messages {
		if m >= maximumServerMessage {
			return fmt.Errorf("server message ID %v not defined (illegal Subscribe call)", m)
		}
		if ch == nil {
			delete(c.Subscriptions, m)
		} else {
			c.Subscriptions[m] = ch
		}
	}
	return c.filterSubscriptions()
}

// MessagePayload is an interface that includes any kind of message the server will
// send to us.
type MessagePayload interface {
	MessageType() ServerMessage
	RawMessage() string
	RawBytes() []byte
}

// ServerMessage is an arbitrary code which identifies specific message types that
// we can receive from the server. This value is passed to the Subscribe method
// and returned by the MessageType method. These values are intended for use
// within an actively-running program but are not guaranteed to remain stable across
// new releases of the code, so they should not be stored and re-used by a later
// execution of the client, nor passed to other programs whose definition of these
// values may not agree.
type ServerMessage byte

// Despite the warning above, we'll do our best to avoid changing these values
// if at all possible.

// ServerMessage values (see the comments accompanying the type definition).
const (
	Accept ServerMessage = iota
	AddAudio
	AddCharacter
	AddDicePresets
	AddImage
	AddObjAttributes
	AdjustView
	Allow
	Auth
	BatchFragment
	Challenge
	CharacterName
	ChatMessage
	Clear
	ClearChat
	ClearFrom
	CombatMode
	Comment
	DefineDicePresets
	DefineDicePresetDelegates
	Denied
	Echo
	Failed
	FilterAudio
	FilterCoreData
	FilterDicePresets
	FilterImages
	Granted
	HitPointAcknowledge
	HitPointRequest
	LoadFrom
	LoadArcObject
	LoadCircleObject
	LoadLineObject
	LoadPolygonObject
	LoadRectangleObject
	LoadSpellAreaOfEffectObject
	LoadTextObject
	LoadTileObject
	Marco
	Mark
	PlaceSomeone
	PlayAudio
	Polo
	Priv
	Protocol
	QueryAudio
	QueryCoreData
	QueryCoreIndex
	QueryDicePresets
	QueryImage
	QueryPeers
	Ready
	Redirect
	RemoveObjAttributes
	RollDice
	RollResult
	Sync
	SyncChat
	TimerAcknowledge
	TimerRequest
	Toolbar
	UpdateClock
	UpdateCoreData
	UpdateCoreIndex
	UpdateDicePresets
	UpdateInitiative
	UpdateObjAttributes
	UpdatePeerList
	UpdateProgress
	UpdateStatusMarker
	UpdateTurn
	UpdateVersions
	World
	UNKNOWN
	ERROR
	maximumServerMessage
)

var ServerMessageByName = map[string]ServerMessage{
	"Accept":                      Accept,
	"AddAudio":                    AddAudio,
	"AddCharacter":                AddCharacter,
	"AddDicePresets":              AddDicePresets,
	"AddImage":                    AddImage,
	"AddObjAttributes":            AddObjAttributes,
	"AdjustView":                  AdjustView,
	"Allow":                       Allow,
	"Auth":                        Auth,
	"BatchFragment":			   BatchFragment,
	"Challenge":                   Challenge,
	"CharacterName":               CharacterName,
	"ChatMessage":                 ChatMessage,
	"Clear":                       Clear,
	"ClearChat":                   ClearChat,
	"ClearFrom":                   ClearFrom,
	"CombatMode":                  CombatMode,
	"Comment":                     Comment,
	"DefineDicePresets":           DefineDicePresets,
	"DefineDicePresetDelegates":   DefineDicePresetDelegates,
	"Denied":                      Denied,
	"Echo":                        Echo,
	"Failed":                      Failed,
	"FilterAudio":                 FilterAudio,
	"FilterCoreData":              FilterCoreData,
	"FilterDicePresets":           FilterDicePresets,
	"FilterImages":                FilterImages,
	"HitPointAcknowledge":         HitPointAcknowledge,
	"HitPointRequest":             HitPointRequest,
	"Granted":                     Granted,
	"LoadFrom":                    LoadFrom,
	"LoadArcObject":               LoadArcObject,
	"LoadCircleObject":            LoadCircleObject,
	"LoadLineObject":              LoadLineObject,
	"LoadPolygonObject":           LoadPolygonObject,
	"LoadRectangleObject":         LoadRectangleObject,
	"LoadSpellAreaOfEffectObject": LoadSpellAreaOfEffectObject,
	"LoadTextObject":              LoadTextObject,
	"LoadTileObject":              LoadTileObject,
	"Marco":                       Marco,
	"Mark":                        Mark,
	"PlaceSomeone":                PlaceSomeone,
	"PlayAudio":                   PlayAudio,
	"Polo":                        Polo,
	"Priv":                        Priv,
	"Protocol":                    Protocol,
	"QueryAudio":                  QueryAudio,
	"QueryCoreData":               QueryCoreData,
	"QueryCoreIndex":              QueryCoreIndex,
	"QueryDicePresets":            QueryDicePresets,
	"QueryImage":                  QueryImage,
	"QueryPeers":                  QueryPeers,
	"Ready":                       Ready,
	"Redirect":                    Redirect,
	"RemoveObjAttributes":         RemoveObjAttributes,
	"RollDice":                    RollDice,
	"RollResult":                  RollResult,
	"Sync":                        Sync,
	"SyncChat":                    SyncChat,
	"TimerAcknowledge":            TimerAcknowledge,
	"TimerRequest":                TimerRequest,
	"Toolbar":                     Toolbar,
	"UpdateClock":                 UpdateClock,
	"UpdateCoreData":              UpdateCoreData,
	"UpdateCoreIndex":             UpdateCoreIndex,
	"UpdateDicePresets":           UpdateDicePresets,
	"UpdateInitiative":            UpdateInitiative,
	"UpdateObjAttributes":         UpdateObjAttributes,
	"UpdatePeerList":              UpdatePeerList,
	"UpdateProgress":              UpdateProgress,
	"UpdateStatusMarker":          UpdateStatusMarker,
	"UpdateTurn":                  UpdateTurn,
	"UpdateVersions":              UpdateVersions,
	"World":                       World,
}

// BaseMessagePayload is not a payload type that you should ever
// encounter directly, but it is included in all other payload
// types. It holds the bare minimum data for any server message.
type BaseMessagePayload struct {
	rawMessage  string        `json:"-"`
	messageType ServerMessage `json:"-"`
}

// RawMessage returns the raw message received from the server before
// it was parsed out into the MessagePayload the client should arguably
// be looking at instead.
//
// The raw message data may be useful for debugging purposes or other
// low-level poking around, though, so we make it available here.
func (p BaseMessagePayload) RawMessage() string { return p.rawMessage }
func (p BaseMessagePayload) RawBytes() []byte   { return []byte(p.rawMessage) }

// MessageType returns the type of message this MessagePayload represents.
// This value will be the same as the ServerMessage value used for the
// Subscribe function, and may be used with channels which receive multiple
// kinds of messages to differentiate them, like so:
//
//	select {
//	case p<-messages:
//	    // This channel may receive a ChatMessage or RollResult.
//	    switch p.MessageType() {
//	    case ChatMessage:
//	        // Do whatever with p.(ChatMessageMessagePayload)
//	    case RollResult:
//	        // Do whatever with p.(RollResultMessagePayload)
//	    default:
//	        // Something bad happened!
//	    }
//	 ...
//	}
//
// You can also use a type switch to accomplish the same thing and avoid
// the explicit type assertions:
//
//	select {
//	case p<-messages:
//	    // This channel may receive a ChatMessage or RollResult.
//	    switch msg := p.(type) {
//	    case ChatMessageMessagePayload:
//	        // Do whatever with msg
//	    case RollResultMessagePayload:
//	        // Do whatever with msg
//	    default:
//	        // Something bad happened!
//	    }
//	 ...
//	}
func (p BaseMessagePayload) MessageType() ServerMessage { return p.messageType }

//TODO//TODO// BatchableMessagePayload is a type of payload which can be broken into pieces to avoid sending excessively long data.
//TODO//TODOtype BatchableMessagePayload struct {
//TODO	Batch        int    `json:",omitempty"`
//TODO	TotalBatches int    `json:",omitempty"`
//TODO	BatchGroup   string `json:",omitempty"`
//TODO	BatchError   string `json:",omitempty"`
//TODO}
//TODO
//TODO// NeedsToBeSplit returns true if we estimate that the payload is likely to exceed the maximum message size allowed.
//TODO//TODOfunc (b BatchableMessagePayload) NeedsToBeSplit() bool {
//TODO	return false // this needs to be overridden by specific types that actually have payloads to examine
//TODO}
//TODO
//TODO// IsBatched returns true if this payload is part of a batched set as opposed to a stand-alone packet.
//TODO//TODOfunc (b BatchableMessagePayload) IsBatched() bool {
//TODO	return b.BatchGroup != ""
//TODO}
//TODO
//TODO// BatchInfo returns the details about the batch for each fragment.
//TODO//TODOfunc (b BatchableMessagePayload) BatchInfo() BatchableMessagePayload {
//TODO	return b
//TODO}
//TODO
//TODO// Split breaks up a payload into multiple parts, returning a slice of them.
//TODO// This must be defined at the derived type level.
//TODO//TODO//func (b BatchableMessagePayload) Split() []any {
//TODO//}
//TODO
//TODO// Reassemble combines fragmented pieces back into a single payload. This must be defined at the derived type level.
//TODO//TODO// func (b BatchableMessagePayload) Reassemble([]any) any {}
//TODO
//TODO// AbortPayload generates a new payload of the same type as the original but with batch fields which indicate that we are
//TODO// bailing out on batching up the payload and won't be completing the operation.
//TODO// This must be defined at the derived type level.
//TODO//TODO// func (b BatchableMessagePayload) AbortPayload(reason string, batchNumber int) any {}

// ErrorMessagePayload describes
// an error which encountered when trying to receive a message.
type ErrorMessagePayload struct {
	BaseMessagePayload
	OriginalMessageType ServerMessage
	Error               error
}

// UnknownMessagePayload describes a server message we received
// but have no idea what it is.
type UnknownMessagePayload struct {
	BaseMessagePayload
}

// ProtocolMessagePayload describes the server's statement of
// what protocol version it implements.
type ProtocolMessagePayload struct {
	BaseMessagePayload
	ProtocolVersion int
}

//
//     _                      _
//    / \   ___ ___ ___ _ __ | |_
//   / _ \ / __/ __/ _ \ '_ \| __|
//  / ___ \ (_| (_|  __/ |_) | |_
// /_/   \_\___\___\___| .__/ \__|
//                     |_|
//

// AcceptMessagePayload holds the information sent by a client requesting
// that the server only send a subset of its possible message types to it.
//
// Clients send this by calling the Subscribe method on their connection.
type AcceptMessagePayload struct {
	BaseMessagePayload

	// Messages is a list of message command words.
	Messages []string `json:",omitempty"`
}

//________________________________________________________________________________
//     _       _     _  ____ _                          _
//    / \   __| | __| |/ ___| |__   __ _ _ __ __ _  ___| |_ ___ _ __
//   / _ \ / _` |/ _` | |   | '_ \ / _` | '__/ _` |/ __| __/ _ \ '__|
//  / ___ \ (_| | (_| | |___| | | | (_| | | | (_| | (__| ||  __/ |
// /_/   \_\__,_|\__,_|\____|_| |_|\__,_|_|  \__,_|\___|\__\___|_|
//

// AddCharacterMessagePayload holds the information sent by the server's AddCharacter
// message to add a new PC to the party. This is not done for most creatures
// and NPCs encountered; it is for the PCs and significant NPCs who are important
// enough to be treated specially by clients (such as being included in menus).
type AddCharacterMessagePayload struct {
	BaseMessagePayload
	PlayerToken
}

//________________________________________________________________________________
//     _       _     _    _             _ _
//    / \   __| | __| |  / \  _   _  __| (_) ___
//   / _ \ / _` |/ _` | / _ \| | | |/ _` | |/ _ \
//  / ___ \ (_| | (_| |/ ___ \ |_| | (_| | | (_) |
// /_/   \_\__,_|\__,_/_/   \_\__,_|\__,_|_|\___/
//

type AudioDefinition struct {
	IsLocalFile bool
	Name        string
	File        string
	Format      string
}

// AddAudioMessagePayload holds the information sent by the server's AddAudio
// message informing the client as to where it can locate an audio clip's data.
//
// Call the AddAudio method to send this message out to others if you know
// of an audio file they should be aware of.
type AddAudioMessagePayload struct {
	BaseMessagePayload
	AudioDefinition
}

// AddAudio informs the server and peers about an image they can use.
func (c *Connection) AddAudio(adef AudioDefinition) error {
	return c.serverConn.Send(AddAudio, adef)
}

//________________________________________________________________________________
//     _       _     _ ___
//    / \   __| | __| |_ _|_ __ ___   __ _  __ _  ___
//   / _ \ / _` |/ _` || || '_ ` _ \ / _` |/ _` |/ _ \
//  / ___ \ (_| | (_| || || | | | | | (_| | (_| |  __/
// /_/   \_\__,_|\__,_|___|_| |_| |_|\__,_|\__, |\___|
//                                         |___/

// AddImageMessagePayload holds the information sent by the server's AddImage
// message informing the client as to where it can locate an image's data.
//
// Call the AddImage method to send this message out to others if you know
// of an image file they should be aware of.
type AddImageMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ImageDefinition
}

//TODOfunc (c AddImageMessagePayload) NeedsToBeSplit() bool {
//TODO	// AI Animation Name Sizes :
//TODO	l := 6               // "AI {}"
//TODO	l += 8 + len(c.Name) // "Name":,
//TODO	if c.Animation != nil {
//TODO		l += 61 // "Animation":{"Frames":99999,"FrameSpeed":99999,"Loops":99999}
//TODO	}
//TODO	l += 11 // "Sizes": []
//TODO	for _, i := range c.Sizes {
//TODO		if i.IsLocalFile {
//TODO			l += 19 // "IsLocalFile":true,
//TODO		}
//TODO		if i.ImageData != nil {
//TODO			l += 15 + len(i.ImageData) // "ImageData":"",
//TODO		}
//TODO		l += 10 + len(i.File) // "File":"",
//TODO		l += 17               // "Zoom":99999.9999
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODO// Split records on Sizes (which are the instances of the images we
//TODO// have at different zoom factors, etc.)
//TODO//
//TODO// 0  Name,
//TODO// 0  Animation->{Frames,FrameSpeed,Loops}
//TODO// 0+ Sizes[File,ImageData,IsLocalFile,Zoom]
//TODOfunc (c AddImageMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Sizes))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Sizes {
//TODO		p := AddImageMessagePayload{
//TODO			ImageDefinition: ImageDefinition{
//TODO				Sizes: []ImageInstance{instance},
//TODO			},
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Sizes),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Name = c.Name
//TODO			p.Animation = c.Animation
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c AddImageMessagePayload) Reassemble(p []any) (any, error) {
//TODO	ai := AddImageMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		img, ok := d.(AddImageMessagePayload)
//TODO		if !ok {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if img.Batch != i {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), img.Batch)
//TODO		}
//TODO		if img.TotalBatches != len(p) {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, img.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			ai.Name = img.Name
//TODO			ai.Sizes = make([]ImageInstance, len(p))
//TODO			ai.Animation = img.Animation
//TODO		}
//TODO		ai.Sizes[i] = img.Sizes[0]
//TODO	}
//TODO	return ai, nil
//TODO}
//TODO
//TODOfunc (c AddImageMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return AddImageMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ImageDefinition: ImageDefinition{
//TODO			Name: c.Name,
//TODO		},
//TODO	}
//TODO}

// AddImage informs the server and peers about an image they can use.
func (c *Connection) AddImage(idef ImageDefinition) error {
	return c.serverConn.Send(AddImage, idef)
}

//     _       _     _  ___  _     _    _   _   _        _ _           _
//    / \   __| | __| |/ _ \| |__ (_)  / \ | |_| |_ _ __(_) |__  _   _| |_ ___  ___
//   / _ \ / _` |/ _` | | | | '_ \| | / _ \| __| __| '__| | '_ \| | | | __/ _ \/ __|
//  / ___ \ (_| | (_| | |_| | |_) | |/ ___ \ |_| |_| |  | | |_) | |_| | ||  __/\__ \
// /_/   \_\__,_|\__,_|\___/|_.__// /_/   \_\__|\__|_|  |_|_.__/ \__,_|\__\___||___/
//                              |__/

// AddObjAttributesMessagePayload holds the information sent by the server's AddObjAttributes
// message. This tells the client to adjust the multi-value attribute
// of the object with the given ID by adding the new values to it.
//
// Call the AddObjAttributes method to send this message out to other clients.
type AddObjAttributesMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ObjID    string
	AttrName string
	Values   []string
}

//TODOfunc (c AddObjAttributesMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 37 + len(c.ObjID) + len(c.AttrName) // OA+ ObjID:x AttrName:x Values:[]
//TODO	for _, v := range c.Values {
//TODO		l += len(v) + 3
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c AddObjAttributesMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Values))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, v := range c.Values {
//TODO		p := AddObjAttributesMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Values),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ObjID = c.ObjID
//TODO			p.AttrName = c.AttrName
//TODO		}
//TODO		p.Values = []string{v}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c AddObjAttributesMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := AddObjAttributesMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(AddObjAttributesMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ObjID = pp.ObjID
//TODO			newp.AttrName = pp.AttrName
//TODO		}
//TODO		for _, v := range pp.Values {
//TODO			newp.Values = append(newp.Values, v)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c AddObjAttributesMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return AddObjAttributesMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ObjID:    c.ObjID,
//TODO		AttrName: c.AttrName,
//TODO	}
//TODO}

// AddObjAttributes informs peers to add a set of string values to the existing
// value of an object attribute. The attribute must be one whose value is a list
// of strings, such as StatusList.
func (c *Connection) AddObjAttributes(objID, attrName string, values []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AddObjAttributes, AddObjAttributesMessagePayload{
		ObjID:    objID,
		AttrName: attrName,
		Values:   values,
	})
}

//     _       _  _           _ __     ___
//    / \   __| |(_)_   _ ___| |\ \   / (_) _____      __
//   / _ \ / _` || | | | / __| __\ \ / /| |/ _ \ \ /\ / /
//  / ___ \ (_| || | |_| \__ \ |_ \ V / | |  __/\ V  V /
// /_/   \_\__,_|/ |\__,_|___/\__| \_/  |_|\___| \_/\_/
//             |__/

// AdjustViewMessagePayload holds the information sent by the server's AdjustView
// message. This tells the client to set its viewable area so that its x and y
// scrollbars are at the given proportion of their full range.
//
// Call the AdjustView method to send this message out to other clients.
type AdjustViewMessagePayload struct {
	BaseMessagePayload
	XView, YView float64 `json:",omitempty"`
	Grid         string  `json:",omitempty"`
}

// AdjustView tells other clients to adjust their scrollbars
// so that the x and y directions are scrolled to xview and
// yview respectively, where those values are a fraction from
// 0.0 to 1.0 indicating the proportion of the full range in
// each direction.
func (c *Connection) AdjustView(xview, yview float64) error {
	return c.AdjustViewToGridLabel(xview, yview, "")
}

// AdjustViewToGridLabel is just like AdjustView but also provides a
// grid label (e.g., A0 for the very top-left of the map) that should be
// made to be at the upper-left of the on-screen display. The xview
// and yview values are also provided for clients who cannot use the grid
// label value.
func (c *Connection) AdjustViewToGridLabel(xview, yview float64, gridLabel string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AdjustView, AdjustViewMessagePayload{
		Grid:  gridLabel,
		XView: xview,
		YView: yview,
	})
}

//     _    _ _
//    / \  | | | _____      __
//   / _ \ | | |/ _ \ \ /\ / /
//  / ___ \| | | (_) \ V  V /
// /_/   \_\_|_|\___/ \_/\_/
//

// AllowMessagePayload holds the data sent by a client when indicating
// which optional features it supports.
type AllowMessagePayload struct {
	BaseMessagePayload

	// List of supported optional feature names
	Features []string `json:",omitempty"`
}

type OptionalFeature byte

const (
	DiceColorBoxes OptionalFeature = iota
	DiceColorLabels
	GMAMarkup
)

// Allow tells the server which optional features this client is
// prepared to accept.
func (c *Connection) Allow(features ...OptionalFeature) error {
	var featureList []string
	if c.Protocol < 333 {
		return nil
	}
	for _, feature := range features {
		switch feature {
		case DiceColorBoxes:
			featureList = append(featureList, "DICE-COLOR-BOXES")
		case DiceColorLabels:
			featureList = append(featureList, "DICE-COLOR-LABELS")
		case GMAMarkup:
			featureList = append(featureList, "GMA-MARKUP")
		default:
			return fmt.Errorf("unknown OptionalFeature code %v", feature)
		}
	}
	return c.serverConn.Send(Allow, AllowMessagePayload{
		Features: featureList,
	})
}

//
//     _         _   _
//    / \  _   _| |_| |__
//   / _ \| | | | __| '_ \
//  / ___ \ |_| | |_| | | |
// /_/   \_\__,_|\__|_| |_|
//

// AuthMessagePayload holds the data sent by a client when authenticating
// to the server.
type AuthMessagePayload struct {
	BaseMessagePayload

	// Client describes the client program (e.g., "mapper 4.0.1")
	Client string `json:",omitempty"`

	// Response gives the binary response to the server's challenge
	Response []byte

	// User gives the username requested by the client. "GM" is privileged. Names beginning with "SYS$" are forbidden.
	User string `json:",omitempty"`

	// Platform gives the platform information the client's coming in on.
	// This should look like "OS version machine"
	Platform string `json:",omitempty"`
}


//  ____        _       _     _____                                     _   
// | __ )  __ _| |_ ___| |__ |  ___| __ __ _  __ _ _ __ ___   ___ _ __ | |_ 
// |  _ \ / _` | __/ __| '_ \| |_ | '__/ _` |/ _` | '_ ` _ \ / _ \ '_ \| __|
// | |_) | (_| | || (__| | | |  _|| | | (_| | (_| | | | | | |  __/ | | | |_ 
// |____/ \__,_|\__\___|_| |_|_|  |_|  \__,_|\__, |_| |_| |_|\___|_| |_|\__|
//                                           |___/                          
// 

// BatchFragmentMessagePayload holds a piece of a larger server message that was too
// large to send in one piece.
type BatchFragmentMessagePayload struct {
	BaseMessagePayload

	// unique identifier for this batch of message fragments
	ID string

	// the command of the original message (first fragment only)
	Command string `json:",omitempty"`

	// Seauence number within the batch
	Part int `json:",omitempty"`

	// Total number of parts to be sent
	Of int 

	// Signal that sending the batch is being abandoned and why
	Error string `json:",omitempty"`

	// Portion of the JSON payload from the original payload
	Data []byte `json:",omitempty"`
}

//   ____           _          _____ _ _
//  / ___|__ _  ___| |__   ___|  ___(_) | ___
// | |   / _` |/ __| '_ \ / _ \ |_  | | |/ _ \
// | |__| (_| | (__| | | |  __/  _| | | |  __/
//  \____\__,_|\___|_| |_|\___|_|   |_|_|\___|
//

// CacheFile asks other clients to be sure they retrieve
// and cache the map file with the given server ID.
func (c *Connection) CacheFile(serverID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(LoadFrom, LoadFromMessagePayload{
		FileDefinition: FileDefinition{
			File:        serverID,
			IsLocalFile: false,
		},
		CacheOnly: true,
	})
}

//
//   ____ _           _ _
//  / ___| |__   __ _| | | ___ _ __   __ _  ___
// | |   | '_ \ / _` | | |/ _ \ '_ \ / _` |/ _ \
// | |___| | | | (_| | | |  __/ | | | (_| |  __/
//  \____|_| |_|\__,_|_|_|\___|_| |_|\__, |\___|
//                                   |___/

type ChallengeMessagePayload struct {
	BaseMessagePayload
	Protocol      int
	Challenge     []byte    `json:",omitempty"`
	Iterations    int       `json:",omitempty"`
	ServerStarted time.Time `json:",omitempty"`
	ServerActive  time.Time `json:",omitempty"`
	ServerTime    time.Time `json:",omitempty"`
	ServerVersion string    `json:",omitempty"`
}

//
//   ____ _                          _            _   _
//  / ___| |__   __ _ _ __ __ _  ___| |_ ___ _ __| \ | | __ _ _ __ ___   ___
// | |   | '_ \ / _` | '__/ _` |/ __| __/ _ \ '__|  \| |/ _` | '_ ` _ \ / _ \
// | |___| | | | (_| | | | (_| | (__| ||  __/ |  | |\  | (_| | | | | | |  __/
//  \____|_| |_|\__,_|_|  \__,_|\___|\__\___|_|  |_| \_|\__,_|_| |_| |_|\___|
//
//

type CharacterNameMessagePayload struct {
	BaseMessagePayload
	NotPlaying bool     `json:",omitempty"`
	Names      []string `json:",omitempty"`
	User       string   `json:",omitempty"`
}

// CharacterName declares our character name on the VTT map.
func (c *Connection) CharacterName(actualName string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(CharacterName, CharacterNameMessagePayload{
		Names: []string{actualName},
	})
}

// CharacterNames declares our character name on the VTT map, to a slice of names.
func (c *Connection) CharacterNames(actualNames []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(CharacterName, CharacterNameMessagePayload{
		Names: actualNames,
	})
}

// IAmNotPlaying declares that this user is not playing any characters.
func (c *Connection) IAmNotPlaying() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(CharacterName, CharacterNameMessagePayload{
		NotPlaying: true,
	})
}

//   ____ _           _   __  __
//  / ___| |__   __ _| |_|  \/  | ___  ___ ___  __ _  __ _  ___
// | |   | '_ \ / _` | __| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \
// | |___| | | | (_| | |_| |  | |  __/\__ \__ \ (_| | (_| |  __/
//  \____|_| |_|\__,_|\__|_|  |_|\___||___/___/\__,_|\__, |\___|
//                                                   |___/

// ChatCommon holds fields common to chat messages and die-roll results.
type ChatCommon struct {
	// True if the peer receiving this was it origin of the message/request
	// as opposed to a peer just getting a copy of the message as it is being
	// broadcast out to everyone else.
	Origin bool `json:",omitempty"`

	// True if the message is being replayed back to the client as opposed to
	// being sent for the first time.
	Replay bool `json:",omitempty"`

	// The name of the person sending the message.
	Sender string `json:",omitempty"`

	// The names of the people the message was explicitly addressed to.
	// This will be nil for global messages.
	Recipients []string `json:",omitempty"`

	// The unique ID number for the chat message.
	MessageID int `json:",omitempty"`

	// True if this is a global message (sent to all users).
	ToAll bool `json:",omitempty"`

	// True if this message was sent only to the GM.
	ToGM bool `json:",omitempty"`

	// The date/time the message was sent
	Sent time.Time `json:",omitempty"`
}

// ChatMessageMessagePayload holds the information sent by the server's ChatMessage
// message. This is a message sent by other players or perhaps by the server itself.
//
// Call the ChatMessage, ChatMessageToAll, or ChatMessageToGM methods to send this message out to other clients.
type ChatMessageMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ChatCommon

	// True if the message contains GMA markup formatting codes
	Markup bool `json:",omitempty"`

	// True if this message should be pinned (persistently in view)
	Pin bool `json:",omitempty"`

	// The text of the chat message we received.
	Text string
}

//TODOfunc (c ChatMessageMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 138 + len(c.Sender) // TO Origin:false Replay:false Sender:x MessageID:i ToAll:false ToGM:false Sent:time[35]
//TODO	l += 16
//TODO	for _, r := range c.Recipients {
//TODO		l += len(r) + 3 // Recipients:[x]
//TODO	}
//TODO	l += 34 + len(c.Text) // Markup:false Pin:false Text:x
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c ChatMessageMessagePayload) Split() []any {
//TODO	fragments := len(c.Recipients)
//TODO	payloads := make([]any, fragments)
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i := range fragments {
//TODO		p := ChatMessageMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: fragments,
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Origin = c.Origin
//TODO			p.Replay = c.Replay
//TODO			p.Sender = c.Sender
//TODO			p.MessageID = c.MessageID
//TODO			p.ToAll = c.ToAll
//TODO			p.ToGM = c.ToGM
//TODO			p.Sent = c.Sent
//TODO			p.Markup = c.Markup
//TODO			p.Pin = c.Pin
//TODO			p.Text = c.Text
//TODO		}
//TODO		if len(c.Recipients) > i {
//TODO			p.Recipients = []string{c.Recipients[i]}
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c ChatMessageMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := ChatMessageMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(ChatMessageMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Origin = pp.Origin
//TODO			newp.Replay = pp.Replay
//TODO			newp.Sender = pp.Sender
//TODO			newp.MessageID = pp.MessageID
//TODO			newp.ToAll = pp.ToAll
//TODO			newp.ToGM = pp.ToGM
//TODO			newp.Sent = pp.Sent
//TODO			newp.Pin = pp.Pin
//TODO			newp.Markup = pp.Markup
//TODO			newp.Text = pp.Text
//TODO		}
//TODO		for _, t := range pp.Recipients {
//TODO			newp.Recipients = append(newp.Recipients, t)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c ChatMessageMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return ChatMessageMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ChatCommon: ChatCommon{
//TODO			MessageID: c.MessageID,
//TODO			Sender:    c.Sender,
//TODO		},
//TODO	}
//TODO}

// ChatMessage sends a message on the chat channel to other
// users. The to paramter is a slice of user names of the people
// who should receive this message.
func (c *Connection) ChatMessage(to []string, message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
		},
		Text: message,
	})
}

// ChatMessageToAll is equivalent to ChatMessage, but is addressed to all users.
func (c *Connection) ChatMessageToAll(message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			ToAll: true,
		},
		Text: message,
	})
}

// ChatMessageToGM is equivalent to ChatMessage, but is addressed only to the GM.
func (c *Connection) ChatMessageToGM(message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			ToGM: true,
		},
		Text: message,
	})
}

// ChatMarkupMessage sends a message on the chat channel to other
// users. The to paramter is a slice of user names of the people
// who should receive this message. The text may contain markup formatting codes
func (c *Connection) ChatMarkupMessage(to []string, message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
		},
		Markup: true,
		Text:   message,
	})
}

// ChatMessageToAll is equivalent to ChatMarkupMessage, but is addressed to all users.
func (c *Connection) ChatMarkupMessageToAll(message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			ToAll: true,
		},
		Markup: true,
		Text:   message,
	})
}

// ChatMarkupMessageToGM is equivalent to ChatMarkupMessage, but is addressed only to the GM.
func (c *Connection) ChatMarkupMessageToGM(message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			ToGM: true,
		},
		Markup: true,
		Text:   message,
	})
}

// ChatMessageWithOptions sends a chat message with all options as parameters.
// If to is an empty slice, it is equivalent to sending to all clients.
func (c *Connection) ChatMessageWithOptions(to []string, toGM, useMarkup, pin bool, message string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ChatMessage, ChatMessageMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
			ToAll:      to == nil,
			ToGM:       toGM,
		},
		Markup: useMarkup,
		Pin:    pin,
		Text:   message,
	})
}

//   ____ _
//  / ___| | ___  __ _ _ __
// | |   | |/ _ \/ _` | '__|
// | |___| |  __/ (_| | |
//  \____|_|\___|\__,_|_|
//

// ClearMessagePayload holds the information sent by the server's Clear
// message. This tells the client to remove one or more objects from its
// canvas.
//
// Call the Clear method to send this message out to other clients.
type ClearMessagePayload struct {
	BaseMessagePayload

	// The ObjID gives the object ID for the object to be removed, or one of
	// the following:
	//   *                    Remove all objects
	//   E*                   Remove all map elements
	//   M*                   Remove all monster tokens
	//   P*                   Remove all player tokens
	//   [<imagename>=]<name> Remove token with given <name>
	ObjID string
}

// Clear tells peers to remove objects from their canvases.
//
//	"*"                  Remove all objects
//	"E*"                 Remove all map elements
//	"M*"                 Remove all monster tokens
//	"P*"                 Remove all player tokens
//	[<imagename>=]<name> Remove token with given <name>
//	<id>                 Remove object with given <id>
func (c *Connection) Clear(objID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Clear, ClearMessagePayload{
		ObjID: objID,
	})
}

//   ____ _                  ____ _           _
//  / ___| | ___  __ _ _ __ / ___| |__   __ _| |_
// | |   | |/ _ \/ _` | '__| |   | '_ \ / _` | __|
// | |___| |  __/ (_| | |  | |___| | | | (_| | |_
//  \____|_|\___|\__,_|_|   \____|_| |_|\__,_|\__|
//

// ClearChatMessagePayload holds the information sent by the server's ClearChat
// message. This tells the client to remove some messages from its chat history.
//
// Call the ClearChat method to send this message out to other clients.
type ClearChatMessagePayload struct {
	BaseMessagePayload

	// User requesting the action, if known.
	RequestedBy string `json:",omitempty"`

	// Don't notify the user of the operation.
	DoSilently bool `json:",omitempty"`

	// If >0, clear all messages with IDs greater than target.
	// If <0, clear most recent -N messages.
	// If 0, clear all messages.
	Target int `json:",omitempty"`

	// Chat message ID of this notice.
	MessageID int `json:",omitempty"`
}

// ClearChat tells peers to remove all messages from their
// chat histories if target is zero. If target>0, then
// all messages with IDs greater than target are removed.
// Otherwise, if target<0 then only the most recent |target|
// messages are kept.
func (c *Connection) ClearChat(target int, silently bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ClearChat, ClearChatMessagePayload{
		DoSilently: silently,
		Target:     target,
	})
}

//   ____ _                 _____
//  / ___| | ___  __ _ _ __|  ___| __ ___  _ __ ___
// | |   | |/ _ \/ _` | '__| |_ | '__/ _ \| '_ ` _ \
// | |___| |  __/ (_| | |  |  _|| | | (_) | | | | | |
//  \____|_|\___|\__,_|_|  |_|  |_|  \___/|_| |_| |_|
//

// ClearFromMessagePayload holds the information sent by the server's ClearFrom
// message. This tells the client to remove all elements mentioned in the specified
// map file.
//
// Call the ClearFrom method to send this message out to other clients.
type ClearFromMessagePayload struct {
	BaseMessagePayload
	FileDefinition
}

// ClearFrom tells all peers to load the map file with the
// given server ID, but to remove from their canvases all
// objects described in the file rather than loading them on.
func (c *Connection) ClearFrom(serverID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(ClearFrom, ClearFromMessagePayload{
		FileDefinition: FileDefinition{
			File:        serverID,
			IsLocalFile: false,
		},
	})
}

//   ____                _           _   __  __           _
//  / ___|___  _ __ ___ | |__   __ _| |_|  \/  | ___   __| | ___
// | |   / _ \| '_ ` _ \| '_ \ / _` | __| |\/| |/ _ \ / _` |/ _ \
// | |__| (_) | | | | | | |_) | (_| | |_| |  | | (_) | (_| |  __/
//  \____\___/|_| |_| |_|_.__/ \__,_|\__|_|  |_|\___/ \__,_|\___|
//

// CombatModeMessagePayload holds the information sent by the server's CombatMode
// message. This tells the client to enter or exit combat (initiative) mode.
//
// Call the CombatMode method to send this message out to other clients.
type CombatModeMessagePayload struct {
	BaseMessagePayload

	// If true, we should be in combat mode.
	Enabled bool `json:",omitempty"`
}

// CombatMode tells all peers to enable or disable combat mode.
func (c *Connection) CombatMode(enabled bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(CombatMode, CombatModeMessagePayload{
		Enabled: enabled,
	})
}

// ToolbarMessagePayload holds the information sent by the server's Toolbar
// message. This tells the client to display or hide its toolbar.
type ToolbarMessagePayload struct {
	BaseMessagePayload
	Enabled bool `json:",omitempty"`
}

// Toolbar tells peers to turn on or off their toolbars.
func (c *Connection) Toolbar(enabled bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Toolbar, ToolbarMessagePayload{
		Enabled: enabled,
	})
}

//   ____                                     _
//  / ___|___  _ __ ___  _ __ ___   ___ _ __ | |_
// | |   / _ \| '_ ` _ \| '_ ` _ \ / _ \ '_ \| __|
// | |__| (_) | | | | | | | | | | |  __/ | | | |_
//  \____\___/|_| |_| |_|_| |_| |_|\___|_| |_|\__|
//

// CommentMessagePayload holds the information sent by the server's Comment
// message. This provides information from the server that the client is
// free to ignore, but may find interesting. Nothing sent in comments is
// critical to the operation of a client. However, some incidental bits
// of information such as an advisement of currently-supported client
// versions and progress gauge data are sent via comments.
type CommentMessagePayload struct {
	BaseMessagePayload
	Text string
}

//   ____               ____        _
//  / ___|___  _ __ ___|  _ \  __ _| |_ __ _
// | |   / _ \| '__/ _ \ | | |/ _` | __/ _` |
// | |__| (_) | | |  __/ |_| | (_| | || (_| |
//  \____\___/|_|  \___|____/ \__,_|\__\__,_|
//
// Functions related to the retrieval of core SRD data from
// the server.
//

// FilterCoreDataMessagePayload holds the request to the server to change
// player visibility of core data items.
type FilterCoreDataMessagePayload struct {
	BaseMessagePayload
	InvertSelection bool `json:",omitempty"`
	IsHidden        bool `json:",omitempty"`
	Type            string
	Filter          string
}

// FilterCoreData requests that the server change the visibility of all core database items
// of the specified type whose code matches the filter regular expression. If isHidden
// is true, those items will be visible to players; otherwise they will be hidden from
// player view.
func (c *Connection) FilterCoreData(itemType, filterRegex string, isHidden bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterCoreData, FilterCoreDataMessagePayload{
		IsHidden: isHidden,
		Type:     itemType,
		Filter:   filterRegex,
	})
}

// FilterCoreDataInverted is like FilterCoreData, but it affects all core database items
// of the given type which do NOT match the filter expression.
func (c *Connection) FilterCoreDataInverted(itemType, filterRegex string, isHidden bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterCoreData, FilterCoreDataMessagePayload{
		IsHidden:        isHidden,
		Type:            itemType,
		Filter:          filterRegex,
		InvertSelection: true,
	})
}

// QueryCoreDataMessagePayload holds the request for a core data item.
type QueryCoreDataMessagePayload struct {
	BaseMessagePayload
	Type      string
	Code      string `json:",omitempty"`
	Name      string `json:",omitempty"`
	RequestID string `json:",omitempty"`
}

// QueryCoreData asks the server to retrieve an item from the core database
// of the specified type whose name and/or code match the strings given here.
// The server will respond with an UpdateCoreData message.
func (c *Connection) QueryCoreData(itemType, code, name string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreData, QueryCoreDataMessagePayload{
		Type: itemType,
		Code: code,
		Name: name,
	})
}

// QueryCoreDataWithID is like QueryCoreData but it also sends an arbitrary ID string
// which will be returned in the server's reply.
func (c *Connection) QueryCoreDataWithID(itemType, code, name, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreData, QueryCoreDataMessagePayload{
		Type:      itemType,
		Code:      code,
		Name:      name,
		RequestID: requestID,
	})
}

// UpdateCoreDataMessagePayload contains the server response to a QueryCoreData request.
type UpdateCoreDataMessagePayload struct {
	BaseMessagePayload
	// If NoSuchEntry is true, none of the other fields should be considered valid except RequestID.
	NoSuchEntry bool   `json:",omitempty"`
	IsHidden    bool   `json:",omitempty"`
	IsLocal     bool   `json:",omitempty"`
	Code        string `json:",omitempty"`
	Name        string `json:",omitempty"`
	Type        string `json:",omitempty"`
	RequestID   string `json:",omitempty"`
}

// QueryCoreIndexMessagePayload holds the request for a core data index.
type QueryCoreIndexMessagePayload struct {
	BaseMessagePayload
	Type      string
	CodeRegex string    `json:",omitempty"`
	NameRegex string    `json:",omitempty"`
	Since     time.Time `json:",omitempty"`
	RequestID string    `json:",omitempty"`
}

// QueryCoreIndex asks the server to retrieve all the names and codes for
// a type of entry from the database.
func (c *Connection) QueryCoreIndex(itemType, codeRegex, nameRegex string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreIndex, QueryCoreIndexMessagePayload{
		Type:      itemType,
		CodeRegex: codeRegex,
		NameRegex: nameRegex,
	})
}

// QueryCoreIndexWithID is like QueryCoreIndex but also sends an arbitrary ID string
// which will be returned in the server's reply.
func (c *Connection) QueryCoreIndexWithID(itemType, codeRegex, nameRegex, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreIndex, QueryCoreIndexMessagePayload{
		Type:      itemType,
		CodeRegex: codeRegex,
		NameRegex: nameRegex,
		RequestID: requestID,
	})
}

// QueryCoreIndexSince is like QueryCoreIndex but limits the responses to those modified since a given date.
func (c *Connection) QueryCoreIndexSince(itemType, codeRegex, nameRegex string, since time.Time) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreIndex, QueryCoreIndexMessagePayload{
		Type:      itemType,
		CodeRegex: codeRegex,
		NameRegex: nameRegex,
		Since:     since,
	})
}

// QueryCoreIndexSinceWithID is like QueryCoreIndex but limits the responses to those modified since a given date,
// and sends a requestID.
func (c *Connection) QueryCoreIndexSinceWithID(itemType, codeRegex, nameRegex string, since time.Time, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryCoreIndex, QueryCoreIndexMessagePayload{
		Type:      itemType,
		CodeRegex: codeRegex,
		NameRegex: nameRegex,
		Since:     since,
		RequestID: requestID,
	})
}

// UpdateCoreIndexMessagePayload contains the server response to a QueryCoreIndex request.
type UpdateCoreIndexMessagePayload struct {
	BaseMessagePayload
	IsDone    bool   `json:",omitempty"`
	N         int    `json:",omitempty"`
	Of        int    `json:",omitempty"`
	Code      string `json:",omitempty"`
	Name      string `json:",omitempty"`
	Type      string `json:",omitempty"`
	RequestID string `json:",omitempty"`
}

//  ____             _          _
// |  _ \  ___ _ __ (_) ___  __| |
// | | | |/ _ \ '_ \| |/ _ \/ _` |
// | |_| |  __/ | | | |  __/ (_| |
// |____/ \___|_| |_|_|\___|\__,_|
//

// DeniedMessagePayload holds the reason the client was denied
// access to the server.
type DeniedMessagePayload struct {
	BaseMessagePayload
	Reason string
}

//  _____     _
// | ____|___| |__   ___
// |  _| / __| '_ \ / _ \
// | |__| (__| | | | (_) |
// |_____\___|_| |_|\___/
//

// EchoMessagePayload holds information the client wants echoed back
// to it. This is typically used to synchronize a client with a server
// by issuing a number of commands and then sending an Echo packet,
// waiting for the server to send back the echo so the client knows
// it's seen the previous messages at that point.
//
// The echo payload may contain an arbitrary boolean, integer, or
// string value named B, I, and S, respectively, for convenience in
// keeping track of the client's state or intentions behind sending
// the echo request. An arbitrary map of named values may also be
// given as the O value.
type EchoMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	B            bool           `json:"b,omitempty"`
	I            int            `json:"i,omitempty"`
	S            string         `json:"s,omitempty"`
	O            map[string]any `json:"o,omitempty"`
	ReceivedTime time.Time      `json:",omitempty"`
	SentTime     time.Time      `json:",omitempty"`
}

//TODOfunc (c EchoMessagePayload) NeedsToBeSplit() bool {
//TODO	if len(c.O) <= 1 {
//TODO		return false
//TODO	}
//TODO
//TODO	l := 8 // "ECHO {}"
//TODO	if c.B {
//TODO		l += 10 // "b": true,
//TODO	}
//TODO	if c.I != 0 {
//TODO		l += len(fmt.Sprintf("%v", c.I)) + 6 // "i": value,
//TODO	}
//TODO	if c.S != "" {
//TODO		l += len(c.S) + 6 // "s": value,
//TODO	}
//TODO	l += 8 // "o": {},
//TODO	for k, v := range c.O {
//TODO		l += len(k) + len(fmt.Sprintf("%v", v)) + 8 // "key": "value", est
//TODO	}
//TODO	return l >= MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c EchoMessagePayload) Split() []any {
//TODO	if len(c.O) < 2 {
//TODO		a := make([]any, 1)
//TODO		a[0] = c
//TODO		return a
//TODO	}
//TODO
//TODO	payloads := make([]any, len(c.O))
//TODO	gid := uuid.NewString()
//TODO	i := 0
//TODO	for k, v := range c.O {
//TODO		p := EchoMessagePayload{}
//TODO		if i == 0 {
//TODO			p.B = c.B
//TODO			p.I = c.I
//TODO			p.S = c.S
//TODO			p.ReceivedTime = c.ReceivedTime
//TODO			p.SentTime = c.SentTime
//TODO		}
//TODO		p.O = make(map[string]any, 1)
//TODO		p.O[k] = v
//TODO		p.TotalBatches = len(c.O)
//TODO		p.Batch = i
//TODO		p.BatchGroup = gid
//TODO		payloads[i] = p
//TODO		i++
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c EchoMessagePayload) Reassemble(p []any) (any, error) {
//TODO	var echo EchoMessagePayload
//TODO
//TODO	for i, d := range p {
//TODO		pkt, ok := d.(EchoMessagePayload)
//TODO		if !ok {
//TODO			return echo, fmt.Errorf("batched %T packet #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO
//TODO		if pkt.Batch != i {
//TODO			return echo, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pkt.Batch)
//TODO		}
//TODO		if pkt.TotalBatches != len(p) {
//TODO			return echo, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pkt.TotalBatches, len(p))
//TODO		}
//TODO		if i == 0 {
//TODO			echo.B = pkt.B
//TODO			echo.I = pkt.I
//TODO			echo.S = pkt.S
//TODO			echo.O = make(map[string]any, len(p))
//TODO			echo.ReceivedTime = pkt.ReceivedTime
//TODO			echo.SentTime = pkt.SentTime
//TODO		}
//TODO
//TODO		for k, v := range pkt.O {
//TODO			echo.O[k] = v
//TODO		}
//TODO	}
//TODO	return echo, nil
//TODO}
//TODO
//TODOfunc (c EchoMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return EchoMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		B: c.B,
//TODO		I: c.I,
//TODO		S: c.S,
//TODO	}
//TODO}

func (c *Connection) EchoString(s string) error {
	if c == nil {
		return fmt.Errorf("nil connection")
	}
	return c.Echo(false, 0, s, nil)
}

func (c *Connection) EchoInt(i int) error {
	if c == nil {
		return fmt.Errorf("nil connection")
	}
	return c.Echo(false, i, "", nil)
}

func (c *Connection) EchoBool(b bool) error {
	if c == nil {
		return fmt.Errorf("nil connection")
	}
	return c.Echo(b, 0, "", nil)
}

func (c *Connection) Echo(b bool, i int, s string, o map[string]any) error {
	if c == nil {
		return fmt.Errorf("nil connection")
	}
	return c.serverConn.Send(Echo, EchoMessagePayload{B: b, I: i, S: s, O: o})
}

//  _____     _ _          _
// |  ___|_ _(_) | ___  __| |
// | |_ / _` | | |/ _ \/ _` |
// |  _| (_| | | |  __/ (_| |
// |_|  \__,_|_|_|\___|\__,_|
//

// FailedMessagePayload holds information about a client request that was
// unsuccessful because the server detected an error with it or the GM decided
// to decline the request.
type FailedMessagePayload struct {
	BaseMessagePayload

	// If true, the request failed because there was something found to be wrong with it.
	IsError bool `json:",omitempty"`

	// If true, the GM simply decided not to honor the request.
	IsDiscretionary bool `json:",omitempty"`

	// The command string for the requested operation.
	Command string

	// The reason given for the failure. Usually, this is all the information you need to give to the end user.
	Reason string

	// The ID of the original request, as supplied with that request.
	RequestID string

	// The user name of the requesting user.
	RequestedBy string `json:",omitempty"`

	// An opaque connection ID meaningful to the server to identify the client from which the request was received.
	RequestingClient string `json:",omitempty"`
}

//  _____ _ _ _            ____  _          ____                     _
// |  ___(_) | |_ ___ _ __|  _ \(_) ___ ___|  _ \ _ __ ___  ___  ___| |_ ___
// | |_  | | | __/ _ \ '__| | | | |/ __/ _ \ |_) | '__/ _ \/ __|/ _ \ __/ __|
// |  _| | | | ||  __/ |  | |_| | | (_|  __/  __/| | |  __/\__ \  __/ |_\__ \
// |_|   |_|_|\__\___|_|  |____/|_|\___\___|_|   |_|  \___||___/\___|\__|___/
//

// FilterDicePresets asks the server to remove all of your
// die-roll presets whose names match the given regular
// expression.
func (c *Connection) FilterDicePresets(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterDicePresets, FilterDicePresetsMessagePayload{
		Filter: re,
	})
}

// FilterDicePresetsFor is like FilterDicePresets but works on another
// user's saved presets (GM only).
func (c *Connection) FilterDicePresetsFor(user, re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterDicePresets, FilterDicePresetsMessagePayload{
		For:    user,
		Filter: re,
	})
}

// FilterGlobalDicePresets is like FilterDicePresets but filters the system-wide global set.
func (c *Connection) FilterGlobalDicePresets(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterDicePresets, FilterDicePresetsMessagePayload{
		Global: true,
		Filter: re,
	})
}

// FilterDicePresetMessagePayload holds the filter expression
// the client sends to the server.
type FilterDicePresetsMessagePayload struct {
	BaseMessagePayload
	Global bool   `json:",omitempty"`
	Filter string `json:",omitempty"`
	For    string `json:",omitempty"`
}

//  _____ _ _ _           ___
// |  ___(_) | |_ ___ _ _|_ _|_ __ ___   __ _  __ _  ___  ___
// | |_  | | | __/ _ \ '__| || '_ ` _ \ / _` |/ _` |/ _ \/ __|
// |  _| | | | ||  __/ |  | || | | | | | (_| | (_| |  __/\__ \
// |_|   |_|_|\__\___|_| |___|_| |_| |_|\__,_|\__, |\___||___/
//                                           |___/
//

// FilterImages asks the server to remove all of your defined images that match
// a regular expression.
func (c *Connection) FilterImages(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterImages, FilterImagesMessagePayload{
		Filter: re,
	})
}

// FilterImagesExcept asks the server to remove all of your defined images that don't match
// a regular expression.
func (c *Connection) FilterImagesExcept(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterImages, FilterImagesMessagePayload{
		KeepMatching: true,
		Filter:       re,
	})
}

// FilterImagesMessagePayload holds the filter expression the client sends to the server.
type FilterImagesMessagePayload struct {
	BaseMessagePayload
	KeepMatching bool   `json:",omitempty"`
	Filter       string `json:",omitempty"`
}

//  _____ _ _ _             _             _ _
// |  ___(_) | |_ ___ _ __ / \  _   _  __| (_) ___
// | |_  | | | __/ _ \ '__/ _ \| | | |/ _` | |/ _ \
// |  _| | | | ||  __/ | / ___ \ |_| | (_| | | (_) |
// |_|   |_|_|\__\___|_|/_/   \_\__,_|\__,_|_|\___/
//

// FilterAudio asks the server to remove all of your defined audio clips that match
// a regular expression.
func (c *Connection) FilterAudio(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterAudio, FilterAudioMessagePayload{
		Filter: re,
	})
}

// FilterAudioExcept asks the server to remove all of your defined audio clips that don't match
// a regular expression.
func (c *Connection) FilterAudioExcept(re string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(FilterAudio, FilterAudioMessagePayload{
		KeepMatching: true,
		Filter:       re,
	})
}

// FilterAudioMessagePayload holds the filter expression the client sends to the server.
type FilterAudioMessagePayload struct {
	BaseMessagePayload
	KeepMatching bool   `json:",omitempty"`
	Filter       string `json:",omitempty"`
}

//
//   ____                 _           _
//  / ___|_ __ __ _ _ __ | |_ ___  __| |
// | |  _| '__/ _` | '_ \| __/ _ \/ _` |
// | |_| | | | (_| | | | | ||  __/ (_| |
//  \____|_|  \__,_|_| |_|\__\___|\__,_|
//

// GrantedMessagePayload holds the response from the server
// informing the client that its access was granted.
type GrantedMessagePayload struct {
	BaseMessagePayload
	User string
}

// .  _   _ _ _   ____       _       _      _        _                        _          _
// . | | | (_) |_|  _ \ ___ (_)_ __ | |_   / \   ___| | ___ __   _____      _| | ___  __| | __ _  ___
// . | |_| | | __| |_) / _ \| | '_ \| __| / _ \ / __| |/ / '_ \ / _ \ \ /\ / / |/ _ \/ _` |/ _` |/ _ \
// . |  _  | | |_|  __/ (_) | | | | | |_ / ___ \ (__|   <| | | | (_) \ V  V /| |  __/ (_| | (_| |  __/
// . |_| |_|_|\__|_|   \___/|_|_| |_|\__/_/   \_\___|_|\_\_| |_|\___/ \_/\_/ |_|\___|\__,_|\__, |\___|
// .                                                                                        |___/
//
// HitPointAcknowledgeMessagePayload conveys to the requesting client
// that their HitPointRequest message was accepted.
type HitPointAcknowledgeMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	RequestID        string
	RequestingClient string `json:",omitempty"`
	RequestedBy      string `json:",omitempty"`
}

//TODOfunc (c HitPointRequestMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 108 + len(c.Description) + len(c.RequestID) + len(c.RequestedBy) + len(c.RequestingClient) // HPREQ Targets[] Description:x RequestedBy:x RequestingClient:x RequestID:x Health:* TmpHP:*
//TODO	for _, t := range c.Targets {
//TODO		l += len(t) + 3
//TODO	}
//TODO	if c.Health == nil {
//TODO		l += 4
//TODO	} else {
//TODO		l += 84 + 7*5 // MaxHP:i LethalDamage:i NonLethalDamage:i AC:i FlatFootedAC:i TouchAC: CMD:i
//TODO	}
//TODO	if c.TmpHP == nil {
//TODO		l += 4
//TODO	} else {
//TODO		l += 33 + 10 + len(c.TmpHP.Expires) // TmpHP:i TmpDamage:i Expires:x
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c HitPointRequestMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Targets))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Targets {
//TODO		p := HitPointRequestMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Targets),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Description = c.Description
//TODO			p.RequestID = c.RequestID
//TODO			p.Health = c.Health
//TODO			p.TmpHP = c.TmpHP
//TODO			p.RequestedBy = c.RequestedBy
//TODO			p.RequestingClient = c.RequestingClient
//TODO		}
//TODO		p.Targets = []string{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c HitPointRequestMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := HitPointRequestMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(HitPointRequestMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Description = pp.Description
//TODO			newp.RequestID = pp.RequestID
//TODO			newp.Health = pp.Health
//TODO			newp.TmpHP = pp.TmpHP
//TODO			newp.RequestedBy = pp.RequestedBy
//TODO			newp.RequestingClient = pp.RequestingClient
//TODO		}
//TODO		for _, t := range pp.Targets {
//TODO			newp.Targets = append(newp.Targets, t)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c HitPointRequestMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return HitPointRequestMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		Description:      c.Description,
//TODO		RequestedBy:      c.RequestedBy,
//TODO		RequestID:        c.RequestID,
//TODO		RequestingClient: c.RequestingClient,
//TODO	}
//TODO}

// .  _   _ _ _   ____       _       _   ____                            _
// . | | | (_) |_|  _ \ ___ (_)_ __ | |_|  _ \ ___  __ _ _   _  ___  ___| |_
// . | |_| | | __| |_) / _ \| | '_ \| __| |_) / _ \/ _` | | | |/ _ \/ __| __|
// . |  _  | | |_|  __/ (_) | | | | | |_|  _ <  __/ (_| | |_| |  __/\__ \ |_
// . |_| |_|_|\__|_|   \___/|_|_| |_|\__|_| \_\___|\__, |\__,_|\___||___/\__|
// .                                                  |_|
//
// HitPointRequestMessagePayload requests that the GM add temporary hit points to a creature (usually a PC).
type HitPointRequestMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	// Simple description to explain the request to the GM
	Description string

	// The creature affected
	Targets []string

	// A unique identifier for this request (recommend using a UUID)
	RequestID string

	// If non-null, request updates to the creature's hit point totals
	Health *HitPointHealthRequest `json:",omitempty"`

	// If non-null, request a new block of temporary hit points
	TmpHP *HitPointTmpHPRequest `json:",omitempty"`

	// The server will fill in this information about the requesting client.
	RequestedBy      string
	RequestingClient string
}

type HitPointHealthRequest struct {
	MaxHP           int `json:",omitempty"`
	LethalDamage    int `json:",omitempty"`
	NonLethalDamage int `json:",omitempty"`
	AC              int `json:",omitempty"`
	FlatFootedAC    int `json:",omitempty"`
	TouchAC         int `json:",omitempty"`
	CMD             int `json:",omitempty"`
}

type HitPointTmpHPRequest struct {
	TmpHP     int
	TmpDamage int `json:",omitempty"`
	Expires   string
}

// HitPointRequest sends a hit point request message to the GM. If approved, the necessary updates will be entered into the system.
func (c *Connection) HitPointUpdateRequest(id, description, target string, maxHP, lethalDamage, nonLethalDamage int) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(HitPointRequest, HitPointRequestMessagePayload{
		Description: description,
		Targets:     []string{target},
		RequestID:   id,
		Health: &HitPointHealthRequest{
			MaxHP:           maxHP,
			LethalDamage:    lethalDamage,
			NonLethalDamage: nonLethalDamage,
		},
	})
}

// TemporaryHitPointRequest sends a temporary hit point request message to the GM. If approved, the necessary updates will be entered into the system.
func (c *Connection) TemporaryHitPointUpdateRequest(id, description string, targets []string, maxHP, damage int, expires string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(HitPointRequest, HitPointRequestMessagePayload{
		Description: description,
		Targets:     targets,
		RequestID:   id,
		TmpHP: &HitPointTmpHPRequest{
			TmpHP:     maxHP,
			TmpDamage: damage,
			Expires:   expires,
		},
	})
}

//  _                    _ _____
// | |    ___   __ _  __| |  ___| __ ___  _ __ ___
// | |   / _ \ / _` |/ _` | |_ | '__/ _ \| '_ ` _ \
// | |__| (_) | (_| | (_| |  _|| | | (_) | | | | | |
// |_____\___/ \__,_|\__,_|_|  |_|  \___/|_| |_| |_|
//

// LoadFromMessagePayload holds the information sent by the server's LoadFrom
// message. This tells the client to open the file named (which may either be
// a local disk file or one retrieved from the server), and either replacing their
// current canvas contents with the elements from that file, or adding those
// elements to the existing contents.
//
// Call the LoadFrom method to send this message out to other clients.
type LoadFromMessagePayload struct {
	BaseMessagePayload
	FileDefinition

	// If true, the client should only pre-load this data into a
	// local cache, but not start displaying these elements yet.
	CacheOnly bool `json:",omitempty"`

	// If true, the elements are merged with the existing map
	// contents rather than replacing them.
	Merge bool `json:",omitempty"`
}

// LoadFrom asks other clients to load a map files from a local
// disk file or from the server. The previous map contents are erased before
// each file is loaded.
//
// If local is true, a local path is specified. This is discouraged in favor
// of storing files on the server.
//
// Otherwise, the path should be the ID for the file stored on the server.
//
// If merge is true, then the current map elements are not deleted first.
// In this case, the newly-loaded elements will be merged with what is already
// on the map.
func (c *Connection) LoadFrom(path string, local bool, merge bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(LoadFrom, LoadFromMessagePayload{
		FileDefinition: FileDefinition{
			File:        path,
			IsLocalFile: local,
		},
		Merge: merge,
	})
}

//  _                    _       ___  _     _           _
// | |    ___   __ _  __| |_/\__/ _ \| |__ (_) ___  ___| |_
// | |   / _ \ / _` |/ _` \    / | | | '_ \| |/ _ \/ __| __|
// | |__| (_) | (_| | (_| /_  _\ |_| | |_) | |  __/ (__| |_
// |_____\___/ \__,_|\__,_| \/  \___/|_.__// |\___|\___|\__|
//                                       |__/
//
// This collection of types hold the message data to load
// individual map elements onto clients.
//

// LoadArcObjectMessagePayload holds the information needed to send an arc element to a map.
type LoadArcObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ArcElement
}

// LoadCircleObjectMessagePayload holds the information needed to send an ellipse element to a map.
type LoadCircleObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	CircleElement
}

// LoadLineObjectMessagePayload holds the information needed to send a line element to a map.
type LoadLineObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	LineElement
}

// LoadPolygonObjectMessagePayload holds the information needed to send a polygon element to a map.
type LoadPolygonObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	PolygonElement
}

// LoadRectangleObjectMessagePayload holds the information needed to send a rectangle element to a map.
type LoadRectangleObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	RectangleElement
}

// LoadSpellAreaOfEffectObjectMessagePayload holds the information needed to send a spell area of effect element to a map.
type LoadSpellAreaOfEffectObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	SpellAreaOfEffectElement
}

// LoadTextObjectMessagePayload holds the information needed to send a text element to a map.
type LoadTextObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	TextElement
}

// LoadTileObjectMessagePayload holds the information needed to send a graphic tile element to a map.
type LoadTileObjectMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	TileElement
}

//TODOfunc (c LoadArcObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 8 // LS-ARC
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 30 + 22 // ArcMode:i Start:f Extent:f
//TODO	l += 13      // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadArcObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadArcObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.ArcMode = c.ArcMode
//TODO			p.Start = c.Start
//TODO			p.Extent = c.Extent
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadArcObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadArcObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadArcObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.ArcMode = pp.ArcMode
//TODO			newp.Start = pp.Start
//TODO			newp.Extent = pp.Extent
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadArcObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadArcObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ArcElement: ArcElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadCircleObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 9 // LS-CIRC
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 13 // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadCircleObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadCircleObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadCircleObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadCircleObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadCircleObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadCircleObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadCircleObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		CircleElement: CircleElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadLineObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 8 // LS-LINE
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 10 // Arrow:i
//TODO	l += 13 // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadLineObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadLineObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.Arrow = c.Arrow
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadLineObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadLineObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadLineObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.Arrow = pp.Arrow
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadLineObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadLineObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		LineElement: LineElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadPolygonObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 8 // LS-POLY
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 20 // Spline:i Join:i
//TODO	l += 13 // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadPolygonObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadPolygonObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.Spline = c.Spline
//TODO			p.Join = c.Join
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadPolygonObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadPolygonObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadPolygonObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.Spline = pp.Spline
//TODO			newp.Join = pp.Join
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadPolygonObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadPolygonObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		PolygonElement: PolygonElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadRectangleObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 8 // LS-ARC
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 13 // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadRectangleObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadRectangleObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadRectangleObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadRectangleObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadRectangleObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadRectangleObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadRectangleObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		RectangleElement: RectangleElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadSpellAreaOfEffectObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 10 // LS-SAOE
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 30 + 22 // AoEShape:i
//TODO	l += 13      // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadSpellAreaOfEffectObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadSpellAreaOfEffectObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.AoEShape = c.AoEShape
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadSpellAreaOfEffectObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadSpellAreaOfEffectObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadSpellAreaOfEffectObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.AoEShape = pp.AoEShape
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadSpellAreaOfEffectObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadSpellAreaOfEffectObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		SpellAreaOfEffectElement: SpellAreaOfEffectElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadTextObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 8 // LS-ARC
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 78 + len(c.Text) + len(c.Font.Family) // Text:x Font:{Family:x Size:f Weight:i Slant:i} Anchor:i
//TODO	l += 13                                    // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadTextObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadTextObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.Anchor = c.Anchor
//TODO			p.Text = c.Text
//TODO			p.Font = c.Font
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadTextObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadTextObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadTextObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.Anchor = pp.Anchor
//TODO			newp.Text = pp.Text
//TODO			newp.Font = pp.Font
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadTextObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadTextObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		TextElement: TextElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO			Text: c.Text,
//TODO		},
//TODO	}
//TODO}
//TODO
//TODOfunc (c LoadTileObjectMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 10 // LS-TILE
//TODO	l += 121 + len(c.ID) + len(c.Stipple) + len(c.Line) + len(c.Fill) + len(c.Layer) + len(c.Group) + 45
//TODO	// ID:x X:x Y:x Z:x Stipple:x Line:x Fill:x Width:i Layer:x Level:i Goup:x Dash:i Hidden:false Locked:false
//TODO	l += 52 + len(c.Image) // Image:x BBHeight:f BBWidth:f
//TODO	l += 13                // Points:[X:f Y:f]
//TODO	l += (10 + 20) * len(c.Points)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c LoadTileObjectMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Points))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Points {
//TODO		p := LoadTileObjectMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Points),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.X = c.X
//TODO			p.Y = c.Y
//TODO			p.Z = c.Z
//TODO			p.Hidden = c.Hidden
//TODO			p.Locked = c.Locked
//TODO			p.Dash = c.Dash
//TODO			p.Width = c.Width
//TODO			p.Level = c.Level
//TODO			p.Line = c.Line
//TODO			p.Fill = c.Fill
//TODO			p.Stipple = c.Stipple
//TODO			p.Layer = c.Layer
//TODO			p.Group = c.Group
//TODO			//
//TODO			p.Image = c.Image
//TODO			p.BBHeight = c.BBHeight
//TODO			p.BBWidth = c.BBWidth
//TODO		}
//TODO		p.Points = []Coordinates{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c LoadTileObjectMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := LoadTileObjectMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(LoadTileObjectMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.X = pp.X
//TODO			newp.Y = pp.Y
//TODO			newp.Z = pp.Z
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.Locked = pp.Locked
//TODO			newp.Dash = pp.Dash
//TODO			newp.Width = pp.Width
//TODO			newp.Level = pp.Level
//TODO			newp.Line = pp.Line
//TODO			newp.Fill = pp.Fill
//TODO			newp.Stipple = pp.Stipple
//TODO			newp.Layer = pp.Layer
//TODO			newp.Group = pp.Group
//TODO			//
//TODO			newp.Image = pp.Image
//TODO			newp.BBHeight = pp.BBHeight
//TODO			newp.BBWidth = pp.BBWidth
//TODO		}
//TODO		for _, pt := range pp.Points {
//TODO			newp.Points = append(newp.Points, pt)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c LoadTileObjectMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return LoadTileObjectMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		TileElement: TileElement{
//TODO			MapElement: MapElement{
//TODO				BaseMapObject: BaseMapObject{
//TODO					ID: c.ID,
//TODO				},
//TODO				Coordinates: Coordinates{
//TODO					X: c.X,
//TODO					Y: c.Y,
//TODO				},
//TODO				Z: c.Z,
//TODO			},
//TODO			Image: c.Image,
//TODO		},
//TODO	}
//TODO}

// LoadObject sends a MapObject to all peers.
// It may be given a value of any of the supported MapObject
// types for map graphic elements (Arc, Circle, Line, Polygon,
// Rectangle, SpellAreaOfEffect, Text, or Tile).
func (c *Connection) LoadObject(mo MapObject) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	switch element := mo.(type) {
	case ArcElement:
		return c.serverConn.Send(LoadArcObject, element)
	case CircleElement:
		return c.serverConn.Send(LoadCircleObject, element)
	case LineElement:
		return c.serverConn.Send(LoadLineObject, element)
	case PolygonElement:
		return c.serverConn.Send(LoadPolygonObject, element)
	case RectangleElement:
		return c.serverConn.Send(LoadRectangleObject, element)
	case SpellAreaOfEffectElement:
		return c.serverConn.Send(LoadSpellAreaOfEffectObject, element)
	case TextElement:
		return c.serverConn.Send(LoadTextObject, element)
	case TileElement:
		return c.serverConn.Send(LoadTileObject, element)
	default:
		return fmt.Errorf("unsupported type passed to LoadObject")
	}
}

//  __  __
// |  \/  | __ _ _ __ ___ ___
// | |\/| |/ _` | '__/ __/ _ \
// | |  | | (_| | | | (_| (_) |
// |_|  |_|\__,_|_|  \___\___/
//

// MarcoMessagePayload holds the information sent by the server's Marco
// message. This is a "ping" message the server periodically sends to all
// clients to ensure they are still responding. A client who receives a
// MARCO message is expected to respond with a POLO message.
//
// If the client doesn't subscribe to Marco messages, the Dial method
// will automatically reply with Polo messages.
type MarcoMessagePayload struct {
	BaseMessagePayload
}

//  __  __            _
// |  \/  | __ _ _ __| | __
// | |\/| |/ _` | '__| |/ /
// | |  | | (_| | |  |   <
// |_|  |_|\__,_|_|  |_|\_\
//

// MarkMessagePayload holds the information sent by the server's Mark
// message. This tells the client to
// visually mark the given map coordinates.
//
// Call the Mark method to send this message out to other clients.
type MarkMessagePayload struct {
	BaseMessagePayload
	Coordinates
}

// Mark tells clients to visibly mark a location centered
// on the given (x, y) coordinates.
func (c *Connection) Mark(x, y float64) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Mark, MarkMessagePayload{
		Coordinates: Coordinates{
			X: x,
			Y: y,
		},
	})
}

//  ____  _                ____
// |  _ \| | __ _  ___ ___/ ___|  ___  _ __ ___   ___  ___  _ __   ___
// | |_) | |/ _` |/ __/ _ \___ \ / _ \| '_ ` _ \ / _ \/ _ \| '_ \ / _ \
// |  __/| | (_| | (_|  __/___) | (_) | | | | | |  __/ (_) | | | |  __/
// |_|   |_|\__,_|\___\___|____/ \___/|_| |_| |_|\___|\___/|_| |_|\___|
//

// PlaceSomeoneMessagePayload holds the information sent by the server's PlaceSomeone
// message. This tells the client to
// introduce a new creature token,
// or if that token is already on the board, update it
// with the new information (usually just moving its location).
//
// Retain any existing attributes in the original which have nil
// values here (notably, this server message never carries health
// stats so that structure will always be nil).
//
// Call the PlaceSomeone method to send this message out to other clients.
type PlaceSomeoneMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	CreatureToken
}

//TODOfunc (c PlaceSomeoneMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 100 + len(c.ID)                                                 // PS ID:x Killed:false Dim:false PolyGM:false CreatureType:i MoveMode:i Reach:i Skin:i
//TODO	l += 112 + len(c.Name) + len(c.Note) + len(c.Size) + len(c.DispSize) // Elev:i Gx:f Gy:f Name:x SkinSize:[x] Note:x Size:x DispSize:x StatusList:[x]
//TODO	for _, s := range c.SkinSize {
//TODO		l += len(s)
//TODO	}
//TODO	for _, s := range c.StatusList {
//TODO		l += len(s)
//TODO	}
//TODO	l += 73 // CustomReach:{Enabled:false Natural:i Extended:i} Targets:[x]
//TODO	for _, t := range c.Targets {
//TODO		l += len(t)
//TODO	}
//TODO	l += 23 // TargetedModifiers:{x:{Type:x Shape:x Color:x Modifiers:[x]}}
//TODO	for tk, tv := range c.TargetedModifiers {
//TODO		l += len(tk) + 6 + 41 + len(tv.Type) + len(tv.Shape) + len(tv.Color)
//TODO		for _, m := range tv.Modifiers {
//TODO			l += len(m)
//TODO		}
//TODO	}
//TODO	if c.AoE == nil {
//TODO		l += 8
//TODO	} else {
//TODO		l += 38 + len(c.AoE.Color) // Aoe*{Radius:f Color:x}
//TODO	}
//TODO	if c.Health == nil {
//TODO		l += 14
//TODO	} else {
//TODO		l += 132                          // Health:*{IsFlatFooted:false IsStable:false MaxHP:i TmpHP:i TmpDamage:i LethalDamage:i NonLethalDamage:i}
//TODO		l += 82 + len(c.Health.Condition) // Con:i HPBlur:i Condition:x AC:i FlatFootedAC:i TouchAC:i CMD:i
//TODO	}
//TODO
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c PlaceSomeoneMessagePayload) Split() []any {
//TODO	fragments := max(len(c.Targets), len(c.TargetedModifiers))
//TODO	payloads := make([]any, fragments)
//TODO	gid := uuid.NewString()
//TODO	tmods := []string{}
//TODO	for k, _ := range c.TargetedModifiers {
//TODO		tmods = append(tmods, k)
//TODO	}
//TODO
//TODO	for i := range fragments {
//TODO		p := PlaceSomeoneMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: fragments,
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ID = c.ID
//TODO			p.Killed = c.Killed
//TODO			p.Dim = c.Dim
//TODO			p.Hidden = c.Hidden
//TODO			p.PolyGM = c.PolyGM
//TODO			p.CreatureType = c.CreatureType
//TODO			p.MoveMode = c.MoveMode
//TODO			p.Reach = c.Reach
//TODO			p.Skin = c.Skin
//TODO			p.Elev = c.Elev
//TODO			p.Gx = c.Gx
//TODO			p.Gy = c.Gy
//TODO			p.Name = c.Name
//TODO			p.Health = c.Health
//TODO			p.SkinSize = c.SkinSize
//TODO			p.Color = c.Color
//TODO			p.Note = c.Note
//TODO			p.Size = c.Size
//TODO			p.DispSize = c.DispSize
//TODO			p.StatusList = c.StatusList
//TODO			p.AoE = c.AoE
//TODO			p.CustomReach = c.CustomReach
//TODO		}
//TODO		if len(c.Targets) > i {
//TODO			p.Targets = []string{c.Targets[i]}
//TODO		}
//TODO		if len(tmods) > i {
//TODO			p.TargetedModifiers = map[string]CustomConditionModifier{tmods[i]: c.TargetedModifiers[tmods[i]]}
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c PlaceSomeoneMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := PlaceSomeoneMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(PlaceSomeoneMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ID = pp.ID
//TODO			newp.Killed = pp.Killed
//TODO			newp.Dim = pp.Dim
//TODO			newp.Hidden = pp.Hidden
//TODO			newp.PolyGM = pp.PolyGM
//TODO			newp.CreatureType = pp.CreatureType
//TODO			newp.MoveMode = pp.MoveMode
//TODO			newp.Reach = pp.Reach
//TODO			newp.Skin = pp.Skin
//TODO			newp.Elev = pp.Elev
//TODO			newp.Gx = pp.Gx
//TODO			newp.Gy = pp.Gy
//TODO			newp.Name = pp.Name
//TODO			newp.Health = pp.Health
//TODO			newp.SkinSize = pp.SkinSize
//TODO			newp.Color = pp.Color
//TODO			newp.Note = pp.Note
//TODO			newp.DispSize = pp.DispSize
//TODO			newp.StatusList = pp.StatusList
//TODO			newp.AoE = pp.AoE
//TODO			newp.CustomReach = pp.CustomReach
//TODO		}
//TODO		for _, t := range pp.Targets {
//TODO			newp.Targets = append(newp.Targets, t)
//TODO		}
//TODO		for k, v := range pp.TargetedModifiers {
//TODO			newp.TargetedModifiers[k] = v
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c PlaceSomeoneMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return PlaceSomeoneMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		CreatureToken: CreatureToken{
//TODO			BaseMapObject: BaseMapObject{
//TODO				ID: c.ID,
//TODO			},
//TODO			Name: c.Name,
//TODO		},
//TODO	}
//TODO}

// PlaceSomeone tells all peers to add a new creature token on their
// maps. The parameter passed must be either a PlayerToken or MonsterToken.
//
// If the creature is already on the map, it will be replaced by the
// new one being presented here. Thus, PlaceSomeone may be used to change
// the name or location of an existing creature, although the preferred
// way to do that would be to use UpdateObjAttributes to change those
// specific attributes of the creature directly.
func (c *Connection) PlaceSomeone(someone any) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	_, isMonster := someone.(MonsterToken)
	_, isPlayer := someone.(PlayerToken)
	if !isMonster && !isPlayer {
		return fmt.Errorf("PlaceSomeone requires a MonsterToken or PlayerToken, not a %T", someone)
	}
	return c.serverConn.Send(PlaceSomeone, someone)
}

//  ____  _                _             _ _
// |  _ \| | __ _ _   _   / \  _   _  __| (_) ___
// | |_) | |/ _` | | | | / _ \| | | |/ _` | |/ _ \
// |  __/| | (_| | |_| |/ ___ \ |_| | (_| | | (_) |
// |_|   |_|\__,_|\__, /_/   \_\__,_|\__,_|_|\___/
//                |___/
//

// PlayAudioMessagePayload contains the message payload for the request to start or
// stop playing an audio clip on a client.
type PlayAudioMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	// Name is the sound clip ID as known by the mapper.
	// This may be "*" to refer to all sounds (e.g., to stop all playing sounds).
	Name string

	// If Loop is true, the sound will be played in a continuous loop.
	Loop bool `json:",omitempty"`

	// If Stop is true, stop playing the sound.
	Stop bool `json:",omitempty"`

	// If IsLocalFilt is true, Name is a local pathname on the client. Otherwise,
	// it is stored server-side and its location may be obtained via the AA server
	// command.
	IsLocalFile bool `json:",omitempty"`

	// If non-empty, Addrs lists the client addresses (as obtained from the Addr field from
	// the CONN response) of the specific clients which should play the sound clip.
	Addrs []string
}

//TODOfunc (c PlayAudioMessagePayload) NeedsToBeSplit() bool {
//TODO	// AI Animation Name Sizes :
//TODO	l := 71 + len(c.Name) // AA Name:x Loop:false Stop:false IsLocalFile:false  Addrs:[x]
//TODO	for _, a := range c.Addrs {
//TODO		l += len(a) + 3
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c PlayAudioMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Addrs))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Addrs {
//TODO		p := PlayAudioMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Addrs),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Name = c.Name
//TODO			p.Loop = c.Loop
//TODO			p.Stop = c.Stop
//TODO			p.IsLocalFile = c.IsLocalFile
//TODO		}
//TODO		p.Addrs = []string{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c PlayAudioMessagePayload) Reassemble(p []any) (any, error) {
//TODO	ai := PlayAudioMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		img, ok := d.(PlayAudioMessagePayload)
//TODO		if !ok {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if img.Batch != i {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), img.Batch)
//TODO		}
//TODO		if img.TotalBatches != len(p) {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, img.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			ai.Name = img.Name
//TODO			ai.Loop = img.Loop
//TODO			ai.Stop = img.Stop
//TODO			ai.IsLocalFile = img.IsLocalFile
//TODO		}
//TODO		for _, a := range img.Addrs {
//TODO			ai.Addrs = append(ai.Addrs, a)
//TODO		}
//TODO	}
//TODO	return ai, nil
//TODO}
//TODO
//TODOfunc (c PlayAudioMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return PlayAudioMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		Name: c.Name,
//TODO	}
//TODO}

// PlayAudio requests that clients start playing a sound.
func (c *Connection) PlayAudio(name string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(PlayAudio, PlayAudioMessagePayload{
		Name: name,
	})
}

// StopAudio requests that clients stop playing a sound.
func (c *Connection) StopAudio(name string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(PlayAudio, PlayAudioMessagePayload{
		Name: name,
		Stop: true,
	})
}

func (c *Connection) PlayAudioX(name string, loop, stop bool, addrs []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(PlayAudio, PlayAudioMessagePayload{
		Name:  name,
		Stop:  stop,
		Loop:  loop,
		Addrs: addrs,
	})
}

//  ____       _
// |  _ \ ___ | | ___
// | |_) / _ \| |/ _ \
// |  __/ (_) | | (_) |
// |_|   \___/|_|\___/
//

// Polo send the client's response to the server's MARCO ping message.
func (c *Connection) Polo() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Polo, nil)
}

type PoloMessagePayload struct {
	BaseMessagePayload
}

//  ____       _
// |  _ \ _ __(_)_   __
// | |_) | '__| \ \ / /
// |  __/| |  | |\ V /
// |_|   |_|  |_| \_/
//

type PrivMessagePayload struct {
	BaseMessagePayload
	Command string
	Reason  string
}

//   ___                        ___
//  / _ \ _   _  ___ _ __ _   _|_ _|_ __ ___   __ _  __ _  ___
// | | | | | | |/ _ \ '__| | | || || '_ ` _ \ / _` |/ _` |/ _ \
// | |_| | |_| |  __/ |  | |_| || || | | | | | (_| | (_| |  __/
//  \__\_\\__,_|\___|_|   \__, |___|_| |_| |_|\__,_|\__, |\___|
//                        |___/                     |___/

// QueryImageMessagePayload holds the information sent by the server's QueryImage
// message. This tells the client
// that a peer wants to know where to find a given
// image and the server didn't know either. If you know the definition
// for that image, reply with an AddImage message of your own.
//
// Call the QueryImage method to send this message out to other clients.
type QueryImageMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ImageDefinition
}

//TODOfunc (c QueryImageMessagePayload) NeedsToBeSplit() bool {
//TODO	// AI Animation Name Sizes :
//TODO	l := 7                 // "AI? {}"
//TODO	l += 8 + len(c.Name)   // "Name":,
//TODO	l += 11                // "Sizes": []
//TODO	l += len(c.Sizes) * 13 // "Zoom": (float)
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c QueryImageMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return QueryImageMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ImageDefinition: ImageDefinition{
//TODO			Name: c.Name,
//TODO		},
//TODO	}
//TODO}
//TODO
//TODO// Split records on Sizes (which are the instances of the images we
//TODO// have at different zoom factors, etc.)
//TODO//
//TODO// 0  Name,
//TODO// 0  Animation->{Frames,FrameSpeed,Loops}
//TODO// 0+ Sizes[File,ImageData,IsLocalFile,Zoom]
//TODOfunc (c QueryImageMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Sizes))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Sizes {
//TODO		p := QueryImageMessagePayload{
//TODO			ImageDefinition: ImageDefinition{
//TODO				Sizes: []ImageInstance{instance},
//TODO			},
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Sizes),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Name = c.Name
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c QueryImageMessagePayload) Reassemble(p []any) (any, error) {
//TODO	ai := QueryImageMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		img, ok := d.(QueryImageMessagePayload)
//TODO		if !ok {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if img.Batch != i {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), img.Batch)
//TODO		}
//TODO		if img.TotalBatches != len(p) {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, img.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			ai.Name = img.Name
//TODO			ai.Sizes = make([]ImageInstance, len(p))
//TODO		}
//TODO		ai.Sizes[i] = img.Sizes[0]
//TODO	}
//TODO	return ai, nil
//TODO}

// QueryImage asks the server and peers if anyone else knows
// where to find the data for the given image name and zoom factor.
// If someone does, you'll receive an AddImage message.
func (c *Connection) QueryImage(idef ImageDefinition) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryImage, idef)
}

//   ___                           _             _ _
//  / _ \ _   _  ___ _ __ _   _   / \  _   _  __| (_) ___
// | | | | | | |/ _ \ '__| | | | / _ \| | | |/ _` | |/ _ \
// | |_| | |_| |  __/ |  | |_| |/ ___ \ |_| | (_| | | (_) |
//  \__\_\\__,_|\___|_|   \__, /_/   \_\__,_|\__,_|_|\___/
//                        |___/
//

// QueryAudioMessagePayload holds the information sent by the server's QueryAudio
// message. This tells the client
// that a peer wants to know where to find a given
// sound file and the server didn't know either. If you know the definition
// for that audio clip, reply with an AddAudio message of your own.
//
// Call the QueryAudio method to send this message out to other clients.
type QueryAudioMessagePayload struct {
	BaseMessagePayload
	AudioDefinition
}

// QueryAudio asks the server and peers if anyone else knows
// where to find the data for the given audio clip name.
// If someone does, you'll receive an AddAudio message.
func (c *Connection) QueryAudio(idef AudioDefinition) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryAudio, idef)
}

//  ____                _
// |  _ \ ___  __ _  __| |_   _
// | |_) / _ \/ _` |/ _` | | | |
// |  _ <  __/ (_| | (_| | |_| |
// |_| \_\___|\__,_|\__,_|\__, |
//                        |___/
//

// ReadyMessagePayload indicates that the server is fully
// ready to interact with the client and all preliminary
// data has been sent to the client.
type ReadyMessagePayload struct {
	BaseMessagePayload
}

//  ____                                ___  _     _
// |  _ \ ___ _ __ ___   _____   _____ / _ \| |__ (_)
// | |_) / _ \ '_ ` _ \ / _ \ \ / / _ \ | | | '_ \| |
// |  _ <  __/ | | | | | (_) \ V /  __/ |_| | |_) | |
// |_| \_\___|_| |_| |_|\___/ \_/ \___|\___/|_.__// |
//                                              |__/
//     _   _   _        _ _           _
//    / \ | |_| |_ _ __(_) |__  _   _| |_ ___  ___
//   / _ \| __| __| '__| | '_ \| | | | __/ _ \/ __|
//  / ___ \ |_| |_| |  | | |_) | |_| | ||  __/\__ \
// /_/   \_\__|\__|_|  |_|_.__/ \__,_|\__\___||___/
//

// RemoveObjAttributesMessagePayload holds the information sent by the server's RemoveObjAttributes
// message. This tells the client
// to adjust the multi-value attribute
// of the object with the given ID by removing the listed values
// from it.
//
// Call the RemoveObjAttributes method to send this message out to other clients.
type RemoveObjAttributesMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	// The ID of the object to be modified
	ObjID string

	// The name of the attribute to modify. Must be one with a []string value.
	AttrName string

	// The values to remove from the attribute.
	Values []string
}

//TODOfunc (c RemoveObjAttributesMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 37 + len(c.ObjID) + len(c.AttrName) // OA- ObjID:x AttrName:x Values:[]
//TODO	for _, v := range c.Values {
//TODO		l += len(v) + 3
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c RemoveObjAttributesMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Values))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, v := range c.Values {
//TODO		p := RemoveObjAttributesMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Values),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ObjID = c.ObjID
//TODO			p.AttrName = c.AttrName
//TODO		}
//TODO		p.Values = []string{v}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c RemoveObjAttributesMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := RemoveObjAttributesMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(RemoveObjAttributesMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ObjID = pp.ObjID
//TODO			newp.AttrName = pp.AttrName
//TODO		}
//TODO		for _, v := range pp.Values {
//TODO			newp.Values = append(newp.Values, v)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c RemoveObjAttributesMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return RemoveObjAttributesMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ObjID:    c.ObjID,
//TODO		AttrName: c.AttrName,
//TODO	}
//TODO}

// RemoveObjAttributes informs peers to remove a set of string values from the existing
// value of an object attribute. The attribute must be one whose value is a list
// of strings, such as StatusList.
func (c *Connection) RemoveObjAttributes(objID, attrName string, values []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RemoveObjAttributes, RemoveObjAttributesMessagePayload{
		ObjID:    objID,
		AttrName: attrName,
		Values:   values,
	})
}

//  ____       _ _ ____  _
// |  _ \ ___ | | |  _ \(_) ___ ___
// | |_) / _ \| | | | | | |/ __/ _ \
// |  _ < (_) | | | |_| | | (_|  __/
// |_| \_\___/|_|_|____/|_|\___\___|
//

// RollDice sends a rollspec such as "d20+12" or "6d6 fire"
// to the server, initiating a die roll using the server's
// built-in facility for that.
//
// This will result in a response in the form of a RollResult
// message. If something went wrong when trying to satisfy
// the request, you'll receive a ChatMessage with an explanation
// instead.
//
// The to parameter lists the users who should receive the
// results of the die roll, in the same way as recipients are
// listed to the ChatMessage function.
//
// The rollspec may have any form that would be accepted to the
// dice.Roll function and dice.DieRoller.DoRoll method. See the dice package for details.
// https://pkg.go.dev/github.com/MadScienceZone/go-gma/v5/dice#DieRoller.DoRoll
//
// Added in version 5.30.0: optional list of option parameters to specify different
// options to the die rolls to avoid needless proliferation of permutations of
// methods for all the different ways we can arrange die rolls.
func (c *Connection) RollDice(to []string, rollspec string, opt ...RollDiceOption) error {
	var options dieRollOptions

	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	for _, o := range opt {
		o(&options)
	}

	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
			ToAll:      options.toGM,
			ToGM:       options.toAll,
		},
		RollSpec:  rollspec,
		RequestID: options.id,
		Targets:   options.targets,
		Type:      options.dtype,
	})
}

type dieRollOptions struct {
	id      string
	toGM    bool
	toAll   bool
	targets []string
	dtype   string
}

type RollDiceOption func(*dieRollOptions)

// WithRollType adds a die roll type designation to the roll request.
func WithRollType(dtype string) RollDiceOption {
	return func(o *dieRollOptions) {
		o.dtype = dtype
	}
}

// WithRollTargets adds creature targets to the roll request.
func WithRollTargets(targets []string) RollDiceOption {
	return func(o *dieRollOptions) {
		o.targets = targets
	}
}

// WithDieRollID adds an ID to a die roll request.
func WithDieRollID(id string) RollDiceOption {
	return func(o *dieRollOptions) {
		o.id = id
	}
}

// RollToAll causes the die roll result to go to all players
func RollToAll() RollDiceOption {
	return func(o *dieRollOptions) {
		o.toAll = true
	}
}

// RollToGM causes the die roll result to go to the GM only
func RollToGM() RollDiceOption {
	return func(o *dieRollOptions) {
		o.toGM = true
	}
}

// RollDiceWithID is identical to RollDice except it passes a user-supplied request ID
// to the server, which will be sent back with the corresponding result message(s).
func (c *Connection) RollDiceWithID(to []string, rollspec string, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
		},
		RollSpec:  rollspec,
		RequestID: requestID,
	})
}

// RollDiceMessagePayload holds the data sent from the client to the
// server when requesting a die roll.
type RollDiceMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ChatCommon

	// If you want to track the results to the requests that created them,
	// put a unique ID here. It will be repeated in the corresponding result(s).
	RequestID string `json:",omitempty"`

	// RollSpec describes the dice to be rolled and any modifiers.
	RollSpec string

	// Creatures this die roll targets
	Targets []string `json:",omitempty"`

	// Die-roll type
	Type string `json:",omitempty"`
}

// Predefined values for Type field of RollDiceMessagePayload
const (
	DTypeAttack      = "attack"
	DTypeAttackFF    = "attack-ff"
	DTypeAttackTouch = "attack-touch"
	DTypeAttackCMD   = "attack-cmd"
	DTypeDamage      = "damage"
)

//TODOfunc (c RollDiceMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 52 + len(c.RequestID) + len(c.RollSpec) + len(c.Type) // D RequestID RollSpec Targets Type {}
//TODO	l += 103 + 10 + 34 + len(c.Sender)                         // ChatCommon: "Origin":false, "Replay":false, "Sender":x, "Recipients":[], "MessageID":int, "ToAll":false, "ToGM":false, "Sent":time[34]
//TODO	for _, r := range c.Recipients {
//TODO		l += len(r)
//TODO	}
//TODO	for _, t := range c.Targets {
//TODO		l += len(t)
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODO// Split records on Sizes (which are the instances of the images we
//TODO// have at different zoom factors, etc.)
//TODO//
//TODO// 0  Name,
//TODO// 0  Animation->{Frames,FrameSpeed,Loops}
//TODO// 0+ Sizes[File,ImageData,IsLocalFile,Zoom]
//TODOfunc (c RollDiceMessagePayload) Split() []any {
//TODO	fragments := max(len(c.Targets), len(c.Recipients))
//TODO	payloads := make([]any, fragments)
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i := range fragments {
//TODO		p := RollDiceMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: fragments,
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Origin = c.Origin
//TODO			p.Replay = c.Replay
//TODO			p.Sender = c.Sender
//TODO			p.MessageID = c.MessageID
//TODO			p.ToAll = c.ToAll
//TODO			p.ToGM = c.ToGM
//TODO			p.Sent = c.Sent
//TODO			p.RequestID = c.RequestID
//TODO			p.RollSpec = c.RollSpec
//TODO			p.Type = c.Type
//TODO		}
//TODO		if len(c.Recipients) > i {
//TODO			p.Recipients = []string{c.Recipients[i]}
//TODO		}
//TODO		if len(c.Targets) > i {
//TODO			p.Targets = []string{c.Targets[i]}
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c RollDiceMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := RollDiceMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(RollDiceMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Origin = pp.Origin
//TODO			newp.Replay = pp.Replay
//TODO			newp.Sender = pp.Sender
//TODO			newp.MessageID = pp.MessageID
//TODO			newp.ToAll = pp.ToAll
//TODO			newp.ToGM = pp.ToGM
//TODO			newp.RequestID = pp.RequestID
//TODO			newp.RollSpec = pp.RollSpec
//TODO			newp.Type = pp.Type
//TODO		}
//TODO		for _, s := range pp.Targets {
//TODO			newp.Targets = append(newp.Targets, s)
//TODO		}
//TODO		for _, s := range pp.Recipients {
//TODO			newp.Recipients = append(newp.Recipients, s)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c RollDiceMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return RollDiceMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ChatCommon: ChatCommon{
//TODO			MessageID: c.MessageID,
//TODO		},
//TODO		RequestID: c.RequestID,
//TODO		RollSpec:  c.RollSpec,
//TODO		Type:      c.Type,
//TODO	}
//TODO}

// RollDiceToAll is equivalent to RollDice, sending the results to all users.
func (c *Connection) RollDiceToAll(rollspec string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			ToAll: true,
		},
		RollSpec: rollspec,
	})
}

// RollDiceToAllWithID is equivalent to RollDiceWithID, sending the results to all users.
func (c *Connection) RollDiceToAllWithID(rollspec, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			ToAll: true,
		},
		RollSpec:  rollspec,
		RequestID: requestID,
	})
}

// RollDiceToGM is equivalent to RollDice, sending the results only to the GM.
func (c *Connection) RollDiceToGM(rollspec string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			ToGM: true,
		},
		RollSpec: rollspec,
	})
}

// RollDiceToGMWithID is equivalent to RollDiceWithID, sending the results only to the GM.
func (c *Connection) RollDiceToGMWithID(rollspec, requestID string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			ToGM: true,
		},
		RollSpec:  rollspec,
		RequestID: requestID,
	})
}

// RollResultMessagePayload holds the information sent by the server's RollResult
// message. This tells the client the results of a die roll.
type RollResultMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	ChatCommon

	// True if there will be more results following this one for the same request
	MoreResults bool `json:",omitempty"`

	// The ID string passed by the user to associate this result with their request (may be blank)
	RequestID string `json:",omitempty"`

	// The title describing the purpose of the die-roll, as set by the user.
	Title string `json:",omitempty"`

	// The die roll result and details behind where it came from.
	Result dice.StructuredResult

	// The creature this roll targeted
	Targets []string `json:",omitempty"`

	// The die-roll type
	Type string `json:",omitempty"`
}

//TODOfunc (c RollResultMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 138 + len(c.Sender) // ROLL Origin:false Replay:false Sender:x MessageID:i ToAll:false ToGM:false Sent:time[35]
//TODO	l += 16
//TODO	for _, r := range c.Recipients {
//TODO		l += len(r) + 3 // Recipients:[x]
//TODO	}
//TODO	l += 42 + len(c.RequestID) + len(c.Title) // MoreResults:false RequestID:x Title:x
//TODO	l += 84                                   // Result:{ResultSuppressed:false InvalidRequest:false Result:i Details:[]{Type:x Value:x}
//TODO	for _, r := range c.Result.Details {
//TODO		l += 22 + len(r.Type) + len(r.Value)
//TODO	}
//TODO	l += 14 // Targets[x]
//TODO	for _, r := range c.Targets {
//TODO		l += len(r) + 3
//TODO	}
//TODO	l += 8 + len(c.Type) // Type:x
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c RollResultMessagePayload) Split() []any {
//TODO	fragments := max(len(c.Targets), len(c.Result.Details), len(c.Recipients))
//TODO	payloads := make([]any, fragments)
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i := range fragments {
//TODO		p := RollResultMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: fragments,
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Origin = c.Origin
//TODO			p.Replay = c.Replay
//TODO			p.Sender = c.Sender
//TODO			p.MessageID = c.MessageID
//TODO			p.ToAll = c.ToAll
//TODO			p.ToGM = c.ToGM
//TODO			p.Sent = c.Sent
//TODO			p.MoreResults = c.MoreResults
//TODO			p.RequestID = c.RequestID
//TODO			p.Title = c.Title
//TODO			p.Type = c.Type
//TODO			p.Result.ResultSuppressed = c.Result.ResultSuppressed
//TODO			p.Result.InvalidRequest = c.Result.InvalidRequest
//TODO			p.Result.Result = c.Result.Result
//TODO		}
//TODO		if len(c.Result.Details) > i {
//TODO			p.Result.Details = dice.StructuredDescriptionSet([]dice.StructuredDescription{c.Result.Details[i]})
//TODO		}
//TODO		if len(c.Recipients) > i {
//TODO			p.Recipients = []string{c.Recipients[i]}
//TODO		}
//TODO		if len(c.Targets) > i {
//TODO			p.Targets = []string{c.Targets[i]}
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c RollResultMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := RollResultMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(RollResultMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Origin = pp.Origin
//TODO			newp.Replay = pp.Replay
//TODO			newp.Sender = pp.Sender
//TODO			newp.MessageID = pp.MessageID
//TODO			newp.ToAll = pp.ToAll
//TODO			newp.ToGM = pp.ToGM
//TODO			newp.Sent = pp.Sent
//TODO			newp.MoreResults = pp.MoreResults
//TODO			newp.RequestID = pp.RequestID
//TODO			newp.Title = pp.Title
//TODO			newp.Type = pp.Type
//TODO			newp.Result.ResultSuppressed = pp.Result.ResultSuppressed
//TODO			newp.Result.InvalidRequest = pp.Result.InvalidRequest
//TODO			newp.Result.Result = pp.Result.Result
//TODO		}
//TODO		for _, t := range pp.Result.Details {
//TODO			newp.Result.Details = append(newp.Result.Details, t)
//TODO		}
//TODO		for _, t := range pp.Recipients {
//TODO			newp.Recipients = append(newp.Recipients, t)
//TODO		}
//TODO		for _, t := range pp.Targets {
//TODO			newp.Targets = append(newp.Targets, t)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c RollResultMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return RollResultMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		RequestID: c.RequestID,
//TODO		ChatCommon: ChatCommon{
//TODO			MessageID: c.MessageID,
//TODO			Sender:    c.Sender,
//TODO		},
//TODO	}
//TODO}

//  ____  _          ____                     _
// |  _ \(_) ___ ___|  _ \ _ __ ___  ___  ___| |_ ___
// | | | | |/ __/ _ \ |_) | '__/ _ \/ __|/ _ \ __/ __|
// | |_| | | (_|  __/  __/| | |  __/\__ \  __/ |_\__ \
// |____/|_|\___\___|_|   |_|  \___||___/\___|\__|___/
//

// DefineDicePresets replaces any existing die-roll presets you have
// stored on the server with the new set passed as the presets parameter.
func (c *Connection) DefineDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresets, DefineDicePresetsMessagePayload{
		Presets: presets,
	})
}

// DefineGlobalDicePresets replaces any existing die-roll presets you have
// stored on the server with the new set passed as the presets parameter, but for the system-wide global set.
func (c *Connection) DefineGlobalDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresets, DefineDicePresetsMessagePayload{
		Global:  true,
		Presets: presets,
	})
}

// DefineDicePresetDelegates changes the current list of users allowed to view and
// change a user's stored presets. The new list replaces any and all previous ones.
func (c *Connection) DefineDicePresetDelegates(delegates []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresetDelegates, DefineDicePresetDelegatesMessagePayload{
		Delegates: delegates,
	})
}

// DefineDicePresetDelegatesFor is just like DefineDicePresetDelegates
// but performs the operation for another user (GM only).
func (c *Connection) DefineDicePresetDelegatesFor(user string, delegates []string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresetDelegates, DefineDicePresetDelegatesMessagePayload{
		For:       user,
		Delegates: delegates,
	})
}

// DefineDicePresetsFor is just like DefineDicePresets but performs the operation
// for another user (GM only).
func (c *Connection) DefineDicePresetsFor(user string, presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresets, DefineDicePresetsMessagePayload{
		For:     user,
		Presets: presets,
	})
}

type DefineDicePresetsMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	Global  bool                 `json:",omitempty"`
	For     string               `json:",omitempty"`
	Presets []dice.DieRollPreset `json:",omitempty"`
}

type DefineDicePresetDelegatesMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	For       string   `json:",omitempty"`
	Delegates []string `json:",omitempty"`
}

//TODOfunc (c DefineDicePresetDelegatesMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 27 + len(c.For) // DDD For:x Delegates[]
//TODO	for _, p := range c.Delegates {
//TODO		l += len(p)
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c DefineDicePresetDelegatesMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Delegates))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Delegates {
//TODO		p := DefineDicePresetDelegatesMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Delegates),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.For = c.For
//TODO		}
//TODO		p.Delegates = []string{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c DefineDicePresetDelegatesMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := DefineDicePresetDelegatesMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(DefineDicePresetDelegatesMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.For = pp.For
//TODO		}
//TODO		for _, pre := range pp.Delegates {
//TODO			newp.Delegates = append(newp.Delegates, pre)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c DefineDicePresetDelegatesMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return DefineDicePresetDelegatesMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		For: c.For,
//TODO	}
//TODO}
//TODOfunc (c DefineDicePresetsMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 40 + len(c.For) // DD For:x Global:false Presets:[]
//TODO	for _, p := range c.Presets {
//TODO		l += 53 + len(p.Name) + len(p.Description) + len(p.DieRollSpec) // Global:false, Name:x, Description:x, DieRollSpec:x
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c DefineDicePresetsMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Presets))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Presets {
//TODO		p := DefineDicePresetsMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Presets),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Global = c.Global
//TODO			p.For = c.For
//TODO		}
//TODO		p.Presets = []dice.DieRollPreset{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c DefineDicePresetsMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := DefineDicePresetsMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(DefineDicePresetsMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Global = pp.Global
//TODO			newp.For = pp.For
//TODO		}
//TODO		for _, pre := range pp.Presets {
//TODO			newp.Presets = append(newp.Presets, pre)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c DefineDicePresetsMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return DefineDicePresetsMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		For:    c.For,
//TODO		Global: c.Global,
//TODO	}
//TODO}
//TODO
//TODOfunc (c AddDicePresetsMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 40 + len(c.For) // DD For:x Global:false Presets:[]
//TODO	for _, p := range c.Presets {
//TODO		l += 53 + len(p.Name) + len(p.Description) + len(p.DieRollSpec) // Global:false, Name:x, Description:x, DieRollSpec:x
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c AddDicePresetsMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Presets))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Presets {
//TODO		p := AddDicePresetsMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Presets),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Global = c.Global
//TODO			p.For = c.For
//TODO		}
//TODO		p.Presets = []dice.DieRollPreset{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c AddDicePresetsMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := AddDicePresetsMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(AddDicePresetsMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Global = pp.Global
//TODO			newp.For = pp.For
//TODO		}
//TODO		for _, pre := range pp.Presets {
//TODO			newp.Presets = append(newp.Presets, pre)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c AddDicePresetsMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return AddDicePresetsMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		For:    c.For,
//TODO		Global: c.Global,
//TODO	}
//TODO}

// AddDicePresets is like DefineDicePresets except that it adds the presets
// passed in to the existing set rather than replacing them.
func (c *Connection) AddDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AddDicePresets, AddDicePresetsMessagePayload{
		Presets: presets,
	})
}

// AddGlobalDicePresets is like DefineGlobalDicePresets except that it adds the presets
// passed in to the existing set rather than replacing them.
func (c *Connection) AddGlobalDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AddDicePresets, AddDicePresetsMessagePayload{
		Global:  true,
		Presets: presets,
	})
}

// AddDicePresetsFor is just like AddDicePresets but performs the operation
// for another user (GM only).
func (c *Connection) AddDicePresetsFor(user string, presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AddDicePresets, AddDicePresetsMessagePayload{
		For:     user,
		Presets: presets,
	})
}

type AddDicePresetsMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	Global  bool                 `json:",omitempty"`
	For     string               `json:",omitempty"`
	Presets []dice.DieRollPreset `json:",omitempty"`
}

// QueryDicePresets requests that the server send you the die-roll
// presets currently stored for you. It will send you an UpdateDicePresets
// message.
func (c *Connection) QueryDicePresets() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryDicePresets, nil)
}

// QueryGlobalDicePresets is like QueryDicePresets but queries only the system-wide set.
func (c *Connection) QueryGlobalDicePresets() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryDicePresets, QueryDicePresetsMessagePayload{Global: true})
}

// QueryDicePresetsFor is like QueryDicePresets but queries presets for a given user.
func (c *Connection) QueryDicePresetsFor(user string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryDicePresets, QueryDicePresetsMessagePayload{For: user})
}

type QueryDicePresetsMessagePayload struct {
	BaseMessagePayload
	Global bool   `json:",omitempty"`
	For    string `json:",omitempty"`
}

// UpdateClockMessagePayload holds the information sent by the server's UpdateClock
// message. This tells the client to update its clock display to the new value.
type UpdateClockMessagePayload struct {
	BaseMessagePayload

	// The clock is now at the given absolute number of
	// seconds from the GMA clock's global epoch.
	Absolute int64

	// The elapsed time counter is now this many seconds from
	// some reference point set by the GM (often the start of
	// combat).
	Relative int64

	// If true and not in combat mode, local clients should
	// keep running the clock in real time.
	Running bool `json:",omitempty"`
}

// UpdateClock informs everyone of the current time
func (c *Connection) UpdateClock(absolute, relative int64, keepRunning bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateClock, UpdateClockMessagePayload{
		Absolute: absolute,
		Relative: relative,
		Running:  keepRunning,
	})
}

// UpdateDicePresetsMessagePayload holds the information sent by the server's UpdateDicePresets
// message. This tells the client to
// accept the die-roll presets
// described here, replacing any previous presets it was
// using.
type UpdateDicePresetsMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	Global      bool `json:",omitempty"`
	Presets     []dice.DieRollPreset
	For         string   `json:",omitempty"`
	DelegateFor []string `json:",omitempty"`
	Delegates   []string `json:",omitempty"`
}

//TODOfunc (c UpdateDicePresetsMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 40 + len(c.For) // DD For:x Global:false Presets:[]
//TODO	l += 34              // DelegateFor:[] Delegates:[]
//TODO	for _, p := range c.DelegateFor {
//TODO		l += len(p)
//TODO	}
//TODO	for _, p := range c.Delegates {
//TODO		l += len(p)
//TODO	}
//TODO	for _, p := range c.Presets {
//TODO		l += 53 + len(p.Name) + len(p.Description) + len(p.DieRollSpec) // Global:false, Name:x, Description:x, DieRollSpec:x
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c UpdateDicePresetsMessagePayload) Split() []any {
//TODO	fragments := max(len(c.Presets), len(c.DelegateFor), len(c.Delegates))
//TODO	payloads := make([]any, fragments)
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i := range fragments {
//TODO		p := UpdateDicePresetsMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: fragments,
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.Global = c.Global
//TODO			p.For = c.For
//TODO		}
//TODO		if len(c.Presets) > i {
//TODO			p.Presets = []dice.DieRollPreset{c.Presets[i]}
//TODO		}
//TODO		if len(c.Delegates) > i {
//TODO			p.Delegates = []string{c.Delegates[i]}
//TODO		}
//TODO		if len(c.DelegateFor) > i {
//TODO			p.DelegateFor = []string{c.DelegateFor[i]}
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c UpdateDicePresetsMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := UpdateDicePresetsMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(UpdateDicePresetsMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.Global = pp.Global
//TODO			newp.For = pp.For
//TODO		}
//TODO		for _, pre := range pp.Presets {
//TODO			newp.Presets = append(newp.Presets, pre)
//TODO		}
//TODO		for _, pre := range pp.Delegates {
//TODO			newp.Delegates = append(newp.Delegates, pre)
//TODO		}
//TODO		for _, pre := range pp.DelegateFor {
//TODO			newp.DelegateFor = append(newp.DelegateFor, pre)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c UpdateDicePresetsMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return AddDicePresetsMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		For:    c.For,
//TODO		Global: c.Global,
//TODO	}
//TODO}

// UpdateInitiativeMessagePayload holds the information sent by the server's UpdateInitiative
// message. This tells the client that the initiative order has been changed. Its current
// notion of the initiative order should be replaced by the one given here.
type UpdateInitiativeMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	InitiativeList []InitiativeSlot
}

//TODOfunc (c UpdateInitiativeMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 24 // IL InitiativeList[]
//TODO	for _, p := range c.InitiativeList {
//TODO		l += 82 + len(p.Name) // Slot:i Name:x IsHolding:false HasReadiedAction:false IsFlatFooted:false CurrentHP:i
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c UpdateInitiativeMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.InitiativeList))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.InitiativeList {
//TODO		p := UpdateInitiativeMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.InitiativeList),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO			InitiativeList: []InitiativeSlot{instance},
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c UpdateInitiativeMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := UpdateInitiativeMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(UpdateInitiativeMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO		for _, l := range pp.InitiativeList {
//TODO			newp.InitiativeList = append(newp.InitiativeList, l)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c UpdateInitiativeMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return UpdateInitiativeMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO	}
//TODO}

// UpdateInitiative informs our peers of a change to the
// inititive order.
func (c *Connection) UpdateInitiative(ilist []InitiativeSlot) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateInitiative, UpdateInitiativeMessagePayload{
		InitiativeList: ilist,
	})
}

// InitiativeSlot describes the creature occupying a given
// slot of the initiative list.
type InitiativeSlot struct {
	// The slot number (currently 0–59, corresponding to the 1/10th second "count" in the initiative round)
	Slot int

	// Deprecated: The current hit point total for the creature.
	// Set creature hit points via the OA command instead.
	CurrentHP int

	// The creature's name as displayed on the map.
	Name string

	// If true, the creature is holding their action.
	IsHolding bool `json:",omitempty"`

	// If true, the creature has a readied action.
	HasReadiedAction bool `json:",omitempty"`

	// It true, the creature is flat-footed.
	IsFlatFooted bool `json:",omitempty"`
}

// UpdateObjAttributesMessagePayload holds the information sent by the server's UpdateObjAttributes
// message. This tells the client to update an existing object
// with new attributes. Any attributes not listed here should
// remain intact.
//
// Call the UpdateObjAttributes method to send this message out to other clients.
type UpdateObjAttributesMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	// The ID of the object to be modified.
	ObjID string

	// A map of attribute name to its new value.
	NewAttrs map[string]any
}

//TODOfunc (c UpdateObjAttributesMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 27 + len(c.ObjID) // OA ObjID:x NewAttrs:{}
//TODO	for k, v := range c.NewAttrs {
//TODO		l += len(k) + len(fmt.Sprintf("%v", v)) + 8
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c UpdateObjAttributesMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.NewAttrs))
//TODO	gid := uuid.NewString()
//TODO
//TODO	i := 0
//TODO	for k, v := range c.NewAttrs {
//TODO		p := UpdateObjAttributesMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.NewAttrs),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ObjID = c.ObjID
//TODO		}
//TODO		p.NewAttrs = map[string]any{k: v}
//TODO		payloads[i] = p
//TODO		i++
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c UpdateObjAttributesMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := UpdateObjAttributesMessagePayload{}
//TODO	newp.NewAttrs = make(map[string]any, len(p))
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(UpdateObjAttributesMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ObjID = pp.ObjID
//TODO		}
//TODO		for k, v := range pp.NewAttrs {
//TODO			newp.NewAttrs[k] = v
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c UpdateObjAttributesMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return UpdateObjAttributesMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		ObjID: c.ObjID,
//TODO	}
//TODO}

// UpdateObjAttributes informs peers that they should modify the
// specified object's attributes which are mentioned in the newAttrs
// map. This maps attribute names to their new values.
func (c *Connection) UpdateObjAttributes(objID string, newAttrs map[string]any) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateObjAttributes, UpdateObjAttributesMessagePayload{
		ObjID:    objID,
		NewAttrs: newAttrs,
	})
}

// UpdatePeerListMessagePayload holds the information sent by the server's UpdatePeerList
// message. This tells the client that the list of
// other connected peers has changed.
type UpdatePeerListMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	PeerList []Peer
}

// Peer describes each peer we can reach via our server connection.
type Peer struct {
	// IP address and port of the peer
	Addr string

	// The username provided by the peer when it authenticated
	User string

	// The peer is not actually playing a character at all.
	NotPlaying bool `json:",omitempty"`

	// The list of creatures controlled if not the same as User.
	AKA []string `json:",omitempty"`

	// A description of the peer client program (provided by that client)
	Client string `json:",omitempty"`

	// How many seconds ago the peer last answered a "still alive?" ping from the server
	LastPolo float64

	// True if the client authenticated successfully
	IsAuthenticated bool `json:",omitempty"`

	// True if this structure describes the connection of this client program
	IsMe bool `json:",omitempty"`
}

func (c UpdatePeerListMessagePayload) NeedsToBeSplit() bool {
	// AI Animation Name Sizes :
	l := 21 // CONN {"PeerList": []}"
	for _, p := range c.PeerList {
		l += 8 + len(p.Addr) // "Addr":,
		l += 8 + len(p.User) // "User":,
		l += 19              // "NotPlaying":false,
		l += 9               // "AKA":[],
		for _, aka := range p.AKA {
			l += len(aka)
		}
		l += 23 // "LastPolo": (float)
		l += 24 // "IsAuthenticated":false,
		l += 8  // "IsMe":false,
	}
	return l > MaxServerMessageSize
}

// Split records on Sizes (which are the instances of the images we
// have at different zoom factors, etc.)
//
// 0  Name,
// 0  Animation->{Frames,FrameSpeed,Loops}
// 0+ Sizes[File,ImageData,IsLocalFile,Zoom]
//TODOfunc (c UpdatePeerListMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.PeerList))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.PeerList {
//TODO		p := UpdatePeerListMessagePayload{
//TODO			PeerList: []Peer{instance},
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.PeerList),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c UpdatePeerListMessagePayload) Reassemble(p []any) (any, error) {
//TODO	ai := UpdatePeerListMessagePayload{
//TODO		PeerList: make([]Peer, len(p)),
//TODO	}
//TODO
//TODO	for i, d := range p {
//TODO		pr, ok := d.(UpdatePeerListMessagePayload)
//TODO		if !ok {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pr.Batch != i {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pr.Batch)
//TODO		}
//TODO		if pr.TotalBatches != len(p) {
//TODO			return ai, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pr.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		ai.PeerList[i] = pr.PeerList[0]
//TODO	}
//TODO	return ai, nil
//TODO}
//TODOfunc (c UpdatePeerListMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return UpdatePeerListMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO	}
//TODO}

// QueryPeers asks the server to send an UpdatePeerList
// message with the current set of peers who are connected
// to the server.
func (c *Connection) QueryPeers() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryPeers, nil)
}

type QueryPeersMessagePayload struct {
	BaseMessagePayload
}

// UpdateProgressMessagePayload holds the information sent by the server's UpdateProgress
// Comment notification. This
// advises the client of the status of an operation
// in progress. The client may wish to display a progress indicator to the
// user.
type UpdateProgressMessagePayload struct {
	BaseMessagePayload

	// If true, we can dispose of the tracked operation
	// and should not expect further updates about it.
	IsDone bool `json:",omitempty"`

	// If true, this is not an operation progress meter, but is a view of
	// the status of a running timer.
	IsTimer bool `json:",omitempty"`

	// The current progress toward MaxValue.
	Value int

	// The maximum expected value for the progress indication.
	// If this is 0, we don't yet know what the maximum will be.
	// Note that this may change from one message to another, if
	// the server realizes its previous estimate was incorrect.
	MaxValue int `json:",omitempty"`

	// Unique identifier for the operation we're tracking
	OperationID string

	// Description of the operation in progress, suitable for display.
	Title string `json:",omitempty"`

	// If this relates to specific characters, their names will be listed here.
	Targets []string `json:",omitempty"`
}

// UpdateStatusMarkerMessagePayload holds the information sent by the server's UpdateStatusMarker
// message. This tells the client
// to add or change a status marker which may be placed
// on creature tokens.
//
// Note: the server usually sends these upon login, which the Connection
// struct stores internally.
type UpdateStatusMarkerMessagePayload struct {
	BaseMessagePayload
	StatusMarkerDefinition
}

// UpdateStatusMarker changes, removes, or adds a status marker to place on
// a creature marker.
func (c *Connection) UpdateStatusMarker(smd StatusMarkerDefinition) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateStatusMarker, smd)
}

// StatusMarkerDefinition describes each creature token status
// that the map clients indicate.
type StatusMarkerDefinition struct {
	// If the token should be transparent when this condition is in effect
	Transparent bool `json:",omitempty"`

	// The name of the condition
	Condition string

	// The shape of the marker to be drawn on the token. This may
	// be one of the following:
	//   |v  small downward-pointing triangle against the left edge
	//   v|  small downward-pointing triangle against the right edge
	//   |o  small circle against the left edge
	//   o|  small circle against the right edge
	//   |<> small diamond against the left edge
	//   <>| small diamond against the right edge
	//   /   slash from upper right to lower left
	//   \   slash from upper left to lower right
	//   //  double slash from upper right to lower left
	//   \\  double slash from upper left to lower right
	//   -   horizontal line through the center of the token
	//   =   double horizontal line through the center
	//   |   vertical line through the center
	//   ||  double vertical line through the center
	//   +   cross (vertical and horizontal lines) through center
	//   #   double cross (vertical and horizontal lines) through center
	//   V   large downward-pointing triangle around token
	//   ^   large upward-pointing triangle around token
	//   <>  large diamond around token
	//   O   large circle around token
	Shape string

	// The color to draw the marker. If the name begins with a pair
	// of hyphens (e.g., "--red") then the marker is drawn with long
	// dashed lines. If it begins with dots (e.g., "..blue") it is
	// drawn with short dashed lines.
	//
	// The special color "*" may be used to indicate that the marker
	// should be drawn in the same color as the creature's threat zone.
	Color string

	// A player-readable description of the effect the condition has on
	// the affected creature.
	Description string `json:",omitempty"`
}

// Text produces a simple text description of a StatusMarkerDefinition structure.
func (c StatusMarkerDefinition) Text() string {
	return fmt.Sprintf("Condition %q: Shape=%q, Color=%q, Description=%q, Transparent=%v", c.Condition, c.Shape, c.Color, c.Description, c.Transparent)
}

// StatusMarkerDefinitions is a map of a condition code name to the full
// description of the marker to use for that condition.
type StatusMarkerDefinitions map[string]StatusMarkerDefinition

// CharacterDefinitions is a map of a character name to their token object.
type CharacterDefinitions map[string]PlayerToken

// Text produces a simple text description of a map of PlayerTokens
func (cs CharacterDefinitions) Text() string {
	var s strings.Builder
	for k, c := range cs {
		fmt.Fprintf(&s, "[%s] %v\n", k, c)
	}
	return s.String()
}

// Text produces a simple text description of a map of StatusMarkerDefinitions
// as a multi-line string.
func (cs StatusMarkerDefinitions) Text() string {
	var s strings.Builder
	for k, c := range cs {
		fmt.Fprintf(&s, "[%s] %s\n", k, c.Text())
	}
	return s.String()
}

// UpdateTurnMessagePayload holds the information sent by the server's UpdateTurn
// message. This tells the client whose turn it is in combat.
type UpdateTurnMessagePayload struct {
	BaseMessagePayload

	// The ObjID of the creature whose turn it is. This may also be one of:
	//   *Monsters*   All monsters are up now.
	//   (empty)      It is no one's turn now.
	//   /regex       All creatures whose names match regex
	ActorID string

	// The time lapsed so far since the start of combat.
	// Count is the initiative slot within the round.
	Hours, Minutes, Seconds, Rounds, Count int
}

// UpdateTurn advances the initiative turn clock for connected clients.
func (c *Connection) UpdateTurn(relative float64, actor string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateTurn, UpdateTurnMessagePayload{
		ActorID: actor,
		// hours, minutes, seconds since start of combat
		Hours:   int(relative) / 3600,
		Minutes: (int(relative) / 60) % 60,
		Seconds: int(relative) % 60,
		// total rounds since start of combat
		Rounds: int(relative) / 6,
		// initiative count since start of round
		Count: int(relative*10) % 60,
	})
}

// Sync requests that the server send the entire game state
// to it.
func (c *Connection) Sync() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Sync, nil)
}

type SyncMessagePayload struct {
	BaseMessagePayload
}

// SyncChat requests that the server (re-)send past messages
// greater than the target message ID (target≥0) or the most
// recent |target| messages (target<0).
func (c *Connection) SyncChat(target int) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(SyncChat, SyncChatMessagePayload{
		Target: target,
	})
}

type SyncChatMessagePayload struct {
	BaseMessagePayload
	Target int `json:",omitempty"`
}

type UpdateVersionsMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload
	Packages []PackageUpdate `json:",omitempty"`
}

//TODOfunc (c UpdateVersionsMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 25 // UPDATES Packages[]
//TODO	for _, p := range c.Packages {
//TODO		l += 57 + len(p.Name) + len(p.VersionPattern) + len(p.MinimumVersion) // Name:x VersionPattern:x MinimumVersion:x Instances[OS:x Arch:x Version:x Token:x]
//TODO		for _, i := range p.Instances {
//TODO			l += 34 + len(i.OS) + len(i.Arch) + len(i.Version) + len(i.Token)
//TODO		}
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c UpdateVersionsMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Packages))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Packages {
//TODO		p := UpdateVersionsMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Packages),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		p.Packages = []PackageUpdate{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c UpdateVersionsMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := UpdateVersionsMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(UpdateVersionsMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO		for _, pkg := range pp.Packages {
//TODO			newp.Packages = append(newp.Packages, pkg)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c UpdateVersionsMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return UpdateVersionsMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO	}
//TODO}

type PackageUpdate struct {
	Name           string
	VersionPattern string         `json:",omitempty"`
	VersionRegex   *regexp.Regexp `json:"-"`
	MinimumVersion string         `json:",omitempty"`
	Instances      []PackageVersion
}

type PackageVersion struct {
	OS      string `json:",omitempty"`
	Arch    string `json:",omitempty"`
	Version string
	Token   string `json:",omitempty"`
}

type ClientSettingsOverrides struct {
	MkdirPath      string `json:",omitempty"`
	ImageBaseURL   string `json:",omitempty"`
	ModuleCode     string `json:",omitempty"`
	SCPDestination string `json:",omitempty"`
	ServerHostname string `json:",omitempty"`
}

type WorldMessagePayload struct {
	BaseMessagePayload
	Calendar       string
	ClientSettings *ClientSettingsOverrides `json:",omitempty"`
}

type RedirectMessagePayload struct {
	BaseMessagePayload
	Host   string
	Port   int
	Reason string `json:",omitempty"`
}

// TimerAcknowledgeMessagePayload conveys to the requesting client
// that their TimerRequest message was accepted.
type TimerAcknowledgeMessagePayload struct {
	BaseMessagePayload
	RequestID        string
	RequestingClient string `json:",omitempty"`
	RequestedBy      string `json:",omitempty"`
}

// TimerRequestMessagePayload carries a client's request to add a timer
// to the GM's time tracker.
type TimerRequestMessagePayload struct {
	BaseMessagePayload
//TODO	BatchableMessagePayload

	// If true, the timer should be visible to the players instead of just the GM
	ShowToAll bool

	// If true, the timer should start running as soon as it's created
	IsRunning bool

	// A unique identifier for this request (recommend using a UUID)
	RequestID string

	// One-line description of the timer's purpose
	Description string

	// Absolute or relative expiration time for this timer.
	Expires string

	// If nonempty, the timer should only be visible to these users.
	Targets []string

	// The server will fill in this information about the requesting client.
	RequestedBy      string
	RequestingClient string
}

//TODOfunc (c TimerRequestMessagePayload) NeedsToBeSplit() bool {
//TODO	l := 123 + len(c.RequestID) + len(c.Description) + len(c.Expires) + len(c.RequestedBy) + len(c.RequestingClient)
//TODO	// TMRQ ShowToAll:false IsRunning:false RequestID:x Description:x Expires:x Targets:[x] RequestedBy:x RequestingClient:x
//TODO	for _, a := range c.Targets {
//TODO		l += len(a) + 3
//TODO	}
//TODO	return l > MaxServerMessageSize
//TODO}
//TODO
//TODOfunc (c TimerRequestMessagePayload) Split() []any {
//TODO	payloads := make([]any, len(c.Targets))
//TODO	gid := uuid.NewString()
//TODO
//TODO	for i, instance := range c.Targets {
//TODO		p := TimerRequestMessagePayload{
//TODO//TODO			BatchableMessagePayload: BatchableMessagePayload{
//TODO				TotalBatches: len(c.Targets),
//TODO				Batch:        i,
//TODO				BatchGroup:   gid,
//TODO			},
//TODO		}
//TODO		if i == 0 {
//TODO			p.ShowToAll = c.ShowToAll
//TODO			p.IsRunning = c.IsRunning
//TODO			p.RequestID = c.RequestID
//TODO			p.Description = c.Description
//TODO			p.Expires = c.Expires
//TODO			p.RequestedBy = c.RequestedBy
//TODO			p.RequestingClient = c.RequestingClient
//TODO		}
//TODO		p.Targets = []string{instance}
//TODO		payloads[i] = p
//TODO	}
//TODO	return payloads
//TODO}
//TODO
//TODOfunc (c TimerRequestMessagePayload) Reassemble(p []any) (any, error) {
//TODO	newp := TimerRequestMessagePayload{}
//TODO
//TODO	for i, d := range p {
//TODO		pp, ok := d.(TimerRequestMessagePayload)
//TODO		if !ok {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d is of type %T", c, i, len(p), d)
//TODO		}
//TODO		if pp.Batch != i {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d of %d claims to be #%d", c, i, len(p), pp.Batch)
//TODO		}
//TODO		if pp.TotalBatches != len(p) {
//TODO			return newp, fmt.Errorf("batched %T packet fragment #%d claims there will be %d batches but %d were collected", c, i, pp.TotalBatches, len(p))
//TODO		}
//TODO
//TODO		if i == 0 {
//TODO			newp.ShowToAll = pp.ShowToAll
//TODO			newp.IsRunning = pp.IsRunning
//TODO			newp.RequestID = pp.RequestID
//TODO			newp.Description = pp.Description
//TODO			newp.Expires = pp.Expires
//TODO			newp.RequestedBy = pp.RequestedBy
//TODO			newp.RequestingClient = pp.RequestingClient
//TODO		}
//TODO		for _, t := range pp.Targets {
//TODO			newp.Targets = append(newp.Targets, t)
//TODO		}
//TODO	}
//TODO	return newp, nil
//TODO}
//TODO
//TODOfunc (c TimerRequestMessagePayload) AbortPayload(reason string, batchNumber int) any {
//TODO	return TimerRequestMessagePayload{
//TODO//TODO		BatchableMessagePayload: BatchableMessagePayload{
//TODO			BatchError:   reason,
//TODO			Batch:        batchNumber,
//TODO			TotalBatches: c.TotalBatches,
//TODO		},
//TODO		RequestID:        c.RequestID,
//TODO		Description:      c.Description,
//TODO		RequestedBy:      c.RequestedBy,
//TODO		RequestingClient: c.RequestingClient,
//TODO	}
//TODO}

// TimerRequest sends a timer requst to the GM. If approved, the new timer will be added
// to the list of things being tracked in the game.

func (c *Connection) TimerRequest(id, description, expires string, targets []string, isRunning, showToAll bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(TimerRequest, TimerRequestMessagePayload{
		ShowToAll:   showToAll,
		IsRunning:   isRunning,
		RequestID:   id,
		Description: description,
		Expires:     expires,
		Targets:     targets,
	})
}

//
// Concurrency and general flow of operation for Dial:
// Dial itself will block until the session with the server is completed.
// Thus, a client program will probably run it in a goroutine, using
// a channel subscribed to ERROR to receive any errors encountered by it
// (otherwise the errors are at least logged).
//
// The Dial call does have some concurrent operations of its own, though,
// to facilitate bidirectional communication with the service without
// stopping the client application or tripping over its own feet.
//
// Dial
// ^ tryConnect
// |   establish socket to server
// |   launch login----------------------------------->login (l<-, s<-, <-s)
// |                                                     receive preamble
// |                                                     negotiate auth
// |   wait for login or cancel (<-l)<--------------------------+
// |     if cancel, close socket to terminate login
// |   abandon if login failed
// | interact (close on exit) (<-m)
// |   launch listen---------------------------------->listen (l<-)
// |   buffer messages from m channel (from app)         read from server (<-s)
// |   send buffered messages (s<-)                      dispatch to chans (*<-)
// |   if cancel, close socket to stop listen            may send too (m<-)
// |   if listen done, stop (<-l)<-------------------------------+
// |   our deferred close upon exit will stop listen
// +-repeat if staying connected
//
// Elsewhere in the client app:
//   send message (m<-)
//   receive subscribed server messages (<-*)
//   call cancel to terminate Dial/login/listen
//

// Dial connects to the server, negotiates the initial sign-on sequence
// with it, and then enters a loop to receive messages from the server
// until the connection is broken or the context is cancelled, at which
// point the Dial method returns.
//
// Dial is designed to be called in a goroutine so it can run in the
// background while the rest of the appliction continues with other
// tasks.
//
// Any errors encountered by the Dial method will be reported to
// the channel subscribed to watch for ERROR messages. If the client
// application did not subscribe to ERROR messages, they will be logged.
//
// Example:
//
//	ctx, cancel := context.Background()
//	server, err := NewConnection("example.org:2323",
//	                             WithAuthenticator(a),
//	                             WithContext(ctx))
//	defer cancel()
//	go server.Dial()
func (c *Connection) Dial() {
	if c == nil {
		return
	}
	var err error
	defer c.Close()

	c.signedOn = false
	for {
		err = c.tryConnect()
		if err == nil {
			// interact will set c.signedOn = true when ready
			if err = c.interact(); err != nil {
				c.Logf("mapper interact failure: %v", err)
			}
			c.signedOn = false
		} else if errors.Is(err, ErrRetryConnection) {
			c.Logf("retrying connection...")
			continue
		}

		if c.Context.Err() != nil || !c.StayConnected {
			break
		}
	}
}

func (c *Connection) tryConnect() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	var err error
	var conn net.Conn
	var i uint

	c.debug(DebugIO, "tryConnect() started")
	defer c.debug(DebugIO, "tryConnect() ended")

	for i = 0; c.Retries == 0 || i < c.Retries; i++ {
		if c.Timeout == 0 {
			var dialer net.Dialer
			conn, err = dialer.DialContext(c.Context, "tcp", c.Endpoint)
		} else {
			conn, err = net.DialTimeout("tcp", c.Endpoint, c.Timeout)
		}

		if err != nil {
			if c.Retries == 0 {
				c.Logf("attempting connection (try %d): %v", i+1, err)
			} else {
				c.Logf("attempting connection (try %d of %d): %v", i+1, c.Retries, err)
			}
		}
	}
	if err != nil {
		c.Logf("no more attempts allowed; giving up!")
		c.LastError = err
		return err
	}

	c.serverConn.conn = conn
	c.serverConn.reader = bufio.NewScanner(conn)
	c.serverConn.writer = bufio.NewWriter(conn)

	loginDone := make(chan error, 1)
	go c.login(loginDone)

syncloop:
	for {
		select {
		case err = <-loginDone:
			if err != nil {
				c.Logf("login process failed: %v", err)
				c.Close()
				c.LastError = err
				return err
			}
			break syncloop

		case <-c.Context.Done():
			c.Logf("context cancelled; closing connections and aborting login...")
			c.Close() // this will abort the scanner in login
			return fmt.Errorf("mapper: connection aborted by termination of context")
		}
	}

	return nil
}

func (c *Connection) login(done chan error) {
	defer close(done)
	if c == nil {
		done <- fmt.Errorf("login called on nil Connection")
		return
	}

	c.debug(DebugIO, "login() started")
	defer c.debug(DebugIO, "login() ended")

	c.Log("initial server negotiation...")
	syncDone := false
	authPending := false
	c.Preamble = nil

	// The first thing we hear from the server MUST be a PROTOCOL command.
	incomingPacket, err := c.serverConn.Receive()
	if err != nil {
		done <- err
		return
	}
	if incomingPacket == nil {
		done <- fmt.Errorf("EOF reading server's greeting")
		return
	}
	p, ok := incomingPacket.(ProtocolMessagePayload)
	if !ok {
		p, ok := incomingPacket.(ErrorMessagePayload)
		if ok {
			c.Logf("unable to begin server negotiation: %v", p.Error)
			done <- p.Error
			return
		}
		c.Logf("unable to begin server negotiation: no PROTOCOL message seen")
		done <- ErrServerProtocolError
		return
	}
	c.Protocol = p.ProtocolVersion
	if c.Protocol < MinimumSupportedMapProtocol {
		c.Logf("unable to connect to mapper with protocol older than %d (server offers %d)", MinimumSupportedMapProtocol, c.Protocol)
		done <- fmt.Errorf("server version %d too old (must be at least %d)", c.Protocol, MinimumSupportedMapProtocol)
		return
	}
	if c.Protocol > MaximumSupportedMapProtocol {
		c.Logf("unable to connect to mapper with protocol newer than %d (server offers %d)", MaximumSupportedMapProtocol, c.Protocol)
		c.Log("** UPGRADE GMA **")
		done <- fmt.Errorf("server version %d too new (must be at most %d)", c.Protocol, MaximumSupportedMapProtocol)
		return
	}

	// Now proceed to get logged in to the server
	for !syncDone {
		incomingPacket, err := c.serverConn.Receive()
		if err != nil {
			done <- err
			break
		}
		if incomingPacket == nil {
			break
		}

		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}

		// Protocol sequence:
		// <- PROTOCOL v
		// <- AC, DSM, REDIRECT, UPDATES, WORLD, // messages
		// <- OK
		// -> AUTH (if authentication required)
		// <- GRANTED or DENIED (if AUTH required and given)
		// <- AC, DSM, REDIRECT, UPDATES, WORLD, // messages
		// <- READY

		switch response := incomingPacket.(type) {
		case ChallengeMessagePayload:
			// OK Protocol=v [Challenge=data] [ServerUptime=time] [ServerActive=time]
			if response.Protocol != c.Protocol {
				c.Logf("server advertised protocol %v initially but then claimed version %v", c.Protocol, response.Protocol)
				done <- fmt.Errorf("server can't make up its mind whether it uses protocol %v or %v", c.Protocol, response.Protocol)
				return
			}

			c.ServerStats.Started = response.ServerStarted
			c.ServerStats.Active = response.ServerActive
			c.ServerStats.ConnectTime = response.ServerTime
			c.ServerStats.ServerVersion = response.ServerVersion

			if response.Challenge != nil {
				if c.Authenticator == nil {
					c.Log("Server requires authentication but no authenticator was provided for the client.")
					done <- ErrAuthenticationRequired
					return
				}
				c.Log("authenticating to server")
				c.Authenticator.Reset()
				authResponse, err := c.Authenticator.AcceptChallengeBytesWithIterations(response.Challenge, response.Iterations)
				if err != nil {
					c.Logf("error accepting server's challenge: %v", err)
					done <- err
					return
				}
				c.serverConn.Send(Auth, AuthMessagePayload{
					Response: authResponse,
					Client:   c.Authenticator.Client,
					User:     c.Authenticator.Username,
				})
				c.Log("authentication sent, awaiting validation.")
				if err := c.serverConn.Flush(); err != nil {
					c.Logf("can't authenticate: %v", err)
				}
				authPending = true
			} else {
				c.Logf("using protocol %d.", c.Protocol)
				c.Log("server sync complete. No authentication requested by server.")
			}
			syncDone = true

		case AddCharacterMessagePayload:
			c.receiveAddCharacter(response)

		case RedirectMessagePayload:
			c.Logf("server requests that we connect instead to %s port %d", response.Host, response.Port)
			if response.Reason != "" {
				c.Logf("reason for server redirect request: %s", response.Reason)
			}
			if err = c.serverConn.conn.Close(); err != nil {
				c.Logf("error closing existing socket to server: %v", err)
			}
			c.PartialReset()
			c.Endpoint = fmt.Sprintf("%s:%d", response.Host, response.Port)
			done <- ErrRetryConnection
			return

		case UpdateStatusMarkerMessagePayload:
			c.receiveDSM(response)

		case WorldMessagePayload:
			c.CalendarSystem = response.Calendar
			c.ClientSettings = nil
			if response.ClientSettings != nil && (response.ClientSettings.MkdirPath != "" ||
				response.ClientSettings.ImageBaseURL != "" ||
				response.ClientSettings.ModuleCode != "" ||
				response.ClientSettings.SCPDestination != "" ||
				response.ClientSettings.ServerHostname != "") {
				c.ClientSettings = &ClientSettingsOverrides{
					MkdirPath:      response.ClientSettings.MkdirPath,
					ImageBaseURL:   response.ClientSettings.ImageBaseURL,
					ModuleCode:     response.ClientSettings.ModuleCode,
					SCPDestination: response.ClientSettings.SCPDestination,
					ServerHostname: response.ClientSettings.ServerHostname,
				}
			}

		case UpdateVersionsMessagePayload:
			c.receiveUpdateVersions(response)

		case CommentMessagePayload:
			c.Preamble = append(c.Preamble, response.Text)

		default:
			c.Log("ignoring unexpected data before server challenge")
		}
	}

	if !syncDone {
		// something happened before we could login.
		done <- fmt.Errorf("unexpected EOF while negotiating login with server")
		return
	}

	// If we're still waiting for authentication results, do that...
	c.debug(DebugIO, "Switched to authentication result scanner")
	for authPending {
		incomingPacket, err := c.serverConn.Receive()
		if err != nil {
			done <- err
			return
		}
		if incomingPacket == nil {
			break
		}
		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}
		switch response := incomingPacket.(type) {
		case DeniedMessagePayload:
			c.Logf("access denied by server: %v", response.Reason)
			done <- ErrAuthenticationFailed
			return

		case GrantedMessagePayload:
			c.Logf("access granted for %s", response.User)
			authPending = false
			if c.Authenticator != nil {
				c.Authenticator.Username = response.User
			}

		case UpdatePeerListMessagePayload, CommentMessagePayload:
			c.Logf("Ignoring message %v while waiting for authentication to complete", incomingPacket.MessageType())

		default:
			c.Logf("unexpected server message %v while waiting for authentication to complete", incomingPacket.MessageType())
		}
	}
	if authPending {
		done <- fmt.Errorf("mapper: unexpected EOF while waiting for authentication to complete")
		return
	}

	// wait for server READY signal, accept incoming preliminary data
waitForReady:
	for {
		incomingPacket, err := c.serverConn.Receive()
		if err != nil {
			done <- err
			return
		}
		if incomingPacket == nil {
			break
		}

		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}

		switch response := incomingPacket.(type) {
		case AddCharacterMessagePayload:
			c.receiveAddCharacter(response)

		case UpdateStatusMarkerMessagePayload:
			c.receiveDSM(response)

		case RedirectMessagePayload:
			c.Logf("server requests that we connect instead to %s port %d", response.Host, response.Port)
			if response.Reason != "" {
				c.Logf("reason for server redirect request: %s", response.Reason)
			}
			if err = c.serverConn.conn.Close(); err != nil {
				c.Logf("error closing existing socket to server: %v", err)
			}
			c.PartialReset()
			c.Endpoint = fmt.Sprintf("%s:%d", response.Host, response.Port)
			done <- ErrRetryConnection
			return

		case WorldMessagePayload:
			c.CalendarSystem = response.Calendar
			c.ClientSettings = nil
			if response.ClientSettings != nil && (response.ClientSettings.MkdirPath != "" ||
				response.ClientSettings.ImageBaseURL != "" ||
				response.ClientSettings.ModuleCode != "" ||
				response.ClientSettings.SCPDestination != "" ||
				response.ClientSettings.ServerHostname != "") {
				c.ClientSettings = &ClientSettingsOverrides{
					MkdirPath:      response.ClientSettings.MkdirPath,
					ImageBaseURL:   response.ClientSettings.ImageBaseURL,
					ModuleCode:     response.ClientSettings.ModuleCode,
					SCPDestination: response.ClientSettings.SCPDestination,
					ServerHostname: response.ClientSettings.ServerHostname,
				}
			}

		case UpdateVersionsMessagePayload:
			c.receiveUpdateVersions(response)

		case CommentMessagePayload:
			c.Preamble = append(c.Preamble, response.Text)

		case ReadyMessagePayload:
			break waitForReady

		default:
			c.Log("ignoring unexpected data before server ready signal")
		}
	}
	c.debug(DebugIO, "Server ready; filtering to subscription list")

	if err := c.filterSubscriptions(); err != nil {
		done <- err
		return
	}

	if c.DebuggingLevel >= 2 {
		c.debug(DebugIO, "Completed server sign-on process")
		if c.Authenticator != nil {
			c.debug(DebugAuth, fmt.Sprintf("Logged in as %s", c.Authenticator.Username))
		}
		c.debug(DebugIO, fmt.Sprintf("Server is using protocol version %d", c.Protocol))
		c.debug(DebugIO, fmt.Sprintf("Defined Characters:\n%s", CharacterDefinitions(c.Characters).Text()))
		c.debug(DebugIO, fmt.Sprintf("Defined Status Markers:\n%s", StatusMarkerDefinitions(c.Conditions).Text()))
		c.debug(DebugIO, "Preamble:\n"+strings.Join(c.Preamble, "\n"))
		c.debug(DebugIO, fmt.Sprintf("Last error: %v", c.LastError))
	}
	if c.ClientSettings != nil {
		c.debug(DebugIO, fmt.Sprintf("Server requests client settings overrides MkdirPath=%s, ImageBaseURL=%s, ModuleCode=%s, SCPDestination=%s, ServerHostname=%s",
			c.ClientSettings.MkdirPath,
			c.ClientSettings.ImageBaseURL,
			c.ClientSettings.ModuleCode,
			c.ClientSettings.SCPDestination,
			c.ClientSettings.ServerHostname,
		))
	}
	if c.ReadySignal != nil {
		c.ReadySignal <- 0
	}
}

func (c *Connection) receiveDSM(d UpdateStatusMarkerMessagePayload) {
	if c != nil {
		c.Conditions[d.Condition] = StatusMarkerDefinition{
			Condition:   d.Condition,
			Shape:       d.Shape,
			Color:       d.Color,
			Description: d.Description,
			Transparent: d.Transparent,
		}
	}
}

func (c *Connection) receiveAddCharacter(d AddCharacterMessagePayload) {
	if c != nil {
		critter := CreatureToken{
			BaseMapObject: BaseMapObject{
				ID: d.ObjID(),
			},
			Name:         d.Name,
			Color:        d.Color,
			Killed:       d.Killed,
			Dim:          d.Dim,
			Hidden:       d.Hidden,
			PolyGM:       d.PolyGM,
			CreatureType: CreatureTypePlayer,
			MoveMode:     d.MoveMode,
			Reach:        d.Reach,
			Elev:         d.Elev,
			Gx:           d.Gx,
			Gy:           d.Gy,
			Note:         d.Note,
			DispSize:     d.DispSize,
			StatusList:   d.StatusList,
			CustomReach: CreatureCustomReach{
				Enabled:  d.CustomReach.Enabled,
				Natural:  d.CustomReach.Natural,
				Extended: d.CustomReach.Extended,
			},
			Targets: d.Targets,
		}
		if err := critter.SetSizes(d.SkinSize, d.Skin, d.Size); err != nil {
			c.Logf("ERROR setting creature %s skinsizes to %v (skin=%v, size=%v): %v",
				d.Name, d.SkinSize, d.Skin, d.Size, err)
		}
		if d.AoE != nil {
			critter.AoE = &RadiusAoE{
				Radius: d.AoE.Radius,
				Color:  d.AoE.Color,
			}
		}
		if d.Health != nil {
			critter.Health = &CreatureHealth{
				IsFlatFooted:    d.Health.IsFlatFooted,
				IsStable:        d.Health.IsStable,
				MaxHP:           d.Health.MaxHP,
				TmpHP:           d.Health.TmpHP,
				TmpDamage:       d.Health.TmpDamage,
				LethalDamage:    d.Health.LethalDamage,
				NonLethalDamage: d.Health.NonLethalDamage,
				Con:             d.Health.Con,
				HPBlur:          d.Health.HPBlur,
				Condition:       d.Health.Condition,
				AC:              d.Health.AC,
				FlatFootedAC:    d.Health.FlatFootedAC,
				TouchAC:         d.Health.TouchAC,
				CMD:             d.Health.CMD,
			}
		}

		c.Characters[d.Name] = PlayerToken{
			CreatureToken: critter,
		}
	}
}

func (c *Connection) receiveUpdateVersions(d UpdateVersionsMessagePayload) {
	if c != nil {
		for _, pkg := range d.Packages {
			c.PackageUpdatesAvailable[pkg.Name] = append(c.PackageUpdatesAvailable[pkg.Name], pkg.Instances...)
		}
	}
}

// listen for, and dispatch, incoming server messages
func (c *Connection) listen(done chan error) {
	if c == nil {
		done <- fmt.Errorf("listen called on nil Connection")
		close(done)
		return
	}
	defer func() {
		close(done)
		c.Log("stopped listening to server")
		c.debug(DebugIO, "listen() ended")
	}()
	c.debug(DebugIO, "listen() started")

	c.Log("listening for server messages to dispatch...")
	for {
		incomingPacket, err := c.serverConn.Receive()
		if err != nil {
			done <- err
			return
		}
		if incomingPacket == nil {
			break
		}

		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}

		switch cmd := incomingPacket.(type) {
		case AddAudioMessagePayload:
			if ch, ok := c.Subscriptions[AddAudio]; ok {
				ch <- cmd
			}

		case AddImageMessagePayload:
			if ch, ok := c.Subscriptions[AddImage]; ok {
				ch <- cmd
			}

		case AddObjAttributesMessagePayload:
			if ch, ok := c.Subscriptions[AddObjAttributes]; ok {
				ch <- cmd
			}

		case AdjustViewMessagePayload:
			if ch, ok := c.Subscriptions[AdjustView]; ok {
				ch <- cmd
			}

		case CharacterNameMessagePayload:
			if ch, ok := c.Subscriptions[CharacterName]; ok {
				ch <- cmd
			}

		case ChatMessageMessagePayload:
			if ch, ok := c.Subscriptions[ChatMessage]; ok {
				ch <- cmd
			}

		case ClearMessagePayload:
			if ch, ok := c.Subscriptions[Clear]; ok {
				ch <- cmd
			}

		case ClearChatMessagePayload:
			if ch, ok := c.Subscriptions[ClearChat]; ok {
				ch <- cmd
			}

		case ClearFromMessagePayload:
			if ch, ok := c.Subscriptions[ClearFrom]; ok {
				ch <- cmd
			}

		case CombatModeMessagePayload:
			if ch, ok := c.Subscriptions[CombatMode]; ok {
				ch <- cmd
			}

		case CommentMessagePayload:
			if ch, ok := c.Subscriptions[Comment]; ok {
				ch <- cmd
			}

		case EchoMessagePayload:
			if ch, ok := c.Subscriptions[Echo]; ok {
				ch <- cmd
			}

		case FailedMessagePayload:
			if ch, ok := c.Subscriptions[Failed]; ok {
				ch <- cmd
			}

		case HitPointAcknowledgeMessagePayload:
			if ch, ok := c.Subscriptions[HitPointAcknowledge]; ok {
				ch <- cmd
			}

		case HitPointRequestMessagePayload:
			if ch, ok := c.Subscriptions[HitPointRequest]; ok {
				ch <- cmd
			}

		case LoadArcObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadArcObject]; ok {
				ch <- cmd
			}

		case LoadCircleObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadCircleObject]; ok {
				ch <- cmd
			}

		case LoadFromMessagePayload:
			if ch, ok := c.Subscriptions[LoadFrom]; ok {
				ch <- cmd
			}

		case LoadLineObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadLineObject]; ok {
				ch <- cmd
			}

		case LoadPolygonObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadPolygonObject]; ok {
				ch <- cmd
			}

		case LoadRectangleObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadRectangleObject]; ok {
				ch <- cmd
			}

		case LoadSpellAreaOfEffectObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadSpellAreaOfEffectObject]; ok {
				ch <- cmd
			}

		case LoadTextObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadTextObject]; ok {
				ch <- cmd
			}

		case LoadTileObjectMessagePayload:
			if ch, ok := c.Subscriptions[LoadTileObject]; ok {
				ch <- cmd
			}

		case MarcoMessagePayload:
			if ch, ok := c.Subscriptions[Marco]; ok {
				ch <- cmd
			} else {
				c.serverConn.Send(Polo, nil)
			}

		case MarkMessagePayload:
			if ch, ok := c.Subscriptions[Mark]; ok {
				ch <- cmd
			}

		case PlayAudioMessagePayload:
			if ch, ok := c.Subscriptions[PlayAudio]; ok {
				ch <- cmd
			}

		case PlaceSomeoneMessagePayload:
			if ch, ok := c.Subscriptions[PlaceSomeone]; ok {
				ch <- cmd
			}

		case PrivMessagePayload:
			if ch, ok := c.Subscriptions[Priv]; ok {
				ch <- cmd
			}

		case QueryAudioMessagePayload:
			if ch, ok := c.Subscriptions[QueryAudio]; ok {
				ch <- cmd
			}

		case QueryImageMessagePayload:
			if ch, ok := c.Subscriptions[QueryImage]; ok {
				ch <- cmd
			}

		case RemoveObjAttributesMessagePayload:
			if ch, ok := c.Subscriptions[RemoveObjAttributes]; ok {
				ch <- cmd
			}

		case RollResultMessagePayload:
			if ch, ok := c.Subscriptions[RollResult]; ok {
				ch <- cmd
			}

		case TimerAcknowledgeMessagePayload:
			if ch, ok := c.Subscriptions[TimerAcknowledge]; ok {
				ch <- cmd
			}

		case TimerRequestMessagePayload:
			if ch, ok := c.Subscriptions[TimerRequest]; ok {
				ch <- cmd
			}

		case ToolbarMessagePayload:
			if ch, ok := c.Subscriptions[Toolbar]; ok {
				ch <- cmd
			}

		case UpdateClockMessagePayload:
			if ch, ok := c.Subscriptions[UpdateClock]; ok {
				ch <- cmd
			}

		case UpdateDicePresetsMessagePayload:
			if ch, ok := c.Subscriptions[UpdateDicePresets]; ok {
				ch <- cmd
			}

		case UpdateInitiativeMessagePayload:
			if ch, ok := c.Subscriptions[UpdateInitiative]; ok {
				ch <- cmd
			}

		case UpdateObjAttributesMessagePayload:
			if ch, ok := c.Subscriptions[UpdateObjAttributes]; ok {
				ch <- cmd
			}

		case UpdatePeerListMessagePayload:
			if ch, ok := c.Subscriptions[UpdatePeerList]; ok {
				ch <- cmd
			}

		case UpdateCoreDataMessagePayload:
			if ch, ok := c.Subscriptions[UpdateCoreData]; ok {
				ch <- cmd
			}

		case UpdateCoreIndexMessagePayload:
			if ch, ok := c.Subscriptions[UpdateCoreIndex]; ok {
				ch <- cmd
			}

		case UpdateProgressMessagePayload:
			if ch, ok := c.Subscriptions[UpdateProgress]; ok {
				ch <- cmd
			}

		case UpdateStatusMarkerMessagePayload:
			c.receiveDSM(cmd)
			if ch, ok := c.Subscriptions[UpdateStatusMarker]; ok {
				ch <- cmd
			}

		case UpdateTurnMessagePayload:
			if ch, ok := c.Subscriptions[UpdateTurn]; ok {
				ch <- cmd
			}

		case AddCharacterMessagePayload, ChallengeMessagePayload,
			GrantedMessagePayload, ProtocolMessagePayload, ReadyMessagePayload,
			UpdateVersionsMessagePayload, RedirectMessagePayload, WorldMessagePayload:

			c.reportError(fmt.Errorf("message type %v should not be sent to client at this stage in the session", cmd.MessageType()))

		case DeniedMessagePayload:
			c.reportError(fmt.Errorf("server has terminated our session: %s", cmd.Reason))
			return

		case AcceptMessagePayload, AddDicePresetsMessagePayload, AllowMessagePayload,
			AuthMessagePayload, DefineDicePresetsMessagePayload, DefineDicePresetDelegatesMessagePayload,
			FilterDicePresetsMessagePayload, FilterImagesMessagePayload, FilterAudioMessagePayload, PoloMessagePayload,
			QueryDicePresetsMessagePayload, QueryPeersMessagePayload,
			RollDiceMessagePayload, SyncMessagePayload, SyncChatMessagePayload:

			c.reportError(fmt.Errorf("message type %v should not be sent to a client (ignored)", cmd.MessageType()))

		default:
			if ch, ok := c.Subscriptions[UNKNOWN]; ok {
				ch <- UnknownMessagePayload{
					BaseMessagePayload: BaseMessagePayload{
						messageType: UNKNOWN,
						rawMessage:  incomingPacket.RawMessage(),
					},
				}
			} else {
				c.Logf("received unknown server message type: \"%v\"", cmd.MessageType())
			}
		}
	}
}

// report any sort of error to the client
func (c *Connection) reportError(e error) {
	if c == nil {
		return
	}
	c.LastError = e
	if ch, ok := c.Subscriptions[ERROR]; ok {
		ch <- ErrorMessagePayload{
			BaseMessagePayload: BaseMessagePayload{
				rawMessage:  "",
				messageType: ERROR,
			},
			Error: e,
		}
	} else {
		c.Logf("mapper error: %v", e)
	}
}

// listen and interact with the service until it's finished,
// then close our connection to it
func (c *Connection) interact() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	defer func() {
		c.signedOn = false
		c.Close()
	}()

	c.debug(DebugIO, "interact() started")
	defer c.debug(DebugIO, "interact() ended")

	listenerDone := make(chan error, 1)
	go c.listen(listenerDone)
	c.signedOn = true
	bufferReadable := make(chan byte, 1)

	for {
		//
		// Receive and buffer any messages to be sent out
		// to the server
		//
		select {
		case <-c.Context.Done():
			c.Log("interact: context done, stopping")
			return nil
		case err := <-listenerDone:
			c.Logf("interact: listener done (%v), stopping", err)
			return err
		case packet := <-c.serverConn.sendChan:
			if len(c.serverConn.sendBuf) == 0 {
				bufferReadable <- 0
			}
			c.serverConn.sendBuf = append(c.serverConn.sendBuf, packet)
		case <-bufferReadable:
			if c.serverConn.writer != nil && len(c.serverConn.sendBuf) > 0 {
				if (c.DebuggingLevel & DebugBinary) != 0 {
					c.debug(DebugBinary, util.Hexdump([]byte(c.serverConn.sendBuf[0])))
				}
				c.debug(DebugIO, fmt.Sprintf("client->%q (%d)", c.serverConn.sendBuf[0], len(c.serverConn.sendBuf)))
				if written, err := c.serverConn.writer.WriteString(c.serverConn.sendBuf[0]); err != nil {
					return fmt.Errorf("only wrote %d of %d bytes: %v", written, len(c.serverConn.sendBuf[0]), err)
				}
				if err := c.serverConn.writer.Flush(); err != nil {
					c.Logf("interact: unable to flush: %v", err)
					return err
				}
				c.serverConn.sendBuf = c.serverConn.sendBuf[1:]
			}
			if len(c.serverConn.sendBuf) > 0 {
				bufferReadable <- 0
			}
		}
	}
}

// Any time the subscription list changes,
// we need to call this to let the server know what kinds of
// messages the client wants to see.
func (c *Connection) filterSubscriptions() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	if !c.IsReady() {
		return nil
	}

	subList := []string{"MARCO", "FAILED", "PRIV"} // these are unconditional
	for msg := range c.Subscriptions {
		switch msg {
		//Accept (client)
		//AddCharacter (forbidden)
		//AddDicePresets (client)
		//Allow (client)
		//Auth (client)
		//Challenge (forbidden)
		//DefineDicePresets (client)
		//DefineDicePresetDelegates (client)
		//Denied (forbidden)
		//Failed (mandatory)
		//FilterCoreData (client)
		//FilterDicePresets (client)
		//FilterImages (client)
		//Granted (forbidden)
		//Marco (mandatory)
		//Polo (client)
		//Priv (mandatory)
		//Protocol (forbidden)
		//QueryCoreData (client)
		//QueryDicePresets (client)
		//QueryPeers (client)
		//Ready (forbidden)
		//Redirect (forbidden)
		//RollDice (client)
		//Sync (client)
		//SyncChat (client)
		//UpdateVersions (forbidden)
		//World (forbidden)

		case AddAudio:
			subList = append(subList, "AA")
		case AddImage:
			subList = append(subList, "AI")
		case AddObjAttributes:
			subList = append(subList, "OA+")
		case AdjustView:
			subList = append(subList, "AV")
		case ChatMessage:
			subList = append(subList, "TO")
		case Clear:
			subList = append(subList, "CLR")
		case ClearChat:
			subList = append(subList, "CC")
		case ClearFrom:
			subList = append(subList, "CLR@")
		case CombatMode:
			subList = append(subList, "CO")
		case Comment:
			subList = append(subList, "//")
		case Echo:
			subList = append(subList, "ECHO")
		case LoadFrom:
			subList = append(subList, "L")
		case LoadArcObject:
			subList = append(subList, "LS-ARC")
		case LoadCircleObject:
			subList = append(subList, "LS-CIRC")
		case LoadLineObject:
			subList = append(subList, "LS-LINE")
		case LoadPolygonObject:
			subList = append(subList, "LS-POLY")
		case LoadRectangleObject:
			subList = append(subList, "LS-RECT")
		case LoadSpellAreaOfEffectObject:
			subList = append(subList, "LS-SAOE")
		case LoadTextObject:
			subList = append(subList, "LS-TEXT")
		case LoadTileObject:
			subList = append(subList, "LS-TILE")
		case Mark:
			subList = append(subList, "MARK")
		case PlaceSomeone:
			subList = append(subList, "PS")
		case PlayAudio:
			subList = append(subList, "SOUND")
		case QueryAudio:
			subList = append(subList, "AA?")
		case QueryImage:
			subList = append(subList, "AI?")
		case RemoveObjAttributes:
			subList = append(subList, "OA-")
		case RollResult:
			subList = append(subList, "ROLL")
		case TimerAcknowledge:
			subList = append(subList, "TMACK")
		case TimerRequest:
			subList = append(subList, "TMRQ")
		case Toolbar:
			subList = append(subList, "TB")
		case UpdateClock:
			subList = append(subList, "CS")
		case UpdateCoreData:
			subList = append(subList, "CORE=")
		case UpdateCoreIndex:
			subList = append(subList, "COREIDX=")
		case UpdateDicePresets:
			subList = append(subList, "DD=")
		case UpdateInitiative:
			subList = append(subList, "IL")
		case UpdateObjAttributes:
			subList = append(subList, "OA")
		case UpdatePeerList:
			subList = append(subList, "CONN")
		case UpdateProgress:
			subList = append(subList, "PROGRESS")
		case UpdateStatusMarker:
			subList = append(subList, "DSM")
		case UpdateTurn:
			subList = append(subList, "I")
		}
	}

	return c.serverConn.Send(Accept, AcceptMessagePayload{
		Messages: subList,
	})
}

//
// Tell the server to send us all possible messages.
//
/*
func (c *Connection) unfilterSubscriptions() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Accept, AcceptMessagePayload{
		Messages: nil,
	})
}
*/

// CheckVersionOf returns the closest match of the requested package
// to the platform we are currently running, or nil if we're already
// on the advertised version.
// on the advertised version.
func (c *Connection) CheckVersionOf(packageName, myVersionNumber string) (*PackageVersion, error) {
	var availableVersion *PackageVersion

	candidates, ok := c.PackageUpdatesAvailable[packageName]
	if !ok {
		return nil, fmt.Errorf("The server provided no upgrade information for package \"%s\"", packageName)
	}
	for _, candidate := range candidates {
		if (candidate.OS == "" || candidate.OS == runtime.GOOS) && (candidate.Arch == "" || candidate.Arch == runtime.GOARCH) {
			if availableVersion != nil && availableVersion.Version != "" && ((candidate.OS != "" && availableVersion.OS == "") || (candidate.Arch != "" && availableVersion.Arch == "")) {
				// found a more specific match, use that instead
				availableVersion = &candidate
			} else if availableVersion == nil || availableVersion.Version == "" {
				availableVersion = &candidate
			}
		}
	}

	return availableVersion, nil
}

// @[00]@| Go-GMA 5.33.0
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2026 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
// @[13]@| points along that historical time line.
// @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[15]@| License as described in the accompanying LICENSE file distributed
// @[16]@| with GMA.
// @[17]@|
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
