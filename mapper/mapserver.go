/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______      __                   #
# (  ____ \(       )(  ___  ) Game      (  ____ \   (  ____ \    /  \                  #
# | (    \/| () () || (   ) | Master's  | (    \/   | (    \/    \/) )                 #
# | |      | || || || (___) | Assistant | (____     | (____        | |                 #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \    (_____ \       | |                 #
# | | \_  )| |   | || (   ) |                 ) )         ) )      | |                 #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\____) ) _  __) (_                #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ (_) \____/                #
#                                                                                      #
########################################################################################
*/

package mapper

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"golang.org/x/exp/slices"
)

// Server-side analogue to the client code.

const IncomingClientPacketBacklog = 64 // How many incoming packets we can buffer waiting to be processed by the server before blocking

type MapServer interface {
	Log(messsage ...any)
	Logf(format string, args ...any)
	GetPersonalCredentials(user string) []byte
	GetClientPreamble() *ClientPreamble
	HandleServerMessage(MessagePayload, *ClientConnection)
	AddClient(*ClientConnection)
	RemoveClient(*ClientConnection)
	SendGameState(*ClientConnection)
}

// ClientPreamble contains information given to each client upon
// connection to the server.
type ClientPreamble struct {
	// Do we send the connecting client a full dump of the current game state?
	SyncData bool

	// Initial commands sent at the start, after authentication, and
	// at the end of the sign-on sequence.
	Preamble  []string
	PostAuth  []string
	PostReady []string
}

//
// ClientConnection describes the connection to a single
// client from the server's point of view.
//
type ClientConnection struct {
	// Client features enabled
	Features struct {
		DiceColorBoxes bool
	}

	// The client's host and port number
	Address string

	// The last time we heard a ping reply from the client
	LastPoloTime time.Time

	// The messages this client wishes to receive
	// (nil means to receive all messages)
	Subscriptions map[ServerMessage]bool

	// Authentication information for this user
	Auth *auth.Authenticator

	// Level of debugging requested for this client
	DebuggingLevel DebugFlags

	Server MapServer

	Conn MapConnection
	D    *dice.DieRoller
}

func NewClientConnection(socket net.Conn, opts ...ClientConnectionOption) (ClientConnection, error) {
	var err error

	newCon := ClientConnection{
		Address: socket.RemoteAddr().String(),
		Conn:    NewMapConnection(socket),
	}
	newCon.Conn.debug = newCon.debug
	newCon.Conn.debugf = newCon.debugf
	newCon.D, err = dice.NewDieRoller()
	if err != nil {
		return newCon, err
	}

	for _, o := range opts {
		if err = o(&newCon); err != nil {
			return newCon, err
		}
	}
	return newCon, nil
}

type ClientConnectionOption func(*ClientConnection) error

func WithServer(s MapServer) ClientConnectionOption {
	return func(c *ClientConnection) error {
		c.Server = s
		return nil
	}
}

func WithDebugFunctions(debugFunc func(DebugFlags, string), debugfFunc func(DebugFlags, string, ...any)) ClientConnectionOption {
	return func(c *ClientConnection) error {
		c.Conn.debug = debugFunc
		c.Conn.debugf = debugfFunc
		return nil
	}
}

func WithClientDebuggingLevel(l DebugFlags) ClientConnectionOption {
	return func(c *ClientConnection) error {
		c.DebuggingLevel = l
		return nil
	}
}

func WithClientAuthenticator(a *auth.Authenticator) ClientConnectionOption {
	return func(c *ClientConnection) error {
		c.Auth = a
		return nil
	}
}

func (c *ClientConnection) clientIdTag() string {
	return "[client " + c.IdTag() + "]"
}

func (c *ClientConnection) IdTag() string {
	if c == nil {
		return "(nil client)"
	}
	if c.Auth == nil || c.Auth.Username == "" {
		return c.Address
	}
	return fmt.Sprintf("%s (%s)", c.Address, c.Auth.Username)
}

func (c *ClientConnection) debug(level DebugFlags, msg string) {
	if c != nil && c.Server != nil && (c.DebuggingLevel&level) != 0 {
		for i, line := range strings.Split(msg, "\n") {
			if line != "" {
				c.Server.Logf("%s DEBUG%s%02d: %s", c.clientIdTag(), DebugFlagNames(level), i, line)
			}
		}
	}
}

func (c *ClientConnection) debugf(level DebugFlags, format string, args ...any) {
	c.debug(level, fmt.Sprintf(format, args...))
}

func (c *ClientConnection) Log(message ...any) {
	if c.Server != nil {
		message = append([]any{c.clientIdTag()}, message...)
		c.Server.Log(message...)
	}
}

func (c *ClientConnection) Logf(format string, args ...any) {
	if c.Server != nil {
		args = append([]any{c.clientIdTag()}, args...)
		c.Server.Logf("%s "+format, args...)
	}
}

func (c *ClientConnection) Close() {
	c.Conn.Close()
}

//
// ServeToClient is intended to be run in its own thread,
// and speaks to one client for the duration of its session.
//
func (c *ClientConnection) ServeToClient(ctx context.Context, serverStarted, lastPing time.Time) {
	if c == nil {
		return
	}
	var err error
	defer c.Close()

	c.debug(DebugIO, "serveToClient() started")
	defer c.debug(DebugIO, "serveToClient() ended")
	loginDone := make(chan error, 1)
	loginctx, loginCancel := context.WithTimeout(ctx, 1*time.Minute)
	defer loginCancel()
	c.LastPoloTime = time.Now()
	go c.loginClient(loginctx, loginDone, serverStarted, lastPing)

syncloop:
	for {
		select {
		case err = <-loginDone:
			if err != nil {
				c.Logf("client login failed: %v", err)
				time.Sleep(2 * time.Second)
				return
			}
			c.debugf(DebugIO, "login successfully completed")
			loginCancel()
			break syncloop

		case <-ctx.Done():
			c.Logf("context cancelled; closing connection to client and aborting login")
			return
		}
	}

	c.Server.AddClient(c)
	defer c.Server.RemoveClient(c)

	// Now we have a fully established connection to the client.
	// wait for each client command and then respond to it.

	// Start a listener which will just watch the incoming socket connection
	// and feed each received packet to the incomingPacket channel.
	// Note that this won't notice a cancelled context until the socket scanner
	// has hit an error, EOF, or has found an input line.

	incomingPacket := make(chan MessagePayload, IncomingClientPacketBacklog)
	clientListenerCtx, cancelClientListener := context.WithCancel(ctx)
	defer cancelClientListener()
	done := make(chan error, 1)

	go func(ctx context.Context, incomingPacket chan MessagePayload, done chan error) {
		c.Log("client listener started")
		defer c.Log("client listener stopped")

		for {
			p, err := c.Conn.Receive()
			if err != nil {
				done <- err
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				incomingPacket <- p
			}
		}
	}(clientListenerCtx, incomingPacket, done)

	// Start a buffer agent which will accept data on sendChan and buffer it locally
	// to send out to the client socket. We don't try to write it to the client here
	// in case the client is slow to receive our output since that could make us block
	// and might possibly lock up the server as other routines wait in line to talk to
	// this client.
	//
	// We want management of the buffer slice to happen only here in this one goroutine,
	// so we also feed the buffered data to the toSend channel as fast as it can accept
	// them (which is tied to the network socket being available and the client's reading
	// speed, so this is where data backs up in the buffer when the client isn't able to
	// accept as fast as we can send.

	toSend := make(chan string, 1)
	clientBufferCtx, cancelClientBuffer := context.WithCancel(ctx)
	defer cancelClientBuffer()

	go func(ctx context.Context) {
		c.Log("client buffer agent started")
		defer c.Log("client buffer agent stopped")

		for {
			select {
			case packet := <-c.Conn.sendChan:
				if len(c.Conn.sendBuf) == 0 {
					select {
					case toSend <- packet:
						c.debugf(DebugIO, "moved packet directly to output channel (buffer empty and channel available now)")
						continue
					default:
					}
				}
				c.Conn.sendBuf = append(c.Conn.sendBuf, packet)
				c.debugf(DebugIO, "moved packet to output buffer (depth %d)", len(c.Conn.sendBuf))

			case <-ctx.Done():
				c.Log("buffer manager context cancelled")
				return

			default:
				if len(c.Conn.sendBuf) > 0 {
					select {
					case toSend <- c.Conn.sendBuf[0]:
						c.Conn.sendBuf = c.Conn.sendBuf[1:]
						c.debugf(DebugIO, "moved packet to output channel (depth %d)", len(c.Conn.sendBuf))
					default:
					}
				}
			}
		}
	}(clientBufferCtx)

	// And now start a client sender which watches the output buffer and shuttles data from
	// the toSend channel to the client socket as fast as we can manage that.  This will block
	// when the client socket is full, which is fine, the buffer manager will collect the data
	// backing up until it's ready again.

	clientSenderCtx, cancelClientSender := context.WithCancel(ctx)
	defer cancelClientSender()

	go func(ctx context.Context, c *ClientConnection) {
		c.Log("client sender started")
		defer c.Log("client sender stopped")

		for {
			if c.Conn.writer == nil {
				c.Log("client writer gone; giving up now")
				return
			}
			select {
			case packet := <-toSend:
				if written, err := c.Conn.writer.WriteString(packet); err != nil {
					c.Logf("error sending %v to client (wrote %d): %v", packet, written, err)
				}
				if err := c.Conn.writer.Flush(); err != nil {
					c.Logf("error sending %v to client (in flush): %v", packet, err)
				}
				c.debugf(DebugIO, "sent %v", packet)
			case <-ctx.Done():
				return
			}
		}
	}(clientSenderCtx, c)

	c.Log("main loop entered")
	defer c.Log("Interaction with client ended")

mainloop:
	for {
		select {
		case <-ctx.Done():
			c.Log("client task signalled to stop")
			break mainloop

		case err := <-done:
			c.Logf("error reading from client: %v", err)
			break mainloop

		case packet := <-incomingPacket:
			// this will block signals to stop this client until processing of the current
			// packet is finished, but that shouldn't deadlock the I/O itself since that's
			// in other goroutines that don't rely on this code.
			if packet == nil {
				c.Log("EOF from client")
				break mainloop
			}
			c.debugf(DebugIO, "received packet %v", packet)
			switch p := packet.(type) {
			case CommentMessagePayload:

			case AddCharacterMessagePayload, ChallengeMessagePayload, ProtocolMessagePayload,
				UpdateDicePresetsMessagePayload, DeniedMessagePayload, GrantedMessagePayload,
				MarcoMessagePayload, PrivMessagePayload, ReadyMessagePayload,
				RollResultMessagePayload, UpdatePeerListMessagePayload, UpdateVersionsMessagePayload,
				WorldMessagePayload:
				c.Conn.Send(Priv, PrivMessagePayload{
					Command: p.RawMessage(),
					Reason:  "I get to send that command, not you.",
				})

			case AuthMessagePayload:
				c.Conn.Send(Priv, PrivMessagePayload{
					Command: p.RawMessage(),
					Reason:  "It's not the right time in our conversation for that.",
				})

			case AcceptMessagePayload:
				if len(p.Messages) == 0 || slices.Index[string](p.Messages, "*") < 0 {
					c.Subscriptions = nil
				} else {
					c.Subscriptions = make(map[ServerMessage]bool)
					for _, message := range p.Messages {
						if msgId, ok := ServerMessageByName[message]; ok {
							c.Subscriptions[msgId] = true
						}
					}
				}

			case AllowMessagePayload:
				c.Features.DiceColorBoxes = false

				for _, feature := range p.Features {
					if feature == "DICE-COLOR-BOXES" {
						c.Features.DiceColorBoxes = true
					}
				}

			case PoloMessagePayload:
				c.LastPoloTime = time.Now()

			default:
				c.Server.HandleServerMessage(packet, c)
			}
		}
	}

}

func (c *ClientConnection) loginClient(ctx context.Context, done chan error, serverStarted, lastPing time.Time) {
	defer close(done)
	if c == nil {
		done <- fmt.Errorf("loginClient called on nil connection")
		return
	}

	c.debug(DebugIO, "loginClient() started")
	defer c.debug(DebugIO, "loginClient() ended")

	c.Log("initial client negotiation...")
	c.debug(DebugIO, "fetching preamble data from generator")
	preamble := c.Server.GetClientPreamble()
	if preamble != nil {
		c.debugf(DebugIO, "got %d initial command(s)", len(preamble.Preamble)+len(preamble.PostAuth)+len(preamble.PostReady))
	} else {
		c.Log("got nil preamble data!")
	}
	c.Conn.Send(Protocol, GMAMapperProtocol)
	if preamble != nil {
		for i, line := range preamble.Preamble {
			c.debugf(DebugIO, "preamble line %d: %s", i, line)
			c.Conn.sendRaw(line)
			if err := c.Conn.Flush(); err != nil {
				done <- err
				return
			}
		}
	}

	// Authentication challenge
	if c.Auth != nil {
		c.debug(DebugIO, "issuing authentication challenge")
		challenge, err := c.Auth.GenerateChallengeBytes()
		if err != nil {
			done <- fmt.Errorf("error generating authentication challenge: %v", err)
			return
		}
		c.Conn.Send(Challenge, ChallengeMessagePayload{
			Protocol:      GMAMapperProtocol,
			Challenge:     challenge,
			ServerStarted: serverStarted,
			ServerActive:  lastPing,
			ServerTime:    time.Now(),
		})
		if err := c.Conn.Flush(); err != nil {
			done <- err
			return
		}

		reply := make(chan AuthMessagePayload, 1)
		go func(reply chan AuthMessagePayload) {
			for {
				packet, err := c.Conn.Receive()
				if err != nil {
					c.Logf("error reading auth response from client: %v; stopping", err)
					return
				}
				if packet == nil {
					c.Log("EOF reading auth response from client; stopping")
					return
				}
				switch p := packet.(type) {
				case ErrorMessagePayload:
					c.Logf("error reading auth response from client: %v", p.Error)
				case PoloMessagePayload:
					continue
				case AuthMessagePayload:
					reply <- p
					return
				}
				c.Logf("Invalid packet of type %T received", packet)
			}
		}(reply)

	awaitUserAuth:
		for {
			select {
			case <-ctx.Done():
				c.Log("Timeout/cancel while waiting for authentication from client")
				c.Conn.Send(Denied, DeniedMessagePayload{Reason: "Life is short indeed / I don't have time for waiting / For you to log in"})
				_ = c.Conn.Flush()
				time.Sleep(1 * time.Second)
				done <- fmt.Errorf("timeout waiting for client auth")
				return

			case packet := <-reply:
				c.debugf(DebugAuth, "received client authentication %v", packet)
				if newSecret := c.Server.GetPersonalCredentials(packet.User); newSecret != nil {
					c.Auth.SetSecret(newSecret)
				}
				success, err := c.Auth.ValidateResponseBytes(packet.Response)
				if err != nil {
					c.Logf("error trying to authenticate: %v", err)
					done <- err
				}
				if success {
					c.Auth.Client = packet.Client
					if c.Auth.GmMode {
						c.Auth.Username = "GM"
						c.Logf("granting GM privileges to client")
					} else {
						if packet.User == "GM" {
							c.Auth.Username = "unknown"
						} else {
							c.Auth.Username = packet.User
						}
					}
					c.Conn.Send(Granted, GrantedMessagePayload{User: c.Auth.Username})
					break awaitUserAuth
				} else {
					c.Conn.Send(Denied, DeniedMessagePayload{Reason: "login incorrect"})
					_ = c.Conn.Flush()
					done <- fmt.Errorf("access denied")
					return
				}
			}
		}
	} else {
		c.debug(DebugIO, "proceeding without authentication")
		c.Conn.Send(Challenge, ChallengeMessagePayload{
			Protocol:      GMAMapperProtocol,
			ServerStarted: serverStarted,
			ServerActive:  lastPing,
		})
		if err := c.Conn.Flush(); err != nil {
			done <- err
		}
	}

	if preamble != nil {
		for i, line := range preamble.PostAuth {
			c.debugf(DebugIO, "post-auth preamble line %d: %s", i, line)
			c.Conn.sendRaw(line)
			if err := c.Conn.Flush(); err != nil {
				done <- err
			}
		}
	}

	c.debug(DebugIO, "signalling end of login step")
	c.Conn.Send(Ready, nil)
	if err := c.Conn.Flush(); err != nil {
		done <- err
	}
	done <- nil // login is done at this point, let the caller start the normal client listener for I/O
	if preamble != nil {
		for i, line := range preamble.PostReady {
			c.debugf(DebugIO, "post-ready preamble line %d: %s", i, line)
			c.Conn.sendRaw(line)
		}
		if preamble.SyncData {
			c.Log("syncing client to current game state...")
			c.Server.SendGameState(c)
			c.Log("syncing done")
		}
	}
}
