/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______     _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )   (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |   | (  )  |              #
# | |      | || || || (___) | Assistant | (____         /   )   | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )    /   _/     |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/(_)(_______)              #
#                                                                                      #
########################################################################################
*/
//
////////////////////////////////////////////////////////////////////////////////////////
//                                                                                    //
//                                  Authenticator                                     //
//                                                                                    //
////////////////////////////////////////////////////////////////////////////////////////

// Package auth calculates the challenge/response for a login attempt, and checks user input to
// determine if it was a valid response to the challenge.
//
// BACKGROUND AND SECURITY NOTES
//
// This authentication system is designed for validating connections between clients
// and servers in our game system and NOTHING ELSE. This is an extremely casual, low-risk
// situation where the main intent is to deflect nuisance connections. It would NOT BE
// SUITABLE for use in most other situations where there is more sensitive information
// to be protected.
//
// In this system, there are two passwords configured into the server. One is shared
// among all players, and is configured into their clients so they may automatically
// connect to the server when playing the game. The other is for the GM's use, and
// serves only to identify his or her clients to the system as being GM clients for
// the purposes of routing GM-only information to them.
//
// If desired, individual passwords may be given to some users which they can use
// to authenticate in order to avoid confusion between players in the game.
// This will reserve their name and prevent other players from impersonating them, since
// their personal password will be required.
//
// While this scheme protects passwords from observation and replay in transit (but not
// man-in-the-middle or other more sophisticated attacks unless further protections are
// placed on the connection itself), the passwords themselves are handled on both client
// and server in plaintext form.
//
// In case you missed it above, DO NOT USE this authenticator for ANYTHING that is worth
// protecting. We only use it to play a game together.
//
// CLIENT-SIDE OPERATION
//
// A client wishing to authenticate should create an Authenticator as
//    a := NewClientAuthenticator(myUserName, mySecretValue, myProgramName)
// where mySecretValue is the password they are using to authenticate (which will be
// the shared secret (password) used by all clients to
// authenticate to the server, the GM's password, or the personal password assigned to the user with the
// name given by myUserName, if one exists); myProgramName is a string describing what
// sort of client this is.
//
// The value in myUserName is arbitrary. If the password is the shared player password,
// then myUserName is taken as the name this player wants to be known by in the game system,
// and may be anything except "GM". If the password is the GM password, then "GM" will be
// used for the username regardless of the value passed as myUserName. If the user has a
// personal password defined for them on the server, then myUserName must match the name
// on file (since that's how the server will know which personal password to check).
//
// If myUserName is empty, an attempt will be made to determine the local user name
// of the process, or will be "anonymous" if that cannot be accomplished. If the
// program name is empty, "unnamed" will be used.
//
// The client then obtains a challenge from the server in the form of a base-64-encoded
// string, which is passed to the AcceptChallenge method:
//     response, err := a.AcceptChallenge(serverChallengeString)
// This provides the response to send back to the server in order to log in.
//
// SERVER-SIDE OPERATION
//
// A server which accepts a client connection should create an Authenticator
// for each such client which will track that session's authentication state.
//     a := &Authenticator{Secret: commonPassword, GmSecret: gmPassword}
// where commonPassword is the password used by all clients to authenticate, and
// gmPassword is the one used by privileged clients.
//
// The server then generates a unique challenge for this client's session by
// calling
//     challenge, err := a.GenerateChallenge()
//
// The generated challenge string is sent to the client, which will reply with
// a response string.
//
// If the server knows it needs to validate against a specific non-GM user's
// personal password, the server needs to update the Authenticator with that
// password by calling
//     a.SetSecret(personalSecret)
// before validating the client's response.
//
// The server validates the authenticity of the response
// by calling
//     ok, err := a.ValidateResponse(clientResponse)
// which will return true if the response is correct for that challenge.
//
// AUTHENTICATION ALGORITHM
//
// Given:
//   C:    the server's challenge as a byte slice
//   h(x): the binary SHA-256 hash digest of byte slice x
//   P:    the password needed to authenticate to the server
//   x‖y:  means to concatenate x and y
//
// To calculate the response to the server's authentication challenge,
// the client must do the following:
//
//   (1) Obtain i from the first 2 bytes of C as a big-endian uint16 value.
//   (2) Calculate D=h(C‖P).
//   (3) Repeat i times: D'=h(P‖D); let D=D'
//
// D is the response to send to the server for validation.
//
// PROTOCOL
//
// Although the auth package itself isn't involved in the client/server protocol
// directly, the way it is used by the map server and its clients uses the following
// protocol:
//
//  (server->client) OK {"Protocol":<v>, "Challenge":"<challenge>"}
// The server's greeting to the client includes this line which gives the server's
// protocol version (<v>) and a base-64 encoding of the binary challenge value (C in the
// algorithm described above)
//
//  (server<-client) AUTH {"Response":"<response>", "User":"<user>", "Client":"<client>"}
// The client's response is sent with this line, where <response> is the base-64
// encoded representation of the response to the challenge (D above), and the optional <user> and
// <client> values are the desired user name and description of the client program.
//
//  (server->client) DENIED {"Reason":"<message>"}
// Server response indicating that the authentication was unsuccessful.
//
//  (server->client) GRANTED {"User":"<user>"}
// Server response indicating that the authentication was successful. The <user> value is
// "GM" if the GM password was used, or the user name supplied by the user, which will be
// the name they're known by inside the game map (usually a character name). If the user
// did not supply a name and did not use the GM password, "anonymous" is returned.
//
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os/user"
)

//////////////////////////////////////////////////////////////////////////////////
//     _         _   _                _   _           _
//    / \  _   _| |_| |__   ___ _ __ | |_(_) ___ __ _| |_ ___  _ __
//   / _ \| | | | __| '_ \ / _ \ '_ \| __| |/ __/ _` | __/ _ \| '__|
//  / ___ \ |_| | |_| | | |  __/ | | | |_| | (_| (_| | || (_) | |
// /_/   \_\__,_|\__|_| |_|\___|_| |_|\__|_|\___\__,_|\__\___/|_|
//

// An Authenticator object holds the authentication challenge/response
// context for a single client transaction.
type Authenticator struct {
	// unprivileged (player) authentication password (client or server)
	Secret []byte

	// privileged (GM) authentication password (server only)
	GmSecret []byte

	// current generated challenge or nil
	Challenge []byte

	// text description of client program/version
	Client string

	// name of user authenticating
	Username string

	// true if GM authentication succeeded (server only)
	GmMode bool
}

//
// SetSecret changes the non-GM secret and disables GM logins for this
// authenticator. (SERVER)
//
// This is typically used on the server's side when it
// determines that the user trying to authenticate has
// their own password configured, so it calls SetSecret
// with that personal password. Subsequent calls to
// ValidateResponse will be based on this personal
// password.
//
// GM logins are disabled since the GM already has their
// own password, so this feature would not apply to them.
//
func (a *Authenticator) SetSecret(secret []byte) {
	a.GmSecret = []byte{}
	a.GmMode = false
	a.Secret = secret
}

//
// Compare two byte arrays for equality.
//
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

//
// GenerateChallenge creates an authentication challenge for this session.
// (SERVER)
//
// This is a 256-bit random nonce which is remembered for
// subsequent operations on this Authenticator instance.
// It does not depend on the password(s).
//
// Since the first two bytes of this challenge value specify the number
// of rounds to apply the hashing function on the client, the 16-bit
// number they represent is constrained to the range [64, 4095] to
// avoid too few rounds or so many rounds that a client may be unnecessarily
// burdened. (The latter is an actual runtime performance issue for one of
// our clients which is implemented as a Tcl script.)
//
// The challenge is returned as a base-64-encoded string.
//
func (a *Authenticator) GenerateChallenge() (string, error) {
	_, err := a.GenerateChallengeBytes()
	if err != nil {
		return "", err
	}
	return a.CurrentChallenge(), nil
}

//
// GenerateChallengeBytes is just like GenerateChallenge but
// returns the challenge as a binary byte slice instead of
// encoding it in base 64.
// (SERVER)
//
func (a *Authenticator) GenerateChallengeBytes() ([]byte, error) {
	a.Challenge = make([]byte, 32)
	a.GmMode = false
	_, err := rand.Read(a.Challenge)
	if err != nil {
		return nil, err
	}

	// The first two bytes of the nonce determine the number of
	// rounds of hashing that are done for the response. We'll
	// limit the number to something in the range [64,4095] to
	// prevent a trivial run or causing a performance issue on
	// the client by requiring too many.
	a.Challenge[0] &= 0x0f
	a.Challenge[1] |= 0x40

	return a.CurrentChallengeBytes(), nil
}

//
// CurrentChallenge returns the last-generated challenge created by
// GenerateChallenge.
// (SERVER)
//
// This is returned as a base-64-encoded string suitable for transmitting
// directly to a client.
//
func (a *Authenticator) CurrentChallenge() string {
	return base64.StdEncoding.EncodeToString(a.Challenge)
}

//
// CurrentChallengeBytes is like CurrentChallenge but returns the value
// as a byte slice instead of base 64 encoding it.
// (SERVER)
//
func (a *Authenticator) CurrentChallengeBytes() []byte {
	return a.Challenge
}

//
// Calculate the correct response for the previously generated challenge
// value (which is stored in the Authenticator instance on which this
// method is called).
//
// The response is:
//    R = H(S || H(S || ... H(C||S)))
//        \________________/
//                 P
//
// Where C = the generated 256-bit binary challenge value
//       S = the user's secret
//    H(x) = the binary SHA-256 hash of x
//       P = the randomly-chosen number of rounds through H(x)
//
// P is a random value from 0-65535 which is obtained by taking the
// first two bytes of C as a 16-bit big-endian binary integer value.
// (The current server will ask for something between 64 and 4095 rounds,
// but the code here will handle the entire unsigned 16-bit range.)
//
// **NOTE** This is not the most robust of challenge-response implementations,
// but deemed adequate for our purposes at the moment. Specifically, it is
// vulnerable to MitM attacks if the communications path is not otherwise
// protected, and long-term observation of authentication attempts may
// provide sufficient insight to determine the secret component of the
// exchange if the communication channel is not otherwise encrypted.
//
// I wouldn't normally say this, but in this very specific case, what we
// are protecting is so very, very unimportant to anyone, and the result
// of obtaining a credential is to either watch us play D&D or interfere
// with moving our pieces on the board, so we're ok with this simple
// approach for the time being. We'll most likely upgrade this to have
// additional protections in the future.
//
func (a *Authenticator) calcResponse(secret []byte) ([]byte, error) {
	var i uint16

	if len(a.Challenge) < 8 {
		return nil, fmt.Errorf("No (or insufficient) challenge value set; unable to determine response")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("No secret value set; unable to determine response")
	}

	passes := binary.BigEndian.Uint16(a.Challenge[0:2])
	h := sha256.New()
	h.Write(a.Challenge)
	h.Write(secret)
	d := h.Sum(nil)
	for i = 0; i < passes; i++ {
		h.Reset()
		h.Write(secret)
		h.Write(d)
		d = h.Sum(nil)
	}

	return d, nil
}

//
// AcceptChallenge takes a server's challenge, stores it internally, and generates an appropriate
// response to it, which is returned as a base-64 encoded string. (CLIENT)
//
func (a *Authenticator) AcceptChallenge(challenge string) (string, error) {
	c, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		return "", fmt.Errorf("bad challenge string: %v", err)
	}

	if len(c) < 2 {
		return "", fmt.Errorf("challenge value is too short")
	}
	response, err := a.AcceptChallengeBytes(c)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(response), nil
}

//
// AcceptChallengeBytes is like Accept Challenge but takes the raw binary
// challenge and emits the raw binay response as []byte slices.
//
func (a *Authenticator) AcceptChallengeBytes(challenge []byte) ([]byte, error) {
	a.Challenge = challenge
	response, err := a.calcResponse(a.Secret)
	if err != nil {
		return nil, fmt.Errorf("unable to generate response: %v", err)
	}
	return response, nil
}

//
// ValidateResponse takes
// a base-64-encoded response string and verifies that the value it encodes
// matches the expected response for the previously-generated challenge.
// (SERVER)
//
// If the response is not correct for the expected user's secret, we try
// again against the GM's secret to see if the user is logging in as the GM
// role. If so, the GmMode member of the Authenticator is set to true.
//
// Returns true if the authentication was successful.
//
func (a *Authenticator) ValidateResponse(response string) (bool, error) {
	binaryResponse, err := base64.StdEncoding.DecodeString(response)
	if err != nil {
		return false, fmt.Errorf("Error decoding client response: %v", err)
	}
	return a.ValidateResponseBytes(binaryResponse)
}

//
// ValidateResponseBytes is like ValidateResponse except that the response
// value passed is in the form of a byte slice instead of being base 64
// encoded.
// (SERVER)
//
func (a *Authenticator) ValidateResponseBytes(binaryResponse []byte) (bool, error) {
	a.GmMode = false
	if len(a.Secret) == 0 {
		return false, fmt.Errorf("No password configured")
	}
	ourResponse, err := a.calcResponse(a.Secret)
	if err != nil {
		return false, fmt.Errorf("Error validating client response: %v", err)
	}
	if bytesEqual(ourResponse, binaryResponse) {
		return true, nil
	}
	if len(a.GmSecret) > 0 {
		ourResponse, err := a.calcResponse(a.GmSecret)
		if err != nil {
			return false, fmt.Errorf("Error validating client response (gm): %v", err)
		}
		if bytesEqual(ourResponse, binaryResponse) {
			a.GmMode = true
			return true, nil
		}
	}
	return false, nil
}

//
// Reset returns an Authenticator instance back to its pre-challenge state
// so that it may be used again for another authentication attempt.
// (CLIENT/SERVER)
//
// Existing secrets and user names are preserved.
//
func (a *Authenticator) Reset() {
	a.Challenge = []byte{}
	a.GmMode = false
}

//
// NewClientAuthenticator creates and returns a pointer to a new Authenticator
// instance, prepared for use by a typical client. (CLIENT)
//
// If the username string is empty, an
// attempt will be made to determine the local username on the system, or the string
// "anonymous" will be used. If the client string is empty, "unnamed" will be used.
//
func NewClientAuthenticator(username string, secret []byte, client string) *Authenticator {
	a := &Authenticator{
		Username: username,
		Client:   client,
	}
	a.Secret = make([]byte, len(secret))
	copy(a.Secret, secret)

	if a.Client == "" {
		a.Client = "unnamed"
	}

	if a.Username == "" {
		u, err := user.Current()
		if err != nil {
			a.Username = "anonymous"
		} else {
			a.Username = u.Name
		}
	}
	return a
}

// @[00]@| GMA 5.2.0
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
