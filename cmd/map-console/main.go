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
#
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
# Copyright (c) 2021 by Steven L. Willoughby, Aloha, Oregon, USA.
# All Rights Reserved.
# Licensed under the terms and conditions of the BSD 3-Clause license.
#
# Based on earlier code by the same author, unreleased for the author's
# personal use; copyright (c) 1992-2019.
#
########################################################################
*/

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/auth"
	"github.com/MadScienceZone/go-gma/v5/dice"
	"github.com/MadScienceZone/go-gma/v5/gma"
	"github.com/MadScienceZone/go-gma/v5/mapper"
	"github.com/MadScienceZone/go-gma/v5/tcllist"
	"github.com/MadScienceZone/go-gma/v5/util"
)

const GMAVersionNumber="5.0.0" //@@##@@

func main() {
	fmt.Printf("GMA mapper console %s\n", GMAVersionNumber)
	log.SetPrefix("map-console: ")

	conf, err := configureApp()
	if err != nil {
		log.Fatalf("unable to set up: %v", err)
	}

	host, ok := conf.Get("host")
	if !ok {
		log.Fatalf("-host is required")
	}
	port, ok := conf.Get("port")
	if !ok {
		log.Fatalf("-port is required")
	}
	user, _ := conf.GetDefault("username", "")
	pass, _ := conf.GetDefault("password", "")

	problems := make(chan mapper.MessagePayload, 10)
	messages := make(chan mapper.MessagePayload, 10)
	done := make(chan int)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conOpts := []mapper.ConnectionOption{
		mapper.WithContext(ctx),
		mapper.WithSubscription(problems, mapper.ERROR, mapper.UNKNOWN),
		mapper.WithSubscription(messages,
			mapper.AddCharacter,
			mapper.AddImage,
			mapper.AddObjAttributes,
			mapper.AdjustView,
			mapper.ChatMessage,
			mapper.Clear,
			mapper.ClearChat,
			mapper.ClearFrom,
			mapper.CombatMode,
			mapper.Comment,
			mapper.LoadFrom,
			mapper.LoadArcObject,
			mapper.LoadCircleObject,
			mapper.LoadLineObject,
			mapper.LoadPolygonObject,
			mapper.LoadRectangleObject,
			mapper.LoadSpellAreaOfEffectObject,
			mapper.LoadTextObject,
			mapper.LoadTileObject,
			mapper.Marco,
			mapper.Mark,
			mapper.PlaceSomeone,
			mapper.QueryImage,
			mapper.RemoveObjAttributes,
			mapper.RollResult,
			mapper.Toolbar,
			mapper.UpdateClock,
			mapper.UpdateDicePresets,
			mapper.UpdateInitiative,
			mapper.UpdateObjAttributes,
			mapper.UpdatePeerList,
			mapper.UpdateProgress,
			mapper.UpdateStatusMarker,
			mapper.UpdateTurn,
		),
		mapper.WithDebugging(5),
	}

	if pass != "" {
		a := auth.NewClientAuthenticator(user, []byte(pass),
			fmt.Sprintf("map-console %s", GMAVersionNumber))
		conOpts = append(conOpts, mapper.WithAuthenticator(a))
	}
	server, conerr := mapper.NewConnection(host+":"+port, conOpts...)
	if conerr != nil {
		log.Fatalf("unable to contact mapper server: %v", conerr)
	}
	go func(done chan int) {
		server.Dial()
		done <- 1
	}(done)

	mono, err := conf.GetBool("mono")
	if err != nil {
		log.Printf("Error in -mono value: %v (assuming true)", err)
		mono = true
	}

	waitCounter := 0
	for !server.IsReady() {
		if waitCounter++; waitCounter > 10 {
			fmt.Println(colorize("Waiting for server to be ready...", "blue", mono))
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println(colorize("Connected to server.", "Green", mono))

	update, err := server.CheckVersionOf("core", GMAVersionNumber)
	if err != nil {
		log.Printf("Error checking for version updates: %v", err)
	} else if update != nil {
		log.Printf("UPDATE AVAILABLE! You are running version %v of GMA.", GMAVersionNumber)
		log.Printf("UPDATE AVAILABLE! Version %v is available for %v on %v.", update.Version, update.OS, update.Arch)
	}

	fmt.Printf("Server protocol %d; using %s calendar.\n", server.Protocol, server.CalendarSystem)
	fmt.Println("Characters Defined:")
	fmt.Println(colorize("NAME----------- ID-------- COLOR----- AREA SIZE", "Blue", mono))
	for _, def := range server.Characters {
		fmt.Println(colorize(fmt.Sprintf("%-15s %-10s %-10s %4s %4s", def.Name, def.ObjID(), def.Color, def.Area, def.Size), "Yellow", mono))
	}

	fmt.Println("Condition Codes from Server:")
	fmt.Println(colorize("CONDITION------ SHAPE COLOR----- DESCRIPTION-----------------------------------", "Blue", mono))
	for _, def := range server.Conditions {
		fmt.Println(colorize(fmt.Sprintf("%-15s %-5s %-10s %.46s", def.Condition, def.Shape, def.Color, def.Description), "Yellow", mono))
	}

	fmt.Println("Available Software Updates:")
	fmt.Println(colorize("PACKAGE--- OS-------- ARCH------ VERSION", "Blue", mono))
	for name, pkg := range server.PackageUpdatesAvailable {
		for _, vers := range pkg {
			fmt.Println(colorize(fmt.Sprintf("%-10s %-10s %-10s %s", name, vers.OS, vers.Arch, vers.Version), "Yellow", mono))
		}
	}

	go readUserInput(mono, cancel, server)

	if server.CalendarSystem == "" {
		// default to command-line argument if server didn't set the calendar
		server.CalendarSystem, _ = conf.GetDefault("calendar", "golarion")
	}
	cal, err := gma.NewCalendar(server.CalendarSystem)
	if err != nil {
		log.Fatalf("Error in calendar tracking: %v", err)
	}

eventloop:
	for {
		select {
		case msg := <-problems:
			switch message := msg.(type) {
			case mapper.ErrorMessagePayload:
				fmt.Println(colorize(fmt.Sprintf("ERROR: %v", message.Error), "Red", mono))
			case mapper.UnknownMessagePayload:
				fmt.Println(colorize(fmt.Sprintf("WARNING: Unknown type of message received from server: %q", msg.RawMessage()), "Red", mono))
			}
		case msg := <-messages:
			describeIncomingMessage(msg, mono, cal, server)
		case <-done:
			log.Printf("Server connection ended.")
			break eventloop
		}
	}
}

func descFields(mono bool, fields ...fieldDesc) string {
	var desc strings.Builder

	for i, f := range fields {
		if i > 0 {
			fmt.Fprint(&desc, colorize(", ", "Blue", mono))
		}

		if f.name == "" {
			fmt.Fprint(&desc, colorize(fmt.Sprintf("%v", f.value), "yellow", mono))
		} else if f.value == nil {
			fmt.Fprint(&desc, colorize(f.name, "Blue", mono))
		} else {
			fmt.Fprint(&desc, colorize(f.name+"=", "Blue", mono))
			switch v := f.value.(type) {
			case *mapper.CreatureHealth:
				fmt.Fprint(&desc, describeObject(mono, v))
			default:
				fmt.Fprint(&desc, colorize(fmt.Sprintf("%v", f.value), "yellow", mono))
			}
		}
	}
	return desc.String()
}

func describeBaseMapObject(mono bool, o mapper.MapElement) string {
	return descFields(mono,
		fieldDesc{"x", o.X},
		fieldDesc{"y", o.Y},
		fieldDesc{"z", o.Z},
		fieldDesc{"points", o.Points},
		fieldDesc{"line", o.Line},
		fieldDesc{"fill", o.Fill},
		fieldDesc{"width", o.Width},
		fieldDesc{"layer", o.Layer},
		fieldDesc{"level", o.Level},
		fieldDesc{"group", o.Group},
		fieldDesc{"dash", o.Dash},
		fieldDesc{"hidden", o.Hidden},
		fieldDesc{"locked", o.Locked},
	)
}

func describeObject(mono bool, obj any) string {
	var desc strings.Builder

	switch o := obj.(type) {
	case *mapper.CreatureHealth:
		if o == nil {
			fmt.Fprint(&desc, colorize("health{nil", "magenta", mono))
		} else {
			fmt.Fprint(&desc, colorize("health{", "magenta", mono))
			fmt.Fprint(&desc, descFields(mono,
				fieldDesc{"maxHP", (*o).MaxHP},
				fieldDesc{"lethal", (*o).LethalDamage},
				fieldDesc{"non", (*o).NonLethalDamage},
				fieldDesc{"con", (*o).Con},
				fieldDesc{"flat", (*o).IsFlatFooted},
				fieldDesc{"stable", (*o).IsStable},
				fieldDesc{"condition", (*o).Condition},
				fieldDesc{"blur", (*o).HPBlur},
			))
		}

	case *mapper.RadiusAoE:
		if o == nil {
			fmt.Fprint(&desc, colorize("radiusAoE{nil", "magenta", mono))
		} else {
			fmt.Fprint(&desc, colorize("radiusAoE{", "magenta", mono))
			fmt.Fprint(&desc, descFields(mono,
				fieldDesc{"r", (*o).Radius},
				fieldDesc{"color", (*o).Color},
			))
		}

	case map[string]any:
		var m []fieldDesc

		for k, v := range o {
			switch vv := v.(type) {
			case *mapper.CreatureHealth:
				m = append(m, fieldDesc{k, describeObject(mono, vv)})
			default:
				m = append(m, fieldDesc{k, fmt.Sprintf("%v", v)})
			}
		}
		fmt.Fprint(&desc, colorize("{", "magenta", mono))
		fmt.Fprint(&desc, descFields(mono, m...))

	case mapper.Coordinates:
		fmt.Fprint(&desc, colorize("Coordinates{", "magenta", mono))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"x", o.X},
			fieldDesc{"y", o.Y},
		))

	case mapper.ArcElement:
		fmt.Fprint(&desc, colorize("arc{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"mode", o.ArcMode},
			fieldDesc{"start", o.Start},
			fieldDesc{"extent", o.Extent},
		))

	case mapper.CircleElement:
		fmt.Fprint(&desc, colorize("circle{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))

	case mapper.LineElement:
		fmt.Fprint(&desc, colorize("line{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"arrow", o.Arrow},
		))

	case mapper.PolygonElement:
		fmt.Fprint(&desc, colorize("poly{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"spline", o.Spline},
			fieldDesc{"join", o.Join},
		))

	case mapper.RectangleElement:
		fmt.Fprint(&desc, colorize("rect{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))

	case mapper.SpellAreaOfEffectElement:
		fmt.Fprint(&desc, colorize("aoe{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"shape", o.AoEShape},
		))

	case mapper.TextElement:
		fmt.Fprint(&desc, colorize("text{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"text", o.Text},
			fieldDesc{"family", o.Font.Family},
			fieldDesc{"size", o.Font.Size},
			fieldDesc{"weight", o.Font.Weight},
			fieldDesc{"slant", o.Font.Slant},
			fieldDesc{"anchor", o.Anchor},
		))

	case mapper.TileElement:
		fmt.Fprint(&desc, colorize("tile{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o.MapElement))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"image", o.Image},
			fieldDesc{"BBHeight", o.BBHeight},
			fieldDesc{"BBWidth", o.BBWidth},
		))

	case mapper.CreatureToken:
		fmt.Fprint(&desc, colorize("creature{", "magenta", mono))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"ID", o.ID},
			fieldDesc{"name", o.Name},
			fieldDesc{"health", describeObject(mono, o.Health)},
			fieldDesc{"gx", o.Gx},
			fieldDesc{"gy", o.Gy},
			fieldDesc{"skin", o.Skin},
			fieldDesc{"skinsize", o.SkinSize},
			fieldDesc{"elev", o.Elev},
			fieldDesc{"color", o.Color},
			fieldDesc{"note", o.Note},
			fieldDesc{"size", o.Size},
			fieldDesc{"area", o.Area},
			fieldDesc{"statuslist", o.StatusList},
			fieldDesc{"aoe", describeObject(mono, o.AoE)},
			fieldDesc{"movemode", o.MoveMode},
			fieldDesc{"reach", o.Reach},
			fieldDesc{"killed", o.Killed},
			fieldDesc{"dim", o.Dim},
			fieldDesc{"type", o.CreatureType},
		))

	case mapper.MapElement:
		fmt.Fprint(&desc, colorize("MapElement{", "magenta", mono))
		fmt.Fprint(&desc, describeBaseMapObject(mono, o))

	case mapper.MapObject:
		fmt.Fprint(&desc, colorize("MapObject{", "magenta", mono))
		fmt.Fprint(&desc, descFields(mono,
			fieldDesc{"ID", o.ObjID()},
			fieldDesc{"data", "(...)"},
		))

	default:
		if obj == nil {
			fmt.Fprint(&desc, colorize("nil", "magenta", mono))
		} else {
			fmt.Fprint(&desc, colorize("{...", "magenta", mono))
		}
	}
	fmt.Fprint(&desc, colorize("}", "magenta", mono))
	return desc.String()
}

//
// describeIncomingMessage applys a standard formatting to present
// a human-readable representation of each server message.
//
type fieldDesc struct {
	name  string
	value any
}

func printFields(mono bool, cmd string, fields ...fieldDesc) {
	if cmd != "" {
		fmt.Print(colorize(cmd+" ", "Cyan", mono))
	}
	fmt.Println(descFields(mono, fields...))
}

func describeIncomingMessage(msg mapper.MessagePayload, mono bool, cal gma.Calendar, server mapper.Connection) {
	switch m := msg.(type) {
	case mapper.AddCharacterMessagePayload:
		printFields(mono, "AddCharacter",
			fieldDesc{"name", m.Name},
			fieldDesc{"id", m.ObjID()},
			fieldDesc{"color", m.Color},
			fieldDesc{"area", m.Area},
			fieldDesc{"size", m.Size},
		)
	case mapper.AddImageMessagePayload:
		for i, inst := range m.Sizes {
			if inst.ImageData != nil {
				printFields(mono, fmt.Sprintf("AddImage %d of %d", i+1, len(m.Sizes)),
					fieldDesc{"name", m.Name},
					fieldDesc{"zoom", inst.Zoom},
					fieldDesc{"data", fmt.Sprintf("(%d bytes)", len(inst.ImageData))},
				)
			} else {
				printFields(mono, fmt.Sprintf("AddImage %d of %d", i+1, len(m.Sizes)),
					fieldDesc{"name", m.Name},
					fieldDesc{"zoom", inst.Zoom},
					fieldDesc{"file", inst.File},
					fieldDesc{"local", inst.IsLocalFile},
				)
			}
		}

	case mapper.AddObjAttributesMessagePayload:
		printFields(mono, "AddObjAttributes",
			fieldDesc{"objID", m.ObjID},
			fieldDesc{"attrname", m.AttrName},
			fieldDesc{"values", m.Values},
		)

	case mapper.AdjustViewMessagePayload:
		printFields(mono, "AdjustView",
			fieldDesc{"xview", fmt.Sprintf("%.2f%%", m.XView*100)},
			fieldDesc{"yview", fmt.Sprintf("%.2f%%", m.YView*100)},
		)

	case mapper.ChatMessageMessagePayload:
		printFields(mono, "ChatMessage",
			fieldDesc{"messageID", m.MessageID},
			fieldDesc{"from", m.Sender},
			fieldDesc{"to", m.Recipients},
			fieldDesc{"toAll", m.ToAll},
			fieldDesc{"toGM", m.ToGM},
			fieldDesc{"text", m.Text},
		)

	case mapper.ClearMessagePayload:
		printFields(mono, "Clear",
			fieldDesc{"objID", m.ObjID},
		)

	case mapper.ClearChatMessagePayload:
		printFields(mono, "ClearChat",
			fieldDesc{"requestedBy", m.RequestedBy},
			fieldDesc{"silent", m.DoSilently},
			fieldDesc{"target", m.Target},
			fieldDesc{"messageID", m.MessageID},
		)

	case mapper.ClearFromMessagePayload:
		printFields(mono, "ClearFrom",
			fieldDesc{"file", m.File},
			fieldDesc{"local", m.IsLocalFile},
		)

	case mapper.CombatModeMessagePayload:
		printFields(mono, "CombatMode",
			fieldDesc{"enabled", m.Enabled},
		)
	case mapper.CommentMessagePayload:
		fmt.Println(
			colorize("//", "Cyan", mono),
			colorize(m.Text, "blue", mono),
		)

	case mapper.LoadFromMessagePayload:
		printFields(mono, "LoadFrom",
			fieldDesc{"file", m.File},
			fieldDesc{"local", m.IsLocalFile},
			fieldDesc{"cache", m.CacheOnly},
			fieldDesc{"merge", m.Merge},
		)

		//	case mapper.LoadObjectMessagePayload:
		//		printFields(mono, "LoadObject",
		//			fieldDesc{"objID", m.ObjID()},
		//			fieldDesc{"obj", describeObject(mono, m)},
		//		)
	case mapper.LoadArcObjectMessagePayload:
		printFields(mono, "LoadArcObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadCircleObjectMessagePayload:
		printFields(mono, "LoadCircleObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadLineObjectMessagePayload:
		printFields(mono, "LoadLineObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadPolygonObjectMessagePayload:
		printFields(mono, "LoadPolygonObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadRectangleObjectMessagePayload:
		printFields(mono, "LoadRectangleObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadSpellAreaOfEffectObjectMessagePayload:
		printFields(mono, "LoadSpellAreaOfEffectObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadTextObjectMessagePayload:
		printFields(mono, "LoadTextObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)
	case mapper.LoadTileObjectMessagePayload:
		printFields(mono, "LoadTileObject",
			fieldDesc{"objID", m.ObjID()},
			fieldDesc{"obj", describeObject(mono, m)},
		)

	case mapper.MarcoMessagePayload:
		fmt.Print(colorize(".", "blue", mono))
		server.Polo()

	case mapper.MarkMessagePayload:
		printFields(mono, "Mark",
			fieldDesc{"X", m.X},
			fieldDesc{"Y", m.Y},
		)

	case mapper.PlaceSomeoneMessagePayload:
		printFields(mono, "PlaceSomeone",
			fieldDesc{"obj", describeObject(mono, m.CreatureToken)},
		)

	case mapper.QueryImageMessagePayload:
		for i, inst := range m.Sizes {
			printFields(mono, fmt.Sprintf("QueryImage %d of %d", i+1, len(m.Sizes)),
				fieldDesc{"name", m.Name},
				fieldDesc{"zoom", inst.Zoom},
				fieldDesc{"file", inst.File},
				fieldDesc{"local", inst.IsLocalFile},
			)
		}

	case mapper.RemoveObjAttributesMessagePayload:
		printFields(mono, "RemoveObjAttributes",
			fieldDesc{"objID", m.ObjID},
			fieldDesc{"attrname", m.AttrName},
			fieldDesc{"values", m.Values},
		)

	case mapper.RollResultMessagePayload:
		printFields(mono, "RollResult",
			fieldDesc{"from", m.Sender},
			fieldDesc{"to", m.Recipients},
			fieldDesc{"messageID", m.MessageID},
			fieldDesc{"toAll", m.ToAll},
			fieldDesc{"toGM", m.ToGM},
			fieldDesc{"title", m.Title},
			fieldDesc{"result", m.Result},
		)

	case mapper.ToolbarMessagePayload:
		printFields(mono, "Toolbar",
			fieldDesc{"enabled", m.Enabled},
		)

	case mapper.UpdateClockMessagePayload:
		cal.SetTimeValue(int64(m.Absolute))
		printFields(mono, "UpdateClock",
			fieldDesc{"absolute", cal.ToString(2)},
			fieldDesc{"relative", cal.DeltaString(int64(m.Relative), false)},
		)

	case mapper.UpdateDicePresetsMessagePayload:
		printFields(mono, "UpdateDicePresets")
		for i, dp := range m.Presets {
			printFields(mono, colorize(fmt.Sprintf("  [%02d] ", i), "Blue", mono),
				fieldDesc{"name", dp.Name},
				fieldDesc{"desc", dp.Description},
				fieldDesc{"spec", dp.DieRollSpec},
			)
		}

	case mapper.UpdateInitiativeMessagePayload:
		printFields(mono, "UpdateInitiative")
		printFields(mono, "",
			fieldDesc{"       NAME----------- HLD RDY FLT -HP SL", nil})
		for i, slot := range m.InitiativeList {
			printFields(mono, "",
				fieldDesc{fmt.Sprintf("  [%02d]", i), fmt.Sprintf("%-15s %s %s %s %3d %2d",
					slot.Name,
					func(b bool) string {
						if b {
							return colorize(" Y ", "green", mono)
						}
						return colorize(" N ", "Red", mono)
					}(slot.IsHolding),
					func(b bool) string {
						if b {
							return colorize(" Y ", "green", mono)
						}
						return colorize(" N ", "Red", mono)
					}(slot.HasReadiedAction),
					func(b bool) string {
						if b {
							return colorize(" Y ", "green", mono)
						}
						return colorize(" N ", "Red", mono)
					}(slot.IsFlatFooted),
					slot.CurrentHP,
					slot.Slot)})
		}

	case mapper.UpdateObjAttributesMessagePayload:
		printFields(mono, "UpdateObjAttributes",
			fieldDesc{"objID", m.ObjID},
			fieldDesc{"attrs", describeObject(mono, m.NewAttrs)},
		)

	case mapper.UpdatePeerListMessagePayload:
		printFields(mono, "UpdatePeerList")
		printFields(mono, "",
			fieldDesc{"       USERNAME------------ ADDRESS-------------- CLIENT------------------- AU ME PING--", nil})
		for i, peer := range m.PeerList {
			printFields(mono, "",
				fieldDesc{fmt.Sprintf("  [%02d]", i), fmt.Sprintf("%-20s %-21s %-25s %s %s %s",
					peer.User, peer.Addr, peer.Client,
					func(b bool) string {
						if b {
							return colorize("Y ", "green", mono)
						}
						return colorize("N ", "Red", mono)
					}(peer.IsAuthenticated),
					func(b bool) string {
						if b {
							return colorize("Y ", "green", mono)
						}
						return colorize("N ", "Red", mono)
					}(peer.IsMe),
					func(p float64) string {
						if p <= 0 {
							return colorize("------", "yellow", mono)
						}
						if p <= 60 {
							return colorize("Active", "green", mono)
						}
						return colorize(fmt.Sprintf("%5.1fs", p), "Red", mono)
					}(peer.LastPolo))})
		}

	case mapper.UpdateProgressMessagePayload:
		printFields(mono, "UpdateProgress",
			fieldDesc{"ID", m.OperationID},
			fieldDesc{"title", m.Title},
			fieldDesc{"value", m.Value},
			fieldDesc{"max", m.MaxValue},
			fieldDesc{"done", m.IsDone},
		)

	case mapper.UpdateStatusMarkerMessagePayload:
		printFields(mono, "UpdateStatusMarker",
			fieldDesc{"condition", m.Condition},
			fieldDesc{"shape", m.Shape},
			fieldDesc{"color", m.Color},
		)

	case mapper.UpdateTurnMessagePayload:
		printFields(mono, "UpdateTurn",
			fieldDesc{"actor", m.ActorID},
			fieldDesc{"elapsed", fmt.Sprintf("%d:%02d:%02d", m.Hours, m.Minutes, m.Seconds)},
			fieldDesc{"rounds", m.Rounds},
			fieldDesc{"slot", m.Count},
		)

	default:
		fmt.Println(colorize(fmt.Sprintf("ERROR: Unhandled server message type: %d %q",
			msg.MessageType(), msg.RawMessage()), "Red", mono))
	}
}

func configureApp() (util.SimpleConfigurationData, error) {
	var defUserName string

	defUser, err := user.Current()
	if err != nil {
		defUserName = "unknown"
	} else {
		defUserName = defUser.Username
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to determine user's home directory: %v", err)
	}

	//
	// command-line parameters
	//
	defConfigPath := filepath.Join(homeDir, ".gma", "mapper", "mapper.conf")
	var Fhost = flag.String("host", "", "hostname of mapper service")
	var Fport = flag.Uint("port", 0, "TCP port of mapper service (default 2323)")
	var Fpass = flag.String("password", "", "Server password (if required)")
	var Frawd = flag.Bool("raw", false, "Report raw data only")
	var Fuser = flag.String("username", "", "Username on server or \"GM\" (default \""+defUserName+"\")")
	var Fconf = flag.String("config", defConfigPath, "Configuration file")
	var Fmono = flag.Bool("mono", false, "Suppress the output of ANSI color codes")
	var Fverb = flag.Bool("verbose", false, "Print extra output about connection")
	var Fcals = flag.String("calendar", "golarion", "Calendar system in use")
	flag.Parse()

	//
	// read in configuration
	//
	var conf util.SimpleConfigurationData
	if *Fconf != "" {
		configFile, err := os.Open(*Fconf)
		if err != nil {
			if *Fconf == defConfigPath && errors.Is(err, fs.ErrNotExist) {
				log.Printf("warning: default configuration file \"%s\" does not exist", *Fconf)
				conf = util.NewSimpleConfigurationData()
			} else {
				return nil, fmt.Errorf("%s: %v", *Fconf, err)
			}
		} else {
			defer configFile.Close()
			conf, err = util.ParseSimpleConfig(configFile)
			if err != nil {
				return nil, err
			}
		}
	} else {
		conf = util.NewSimpleConfigurationData()
	}

	// Override configuration file settings from command-line
	// options
	if *Fhost != "" {
		conf.Set("host", *Fhost)
	}
	if *Fport != 0 {
		conf.SetInt("port", int(*Fport))
	}
	if *Fpass != "" {
		conf.Set("password", *Fpass)
	}
	if *Fuser != "" {
		conf.Set("username", *Fuser)
	}
	if *Frawd {
		conf.Set("raw", "1")
	}
	if *Fverb {
		conf.Set("verbose", "1")
	}
	if *Fmono {
		conf.Set("mono", "1")
	}
	if *Fcals != "" {
		conf.Set("calendar", *Fcals)
	}

	// Sanity check and defaults
	u, ok := conf.Get("username")
	if !ok || u == "" {
		conf.Set("username", defUserName)
	}

	p, err := conf.GetIntDefault("port", 0)
	if err != nil {
		return nil, fmt.Errorf("port value: %v", err)
	}
	if p <= 0 {
		conf.SetInt("port", 2323)
	}

	u, ok = conf.Get("host")
	if !ok {
		return nil, fmt.Errorf("host value is required")
	}

	return conf, nil
}

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

//
// readUserInput takes lines of input from the user which correspond to
// map server requests. While the Python version of the map-console
// just sent the input as-is, trusting the user to form a protocol-conforming
// string, we will parse out the user's request, make sure it is correct,
// and then make that request to the server.
//
// It would be nice to accept a higher-level abstraction of the commands
// but for now, like the earlier Python map-console, we expect the user
// to type the command in the form it would be sent to the server.
// See the mapper protocol as described in mapper(6) for details.
//
func readUserInput(mono bool, cancel context.CancelFunc, server mapper.Connection) {
	fmt.Printf("map-console> ")
	scanner := bufio.NewScanner(os.Stdin)

inputloop:
	for scanner.Scan() {
		fields, err := tcllist.ParseTclList(scanner.Text())
		if err != nil {
			fmt.Println(colorize(fmt.Sprintf("ERROR: Unrecognized input: %v", err), "Red", mono))
		} else if len(fields) > 0 {
		handle_input:
			switch strings.ToUpper(fields[0]) {
			case "HELP", "?":
				fmt.Println(`Command summary:
AI <name> <size> <filename>             Upload image from local file
AI? <name> <size>                       Ask for definition of image
AI@ <name> <size> <serverid>            Advertise image stored on server
AV <xfrac> <yfrac>                      Adjust view to fraction of each axis
CC <silent?> <target>                   Clear chat history
CLR <id>|*|E*|M*|P*|[<image>=]<name>    Clear specified element(s) from canvas
CLR@ <serverid>                         Remove all elements from server file
CO <bool>                               Enable/disable combat mode
CS <abs> <rel>                          Set game clock
D {<recip>|@|*|% ...} <dice>            Roll dice to users (*=all, %=GM)
DD {{<name> <desc> <dice>} ...}         Replace dice preset list
DD+ {{<name> <desc> <dice>} ...}        Add to dice preset list
DD/ <regex>                             Delete all presets whose names match RE
DR                                      Request die roll preset
L <filename>                            Tell clients to load local file to mapper
L@ <filename>                           Tell clients to load server file
LS <filename>                           Upload contents of local file to all
M <filename>                            Tell clients to merge local file to mapper
M? <serverid>                           Ensure local cache of server file
M@ <serverid>                           Tell clients to merge contents of server file to canvas
MARK <x> <y>                            Show visual marker at coordinates
OA <id> {<k0> <v0> ... <kN> <vN>}       Set object attribute(s)
OA+ <id> <key> {<v0> <v1> ... <vN>}     Add to list-type object attribute
OA- <id> <key> {<v0> <v1> ... <vN>}     Remove from list-type object attribute
POLO                                    Answer server ping request
PS <id> <color> <name> <area> <size> player|monster <x> <y> <reach>  
QUIT|EXIT                               Exit the client
SYNC [CHAT [<target>]]                  Sync server content / chat history
TO {<recip>|@|*|% ...} <message>        Send chat message
/CONN`)
			case "//":
				// ignore
			case "AC", "ACCEPT", "DENIED", "DSM", "GRANTED", "I", "IL", "MARCO", "OK", "PRIV", "ROLL", "CONN", "CONN:", "CONN.":
				// server messages
				fmt.Println(colorize(fmt.Sprintf("ERROR: %s is not for clients to send.", fields[0]), "Red", mono))

			case "AI":
				// AI name size file
				v, err := tcllist.ConvertTypes(fields, "ssfs")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				data, err := os.ReadFile(v[3].(string))
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("I/O ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AddImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64), ImageData: data},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AI:", "AI.", "AUTH", "DD=", "DD:", "DD.", "LS:", "LS.":
				// AI: AI. DD= DD: DD. LS: LS. internal protocol commands
				// AUTH response [user client]
				fmt.Println(colorize(fmt.Sprintf("ERROR: %s should not be typed directly here.", fields[0]), "Red", mono))

			case "AI?":
				// AI? name size
				v, err := tcllist.ConvertTypes(fields, "ssf")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.QueryImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64)},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AI@":
				// AI@ name size id
				v, err := tcllist.ConvertTypes(fields, "ssfs")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AddImage(mapper.ImageDefinition{
					Name: v[1].(string),
					Sizes: []mapper.ImageInstance{
						{Zoom: v[2].(float64), File: v[3].(string), IsLocalFile: false},
					},
				}); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "AV":
				// AV x y
				v, err := tcllist.ConvertTypes(fields, "sff")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.AdjustView(v[1].(float64), v[2].(float64)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CC":
				// CC silent? target
				v, err := tcllist.ConvertTypes(fields, "s?i")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.ClearChat(v[2].(int), v[1].(bool)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CLR":
				// CLR id|*|E*|M*|P*|[imagename=]name
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: CLR <id>", "Red", mono))
					break
				}
				if err := server.Clear(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CLR@":
				// CLR@ id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: CLR@ <serverid>", "Red", mono))
					break
				}
				if err := server.ClearFrom(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CO":
				// CO state
				v, err := tcllist.ConvertTypes(fields, "s?")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.CombatMode(v[1].(bool)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "CS":
				// CS abs rel
				fmt.Println(colorize("Sorry, CS is not yet implemented for the console.", "Red", mono))

			case "D":
				// D reciplist dice
				if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: D <recip>|*|% <dice>", "Red", mono))
					break
				}
				recips, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in recipient list: %v", err), "Red", mono))
					break
				}
				if err := server.RollDice(recips, fields[2]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "DD", "DD+":
				// DD {{name desc dice} ...}
				// DD+ {{name desc dice} ...}
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DD[+] {{<name> <desc> <dice>} ...}", "Red", mono))
					break
				}
				p, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in preset list: %v", err), "Red", mono))
					break
				}
				var presetList []dice.DieRollPreset
				for i, ps := range p {
					pl, err := tcllist.Parse(ps, "sss")
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("ERROR in preset list, #%d: %v", i+1, err), "Red", mono))
						break handle_input
					}
					presetList = append(presetList, dice.DieRollPreset{
						Name:        pl[0].(string),
						Description: pl[1].(string),
						DieRollSpec: pl[2].(string),
					})
				}
				if fields[0] == "DD" {
					if err := server.DefineDicePresets(presetList); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.AddDicePresets(presetList); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "DD/":
				// DD/ regex
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DD/ <regex>", "Red", mono))
					break
				}
				if err := server.FilterDicePresets(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "DR":
				// DR
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: DR", "Red", mono))
					break
				}
				if err := server.QueryDicePresets(); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "L":
				// L filename
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: L <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], true, false); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "L@":
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: L@ <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], false, false); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "LS":
				// LS
				fmt.Println(colorize("LS not supported", "Red", mono))
				/*
					if len(fields) != 2 {
						fmt.Println(colorize("usage ERROR: wrong number of fields: LS <file>", "Red", mono))
						break
					}
					data, err := readLines(fields[1])
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("I/O ERROR: %v", err), "Red", mono))
						break
					}
					log.Printf("Reading objects from local mapper file %s", fields[1])
					objects, images, files, err := mapper.ParseObjects(data)
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("parser ERROR: %v", err), "Red", mono))
						break
					}
					log.Printf("Found %d object%s, %d image%s, and %d file%s in %s",
						len(objects), plural(len(objects)),
						len(images), plural(len(images)),
						len(files), plural(len(files)),
						fields[1],
					)
					if len(objects) > 0 {
						fmt.Print("Sending objects...")
						for i, o := range objects {
							if i%10 == 9 {
								fmt.Print(".")
							}
							if err := server.LoadObject(o); err != nil {
								fmt.Println(colorize(fmt.Sprintf("server ERROR sending object #%d (%s): %v",
									i+1, o.ObjID(), err), "Red", mono))
								break handle_input
							}
						}
						fmt.Println("done")
					}
					if len(images) > 0 {
						fmt.Print("Sending images...")
						for id, image := range images {
							if err := server.AddImage(image); err != nil {
								fmt.Println(colorize(fmt.Sprintf("server ERROR sending image %s: %v",
									id, err), "Red", mono))
								break handle_input
							}
						}
						fmt.Println("done")
					}
					if len(files) > 0 {
						fmt.Print("Sending files...")
						for i, f := range files {
							if i%10 == 9 {
								fmt.Print(".")
							}
							if f.IsLocalFile {
								log.Printf("%s include local file definition %s which doesn't make sense. Ignoring this.", fields[1], f.File)
							} else {
								if err := server.CacheFile(f.File); err != nil {
									fmt.Println(colorize(fmt.Sprintf("server ERROR sending file #%d: %v",
										i+1, err), "Red", mono))
									break handle_input
								}
							}
						}
						fmt.Println("done")
					}
				*/

			case "M":
				// M filenames
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], true, true); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "M?":
				// M" id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M: <serverID>", "Red", mono))
					break
				}
				if err := server.CacheFile(fields[1]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "M@":
				// M@ id
				if len(fields) != 2 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: M@ <file>", "Red", mono))
					break
				}
				if err := server.LoadFrom(fields[1], false, true); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "MARK":
				// MARK x y
				v, err := tcllist.ConvertTypes(fields, "sff")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				if err := server.Mark(v[1].(float64), v[2].(float64)); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "OA":
				// OA id kvlist
				// It just happens to be a side effect of the protocol definition that we
				// can accept string values for the attributes here regardless of their
				// actual types, so we need not worry about converting them now.
				//
				// TODO: it might still be prudent to check anyway, but the point of the
				// console isn't to have all safety belts engaged for the user.

				if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: OA <id> {<k> <v> ...}", "Red", mono))
					break
				}
				alist, err := tcllist.ParseTclList(fields[2])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: can't parse kv list: %v", err), "Red", mono))
					break
				}
				if (len(alist) % 2) != 0 {
					fmt.Println(colorize("usage ERROR: kv list must have an even number of elements", "Red", mono))
					break
				}
				attrs := make(map[string]any)
				for i := 0; i < len(alist); i += 2 {
					attrs[alist[i]] = alist[i+1]
				}
				if err := server.UpdateObjAttributes(fields[1], attrs); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "OA+", "OA-":
				// OA+ id key vlist
				// OA- id key vlist
				if len(fields) != 4 {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: wrong number of fields: %s <id> <k> {<v> ...}", fields[0]), "Red", mono))
					break
				}
				vlist, err := tcllist.ParseTclList(fields[3])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: can't parse value list: %v", err), "Red", mono))
					break
				}
				if fields[0] == "OA+" {
					if err := server.AddObjAttributes(fields[1], fields[2], vlist); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				} else {
					if err := server.RemoveObjAttributes(fields[1], fields[2], vlist); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break
					}
				}

			case "POLO":
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: POLO", "Red", mono))
					break
				}
				server.Polo()

			case "PS":
				// PS id color name area size player|monster x y reach
				// 0  1  2     3    4    5    6              7 8 9
				v, err := tcllist.ConvertTypes(fields, "sssssssffi")
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
					break
				}
				c := mapper.CreatureToken{
					CreatureType: mapper.CreatureTypeUnknown,
				}
				c.ID = v[1].(string)
				c.Color = v[2].(string)
				c.Name = v[3].(string)
				c.Area = v[4].(string)
				c.Size = v[5].(string)
				c.Gx = v[7].(float64)
				c.Gy = v[8].(float64)
				c.Reach = v[9].(int)

				switch v[6].(string) {
				case "player":
					c.CreatureType = mapper.CreatureTypePlayer
					if err := server.PlaceSomeone(mapper.PlayerToken{
						CreatureToken: c,
					}); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case "monster":
					c.CreatureType = mapper.CreatureTypeMonster
					if err := server.PlaceSomeone(mapper.MonsterToken{
						CreatureToken: c,
					}); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				default:
					fmt.Println(colorize("usage ERROR: creature type must be \"monster\" or \"player\"", "Red", mono))
				}

			case "SYNC":
				// SYNC [CHAT [target]]
				switch len(fields) {
				case 1:
					if err := server.Sync(); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case 2:
					if err := server.SyncChat(0); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				case 3:
					v, err := tcllist.ConvertTypes(fields, "ssi")
					if err != nil {
						fmt.Println(colorize(fmt.Sprintf("usage ERROR: %v", err), "Red", mono))
						break handle_input
					}
					if err := server.SyncChat(v[2].(int)); err != nil {
						fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
						break handle_input
					}

				default:
					fmt.Println(colorize("usage ERROR: SYNC [CHAT [target]]", "Red", mono))
				}

			case "TO":
				// TO recips message
				if len(fields) != 3 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: TO <recip>|*|% <message>", "Red", mono))
					break
				}
				recips, err := tcllist.ParseTclList(fields[1])
				if err != nil {
					fmt.Println(colorize(fmt.Sprintf("ERROR in recipient list: %v", err), "Red", mono))
					break
				}
				if err := server.ChatMessage(recips, fields[2]); err != nil {
					fmt.Println(colorize(fmt.Sprintf("server ERROR: %v", err), "Red", mono))
					break
				}

			case "/CONN":
				if len(fields) != 1 {
					fmt.Println(colorize("usage ERROR: wrong number of fields: /CONN", "Red", mono))
					break
				}
				server.QueryPeers()

			case "EXIT", "QUIT":
				// stop this client
				break inputloop

			default:
				fmt.Println(colorize(fmt.Sprintf("ERROR: Unrecognized command: %s", fields[0]), "Red", mono))
			}
		}
		fmt.Printf("map-console> ")
	}
	fmt.Println(colorize("Shutting down...", "Yellow", mono))
	cancel()
}

func colorize(text, color string, mono bool) string {
	if mono {
		return text
	}
	var prefix string
	switch color {
	case "blue":
		prefix = "34"
	case "Blue":
		prefix = "1;34"
	case "cyan":
		prefix = "36"
	case "Cyan":
		prefix = "1;36"
	case "green":
		prefix = "32"
	case "Green":
		prefix = "1;32"
	case "magenta":
		prefix = "35"
	case "Magenta":
		prefix = "1;35"
	case "red":
		prefix = "31"
	case "Red":
		prefix = "1;31"
	case "yellow":
		prefix = "33"
	case "Yellow":
		prefix = "1;33"
	}
	return "\x1b[" + prefix + "m" + text + "\x1b[0m"
}

/*
# @[00]@| GMA 5.0.0
# @[01]@|
# @[10]@| Copyright © 1992–2022 by Steven L. Willoughby (AKA MadScienceZone)
# @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
# @[12]@| Aloha, Oregon, USA. All Rights Reserved.
# @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
# @[14]@| License as described in the accompanying LICENSE file distributed
# @[15]@| with GMA.
# @[16]@|
# @[20]@| Redistribution and use in source and binary forms, with or without
# @[21]@| modification, are permitted provided that the following conditions
# @[22]@| are met:
# @[23]@| 1. Redistributions of source code must retain the above copyright
# @[24]@|    notice, this list of conditions and the following disclaimer.
# @[25]@| 2. Redistributions in binary form must reproduce the above copy-
# @[26]@|    right notice, this list of conditions and the following dis-
# @[27]@|    claimer in the documentation and/or other materials provided
# @[28]@|    with the distribution.
# @[29]@| 3. Neither the name of the copyright holder nor the names of its
# @[30]@|    contributors may be used to endorse or promote products derived
# @[31]@|    from this software without specific prior written permission.
# @[32]@|
# @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
# @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
# @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
# @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
# @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
# @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
# @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
# @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
# @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
# @[45]@| SUCH DAMAGE.
# @[46]@|
# @[50]@| This software is not intended for any use or application in which
# @[51]@| the safety of lives or property would be at risk due to failure or
# @[52]@| defect of the software.
*/
