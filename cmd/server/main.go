/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______      _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \    (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \   | (  )  |              #
# | |      | || || || (___) | Assistant | (____        ___) /   | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )         ) \   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\___/  / _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ (_)(_______)              #
#                                                                                      #
########################################################################################
*/
//
////////////////////////////////////////////////////////////////////////////////////////
//                                                                                    //
//                                     MapService                                     //
//                                                                                    //
// Inter-map communication service.  Transmits map updates to other maps and allows   //
// API callers to inject events to be sent to all maps.                               //
//                                                                                    //
// This is a re-implementation from scratch of the Python GMA game server (as         //
// originally implemented in the Mapper.MapService module), in the Go language, in    //
// the hopes that this will provide better performance than the Python version.  It   //
// was also done out of personal interest to explore Go design features for a server  //
// such as this one.                                                                  //
//                                                                                    //
////////////////////////////////////////////////////////////////////////////////////////

/*
Server coordinates the actions of connected mappers, the
GM console, and other clients.

The individual mapper(6) clients used by players in a game may keep in contact with one another so that they all display the same contents.
A change made on one client (moving a creature token or adding a room, say) appears on all the others.
This is accomplished by starting a server process and having all of the mapper clients connect to it via their −−host and −−port options.

Once connected, the server will send an initial greeting that may define a list of player character tokens to appear on the mapper context menus, or any other useful information the clients need to have at startup time.
It may, at the GM’s option, even initialize the client to show the full current game state.

From that point forward, the server relays traffic between the clients, so they communicate with each other via the service.
The server also tracks the commands it sees, so that it maintains a notion of the current state of the game.
Clients may re-sync with the server in case they restart or otherwise miss any updates so they match the server’s state.
The server may respond directly to some client queries if it knows the answer rather than referring the query to the other clients.

To guard against nuisance or malicious port scans and other superfluous connections, the server will automatically drop any clients which don’t authenticate within a short time.
(In actual production use, we have observed some automated agents which connected and then sat idle for hours, if we didn’t terminate their connections. This prevents that.)

Options:
   server [−debug flags] [−endpoint [hostname]:port] [−init−file path] [−log−file path] [−password−file path] −sqlite path [−telemetry−log path]

   -debug flags
      Add debugging information to the log file. The flags value is a comma-separated
      list of debugging information to be included, from the following list:
         all      All possible debugging information.
         none     No debugging information (this cancels any previously-specified
                  debug flags, but more may be added after this).
         auth     Authentication operations.
         db       Database operations.
         events   Background events.
         i/o      Input/output operations.
         init     Client initialization.
         messages Message traffic between the server and clients.
         misc     Miscellaneous debugging.
         state    Changes to the game state.

   -endpoint [hostname]:port
      Accept incoming client connections on the specified TCP port. (Default ":2323")

   -init-file path
      Initialization file which controls the initial client negotiation upon first
      connection to the server.

   -log-file path
      Write a log of server actions to the specified file. (Default "-", which means
      to send to standard output.)

   -password-file path
      Enable server authentication with the set of passwords in the specified file.
      Each line of the file holds a plaintext password, in the following format:
          general-user-password
          gm-only-password
          user1:password1
          user2:password2
          user3:password3
      Only the first line is required.

   -sqlite path
      Specifies the file name of a sqlite database used to keep persistent data used
      by the server. If path does not exist, server will create a new database with that
      name.

   -telemetry-log path
      If server was compiled to send performance telemetry data, a debugging log of that
      data is recorded in the specified file.

See the full documentation in the accompanying manual file man/man6/server.6.pdf (or run “gma man go server” if you have the GMA Core package installed as well as Go-GMA).

See also the server protocol specification in the man/man7/mapper-protocol.7.pdf of the GMA-Mapper package (or run “gma man mapper-protocol”). This is also printed in Appendix F of the GMA Game Master's Guide.
*/
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadScienceZone/go-gma/v5/mapper"
	"github.com/newrelic/go-agent/v3/newrelic"
)

//
// Auto-configured values
//

const GoVersionNumber="5.3.0" // @@##@@

//
// eventMonitor responds to signals and timers that affect our overall operation
// independent of client requests.
//
func eventMonitor(sigChan chan os.Signal, stopChan chan int, app *Application) {
	ping_signal := time.NewTicker(1 * time.Minute)
	app.LastPing = time.Now()

	for {
		select {
		case s := <-sigChan:
			app.Logf("received signal %v", s)
			switch s {
			case syscall.SIGHUP:
				app.Debug(DebugEvents, "SIGHUP; sending STOP signal to application")
				stopChan <- 1

			case syscall.SIGUSR1:
				app.Debug(DebugEvents, "SIGUSR1; reloading configuration data")
				app.clientPreamble.reload <- 0
				if err := app.refreshAuthenticator(); err != nil {
					app.Logf("WARNING: authenticator initialization file reload failed: %v", err)
					app.Log("WARNING: client credentials may be incomplete or incorrect now")
				}

			case syscall.SIGUSR2:
				app.Debug(DebugEvents, "SIGUSR2 (dump database out to logfile)")
				if err := app.LogDatabaseContents(); err != nil {
					app.Logf("Error dumping database: %v", err)
				}

			case syscall.SIGINT:
				app.Debug(DebugEvents, "SIGINT; sending STOP signal to application")
				stopChan <- 1
				// Make a quick effort to shut down as fast as possible
				// by terminating all client connections immediately.
				//				log.Printf("EMERGENCY SHUTDOWN INITIATED")
				//				ms.AcceptIncoming = false
				//				for i, client := range ms.Clients {
				//					log.Printf("Terminating client %v from %s", i, client.ClientAddr)
				//					client.Connection.Close()
				//				}
			}

		case <-ping_signal.C:
			app.Debug(DebugEvents, "ping timer expired")
			app.LastPing = time.Now()
			app.SendToAll(mapper.Marco, nil)
		}
	}
}

func generateMessageIDs(logf func(format string, args ...any), c chan int) {
	// Start off with the time on the clock, on the assumption
	// that on average there won't be more than a chat message per
	// second since the server was started, so when the server is
	// restarted, this should give us a safe margin to start a new
	// set of IDs. It's simplistic, but works for our purposes.
	var nextMessageID int = int(time.Now().Unix())
	logf("starting messsageID generator at %v", nextMessageID)
	defer logf("stopping messageID generator")

	// Now just feed these numbers to the channel as fast as they are
	// consumed.
	for {
		c <- nextMessageID
		nextMessageID++
	}
}

func main() {
	var nrApp *newrelic.Application
	var err error

	app := *NewApplication()
	app.ServerStarted = time.Now()
	app.LastPing = time.Now()
	app.Logger.SetPrefix("go-gma-server: ")
	if err := app.GetAppOptions(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %v\n", err)
		os.Exit(1)
	}
	app.Logf("Server %s started", GoVersionNumber)
	app.Logf("Implements protocol %d (library supports minimum %d, maximum %d)",
		mapper.GMAMapperProtocol,
		mapper.MinimumSupportedMapProtocol,
		mapper.MaximumSupportedMapProtocol)

	go generateMessageIDs(app.Logf, app.MessageIDGenerator)
	go app.managePreambleData()
	go app.manageClientList()
	go app.manageGameState()
	go app.announceClients()

	/* instrumentation */
	// set the following environment variables for the New Relic
	// Go Agent:
	//    NEW_RELIC_APP_NAME = the name you want to appear in the datasets
	//    NEW_RELIC_LICENSE_KEY = your license key
	//    NEW_RELIC_METADATA_RELEASE_TAG = application release
	//
	if InstrumentCode {
		app.Log("application performance metrics telemetry reporting enabled")
		if err = os.Setenv("NEW_RELIC_METADATA_SERVICE_VERSION", GoVersionNumber); err != nil {
			app.Logf("unable to set version metadata: %v", err)
		}
		nrApp, err = newrelic.NewApplication(
			newrelic.ConfigAppName("gma-server"),
			newrelic.ConfigFromEnvironment(),
			newrelic.ConfigCodeLevelMetricsEnabled(true),
			newrelic.ConfigCodeLevelMetricsPathPrefixes("go-gma/"),
			newrelic.ConfigCodeLevelMetricsRedactPathPrefixes(false),
			newrelic.ConfigDebugLogger(app.NrLogFile),
		)
		if err != nil {
			app.Logf("unable to start instrumentation: %v", err)
			os.Exit(1)
		}
		defer func() {
			app.Logf("waiting for instrumentation to finish (max 30 sec) ...")
			nrApp.Shutdown(30 * time.Second)
		}()
	}
	/*
		for {
			func() {
				if InstrumentCode {
					defer nrApp.StartTransaction("testing").End()
				}
				time.Sleep(10 * time.Second)
			}()
		}
	*/

	if err := app.dbOpen(); err != nil {
		app.Logf("unable to open database: %v", err)
		os.Exit(1)
	}
	defer app.dbClose()

	// TODO instrumentation
	/*
		txn := nrapp.StartTransaction("background")
		defer txn.End()
		// do stuff
	*/

	// start listening to incoming port
	incoming, err := net.Listen("tcp", app.Endpoint)
	if err != nil {
		app.Logf("unable to open incoming TCP %s: %v", app.Endpoint, err)
		os.Exit(2)
	}
	app.Logf("Listening on %s", app.Endpoint)
	defer func() {
		if err := incoming.Close(); err != nil {
			app.Logf("failure closing incoming socket: %v", err)
		}
	}()

	sigChannel := make(chan os.Signal, 1)
	stopChannel := make(chan int, 1)
	signal.Notify(sigChannel, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT)

	//expiredClients := make(chan *ClientConnection, 16)
	go eventMonitor(sigChannel, stopChannel, &app)
	go acceptIncomingConnections(incoming, &app)

	<-stopChannel
	app.Log("received STOP signal; shutting down")
	app.Log("server shut down")
}

func acceptIncomingConnections(incoming net.Listener, app *Application) {
	for {
		app.Debug(DebugIO, "waiting for next incoming client")
		client, err := incoming.Accept()
		if err != nil {
			app.Logf("incoming connection: %v", err)
			continue
		}
		app.Debugf(DebugIO, "client connected from %v", client.RemoteAddr())

		auth, err := app.newClientAuthenticator("")
		if err != nil {
			app.Logf("unable to set up client authentication: %v", err)
			client.Close()
			continue
		}

		ourDebugFlags := DebugFlagNameSlice(app.DebugLevel)
		debugFlags, _ := mapper.NamedDebugFlags(ourDebugFlags...)

		newConnection, err := mapper.NewClientConnection(client,
			mapper.WithServer(app),
			mapper.WithClientDebuggingLevel(debugFlags),
			mapper.WithClientAuthenticator(auth),
		)
		if err != nil {
			app.Logf("unable to initialize client session: %v", err)
			client.Close()
			continue
		}
		go newConnection.ServeToClient(context.Background(), app.ServerStarted, app.LastPing)
	}
}

// @[00]@| Go-GMA 5.3.0
// @[01]@|
// @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
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
