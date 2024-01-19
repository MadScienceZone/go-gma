/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __    _______     _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \  (  ____ \   (  __   )       #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) | (    \/   | (  )  |       #
# | |      | || || || (___) | Assistant | (____        | | | (____     | | /   |       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | | (_____ \    | (/ /) |       #
# | | \_  )| |   | || (   ) |                 ) )      | |       ) )   |   / | |       #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_/\____) ) _ |  (__) |       #
# (_______)|/     \||/     \| Client    \______/ (_) \____/\______/ (_)(_______)       #
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
	"github.com/MadScienceZone/go-gma/v5/util"
	"github.com/newrelic/go-agent/v3/newrelic"
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
	GetAllowedClients() []PackageUpdate
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
// If the ctx context value is cancelled, the connection to the client will be closed and this routin will exit.
//
func (c *ClientConnection) ServeToClient(ctx context.Context, serverStarted, lastPing time.Time, nrApp *newrelic.Application) {
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

	select {
	case err = <-loginDone:
		if err != nil {
			c.Logf("client login failed: %v", err)
			time.Sleep(2 * time.Second)
			return
		}
		c.debugf(DebugIO, "login successfully completed")
		loginCancel()

	case <-ctx.Done():
		c.Logf("context cancelled; closing connection to client and aborting login")
		return
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
	//
	// To avoid pegging the CPU in a tight buzz loop which constantly keeps checking to
	// see if anything showed up in the buffer (*cough*pre-v8.5.3*cough*), we use the
	// bufferReadable channel as a sort of simple semaphore that (due to its blocking
	// behavior) we can wait for in the select, so this goroutine can fully sleep until either the semaphore (channel)
	// lights up to indicate the buffer needs service, or an
	// incoming message is received.  Sure, we could make a fancy
	// queue type which accepts an arbitrary amount of data and
	// blocks if we try to read from it when it's empty, but since the
	// buffer is managed just right here (and in a corresponding
	// bit of code on the client side), this is a simpler approach
	// that works for what we're doing now. It's possible this
	// will evolve later into something more sophisticated such as what was just described.

	toSend := make(chan string, 1)
	clientBufferCtx, cancelClientBuffer := context.WithCancel(ctx)
	defer cancelClientBuffer()

	go func(ctx context.Context) {
		c.Log("client buffer agent started")
		defer c.Log("client buffer agent stopped")
		bufferReadable := make(chan byte, 1)

		for {
			select {
			case packet := <-c.Conn.sendChan:
				if len(c.Conn.sendBuf) == 0 {
					select {
					case toSend <- packet:
						c.debugf(DebugIO, "moved packet directly to output channel (buffer empty and channel available now)")
						continue
					default:
						// buffer was empty, so populate it and signal that it has content
						c.Conn.sendBuf = append(c.Conn.sendBuf, packet)
						c.debugf(DebugIO, "moved packet to empty output buffer (depth %d)", len(c.Conn.sendBuf))
						bufferReadable <- 0
					}
				} else {
					// buffer already has contents, just add to it
					c.Conn.sendBuf = append(c.Conn.sendBuf, packet)
					c.debugf(DebugIO, "moved packet to existing output buffer (depth %d)", len(c.Conn.sendBuf))
				}

			case <-bufferReadable:
				if len(c.Conn.sendBuf) > 0 {
					select {
					case toSend <- c.Conn.sendBuf[0]:
						c.Conn.sendBuf = c.Conn.sendBuf[1:]
						c.debugf(DebugIO, "moved packet to output channel (depth %d)", len(c.Conn.sendBuf))
					default:
					}
				}
				if len(c.Conn.sendBuf) > 0 {
					bufferReadable <- 0
				}

			case <-ctx.Done():
				c.Log("buffer manager context cancelled")
				return
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
			func() {
				if InstrumentCode {
					if nrApp != nil {
						txn := nrApp.StartTransaction("handle_request")
						defer txn.End()
					}
				}
				switch p := packet.(type) {
				case CommentMessagePayload:

				case AddCharacterMessagePayload, ChallengeMessagePayload, ProtocolMessagePayload,
					UpdateDicePresetsMessagePayload, DeniedMessagePayload, GrantedMessagePayload,
					MarcoMessagePayload, PrivMessagePayload, ReadyMessagePayload, RedirectMessagePayload,
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
					if len(p.Messages) == 0 || slices.Index(p.Messages, "*") < 0 {
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

				case EchoMessagePayload:
					p.ReceivedTime = time.Now()
					c.Server.HandleServerMessage(p, c)

				default:
					c.Server.HandleServerMessage(packet, c)
				}
			}()
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
			if strings.HasPrefix(line, "REDIRECT ") {
				c.debugf(DebugIO, "preamble includes REDIRECT statement; not continuing further")
				time.Sleep(time.Second * 5)
				c.Conn.sendRaw("// Disconnecting now. See you on the other server!")
				c.Conn.Flush()
				time.Sleep(time.Second * 2)
				done <- fmt.Errorf("login cancelled due to redirect")
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
			ServerVersion: GoVersionNumber,
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
				allowed := c.Server.GetAllowedClients()

				if packet.Client != "" && allowed != nil {
					c.debugf(DebugAuth, "checking for allowed client version")
					for _, allowedClient := range allowed {
						if allowedClient.VersionRegex == nil || allowedClient.MinimumVersion == "" {
							c.debugf(DebugAuth, "no minimum version set for %s, skipping", allowedClient.Name)
							continue
						}

						fields := allowedClient.VersionRegex.FindStringSubmatch(packet.Client)
						if fields == nil {
							c.debugf(DebugAuth, "client %s does not match pattern %s for %s, trying next package", packet.Client, allowedClient.VersionPattern, allowedClient.Name)
							continue
						}

						if len(fields) != 2 {
							c.debugf(DebugAuth, "package %s pattern %s is invalid: MUST have exactly one capturing group", allowedClient.Name, allowedClient.VersionPattern)
							continue
						}

						if fields[1] == "" {
							c.debugf(DebugAuth, "client %s matches pattern %s for %s, but does not announce its version; denied", packet.Client, allowedClient.VersionPattern, allowedClient.Name)
							c.Conn.Send(Denied, DeniedMessagePayload{Reason: "disallowed client version"})
							_ = c.Conn.Flush()
							done <- fmt.Errorf("client denied")
							return
						}

						relVer, err := util.VersionCompare(fields[1], allowedClient.MinimumVersion)
						if err != nil {
							c.debugf(DebugAuth, "Error parsing client version %s and minimum version %s: %v", fields[1], allowedClient.MinimumVersion, err)
							c.Conn.Send(Denied, DeniedMessagePayload{Reason: "unable to understand client version"})
							_ = c.Conn.Flush()
							done <- fmt.Errorf("client version error")
							return
						}

						if relVer < 0 {
							c.debugf(DebugAuth, "%s client version %s is older than minimum version %s; denied", allowedClient.Name, fields[1], allowedClient.MinimumVersion)
							c.Conn.Send(Denied, DeniedMessagePayload{Reason: allowedClient.Name + " client is older than minimum allowed version"})
							_ = c.Conn.Flush()
							done <- fmt.Errorf("client version not allowed")
							return
						} else {
							c.debugf(DebugAuth, "%s client version %s is allowed", allowedClient.Name, fields[1])
						}
						break
					}
				}

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
			ServerVersion: GoVersionNumber,
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
