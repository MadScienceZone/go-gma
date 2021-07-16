/*
########################################################################################
#  _______  _______  _______                ___       ______      _______              #
# (  ____ \(       )(  ___  )              /   )     / ___  \    (  ____ \             #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   | (    \/             #
# | |      | || || || (___) |            / (_) (_       ___) /   | (____               #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (    (_____ \              #
# | | \_  )| |   | || (   ) | Game           ) (           ) \         ) )             #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _ /\____) )             #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_)\______/              #
#                                                                                      #
########################################################################################
*/

//
// Client interface for the mapper service.
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
// calling the Dial() method in this package. This function will
// sign on to the server and then enter a loop, sending incoming
// server messages back on the channel(s) established via the
// Subscribe() method. Dial() returns when the session with the
// server has terminated.
//
// Typically, an application will invoke the Dial() method in a
// goroutine. Calling the associated context's cancel function
// will signal that we want to stop talking to the server, resulting
// in the termination of the running Dial() method.
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
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/fizban-of-ragnarok/go-gma/v4/auth"
	"github.com/fizban-of-ragnarok/go-gma/v4/dice"
	"github.com/fizban-of-ragnarok/go-gma/v4/tcllist"
	"github.com/fizban-of-ragnarok/go-gma/v4/util"
)

//
// The GMA Mapper Protocol version number current as of this build,
// and protocol versions supported by this code.
//
const (
	GMAMapperProtocol=332     // @@##@@ auto-configured
	GMAVersionNumber="4.3.5" // @@##@@ auto-configured
	MINIMUM_SUPPORTED_MAP_PROTOCOL = 332
	MAXIMUM_SUPPORTED_MAP_PROTOCOL = 332
)

func init() {
	if MINIMUM_SUPPORTED_MAP_PROTOCOL > GMAMapperProtocol || MAXIMUM_SUPPORTED_MAP_PROTOCOL < GMAMapperProtocol {
		if MINIMUM_SUPPORTED_MAP_PROTOCOL == MAXIMUM_SUPPORTED_MAP_PROTOCOL {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapclient only supports mapper protocol %v, but version %v was the official one when this package was released!", MINIMUM_SUPPORTED_MAP_PROTOCOL, GMAMapperProtocol))
		} else {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapclient only supports mapper protocols %v-%v, but version %v was the official one when this package was released!", MINIMUM_SUPPORTED_MAP_PROTOCOL, MAXIMUM_SUPPORTED_MAP_PROTOCOL, GMAMapperProtocol))
		}
	}
}

// This is the error returned when the server requires authentication but we didn't provide any.
var AuthenticationRequired = errors.New("authenticator required for connection")

// This is the error returned when our authentication was rejected by the server.
var AuthenticationFailed = errors.New("access denied to server")

//
// A connection to the server is described by the Connection
// type.
//
type Connection struct {
	// The context for our session, either one we created in the
	// NewConnection() function or one we received from the caller.
	Context context.Context

	// The server endpoint, in any form acceptable to the net.Dial()
	// function.
	Endpoint string

	// If this is non-nil, we will use this to identify the user
	// to the server.
	Authenticator *auth.Authenticator

	// Server message subscriptions currently in effect.
	Subscriptions map[ServerMessage]chan MessagePayload

	// If nonzero, we will re-try a failing connection this many
	// times before giving up on the server. Otherwise we will keep
	// trying forever.
	Retries uint

	// If nonzero, our connection attempts will timeout after the
	// specified time interval. Otherwise they will wait indefinitely.
	Timeout time.Duration

	// We will log informational messages here as we work.
	Logger *log.Logger

	// The server's protocol version number.
	Protocol int

	// Characters received from the server.
	Characters map[string]CharacterDefinition

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

	// The last error encountered while communicating with the server.
	LastError error

	// The verbosity level of debugging log messages.
	DebuggingLevel uint

	server   net.Conn       // network socket to the server
	reader   *bufio.Scanner // read interface to server
	writer   *bufio.Writer  // write interface to server
	sendChan chan string    // outgoing packets go through this channel
	sendBuf  []string       // internal buffer of outgoing packets
	signedOn bool           // do we have an active session now?

	// If true, we will always try to reconnect to the server if we
	// lose our connection.
	StayConnected bool
}

//
// IsReady returns true if the connection to the server
// has completed and authentication was successful, so
// the connection is ready for interactive use.
//
func (c *Connection) IsReady() bool {
	return c.reader != nil && c.writer != nil && c.signedOn
}

//
// A CharacterDefinition describes a PC known as a regular player of the
// game system.
//
type CharacterDefinition struct {
	// Character name as appears on the map.
	Name string

	// ObjID to use for the character rather than generating one.
	ObjID string

	// Color to use to draw the threat zone around the character.
	Color string

	// Size codes of the threat area and creature token size.
	Area, Size string
}

//
// Text produces a simple text description of the receiving CharacterDefinition value.
//
func (c CharacterDefinition) Text() string {
	return fmt.Sprintf("ID %q; Name %q; Zone color %q; Size %q; Threat Space %q",
		c.ObjID, c.Name, c.Color, c.Size, c.Area)
}

type CharacterDefinitions map[string]CharacterDefinition

//
// Text describes the PCs in the receiving slice of CharacterDefinitions in a
// simple multi-line text form.
//
func (cs CharacterDefinitions) Text() string {
	var s strings.Builder
	for k, c := range cs {
		fmt.Fprintf(&s, "[%s] %s\n", k, c.Text())
	}
	return s.String()
}

//
// Options which can be added to the NewConnection() function.
//
type connectionOption func(conn *Connection) error

//
// WithContext modifies the behavior of the NewConnection() function
// by supplying a context for this connection, which may be used to
// signal the Dial() method that the connection to the server should
// be terminated.
//
// N.B.: When making the initial TCP connection to the server,
// if there is a timeout value specified via WithTimeout(), then
// a hanging connection will terminate when that timer expires,
// regardless of the context. Otherwise, the connection will wait
// indefinitely to complete OR until the context is cancelled.
//
func WithContext(ctx context.Context) connectionOption {
	return func(c *Connection) error {
		c.Context = ctx
		return nil
	}
}

//
// WithSubscription modifies the behavior of the NewConnection() function
// by adding a server message subscription to the connection just as if
// the Subscribe() method had been called on the connection value.
//
func WithSubscription(ch chan MessagePayload, messages ...ServerMessage) connectionOption {
	return func(c *Connection) error {
		return c.Subscribe(ch, messages...)
	}
}

//
// WithAuthenticator modifies the behavior of the NewConnection() function
// by adding an authenticator which will be used to identify the client
// to the server.
//
func WithAuthenticator(a *auth.Authenticator) connectionOption {
	return func(c *Connection) error {
		c.Authenticator = a
		return nil
	}
}

//
// WithLogger modifies the behavior of the NewConnection() function
// by specifying a custom logger instead of the default one.
//
func WithLogger(l *log.Logger) connectionOption {
	return func(c *Connection) error {
		c.Logger = l
		return nil
	}
}

//
// WithTimeout specifies the time to allow when making the TCP
// connection to the server. After this time expires, the attempt
// is abandoned (but may be retried based on the value of
// WithRetries(), if any).
//
// N.B.: When making the initial TCP connection to the server,
// if there is a timeout value specified via WithTimeout(), then
// a hanging connection will terminate when that timer expires,
// regardless of the context (although a canceled context will
// stop retry attempts). Otherwise, the connection will wait
// indefinitely to complete OR until the context is cancelled.
//
func WithTimeout(t time.Duration) connectionOption {
	return func(c *Connection) error {
		c.Timeout = t
		return nil
	}
}

//
// WithRetries modifies the connection so that failed attempts
// to make the TCP connection to the server will be tried again
// the given number of times.
//
// Setting this to 0 means to retry infinitely many times.
//
func WithRetries(n uint) connectionOption {
	return func(c *Connection) error {
		c.Retries = n
		return nil
	}
}

//
// StayConnected modifies the connection so that the Dial()
// method will never return until the context is cancelled
// or it failed to contact the server at all.
//
// If the parameter is true, the stay-connected mode is enabled;
// if false, it is disabled (the default).
//
// In other words, with this option in effect, if the server's
// connection is lost, Dial() will simply try to reconnect and
// continue operations.
//
func StayConnected(enable bool) connectionOption {
	return func(c *Connection) error {
		c.StayConnected = enable
		return nil
	}
}

//
// WithDebugging modifies the connection so that its operations
// are logged to varying levels of verbosity:
//   0 - no extra logging
//   1 - each incoming/outgoing message is logged
//   2 - more internal state is exposed
//   3 - the full data for each incoming/outgoing message is logged
//
func WithDebugging(level uint) connectionOption {
	return func(c *Connection) error {
		c.DebuggingLevel = level
		return nil
	}
}

//
// Create a new server connection value which can then be used to
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
func NewConnection(endpoint string, opts ...connectionOption) (Connection, error) {
	newCon := Connection{
		Context:       context.Background(),
		Endpoint:      endpoint,
		Subscriptions: make(map[ServerMessage]chan MessagePayload),
		Characters:    make(map[string]CharacterDefinition),
		Conditions:    make(map[string]StatusMarkerDefinition),
		Retries:       1,
		Logger:        log.Default(),
		sendChan:      make(chan string, 16),
		Gauges:        make(map[string]*UpdateProgressMessagePayload),
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
func (c *Connection) debug(level uint, msg string) {
	if c.DebuggingLevel >= level {
		for i, line := range strings.Split(msg, "\n") {
			if line != "" {
				c.Logger.Printf("DEBUG%d.%02d: %s", level, i, line)
			}
		}
	}
}

//
// Close terminates the connection to the server.
// Note that the Dial() function normally closes the connection
// before it returns, so calling this explicitly should not
// normally be necessary.
//
// Calling Close() will result in the Dial() function stopping
// due to the connection disappearing, but it is better to cancel
// the context being watched by Dial() instead.
//
func (c *Connection) Close() {
	c.debug(1, "Close()")
	c.server.Close()
}

//
// Subscribe arranges for server messages to be sent to the specified channel
// when they arrive.
//
// If multiple messages are specified, they are all directed to send their payloads
// to the channel, which may used the MessageType() method to differentiate what
// kind of payload was sent.
//
// This method may be called multiple times for the same channel, in which case
// the specified message(s) are added to the set which sends to that channel.
//
// If another Subscribe() method is called with the same ServerMessage that a
// previous Subscribe() mentioned, that will change the subscription for that
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
//   cm := make(chan<-MessagePayload, 1)
//   service, err := NewConnection(endpoint)
//   err = service.Subscribe(cm, ChatMessage)
//
func (conn *Connection) Subscribe(ch chan MessagePayload, messages ...ServerMessage) error {
	for _, m := range messages {
		if m >= maximumServerMessage {
			return fmt.Errorf("server message ID %v not defined (illegal Subscribe call)", m)
		}
		if ch == nil {
			delete(conn.Subscriptions, m)
		} else {
			conn.Subscriptions[m] = ch
		}
	}
	return conn.filterSubscriptions()
}

//
// Incoming server messages are described by a structure of
// a type implementing MessagePayload.
//
type MessagePayload interface {
	MessageType() ServerMessage
	RawMessage() []string
}

type ServerMessage byte

//
// These are the server messages to which a client may subscribe.
//
const (
	AddCharacter = iota
	AddImage
	AddObjAttributes
	AdjustView
	ChatMessage
	Clear
	ClearChat
	ClearFrom
	CombatMode
	Comment
	LoadFrom
	LoadObject
	Marco
	Mark
	PlaceSomeone
	QueryImage
	RemoveObjAttributes
	RollResult
	Toolbar
	UpdateClock
	UpdateDicePresets
	UpdateInitiative
	UpdateObjAttributes
	UpdatePeerList
	UpdateProgress
	UpdateStatusMarker
	UpdateTurn
	UNKNOWN
	ERROR
	maximumServerMessage
)

//
// The bare minimum payload for any server message.
//
type BaseMessagePayload struct {
	rawMessage  []string
	messageType ServerMessage
}

func (p BaseMessagePayload) RawMessage() []string       { return p.rawMessage }
func (p BaseMessagePayload) MessageType() ServerMessage { return p.messageType }

//
// An error encountered when trying to receive a message.
//
type ErrorMessagePayload struct {
	BaseMessagePayload
	OriginalMessageType ServerMessage
	Error               error
}

//
// We received a server message but have no idea what it is.
//
type UnknownMessagePayload struct {
	BaseMessagePayload
}

//________________________________________________________________________________
//     _       _     _  ____ _                          _
//    / \   __| | __| |/ ___| |__   __ _ _ __ __ _  ___| |_ ___ _ __
//   / _ \ / _` |/ _` | |   | '_ \ / _` | '__/ _` |/ __| __/ _ \ '__|
//  / ___ \ (_| | (_| | |___| | | | (_| | | | (_| | (__| ||  __/ |
// /_/   \_\__,_|\__,_|\____|_| |_|\__,_|_|  \__,_|\___|\__\___|_|
//

//
// AddCharacter message: add a PC to the party
//
type AddCharacterMessagePayload struct {
	BaseMessagePayload
	CharacterDefinition
}

/* Clients may not send these */

//________________________________________________________________________________
//     _       _     _ ___
//    / \   __| | __| |_ _|_ __ ___   __ _  __ _  ___
//   / _ \ / _` |/ _` || || '_ ` _ \ / _` |/ _` |/ _ \
//  / ___ \ (_| | (_| || || | | | | | (_| | (_| |  __/
// /_/   \_\__,_|\__,_|___|_| |_| |_|\__,_|\__, |\___|
//                                         |___/

//
// AddImage message: the client should locally note the definition
// of an image for later reference.
//
type AddImageMessagePayload struct {
	BaseMessagePayload

	// The image definition received from the server.
	ImageDefinition

	// If non-nil, this holds the image data received directly
	// from the server. This usage is not recommended but still
	// supported. In this case the "File" member of the ImageDefinition
	// will be empty.
	ImageData []byte
}

//
// AddImage informs the server and peers about an image they can use.
//
func (c *Connection) AddImage(idef ImageDefinition) error {
	if idef.IsLocalFile {
		return fmt.Errorf("sending local files is not supported. Upload image to server first")
	}
	return c.send("AI@", idef.Name, idef.Zoom, idef.File)
}

//
// AddImageData sends binary image data to peer clients.
//
// Generally, it is better to store an image file on the server, and
// then call AddImage() to point others to find the image there, since
// that will be more efficient than sending it through the mapper
// protocol.
//
func (c *Connection) AddImageData(idef ImageDefinition, data []byte) error {
	var dataBlocks []string
	encoded := base64.StdEncoding.EncodeToString(data)
	for len(encoded) > 100 {
		dataBlocks = append(dataBlocks, encoded[0:100])
		encoded = encoded[100:]
	}
	dataBlocks = append(dataBlocks, encoded)

	if err := c.send("AI", idef.Name, idef.Zoom); err != nil {
		_ = c.send("AI.", 0) // best effort but doesn't matter if this actually goes through
		return err
	}
	for _, dataBlock := range dataBlocks {
		if err := c.send("AI:", dataBlock); err != nil {
			_ = c.send("AI.", 0)
			return err
		}
	}
	return c.send("AI.", len(dataBlocks), streamChecksumStrings(dataBlocks))
}

//     _       _     _  ___  _     _    _   _   _        _ _           _
//    / \   __| | __| |/ _ \| |__ (_)  / \ | |_| |_ _ __(_) |__  _   _| |_ ___  ___
//   / _ \ / _` |/ _` | | | | '_ \| | / _ \| __| __| '__| | '_ \| | | | __/ _ \/ __|
//  / ___ \ (_| | (_| | |_| | |_) | |/ ___ \ |_| |_| |  | | |_) | |_| | ||  __/\__ \
// /_/   \_\__,_|\__,_|\___/|_.__// /_/   \_\__|\__|_|  |_|_.__/ \__,_|\__\___||___/
//                              |__/

//
// AddObjAttributes message: Adjust the multi-value attribute
// of the object with the given ID by adding the new values
// to it.
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
// of strings, such as STATUSLIST.
//
func (c *Connection) AddObjAttributes(objID, attrName string, values []string) error {
	return c.send("OA+", objID, strings.ToUpper(attrName), values)
}

//     _       _  _           _ __     ___
//    / \   __| |(_)_   _ ___| |\ \   / (_) _____      __
//   / _ \ / _` || | | | / __| __\ \ / /| |/ _ \ \ /\ / /
//  / ___ \ (_| || | |_| \__ \ |_ \ V / | |  __/\ V  V /
// /_/   \_\__,_|/ |\__,_|___/\__| \_/  |_|\___| \_/\_/
//             |__/

//
// AdjustView message: Change your displayed map view to the given
// fractions of the full canvas size.
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
	return c.send("AV", xview, yview)
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
	return c.send("M?", serverID)
}

//   ____ _           _   __  __
//  / ___| |__   __ _| |_|  \/  | ___  ___ ___  __ _  __ _  ___
// | |   | '_ \ / _` | __| |\/| |/ _ \/ __/ __|/ _` |/ _` |/ _ \
// | |___| | | | (_| | |_| |  | |  __/\__ \__ \ (_| | (_| |  __/
//  \____|_| |_|\__,_|\__|_|  |_|\___||___/___/\__,_|\__, |\___|
//                                                   |___/

//
// Fields common to chat messages and die-roll results.
//
type ChatCommon struct {
	// The name of the person sending the message.
	Sender string

	// The names of the people the message was explicitly addressed to.
	// This will be nil for global messages.
	Recipients []string

	// The unique ID number for the chat message.
	MessageID int

	// True if this is a global message (sent to all users).
	ToAll bool

	// True if this message was sent only to the GM.
	ToGM bool
}

//
// ChatMessage message: A chat message was received from the server
// for the client to display.
//
type ChatMessageMessagePayload struct {
	BaseMessagePayload
	ChatCommon

	// The text of the chat message we received.
	Text string
}

const (
	// ToGMOnly as one of the recipients to a ChatMessage()
	// or Roll() method will cause the message or die-roll
	// result to be sent to the GM only.
	ToGMOnly = "%" // ChatMessage recipient is GM only

	// ToAll as one of the recipients to a ChatMessage()
	// or Roll() method will cause the message to go to
	// all connected clients.
	ToAll = "*" // ChatMessage recipients are all users
)

//
// ChatMessage sends a message on the chat channel to other
// users. The to paramter is a slice of user names of the people
// who should receive this message. Any of them may also be the
// special values
//    ToGMOnly  -- ignore any other names on the list. Send only to GM.
//    ToAll     -- send to all users.
//
func (c *Connection) ChatMessage(to []string, message string) error {
	return c.send("TO", "", to, message)
}

//
// ChatMessageToAll is equivalent to ChatMessage, but is addressed to all users.
//
func (c *Connection) ChatMessageToAll(message string) error {
	return c.send("TO", "", ToAll, message)
}

//
// ChatMessageToGM is equivalent to ChatMessage, but is addressed only to the GM.
//
func (c *Connection) ChatMessageToGM(message string) error {
	return c.send("TO", "", ToGMOnly, message)
}

//   ____ _
//  / ___| | ___  __ _ _ __
// | |   | |/ _ \/ _` | '__|
// | |___| |  __/ (_| | |
//  \____|_|\___|\__,_|_|
//

//
// Clear message: The client should remove the given object from
// its map. This ID may also be one of the following:
//   *                    Remove all objects
//   E*                   Remove all map elements
//   M*                   Remove all monster tokens
//   P*                   Remove all player tokens
//   [<imagename>=]<name> Remove token with given <name>
//
type ClearMessagePayload struct {
	BaseMessagePayload
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
	return c.send("CLR", objID)
}

//   ____ _                  ____ _           _
//  / ___| | ___  __ _ _ __ / ___| |__   __ _| |_
// | |   | |/ _ \/ _` | '__| |   | '_ \ / _` | __|
// | |___| |  __/ (_| | |  | |___| | | | (_| | |_
//  \____|_|\___|\__,_|_|   \____|_| |_|\__,_|\__|
//

//
// ClearChat message: clear the client's chat history
//
type ClearChatMessagePayload struct {
	BaseMessagePayload

	// User requesting the action, if known.
	RequestedBy string

	// Don't notify the user of the operation.
	DoSilently bool

	// If >0, clear all messages with IDs greater than target.
	// If <0, clear most recent -N messages.
	// If 0, clear all messages.
	Target int

	// Chat message ID of this notice.
	MessageID int
}

//
// ClearChat tells peers to remove all messages from their
// chat histories if target is zero. If target>0, then
// all messages with IDs greater than target are removed.
// Otherwise, if target<0 then only the most recent |target|
// messages are kept.
//
func (c *Connection) ClearChat(target int, silently bool) error {
	by := ""
	if silently {
		by = "*"
	} else if c.Authenticator != nil {
		by = c.Authenticator.Username
	}

	return c.send("CC", by, target)
}

//   ____ _                 _____
//  / ___| | ___  __ _ _ __|  ___| __ ___  _ __ ___
// | |   | |/ _ \/ _` | '__| |_ | '__/ _ \| '_ ` _ \
// | |___| |  __/ (_| | |  |  _|| | | (_) | | | | | |
//  \____|_|\___|\__,_|_|  |_|  |_|  \___/|_| |_| |_|
//

//
// ClearFrom message: remove all elements in the map file
// referenced.
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
	return c.send("CLR@", serverID)
}

//   ____                _           _   __  __           _
//  / ___|___  _ __ ___ | |__   __ _| |_|  \/  | ___   __| | ___
// | |   / _ \| '_ ` _ \| '_ \ / _` | __| |\/| |/ _ \ / _` |/ _ \
// | |__| (_) | | | | | | |_) | (_| | |_| |  | | (_) | (_| |  __/
//  \____\___/|_| |_| |_|_.__/ \__,_|\__|_|  |_|\___/ \__,_|\___|
//

//
// CombatMode message: indicate if combat mode should be in effect.
//
type CombatModeMessagePayload struct {
	BaseMessagePayload
	Enabled bool
}

//
// CombatMode tells all peers to enable or disable combat mode.
//
func (c *Connection) CombatMode(enabled bool) error {
	return c.send("CO", enabled)
}

//
// Toolbar message: indicate if the client's toolbar should be displayed.
//
type ToolbarMessagePayload struct {
	BaseMessagePayload
	Enabled bool
}

//   ____                                     _
//  / ___|___  _ __ ___  _ __ ___   ___ _ __ | |_
// | |   / _ \| '_ ` _ \| '_ ` _ \ / _ \ '_ \| __|
// | |__| (_) | | | | | | | | | | |  __/ | | | |_
//  \____\___/|_| |_| |_|_| |_| |_|\___|_| |_|\__|
//

//
// Comment message: a server comment to the client. The client is
// free to ignore these.
//
type CommentMessagePayload struct {
	BaseMessagePayload
	Text string
}

/* Clients don't send these */

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
	return c.send("DD/", re)
}

//  _                    _ _____
// | |    ___   __ _  __| |  ___| __ ___  _ __ ___
// | |   / _ \ / _` |/ _` | |_ | '__/ _ \| '_ ` _ \
// | |__| (_) | (_| | (_| |  _|| | | (_) | | | | | |
// |_____\___/ \__,_|\__,_|_|  |_|  \___/|_| |_| |_|
//

//
// LoadFrom message: load elements from the given file.
//
type LoadFromMessagePayload struct {
	BaseMessagePayload
	FileDefinition

	// If true, the client should only pre-load this data into a
	// local cache, but not start displaying these elements yet.
	CacheOnly bool

	// If true, the elements are merged with the existing map
	// contents rather than replacing them.
	Merge bool
}

//
// LoadFrom asks other clients to load a map files from a local
// disk file or from the server. The previous map contents are erased before
// each file is loaded.
//
// If local is true, a local path is specified. This is DEPRECATED.
//
// Otherwise, the path should be the ID for the file stored on the server.
//
// If merge is true, then the current map elements are not deleted first.
// In this case, the newly-loaded elements will be merged with what is already
// on the map.
//
func (c *Connection) LoadFrom(path string, local bool, merge bool) error {
	if merge {
		if local {
			return c.send("M", path)
		} else {
			return c.send("M@", path)
		}
	} else {
		if local {
			return c.send("L", path)
		} else {
			if err := c.send("CLR", "E*"); err != nil {
				return err
			}
			return c.send("M@", path)
		}
	}
}

//  _                    _  ___  _     _           _
// | |    ___   __ _  __| |/ _ \| |__ (_) ___  ___| |_
// | |   / _ \ / _` |/ _` | | | | '_ \| |/ _ \/ __| __|
// | |__| (_) | (_| | (_| | |_| | |_) | |  __/ (__| |_
// |_____\___/ \__,_|\__,_|\___/|_.__// |\___|\___|\__|
//                                  |__/

//
// LoadObject message: load an object from the server.
//
type LoadObjectMessagePayload struct {
	BaseMessagePayload
	MapObject
}

//
// Calculate the checksum for a data stream
// of already-broken-out fields as a [][]string
// (slice of lines, each of which is a slice of fields).
//
// The checksum is returned as a base-64-encoded string.
//
func streamChecksum(data [][]string) string {
	ck := sha256.New()
	for _, item := range data {
		for i, field := range item {
			if i > 0 {
				ck.Write([]byte{' '})
			}
			ck.Write([]byte(field))
		}
	}
	return base64.StdEncoding.EncodeToString(ck.Sum(nil))
}

//
// Calculate the checksum for a data stream
// of raw lines (not broken out into fields).
//
// The checksum is returned as a base-64-encoded string.
//
func streamChecksumStrings(data []string) string {
	ck := sha256.New()
	for _, item := range data {
		ck.Write([]byte(item))
	}
	return base64.StdEncoding.EncodeToString(ck.Sum(nil))
}

//
// LoadObject sends a MapObject to all peers.
//
func (c *Connection) LoadObject(me MapObject) error {
	data, err := SaveObjects([]MapObject{me}, nil, nil)
	if err != nil {
		return fmt.Errorf("Unable to send map object: %v", err)
	}

	if err := c.send("LS"); err != nil {
		return err
	}
	for _, a := range data {
		if err := c.send("LS:", a); err != nil {
			c.send("LS.", "0")
			return err
		}
	}
	c.send("LS.", fmt.Sprintf("%d", len(data)), streamChecksumStrings(data))
	return nil
}

//  __  __
// |  \/  | __ _ _ __ ___ ___
// | |\/| |/ _` | '__/ __/ _ \
// | |  | | (_| | | | (_| (_) |
// |_|  |_|\__,_|_|  \___\___/
//

//
// Marco message: the server is asking if we are still
// alive and responding. Reply by sending a Polo message.
//
type MarcoMessagePayload struct {
	BaseMessagePayload
}

/* clients don't send these */

//  __  __            _
// |  \/  | __ _ _ __| | __
// | |\/| |/ _` | '__| |/ /
// | |  | | (_| | |  |   <
// |_|  |_|\__,_|_|  |_|\_\
//

//
// Mark message: visually mark the given map coordinates.
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
	return c.send("MARK", x, y)
}

//  ____  _                ____
// |  _ \| | __ _  ___ ___/ ___|  ___  _ __ ___   ___  ___  _ __   ___
// | |_) | |/ _` |/ __/ _ \___ \ / _ \| '_ ` _ \ / _ \/ _ \| '_ \ / _ \
// |  __/| | (_| | (_|  __/___) | (_) | | | | | |  __/ (_) | | | |  __/
// |_|   |_|\__,_|\___\___|____/ \___/|_| |_| |_|\___|\___/|_| |_|\___|
//

//
// PlaceSomeone message: introduce a new creature token,
// or if that token is already on the board, update it
// with the new information (usually just moving its location).
//
// Retain any existing attributes in the original which have nil
// values here (notably, this server message never carries health
// stats so that structure will always be nil).
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
// way to do that would be to use UpdateObjAttributes() to change those
// specific attributes of the creature directly.
//
func (c *Connection) PlaceSomeone(someone interface{}) error {
	if player, ok := someone.(PlayerToken); ok {
		return c.send("PS", player.ObjID(), player.Color, player.Name,
			player.Area, player.Size, "player", player.Gx, player.Gy, player.Reach)
	}

	if monster, ok := someone.(MonsterToken); ok {
		return c.send("PS", monster.ObjID(), monster.Color, monster.Name,
			monster.Area, monster.Size, "monster", monster.Gx, monster.Gy, monster.Reach)
	}
	return fmt.Errorf("PlaceSomeone: argument not a PlayerToken or MonsterToken")
}

//   ___                        ___
//  / _ \ _   _  ___ _ __ _   _|_ _|_ __ ___   __ _  __ _  ___
// | | | | | | |/ _ \ '__| | | || || '_ ` _ \ / _` |/ _` |/ _ \
// | |_| | |_| |  __/ |  | |_| || || | | | | | (_| | (_| |  __/
//  \__\_\\__,_|\___|_|   \__, |___|_| |_| |_|\__,_|\__, |\___|
//                        |___/                     |___/

//
// QueryImage message: a peer wants to know where to find a given
// image and the server didn't know either. If you know the definition
// for that image, reply with an AddImage message of your own.
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
	return c.send("AI?", idef.Name, idef.Zoom)
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
// RemoveObjAttributes message: Adjust the multi-value attribute
// of the object with the given ID by removing the new values
// from it.
//
type RemoveObjAttributesMessagePayload struct {
	BaseMessagePayload
	ObjID    string
	AttrName string
	Values   []string
}

//
// RemoveObjAttributes informs peers to remove a set of string values from the existing
// value of an object attribute. The attribute must be one whose value is a list
// of strings, such as STATUSLIST.
//
func (c *Connection) RemoveObjAttributes(objID, attrName string, values []string) error {
	return c.send("OA-", objID, strings.ToUpper(attrName), values)
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
// listed to the ChatMessage() function.
//
// The rollspec may have any form that would be accepted to the
// dice.Roll function and dice.DieRoller.DoRoll method. See the dice package for details.
// https://pkg.go.dev/github.com/fizban-of-ragnarok/go-gma/v4/dice#DieRoller.DoRoll
//
func (c *Connection) RollDice(to []string, rollspec string) error {
	return c.send("D", to, rollspec)
}

//
// RollDiceToAll is equivalent to RollDice addressed to all users.
//
func (c *Connection) RollDiceToAll(message string) error {
	return c.send("D", ToAll, message)
}

//
// RollDiceToGM is equivalent to RollDice addressed only to the GM.
// This is a "blind" roll; only the GM will see the result.
//
func (c *Connection) RollDiceToGM(message string) error {
	return c.send("D", ToGMOnly, message)
}

//
// RollResult message: the server has rolled dice for someone
// and is sending out the results of that roll.
//
type RollResultMessagePayload struct {
	BaseMessagePayload
	ChatCommon

	// The title describing the purpose of the die-roll, as set by the user.
	Title string

	// The die roll result and details behind where it came from.
	Result dice.StructuredResult
}

//  ____  _          ____                     _
// |  _ \(_) ___ ___|  _ \ _ __ ___  ___  ___| |_ ___
// | | | | |/ __/ _ \ |_) | '__/ _ \/ __|/ _ \ __/ __|
// | |_| | | (_|  __/  __/| | |  __/\__ \  __/ |_\__ \
// |____/|_|\___\___|_|   |_|  \___||___/\___|\__|___/
//

type DicePreset struct {
	// The name of the preset
	Name string

	// Description of the preset
	Description string

	// The die-roll specification string
	RollSpec string
}

//
// DefineDicePresets replaces any existing die-roll presets you have
// stored on the server with the new set passed as the presets parameter.
//
func (c *Connection) DefineDicePresets(presets []DicePreset) error {
	var plist [][]string
	for _, p := range presets {
		plist = append(plist, []string{p.Name, p.Description, p.RollSpec})
	}
	return c.send("DD", plist)
}

//
// AddDicePresets is like DefineDicePresets except that it adds the presets
// passed in to the existing set rather than replacing them.
//
func (c *Connection) AddDicePresets(presets []DicePreset) error {
	var plist [][]string
	for _, p := range presets {
		plist = append(plist, []string{p.Name, p.Description, p.RollSpec})
	}
	return c.send("DD+", plist)
}

//
// QueryDicePresets requests that the server send you the die-roll
// presets currently stored for you. It will send you an UpdateDicePresets
// message.
//
func (c *Connection) QueryDicePresets() error {
	return c.send("DR")
}

//
// UpdateClock message: change the game clock
//
type UpdateClockMessagePayload struct {
	BaseMessagePayload

	// The clock is now at the given absolute number of
	// seconds from the GMA clock's global epoch.
	Absolute float64

	// The elapsed time counter is now this many seconds from
	// some epoch set by the GM.
	Relative float64
}

//
// UpdateDicePresets message: the client should accept the die-roll presets
// described here, replacing any previous presets it was
// using.
//
type UpdateDicePresetsMessagePayload struct {
	BaseMessagePayload
	Presets []DieRollPreset
}

type DieRollPreset struct {
	// The name by which this die-roll preset is identified to the user.
	// This must be unique among that user's presets.
	//
	// Clients typically
	// sort these names before displaying them.
	// Note that if a vertical bar ("|") appears in the name, all text
	// up to and including the bar are suppressed from display. This allows
	// for the displayed names to be forced into a particular order on-screen,
	// and allow a set of presets to appear to have the same name from the user's
	// point of view.
	Name string

	// A text description of the purpose for this die-roll specification.
	Description string

	// The die-roll specification to send to the server. This must be in a
	// form acceptable to the dice.Roll function. For details, see
	// https://pkg.go.dev/github.com/fizban-of-ragnarok/go-gma/v4/dice#DieRoller.DoRoll
	DieRollSpec string
}

//
// UpdateInitiative message: updates the initiative order listing
// for all combatants.
//
type UpdateInitiativeMessagePayload struct {
	BaseMessagePayload
	InitiativeList []InitiativeSlot
}

//
// An InitiativeSlot describes the creature occupying a given
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
	IsHolding bool

	// If true, the creature has a readied action.
	HasReadiedAction bool

	// It true, the creature is flat-footed.
	IsFlatFooted bool
}

//
// UpdateObjAttributes message: updates an existing object
// with new attributes. Any attributes not listed here should
// remain intact.
//
type UpdateObjAttributesMessagePayload struct {
	BaseMessagePayload
	ObjID    string
	NewAttrs map[string]interface{}
}

//
// UpdateObjAttributes informs peers that they should modify the
// specified object's attributes which are mentioned in the newAttrs
// map. This maps attribute names to their new values.
//
func (c *Connection) UpdateObjAttributes(objID string, newAttrs map[string]interface{}) error {
	var kvlist []interface{}
	for k, v := range newAttrs {
		kvlist = append(kvlist, strings.ToUpper(k))
		kvlist = append(kvlist, v)
	}
	kvField, err := tcllist.ToDeepTclString(kvlist...)
	if err != nil {
		return fmt.Errorf("Error in newAttrs list: %v", err)
	}
	return c.send("OA", objID, kvField)
}

//
// UpdatePeerList message: notifies the client that the list of
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
	Client string

	// How many seconds ago the peer last answered a "still alive?" ping from the server
	LastPolo float64

	// True if the client authenticated successfully
	IsAuthenticated bool

	// True if this structure describes the connection of this client program
	IsMe bool

	// True if the peer is running as the "main" or "primary" client
	IsMain bool

	// True if the peer client is not paying attention to any incoming
	// messages
	IsWriteOnly bool
}

//
// QueryPeers asks the server to send an UpdatePeerList
// message with the current set of peers who are connected
// to the server.
//
func (c *Connection) QueryPeers() error {
	return c.send("/CONN")
}

//
// UpdateProgress message: advises the client of the status of an operation
// in progress. The client may wish to display a progress indicator to the
// user.
//
type UpdateProgressMessagePayload struct {
	BaseMessagePayload

	// Unique identifier for the operation we're tracking
	OperationID string

	// Description of the operation in progress, suitable for display.
	Title string

	// The current progress toward MaxValue.
	Value int

	// The maximum expected value for the progress indication.
	// If this is 0, we don't yet know what the maximum will be.
	// Note that this may change from one message to another, if
	// the server realizes its previous estimate was incorrect.
	MaxValue int

	// If true, we can dispose of the tracked operation
	// and should not expect further updates about it.
	IsDone bool
}

//
// UpdateStatusMarker message: add or change a status marker to place
// on creature tokens.
//
// Note: the server usually sends these upon login, which the Connection
// struct stores internally. When this message is received, the Connection's
// status marker list is updated regardless of whether the client is subscribed
// to this message (which is may be if some overt action is required on its part
// to (re-)define the status marker).
//
type UpdateStatusMarkerMessagePayload struct {
	BaseMessagePayload
	StatusMarkerDefinition
}

//
// A StatusMarkerDefinition describes each creature token status
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
}

//
// Text produces a simple text description of a StatusMarkerDefinition structure.
//
func (c StatusMarkerDefinition) Text() string {
	return fmt.Sprintf("Condition %q: Shape=%q, Color=%q", c.Condition, c.Shape, c.Color)
}

type StatusMarkerDefinitions map[string]StatusMarkerDefinition

//
// Text produces a simple text description of a slice of StatusMarkerDefinitions
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
// UpdateTurn message: declares whose turn it is in combat.
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
// WriteOnly informs the server that from this point forward,
// we will not be interested in receiving any messages. We will
// only be sending.
//
// Regardless of sending this, clients should still read any
// data that is sent anyway. The Dial() method does in fact do this
// in case a misbehaving server doesn't fully respect the WriteOnly
// request.
//
func (c *Connection) WriteOnly() error {
	return c.send("NO")
}

//
// Sync requests that the server send the entire game state
// to it.
//
func (c *Connection) Sync() error {
	return c.send("SYNC")
}

//
// SyncChat requests that the server (re-)send past messages
// greater than the target message ID (target≥0) or the most
// recent |target| messages (target<0).
//
func (c *Connection) SyncChat(target int) error {
	return c.send("SYNC", "CHAT", target)
}

//
// Concurrency and general flow of operation for Dial():
// Dial() itself will block until the session with the server is completed.
// Thus, a client program will probably run it in a goroutine, using
// a channel subscribed to ERROR to receive any errors encountered by it
// (otherwise the errors are at least logged).
//
// The Dial() call does have some concurrent operations of its own, though,
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
// point the Dial() method returns.
//
// Dial is designed to be called in a goroutine so it can run in the
// background while the rest of the appliction continues with other
// tasks.
//
// Any errors encountered by the Dial() method will be reported on
// the channel being watched for ERROR events.
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
	var err error
	defer c.Close()

	c.signedOn = false
	for {
		err = c.tryConnect()
		if err == nil {
			c.signedOn = true
			if err = c.interact(); err != nil {
				c.Logger.Printf("mapper: INTERACT FAILURE: %v", err)
			}
			c.signedOn = false
		}

		if c.Context.Err() != nil || !c.StayConnected {
			break
		}
	}
}

func (c *Connection) tryConnect() error {
	var err error
	var conn net.Conn
	var i uint

	c.debug(2, "tryConnect() started")
	defer c.debug(2, "tryConnect() ended")

	for i = 0; c.Retries == 0 || i < c.Retries; i++ {
		if c.Timeout == 0 {
			var dialer net.Dialer
			conn, err = dialer.DialContext(c.Context, "tcp", c.Endpoint)
		} else {
			conn, err = net.DialTimeout("tcp", c.Endpoint, c.Timeout)
		}

		if err != nil {
			if c.Retries == 0 {
				c.Logger.Printf("mapper: Attempting connection (try %d): %v", i+1, err)
			} else {
				c.Logger.Printf("mapper: Attempting connection (try %d of %d): %v", i+1, c.Retries, err)
			}
		}
	}
	if err != nil {
		c.Logger.Printf("mapper: No more attempts allowed; giving up!")
		c.LastError = err
		return err
	}

	c.server = conn
	c.reader = bufio.NewScanner(conn)
	c.writer = bufio.NewWriter(conn)

	loginDone := make(chan error, 1)
	go c.login(loginDone)

syncloop:
	for {
		select {
		case err = <-loginDone:
			if err != nil {
				c.Logger.Printf("mapper: login process failed: %v", err)
				c.Close()
				c.LastError = err
				return err
			}
			break syncloop

		case <-c.Context.Done():
			c.Logger.Printf("mapper: context cancelled; closing connections and aborting login...")
			c.Close() // this will abort the scanner in login()
			return fmt.Errorf("mapper: connection aborted by termination of context")
		}
	}

	return nil
}

func (c *Connection) login(done chan error) {
	defer close(done)

	c.debug(2, "login() started")
	defer c.debug(2, "login() ended")

	c.Logger.Printf("mapper: Initial server negotiation...")
	syncDone := false
	authPending := false
	recCount := 0
	c.Preamble = nil

	for !syncDone && c.reader.Scan() {
		if c.DebuggingLevel >= 3 {
			c.debug(3, util.Hexdump(c.reader.Bytes()))
		}
		f, err := tcllist.ParseTclList(c.reader.Text())
		if err != nil {
			c.Logger.Printf("mapper: unable to parse message from server: %v", err)
			done <- err
			return
		}
		c.debug(1, "server->"+f[0])

		if len(f) == 0 {
			// empty line is ok, we just ignore it
			continue
		}

		if f[0] == "OK" {
			// Unlike the Python version, we won't even accept a server
			// too old to have a protocol version or authentication
			// support at all. That's clearly below our minimum supported
			// protocol level by definition.
			fv, err := tcllist.ConvertTypes(f, "si*")
			if err != nil {
				c.Logger.Printf("mapper: error in server greeting (%v): %v", f, err)
				done <- fmt.Errorf("server greeting %v could not be parsed: %v", f, err)
				return
			}
			c.Protocol = fv[1].(int)

			if c.Protocol < MINIMUM_SUPPORTED_MAP_PROTOCOL {
				c.Logger.Printf("mapper: Unable to connect to mapper with protocol older than %d (server offers %d)", MINIMUM_SUPPORTED_MAP_PROTOCOL, c.Protocol)
				done <- fmt.Errorf("server version %d too old (must be at least %d)", c.Protocol, MINIMUM_SUPPORTED_MAP_PROTOCOL)
				return
			}
			if c.Protocol > MAXIMUM_SUPPORTED_MAP_PROTOCOL {
				c.Logger.Printf("mapper: Unable to connect to mapper with protocol newer than %d (server offers %d)", MAXIMUM_SUPPORTED_MAP_PROTOCOL, c.Protocol)
				c.Logger.Printf("mapper: ** UPGRADE GMA **")
				done <- fmt.Errorf("server version %d too new (must be at most %d)", c.Protocol, MAXIMUM_SUPPORTED_MAP_PROTOCOL)
				return
			}
			if c.Protocol >= 321 && len(f) >= 3 {
				// Authenticate user (protocol 321 was the first that supported
				// authentication)
				if c.Authenticator == nil {
					c.Logger.Printf("mapper: Server requires authentication but no authenticator was provided for the client.")
					done <- AuthenticationRequired
					return
				}
				c.Logger.Printf("mapper: authenticating to server")
				c.Authenticator.Reset()
				response, err := c.Authenticator.AcceptChallenge(f[2])
				if err != nil {
					c.Logger.Printf("mapper: Error accepting server's challenge: %v", err)
					done <- err
					return
				}
				c.rawSend("AUTH", response, c.Authenticator.Username, c.Authenticator.Client)
				c.Logger.Printf("mapper: authentication sent. Awaiting validation.")
				authPending = true
			} else {
				c.Logger.Printf("mapper: using protocol %d.", c.Protocol)
				c.Logger.Printf("mapper: server sync complete. No authentication requested by server.")
			}
			syncDone = true
		} else {
			// Digest all the preamble content before the authentication
			// challenge or end-of-greeting
			recCount++
			switch f[0] {
			case "AC": // AC name id color area size
				_, err := tcllist.ConvertTypes(f, "ssssss")
				if err != nil {
					c.Logger.Printf("mapper: INVALID server AC data: %v: %v", f, err)
					continue
				}
				// add to character list
				c.Characters[f[1]] = CharacterDefinition{
					Name:  f[1],
					ObjID: f[2],
					Color: f[3],
					Area:  f[4],
					Size:  f[5],
				}
				c.Logger.Printf("mapper: sync %02d: Added %s", recCount, f[1])

			case "DSM": // DSM name shape color
				if err := c.receiveDSM(f); err != nil {
					c.Logger.Printf("mapper: error in DSM data: %v", err)
				}
				c.Logger.Printf("mapper: sync %02d: Added condition %s", recCount, f[1])

			case "//":
				if len(f) >= 5 && f[1] == "CORE" && f[2] == "UPDATE" && f[3] == "//" {
					// CORE UPDATE // version
					advertisedVersion := f[4]
					d, err := util.VersionCompare(GMAVersionNumber, advertisedVersion)
					if err != nil {
						c.Logger.Printf("mapper: Can't compare version %s vs %s: %v", GMAVersionNumber, advertisedVersion, err)
						continue
					}

					if d > 0 {
						c.Logger.Printf("mapper: **NOTE** You are running a client with GMA Core API library version %s, which is ahead of the latest advertised version (%s) on your server.", GMAVersionNumber, advertisedVersion)
						c.Logger.Printf("mapper: This may mean you are working on an experimental version, or that your GM isn't using the latest version.")
						c.Logger.Printf("mapper: If you did not intend for this to be the case, you might want to check with your GM to be sure your client is compatible.")
					} else if d < 0 {
						c.Logger.Printf("mapper: **NOTE** An update for the GMA Core API is available. You are using %s, but your server is advertising version %s.", GMAVersionNumber, advertisedVersion)
					}
					c.Logger.Printf("mapper: sync %02d: Noted Core API version %s", recCount, advertisedVersion)
				} else if len(f) >= 4 && f[2] == "UPDATE" && f[3] == "//" {
					c.Logger.Printf("mapper: sync %02d: Noted other client version", recCount)
				} else {
					c.Logger.Printf("mapper: sync %02d...", recCount)
				}
				c.Preamble = append(c.Preamble, c.reader.Text())

			default:
				c.Preamble = append(c.Preamble, c.reader.Text())
				c.Logger.Printf("mapper: sync %02d...", recCount)
			}
		}
	}
	if err := c.reader.Err(); err != nil {
		done <- err
		return
	}

	// If we're still waiting for authentication results, do that...
	c.debug(2, "Switched to authentication result scanner")
	for authPending && c.reader.Scan() {
		if c.DebuggingLevel >= 3 {
			c.debug(3, util.Hexdump(c.reader.Bytes()))
		}
		f, err := tcllist.ParseTclList(c.reader.Text())
		if err != nil {
			c.Logger.Printf("mapper: unable to parse message from server: %v", err)
			done <- err
			return
		}
		c.debug(1, "server->"+f[0])

		if len(f) == 0 {
			// empty line is ok, we just ignore it
			continue
		}

		switch f[0] {
		case "DENIED":
			if len(f) > 1 {
				c.Logger.Printf("mapper: access denied by server: %v", f[1])
			} else {
				c.Logger.Printf("mapper: access denied by server")
			}
			done <- AuthenticationFailed
			return

		case "GRANTED":
			if len(f) > 1 {
				c.Logger.Printf("mapper: access granted for %s", f[1])
				if c.Authenticator != nil {
					c.Authenticator.Username = f[1]
				}
			} else {
				c.Logger.Printf("mapper: access granted")
			}
			authPending = false

		case "//", "CONN", "CONN:", "CONN.":
			// Ignore

		default:
			c.Logger.Printf("mapper: unexpected server message %v while waiting for authentication to complete", f)
		}
	}
	if err := c.reader.Err(); err != nil {
		done <- err
		return
	}

	if err := c.filterSubscriptions(); err != nil {
		done <- err
		return
	}

	if c.DebuggingLevel >= 2 {
		c.debug(2, "Completed server sign-on process")
		if c.Authenticator != nil {
			c.debug(2, fmt.Sprintf("Logged in as %s", c.Authenticator.Username))
		}
		c.debug(2, fmt.Sprintf("Server is using protocol version %d", c.Protocol))
		c.debug(2, fmt.Sprintf("Defined Characters:\n%s", CharacterDefinitions(c.Characters).Text()))
		c.debug(2, fmt.Sprintf("Defined Status Markers:\n%s", StatusMarkerDefinitions(c.Conditions).Text()))
		c.debug(2, "Preamble:\n"+strings.Join(c.Preamble, "\n"))
		c.debug(2, fmt.Sprintf("Last error: %v", c.LastError))
	}
}

func (c *Connection) receiveDSM(f []string) error {
	_, err := tcllist.ConvertTypes(f, "ssss")
	if err != nil {
		return fmt.Errorf("%v: %v", f, err)
	}
	// add to status list
	c.Conditions[f[1]] = StatusMarkerDefinition{
		Condition: f[1],
		Shape:     f[2],
		Color:     f[3],
	}
	return nil
}

//
// The official protocol spec is the mapper(6) manpage. This is a summary only.
//
// ->Comment              // ...
//                        // MAPPER UPDATE // <version> <file>
//                        // CORE UPDATE // <version> [<file>]
// ->UpdateProgress       // BEGIN <id> <max>|* <title>
// ->UpdateProgress       // UPDATE <id> <value> [<newmax>]
// ->UpdateProgress       // END <id>
// ->AddCharacter         AC <name> <id> <color> <area> <size>
// <-(Subscribe)          ACCEPT <msglist>
// <>AddImage             AI <name> <size>
//                        AI: <data>			(repeated)
//                        AI. <#lines> <sha256>
//                        AI@ <name> <size> <serverid>
// <>QueryImage           AI? <name> <size>
// <-(login)              AUTH <response> [<user>|GM <client>]
// <>AdjustView           AV <xview> <yview>
// <>Clear                CLR *|E*|M*|P*|[<imagename>=]<name>|<objID>
// <>ClearChat            CC *|<user> [""|<newmax>|-<#recents> [<messageID>]]
// <>ClearFrom            CLR@ <serverid>
// <>CombatMode           CO 0|1
// ->UpdateClock          CS <abs> <rel>
// <-RollDice             D {<recipient>|@|*|% ...} <spec>
// <-DefineDicePresets    DD {{<name> <description> <spec>} ...}     (replace)
// <-DefineDicePresets    DD+ {{<name> <description> <spec>} ...}    (append)
// <-FilterDicePresets    DD/ <regex>
// ->UpdateDicePresets    DD=
//                        DD: <i> <name> <description> <spec>      (repeated)
//                        DD. <#lines> <sha256>
// ->(login)              DENIED [<message>]
// <-QueryDicePresets     DR
// ->UpdateStatusMarker   DSM <condition> <shape> <color>
// ->(login)              GRANTED <name>|GM
// ->UpdateTurn           I {<r> <c> <s> <m> <h>} <id>|""|*Monsters*|/<regex>
// ->UpdateInitiative     IL {{<name> <hold> <ready> <hp> <flat> <slot#>} ...}
// <>LoadFrom             L {<path> ...}              (clear map before each)
//                        M {<path> ...}              (merge to map)
//                        M@ <serverid>               (merge to map)
// <>LoadObject           LS
//                        LS: <data>                  (repeated)
//                        LS. <#lines> <sha256>
//                        LS. 0                       (NEW: cancel LS)
// <>LoadFrom             M? <serverid>
// ->Marco                MARCO
// <>Mark                 MARK <x> <y>
// <-WriteOnly            NO
// <>UpdateObjAttributes  OA <objid> {<key> <value ...}
// <>AddObjAttributes     OA+ <objid> <key> {<value> ...}
// <>RemoveObjAttributes  OA- <objid> <key> {<value> ...}
// ->(login)              OK <version> [<challenge>]
// ->                     PRIV <message>
// <-Polo                 POLO
// <>PlaceSomeone         PS <id> <color> <name> <area> <size> player|monster <gx> <gy> <reach>
// ->RollResult           ROLL <from> {<recipient> ...} <title> <result> {{<type> <value>} ...} <messageid>
// <-Sync                 SYNC
// <-SyncChat             SYNC CHAT [-<#recent>|<since>]
// ->Toolbar              TB 0|1
// <>ChatMessage          TO <from> {<recipient>|@|*|% ...} <message> [<messageid>]
// <-QueryPeers           /CONN
// ->UpdatePeerList       CONN
//                        CONN: <i> you|peer <addr> <user> <client> <auth> <primary> <writeonly> <lastseen>
//                        CONN. <#lines> <sha256>

func (c *Connection) finishStream(f []string, started bool, dataLen int, checksum string) (bool, error) {
	fv, err := tcllist.ConvertTypes(f, "si*")
	if err != nil {
		return false, fmt.Errorf("cannot parse stream-end message %q: %v", f, err)
	}
	if !started {
		return false, fmt.Errorf("stream-end encountered before stream-begin: %q", f)
	}
	if fv[1].(int) == 0 {
		return false, nil
	}
	if len(f) > 2 && checksum != f[2] {
		return false, fmt.Errorf("stream data transfer checksum error")
	}
	if dataLen != fv[1].(int) {
		return false, fmt.Errorf("stream transfer had %d elements but expected %d", dataLen, fv[1].(int))
	}
	return true, nil
}

//
// listen for, and dispatch, incoming server messages
//
func (c *Connection) listen(done chan error) {
	defer func() {
		close(done)
		c.Logger.Printf("mapper: stopped listening to server")
		c.debug(2, "listen() ended")
	}()
	c.debug(2, "listen() started")

	var (
		imageDataBuffer []string
		imageDataDef    ImageDefinition
		currentImage    string
		presetBuffer    [][]string
		currentPreset   string
		peerBuffer      [][]string
		currentPeer     string
		objBuffer       []string
		currentObj      string
		rPeerBuffer     []string
		rPresetBuffer   []string
	)

	strike := 0
	c.Logger.Printf("mapper: listening for server messages to dispatch...")
	for c.reader.Scan() {
		if c.DebuggingLevel >= 3 {
			c.debug(3, util.Hexdump(c.reader.Bytes()))
		}
		f, err := tcllist.ParseTclList(c.reader.Text())
		if err != nil {
			c.reportError(fmt.Errorf("mapper: unable to parse message \"%s\" from server: %v", c.reader.Text(), err))
			if strike > 3 {
				c.reportError(fmt.Errorf("mapper: giving up"))
				done <- err
				return
			}
			strike++
			continue
		} else {
			strike = 0
		}
		c.debug(1, "server->"+f[0])

		if len(f) == 0 {
			continue // skip blank lines
		}

		payload := BaseMessagePayload{
			rawMessage: f,
		}
		if c.DebuggingLevel >= 3 {
			c.debug(3, fmt.Sprintf("Payload: %q", f))
		}

	payloadSelection:
		switch f[0] {
		case "//":
			//    ____   ___ ___  __  __ __  __ ___ _  _ _____ ___
			//   / / /  / __/ _ \|  \/  |  \/  | __| \| |_   _/ __|
			//  / / /  | (_| (_) | |\/| | |\/| | _|| .` | | | \__ \
			// /_/_/    \___\___/|_|  |_|_|  |_|___|_|\_| |_| |___/
			//
			// // BEGIN  <id> <max>|* <title>
			// // UPDATE <id> <value> [<newmax>]
			// // END    <id>
			//
			if len(f) > 2 && (f[1] == "BEGIN" || f[1] == "UPDATE" || f[1] == "END") {
				payload.messageType = UpdateProgress
				gauge, ok := c.Gauges[f[2]]
				if !ok {
					gauge = &UpdateProgressMessagePayload{
						OperationID: f[2],
						Title:       "(Unnamed progress gauge)",
					}
					c.Gauges[f[2]] = gauge
				}
				if len(f) == 5 && f[1] == "BEGIN" {
					//
					// BEGIN: define new gauge
					//
					gauge.Title = f[4]
					gauge.Value = 0
					gauge.IsDone = false
					if f[3] == "*" {
						gauge.MaxValue = 0
					} else {
						v, err := strconv.Atoi(f[3])
						if err != nil {
							c.reportError(fmt.Errorf("mapper: progress ID %s: invalid max value \"%s\": %v", f[2], f[3], err))
						} else {
							gauge.MaxValue = v
						}
					}
				} else if len(f) >= 4 && f[1] == "UPDATE" {
					//
					// UPDATE: advance the gauge
					//
					v, err := strconv.Atoi(f[3])
					if err != nil {
						c.reportError(fmt.Errorf("mapper: progress ID %s: invalid updated value \"%s\": %v", f[2], f[3], err))
					} else {
						gauge.Value = v
					}
					if len(f) > 4 {
						if f[4] == "*" {
							gauge.MaxValue = 0
						} else {
							v, err := strconv.Atoi(f[4])
							if err != nil {
								c.reportError(fmt.Errorf("mapper: progress ID %s: invalid max value \"%s\": %v", f[2], f[4], err))
							} else {
								gauge.MaxValue = v
							}
						}
					}
				} else if f[1] == "END" {
					//
					// END: stop tracking this gauge
					//
					gauge.IsDone = true
					delete(c.Gauges, f[2])
				} else {
					// it's a comment
					payload.messageType = Comment
					ch, ok := c.Subscriptions[Comment]
					if ok {
						ch <- CommentMessagePayload{
							BaseMessagePayload: payload,
							Text:               strings.Join(f[1:], " "),
						}
					}
					continue
				}

				ch, ok := c.Subscriptions[UpdateProgress]
				if ok {
					ch <- *gauge
				}
			} else {
				//
				// regular comment
				//
				payload.messageType = Comment
				ch, ok := c.Subscriptions[Comment]
				if ok {
					ch <- CommentMessagePayload{
						BaseMessagePayload: payload,
						Text:               strings.Join(f[1:], " "),
					}
				}
			}

		case "AC":
			//    _   ___
			//   /_\ / __|
			//  / _ \ (__
			// /_/ \_\___|
			//
			// AC <name> <id> <color> <area> <size>
			//
			if len(f) != 6 {
				c.reportError(fmt.Errorf("mapper: AddCharacter message field count %d wrong", len(f)))
				continue
			}
			// add to character list
			c.Characters[f[1]] = CharacterDefinition{
				Name:  f[1],
				ObjID: f[2],
				Color: f[3],
				Area:  f[4],
				Size:  f[5],
			}
			ch, ok := c.Subscriptions[AddCharacter]
			if ok {
				payload.messageType = AddCharacter
				ch <- AddCharacterMessagePayload{
					BaseMessagePayload: payload,
					CharacterDefinition: CharacterDefinition{
						Name:  f[1],
						ObjID: f[2],
						Color: f[3],
						Area:  f[4],
						Size:  f[5],
					},
				}
			}

		case "AI":
			//    _   ___
			//   /_\ |_ _|
			//  / _ \ | |
			// /_/ \_\___|
			//
			// AI <name> <size>
			// AI: <data>
			// AI. <#lines> <sha256>
			//
			_, ok := c.Subscriptions[AddImage]
			if ok {
				if currentImage != "" {
					c.reportError(fmt.Errorf("AI encountered before previous one ended"))
				}
				fv, err := tcllist.ConvertTypes(f, "ssf")
				if err != nil {
					c.reportError(fmt.Errorf("cannot parse AI message from server: %v", err))
					break
				}
				currentImage = c.reader.Text()
				imageDataBuffer = nil
				imageDataDef.Name = fv[1].(string)
				imageDataDef.Zoom = fv[2].(float64)
				if imageDataDef.Zoom == 0 || imageDataDef.Name == "" {
					c.reportError(fmt.Errorf("cannot parse AI message from server: data out of range"))
					break
				}
			}

		case "AI:":
			_, ok := c.Subscriptions[AddImage]
			if ok {
				if currentImage == "" {
					c.reportError(fmt.Errorf("AI: message received before AI"))
					break
				}
				if len(f) != 2 {
					c.reportError(fmt.Errorf("AI: message field count wrong (%d)", len(f)))
					break
				}
				imageDataBuffer = append(imageDataBuffer, f[1])
			}

		case "AI.":
			ch, ok := c.Subscriptions[AddImage]
			if ok {
				imgOk, err := c.finishStream(f, currentImage != "", len(imageDataBuffer),
					streamChecksumStrings(imageDataBuffer))
				if err != nil {
					c.reportError(fmt.Errorf("bad AddImage transfer: %v", err))
				}
				if imgOk {
					data, err := base64.StdEncoding.DecodeString(strings.Join(imageDataBuffer, ""))
					if err != nil {
						c.reportError(fmt.Errorf("Image data could not be decoded: %v", err))
					} else {
						r := make([]string, 0, 2+len(imageDataBuffer))
						r = append(r, currentImage)
						r = append(r, imageDataBuffer...)
						r = append(r, payload.rawMessage...)
						payload.rawMessage = r
						payload.messageType = AddImage
						ch <- AddImageMessagePayload{
							BaseMessagePayload: payload,
							ImageDefinition: ImageDefinition{
								Zoom:        imageDataDef.Zoom,
								Name:        imageDataDef.Name,
								IsLocalFile: false,
							},
							ImageData: data,
						}
					}
				}
				imageDataBuffer = nil
				imageDataDef = ImageDefinition{}
				currentImage = ""
			}

		case "AI@":
			//    _   ___  ____
			//   /_\ |_ _|/ __ \
			//  / _ \ | |/ / _` |
			// /_/ \_\___\ \__,_|
			//            \____/
			//
			// AI@ <name> <size> <serverID>
			//
			ch, ok := c.Subscriptions[AddImage]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ssfs")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid AddImage (AI@) message: %v", err))
					break
				}
				payload.messageType = AddImage
				ch <- AddImageMessagePayload{
					BaseMessagePayload: payload,
					ImageDefinition: ImageDefinition{
						Zoom:        fv[2].(float64),
						Name:        fv[1].(string),
						File:        fv[3].(string),
						IsLocalFile: false,
					},
				}
			}

		case "AI?":
			//    _   ___ ___
			//   /_\ |_ _|__ \
			//  / _ \ | |  /_/
			// /_/ \_\___|(_)
			//
			// AI? <name> <size>
			//
			ch, ok := c.Subscriptions[QueryImage]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ssf")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid QueryImage message: %v", err))
					break
				}
				payload.messageType = QueryImage
				ch <- QueryImageMessagePayload{
					BaseMessagePayload: payload,
					ImageDefinition: ImageDefinition{
						Zoom: fv[2].(float64),
						Name: fv[1].(string),
					},
				}
			}

		case "AV":
			//    ___   __
			//   /_\ \ / /
			//  / _ \ V /
			// /_/ \_\_/
			//
			// AV <x> <y>
			//
			ch, ok := c.Subscriptions[AdjustView]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "sff")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid AdjustView message: %v", err))
					break
				}
				payload.messageType = AdjustView
				ch <- AdjustViewMessagePayload{
					BaseMessagePayload: payload,
					XView:              fv[1].(float64),
					YView:              fv[2].(float64),
				}
			}

		case "CLR":
			//   ___ _    ___
			//  / __| |  | _ \
			// | (__| |__|   /
			//  \___|____|_|_\
			//
			// CLR *|E*|M*|P*|[<image>=]<name>|<id>
			//
			ch, ok := c.Subscriptions[Clear]
			if ok {
				if len(f) != 2 {
					c.reportError(fmt.Errorf("Invalid Clear message: parameter list length %d", len(f)))
				} else {
					payload.messageType = Clear
					ch <- ClearMessagePayload{
						BaseMessagePayload: payload,
						ObjID:              f[1],
					}
				}
			}

		case "CC":
			//   ___ ___
			//  / __/ __|
			// | (_| (__
			//  \___\___|
			//
			// CC *|<user> ""|<newmax>|-<#recents> <messageID>
			//
			ch, ok := c.Subscriptions[ClearChat]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ssIi")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid ClearChat message: %v", err))
				} else {
					payload.messageType = ClearChat
					by := fv[1].(string)
					silent := false
					if by == "*" {
						by = ""
						silent = true
					}
					ch <- ClearChatMessagePayload{
						BaseMessagePayload: payload,
						RequestedBy:        by,
						DoSilently:         silent,
						Target:             fv[2].(int),
						MessageID:          fv[3].(int),
					}
				}
			}

		case "CLR@":
			//   ___ _    ___  ____
			//  / __| |  | _ \/ __ \
			// | (__| |__|   / / _` |
			//  \___|____|_|_\ \__,_|
			//                \____/
			// CLR@ <id>
			//
			ch, ok := c.Subscriptions[ClearFrom]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ss")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid ClearFrom message: %v", err))
				} else {
					payload.messageType = ClearFrom
					ch <- ClearFromMessagePayload{
						BaseMessagePayload: payload,
						FileDefinition: FileDefinition{
							File:        fv[1].(string),
							IsLocalFile: false,
						},
					}
				}
			}

		case "CO":
			//   ___ ___
			//  / __/ _ \
			// | (_| (_) |
			//  \___\___/
			//
			// CO 0|1
			//
			ch, ok := c.Subscriptions[CombatMode]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "s?")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid CombatMode message: %v", err))
				} else {
					payload.messageType = CombatMode
					ch <- CombatModeMessagePayload{
						BaseMessagePayload: payload,
						Enabled:            fv[1].(bool),
					}
				}
			}

		case "CONN":
			//   ___ ___  _  _ _  _
			//  / __/ _ \| \| | \| |
			// | (_| (_) | .` | .` |
			//  \___\___/|_|\_|_|\_|
			//
			// CONN
			// CONN: <i> you|peer <addr> <user> <client> <auth> <primary> <writeonly> <lastseen>
			// CONN. <#peers> <sha256>
			//
			_, ok := c.Subscriptions[UpdatePeerList]
			if ok {
				if currentPeer != "" {
					c.reportError(fmt.Errorf("CONN encountered before previous one ended"))
				}
				presetBuffer = nil
				rPresetBuffer = nil
				currentPeer = c.reader.Text()
			}

		case "CONN:":
			_, ok := c.Subscriptions[UpdatePeerList]
			if ok {
				if currentPeer == "" {
					c.reportError(fmt.Errorf("CONN: enountered before CONN"))
					break
				}
				if len(f) != 10 {
					c.reportError(fmt.Errorf("CONN: message field count wrong (%d)", len(f)))
					break
				}
				thisSet := make([]string, 10)
				copy(thisSet, f)
				peerBuffer = append(peerBuffer, thisSet)
				rPeerBuffer = append(rPeerBuffer, c.reader.Text())
			}

		case "CONN.":
			ch, ok := c.Subscriptions[UpdatePeerList]
			if ok {
				ccOk, err := c.finishStream(f, currentPeer != "", len(peerBuffer), streamChecksum(peerBuffer))
				if err != nil {
					c.reportError(fmt.Errorf("bad peer list transfer: %v", err))
				}
				plist := make([]Peer, 0, len(peerBuffer))
				if ccOk {
					for i, preset := range peerBuffer {
						pv, err := tcllist.ConvertTypes(preset, "sissss???f")
						if err != nil {
							c.reportError(fmt.Errorf("bad peer list transfer: %v", err))
							ccOk = false
							break
						}
						if i != pv[1].(int) {
							c.reportError(fmt.Errorf("peer list transfer sequence error %d!=%d: %v", i, pv[1].(int), err))
							ccOk = false
							break
						}
						plist = append(plist, Peer{
							Addr:            pv[3].(string),
							User:            pv[4].(string),
							Client:          pv[5].(string),
							LastPolo:        pv[9].(float64),
							IsAuthenticated: pv[6].(bool),
							IsMe:            pv[2].(string) == "you",
							IsMain:          pv[7].(bool),
							IsWriteOnly:     pv[8].(bool),
						})
					}
				}
				if ccOk {
					rm := make([]string, 0, 2+len(rPeerBuffer))
					rm = append(rm, currentPeer)
					rm = append(rm, rPeerBuffer...)
					rm = append(rm, payload.rawMessage...)
					payload.rawMessage = rm
					payload.messageType = UpdatePeerList
					ch <- UpdatePeerListMessagePayload{
						BaseMessagePayload: payload,
						PeerList:           plist,
					}
				}
				currentPeer = ""
				peerBuffer = nil
				rPeerBuffer = nil
			}

		case "CS":
			//   ___ ___
			//  / __/ __|
			// | (__\__ \
			//  \___|___/
			//
			// CS <abs> <rel>
			//
			ch, ok := c.Subscriptions[UpdateClock]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "sff")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid UpdateClock message: %v", err))
				} else {
					payload.messageType = UpdateClock
					ch <- UpdateClockMessagePayload{
						BaseMessagePayload: payload,
						Absolute:           fv[1].(float64),
						Relative:           fv[2].(float64),
					}
				}
			}

		case "DD=":
			//  ___  ___
			// |   \|   \ ___
			// | |) | |) |___|
			// |___/|___/|___|
			//
			// DD=
			// DD: <i> <name> <desc> <spec>
			// DD. <#defs> <sha256>
			//
			_, ok := c.Subscriptions[UpdateDicePresets]
			if ok {
				if currentPreset != "" {
					c.reportError(fmt.Errorf("UpdateDicePresets message started before previous one ended"))
				}
				presetBuffer = nil
				rPresetBuffer = nil
				currentPreset = c.reader.Text()
			}

		case "DD:":
			_, ok := c.Subscriptions[UpdateDicePresets]
			if ok {
				if currentPreset == "" {
					c.reportError(fmt.Errorf("UpdateDicePresets message without start-of-stream"))
					break
				}

				if len(f) != 5 {
					c.reportError(fmt.Errorf("UpdateDicePresets message field count wrong (%d)", len(f)))
					break
				}
				thisSet := make([]string, 5)
				copy(thisSet, f)
				presetBuffer = append(presetBuffer, thisSet)
				rPresetBuffer = append(rPresetBuffer, c.reader.Text())
			}

		case "DD.":
			ch, ok := c.Subscriptions[UpdateDicePresets]
			if ok {
				ddOk, err := c.finishStream(f, currentPreset != "", len(presetBuffer), streamChecksum(presetBuffer))
				if err != nil {
					c.reportError(fmt.Errorf("bad preset transfer: %v", err))
				}
				plist := make([]DieRollPreset, 0, len(presetBuffer))
				if ddOk {
					for i, preset := range presetBuffer {
						pv, err := tcllist.ConvertTypes(preset, "sisss")
						if err != nil {
							c.reportError(fmt.Errorf("bad preset transfer: %v", err))
							ddOk = false
							break
						}
						if i != pv[1].(int) {
							c.reportError(fmt.Errorf("preset transfer sequence error %d!=%d: %v", i, pv[1].(int), err))
							ddOk = false
							break
						}
						plist = append(plist, DieRollPreset{
							Name:        pv[2].(string),
							Description: pv[3].(string),
							DieRollSpec: pv[4].(string),
						})
					}
				}
				if ddOk {
					r := make([]string, 0, len(rPresetBuffer)+2)
					r = append(r, currentPreset)
					r = append(r, rPresetBuffer...)
					r = append(r, payload.rawMessage...)
					payload.rawMessage = r
					payload.messageType = UpdateDicePresets
					ch <- UpdateDicePresetsMessagePayload{
						BaseMessagePayload: payload,
						Presets:            plist,
					}
				}
				presetBuffer = nil
				rPresetBuffer = nil
				currentPreset = ""
			}

		case "DENIED":
			//  ___  ___ _  _ ___ ___ ___
			// |   \| __| \| |_ _| __|   \
			// | |) | _|| .` || || _|| |) |
			// |___/|___|_|\_|___|___|___/
			//
			if len(f) > 1 {
				c.reportError(fmt.Errorf("server denied access: %s", f[1]))
			} else {
				c.reportError(fmt.Errorf("server denied access"))
			}

		case "DSM":
			//  ___  ___ __  __
			// |   \/ __|  \/  |
			// | |) \__ \ |\/| |
			// |___/|___/_|  |_|
			//
			// DSM <cond> <shape> <color>
			//
			if err := c.receiveDSM(f); err != nil {
				c.reportError(fmt.Errorf("error in UpdateStatusMarker: %v", err))
				break
			}
			ch, ok := c.Subscriptions[UpdateStatusMarker]
			if ok {
				if len(f) != 4 {
					// This should have been caught by receiveDSM.
					c.reportError(fmt.Errorf("error in UpdateStatusMarker: field count %d", len(f)))
					break
				}
				payload.messageType = UpdateStatusMarker
				ch <- UpdateStatusMarkerMessagePayload{
					BaseMessagePayload: payload,
					StatusMarkerDefinition: StatusMarkerDefinition{
						Condition: f[1],
						Shape:     f[2],
						Color:     f[3],
					},
				}
			}

		case "I":
			//  ___
			// |_ _|
			//  | |
			// |___|
			//
			// I {<r> <c> <s> <m> <h>} <id>|""|*Monsters*|/<regex>
			//
			ch, ok := c.Subscriptions[UpdateTurn]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "sss")
				if err != nil {
					c.reportError(fmt.Errorf("invalid UpdateTurn message: %v", err))
					break
				}
				breakdown, err := tcllist.Parse(fv[1].(string), "iiiii")
				if err != nil {
					c.reportError(fmt.Errorf("invalid UpdateTurn time structure: %v", err))
					break
				}

				payload.messageType = UpdateTurn
				ch <- UpdateTurnMessagePayload{
					BaseMessagePayload: payload,
					ActorID:            fv[2].(string),
					Hours:              breakdown[4].(int),
					Minutes:            breakdown[3].(int),
					Seconds:            breakdown[2].(int),
					Rounds:             breakdown[0].(int),
					Count:              breakdown[1].(int),
				}
			}

		case "IL":
			//  ___ _
			// |_ _| |
			//  | || |__
			// |___|____|
			//
			// IL {{<name> <hold?> <ready?> <hp> <flat?> <slot#>} ...}
			//

			ch, ok := c.Subscriptions[UpdateInitiative]
			if ok {
				if len(f) != 2 {
					c.reportError(fmt.Errorf("invalid UpdateInitiative message: field count %d", len(f)))
					break
				}
				entries, err := tcllist.ParseTclList(f[1])
				if err != nil {
					c.reportError(fmt.Errorf("invalid UpdateInitiative structure: %v", err))
					break
				}

				var ilist []InitiativeSlot
				for i, entry := range entries {
					fv, err := tcllist.Parse(entry, "s??i?i")
					if err != nil {
						c.reportError(fmt.Errorf("invalid UpdateInitiative slot #%d: %v", i, err))
						break
					}
					ilist = append(ilist, InitiativeSlot{
						Name:             fv[0].(string),
						IsHolding:        fv[1].(bool),
						HasReadiedAction: fv[2].(bool),
						CurrentHP:        fv[3].(int),
						IsFlatFooted:     fv[4].(bool),
						Slot:             fv[5].(int),
					})
				}

				payload.messageType = UpdateInitiative
				ch <- UpdateInitiativeMessagePayload{
					BaseMessagePayload: payload,
					InitiativeList:     ilist,
				}
			}

		case "L":
			//  _
			// | |
			// | |__
			// |____|
			//
			// L {<path> ...}
			//
			ch, ok := c.Subscriptions[LoadFrom]
			if ok {
				files, err := tcllist.ParseTclList(f[1])
				if err != nil {
					c.reportError(fmt.Errorf("invalid LoadFrom file list: %v", err))
					break
				}

				for _, file := range files {
					payload.messageType = LoadFrom
					ch <- LoadFromMessagePayload{
						BaseMessagePayload: payload,
						FileDefinition: FileDefinition{
							File:        file,
							IsLocalFile: true,
						},
					}
				}
			}

		case "LS":
			//  _    ___
			// | |  / __|
			// | |__\__ \
			// |____|___/
			//
			// LS
			// LS: <data>
			// LS. <#lines> <sha256>
			//
			_, ok := c.Subscriptions[LoadObject]
			_, ok2 := c.Subscriptions[AddImage]
			_, ok3 := c.Subscriptions[LoadFrom]
			if ok || ok2 || ok3 {
				if currentObj != "" {
					c.reportError(fmt.Errorf("LoadObject started before previous one ended"))
				}
				objBuffer = nil
				currentObj = c.reader.Text()
			}

		case "LS:":
			_, ok := c.Subscriptions[LoadObject]
			_, ok2 := c.Subscriptions[AddImage]
			_, ok3 := c.Subscriptions[LoadFrom]
			if ok || ok2 || ok3 {
				if currentObj == "" {
					c.reportError(fmt.Errorf("LoadObject message received before start-of-stream"))
					break
				}
				if len(f) != 2 {
					c.reportError(fmt.Errorf("LS: message field count wrong (%d)", len(f)))
					break
				}
				objBuffer = append(objBuffer, f[1])
			}

		case "LS.":
			ch, ok := c.Subscriptions[LoadObject]
			ich, iok := c.Subscriptions[AddImage]
			fch, fok := c.Subscriptions[LoadFrom]
			if ok || iok || fok {
				imgOk, err := c.finishStream(f, currentObj != "", len(objBuffer), streamChecksumStrings(objBuffer))
				if err != nil {
					c.reportError(fmt.Errorf("bad LoadObject transfer: %v", err))
				}
				if imgOk {
					objs, images, files, err := ParseObjects(objBuffer)
					if err != nil {
						c.reportError(fmt.Errorf("bad LoadObject transfer: %v", err))
						break
					}
					r := make([]string, 0, 2+len(objBuffer))
					r = append(r, currentObj)
					r = append(r, objBuffer...)
					r = append(r, payload.rawMessage...)
					for _, o := range objs {
						payload.rawMessage = r
						payload.messageType = LoadObject
						ch <- LoadObjectMessagePayload{
							BaseMessagePayload: payload,
							MapObject:          o,
						}
					}
					if iok {
						for _, im := range images {
							payload.rawMessage = r
							payload.messageType = AddImage
							ich <- AddImageMessagePayload{
								BaseMessagePayload: payload,
								ImageDefinition:    im,
							}
						}
					}
					if fok {
						for _, fi := range files {
							payload.rawMessage = r
							payload.messageType = LoadFrom
							fch <- LoadFromMessagePayload{
								BaseMessagePayload: payload,
								FileDefinition:     fi,
							}
						}
					}
				}
				currentObj = ""
				objBuffer = nil
			}

		case "M":
			//  __  __
			// |  \/  |
			// | |\/| |
			// |_|  |_|
			//
			// M {path ...}
			//
			ch, ok := c.Subscriptions[LoadFrom]
			if ok {
				files, err := tcllist.ParseTclList(f[1])
				if err != nil {
					c.reportError(fmt.Errorf("invalid LoadFrom file list: %v", err))
					break
				}

				for _, file := range files {
					payload.messageType = LoadFrom
					ch <- LoadFromMessagePayload{
						BaseMessagePayload: payload,
						FileDefinition: FileDefinition{
							File:        file,
							IsLocalFile: true,
						},
						Merge: true,
					}
				}
			}

		case "M@":
			//  __  __  ____
			// |  \/  |/ __ \
			// | |\/| / / _` |
			// |_|  |_\ \__,_|
			//         \____/
			//
			// M@ <id>
			//
			ch, ok := c.Subscriptions[LoadFrom]
			if ok {
				if len(f) != 2 {
					c.reportError(fmt.Errorf("invalid LoadFrom (M@) file list field count: %d", len(f)))
					break
				}
				payload.messageType = LoadFrom
				ch <- LoadFromMessagePayload{
					BaseMessagePayload: payload,
					FileDefinition: FileDefinition{
						File:        f[1],
						IsLocalFile: false,
					},
					Merge: true,
				}
			}

		case "M?":
			//  __  __ ___
			// |  \/  |__ \
			// | |\/| | /_/
			// |_|  |_|(_)
			//
			// M? <id>
			//
			ch, ok := c.Subscriptions[LoadFrom]
			if ok {
				if len(f) != 2 {
					c.reportError(fmt.Errorf("invalid LoadFrom (M?) message field count: %d", len(f)))
					break
				}
				payload.messageType = LoadFrom
				ch <- LoadFromMessagePayload{
					BaseMessagePayload: payload,
					FileDefinition: FileDefinition{
						File:        f[1],
						IsLocalFile: false,
					},
					CacheOnly: true,
				}
			}

		case "MARCO":
			//  __  __   _   ___  ___ ___
			// |  \/  | /_\ | _ \/ __/ _ \
			// | |\/| |/ _ \|   / (_| (_) |
			// |_|  |_/_/ \_\_|_\\___\___/
			//
			ch, ok := c.Subscriptions[Marco]
			if ok {
				payload.messageType = Marco
				ch <- MarcoMessagePayload{
					BaseMessagePayload: payload,
				}
			} else {
				// if the user isn't catching these, we'll respond
				// back to the server ourselves
				c.send("POLO")
			}

		case "MARK":
			//  __  __   _   ___ _  __
			// |  \/  | /_\ | _ \ |/ /
			// | |\/| |/ _ \|   / ' <
			// |_|  |_/_/ \_\_|_\_|\_\
			//
			// MARK <x> <y>
			//
			ch, ok := c.Subscriptions[Mark]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "sff")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid Mark message: %v", err))
				} else {
					payload.messageType = Mark
					ch <- MarkMessagePayload{
						BaseMessagePayload: payload,
						Coordinates: Coordinates{
							X: fv[1].(float64),
							Y: fv[2].(float64),
						},
					}
				}
			}

		case "OA":
			//   ___   _
			//  / _ \ /_\
			// | (_) / _ \
			//  \___/_/ \_\
			//
			// OA <id> {<key> <value> ...}
			//
			// XXX: Caveat: There is an ambiguous case for the SIZE
			// attribute. Since we don't know the type of object these
			// attributes are going into, we will take the most common
			// case and assume SIZE will be a creature attribute and return
			// it as a string value.
			// The only other object which uses SIZE is a tile object, and
			// I propose changing the attribute name for tiles to something
			// distinct.
			//
			ch, ok := c.Subscriptions[UpdateObjAttributes]
			if ok {
				if len(f) != 3 {
					c.reportError(fmt.Errorf("Invalid UpdateObjAttributes message: parameter list length %d", len(f)))
					break
				}
				kvlist, err := tcllist.ParseTclList(f[2])
				if err != nil {
					c.reportError(fmt.Errorf("Invalid UpdateObjAttributes message: %v", err))
					break
				}
				if (len(kvlist) % 2) != 0 {
					c.reportError(fmt.Errorf("Invalid UpdateObjAttributes message: kvlist length not even"))
					break
				}
				if len(kvlist) > 0 {
					payload.messageType = UpdateObjAttributes
					p := UpdateObjAttributesMessagePayload{
						BaseMessagePayload: payload,
						ObjID:              f[1],
						NewAttrs:           make(map[string]interface{}),
					}
					for i := 0; i < len(kvlist); i += 2 {
						atype, ok := attributeType(kvlist[i])
						if !ok {
							c.reportError(fmt.Errorf("Invalid %s attribute in UpdateObjAttributes; not recognized, assuming string value", kvlist[i]))
							atype = "string"
						}

						switch atype {
						case "enum":
							var b byte
							b, ok := enumToByte(kvlist[i], kvlist[i+1])
							if !ok {
								c.reportError(fmt.Errorf("Invalid %s value %s in UpdateObjAttributes; using default instead", kvlist[i], kvlist[i+1]))
								b = 0
							}
							p.NewAttrs[kvlist[i]] = b

						case "*RadiusAoE":
							var aep *RadiusAoE
							aoe, err := tcllist.Parse(kvlist[i+1], "sfs")
							if err != nil {
								c.reportError(fmt.Errorf("Invalid %s value %s in UpdateObjAttributes: %v; using nil", kvlist[i], kvlist[i+1], err))
							} else if aoe[0].(string) != "radius" {
								c.reportError(fmt.Errorf("Invalid %s value %s in UpdateObjAttributes: unknown radius type; using nil", kvlist[i], kvlist[i+1]))
							} else {
								aep = &RadiusAoE{
									Radius: aoe[1].(float64),
									Color:  aoe[2].(string),
								}
							}
							p.NewAttrs[kvlist[i]] = aep

						case "[]Coordinates":
							var coords []Coordinates
							cs, err := tcllist.ParseTclList(kvlist[i+1])
							if err != nil {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
							} else if len(cs)%2 != 0 {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: odd number of values", kvlist[i], kvlist[i+1]))
							} else {
								for j := 0; j < len(cs); j += 2 {
									px, err := strconv.ParseFloat(cs[j], 64)
									if err != nil {
										c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes at [%d](%s): %v", kvlist[i], kvlist[i+1], j, cs[j], err))
										px = 0.0
									}

									py, err := strconv.ParseFloat(cs[j+1], 64)
									if err != nil {
										c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes at [%d](%s): %v", kvlist[i], kvlist[i+1], j+1, cs[j+1], err))
										py = 0.0
									}

									coords = append(coords, Coordinates{
										X: px,
										Y: py,
									})
								}
								p.NewAttrs[kvlist[i]] = coords
							}

						case "*CreatureHealth":
							hp, err := newHealth(kvlist[i+1], nil)
							if err != nil {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
								hp = nil
							}
							p.NewAttrs[kvlist[i]] = hp

						case "[]string":
							ss, err := tcllist.ParseTclList(kvlist[i+1])
							if err != nil {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
							}
							p.NewAttrs[kvlist[i]] = ss

						case "float64":
							v, err := strconv.ParseFloat(kvlist[i+1], 64)
							if err != nil {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
								v = 0.0
							}
							p.NewAttrs[kvlist[i]] = v

						case "bool":
							var b bool
							if kvlist[i+1] == "" {
								b = false
							} else {
								var err error
								b, err = strconv.ParseBool(kvlist[i+1])
								if err != nil {
									c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
									b = false
								}
							}
							p.NewAttrs[kvlist[i]] = b

						case "int":
							var v int
							v, err := strconv.Atoi(kvlist[i+1])
							if err != nil {
								c.reportError(fmt.Errorf("invalid %s value %s in UpdateObjAttributes: %v", kvlist[i], kvlist[i+1], err))
								v = 0
							}
							p.NewAttrs[kvlist[i]] = v

						case "string":
							p.NewAttrs[kvlist[i]] = kvlist[i+1]

						default:
							c.reportError(fmt.Errorf("Internal error: Unrecognized type %s for attribute %s in UpdateObjAttributes; using string instead", atype, kvlist[i]))
							p.NewAttrs[kvlist[i]] = kvlist[i+1]
						}
					}
					ch <- p
				}
			}

		case "OA+":
			//   ___   _    _
			//  / _ \ /_\ _| |_
			// | (_) / _ \_   _|
			//  \___/_/ \_\|_|
			//
			// OA+ <objid> <key> {<value> ...}
			ch, ok := c.Subscriptions[AddObjAttributes]
			if ok {
				if len(f) != 4 {
					c.reportError(fmt.Errorf("Invalid AddObjAttributes message: parameter list length %d", len(f)))
					break
				}
				vlist, err := tcllist.ParseTclList(f[3])
				if err != nil {
					c.reportError(fmt.Errorf("Invalid AddObjAttributes message: %v", err))
					break
				}
				payload.messageType = AddObjAttributes
				ch <- AddObjAttributesMessagePayload{
					BaseMessagePayload: payload,
					ObjID:              f[1],
					AttrName:           f[2],
					Values:             vlist,
				}
			}

		case "OA-":
			//   ___     _
			//  / _ \   / \
			// | | | | / _ \  _____
			// | |_| |/ ___ \|_____|
			//  \___//_/   \_\
			//
			// OA- <objid> <key> {<value> ...}
			ch, ok := c.Subscriptions[RemoveObjAttributes]
			if ok {
				if len(f) != 4 {
					c.reportError(fmt.Errorf("Invalid RemoveObjAttributes message: parameter list length %d", len(f)))
					break
				}
				vlist, err := tcllist.ParseTclList(f[3])
				if err != nil {
					c.reportError(fmt.Errorf("Invalid RemoveObjAttributes message: %v", err))
					break
				}
				payload.messageType = RemoveObjAttributes
				ch <- RemoveObjAttributesMessagePayload{
					BaseMessagePayload: payload,
					ObjID:              f[1],
					AttrName:           f[2],
					Values:             vlist,
				}
			}

		case "PRIV":
			//  ___ ___ _____   __
			// | _ \ _ \_ _\ \ / /
			// |  _/   /| | \ V /
			// |_| |_|_\___| \_/
			//
			if len(f) > 1 {
				c.reportError(fmt.Errorf("privileged server operation denied: %s", f[1]))
			} else {
				c.reportError(fmt.Errorf("privileged server operation denied"))
			}

		case "PS":
			//  ___  ___
			// | _ \/ __|
			// |  _/\__ \
			// |_|  |___/
			//
			// PS <id> <color> <name> <area> <size> player|monster <gx> <gy> <reach?>
			//
			ch, ok := c.Subscriptions[PlaceSomeone]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "sssssssff?")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid PlaceSomeone message: %v", err))
					break
				}

				var ct byte
				ct = CreatureTypeUnknown
				switch fv[6].(string) {
				case "player":
					ct = CreatureTypePlayer
				case "monster":
					ct = CreatureTypeMonster
				default:
					c.reportError(fmt.Errorf("Invalid creature type %s", fv[6].(string)))
					break payloadSelection
				}

				payload.messageType = PlaceSomeone
				ch <- PlaceSomeoneMessagePayload{
					BaseMessagePayload: payload,
					CreatureToken: CreatureToken{
						BaseMapObject: BaseMapObject{
							ID: fv[1].(string),
						},
						Color:        fv[2].(string),
						Name:         fv[3].(string),
						Area:         fv[4].(string),
						Size:         fv[5].(string),
						CreatureType: ct,
						Gx:           fv[7].(float64),
						Gy:           fv[8].(float64),
						Reach:        fv[9].(bool),
					},
				}
			}

		case "ROLL":
			//  ___  ___  _    _
			// | _ \/ _ \| |  | |
			// |   / (_) | |__| |__
			// |_|_\\___/|____|____|
			//
			// ROLL <from> {<recipient> ...} <title> <result> {{<type> <value>} ...} <messageID>
			//
			ch, ok := c.Subscriptions[RollResult]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ssssisi")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid RollResult message: %v", err))
					break
				}

				var recipients []string
				var toGM bool
				var toAll bool
				var dlist []dice.StructuredDescription

				recips, err := tcllist.ParseTclList(fv[2].(string))
				if err != nil {
					c.reportError(fmt.Errorf("Invalid RollResult recipient list: %v", err))
					break
				}
				for _, recipient := range recips {
					switch recipient {
					case ToGMOnly:
						toGM = true
					case ToAll:
						toAll = true
					}
					recipients = append(recipients, recipient)
				}
				details, err := tcllist.ParseTclList(fv[5].(string))
				if err != nil {
					c.reportError(fmt.Errorf("Invalid503-574-4067 RollResult detail list: %v", err))
					break
				}
				for i, det := range details {
					ds, err := tcllist.ParseTclList(det)
					if err != nil || len(ds) != 2 {
						c.reportError(fmt.Errorf("Invalid RollResult detail element #%d \"%s\"", i, det))
						break payloadSelection
					}

					dlist = append(dlist, dice.StructuredDescription{
						Type:  ds[0],
						Value: ds[1],
					})
				}

				payload.messageType = RollResult
				ch <- RollResultMessagePayload{
					BaseMessagePayload: payload,
					ChatCommon: ChatCommon{
						Sender:     fv[1].(string),
						Recipients: recipients,
						MessageID:  fv[6].(int),
						ToAll:      toAll,
						ToGM:       toGM,
					},
					Title: fv[3].(string),
					Result: dice.StructuredResult{
						Result:  fv[4].(int),
						Details: dlist,
					},
				}
			}

		case "TB":
			//  _____ ___
			// |_   _| _ )
			//   | | | _ \
			//   |_| |___/
			//
			// TB 0|1
			ch, ok := c.Subscriptions[Toolbar]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "s?")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid Toolbar message: %v", err))
				} else {
					payload.messageType = Toolbar
					ch <- ToolbarMessagePayload{
						BaseMessagePayload: payload,
						Enabled:            fv[1].(bool),
					}
				}
			}

		case "TO":
			//  _____ ___
			// |_   _/ _ \
			//   | || (_) |
			//   |_| \___/
			//
			// TO <from> {<recipient> ...} <message> <messageID>
			//
			ch, ok := c.Subscriptions[ChatMessage]
			if ok {
				fv, err := tcllist.ConvertTypes(f, "ssssi")
				if err != nil {
					c.reportError(fmt.Errorf("Invalid ChatMessage message: %v", err))
					break
				}

				var recipients []string
				var toGM bool
				var toAll bool

				recips, err := tcllist.ParseTclList(fv[2].(string))
				if err != nil {
					c.reportError(fmt.Errorf("Invalid ChatMessage recipient list: %v", err))
					break
				}
				for _, recipient := range recips {
					switch recipient {
					case ToGMOnly:
						toGM = true
					case ToAll:
						toAll = true
					}
					recipients = append(recipients, recipient)
				}

				payload.messageType = ChatMessage
				ch <- ChatMessageMessagePayload{
					BaseMessagePayload: payload,
					ChatCommon: ChatCommon{
						Sender:     fv[1].(string),
						Recipients: recipients,
						MessageID:  fv[4].(int),
						ToAll:      toAll,
						ToGM:       toGM,
					},
					Text: fv[3].(string),
				}
			}

		default:
			//  _   _ _  _ _  ___  _  _____      ___  _
			// | | | | \| | |/ / \| |/ _ \ \    / / \| |
			// | |_| | .` | ' <| .` | (_) \ \/\/ /| .` |
			//  \___/|_|\_|_|\_\_|\_|\___/ \_/\_/ |_|\_|
			//
			ch, ok := c.Subscriptions[UNKNOWN]
			if ok {
				payload.messageType = UNKNOWN
				ch <- UnknownMessagePayload{
					BaseMessagePayload: payload,
				}
			} else {
				c.Logger.Printf("received unknown server message type: \"%v\"", f)
			}
		}
	}
	if err := c.reader.Err(); err != nil {
		done <- err
		return
	}
}

//
// report any sort of error to the client
//
func (c *Connection) reportError(e error) {
	c.LastError = e
	ch, ok := c.Subscriptions[ERROR]
	if ok {
		ch <- ErrorMessagePayload{
			BaseMessagePayload: BaseMessagePayload{
				rawMessage:  nil,
				messageType: ERROR,
			},
			Error: e,
		}
	} else {
		c.Logger.Printf("mapper error: %v", e)
	}
}

//
// listen and interact with the service until it's finished,
// then close our connection to it
//
func (c *Connection) interact() error {
	defer c.Close()
	c.debug(2, "interact() started")
	defer c.debug(2, "interact() ended")

	listenerDone := make(chan error, 1)
	go c.listen(listenerDone)

	for {
		//
		// Receive and buffer any messages to be sent out
		// to the server
		//
		select {
		case <-c.Context.Done():
			c.Logger.Printf("interact: context done, stopping")
			return nil
		case err := <-listenerDone:
			c.Logger.Printf("interact: listener done (%v), stopping", err)
			return err
		case packet := <-c.sendChan:
			c.sendBuf = append(c.sendBuf, packet)
		default:
		}
		//
		// Send the next outgoing message in the buffer
		//
		if c.writer != nil && len(c.sendBuf) > 0 {
			if c.DebuggingLevel >= 3 {
				c.debug(3, util.Hexdump([]byte(c.sendBuf[0])))
			}
			if c.DebuggingLevel >= 1 {
				c.debug(1, fmt.Sprintf("client->%q (%d)", c.sendBuf[0], len(c.sendBuf)))
			}
			if written, err := c.writer.WriteString(c.sendBuf[0]); err != nil {
				return fmt.Errorf("only wrote %d of %d bytes: %v", written, len(c.sendBuf[0]), err)
			}
			if err := c.writer.Flush(); err != nil {
				c.Logger.Printf("interact: unable to flush: %v", err)
				return err
			}
			c.sendBuf = c.sendBuf[1:]
		}
	}
}

//
// send a packet formed from the parameters passed here
// to the server. The server connection must be fully
// established at this point as this notifies the interact
// routine via the c.sendChan channel to pass the packet
// on to the server.
//
func (c *Connection) send(fields ...interface{}) error {
	packet, err := tcllist.ToDeepTclString(fields...)
	if err != nil {
		return err
	}
	if strings.ContainsAny(packet, "\n\r") {
		return fmt.Errorf("sent data may not contain a newline")
	}
	packet += "\n"
	select {
	case c.sendChan <- packet:
	default:
		return fmt.Errorf("unable to send to server (Dial() not running?)")
	}
	return nil
}

//
// send a packet formed from the string parameters passed here
// to the server directly. This should be used when the connection
// is NOT yet established fully, since this talks to the server
// directly on the assumption that the interact routine isn't
// available yet to handle this.
//
func (c *Connection) rawSend(fields ...string) error {
	packet, err := tcllist.ToTclString(fields)
	if err != nil {
		return err
	}
	packet += "\n"
	if c.DebuggingLevel >= 3 {
		c.debug(3, util.Hexdump([]byte(packet)))
	}
	if c.DebuggingLevel >= 1 {
		c.debug(1, fmt.Sprintf("client->%q (raw)", packet))
	}
	if written, err := c.writer.WriteString(packet); err != nil {
		return fmt.Errorf("only wrote %d of %d bytes: %v", written, len(packet), err)
	}
	if err := c.writer.Flush(); err != nil {
		c.Logger.Printf("rawSend: unable to flush: %v", err)
		return err
	}
	return nil
}

//
// Any time the subscription list changes,
// we need to call this to let the server know what kinds of
// messages the client wants to see.
//
func (c *Connection) filterSubscriptions() error {
	subList := []string{"AC", "DSM", "MARCO"} // these are unconditional
	for msg, _ := range c.Subscriptions {
		switch msg {
		case AddImage:
			subList = append(subList, "AI", "AI:", "AI.", "AI@", "LS", "LS:", "LS.")
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
		case Comment, UpdateProgress:
			subList = append(subList, "//")
		case LoadFrom:
			subList = append(subList, "L", "M", "M@", "M?", "LS", "LS:", "LS.")
		case LoadObject:
			subList = append(subList, "LS", "LS:", "LS.")
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
			subList = append(subList, "DD=", "DD:", "DD.")
		case UpdateInitiative:
			subList = append(subList, "IL")
		case UpdateObjAttributes:
			subList = append(subList, "OA")
		case UpdatePeerList:
			subList = append(subList, "CONN", "CONN:", "CONN.")
		case UpdateTurn:
			subList = append(subList, "I")
		}
	}

	sl, err := tcllist.ToTclString(subList)
	if err != nil {
		return err
	}

	return c.send("ACCEPT", sl)
}

//
// Tell the server to send us all possible messages.
//
func (c *Connection) unfilterSubscriptions() error {
	return c.send("ACCEPT", "*")
}

// @[00]@| GMA 4.3.5
// @[01]@|
// @[10]@| Copyright © 1992–2021 by Steven L. Willoughby
// @[11]@| (AKA Software Alchemy), Aloha, Oregon, USA. All Rights Reserved.
// @[12]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[13]@| License as described in the accompanying LICENSE file distributed
// @[14]@| with GMA.
// @[15]@|
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
