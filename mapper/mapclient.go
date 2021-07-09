/*
########################################################################################
#  _______  _______  _______                ___       ______      ______               #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___  \              #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   \  \             #
# | |      | || || || (___) |            / (_) (_       ___) /      ___) /             #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (      (___ (              #
# | | \_  )| |   | || (   ) | Game           ) (           ) \         ) \             #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _ /\___/  /             #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_)\______/              #
#                                                                                      #
########################################################################################
*/

// NewConnection(endpoint string, ...connOptions)
// .WithContext(ctx)
// .WithSubscription(c, event...)
// .WithAuthenticator(a)
//
// conn.Subscribe(c chan<-interface{}, event ...ServerMessage)
//   Receive server message on channel c when any of the named messages
//   arrive here. You may call this multiple times to add more events
//   to a channel.
//
// conn.Dial()
// conn.<message>(args...)
// conn.Context		(context object -- will use a default if one isn't given)
//

//
// Client interface for the mapper service.
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

import (
	"bufio"
	"context"
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
	GMAMapperProtocol              = 332     // @@##@@ auto-configured
	GMAVersionNumber               = "4.3.3" // @@##@@ auto-configured
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
	Subscriptions map[ServerMessage]chan<- MessagePayload

	// If true, we will always try to reconnect to the server if we
	// lose our connection.
	StayConnected bool

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

	// Private fields
	server   net.Conn
	reader   *bufio.Scanner
	writer   *bufio.Writer
	sendChan chan string
	sendBuf  []string
}

type CharacterDefinition struct {
	Name  string
	ObjID string
	Color string
	Area  string
	Size  string
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
func WithSubscription(ch chan<- MessagePayload, messages ...ServerMessage) connectionOption {
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
// Create a new server connection value which can then be used to
// manage our communication with the server.
//
// After the endpoint, you may specify any of the following options
// to define the behavior desired for this connection:
//   StayConnected(bool)
//   WithAuthenticator(a)
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
		Subscriptions: make(map[ServerMessage]chan<- MessagePayload),
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

func (c *Connection) Close() {
	c.server.Close()
	c.reader = nil
	c.writer = nil
	// TODO
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
func (conn *Connection) Subscribe(ch chan<- MessagePayload, messages ...ServerMessage) error {
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

//
// These are the server message to which a client may subscribe.
//
type ServerMessage byte

const (
	AddCharacter = iota
	AddImage
	AddObjAttributes
	AdjustView
	CacheFile
	ChatMessage
	Clear
	ClearChat
	ClearFrom
	CombatMode
	Comment
	//N	DefineDicePresets
	//N? FilterDicePresets
	//->LoadFrom LoadFile
	LoadFrom
	LoadObject
	Marco
	Mark
	PlaceSomeone
	//N	Polo
	//N QueryDicePresets
	QueryImage
	//N	QueryPeers
	RemoveObjAttributes
	//N	RollDice
	RollResult
	//N	Sync
	//N	SyncChat
	Toolbar
	UpdateClock
	UpdateDicePresets
	UpdateInitiative
	UpdateObjAttributes
	UpdatePeerList
	UpdateProgress
	UpdateStatusMarker
	UpdateTurn
	//N	WriteOnly
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
// Tell the server and peers about an image they can use.
//
func (c Connection) AddImage(idef ImageDefinition) error {
	if idef.IsLocalFile {
		return fmt.Errorf("Sending non-local files is not supported by this version.")
	}
	return c.send("AI@", idef.Name, fmt.Sprintf("%g", idef.Zoom), idef.File)
}

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
// AdjustView message: Change your displayed map view to the given
// fractions of the full canvas size.
//
type AdjustViewMessagePayload struct {
	BaseMessagePayload
	XView, YView float64
}

//
// CacheFile message: The client should take note of the given file
// which may be referred to in the future. It is recommended that the
// client pre-fetch the file into its cache.
//
type CacheFileMessagePayload struct {
	BaseMessagePayload
	FileDefinition
}

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

//
// Clear message: The client should remove the given object from
// its map. This ID may also be one of the following:
//   *   Remove all objects
//   E*  Remove all map elements
//   M*  Remove all monster tokens
//   P*  Remove all player tokens
//
type ClearMessagePayload struct {
	BaseMessagePayload
	ObjID string
}

//
// ClearChat message: clear the client's chat history
// XXX is this just a notice of the action?
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
// ClearFrom message: remove all elements in the map file
// referenced.
//
type ClearFromMessagePayload struct {
	BaseMessagePayload
	FileDefinition
}

//
// CombatMode message: indicate if combat mode should be in effect.
//
type CombatModeMessagePayload struct {
	BaseMessagePayload
	Enabled bool
}

//
// Comment message: a server comment to the client. The client is
// free to ignore these.
//
type CommentMessagePayload struct {
	BaseMessagePayload
	Text string
}

//
// FilterDicePresets message: remove die-roll presets whose names
// match a regular expression.
//
/*
type FilterDicePresetsMessagePayload struct {
	BaseMessagePayload
	NamePattern string
}
*/

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
// LoadObject message: load an object from the server.
//
type LoadObjectMessagePayload struct {
	BaseMessagePayload
	MapObject
}

//
// Marco message: the server is asking if we are still
// alive and responding. Reply by sending a Polo message.
//
// TODO: default handler for Marco should auto-send Polo.
//
type MarcoMessagePayload struct {
	BaseMessagePayload
}

//
// Mark message: visually mark the given map coordinates.
//
type MarkMessagePayload struct {
	BaseMessagePayload
	Coordinates
}

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
// QueryImage message: a peer wants to know where to find a given
// image and the server didn't know either. If you know the definition
// for that image, reply with an AddImage message of your own.
//
type QueryImageMessagePayload struct {
	BaseMessagePayload
	ImageDefinition
}

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

	// Was this die-roll sent for the GM to see only?
	BlindToGM bool
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
	Name        string
	Description string
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

type InitiativeSlot struct {
	Slot             int
	CurrentHP        int
	Name             string
	IsHolding        bool
	HasReadiedAction bool
	IsFlatFooted     bool
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
// UpdatePeerList message: notifies the client that the list of
// other connected peers has changed.
//
type UpdatePeerListMessagePayload struct {
	BaseMessagePayload
	PeerList []Peer
}
type Peer struct {
	Addr            string
	User            string
	Client          string
	LastPolo        int
	IsAuthenticated bool
	IsMe            bool
	IsMain          bool
	IsWriteOnly     bool
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

type StatusMarkerDefinition struct {
	Condition string
	Shape     string
	Color     string
}

//
// UpdateTurn message: declares whose turn it is in combat.
//
type UpdateTurnMessagePayload struct {
	BaseMessagePayload

	// The ObjID of the creature whose turn it is. This may also be one of:
	//   *Monsters*   All monsters are up now.
	//   (empty)      It is no one's turn now.
	ActorID string

	// The time lapsed so far since the start of combat.
	// Count is the initiative slot within the round.
	Hours, Minutes, Seconds, Rounds, Count int
}

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

	for {
		err = c.tryConnect()
		if err == nil {
			err = c.interact()
			// XXX err
		}

		if c.Context.Err() != nil || !c.StayConnected {
			break
		}
	}
	// XXX close out all our stuff
}

func (c *Connection) tryConnect() error {
	var err error
	var conn net.Conn
	var i uint

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
				return err
			}
			break syncloop

		case <-c.Context.Done():
			c.Logger.Printf("mapper: context cancelled; closing connections and aborting login...")
			c.Close() // this will abort the scanner in login()
			return fmt.Errorf("mapper: connection aborted by termination of context")
		}
	}

	c.Logger.Printf("mapper: login process finished")
	return nil
}

//
// We will block here on the assumption that the caller
// will put us in a goroutine to meet their own concurrency
// needs, if any.
//
func (c *Connection) login(done chan error) {
	defer close(done)

	c.Logger.Printf("mapper: Initial server negotiation...")
	syncDone := false
	authPending := false
	recCount := 0
	c.Preamble = nil

	for !syncDone && c.reader.Scan() {
		f, err := tcllist.ParseTclList(c.reader.Text())
		if err != nil {
			c.Logger.Printf("mapper: unable to parse message from server: %v", err)
			done <- err
			return
		}

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
				continue // XXX should this be fatal?
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
					done <- fmt.Errorf("authenticator required for connection")
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
				c.send("AUTH", response, c.Authenticator.Username, c.Authenticator.Client)
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
				_, err := tcllist.ConvertTypes(f, "ssss")
				if err != nil {
					c.Logger.Printf("mapper: INVALID server DSM data: %v: %v", f, err)
					continue
				}
				// add to status list
				c.Conditions[f[1]] = StatusMarkerDefinition{
					Condition: f[1],
					Shape:     f[2],
					Color:     f[3],
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
					c.Logger.Printf("mapper: sync %02d: Notedc Core API version %s", recCount, advertisedVersion)
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
	for authPending && c.reader.Scan() {
		f, err := tcllist.ParseTclList(c.reader.Text())
		if err != nil {
			c.Logger.Printf("mapper: unable to parse message from server: %v", err)
			done <- err
			return
		}

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
			done <- fmt.Errorf("server access denied")
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

		default:
			c.Logger.Printf("mapper: unexpected server message %v while waiting for authentication to complete", f)
		}
	}
	if err := c.reader.Err(); err != nil {
		done <- err
		return
	}

	c.filterSubscriptions()
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
// ?>ClearChat            CC *|<user> [""|<newmax>|-<#recents> [<messageID>]]
// <>Clear                CLR <objid>|*|E*|M*|P*|[<imagename>=]<name>
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
// ->CacheFile            M? <serverid>
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

//
// listen for, and dispatch, incoming server messages
//
func (c *Connection) listen(done chan error) {
	defer func() {
		close(done)
		c.Logger.Printf("mapper: stopped listening to server")
	}()

	strike := 0
	c.Logger.Printf("mapper: listening for server messages to dispatch...")
	for c.reader.Scan() {
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

		if len(f) == 0 {
			continue // skip blank lines
		}

		payload := BaseMessagePayload{
			rawMessage: f,
		}

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

		case "MARCO":
			//  __  __   _   ___  ___ ___
			// |  \/  | /_\ | _ \/ __/ _ \
			// | |\/| |/ _ \|   / (_| (_) |
			// |_|  |_/_/ \_\_|_\\___\___/
			//
			payload.messageType = Marco
			ch, ok := c.Subscriptions[Marco]
			if ok {
				ch <- MarcoMessagePayload{
					BaseMessagePayload: payload,
				}
			} else {
				// if the user isn't catching these, we'll respond
				// back to the server ourselves
				c.send("POLO")
			}

		default:
			//  _   _ _  _ _  ___  _  _____      ___  _
			// | | | | \| | |/ / \| |/ _ \ \    / / \| |
			// | |_| | .` | ' <| .` | (_) \ \/\/ /| .` |
			//  \___/|_|\_|_|\_\_|\_|\___/ \_/\_/ |_|\_|
			//
			payload.messageType = UNKNOWN
			ch, ok := c.Subscriptions[UNKNOWN]
			if ok {
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

	listenerDone := make(chan error, 1)
	go c.listen(listenerDone)

	for {
		//
		// Receive and buffer any messages to be sent out
		// to the server
		//
		select {
		case <-c.Context.Done():
			return nil
		case err := <-listenerDone:
			return err
		case packet := <-c.sendChan:
			c.sendBuf = append(c.sendBuf, packet)
		default:
		}
		//
		// Send the next outgoing message in the buffer
		//
		if c.writer != nil && len(c.sendBuf) > 0 {
			if written, err := c.writer.WriteString(c.sendBuf[0]); err != nil {
				return fmt.Errorf("only wrote %d of %d bytes: %v", written, len(c.sendBuf[0]), err)
			}
			if err := c.writer.Flush(); err != nil {
				return err
			}
			c.sendBuf = c.sendBuf[1:]
		}
	}
}

func (c Connection) send(fields ...string) error {
	packet, err := tcllist.ToTclString(fields)
	if err != nil {
		return err
	}
	packet += "\n"
	select {
	case c.sendChan <- packet:
	default:
		return fmt.Errorf("unable to send to server (Dial() not running?)")
	}
	return nil
}

func (c *Connection) filterSubscriptions() error {
	subList := []string{"AC", "DSM", "MARCO"} // these are unconditional
	for msg, _ := range c.Subscriptions {
		// XXX double-check these as we work on interact()
		switch msg {
		case AddImage:
			subList = append(subList, "AI", "AI:", "AI.", "AI@", "LS", "LS:", "LS.")
		case AddObjAttributes:
			subList = append(subList, "OA+")
		case AdjustView:
			subList = append(subList, "AV")
		case CacheFile:
			subList = append(subList, "M?")
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
			subList = append(subList, "L", "M", "M@")
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

	c.send("ACCEPT", sl)
	return nil
}

/*
//
// Send the provided fields to the server
//
func (c *Client) send(f ...string) error {
	cmd, err := tcllist.ToTclString(f)
	if err != nil {
		c.Logger.Printf("mapclient: error preparing data to send: %v", err)
		return err
	}
	if _, err := c.writer.WriteString(cmd + "\n"); err != nil {
		c.Logger.Printf("mapclient: error sending data \"%s\" to server: %v", cmd, err)
		return err
	}
	if err := c.writer.Flush(); err != nil {
		c.Logger.Printf("mapclient: error flushing data to server: %v", err)
		return err
	}
	return nil
}

func (c *Client) Close() error {
	if err := c.connection.Close(); err != nil {
		c.Logger.Printf("mapclient: Error closing connection to server: %v", err)
		return err
	}
	c.connection = nil
	return nil
}

type Logger interface {
	Printf(string, ...interface{})
}

type dialOptionSet struct {
	sleepFunction   func(time.Duration)
	timeout         time.Duration
	authenticator   *auth.Authenticator
	initialCommands []string
	retries         int
	debugProtocol   bool
	logger          *Logger
}

type dialOption func(*dialOptionSet)

func WithAuthenticator(a *auth.Authenticator) dialOption {
	return func(o *dialOptionSet) {
		o.authenticator = a
	}
}

func WithInitialCommands(cmds []string) dialOption {
	return func(o *dialOptionSet) {
		o.initialCommands = make([]string, len(cmds))
		for i, c := range cmds {
			o.initialCommands[i] = c
		}
	}
}

func WithLogger(l *Logger) dialOption {
	return func(o *dialOptionSet) {
		o.logger = l
	}
}

func WithTimeout(n time.Duration) dialOption {
	return func(o *dialOptionSet) {
		o.timeout = n
	}
}

func WithRetries(n int) dialOption {
	return func(o *dialOptionSet) {
		o.retries = n
	}
}

func WithSleepFunction(f func(time.Duration)) dialOption {
	return func(o *dialOptionSet) {
		o.sleepFunction = f
	}
}

func DebugProtocol(o *dialOptionSet) {
	o.debugProtocol = true
}
*/

/*
       When  instantiated, a SocketInterface object s establishes a connection
       to the specified host, logs in, and then becomes the point  of  contact
       to interact with the mapper service. The constructor and methods avail‐
       able are described below.
              s=SocketInterface(maphost, authenticator=None, initial_com‐
              mands=None, retries=1, diag_callback=None, sleep_callback=None,
              protocol_debug=None, subscribe_to=None, pass_callback=None,
              use_tk=False)

       Establishes a connection to the service at maphost (which is a tuple of
       (hostname_or_IP, port)).  If the connection fails, it  will  try  again
       until the number of retries specified has been exhausted. It will sleep
       for 1 second between attempts. Once the  connection  is  made,  if  the
       server  requires  authentication  and an authenticator object is given,
       that object will be used to negotiate authentication with  the  server.
       This  should  be  an instance of an Mapper.Authentication.Authenticator
       object.  If the password stored in authenticator is a  single  question
       mark  (“?”), then the user is prompted for their actual password, which
       is then inserted into the authenticator object before  using  it.   The
       server’s initial set of configuration data and/or commands are received
       and stored in the new SocketInterface object as appropriate.

       If initial_commands is provided, they are sent to  the  server  immedi‐
ately.

       If diag_callback is given, then all diagnostic messages will be sent to
       that function as the  sole  parameter.  Otherwise,  in  some  cases  an
       attempt  will  be  made to alert the user using a Tk dialog box, but if
       that fails (such as, the application does not have a Tk interface  run‐
       ning),  it  will fall back to printing diagnostic messages to the stan‐
       dard output.

       If sleep_callback is given, then any delays needed by the  object  will
       be performed by calling the provided function with two parameters: num‐
       ber of seconds to delay,  and  a  callable  function  which  should  be
       invoked when that delay has expired. Otherwise, sleep() will be called.

       If subscribe_to is specified, its value will  be  passed  to  the  sub‐
       scribe()  method  to  limit the set of messages the server will send to
       this client.

       If debugging a client, you can pass an open  file‐like  object  to  the
       protocol_debug  parameter.  Extra debugging information will be written
       to that file.

       If pass_callback is given, then any time the socket interface needs  to
       prompt  the  user for a password, it will invoke the function passed as
       the value to this parameter, with two arguments: a title and a  prompt.
       The value returned by this callback function should be the user’s pass‐
       word. If this is not given, at attempt will be made  to  prompt  for  a
       password  using  a Tk dialog box, and if that is unsuccessful, the user
       will be prompted to type it on the standard input.


       If use_tk has a true value, the Tk features will be  attempted;  other‐
       wise they never will.

       The following attributes are available after the object is constructed:

       s.protocol    The protocol version number used by the server.

       s.characters  A dictionary mapping  character  names  to  corresponding
                     MapCharacter  objects  describing  the  main  party (what
                     would be on the quick placement list of the Mapper tool’s
                     pop‐up context menu).

       s.conditions  A  dictionary mapping the name of a condition to a corre‐
                     sponding MapCondition object describing how a  particular
                     condition is to be drawn on a map display.

       s.subscribed_messages
                     The  set  of all server messages this client is currently
                     subscribed to receive.

       s.preamble    The initial greeting sent by the server.

   High‐Level Interface
       Once a SocketInterface object s is constructed, it should be logged  in
       to  a  server  and  ready  to exchange commands with it.  The following
       methods expose the API to the caller to interact with the map service:

       s.protocol_debug(handle=None)
              You can call this to turn on or off debugging  on  the  fly.  It
              performs  just  as the protocol_debug parameter to the construc‐
              tor.

       s.dispatch(handler, verbose=False, timeout=0)
              This method is analogous to the lower‐level  read_poll()  method
              (and  the  verbose  and  timeout  parameters are handled identi‐
              cally), but rather than simply returning the data  sent  by  the
              server as a string, it interprets the data according to the map‐
              per protocol and invokes the  corresponding  handler  method  in
              your handler object.

              To  make  one of these, create your own handler subclass derived
              from MapperProtocolHandler.  The base  class  is  full  of  stub
              methods  which  ignore the incoming server messages. In your own
              class, simply override all of the ones you care  to  receive  in
              your client.

              There  will  be  a method corresponding to each of the following
              methods used to send data to the server, with identical  calling
              semantics.   Thus,  for example, given that you can call s.tool‐
              bar(flag) to order peers to turn on or off  their  toolbars,  if
              you  receive  the toolbar command from the server, dispatch will
              invoke the method handler.toolbar(flag) for your handler  object
              to respond in whatever way you wish.

              In  addition to the protocol methods, two additional methods are
              defined:

              handler.unknown(command, args)
                     The server sent a command that we  don’t  recognize.  The
                     command  (as  a  string)  and its arguments (as a list of
                     strings) are passed as parameters. The  default  behavior
                     is to ignore this event.

              handler.error(description, data)
                     An  error was encountered as described in the description
                     parameter while trying to  process  the  incoming  server
                     data.  The raw data string as received is passed as data.
                     The default behavior is to ignore these errors.

       s.add_image(name, size, data=None, server_id=None)
              Define a graphic image for use by the mapper clients.  The  name
              is  what  the  clients  will use to refer to the image and these
              names must be globally unique (more precisely,  the  combination
              of name and size must be unique). In theory, this will replace a
              previous image definition with the  same  name  but  there’s  no
              guarantee  any  client will do so if it already had the previous
              one cached. This does not permanently store  the  image  server‐
              side.  It  only sends the data to clients listening at the time,
              or who later sync the data from the  server  (since  the  server
              will  keep  the  image definition in its state memory until that
              state is cleared).

              The size is a floating‐point value that gives the  magnification
              (zoom  level)  that this data defines for the image called name.
              The mapper currently supports values of 0.25, 0.5, 1.0, 2.0, and
              4.0  but  may  support others in the future. The data is a bytes

*/

// @[00]@| GMA 4.3.3
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
