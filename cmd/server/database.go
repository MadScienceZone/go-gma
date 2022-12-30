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
// Database subsystem for the map server. This stores the persistent data the server
// needs to maintain between sessions. Note that the game state is now considered
// too ephemeral to pay the cost of constantly writing it to the database.
//

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/mapper"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
)

func (a *Application) dbOpen() error {
	var err error

	if a.DatabaseName == "" {
		a.sqldb = nil
		return nil
	}

	if _, err = os.Stat(a.DatabaseName); os.IsNotExist(err) {
		// database doesn't exist yet; create a new one

		a.Logf("no existing sqlite3 database \"%s\" found--creating a new one", a.DatabaseName)
		a.sqldb, err = sql.Open("sqlite3", "file:"+a.DatabaseName)
		if err != nil {
			a.Logf("unable to create sqlite3 database %s: %v", a.DatabaseName, err)
			return err
		}

		_, err = a.sqldb.Exec(`
			create table dicepresets (
				user        text    not null,
				name        text    not null,
				description text    not null,
				rollspec    text    not null,
					primary key (user, name)
			);
			create table chats (
				msgid   integer primary key,
				msgtype integer,
				rawdata text    not null
			);
			create table images (
				name	text	not null,
				zoom    real    not null,
				location text   not null,
				islocal integer(1) not null,
					primary key (name,zoom)
		);`)

		if err != nil {
			a.Logf("unable to create sqlite3 database %s contents: %v", a.DatabaseName, err)
			return err
		}
	} else {
		a.sqldb, err = sql.Open("sqlite3", "file:"+a.DatabaseName)
	}
	return err
}

func (a *Application) dbClose() error {
	if a.sqldb == nil {
		return nil
	}
	return a.sqldb.Close()
}

func (a *Application) StoreImageData(imageName string, img mapper.ImageInstance) error {
	result, err := a.sqldb.Exec(`REPLACE INTO images (name, zoom, location, islocal) VALUES (?, ?, ?, ?);`, imageName, img.Zoom, img.File, img.IsLocalFile)
	if err != nil {
		return err
	}
	a.debugDbAffected(result, fmt.Sprintf("stored image record \"%s\"@%v local=%v, ID=%v", imageName, img.Zoom, img.IsLocalFile, img.File))
	return nil
}

func (a *Application) debugDbAffected(result sql.Result, msg string) {
	affected, err := result.RowsAffected()
	if err != nil {
		a.Debugf(DebugDB, "%s, (unable to examine results: %v)", msg, err)
	} else {
		a.Debugf(DebugDB, "%s, rows affected=%d", msg, affected)
	}
}

func (a *Application) ClearChatHistory(target int) error {
	var result sql.Result
	var err error

	if target == 0 {
		// clear everything
		result, err = a.sqldb.Exec(`delete from chats`)
	} else if target < 0 {
		// clear all but most recent -target messages
		result, err = a.sqldb.Exec(`delete from chats where msgid not in (select msgid from chats order by msgid desc limit ?)`, -target)
	} else {
		// clear all messages earlier than target
		result, err = a.sqldb.Exec(`delete from chats where msgid < ?`, target)
	}
	if err != nil {
		return err
	}
	a.debugDbAffected(result, fmt.Sprintf("clear chat history target=%d", target))
	return nil
}

func (a *Application) QueryImageData(img mapper.ImageDefinition) (mapper.ImageDefinition, error) {
	var resultSet mapper.ImageDefinition

	a.Debugf(DebugDB, "query of image \"%s\"", img.Name)
	rows, err := a.sqldb.Query(`SELECT zoom, location, islocal FROM images WHERE name=?`, img.Name)
	if err != nil {
		return resultSet, err
	}
	defer rows.Close()

	resultSet.Name = img.Name
	for rows.Next() {
		var instance mapper.ImageInstance
		var isLocal int

		if err := rows.Scan(&instance.Zoom, &instance.File, &isLocal); err != nil {
			return resultSet, err
		}
		if isLocal != 0 {
			instance.IsLocalFile = true
		}
		resultSet.Sizes = append(resultSet.Sizes, instance)
		a.Debugf(DebugDB, "result: \"%s\"@%v from \"%s\" (local=%v)", img.Name, instance.Zoom, instance.File, instance.IsLocalFile)
	}
	return resultSet, rows.Err()
}

func (a *Application) QueryChatHistory(target int, requester *mapper.ClientConnection) error {
	var rows *sql.Rows
	var err error

	if requester == nil || requester.Auth == nil || requester.Auth.Username == "" {
		return fmt.Errorf("query of chat history denied to unauthenticated requester")
	}

	a.Debugf(DebugDB, "query of chat history target=%d", target)
	if target == 0 {
		rows, err = a.sqldb.Query(`SELECT msgid, msgtype, rawdata FROM chats`)
	} else if target < 0 {
		rows, err = a.sqldb.Query(`SELECT msgid, msgtype, rawdata FROM chats WHERE msgid not in (select msgid from chats order by msgid desc limit ?)`, -target)
	} else {
		rows, err = a.sqldb.Query(`SELECT msgid, msgtype, rawdata FROM chats WHERE msgid > ?`, target)
	}

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var msgid int
		var msgtype int
		var jdata string

		if err := rows.Scan(&msgid, &msgtype, &jdata); err != nil {
			return err
		}

		switch msgtype {
		case int(mapper.ClearChat):
			var cc mapper.ClearChatMessagePayload
			if err := json.Unmarshal([]byte(jdata), &cc); err != nil {
				return err
			}
			requester.Conn.Send(mapper.ClearChat, cc)

		case int(mapper.ChatMessage):
			var chat mapper.ChatMessageMessagePayload
			if err := json.Unmarshal([]byte(jdata), &chat); err != nil {
				return err
			}
			if chat.ToAll || (chat.ToGM && requester.Auth.GmMode) || slices.Contains[string](chat.Recipients, requester.Auth.Username) {
				requester.Conn.Send(mapper.ChatMessage, chat)
			}

		case int(mapper.RollResult):
			var rr mapper.RollResultMessagePayload
			if err := json.Unmarshal([]byte(jdata), &rr); err != nil {
				return err
			}
			if rr.ToAll || (rr.ToGM && requester.Auth.GmMode) || slices.Contains[string](rr.Recipients, requester.Auth.Username) {
				requester.Conn.Send(mapper.RollResult, rr)
			}

		default:
			a.Logf("Found item of type %v in chat history (ignored)", msgtype)
		}
	}
	return rows.Err()
}

func (a *Application) StoreDicePresets(user string, presets []dice.DieRollPreset, deleteOld bool) error {
	if deleteOld {
		a.Debugf(DebugDB, "removing existing die-roll presets for %s", user)
		result, err := a.sqldb.Exec(`delete from dicepresets where user = ?`, user)
		if err != nil {
			return err
		}
		a.debugDbAffected(result, fmt.Sprintf("clear old presets for %s", user))
	}

	for i, preset := range presets {
		a.Debugf(DebugDB, "adding new preset %s for %s", preset.Name, user)
		result, err := a.sqldb.Exec(`
			replace into dicepresets (user, name, description, rollspec) 
				values (?, ?, ?, ?)`,
			user, preset.Name, preset.Description, preset.DieRollSpec)
		if err != nil {
			return err
		}
		a.debugDbAffected(result, fmt.Sprintf("add preset #%d for %s", i, user))
	}
	return nil
}

func (a *Application) FilterDicePresets(user string, f mapper.FilterDicePresetsMessagePayload) error {
	a.Debugf(DebugDB, "removing existing die-roll presets for %s matching /%s/", user, f.Filter)
	result, err := a.sqldb.Exec(`delete from dicepresets where user = ? and name regexp ?`, user, f.Filter)
	if err != nil {
		return err
	}
	a.debugDbAffected(result, fmt.Sprintf("filter presets for %s", user))
	return nil
}

func (a *Application) SendDicePresets(user string) error {
	rows, err := a.sqldb.Query(`select name, description, rollspec from dicepresets where user = ?`, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	var pset mapper.UpdateDicePresetsMessagePayload

	for rows.Next() {
		var preset dice.DieRollPreset
		if err := rows.Scan(&preset.Name, &preset.Description, &preset.DieRollSpec); err != nil {
			return err
		}
		pset.Presets = append(pset.Presets, preset)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, peer := range a.GetClients() {
		if peer.Auth != nil && peer.Auth.Username == user {
			peer.Conn.Send(mapper.UpdateDicePresets, pset)
		}
	}
	return nil
}

func (a *Application) AddToChatHistory(id int, chatType mapper.ServerMessage, chatData any) error {
	jdata, err := json.Marshal(chatData)
	if err != nil {
		return err
	}
	result, err := a.sqldb.Exec(`insert into chats (msgid, msgtype, rawdata) values (?, ?, ?)`, id, int(chatType), string(jdata))
	if err != nil {
		return err
	}
	a.debugDbAffected(result, "add to chat history")
	return nil
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