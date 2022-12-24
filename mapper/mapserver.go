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

package mapper

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"golang.org/x/exp/slices"
)

// Server-side analogue to the client code.

type MapServer interface {
	Log(messsage ...any)
	Logf(format string, args ...any)
	GetPreamble() ([]string, []string, []string, bool)
	GetPersonalCredentials(user string) []byte
	HandleServerMessage(payload MessagePayload)
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

	server MapServer
	conn   MapConnection
}

func NewClientConnection(socket net.Conn, opts ...ClientConnectionOption) (ClientConnection, error) {
	newCon := ClientConnection{
		Address: socket.RemoteAddr().String(),
		conn:    NewMapConnection(socket),
	}

	for _, o := range opts {
		if err := o(&newCon); err != nil {
			return newCon, err
		}
	}
	return newCon, nil
}

type ClientConnectionOption func(*ClientConnection) error

func WithServer(s MapServer) ClientConnectionOption {
	return func(c *ClientConnection) error {
		c.server = s
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
	return "[client " + c.idTag() + "]"
}

func (c *ClientConnection) idTag() string {
	if c == nil {
		return "(nil client)"
	}
	if c.Auth == nil || c.Auth.Username == "" {
		return c.Address
	}
	return fmt.Sprintf("%s (%s)", c.Address, c.Auth.Username)
}

func (c *ClientConnection) debug(level DebugFlags, msg string) {
	if c != nil && c.server != nil && (c.DebuggingLevel&level) != 0 {
		for i, line := range strings.Split(msg, "\n") {
			if line != "" {
				c.server.Logf("%s DEBUG%s%02d: %s", c.clientIdTag(), DebugFlagNames(level), i, line)
			}
		}
	}
}

func (c *ClientConnection) debugf(level DebugFlags, format string, args ...any) {
	c.debug(level, fmt.Sprintf(format, args...))
}

func (c *ClientConnection) Log(message ...any) {
	if c.server != nil {
		message = append([]any{c.clientIdTag()}, message...)
		c.server.Log(message...)
	}
}

func (c *ClientConnection) Logf(format string, args ...any) {
	if c.server != nil {
		args = append([]any{c.clientIdTag()}, args...)
		c.server.Logf("%s "+format, args...)
	}
}

func (c *ClientConnection) Close() {
	c.conn.Close()
}

//func (c *ClientConnection) EmergencyReject(message string) {
//	c.Logf("performing emergency reject (%s)", message)
//	c.Conn.Send(mapper.Protocol, mapper.GMAMapperProtocol)
//	c.Conn.Send(mapper.Denied, mapper.DeniedMessagePayload{Reason: message})
//	c.Conn.Close()
//}
//
//
// serveToClient is intended to be run in its own thread,
// and speaks to one client for the duration of its session.
//
func (c *ClientConnection) ServeToClient(ctx context.Context) {
	if c == nil {
		return
	}
	var err error
	defer c.Close()

	c.debug(DebugIO, "serveToClient() started")
	defer c.debug(DebugIO, "serveToClient() ended")
	loginDone := make(chan error, 1)
	loginctx, _ := context.WithTimeout(ctx, 1*time.Minute)
	go c.loginClient(loginctx, loginDone)

syncloop:
	for {
		select {
		case err = <-loginDone:
			if err != nil {
				c.Logf("client login failed: %v", err)
				return
			}
			break syncloop

		case <-ctx.Done():
			c.Logf("context cancelled; closing connection to client and aborting login")
			return

		case <-loginctx.Done():
			c.Logf("timeout/cancel of login negotiation")
			return
		}
	}

	// Now we have a fully established connection to the client.
	// wait for each client command and then respond to it.
	incomingPacket := make(chan MessagePayload)
	go func(incomingPacket chan MessagePayload, done chan error) {
		for {
			incomingPacket <- c.conn.Receive(done)
		}
	}(incomingPacket, done)

mainloop:
	for {
		select {
		case <-loginctx.Done():
			// we no longer care

		case <-ctx.Done():
			// ok, this one we care about.
			c.Logf("client task signalled to stop")
			break mainloop

		case packet := <-incomingPacket:
			switch p := packet.(type) {
			case CommentMessagePayload:

			case AddCharacterMessagePayload, ChallengeMessagePayload, ProtocolMessagePayload,
				UpdateDicePresetsMessagePayload, DeniedMessagePayload, GrantedMessagePayload,
				MarcoMessagePayload, PrivMessagePayload, ReadyMessagePayload,
				RollResultMessagePayload, UpdateVersionsMessagePayload, WorldMessagePayload:
				c.conn.Send(Priv, PrivMessagePayload{
					Command: p.RawMessage(),
					Reason:  "I get to send that command, not you.",
				})

			case AuthMessagepayload:
				c.conn.Send(Priv, PrivMessagePayload{
					Command: p.RawMessage(),
					Reason:  "It's not the right time in our conversation for that.",
				})

			case AcceptMessagePayload:
				if len(p.Messages) == 0 || slices.Index[string](p.Messages, "*") < 0 {
					c.Subscriptions = nil
				} else {
					c.Subscriptions = make(map[ServerMessage]bool)
					for _, message := range c.Messages {
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
				c.server.HandleServerMessage(payload)
			}
		}
	}

	c.Logf("Interaction with client ended")
}

func (c *ClientConnection) loginClient(ctx context.Context, done chan error) {
	defer close(done)
	if c == nil {
		done <- fmt.Errorf("loginClient called on nil connection")
		return
	}

	c.debug(DebugIO, "loginClient() started")
	defer c.debug(DebugIO, "loginClient() ended")

	c.Log("initial client negotiation...")
	preamble, postAuth, postReady, syncData := c.server.GetPreamble()
	c.conn.Send(Protocol, GMAMapperProtocol)
	for i, line := range preamble {
		c.debugf(DebugIO, "preamble line %d: %s", i, line)
		c.conn.sendRaw(line)
		if err := c.conn.Flush(); err != nil {
			done <- err
			return
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
		c.conn.Send(Challenge, ChallengeMessagePayload{
			Protocol:  GMAMapperProtocol,
			Challenge: challenge,
		})
		if err := c.conn.Flush(); err != nil {
			done <- err
			return
		}

		reply := make(chan AuthMessagePayload)
		go func(reply chan AuthMessagePayload) {
			for {
				packet := c.conn.Receive(done)
				if packet == nil {
					c.Log("error reading auth response from client; stopping")
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
				c.conn.Send(Denied, DeniedMessagePayload{Reason: "Login timed out"})
				_ = c.conn.Flush()
				done <- fmt.Errorf("timeout waiting for client auth")
				return

			case packet := <-reply:
				c.debugf(DebugAuth, "received client authentication %v", packet)
				if newSecret := c.server.GetPersonalCredentials(packet.User); newSecret != nil {
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
					c.conn.Send(Granted, GrantedMessagePayload{User: c.Auth.Username})
					break awaitUserAuth
				} else {
					c.conn.Send(Denied, DeniedMessagePayload{Reason: "login incorrect"})
					_ = c.conn.Flush()
					done <- fmt.Errorf("access denied")
					return
				}
			}
		}

		//TODO Flush() should allow a timeout

	} else {
		c.debug(DebugIO, "proceeding without authentication")
		c.conn.Send(Challenge, ChallengeMessagePayload{
			Protocol: GMAMapperProtocol,
		})
		if err := c.conn.Flush(); err != nil {
			done <- err
		}
	}

	for i, line := range postAuth {
		c.debugf(DebugIO, "post-auth preamble line %d: %s", i, line)
		c.conn.sendRaw(line)
		if err := c.conn.Flush(); err != nil {
			done <- err
		}
	}

	c.debug(DebugIO, "signalling end of login step")
	c.conn.Send(Ready, nil)
	if err := c.conn.Flush(); err != nil {
		done <- err
	}
	for i, line := range postReady {
		c.debugf(DebugIO, "post-ready preamble line %d: %s", i, line)
		c.conn.sendRaw(line)
		if err := c.conn.Flush(); err != nil {
			done <- err
		}
	}
	if syncData {
		c.Log("syncing client to current game state...")
		// TODO
		c.Log("syncing done")
	}
	done <- nil
}

//	go app.ioRunner(&client)
//	app.Debugf(DebugIO, "serveToClient from %v, conn=%q", client.Address, client.Conn)
//	client.Debug(DebugIO, "starting session for client")
//	client.Conn.Send(mapper.Protocol, mapper.GMAMapperProtocol)
//	client.Debug(DebugIO, "end of session")
//}
//
//func (app *Application) ioRunner(client *ClientConnection) {
//	for {
//		select {
//		case packet := <-client.Conn.sendChan:
//			client.Conn.sendBuf = append(client.Conn.sendBuf, packet)
//		default:
//			if client.Conn.writer != nil && len(client.Conn.sendBuf) > 0 {
//				client.Debug(DebugIO, util.Hexdump([]byte(client.Conn.sendBuf[0])))
//				client.Debugf(DebugIO, "client->%q (%d)", client.Conn.sendBuf[0], len(client.Conn.sendBuf))
//				if written, err := client.Conn.writer.WriteString(client.Conn.sendBuf[0]); err != nil {
//					client.Logf("only wrote %d of %d bytes: %v", written, len(client.Conn.sendBuf[0]), err)
//					// TODO abort?
//				}
//				if err := client.Conn.writer.Flush(); err != nil {
//					client.Logf("ioRunner: unable to flush: %v", err)
//					// TODO abort?
//				}
//				client.Conn.sendBuf = client.Conn.sendBuf[1:]
//			}
//		}
//	}
//}
