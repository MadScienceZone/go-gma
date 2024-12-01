/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  _______      __          #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )(  ____ \    /  \         #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  || (    \/    \/) )        #
# | |      | || || || (___) | Assistant | (____         /   )| (____        | |        #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   / (_____ \       | |        #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) )      | |        #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _  __) (_       #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_) \____/       #
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
	"regexp"
	"strings"

	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/mapper"
	"golang.org/x/exp/slices"
)

const (
	MsgTypeClearChat   = 0
	MsgTypeChatMessage = 1
	MsgTypeRollResult  = 2
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
		a.sqldb, err = sql.Open(DatabaseDriver, "file:"+a.DatabaseName)
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
			create table delegates (
				user        text    not null,
				delegate    text    not null,
					primary key (user, delegate)
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
				frames integer not null default 0,
				speed integer not null default 0,
				loops integer not null default 0,
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

func (a *Application) StoreImageData(imageName string, img mapper.ImageInstance, anim *mapper.ImageAnimation) error {
	if anim == nil {
		result, err := a.sqldb.Exec(`REPLACE INTO images (name, zoom, location, islocal) VALUES (?, ?, ?, ?);`, imageName, img.Zoom, img.File, img.IsLocalFile)
		if err != nil {
			return err
		}

		a.debugDbAffected(result, fmt.Sprintf("stored image record \"%s\"@%v local=%v, ID=%v", imageName, img.Zoom, img.IsLocalFile, img.File))
	} else {
		result, err := a.sqldb.Exec(`REPLACE INTO images (name, zoom, location, islocal, frames, speed, loops) VALUES (?, ?, ?, ?, ?, ?, ?);`,
			imageName, img.Zoom, img.File, img.IsLocalFile,
			anim.Frames, anim.FrameSpeed, anim.Loops,
		)
		if err != nil {
			return err
		}

		a.debugDbAffected(result, fmt.Sprintf("stored image record \"%s\"@%v local=%v, ID=%v, frames=%d, speed=%d mS/frame, loops=%d", imageName, img.Zoom, img.IsLocalFile, img.File, anim.Frames, anim.FrameSpeed, anim.Loops))
	}

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
	rows, err := a.sqldb.Query(`SELECT zoom, location, islocal, frames, speed, loops FROM images WHERE name=?`, img.Name)
	if err != nil {
		return resultSet, err
	}
	defer rows.Close()

	resultSet.Name = img.Name

	for rows.Next() {
		var instance mapper.ImageInstance
		var isLocal int
		var aframes int
		var aspeed int
		var aloops int

		if err := rows.Scan(&instance.Zoom, &instance.File, &isLocal, &aframes, &aspeed, &aloops); err != nil {
			return resultSet, err
		}
		if isLocal != 0 {
			instance.IsLocalFile = true
		}
		resultSet.Sizes = append(resultSet.Sizes, instance)
		if resultSet.Animation == nil && (aframes != 0 || aspeed != 0 || aloops != 0) {
			resultSet.Animation = &mapper.ImageAnimation{
				Frames:     aframes,
				FrameSpeed: aspeed,
				Loops:      aloops,
			}
		}
		a.Debugf(DebugDB, "result: \"%s\"@%v from \"%s\" (local=%v)", img.Name, instance.Zoom, instance.File, instance.IsLocalFile)
	}
	return resultSet, rows.Err()
}

func (a *Application) QueryPresetDelegates(user string) ([]string, error) {
	var delegates []string

	a.Debugf(DebugDB, "query of delegates for %s", user)
	rows, err := a.sqldb.Query(`SELECT delegate FROM delegates WHERE user=?`, user)
	if err != nil {
		return delegates, err
	}
	defer rows.Close()

	for rows.Next() {
		var d string

		if err := rows.Scan(&d); err != nil {
			return delegates, err
		}
		delegates = append(delegates, d)
		a.Debugf(DebugDB, "result: %s", d)
	}
	return delegates, rows.Err()
}

func (a *Application) QueryPresetDelegateFor(user string) ([]string, error) {
	var delegates []string

	a.Debugf(DebugDB, "query of who %s is a delegate for", user)
	rows, err := a.sqldb.Query(`SELECT user FROM delegates WHERE delegate=?`, user)
	if err != nil {
		return delegates, err
	}
	defer rows.Close()

	for rows.Next() {
		var d string

		if err := rows.Scan(&d); err != nil {
			return delegates, err
		}
		delegates = append(delegates, d)
		a.Debugf(DebugDB, "result: %s", d)
	}
	return delegates, rows.Err()
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
		rows, err = a.sqldb.Query(`SELECT msgid, msgtype, rawdata FROM chats WHERE msgid in (
				select msgid from chats order by msgid desc limit ?) order by msgid asc`, -target)
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
		case MsgTypeClearChat:
			var cc mapper.ClearChatMessagePayload
			if err := json.Unmarshal([]byte(jdata), &cc); err != nil {
				return err
			}
			requester.Conn.Send(mapper.ClearChat, cc)

		case MsgTypeChatMessage:
			var chat mapper.ChatMessageMessagePayload
			if err := json.Unmarshal([]byte(jdata), &chat); err != nil {
				return err
			}
			if chat.ToAll || (chat.ToGM && requester.Auth.GmMode) || slices.Contains(chat.Recipients, requester.Auth.Username) {
				requester.Conn.Send(mapper.ChatMessage, chat)
			}

		case MsgTypeRollResult:
			var rr mapper.RollResultMessagePayload
			if err := json.Unmarshal([]byte(jdata), &rr); err != nil {
				return err
			}
			if rr.ToAll || (rr.ToGM && requester.Auth.GmMode) || slices.Contains(rr.Recipients, requester.Auth.Username) {
				requester.Conn.Send(mapper.RollResult, rr)
			}

		default:
			a.Logf("Found item of type %v in chat history (ignored)", msgtype)
		}
	}
	return rows.Err()
}

func (a *Application) StoreDicePresetDelegates(user string, delegates []string) error {
	result, err := a.sqldb.Exec(`delete from delegates where user = ?`, user)
	if err != nil {
		return err
	}
	a.debugDbAffected(result, fmt.Sprintf("clear old delegates for %s", user))

	for i, delegate := range delegates {
		a.Debugf(DebugDB, "adding die-roll delegate %s for %s", delegate, user)
		result, err := a.sqldb.Exec("insert into delegates (user, delegate) values (?, ?)", user, delegate)
		if err != nil {
			return err
		}
		a.debugDbAffected(result, fmt.Sprintf("add delegate #%d (%s) for %s", i, delegate, user))
	}
	return nil
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
	var namesToDelete []string

	a.Debugf(DebugDB, "removing existing die-roll presets for %s matching /%s/", user, f.Filter)
	filter, err := regexp.Compile(f.Filter)
	if err != nil {
		return err
	}

	rows, err := a.sqldb.Query(`select name from dicepresets where user = ?`, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var thisName string
		if err := rows.Scan(&thisName); err != nil {
			return err
		}
		if filter.MatchString(thisName) {
			namesToDelete = append(namesToDelete, thisName)
		}
	}
	if len(namesToDelete) > 0 {
		a.Debugf(DebugDB, "--filter pattern matches %v row(s)", len(namesToDelete))

		for _, name := range namesToDelete {
			_, err := a.sqldb.Exec(`delete from dicepresets where user = ? and name = ?`, user, name)
			if err != nil {
				return err
			}
		}
	} else {
		a.Debugf(DebugDB, "--filter matched no presets")
	}
	return nil
}

func (a *Application) SendDicePresets(user string) error {
	delegates, err := a.QueryPresetDelegates(user)
	if err != nil {
		return err
	}
	delegateFor, err := a.QueryPresetDelegateFor(user)
	if err != nil {
		return err
	}

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
	pset.Delegates = delegates
	pset.DelegateFor = delegateFor
	pset.For = user

	for _, peer := range a.GetClients() {
		if peer.Auth != nil && (peer.Auth.Username == user || slices.Contains(delegates, peer.Auth.Username)) {
			peer.Conn.Send(mapper.UpdateDicePresets, pset)
		}
	}
	return nil
}

func (a *Application) AddToChatHistory(id int, chatType mapper.ServerMessage, chatData any) error {
	var dbMessageType int

	switch chatType {
	case mapper.ClearChat:
		dbMessageType = 0
	case mapper.ChatMessage:
		dbMessageType = 1
	case mapper.RollResult:
		dbMessageType = 2
	default:
		a.Logf("ERROR in AddToChatHistory: Invalid chatType value %v (entry not added to history database)", chatType)
		return fmt.Errorf("invalid chatType value %v", chatType)
	}

	jdata, err := json.Marshal(chatData)
	if err != nil {
		return err
	}
	result, err := a.sqldb.Exec(`insert into chats (msgid, msgtype, rawdata) values (?, ?, ?)`, id, dbMessageType, string(jdata))
	if err != nil {
		return err
	}
	a.debugDbAffected(result, "add to chat history")
	return nil
}

func (a *Application) LogDatabaseContents() error {
	a.Log("Database Contents:")

	dumpTable := func(title, table string, fields ...string) error {
		queryString := "select " + strings.Join(fields, ",") + " from " + table
		a.Logf("-%s (query=%s)", title, queryString)
		rows, err := a.sqldb.Query(queryString)
		if err != nil {
			return err
		}
		defer rows.Close()
		dest := make([]any, len(fields))
		values := make([]string, len(fields))
		for i := range fields {
			dest[i] = &values[i]
		}
		for rows.Next() {
			if err := rows.Scan(dest...); err != nil {
				return err
			}
			a.Logf("--%q", values)
		}
		return nil
	}

	if err := dumpTable("dice presets", "dicepresets", "user", "name", "description", "rollspec"); err != nil {
		return err
	}
	if err := dumpTable("chat history", "chats", "msgid", "msgtype", "rawdata"); err != nil {
		return err
	}
	if err := dumpTable("images known", "images", "name", "zoom", "location", "islocal"); err != nil {
		return err
	}
	return nil
}

//
// Remove all stored image definitions matching a regular expression
//
func (a *Application) FilterImages(f mapper.FilterImagesMessagePayload) error {
	var namesToDelete []string

	if f.KeepMatching {
		a.Debugf(DebugDB, "removing existing images NOT matching /%s/", f.Filter)
	} else {
		a.Debugf(DebugDB, "removing existing images matching /%s/", f.Filter)
	}

	filter, err := regexp.Compile(f.Filter)
	if err != nil {
		return err
	}

	rows, err := a.sqldb.Query(`select name from images`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var thisName string
		if err := rows.Scan(&thisName); err != nil {
			return err
		}
		matches := filter.MatchString(thisName)
		if (f.KeepMatching && !matches) || (!f.KeepMatching && matches) {
			namesToDelete = append(namesToDelete, thisName)
		}
	}
	if len(namesToDelete) > 0 {
		a.Debugf(DebugDB, "--filter pattern matches %v row(s) to be deleted", len(namesToDelete))

		for _, name := range namesToDelete {
			_, err := a.sqldb.Exec(`delete from images where name = ?`, name)
			if err != nil {
				return err
			}
		}
	} else {
		a.Debugf(DebugDB, "--filter matched no images")
	}
	return nil
}

// @[00]@| Go-GMA 5.25.1
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2024 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
// @[13]@| points along that historical time line.
// @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[15]@| License as described in the accompanying LICENSE file distributed
// @[16]@| with GMA.
// @[17]@|
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
