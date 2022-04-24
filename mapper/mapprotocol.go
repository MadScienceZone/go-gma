//
// Client/Server protocol functions.
//
// Clients and the server send messages to each other without
// waiting for any response (i.e., all communication is asynchronous).
// Regardless of direction, a message has the format
//
// COMMAND-WORD [JSON] \n
//
// Where COMMAND-WORD is a plain text identifier which indicates what
// message is being sent, and JSON is an optional JSON-formatted data
// structure appropriate to that message. The entire message is sent
// in a single newline-terminated line of text. If JSON includes fields
// not expected for that command, those fields are silently ignored.
// If it is missing any expected fields (or the entire JSON object is
// omitted entirely), the missing fields are assumed to have an appropriate
// "zero" value.
//

package mapper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

// ErrProtocol is the error returned when there is a protocol-level issue.
// This generally indicates a bug in the code, not a communications issue.
var ErrProtocol = errors.New("internal protocol error")

type MapConnection struct {
	conn     net.Conn       // network socket
	reader   *bufio.Scanner // read interface to socket
	writer   *bufio.Writer  // write interface to socket
	sendBuf  []string       // internal buffer of outgoing packets
	sendChan chan string    // outgoing packets go through this channel
}

func (c *MapConnection) Close() {
	c.conn.Close()
}

//
// send sends a message to the peer using the mapper protocol.
//
func (c *MapConnection) send(command ServerMessage, data interface{}) error {
	switch command {
	case Accept:
		if msgs, ok := data.(AcceptMessagePayload); ok {
			return c.sendJSON("ACCEPT", msgs)
		}
	case AddCharacter:
		if ac, ok := data.(AddCharacterMessagePayload); ok {
			return c.sendJSON("AC", ac)
		}
	case AddDicePresets:
		if ad, ok := data.(AddDicePresetsMessagePayload); ok {
			return c.sendJSON("DD+", ad)
		}
	case AddImage:
		if ai, ok := data.(ImageDefinition); ok {
			return c.sendJSON("AI", ai)
		}
		if ai, ok := data.(AddImageMessagePayload); ok {
			return c.sendJSON("AI", ai)
		}
	case AddObjAttributes:
		if oa, ok := data.(AddObjAttributesMessagePayload); ok {
			return c.sendJSON("OA+", oa)
		}
	case AdjustView:
		if av, ok := data.(AdjustViewMessagePayload); ok {
			return c.sendJSON("AV", av)
		}
	case Auth:
		if au, ok := data.(AuthMessagePayload); ok {
			return c.sendJSON("AUTH", au)
		}
	case Challenge:
		if ch, ok := data.(ChallengeMessagePayload); ok {
			return c.sendJSON("OK", ch)
		}
	case ChatMessage:
		if ch, ok := data.(ChatMessageMessagePayload); ok {
			return c.sendJSON("TO", ch)
		}
	case Clear:
		if cl, ok := data.(ClearMessagePayload); ok {
			return c.sendJSON("CLR", cl)
		}
	case ClearChat:
		if cc, ok := data.(ClearChatMessagePayload); ok {
			return c.sendJSON("CC", cc)
		}
	case ClearFrom:
		if cf, ok := data.(ClearFromMessagePayload); ok {
			return c.sendJSON("CLR@", cf)
		}
	case CombatMode:
		if cm, ok := data.(CombatModeMessagePayload); ok {
			return c.sendJSON("CO", cm)
		}
	case Comment:
		if data == nil {
			return c.sendln("//", "")
		}
		if s, ok := data.(string); ok {
			return c.sendln("//", s)
		}
	case DefineDicePresets:
		if dd, ok := data.(DefineDicePresetsMessagePayload); ok {
			return c.sendJSON("DD", dd)
		}
	case Denied:
		if reason, ok := data.(DeniedMessagePayload); ok {
			return c.sendJSON("DENIED", reason)
		}
	case FilterDicePresets:
		if fi, ok := data.(FilterDicePresetsMessagePayload); ok {
			return c.sendJSON("DD/", fi)
		}
	case Granted:
		if reason, ok := data.(GrantedMessagePayload); ok {
			return c.sendJSON("GRANTED", reason)
		}
	case LoadFrom:
		if lf, ok := data.(LoadFromMessagePayload); ok {
			return c.sendJSON("L", lf)
		}
	case LoadArcObject:
		if ob, ok := data.(ArcElement); ok {
			return c.sendJSON("LS-ARC", ob)
		}
		if ob, ok := data.(LoadArcObjectMessagePayload); ok {
			return c.sendJSON("LS-ARC", ob)
		}
	case LoadCircleObject:
		if ob, ok := data.(CircleElement); ok {
			return c.sendJSON("LS-CIRC", ob)
		}
		if ob, ok := data.(LoadCircleObjectMessagePayload); ok {
			return c.sendJSON("LS-CIRC", ob)
		}
	case LoadLineObject:
		if ob, ok := data.(LineElement); ok {
			return c.sendJSON("LS-LINE", ob)
		}
		if ob, ok := data.(LoadLineObjectMessagePayload); ok {
			return c.sendJSON("LS-LINE", ob)
		}
	case LoadPolygonObject:
		if ob, ok := data.(PolygonElement); ok {
			return c.sendJSON("LS-POLY", ob)
		}
		if ob, ok := data.(LoadPolygonObjectMessagePayload); ok {
			return c.sendJSON("LS-POLY", ob)
		}
	case LoadRectangleObject:
		if ob, ok := data.(RectangleElement); ok {
			return c.sendJSON("LS-RECT", ob)
		}
		if ob, ok := data.(LoadRectangleObjectMessagePayload); ok {
			return c.sendJSON("LS-RECT", ob)
		}
	case LoadSpellAreaOfEffectObject:
		if ob, ok := data.(SpellAreaOfEffectElement); ok {
			return c.sendJSON("LS-SAOE", ob)
		}
		if ob, ok := data.(LoadSpellAreaOfEffectObjectMessagePayload); ok {
			return c.sendJSON("LS-SAOE", ob)
		}
	case LoadTextObject:
		if ob, ok := data.(TextElement); ok {
			return c.sendJSON("LS-TEXT", ob)
		}
		if ob, ok := data.(LoadTextObjectMessagePayload); ok {
			return c.sendJSON("LS-TEXT", ob)
		}
	case LoadTileObject:
		if ob, ok := data.(TileElement); ok {
			return c.sendJSON("LS-TEXT", ob)
		}
		if ob, ok := data.(LoadTileObjectMessagePayload); ok {
			return c.sendJSON("LS-TILE", ob)
		}
	case Marco:
		return c.sendln("MARCO", "")
	case Mark:
		if mk, ok := data.(MarkMessagePayload); ok {
			return c.sendJSON("MARK", mk)
		}
	case PlaceSomeone:
		if ps, ok := data.(MonsterToken); ok {
			return c.sendJSON("PS", ps)
		}
		if ps, ok := data.(PlayerToken); ok {
			return c.sendJSON("PS", ps)
		}
		if ps, ok := data.(CreatureToken); ok {
			return c.sendJSON("PS", ps)
		}
		if ps, ok := data.(PlaceSomeoneMessagePayload); ok {
			return c.sendJSON("PS", ps)
		}
	case Polo:
		return c.sendln("POLO", "")
	case Priv:
		if reason, ok := data.(PrivMessagePayload); ok {
			return c.sendJSON("PRIV", reason)
		}
	case QueryDicePresets:
		return c.sendln("DR", "")
	case QueryImage:
		if qi, ok := data.(ImageDefinition); ok {
			return c.sendJSON("AI?", qi)
		}
		if qi, ok := data.(QueryImageMessagePayload); ok {
			return c.sendJSON("AI?", qi)
		}
	case QueryPeers:
		return c.sendln("/CONN", "")
	case Ready:
		return c.sendln("READY", "")
	case RemoveObjAttributes:
		if oa, ok := data.(RemoveObjAttributesMessagePayload); ok {
			return c.sendJSON("OA-", oa)
		}
	case RollDice:
		if rd, ok := data.(RollDiceMessagePayload); ok {
			return c.sendJSON("D", rd)
		}
	case RollResult:
		if rd, ok := data.(RollResultMessagePayload); ok {
			return c.sendJSON("ROLL", rd)
		}
	case Sync:
		return c.sendln("SYNC", "")
	case SyncChat:
		if sc, ok := data.(SyncChatMessagePayload); ok {
			return c.sendJSON("SYNC-CHAT", sc)
		}
	case Toolbar:
		if tb, ok := data.(ToolbarMessagePayload); ok {
			return c.sendJSON("TB", tb)
		}
	case UpdateClock:
		if uc, ok := data.(UpdateClockMessagePayload); ok {
			return c.sendJSON("CS", uc)
		}
	case UpdateDicePresets:
		if dd, ok := data.(UpdateDicePresetsMessagePayload); ok {
			return c.sendJSON("DD=", dd)
		}
	case UpdateInitiative:
		if i, ok := data.(UpdateInitiativeMessagePayload); ok {
			return c.sendJSON("IL", i)
		}
	case UpdateObjAttributes:
		if oa, ok := data.(UpdateObjAttributesMessagePayload); ok {
			return c.sendJSON("OA", oa)
		}
	case UpdatePeerList:
		if up, ok := data.(UpdatePeerListMessagePayload); ok {
			return c.sendJSON("CONN", up)
		}
	case UpdateProgress:
		if up, ok := data.(UpdateProgressMessagePayload); ok {
			return c.sendJSON("PROGRESS", up)
		}
	case UpdateStatusMarker:
		if sm, ok := data.(StatusMarkerDefinition); ok {
			return c.sendJSON("DSM", sm)
		}
		if sm, ok := data.(UpdateStatusMarkerMessagePayload); ok {
			return c.sendJSON("DSM", sm)
		}
	case UpdateTurn:
		if tu, ok := data.(UpdateTurnMessagePayload); ok {
			return c.sendJSON("I", tu)
		}
	case UpdateVersions:
		if up, ok := data.(UpdateVersionsMessagePayload); ok {
			return c.sendJSON("AUTH", up)
		}
	case World:
		if wo, ok := data.(WorldMessagePayload); ok {
			return c.sendJSON("WORLD", wo)
		}
	case WriteOnly:
		return c.sendln("NO", "")
	}

	return ErrProtocol
}

func (c *MapConnection) sendJSON(commandWord string, data interface{}) error {
	if data == nil {
		return c.sendln(commandWord, "")
	}
	if j, err := json.Marshal(data); err != nil {
		return c.sendln(commandWord, string(j))
	}
	return ErrProtocol
}

func (c *MapConnection) sendln(commandWord, data string) error {
	if strings.ContainsAny(data, "\n\r") {
		return fmt.Errorf("protocol error: outgoing data packet may not contain newlines")
	}
	var packet strings.Builder

	packet.WriteString(commandWord)
	if data != "" {
		packet.WriteString(" ")
		packet.WriteString(data)
	}
	packet.WriteString("\n")

	select {
	case c.sendChan <- packet.String():
	default:
		return fmt.Errorf("unable to send to server (Dial() not running?")
	}
	return nil
}

//
// receive waits for a message to arrive on its input then returns it.
//
func (c *MapConnection) receive(done chan error) MessagePayload {
	var err error
	if !c.reader.Scan() {
		if err = c.reader.Err(); err != nil {
			done <- err
		}
		return nil
	}

	// Comments are anything starting with "//"
	// The input line is in the form COMMAND-WORD [JSON] \n
	payload := BaseMessagePayload{
		rawMessage: c.reader.Text(),
	}
	commandWord, jsonString, hasJsonPart := strings.Cut(c.reader.Text(), " ")
	if strings.Index(commandWord, "//") == 0 {
		payload.messageType = Comment
		return CommentMessagePayload{
			BaseMessagePayload: payload,
			Text:               c.reader.Text()[2:],
		}
	}

	switch commandWord {
	case "AC":
		p := AddCharacterMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddCharacter
		return p

	case "ACCEPT":
		p := AcceptMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Accept
		return p

	case "AI":
		p := AddImageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddImage
		return p

	case "AI?":
		p := QueryImageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryImage
		return p

	case "AUTH":
		p := AuthMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Auth
		return p

	case "AV":
		p := AdjustViewMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AdjustView
		return p

	case "CC":
		p := ClearChatMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ClearChat
		return p

	case "CLR":
		p := ClearMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Clear
		return p

	case "CLR@":
		p := ClearFromMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ClearFrom
		return p

	case "CO":
		p := CombatModeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = CombatMode
		return p

	case "CONN":
		p := UpdatePeerListMessagePayload{BaseMessagePayload: payload}
		p.messageType = UpdatePeerList
		return p

	case "CS":
		p := UpdateClockMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateClock
		return p

	case "D":
		p := RollDiceMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RollDice
		return p

	case "DD":
		p := DefineDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = DefineDicePresets
		return p

	case "DD+":
		p := AddDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddDicePresets
		return p

	case "DD/":
		p := FilterDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = FilterDicePresets
		return p

	case "DD=":
		p := UpdateDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateDicePresets
		return p

	case "DENIED":
		p := DeniedMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Denied
		return p

	case "DR":
		p := QueryDicePresetsMessagePayload{BaseMessagePayload: payload}
		p.messageType = QueryDicePresets
		return p

	case "DSM":
		p := UpdateStatusMarkerMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateStatusMarker
		return p

	case "GRANTED":
		p := GrantedMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Granted
		return p

	case "I":
		p := UpdateTurnMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateTurn
		return p

	case "IL":
		p := UpdateInitiativeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateInitiative
		return p

	case "L":
		p := LoadFromMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadFrom
		return p

	case "LS-ARC":
		p := LoadArcObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadArcObject
		return p

	case "LS-CIRC":
		p := LoadCircleObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadCircleObject
		return p

	case "LS-LINE":
		p := LoadLineObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadLineObject
		return p

	case "LS-POLY":
		p := LoadPolygonObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadPolygonObject
		return p

	case "LS-RECT":
		p := LoadRectangleObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadRectangleObject
		return p

	case "LS-SAOE":
		p := LoadSpellAreaOfEffectObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadSpellAreaOfEffectObject
		return p

	case "LS-TEXT":
		p := LoadTextObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadTextObject
		return p

	case "LS-TILE":
		p := LoadTileObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadTileObject
		return p

	case "MARCO":
		p := MarcoMessagePayload{BaseMessagePayload: payload}
		p.messageType = Marco
		return p

	case "MARK":
		p := MarkMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Mark
		return p

	case "OA":
		p := UpdateObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateObjAttributes
		return p

	case "OA+":
		p := AddObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddObjAttributes
		return p

	case "OA-":
		p := RemoveObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RemoveObjAttributes
		return p

	case "OK":
		p := ChallengeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Challenge
		return p

	case "PRIV":
		p := PrivMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Priv
		return p

	case "POLO":
		p := PoloMessagePayload{BaseMessagePayload: payload}
		p.messageType = Polo
		return p

	case "PROGRESS":
		p := UpdateProgressMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateProgress
		return p

	case "PS":
		p := PlaceSomeoneMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = PlaceSomeone
		return p

	case "READY":
		p := ReadyMessagePayload{BaseMessagePayload: payload}
		p.messageType = Ready
		return p

	case "ROLL":
		p := RollResultMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RollResult
		return p

	case "SYNC":
		p := SyncMessagePayload{BaseMessagePayload: payload}
		p.messageType = Sync
		return p

	case "SYNC-CHAT":
		p := SyncChatMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = SyncChat
		return p

	case "TB":
		p := ToolbarMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Toolbar
		return p

	case "TO":
		p := ChatMessageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ChatMessage
		return p

	case "UPDATES":
		p := UpdateVersionsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateVersions
		return p

	case "WORLD":
		p := WorldMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = World
		return p

	case "/CONN":
		p := QueryPeersMessagePayload{BaseMessagePayload: payload}
		p.messageType = QueryPeers
		return p

	default:
		payload.messageType = UNKNOWN
		return payload
	}

	if err != nil {
		payload.messageType = ERROR
		return ErrorMessagePayload{
			BaseMessagePayload: payload,
			Error:              err,
		}
	}

	return nil
}
