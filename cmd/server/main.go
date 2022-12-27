/*
########################################################################################
#  _______  _______  _______             _______     _______     _______         _____ #
# (  ____ \(       )(  ___  )           (  ____ \   (  __   )   (  __   )       (  ___ #
# | (    \/| () () || (   ) |           | (    \/   | (  )  |   | (  )  |       | (    #
# | |      | || || || (___) |           | (____     | | /   |   | | /   | _____ | (___ #
# | | ____ | |(_)| ||  ___  |           (_____ \    | (/ /) |   | (/ /) |(_____)|  ___ #
# | | \_  )| |   | || (   ) | Game            ) )   |   / | |   |   / | |       | (    #
# | (___) || )   ( || )   ( | Master's  /\____) ) _ |  (__) | _ |  (__) |       | )    #
# (_______)|/     \||/     \| Assistant \______/ (_)(_______)(_)(_______)       |/     #
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

const GMAVersionNumber="5.0.0-alpha.1" // @@##@@

//
// eventMonitor responds to signals and timers that affect our overall operation
// independent of client requests.
//
/*
func eventMonitor(sig_chan chan os.Signal, stop_chan chan int, ms *mapservice.MapService, saveInterval int) {
*/
func eventMonitor(sigChan chan os.Signal, stopChan chan int, app *Application) {
	ping_signal := time.NewTicker(1 * time.Minute)

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

				//				ms.DumpState()

			case syscall.SIGUSR2:
				app.Debug(DebugEvents, "SIGUSR2 (currently not used by the server)")

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
			app.SendToAll(mapper.Marco, nil)
			//			any_connections := ms.PingAll()
			//			if any_connections {
			//				if report_interval > 1 {
			//					report_interval = 1
			//					log.Printf("Activity detected; reset ping timer to 1 minute")
			//					ping_signal.Reset(1 * time.Minute)
			//				}
			//			} else {
			//				new_interval := report_interval * 2
			//				if new_interval > 60 {
			//					new_interval = 60
			//				}
			//				if new_interval != report_interval {
			//					report_interval = new_interval
			//					log.Printf("No connections; backing off ping timer to %d minutes", new_interval)
			//					ping_signal.Reset(time.Duration(new_interval) * time.Minute)
			//				}
			//			}
		}
	}
}

func generateMessageIDs(c chan int) {
	// Start off with the time on the clock, on the assumption
	// that on average there won't be more than a chat message per
	// second since the server was started, so when the server is
	// restarted, this should give us a safe margin to start a new
	// set of IDs. It's simplistic, but works for our purposes.
	var nextMessageID int = int(time.Now().Unix())

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

	app := NewApplication()
	app.Logger.SetPrefix("go-gma-server: ")
	if err := app.GetAppOptions(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %v\n", err)
		os.Exit(1)
	}
	app.Logf("Server %s started", GMAVersionNumber)
	app.Logf("Implements protocol %d (minimum %d, maximum %d)",
		mapper.GMAMapperProtocol,
		mapper.MinimumSupportedMapProtocol,
		mapper.MaximumSupportedMapProtocol)

	go generateMessageIDs(app.MessageIDGenerator)
	go app.managePreambleData()
	go app.manageClientList()
	go app.manageGameState()

	/* instrumentation */
	// set the following environment variables for the New Relic
	// Go Agent:
	//    NEW_RELIC_APP_NAME = the name you want to appear in the datasets
	//    NEW_RELIC_LICENSE_KEY = your license key
	//
	if InstrumentCode {
		app.Log("application performance metrics telemetry reporting enabled")
		nrApp, err = newrelic.NewApplication(
			newrelic.ConfigAppName("go-gma-server"),
			newrelic.ConfigFromEnvironment(),
			newrelic.ConfigDebugLogger(os.Stdout),
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
		go newConnection.ServeToClient(context.Background())
	}
}

// @[00]@| GMA 5.0.0-alpha.1
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
