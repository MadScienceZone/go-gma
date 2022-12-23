/*
########################################################################################
#  _______  _______  _______             _______     _______     _______               #
# (  ____ \(       )(  ___  )           (  ____ \   (  __   )   (  __   )              #
# | (    \/| () () || (   ) |           | (    \/   | (  )  |   | (  )  |              #
# | |      | || || || (___) |           | (____     | | /   |   | | /   |              #
# | | ____ | |(_)| ||  ___  |           (_____ \    | (/ /) |   | (/ /) |              #
# | | \_  )| |   | || (   ) | Game            ) )   |   / | |   |   / | |              #
# | (___) || )   ( || )   ( | Master's  /\____) ) _ |  (__) | _ |  (__) |              #
# (_______)|/     \||/     \| Assistant \______/ (_)(_______)(_)(_______)              #
#                                                                                      #
########################################################################################
*/

//
// Package mapper implements a standard client interface for the mapper service.
//
// EXPERIMENTAL CODE
//
// THIS PACKAGE IS STILL A WORK IN PROGRESS and has not been
// completely tested yet. Although GMA generally is a stable
// product, this module of it is new, and is not.
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
//
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
	"runtime"
	"strings"
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

//
// Debugging information is enabled by selecting a nummber
// of discrete topics which you want logged as the application
// runs (previous versions used a "level" of verbosity which
// doesn't provide the better granularity this version provides
// to just get the info you want.
//
type DebugFlags uint64

const (
	DebugAuth DebugFlags = 1 << iota
	DebugBinary
	DebugEvents
	DebugIO
	DebugMisc
	DebugAll DebugFlags = 0xffffffff
)

//
// DebugFlagNames returns a string representation of
// the debugging flags (topics) stored in the DebugFlags
// value passed in.
//
func DebugFlagNames(flags DebugFlags) string {
	if flags == 0 {
		return "<none>"
	}
	if flags == DebugAll {
		return "<all>"
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
		{bits: DebugMisc, name: "misc"},
	} {
		if (flags & f.bits) != 0 {
			list = append(list, f.name)
		}
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
			case "misc":
				d |= DebugMisc
			default:
				return 0, fmt.Errorf("unknown debug flag name \"%s\"", flag)
			}
		}
	}
	return d, nil
}

//
// Connection describes a connection to the server. These are
// created with NewConnection and then send methods such as
// Subscribe and Dial.
//
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
}

//
// Log writes data to our log destination.
//
func (c *Connection) Log(message ...any) {
	if c != nil && c.Logger != nil {
		message = append([]any{"[client] "}, message...)
		c.Logger.Print(message...)
	}
}

//
// Logf writes data to our log destination.
//
func (c *Connection) Logf(format string, data ...any) {
	if c != nil && c.Logger != nil {
		c.Logger.Printf("[client] "+format, data...)
	}
}

//
// IsReady returns true if the connection to the server
// has completed and authentication was successful, so
// the connection is ready for interactive use.
//
func (c *Connection) IsReady() bool {
	return c != nil && c.serverConn.IsReady() && c.signedOn
}

//
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
//
func WithContext(ctx context.Context) func(*Connection) error {
	return func(c *Connection) error {
		c.Context = ctx
		return nil
	}
}

//
// ConnectionOption is an option to be passed to the NewConnection
// function.
//
type ConnectionOption func(*Connection) error

//
// WithSubscription modifies the behavior of the NewConnection function
// by adding a server message subscription to the connection just as if
// the Subscribe method had been called on the connection value.
//
// For example, this:
//   server, err := NewConnection(endpoint,
//                    WithSubscription(chats, ChatMessage, RollResult),
//                    WithSubscription(oops, ERROR, UNKNOWN))
//   go server.Dial()
// is equivalent to this:
//   server, err := NewConnection(endpoint)
//   err = server.Subscribe(chats, ChatMessage, RollResult)
//   err = server.Subscribe(oops, ERROR, UNKNOWN)
//   go server.Dial()
// (Of course, real production code should check the returned error values.)
//
func WithSubscription(ch chan MessagePayload, messages ...ServerMessage) ConnectionOption {
	return func(c *Connection) error {
		return c.Subscribe(ch, messages...)
	}
}

//
// WithAuthenticator modifies the behavior of the NewConnection function
// by adding an authenticator which will be used to identify the client
// to the server. If this option is not given, no attempt will be made
// to authenticate, which is only appropriate for servers which do not
// require authentication. (Which, hopefully, won't be the case anyway.)
//
func WithAuthenticator(a *auth.Authenticator) ConnectionOption {
	return func(c *Connection) error {
		c.Authenticator = a
		return nil
	}
}

//
// WithLogger modifies the behavior of the NewConnection function
// by specifying a custom logger instead of the default one for
// the Connection to use during its operations.
//
func WithLogger(l *log.Logger) ConnectionOption {
	return func(c *Connection) error {
		c.Logger = l
		return nil
	}
}

//
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
//
func WithTimeout(t time.Duration) ConnectionOption {
	return func(c *Connection) error {
		c.Timeout = t
		return nil
	}
}

//
// WithRetries modifies the behavior of the NewConnection function
// to indicate how many times the Dial method should try to
// establish a connection to the server before giving up.
//
// Setting this to 0 means to retry infinitely many times.
// The default is to make a single attempt to connect to the
// server.
//
func WithRetries(n uint) ConnectionOption {
	return func(c *Connection) error {
		c.Retries = n
		return nil
	}
}

//
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
//
func StayConnected(enable bool) ConnectionOption {
	return func(c *Connection) error {
		c.StayConnected = enable
		return nil
	}
}

//
// WithDebugging modifies the behavior of the NewConnection function
// so that the operations of the Connection's interaction with the
// server are logged to varying levels of verbosity.
//
func WithDebugging(flags DebugFlags) ConnectionOption {
	return func(c *Connection) error {
		c.DebuggingLevel = flags
		return nil
	}
}

//
// NewConnection creates a new server connection value which can then be used to
// manage our communication with the server.
//
// After the endpoint, you may specify any of the following options
// to define the behavior desired for this connection:
//   StayConnected(bool)
//   WithAuthenticator(a)
//   WithDebugging(level)
//   WithContext(ctx)
//   WithLogger(l)
//   WithRetries(n)
//   WithSubscription(ch, msgs...)
//   WithTimeout(t)
//
// Example:
//   a := NewClientAuthenticator("fred", []byte("sekret"), "some random client")
//   ctx, cancel := context.Background()
//   defer cancel()
//
//   messages := make(chan MessagePayload, 10)
//   problems := make(chan MessagePayload, 10)
//
//   server, err := NewConnection("mygame.example.org:2323",
//                     WithAuthenticator(a),
//                     WithContext(ctx),
//                     StayConnected(true),
//                     WithSubscription(messages, ChatMessage, RollResult),
//                     WithSubscription(problems, ERROR, UNKNOWN))
//   if err != nil {
//      log.Fatalf("can't reach the server: %v", err)
//   }
//   go server.Dial()
//
func NewConnection(endpoint string, opts ...ConnectionOption) (Connection, error) {
	newCon := Connection{
		Context:                 context.Background(),
		Endpoint:                endpoint,
		Subscriptions:           make(map[ServerMessage]chan MessagePayload),
		Characters:              make(map[string]PlayerToken),
		Conditions:              make(map[string]StatusMarkerDefinition),
		Retries:                 1,
		Logger:                  log.Default(),
		Gauges:                  make(map[string]*UpdateProgressMessagePayload),
		PackageUpdatesAvailable: make(map[string][]PackageVersion),
		serverConn: MapConnection{
			sendChan: make(chan string, 16),
		},
	}

	for _, o := range opts {
		if err := o(&newCon); err != nil {
			return newCon, err
		}
	}

	return newCon, nil
}

//
// Log debugging info at the given level.
//
func (c *Connection) debug(level DebugFlags, msg string) {
	if c != nil && (c.DebuggingLevel&level) != 0 {
		for i, line := range strings.Split(msg, "\n") {
			if line != "" {
				c.Logf("DEBUG%s%02d: %s", DebugFlagNames(level), i, line)
			}
		}
	}
}

//
// Close terminates the connection to the server.
// Note that the Dial function normally closes the connection
// before it returns, so calling this explicitly should not
// normally be necessary.
//
// Calling Close will result in the Dial function stopping
// due to the connection disappearing, but it is better to cancel
// the context being watched by Dial instead.
//
func (c *Connection) Close() {
	if c != nil {
		c.debug(DebugIO, "Close()")
		c.serverConn.Close()
	}
}

//
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
//   Marco:   Auto-reply with Polo
//   ERROR:   Log a message
//   UNKNOWN: Log a message
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
//   cm := make(chan MessagePayload, 1)
//   service, err := NewConnection(endpoint)
//   err = service.Subscribe(cm, ChatMessage)
//
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

//
// MessagePayload is an interface that includes any kind of message the server will
// send to us.
//
type MessagePayload interface {
	MessageType() ServerMessage
	RawMessage() string
	RawBytes() []byte
}

//
// ServerMessage is an arbitrary code which identifies specific message types that
// we can receive from the server. This value is passed to the Subscribe method
// and returned by the MessageType method. These values are intended for use
// within an actively-running program but are not guaranteed to remain stable across
// new releases of the code, so they should not be stored and re-used by a later
// execution of the client, nor passed to other programs whose definition of these
// values may not agree.
//
type ServerMessage byte

// Despite the warning above, we'll do our best to avoid changing these values
// if at all possible.

//
// ServerMessage values (see the comments accompanying the type definition).
//
const (
	Accept ServerMessage = iota
	AddCharacter
	AddDicePresets
	AddImage
	AddObjAttributes
	AdjustView
	Allow
	Auth
	Challenge
	ChatMessage
	Clear
	ClearChat
	ClearFrom
	CombatMode
	Comment
	DefineDicePresets
	Denied
	FilterDicePresets
	Granted
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
	Polo
	Priv
	Protocol
	QueryDicePresets
	QueryImage
	QueryPeers
	Ready
	RemoveObjAttributes
	RollDice
	RollResult
	Sync
	SyncChat
	Toolbar
	UpdateClock
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

//
// BaseMessagePayload is not a payload type that you should ever
// encounter directly, but it is included in all other payload
// types. It holds the bare minimum data for any server message.
//
type BaseMessagePayload struct {
	rawMessage  string        `json:"-"`
	messageType ServerMessage `json:"-"`
}

//
// RawMessage returns the raw message received from the server before
// it was parsed out into the MessagePayload the client should arguably
// be looking at instead.
//
// The raw message data may be useful for debugging purposes or other
// low-level poking around, though, so we make it available here.
//
func (p BaseMessagePayload) RawMessage() string { return p.rawMessage }
func (p BaseMessagePayload) RawBytes() []byte   { return []byte(p.rawMessage) }

//
// MessageType returns the type of message this MessagePayload represents.
// This value will be the same as the ServerMessage value used for the
// Subscribe function, and may be used with channels which receive multiple
// kinds of messages to differentiate them, like so:
//
//   select {
//   case p<-messages:
//       // This channel may receive a ChatMessage or RollResult.
//       switch p.MessageType() {
//       case ChatMessage:
//           // Do whatever with p.(ChatMessageMessagePayload)
//       case RollResult:
//           // Do whatever with p.(RollResultMessagePayload)
//       default:
//           // Something bad happened!
//       }
//    ...
//   }
//
// You can also use a type switch to accomplish the same thing and avoid
// the explicit type assertions:
//   select {
//   case p<-messages:
//       // This channel may receive a ChatMessage or RollResult.
//       switch msg := p.(type) {
//       case ChatMessageMessagePayload:
//           // Do whatever with msg
//       case RollResultMessagePayload:
//           // Do whatever with msg
//       default:
//           // Something bad happened!
//       }
//    ...
//   }
//
func (p BaseMessagePayload) MessageType() ServerMessage { return p.messageType }

//
// ErrorMessagePayload describes
// an error which encountered when trying to receive a message.
//
type ErrorMessagePayload struct {
	BaseMessagePayload
	OriginalMessageType ServerMessage
	Error               error
}

//
// UnknownMessagePayload describes a server message we received
// but have no idea what it is.
//
type UnknownMessagePayload struct {
	BaseMessagePayload
}

//
// ProtocolMessagePayload describes the server's statement of
// what protocol version it implements.
//
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

//
// AcceptMessagePayload holds the information sent by a client requesting
// that the server only send a subset of its possible message types to it.
//
// Clients send this by calling the Subscribe method on their connection.
//
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

//
// AddCharacterMessagePayload holds the information sent by the server's AddCharacter
// message to add a new PC to the party. This is not done for most creatures
// and NPCs encountered; it is for the PCs and significant NPCs who are important
// enough to be treated specially by clients (such as being included in menus).
//
type AddCharacterMessagePayload struct {
	BaseMessagePayload
	PlayerToken
}

//________________________________________________________________________________
//     _       _     _ ___
//    / \   __| | __| |_ _|_ __ ___   __ _  __ _  ___
//   / _ \ / _` |/ _` || || '_ ` _ \ / _` |/ _` |/ _ \
//  / ___ \ (_| | (_| || || | | | | | (_| | (_| |  __/
// /_/   \_\__,_|\__,_|___|_| |_| |_|\__,_|\__, |\___|
//                                         |___/

//
// AddImageMessagePayload holds the information sent by the server's AddImage
// message informing the client as to where it can locate an image's data.
//
// Call the AddImage method to send this message out to others if you know
// of an image file they should be aware of.
//
type AddImageMessagePayload struct {
	BaseMessagePayload
	ImageDefinition
}

//
// AddImage informs the server and peers about an image they can use.
//
func (c *Connection) AddImage(idef ImageDefinition) error {
	return c.serverConn.Send(AddImage, idef)
}

//     _       _     _  ___  _     _    _   _   _        _ _           _
//    / \   __| | __| |/ _ \| |__ (_)  / \ | |_| |_ _ __(_) |__  _   _| |_ ___  ___
//   / _ \ / _` |/ _` | | | | '_ \| | / _ \| __| __| '__| | '_ \| | | | __/ _ \/ __|
//  / ___ \ (_| | (_| | |_| | |_) | |/ ___ \ |_| |_| |  | | |_) | |_| | ||  __/\__ \
// /_/   \_\__,_|\__,_|\___/|_.__// /_/   \_\__|\__|_|  |_|_.__/ \__,_|\__\___||___/
//                              |__/

//
// AddObjAttributesMessagePayload holds the information sent by the server's AddObjAttributes
// message. This tells the client to adjust the multi-value attribute
// of the object with the given ID by adding the new values to it.
//
// Call the AddObjAttributes method to send this message out to other clients.
//
type AddObjAttributesMessagePayload struct {
	BaseMessagePayload
	ObjID    string
	AttrName string
	Values   []string
}

//
// AddObjAttributes informs peers to add a set of string values to the existing
// value of an object attribute. The attribute must be one whose value is a list
// of strings, such as StatusList.
//
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

//
// AdjustViewMessagePayload holds the information sent by the server's AdjustView
// message. This tells the client to set its viewable area so that its x and y
// scrollbars are at the given proportion of their full range.
//
// Call the AdjustView method to send this message out to other clients.
//
type AdjustViewMessagePayload struct {
	BaseMessagePayload
	XView, YView float64
}

//
// AdjustView tells other clients to adjust their scrollbars
// so that the x and y directions are scrolled to xview and
// yview respectively, where those values are a fraction from
// 0.0 to 1.0 indicating the proportion of the full range in
// each direction.
//
func (c *Connection) AdjustView(xview, yview float64) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AdjustView, AdjustViewMessagePayload{
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

//
// AllowMessagePayload holds the data sent by a client when indicating
// which optional features it supports.
//
type AllowMessagePayload struct {
	BaseMessagePayload

	// List of supported optional feature names
	Features []string `json:",omitempty"`
}

type OptionalFeature byte

const (
	DiceColorBoxes OptionalFeature = iota
)

//
// Allow tells the server which optional features this client is
// prepared to accept.
//
func (c *Connection) Allow(features ...OptionalFeature) error {
	var featureList []string
	if c.Protocol < 333 {
		return nil
	}
	for _, feature := range features {
		switch feature {
		case DiceColorBoxes:
			featureList = append(featureList, "DICE-COLOR-BOXES")
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

//
// AuthMessagePayload holds the data sent by a client when authenticating
// to the server.
//
type AuthMessagePayload struct {
	BaseMessagePayload

	// Client describes the client program (e.g., "mapper 4.0.1")
	Client string `json:",omitempty"`

	// Response gives the binary response to the server's challenge
	Response []byte

	// User gives the username requested by the client
	User string `json:",omitempty"`
}

//   ____           _          _____ _ _
//  / ___|__ _  ___| |__   ___|  ___(_) | ___
// | |   / _` |/ __| '_ \ / _ \ |_  | | |/ _ \
// | |__| (_| | (__| | | |  __/  _| | | |  __/
//  \____\__,_|\___|_| |_|\___|_|   |_|_|\___|
//

//
// CacheFile asks other clients to be sure they retrieve
// and cache the map file with the given server ID.
//
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
	Protocol  int
	Challenge []byte `json:",omitempty"`
}

//   ____ _           _   __  __
//  / ___| |__   __ _| |_|  \/  | ___  ___ ___  __ _  __ _  ___
// | |   | '_ \ / _` | __| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \
// | |___| | | | (_| | |_| |  | |  __/\__ \__ \ (_| | (_| |  __/
//  \____|_| |_|\__,_|\__|_|  |_|\___||___/___/\__,_|\__, |\___|
//                                                   |___/

//
// ChatCommon holds fields common to chat messages and die-roll results.
//
type ChatCommon struct {
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
}

//
// ChatMessageMessagePayload holds the information sent by the server's ChatMessage
// message. This is a message sent by other players or perhaps by the server itself.
//
// Call the ChatMessage, ChatMessageToAll, or ChatMessageToGM methods to send this message out to other clients.
//
type ChatMessageMessagePayload struct {
	BaseMessagePayload
	ChatCommon

	// The text of the chat message we received.
	Text string
}

//
// ChatMessage sends a message on the chat channel to other
// users. The to paramter is a slice of user names of the people
// who should receive this message.
//
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

//
// ChatMessageToAll is equivalent to ChatMessage, but is addressed to all users.
//
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

//
// ChatMessageToGM is equivalent to ChatMessage, but is addressed only to the GM.
//
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

//   ____ _
//  / ___| | ___  __ _ _ __
// | |   | |/ _ \/ _` | '__|
// | |___| |  __/ (_| | |
//  \____|_|\___|\__,_|_|
//

//
// ClearMessagePayload holds the information sent by the server's Clear
// message. This tells the client to remove one or more objects from its
// canvas.
//
// Call the Clear method to send this message out to other clients.
//
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

//
// Clear tells peers to remove objects from their canvases.
// The objID may be one of the following:
//   *                    Remove all objects
//   E*                   Remove all map elements
//   M*                   Remove all monster tokens
//   P*                   Remove all player tokens
//   [<imagename>=]<name> Remove token with given <name>
//   <id>                 Remove object with given <id>
//
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

//
// ClearChatMessagePayload holds the information sent by the server's ClearChat
// message. This tells the client to remove some messages from its chat history.
//
// Call the ClearChat method to send this message out to other clients.
//
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

//
// ClearChat tells peers to remove all messages from their
// chat histories if target is zero. If target>0, then
// all messages with IDs greater than target are removed.
// Otherwise, if target<0 then only the most recent |target|
// messages are kept.
//
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

//
// ClearFromMessagePayload holds the information sent by the server's ClearFrom
// message. This tells the client to remove all elements mentioned in the specified
// map file.
//
// Call the ClearFrom method to send this message out to other clients.
//
type ClearFromMessagePayload struct {
	BaseMessagePayload
	FileDefinition
}

//
// ClearFrom tells all peers to load the map file with the
// given server ID, but to remove from their canvases all
// objects described in the file rather than loading them on.
//
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

//
// CombatModeMessagePayload holds the information sent by the server's CombatMode
// message. This tells the client to enter or exit combat (initiative) mode.
//
// Call the CombatMode method to send this message out to other clients.
//
type CombatModeMessagePayload struct {
	BaseMessagePayload

	// If true, we should be in combat mode.
	Enabled bool
}

//
// CombatMode tells all peers to enable or disable combat mode.
//
func (c *Connection) CombatMode(enabled bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(CombatMode, CombatModeMessagePayload{
		Enabled: enabled,
	})
}

//
// ToolbarMessagePayload holds the information sent by the server's Toolbar
// message. This tells the client to display or hide its toolbar.
//
type ToolbarMessagePayload struct {
	BaseMessagePayload
	Enabled bool
}

//
// Toolbar tells peers to turn on or off their toolbars.
//
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

//
// CommentMessagePayload holds the information sent by the server's Comment
// message. This provides information from the server that the client is
// free to ignore, but may find interesting. Nothing sent in comments is
// critical to the operation of a client. However, some incidental bits
// of information such as an advisement of currently-supported client
// versions and progress gauge data are sent via comments.
//
type CommentMessagePayload struct {
	BaseMessagePayload
	Text string
}

//  ____             _          _
// |  _ \  ___ _ __ (_) ___  __| |
// | | | |/ _ \ '_ \| |/ _ \/ _` |
// | |_| |  __/ | | | |  __/ (_| |
// |____/ \___|_| |_|_|\___|\__,_|
//

//
// DeniedMessagePayload holds the reason the client was denied
// access to the server.
//
type DeniedMessagePayload struct {
	BaseMessagePayload
	Reason string
}

//  _____ _ _ _            ____  _          ____                     _
// |  ___(_) | |_ ___ _ __|  _ \(_) ___ ___|  _ \ _ __ ___  ___  ___| |_ ___
// | |_  | | | __/ _ \ '__| | | | |/ __/ _ \ |_) | '__/ _ \/ __|/ _ \ __/ __|
// |  _| | | | ||  __/ |  | |_| | | (_|  __/  __/| | |  __/\__ \  __/ |_\__ \
// |_|   |_|_|\__\___|_|  |____/|_|\___\___|_|   |_|  \___||___/\___|\__|___/
//

//
// FilterDicePresets asks the server to remove all of your
// die-roll presets whose names match the given regular
// expression.
//
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

//
// FilterDicePresetMessagePayload holds the filter expression
// the client sends to the server.
//
type FilterDicePresetsMessagePayload struct {
	BaseMessagePayload
	Filter string `json:",omitempty"`
	For    string `json:",omitempty"`
}

//
//   ____                 _           _
//  / ___|_ __ __ _ _ __ | |_ ___  __| |
// | |  _| '__/ _` | '_ \| __/ _ \/ _` |
// | |_| | | | (_| | | | | ||  __/ (_| |
//  \____|_|  \__,_|_| |_|\__\___|\__,_|
//

//
// GrantedMessagePayload holds the response from the server
// informing the client that its access was granted.
//
type GrantedMessagePayload struct {
	BaseMessagePayload
	User string
}

//  _                    _ _____
// | |    ___   __ _  __| |  ___| __ ___  _ __ ___
// | |   / _ \ / _` |/ _` | |_ | '__/ _ \| '_ ` _ \
// | |__| (_) | (_| | (_| |  _|| | | (_) | | | | | |
// |_____\___/ \__,_|\__,_|_|  |_|  \___/|_| |_| |_|
//

//
// LoadFromMessagePayload holds the information sent by the server's LoadFrom
// message. This tells the client to open the file named (which may either be
// a local disk file or one retrieved from the server), and either replacing their
// current canvas contents with the elements from that file, or adding those
// elements to the existing contents.
//
// Call the LoadFrom method to send this message out to other clients.
//
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

//
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
//
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

//
// LoadArcObjectMessagePayload holds the information needed to send an arc element to a map.
//
type LoadArcObjectMessagePayload struct {
	BaseMessagePayload
	ArcElement
}

//
// LoadCircleObjectMessagePayload holds the information needed to send an ellipse element to a map.
//
type LoadCircleObjectMessagePayload struct {
	BaseMessagePayload
	CircleElement
}

//
// LoadLineObjectMessagePayload holds the information needed to send a line element to a map.
//
type LoadLineObjectMessagePayload struct {
	BaseMessagePayload
	LineElement
}

//
// LoadPolygonObjectMessagePayload holds the information needed to send a polygon element to a map.
//
type LoadPolygonObjectMessagePayload struct {
	BaseMessagePayload
	PolygonElement
}

//
// LoadRectangleObjectMessagePayload holds the information needed to send a rectangle element to a map.
//
type LoadRectangleObjectMessagePayload struct {
	BaseMessagePayload
	RectangleElement
}

//
// LoadSpellAreaOfEffectObjectMessagePayload holds the information needed to send a spell area of effect element to a map.
//
type LoadSpellAreaOfEffectObjectMessagePayload struct {
	BaseMessagePayload
	SpellAreaOfEffectElement
}

//
// LoadTextObjectMessagePayload holds the information needed to send a text element to a map.
//
type LoadTextObjectMessagePayload struct {
	BaseMessagePayload
	TextElement
}

//
// LoadTileObjectMessagePayload holds the information needed to send a graphic tile element to a map.
//
type LoadTileObjectMessagePayload struct {
	BaseMessagePayload
	TileElement
}

//
// LoadObject sends a MapObject to all peers.
// It may be given a value of any of the supported MapObject
// types for map graphic elements (Arc, Circle, Line, Polygon,
// Rectangle, SpellAreaOfEffect, Text, or Tile).
//
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

//
// MarcoMessagePayload holds the information sent by the server's Marco
// message. This is a "ping" message the server periodically sends to all
// clients to ensure they are still responding. A client who receives a
// MARCO message is expected to respond with a POLO message.
//
// If the client doesn't subscribe to Marco messages, the Dial method
// will automatically reply with Polo messages.
//
type MarcoMessagePayload struct {
	BaseMessagePayload
}

//  __  __            _
// |  \/  | __ _ _ __| | __
// | |\/| |/ _` | '__| |/ /
// | |  | | (_| | |  |   <
// |_|  |_|\__,_|_|  |_|\_\
//

//
// MarkMessagePayload holds the information sent by the server's Mark
// message. This tells the client to
// visually mark the given map coordinates.
//
// Call the Mark method to send this message out to other clients.
//
type MarkMessagePayload struct {
	BaseMessagePayload
	Coordinates
}

//
// Mark tells clients to visibly mark a location centered
// on the given (x, y) coordinates.
//
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
//
type PlaceSomeoneMessagePayload struct {
	BaseMessagePayload
	CreatureToken
}

//
// PlaceSomeone tells all peers to add a new creature token on their
// maps. The parameter passed must be either a PlayerToken or MonsterToken.
//
// If the creature is already on the map, it will be replaced by the
// new one being presented here. Thus, PlaceSomeone may be used to change
// the name or location of an existing creature, although the preferred
// way to do that would be to use UpdateObjAttributes to change those
// specific attributes of the creature directly.
//
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

//  ____       _
// |  _ \ ___ | | ___
// | |_) / _ \| |/ _ \
// |  __/ (_) | | (_) |
// |_|   \___/|_|\___/
//

//
// Polo send the client's response to the server's MARCO ping message.
//
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

//
// QueryImageMessagePayload holds the information sent by the server's QueryImage
// message. This tells the client
// that a peer wants to know where to find a given
// image and the server didn't know either. If you know the definition
// for that image, reply with an AddImage message of your own.
//
// Call the QueryImage method to send this message out to other clients.
//
type QueryImageMessagePayload struct {
	BaseMessagePayload
	ImageDefinition
}

//
// QueryImage asks the server and peers if anyone else knows
// where to find the data for the given image name and zoom factor.
// If someone does, you'll receive an AddImage message.
//
func (c *Connection) QueryImage(idef ImageDefinition) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryImage, idef)
}

//  ____                _
// |  _ \ ___  __ _  __| |_   _
// | |_) / _ \/ _` |/ _` | | | |
// |  _ <  __/ (_| | (_| | |_| |
// |_| \_\___|\__,_|\__,_|\__, |
//                        |___/
//

//
// ReadyMessagePayload indicates that the server is fully
// ready to interact with the client and all preliminary
// data has been sent to the client.
//
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

//
// RemoveObjAttributesMessagePayload holds the information sent by the server's RemoveObjAttributes
// message. This tells the client
// to adjust the multi-value attribute
// of the object with the given ID by removing the listed values
// from it.
//
// Call the RemoveObjAttributes method to send this message out to other clients.
//
type RemoveObjAttributesMessagePayload struct {
	BaseMessagePayload

	// The ID of the object to be modified
	ObjID string

	// The name of the attribute to modify. Must be one with a []string value.
	AttrName string

	// The values to remove from the attribute.
	Values []string
}

//
// RemoveObjAttributes informs peers to remove a set of string values from the existing
// value of an object attribute. The attribute must be one whose value is a list
// of strings, such as StatusList.
//
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
func (c *Connection) RollDice(to []string, rollspec string) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(RollDice, RollDiceMessagePayload{
		ChatCommon: ChatCommon{
			Recipients: to,
		},
		RollSpec: rollspec,
	})
}

//
// RollDiceMessagePayload holds the data sent from the client to the
// server when requesting a die roll.
//
type RollDiceMessagePayload struct {
	BaseMessagePayload
	ChatCommon

	// RollSpec describes the dice to be rolled and any modifiers.
	RollSpec string
}

//
// RollDiceToAll is equivalent to RollDice, sending the results to all users.
//
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

//
// RollDiceToGM is equivalent to RollDice, sending the results only to the GM.
//
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

//
// RollResultMessagePayload holds the information sent by the server's RollResult
// message. This tells the client the results of a die roll.
//
type RollResultMessagePayload struct {
	BaseMessagePayload
	ChatCommon

	// The title describing the purpose of the die-roll, as set by the user.
	Title string `json:",omitempty"`

	// The die roll result and details behind where it came from.
	Result dice.StructuredResult
}

//  ____  _          ____                     _
// |  _ \(_) ___ ___|  _ \ _ __ ___  ___  ___| |_ ___
// | | | | |/ __/ _ \ |_) | '__/ _ \/ __|/ _ \ __/ __|
// | |_| | | (_|  __/  __/| | |  __/\__ \  __/ |_\__ \
// |____/|_|\___\___|_|   |_|  \___||___/\___|\__|___/
//

//
// DefineDicePresets replaces any existing die-roll presets you have
// stored on the server with the new set passed as the presets parameter.
//
func (c *Connection) DefineDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(DefineDicePresets, DefineDicePresetsMessagePayload{
		Presets: presets,
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
	For     string               `json:",omitempty"`
	Presets []dice.DieRollPreset `json:",omitempty"`
}

//
// AddDicePresets is like DefineDicePresets except that it adds the presets
// passed in to the existing set rather than replacing them.
//
func (c *Connection) AddDicePresets(presets []dice.DieRollPreset) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(AddDicePresets, AddDicePresetsMessagePayload{
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
	For     string               `json:",omitempty"`
	Presets []dice.DieRollPreset `json:",omitempty"`
}

//
// QueryDicePresets requests that the server send you the die-roll
// presets currently stored for you. It will send you an UpdateDicePresets
// message.
//
func (c *Connection) QueryDicePresets() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryDicePresets, nil)
}

type QueryDicePresetsMessagePayload struct {
	BaseMessagePayload
	For string `json:",omitempty"`
}

//
// UpdateClockMessagePayload holds the information sent by the server's UpdateClock
// message. This tells the client to update its clock display to the new value.
//
type UpdateClockMessagePayload struct {
	BaseMessagePayload

	// The clock is now at the given absolute number of
	// seconds from the GMA clock's global epoch.
	Absolute float64

	// The elapsed time counter is now this many seconds from
	// some reference point set by the GM (often the start of
	// combat).
	Relative float64

	// If true and not in combat mode, local clients should
	// keep running the clock in real time.
	Running bool
}

//
// UpdateClock informs everyone of the current time
//
func (c *Connection) UpdateClock(absolute, relative float64, keepRunning bool) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateClock, UpdateClockMessagePayload{
		Absolute: absolute,
		Relative: relative,
		Running:  keepRunning,
	})
}

//
// UpdateDicePresetsMessagePayload holds the information sent by the server's UpdateDicePresets
// message. This tells the client to
// accept the die-roll presets
// described here, replacing any previous presets it was
// using.
//
type UpdateDicePresetsMessagePayload struct {
	BaseMessagePayload
	Presets []dice.DieRollPreset
}

//
// UpdateInitiativeMessagePayload holds the information sent by the server's UpdateInitiative
// message. This tells the client that the initiative order has been changed. Its current
// notion of the initiative order should be replaced by the one given here.
//
type UpdateInitiativeMessagePayload struct {
	BaseMessagePayload
	InitiativeList []InitiativeSlot
}

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

//
// InitiativeSlot describes the creature occupying a given
// slot of the initiative list.
//
type InitiativeSlot struct {
	// The slot number (currently 0–59, corresponding to the 1/10th second "count" in the initiative round)
	Slot int

	// The current hit point total for the creature.
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

//
// UpdateObjAttributesMessagePayload holds the information sent by the server's UpdateObjAttributes
// message. This tells the client to update an existing object
// with new attributes. Any attributes not listed here should
// remain intact.
//
// Call the UpdateObjAttributes method to send this message out to other clients.
//
type UpdateObjAttributesMessagePayload struct {
	BaseMessagePayload

	// The ID of the object to be modified.
	ObjID string

	// A map of attribute name to its new value.
	NewAttrs map[string]any
}

//
// UpdateObjAttributes informs peers that they should modify the
// specified object's attributes which are mentioned in the newAttrs
// map. This maps attribute names to their new values.
//
func (c *Connection) UpdateObjAttributes(objID string, newAttrs map[string]any) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateObjAttributes, UpdateObjAttributesMessagePayload{
		ObjID:    objID,
		NewAttrs: newAttrs,
	})
}

//
// UpdatePeerListMessagePayload holds the information sent by the server's UpdatePeerList
// message. This tells the client that the list of
// other connected peers has changed.
//
type UpdatePeerListMessagePayload struct {
	BaseMessagePayload
	PeerList []Peer
}

//
// Peer describes each peer we can reach via our server connection.
//
type Peer struct {
	// IP address and port of the peer
	Addr string

	// The username provided by the peer when it authenticated
	User string

	// A description of the peer client program (provided by that client)
	Client string `json:",omitempty"`

	// How many seconds ago the peer last answered a "still alive?" ping from the server
	LastPolo float64

	// True if the client authenticated successfully
	IsAuthenticated bool `json:",omitempty"`

	// True if this structure describes the connection of this client program
	IsMe bool `json:",omitempty"`
}

//
// QueryPeers asks the server to send an UpdatePeerList
// message with the current set of peers who are connected
// to the server.
//
func (c *Connection) QueryPeers() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(QueryPeers, nil)
}

type QueryPeersMessagePayload struct {
	BaseMessagePayload
}

//
// UpdateProgressMessagePayload holds the information sent by the server's UpdateProgress
// Comment notification. This
// advises the client of the status of an operation
// in progress. The client may wish to display a progress indicator to the
// user.
//
type UpdateProgressMessagePayload struct {
	BaseMessagePayload

	// If true, we can dispose of the tracked operation
	// and should not expect further updates about it.
	IsDone bool `json:",omitempty"`

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
}

//
// UpdateStatusMarkerMessagePayload holds the information sent by the server's UpdateStatusMarker
// message. This tells the client
// to add or change a status marker which may be placed
// on creature tokens.
//
// Note: the server usually sends these upon login, which the Connection
// struct stores internally.
//
type UpdateStatusMarkerMessagePayload struct {
	BaseMessagePayload
	StatusMarkerDefinition
}

//
// UpdateStatusMarker changes, removes, or adds a status marker to place on
// a creature marker.
//
func (c *Connection) UpdateStatusMarker(smd StatusMarkerDefinition) error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(UpdateStatusMarker, smd)
}

//
// StatusMarkerDefinition describes each creature token status
// that the map clients indicate.
//
type StatusMarkerDefinition struct {
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

//
// Text produces a simple text description of a StatusMarkerDefinition structure.
//
func (c StatusMarkerDefinition) Text() string {
	return fmt.Sprintf("Condition %q: Shape=%q, Color=%q, Description=%q", c.Condition, c.Shape, c.Color, c.Description)
}

//
// StatusMarkerDefinitions is a map of a condition code name to the full
// description of the marker to use for that condition.
//
type StatusMarkerDefinitions map[string]StatusMarkerDefinition

//
// CharacterDefinitions is a map of a character name to their token object.
//
type CharacterDefinitions map[string]PlayerToken

// Text produces a simple text description of a map of PlayerTokens
func (cs CharacterDefinitions) Text() string {
	var s strings.Builder
	for k, c := range cs {
		fmt.Fprintf(&s, "[%s] %v\n", k, c)
	}
	return s.String()
}

//
// Text produces a simple text description of a map of StatusMarkerDefinitions
// as a multi-line string.
//
func (cs StatusMarkerDefinitions) Text() string {
	var s strings.Builder
	for k, c := range cs {
		fmt.Fprintf(&s, "[%s] %s\n", k, c.Text())
	}
	return s.String()
}

//
// UpdateTurnMessagePayload holds the information sent by the server's UpdateTurn
// message. This tells the client whose turn it is in combat.
//
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

//
// UpdateTurn advances the initiative turn clock for connected clients.
//
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

//
// Sync requests that the server send the entire game state
// to it.
//
func (c *Connection) Sync() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	return c.serverConn.Send(Sync, nil)
}

type SyncMessagePayload struct {
	BaseMessagePayload
}

//
// SyncChat requests that the server (re-)send past messages
// greater than the target message ID (target≥0) or the most
// recent |target| messages (target<0).
//
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
	Packages []PackageUpdate `json:",omitempty"`
}

type PackageUpdate struct {
	Name      string
	Instances []PackageVersion
}

type PackageVersion struct {
	OS      string `json:",omitempty"`
	Arch    string `json:",omitempty"`
	Version string
	Token   string `json:",omitempty"`
}

type WorldMessagePayload struct {
	BaseMessagePayload
	Calendar string
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
//   ctx, cancel := context.Background()
//   server, err := NewConnection("example.org:2323",
//                                WithAuthenticator(a),
//                                WithContext(ctx))
//   defer cancel()
//   go server.Dial()
//
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
	incomingPacket := c.serverConn.Receive(done)
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
		incomingPacket := c.serverConn.Receive(done)
		if incomingPacket == nil {
			break
		}

		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}

		// Protocol sequence:
		// <- PROTOCOL v
		// <- AC, DSM, UPDATES, WORLD, // messages
		// <- OK
		// -> AUTH (if authentication required)
		// <- GRANTED or DENIED (if AUTH required and given)
		// <- AC, DSM, UPDATES, WORLD, // messages
		// <- READY

		switch response := incomingPacket.(type) {
		case ChallengeMessagePayload:
			// OK Protocol=v [Challenge=data]
			if response.Protocol != c.Protocol {
				c.Logf("server advertised protocol %v initially but then claimed version %v", c.Protocol, response.Protocol)
				done <- fmt.Errorf("server can't make up its mind whether it uses protocol %v or %v", c.Protocol, response.Protocol)
				return
			}

			if response.Challenge != nil {
				if c.Authenticator == nil {
					c.Log("Server requires authentication but no authenticator was provided for the client.")
					done <- ErrAuthenticationRequired
					return
				}
				c.Log("authenticating to server")
				c.Authenticator.Reset()
				authResponse, err := c.Authenticator.AcceptChallengeBytes(response.Challenge)
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

		case UpdateStatusMarkerMessagePayload:
			c.receiveDSM(response)

		case WorldMessagePayload:
			c.CalendarSystem = response.Calendar

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
		incomingPacket := c.serverConn.Receive(done)
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
		incomingPacket := c.serverConn.Receive(done)
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

		case WorldMessagePayload:
			c.CalendarSystem = response.Calendar

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
}

func (c *Connection) receiveDSM(d UpdateStatusMarkerMessagePayload) {
	if c != nil {
		c.Conditions[d.Condition] = StatusMarkerDefinition{
			Condition:   d.Condition,
			Shape:       d.Shape,
			Color:       d.Color,
			Description: d.Description,
		}
	}
}

func (c *Connection) receiveAddCharacter(d AddCharacterMessagePayload) {
	if c != nil {
		c.Characters[d.Name] = PlayerToken{
			CreatureToken: CreatureToken{
				BaseMapObject: BaseMapObject{
					ID: d.ObjID(),
				},
				Name:  d.Name,
				Color: d.Color,
				Area:  d.Area,
				Size:  d.Size,
			},
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

//
// listen for, and dispatch, incoming server messages
//
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
		incomingPacket := c.serverConn.Receive(done)
		if incomingPacket == nil {
			break
		}

		if (c.DebuggingLevel & DebugBinary) != 0 {
			c.debug(DebugBinary, util.Hexdump(incomingPacket.RawBytes()))
		}

		switch cmd := incomingPacket.(type) {
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

		case PlaceSomeoneMessagePayload:
			if ch, ok := c.Subscriptions[PlaceSomeone]; ok {
				ch <- cmd
			}

		case PrivMessagePayload:
			if ch, ok := c.Subscriptions[Priv]; ok {
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

		case AddCharacterMessagePayload, ChallengeMessagePayload, DeniedMessagePayload,
			GrantedMessagePayload, ProtocolMessagePayload, ReadyMessagePayload,
			UpdateVersionsMessagePayload, WorldMessagePayload:

			c.reportError(fmt.Errorf("message type %v should not be sent to client at this stage in the session", cmd.MessageType()))

		case AcceptMessagePayload, AddDicePresetsMessagePayload, AllowMessagePayload,
			AuthMessagePayload, DefineDicePresetsMessagePayload,
			FilterDicePresetsMessagePayload, PoloMessagePayload,
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

//
// report any sort of error to the client
//
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

//
// listen and interact with the service until it's finished,
// then close our connection to it
//
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
			c.serverConn.sendBuf = append(c.serverConn.sendBuf, packet)
		default:
		}
		//
		// Send the next outgoing message in the buffer
		//
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
	}
}

//
// Any time the subscription list changes,
// we need to call this to let the server know what kinds of
// messages the client wants to see.
//
func (c *Connection) filterSubscriptions() error {
	if c == nil {
		return fmt.Errorf("nil Connection")
	}
	if !c.IsReady() {
		return nil
	}

	subList := []string{"MARCO", "PRIV"} // these are unconditional
	for msg := range c.Subscriptions {
		switch msg {
		//Accept (client)
		//AddCharacter (forbidden)
		//AddDicePresets (client)
		//Allow (client)
		//Auth (client)
		//Challenge (forbidden)
		//DefineDicePresets (client)
		//Denied (forbidden)
		//FilterDicePresets (client)
		//Granted (forbidden)
		//Marco (mandatory)
		//Polo (client)
		//Priv (mandatory)
		//Protocol (forbidden)
		//QueryDicePresets (client)
		//QueryPeers (client)
		//Ready (forbidden)
		//RollDice (client)
		//Sync (client)
		//SyncChat (client)
		//UpdateVersions (forbidden)
		//World (forbidden)

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
		case QueryImage:
			subList = append(subList, "AI?")
		case RemoveObjAttributes:
			subList = append(subList, "OA-")
		case RollResult:
			subList = append(subList, "ROLL")
		case Toolbar:
			subList = append(subList, "TB")
		case UpdateClock:
			subList = append(subList, "CS")
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

//
// CheckVersionOf returns the closest match of the requested package
// to the platform we are currently running, or nil if we're already
// on the advertised version.
//
func (c *Connection) CheckVersionOf(packageName, myVersionNumber string) (*PackageVersion, error) {
	var availableVersion *PackageVersion

	candidates, ok := c.PackageUpdatesAvailable[packageName]
	if !ok {
		return nil, fmt.Errorf("The server provided no upgrade information for package \"%s\"", packageName)
	}
	for _, candidate := range candidates {
		if (candidate.OS == "" || candidate.OS == runtime.GOOS) && (candidate.Arch == "" || candidate.Arch == runtime.GOARCH) {
			if availableVersion.Version != "" && ((candidate.OS != "" && availableVersion.OS == "") || (candidate.Arch != "" && availableVersion.Arch == "")) {
				// found a more specific match, use that instead
				availableVersion = &candidate
			} else if availableVersion.Version == "" {
				availableVersion = &candidate
			}
		}
	}

	return availableVersion, nil
}

// @[00]@| GMA 5.0.0
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
