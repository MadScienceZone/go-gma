/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __    ______      _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \  / ___  \    (  __   )       #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) \/   )  )   | (  )  |       #
# | |      | || || || (___) | Assistant | (____        | |     /  /    | | /   |       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | |    /  /     | (/ /) |       #
# | | \_  )| |   | || (   ) |                 ) )      | |   /  /      |   / | |       #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_ /  /     _ |  (__) |       #
# (_______)|/     \||/     \| Client    \______/ (_) \____/ \_/     (_)(_______)       #
#                                                                                      #
########################################################################################
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2024 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

/*
Session-stats collects statistics about the sessions of a campaign, which are stored
conveniently in a user-editable JSON file, and reformats them into HTML suitable
for posting to a campaign website.
The JSON file consists of an object with the following field:

	   game_sessions
	      This is a list of objects representing each game, each with the fields:
		     date     The game date in mm-ddd-yyyy format
			 video    The YouTube video link (just the token after "v=" in the URL)
			 duration Game session length, e.g. 4h30m5s
			 title    Title for the game session

More data values may be defined in the future as this program grows.

These are output as an HTML table with total game play time at the bottom.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"time"
	"github.com/MadScienceZone/go-gma/v5/text"
)

type SessionStats struct {
	GameSessions []GameSession `json:"game_sessions"`
}

type GameSession struct {
	Date       GameDate `json:"date"`
	VideoToken string   `json:"video,omitempty"`
	Duration   GameTime `json:"duration"`
	Title      string   `json:"title"`
	WorldDates string   `json:"world_dates,omitempty"`
	BookNumber int      `json:"book"`
	Synopsis   string   `json:"synopsis,omitempty"`
	ForumURL   string	`json:"url,omitempty"`
}

type GameDate struct {
	time.Time
}

type GameTime struct {
	time.Duration
}

func (d *GameDate) UnmarshalJSON(b []byte) error {
	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		return err
	}

	date, err := time.Parse("02-Jan-2006", dateStr)
	if err != nil {
		return err
	}
	d.Time = date
	return nil
}

func (d *GameTime) UnmarshalJSON(b []byte) error {
	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		return err
	}

	date, err := time.ParseDuration(dateStr)
	if err != nil {
		return err
	}
	d.Duration = date
	return nil
}

func main() {
	var generateSynopsis = flag.Bool("s", false, "generate synopsis of games")
	var generateVidList = flag.Bool("v", false, "generate list of video links for games")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println(len(flag.Args()), flag.Args())
		fmt.Printf("Usage: %s [-sv] json-file\n", os.Args[0])
		os.Exit(1)
	}

	gameData, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	var stats SessionStats
	if err := json.Unmarshal(gameData, &stats); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	if *generateVidList {
		generateGameSummary(stats)
	}
	if *generateSynopsis {
		generateGameSynopsis(stats)
	}
}

func generateGameSynopsis(stats SessionStats) {
	fmt.Println(`[html]
<link rel="stylesheet" href="/gpbp/local.css" />
<table class="pftable">
	<thead>
<th><b>Session</b></th>
<th><b>Game Date</b></th>
<th><b>Name</b></th>
<th><b>Campaign Dates</b></th>
<th><b>Synopsis</b></th></tr>
</thead>
<tbody>`)
	current_book := 0
	extra := ""
	for i, session := range stats.GameSessions {
		if current_book < session.BookNumber {
			current_book = session.BookNumber
			if current_book == 1 {
				extra = "<b>Start of Age of Worms Campaign.</b> "
			} else {
				bookRoman, err := text.ToRoman(current_book)
				if err == nil {
					extra = "<b>Start of Book " + bookRoman + "</b> "
				} else {
					extra = fmt.Sprintf("<b>Start of Book %d</b> ", current_book)
				}
			}
		} else {
			extra = ""
		}
		fmt.Printf(`<tr>
<td align=center valign=top><a href="%s">%d</a></td>
<td align=center valign=top><a href="%s">%s</a></td>
<td valign=top>%s</td>
<td valign=top>%s</td>
<td valign=top>%s%s</td></tr>`,
			session.ForumURL, i+1,
			session.ForumURL, session.Date.Format("02-Jan-2006"),
			html.EscapeString(session.Title),
			session.WorldDates,
			extra, session.Synopsis,
		)
	}

	fmt.Println(`	</tbody>
</table>
[/html]`)
}

func generateGameSummary(stats SessionStats) {
	fmt.Println(`[html]
<table class="pftable">
	<thead>
		<tr><th>Session</th><th>Date</th><th>Video Duration/Link/Title</th></tr>
	</thead>
	<tbody>`)
	var totalSessions int
	var totalDuration time.Duration
	for i, session := range stats.GameSessions {
		fmt.Printf("\t\t<tr><td align=right>%d</td><td align=center>%s</td><td><a target=\"_blank\" href=\"https://www.youtube.com/watch?v=%s&t=0s\">%d:%02d:%02d %s</a></td></tr>\n",
			i+1, session.Date.Format("02-Jan-2006"), session.VideoToken,
			int(session.Duration.Hours()),
			int(session.Duration.Minutes())%60,
			int(session.Duration.Seconds())%60,
			html.EscapeString(session.Title))
		totalSessions = i + 1
		totalDuration += session.Duration.Duration
	}
	fmt.Printf("\t\t<tr class=\"pftablesummary\"><td colspan=3><b>Totals:</b></td></tr>\n")
	fmt.Printf("\t\t<tr class=\"pftablesummary\"><td align=right><b>%d</b></td><td></td><td><b>%dd, %02d:%02d:%02d</b></td></tr>\n",
		totalSessions,
		int(totalDuration.Hours()/24),
		int(totalDuration.Hours())%24,
		int(totalDuration.Minutes())%60,
		int(totalDuration.Seconds())%60)
	fmt.Println(`	</tbody>
</table>
[/html]`)
}
