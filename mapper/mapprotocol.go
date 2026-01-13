/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______   ______      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \ / ___  \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \\/   \  \   | (  )  |     #
# | |      | || || || (___) | Assistant | (____        ___) /   ___) /   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (   (___ (    | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )         ) \      ) \   |   / | |     #
# | (___) || )   ( || )   ( |           /\____) ) _ /\___/  //\___/  / _ |  (__) |     #
# (_______)|/     \||/     \|           \______/ (_)\______/ \______/ (_)(_______)     #
#                                                                                      #
########################################################################################
*/

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
	"strconv"
	"strings"
	"sync"
	"time"
)

// The GMA Mapper Protocol version number current as of this build,
// and protocol versions supported by this code.
const (
	GMAMapperProtocol           = 422      // @@##@@ auto-configured
	GoVersionNumber             = "5.32.1" // @@##@@ auto-configured
	MinimumSupportedMapProtocol = 400
	MaximumSupportedMapProtocol = 423
	MaxServerMessageSize        = 60 * 1024 // don't send server messages bigger than this
)

func init() {
	if MinimumSupportedMapProtocol > GMAMapperProtocol || MaximumSupportedMapProtocol < GMAMapperProtocol {
		if MinimumSupportedMapProtocol == MaximumSupportedMapProtocol {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapclient only supports mapper protocol %v, but version %v was the official one when this package was released!", MinimumSupportedMapProtocol, GMAMapperProtocol))
		} else {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapclient only supports mapper protocols %v-%v, but version %v was the official one when this package was released!", MinimumSupportedMapProtocol, MaximumSupportedMapProtocol, GMAMapperProtocol))
		}
	}
}

// ErrProtocol is the error returned when there is a protocol-level issue.
// This generally indicates a bug in the code, not a communications issue.
var ErrProtocol = errors.New("internal protocol error")

type MapConnection struct {
	conn     net.Conn               // network socket
	reader   *bufio.Scanner         // read interface to socket
	writer   *bufio.Writer          // write interface to socket
	sendBuf  []string               // internal buffer of outgoing packets
	sendChan chan string            // outgoing packets go through this channel
	batches  map[string]map[int]any // storage for incoming batched packets	(batchID->batch#->packet)
	bLock    *sync.Mutex            // mutex protecting batches
	debug    func(DebugFlags, string)
	debugf   func(DebugFlags, string, ...any)
}

// RetrieveBatches retrieves all the batches belonging to a set and removes them from storage
func (m *MapConnection) RetrieveBatches(packet any) ([]any, error) {
	if b, isBatch := packet.(BatchableMessagePayload); isBatch {
		m.bLock.Lock()
		defer m.bLock.Unlock()

		storage := m.batches[b.BatchGroup]
		storageLen := len(storage)
		if storageLen != b.TotalBatches {
			delete(m.batches, b.BatchGroup)
			return nil, fmt.Errorf("incomplete or corrupt batched payload: expected %d, received %d", b.TotalBatches, storageLen)
		}

		packets := make([]any, 0)
		for i := range storageLen {
			packets = append(packets, storage[i])
		}
		delete(m.batches, b.BatchGroup)
		return packets, nil
	}
	return nil, fmt.Errorf("incoming packet does not appear to be a batch we can unpack")
}

// StashBatch stashes an incoming message payload which is part of a batched set, assuming we'll assemble all of the
// pieces later. It returns true if we are still expecting more to arrive and an error if one occurred.
// If an error is returned, the meaning of the boolean return value is undefined.
func (m *MapConnection) StashBatch(packet any) (bool, error) {
	if b, isBatch := packet.(BatchableMessagePayload); isBatch {
		if b.BatchGroup == "" {
			return false, fmt.Errorf("missing BatchGroup")
		}

		m.bLock.Lock()
		defer m.bLock.Unlock()

		if m.batches == nil {
			m.batches = make(map[string]map[int]any)
		}
		if m.batches[b.BatchGroup] == nil {
			m.batches[b.BatchGroup] = make(map[int]any)
		} else {
			m.batches[b.BatchGroup][b.Batch] = packet
		}
		return b.TotalBatches > len(m.batches[b.BatchGroup]), nil
	}
	return false, fmt.Errorf("packet does not appear to be part of a batch")
}

func (m *MapConnection) IsReady() bool {
	return m != nil && m.reader != nil && m.writer != nil
}

func NewMapConnection(c net.Conn) MapConnection {
	return MapConnection{
		conn:     c,
		reader:   bufio.NewScanner(c),
		writer:   bufio.NewWriter(c),
		sendChan: make(chan string, 50),
	}
}

func (c *MapConnection) Close() {
	if c != nil && c.conn != nil {
		c.conn.Close()
	}
}

// Batchable is any payload which may be split up into multiple batches.
type Batchable interface {
	NeedsToBeSplit() bool                            // does this payload need to be split up because it's already too large to be sent as it is?
	IsBatched() bool                                 // is this payload part of an incoming batch of payloads?
	Split() []any                                    // split up the message, returning the slice of batched payloads
	AbortPayload(reason string, batchNumber int) any // generate an abort payload
	Reassemble([]any) error                          // reassemble a slice of batches into a single payload structure
}

// SendEchoWithTimestamp is identical to Send, but only takes an EchoMessagePayload parameter
// and writes the SentTime value into it as it sends it out.
func (c *MapConnection) SendEchoWithTimestamp(command ServerMessage, data EchoMessagePayload) error {
	data.SentTime = time.Now()
	return c.Send(command, data)
}

// Send sends a message to the peer using the mapper protocol.
func (c *MapConnection) Send(command ServerMessage, data any) error {
	if c == nil {
		return fmt.Errorf("nil MapConnection")
	}

	switch command {
	case Accept:
		if msgs, ok := data.(AcceptMessagePayload); ok {
			return c.sendJSON("ACCEPT", msgs)
		}
	case AddAudio:
		if aa, ok := data.(AudioDefinition); ok {
			return c.sendJSON("AA", aa)
		}
		if aa, ok := data.(AddAudioMessagePayload); ok {
			return c.sendJSON("AA", aa)
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
	case Allow:
		if al, ok := data.(AllowMessagePayload); ok {
			return c.sendJSON("ALLOW", al)
		}
	case Auth:
		if au, ok := data.(AuthMessagePayload); ok {
			return c.sendJSON("AUTH", au)
		}
	case Challenge:
		if ch, ok := data.(ChallengeMessagePayload); ok {
			return c.sendJSON("OK", ch)
		}
	case CharacterName:
		if cn, ok := data.(CharacterNameMessagePayload); ok {
			return c.sendJSON("AKA", cn)
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
	case DefineDicePresetDelegates:
		if dd, ok := data.(DefineDicePresetDelegatesMessagePayload); ok {
			return c.sendJSON("DDD", dd)
		}
	case Denied:
		if reason, ok := data.(DeniedMessagePayload); ok {
			return c.sendJSON("DENIED", reason)
		}
	case Echo:
		if e, ok := data.(EchoMessagePayload); ok {
			return c.sendJSON("ECHO", e)
		}
	case Failed:
		if fa, ok := data.(FailedMessagePayload); ok {
			return c.sendJSON("FAILED", fa)
		}
	case FilterAudio:
		if fi, ok := data.(FilterAudioMessagePayload); ok {
			return c.sendJSON("AA/", fi)
		}
	case FilterCoreData:
		if fi, ok := data.(FilterCoreDataMessagePayload); ok {
			return c.sendJSON("CORE/", fi)
		}
	case FilterDicePresets:
		if fi, ok := data.(FilterDicePresetsMessagePayload); ok {
			return c.sendJSON("DD/", fi)
		}
	case FilterImages:
		if fi, ok := data.(FilterImagesMessagePayload); ok {
			return c.sendJSON("AI/", fi)
		}
	case Granted:
		if reason, ok := data.(GrantedMessagePayload); ok {
			return c.sendJSON("GRANTED", reason)
		}
	case HitPointAcknowledge:
		if ha, ok := data.(HitPointAcknowledgeMessagePayload); ok {
			return c.sendJSON("HPACK", ha)
		}
	case HitPointRequest:
		if hr, ok := data.(HitPointRequestMessagePayload); ok {
			return c.sendJSON("HPREQ", hr)
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
	case PlayAudio:
		if pa, ok := data.(PlayAudioMessagePayload); ok {
			return c.sendJSON("SOUND", pa)
		}
	case Polo:
		return c.sendln("POLO", "")
	case Priv:
		if reason, ok := data.(PrivMessagePayload); ok {
			return c.sendJSON("PRIV", reason)
		}
	case Protocol:
		return c.sendln("PROTOCOL", fmt.Sprintf("%v", data))
	case QueryAudio:
		if qi, ok := data.(AudioDefinition); ok {
			return c.sendJSON("AA?", qi)
		}
		if qi, ok := data.(QueryAudioMessagePayload); ok {
			return c.sendJSON("AA?", qi)
		}
	case QueryCoreData:
		if q, ok := data.(QueryCoreDataMessagePayload); ok {
			return c.sendJSON("CORE", q)
		}
	case QueryCoreIndex:
		if q, ok := data.(QueryCoreIndexMessagePayload); ok {
			return c.sendJSON("COREIDX", q)
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
	case Redirect:
		if red, ok := data.(RedirectMessagePayload); ok {
			return c.sendJSON("REDIRECT", red)
		}
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
	case TimerAcknowledge:
		if ta, ok := data.(TimerAcknowledgeMessagePayload); ok {
			return c.sendJSON("TMACK", ta)
		}
	case TimerRequest:
		if tr, ok := data.(TimerRequestMessagePayload); ok {
			return c.sendJSON("TMRQ", tr)
		}
	case Toolbar:
		if tb, ok := data.(ToolbarMessagePayload); ok {
			return c.sendJSON("TB", tb)
		}
	case UpdateClock:
		if uc, ok := data.(UpdateClockMessagePayload); ok {
			return c.sendJSON("CS", uc)
		}
	case UpdateCoreData:
		if uc, ok := data.(UpdateCoreDataMessagePayload); ok {
			return c.sendJSON("CORE=", uc)
		}
	case UpdateCoreIndex:
		if uc, ok := data.(UpdateCoreIndexMessagePayload); ok {
			return c.sendJSON("COREIDX=", uc)
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
			return c.sendJSON("UPDATES", up)
		}
	case World:
		if wo, ok := data.(WorldMessagePayload); ok {
			return c.sendJSON("WORLD", wo)
		}
	}
	return fmt.Errorf("send: invalid command or data type")
}

func (c *MapConnection) sendJSON(commandWord string, data any) error {
	var err error
	if c == nil {
		return fmt.Errorf("nil MapConnection")
	}
	if data == nil {
		return c.sendln(commandWord, "")
	}

	bail := func(b Batchable, reason string, origError error, batch int) error {
		j, err := json.Marshal(b.AbortPayload(fmt.Sprintf("Error: %s: %v", reason, origError), batch))
		if err != nil {
			return fmt.Errorf("send error %v for batch %d: %v", origError, batch, err)
		}
		return c.sendln(commandWord, string(j))
	}

	splitIntoBatches := func(b Batchable) error {
		for i, batch := range b.Split() {
			j, err := json.Marshal(batch)
			if err != nil {
				return bail(b, "marshaling payload", err, i)
			}
			err = c.sendln(commandWord, string(j))
			if err != nil {
				return bail(b, "sending payload", err, i)
			}
		}
		return nil
	}

	if b, isBatchable := data.(Batchable); isBatchable && b.NeedsToBeSplit() {
		return splitIntoBatches(b)
	}

	if j, err := json.Marshal(data); err == nil {
		sj := string(j)
		if len(sj)+len(commandWord)+2 > MaxServerMessageSize {
			if b, isBatchable := data.(Batchable); isBatchable {
				// the up-front batchable check didn't predict we needed to do this but we ended up here anyway,
				// perhaps because of the cost of character encoding or something. Let's split it up now.
				return splitIntoBatches(b)
			}
			// Otherwise this will fail, but we handle that case in sendln...
		}
		return c.sendln(commandWord, sj)
	}
	return fmt.Errorf("send: %v", err)
}

func (c *MapConnection) sendln(commandWord, data string) error {
	if c == nil {
		return fmt.Errorf("nil MapConnection")
	}
	if c.debugf != nil {
		c.debugf(DebugIO|DebugMessages, "->%s %s", commandWord, data)
	}

	if strings.ContainsAny(data, "\n\r") {
		return fmt.Errorf("protocol error: outgoing data packet may not contain newlines")
	}

	if len(data)+len(commandWord)+2 > MaxServerMessageSize {
		c.sendChan <- fmt.Sprintf("FAILED {\"Command\": \"%s\",\"Reason\":\"Transmission failed for server message; payload length %d exceeds maximum allowed\"}\n", commandWord, len(data))
		return fmt.Errorf("protocol error: outgoing data packet length %d would exceed maximum allowed", len(data))
	}

	var packet strings.Builder
	packet.WriteString(commandWord)
	if data != "" {
		packet.WriteString(" ")
		packet.WriteString(data)
	}
	packet.WriteString("\n")

	//	select {
	//	case c.sendChan <- packet.String():
	//	default:
	//		return fmt.Errorf("unable to send to server (Dial() not running or data backed up?")
	//	}
	c.sendChan <- packet.String()
	return nil
}

// blocking raw data sent to other side
func (c *MapConnection) sendRaw(data string) error {
	if c != nil {
		c.sendChan <- data + "\n"
	}
	return nil
}

// UNSAFEsendRaw will send raw data to the server without any checks or controls.
// Use this function at your own risk. If you don't phrase the data perfectly, the server will
// not understand your request. This is intended only for testing purposes including manually
// communicating with the server for debugging.
func (c *MapConnection) UNSAFEsendRaw(data string) error {
	return c.sendRaw(data)
}

// UNSAFEsendRaw will send raw data to the server without any checks or controls.
// Use this function at your own risk. If you don't phrase the data perfectly, the server will
// not understand your request. This is intended only for testing purposes including manually
// communicating with the server for debugging.
func (c *Connection) UNSAFEsendRaw(data string) error {
	return c.serverConn.sendRaw(data)
}

// Receive waits for a message to arrive on the MapConnection's input then returns it.
func (c *MapConnection) Receive() (MessagePayload, error) {
	var err error
	var rescan bool

	if c == nil {
		return nil, fmt.Errorf("Receive called on nil MapConnection")
	}

rescan_input:
	if !c.reader.Scan() {
		//c.debug(DebugIO, "Receive: scan failed; stopping")
		if err = c.reader.Err(); err != nil {
			//c.debugf(DebugIO, "Receive: scan failed with %v", err)
			return nil, err
		}
		return nil, nil
	}

	// Comments are anything starting with "//"
	// The input line is in the form COMMAND-WORD [JSON] \n
	c.debugf(DebugIO|DebugMessages, "<-%v", c.reader.Text())
	payload := BaseMessagePayload{
		rawMessage: c.reader.Text(),
	}
	commandWord, jsonString, hasJsonPart := strings.Cut(c.reader.Text(), " ")
	if strings.Index(commandWord, "//") == 0 {
		payload.messageType = Comment
		return CommentMessagePayload{
			BaseMessagePayload: payload,
			Text:               c.reader.Text()[2:],
		}, nil
	}

	handleBatching := func(p any) any {
		var moreRemaining bool
		rescan = false
		if b, isBatchable := p.(Batchable); isBatchable && b.IsBatched() {
			moreRemaining, err = c.StashBatch(p)
			if err != nil {
				return p
			}
			if moreRemaining {
				rescan = true
				return nil
			}
			packets, err := c.RetrieveBatches(p)
			if err != nil {
				return p
			}
			p = b.Reassemble(packets)
		}
		return p
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
		p = handleBatching(p).(AddCharacterMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "ACCEPT":
		p := AcceptMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Accept
		p = handleBatching(p).(AcceptMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AA":
		p := AddAudioMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddAudio
		p = handleBatching(p).(AddAudioMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AA?":
		p := QueryAudioMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryAudio
		p = handleBatching(p).(QueryAudioMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AA/":
		p := FilterAudioMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = FilterAudio
		p = handleBatching(p).(FilterAudioMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AI":
		p := AddImageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddImage
		p = handleBatching(p).(AddImageMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AI?":
		p := QueryImageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryImage
		p = handleBatching(p).(QueryImageMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AI/":
		p := FilterImagesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = FilterImages
		p = handleBatching(p).(FilterImagesMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AKA":
		p := CharacterNameMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = CharacterName
		p = handleBatching(p).(CharacterNameMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "ALLOW":
		p := AllowMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Allow
		p = handleBatching(p).(AllowMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "AUTH":
		p := AuthMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Auth
		if err != nil {
			break
		}
		return p, nil

	case "AV":
		p := AdjustViewMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AdjustView
		if err != nil {
			break
		}
		return p, nil

	case "CC":
		p := ClearChatMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ClearChat
		if err != nil {
			break
		}
		return p, nil

	case "CLR":
		p := ClearMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Clear
		if err != nil {
			break
		}
		return p, nil

	case "CLR@":
		p := ClearFromMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ClearFrom
		if err != nil {
			break
		}
		return p, nil

	case "CO":
		p := CombatModeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = CombatMode
		if err != nil {
			break
		}
		return p, nil

	case "CONN":
		p := UpdatePeerListMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdatePeerList
		p = handleBatching(p).(UpdatePeerListMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "CORE":
		p := QueryCoreDataMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryCoreData
		p = handleBatching(p).(QueryCoreDataMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "COREIDX":
		p := QueryCoreIndexMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryCoreIndex
		p = handleBatching(p).(QueryCoreIndexMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "CORE/":
		p := FilterCoreDataMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = FilterCoreData
		p = handleBatching(p).(FilterCoreDataMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "CORE=":
		p := UpdateCoreDataMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateCoreData
		p = handleBatching(p).(UpdateCoreDataMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "COREIDX=":
		p := UpdateCoreIndexMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateCoreIndex
		p = handleBatching(p).(UpdateCoreIndexMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "CS":
		p := UpdateClockMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateClock
		if err != nil {
			break
		}
		return p, nil

	case "D":
		p := RollDiceMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RollDice
		p = handleBatching(p).(RollDiceMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DD":
		p := DefineDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = DefineDicePresets
		p = handleBatching(p).(DefineDicePresetsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DDD":
		p := DefineDicePresetDelegatesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = DefineDicePresetDelegates
		p = handleBatching(p).(DefineDicePresetDelegatesMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DD+":
		p := AddDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddDicePresets
		p = handleBatching(p).(AddDicePresetsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DD/":
		p := FilterDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = FilterDicePresets
		p = handleBatching(p).(FilterDicePresetsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DD=":
		p := UpdateDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateDicePresets
		p = handleBatching(p).(UpdateDicePresetsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DENIED":
		p := DeniedMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Denied
		if err != nil {
			break
		}
		return p, nil

	case "DR":
		p := QueryDicePresetsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = QueryDicePresets
		p = handleBatching(p).(QueryDicePresetsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "DSM":
		p := UpdateStatusMarkerMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateStatusMarker
		p = handleBatching(p).(UpdateStatusMarkerMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "ECHO":
		p := EchoMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Echo
		p = handleBatching(p).(EchoMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "FAILED":
		p := FailedMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Failed
		if err != nil {
			break
		}
		return p, nil

	case "GRANTED":
		p := GrantedMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Granted
		if err != nil {
			break
		}
		return p, nil

	case "HPACK":
		p := HitPointAcknowledgeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = HitPointAcknowledge
		p = handleBatching(p).(HitPointAcknowledgeMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "HPREQ":
		p := HitPointRequestMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = HitPointRequest
		p = handleBatching(p).(HitPointRequestMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "I":
		p := UpdateTurnMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateTurn
		p = handleBatching(p).(UpdateTurnMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "IL":
		p := UpdateInitiativeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateInitiative
		p = handleBatching(p).(UpdateInitiativeMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "L":
		p := LoadFromMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadFrom
		p = handleBatching(p).(LoadFromMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-ARC":
		p := LoadArcObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadArcObject
		p = handleBatching(p).(LoadArcObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-CIRC":
		p := LoadCircleObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadCircleObject
		p = handleBatching(p).(LoadCircleObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-LINE":
		p := LoadLineObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadLineObject
		p = handleBatching(p).(LoadLineObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-POLY":
		p := LoadPolygonObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadPolygonObject
		p = handleBatching(p).(LoadPolygonObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-RECT":
		p := LoadRectangleObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadRectangleObject
		p = handleBatching(p).(LoadRectangleObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-SAOE":
		p := LoadSpellAreaOfEffectObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadSpellAreaOfEffectObject
		p = handleBatching(p).(LoadSpellAreaOfEffectObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-TEXT":
		p := LoadTextObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadTextObject
		p = handleBatching(p).(LoadTextObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "LS-TILE":
		p := LoadTileObjectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = LoadTileObject
		p = handleBatching(p).(LoadTileObjectMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "MARCO":
		p := MarcoMessagePayload{BaseMessagePayload: payload}
		p.messageType = Marco
		if err != nil {
			break
		}
		return p, nil

	case "MARK":
		p := MarkMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Mark
		if err != nil {
			break
		}
		return p, nil

	case "OA":
		p := UpdateObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateObjAttributes
		p = handleBatching(p).(UpdateObjAttributesMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "OA+":
		p := AddObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = AddObjAttributes
		p = handleBatching(p).(AddObjAttributesMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "OA-":
		p := RemoveObjAttributesMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RemoveObjAttributes
		p = handleBatching(p).(RemoveObjAttributesMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "OK":
		p := ChallengeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Challenge
		if err != nil {
			break
		}
		return p, nil

	case "POLO":
		p := PoloMessagePayload{BaseMessagePayload: payload}
		p.messageType = Polo
		if err != nil {
			break
		}
		return p, nil

	case "PRIV":
		p := PrivMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Priv
		if err != nil {
			break
		}
		return p, nil

	case "PROGRESS":
		p := UpdateProgressMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateProgress
		if err != nil {
			break
		}
		return p, nil

	case "PROTOCOL":
		p := ProtocolMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			// not really JSON for this command; just the protocol version as an integer
			p.ProtocolVersion, err = strconv.Atoi(jsonString)
			if err != nil {
				break
			}
		} else {
			err = fmt.Errorf("Server PROTOCOL command invalid (no version value)")
			break
		}
		return p, nil

	case "PS":
		p := PlaceSomeoneMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = PlaceSomeone
		p = handleBatching(p).(PlaceSomeoneMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "READY":
		p := ReadyMessagePayload{BaseMessagePayload: payload}
		p.messageType = Ready
		if err != nil {
			break
		}
		return p, nil

	case "REDIRECT":
		p := RedirectMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Redirect
		if err != nil {
			break
		}
		return p, nil

	case "ROLL":
		p := RollResultMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = RollResult
		p = handleBatching(p).(RollResultMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "SOUND":
		p := PlayAudioMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = PlayAudio
		p = handleBatching(p).(PlayAudioMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "SYNC":
		p := SyncMessagePayload{BaseMessagePayload: payload}
		p.messageType = Sync
		p = handleBatching(p).(SyncMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "SYNC-CHAT":
		p := SyncChatMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = SyncChat
		p = handleBatching(p).(SyncChatMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "TB":
		p := ToolbarMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = Toolbar
		if err != nil {
			break
		}
		return p, nil

	case "TMACK":
		p := TimerAcknowledgeMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = TimerAcknowledge
		p = handleBatching(p).(TimerAcknowledgeMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "TMRQ":
		p := TimerRequestMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = TimerRequest
		p = handleBatching(p).(TimerRequestMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "TO":
		p := ChatMessageMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = ChatMessage
		p = handleBatching(p).(ChatMessageMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "UPDATES":
		p := UpdateVersionsMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = UpdateVersions
		p = handleBatching(p).(UpdateVersionsMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "WORLD":
		p := WorldMessagePayload{BaseMessagePayload: payload}
		if hasJsonPart {
			if err = json.Unmarshal([]byte(jsonString), &p); err != nil {
				break
			}
		}
		p.messageType = World
		p = handleBatching(p).(WorldMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	case "/CONN":
		p := QueryPeersMessagePayload{BaseMessagePayload: payload}
		p.messageType = QueryPeers
		p = handleBatching(p).(QueryPeersMessagePayload)
		if rescan {
			goto rescan_input
		}
		if err != nil {
			break
		}
		return p, nil

	default:
		payload.messageType = UNKNOWN
		return payload, nil
	}

	if err != nil {
		payload.messageType = ERROR
		return ErrorMessagePayload{
			BaseMessagePayload: payload,
			Error:              err,
		}, nil
	}

	c.debug(DebugIO, "unable to cope with message, returning nil")
	return nil, fmt.Errorf("bailing out, unable to cope with received packet")
}

// Send out all waiting outbound messages and then return
// DO NOT CALL this if you have a writer routine still running that is managing the sendBuf slice.
func (c *MapConnection) Flush() error {
	// receive all the messages still in the channel
	if c.debug != nil {
		c.debug(DebugIO, "flushing output")
	}
	for {
		select {
		case packet := <-c.sendChan:
			c.sendBuf = append(c.sendBuf, packet)
			if c.debugf != nil {
				c.debugf(DebugIO, "flush: moved %v out to sendBuf (depth %d)", packet, len(c.sendBuf))
			}
		default:
			if c.writer == nil || len(c.sendBuf) == 0 {
				if c.debugf != nil {
					c.debugf(DebugIO, "flush: terminating (writer=%v, sendBuf=%v)", c.writer != nil, c.sendBuf)
				}
				return nil
			}
			if written, err := c.writer.WriteString(c.sendBuf[0]); err != nil {
				return fmt.Errorf("error sending \"%s\" (wrote %d): %v", c.sendBuf[0], written, err)
			}
			if err := c.writer.Flush(); err != nil {
				return fmt.Errorf("error sending \"%s\" (in flush): %v", c.sendBuf[0], err)
			}
			if c.debugf != nil {
				c.debugf(DebugIO, "flush: sent %v (depth now %d)", c.sendBuf[0], len(c.sendBuf)-1)
			}
			c.sendBuf = c.sendBuf[1:]
		}
	}
}
