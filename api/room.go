package api

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/umarbektokyo/matetra-engine/engine"
	"github.com/umarbektokyo/matetra-engine/model"
)

type PlayerConnection struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	PlayerID int
	Username string
}

type Room struct {
	Code        string
	Game        *engine.Game
	connections map[int]*PlayerConnection
	connMu      sync.RWMutex
	nextConnID  int
	hub         *Hub
}

func newRoom(code string, settings CreateRoomPayload) *Room {
	gs := model.GameSettings{
		WinCondition: model.WinCondition(settings.WinCondition),
		TurnLimit:    settings.TurnLimit,
		HandSize:     settings.HandSize,
	}
	if gs.HandSize <= 0 {
		gs.HandSize = 6
	}

	g := engine.New(code, gs)
	if err := g.LoadCards(); err != nil {
		log.Printf("failed to load cards: %v", err)
	}

	return &Room{
		Code:        code,
		Game:        g,
		connections: make(map[int]*PlayerConnection),
	}
}

func (r *Room) AddConnection(conn *websocket.Conn, hub *Hub, username, hash string) {
	r.connMu.Lock()
	pc := &PlayerConnection{conn: conn, PlayerID: -1, Username: username}

	if pid, err := r.Game.Reconnect(username, hash); err == nil {
		pc.PlayerID = pid
		log.Printf("[%s] @%s reconnected as player %d", r.Code, username, pid)
	}

	id := r.nextConnID
	r.connections[id] = pc
	r.nextConnID++
	r.connMu.Unlock()

	if pc.PlayerID >= 0 {
		r.sendResponse(pc, "RECONNECTED", map[string]any{
			"playerID": pc.PlayerID,
			"name":     username,
		})
		r.sendState(pc)
	}

	log.Printf("[%s] @%s connected (conn %d)", r.Code, username, id)
	go r.readMessages(id, pc, hub, username, hash)
}

func (r *Room) readMessages(connID int, pc *PlayerConnection, hub *Hub, username, hash string) {
	defer func() {
		pc.conn.Close()
		if pc.PlayerID >= 0 {
			r.Game.SetPlayerOffline(pc.PlayerID)
		}
		r.connMu.Lock()
		delete(r.connections, connID)
		empty := len(r.connections) == 0
		r.connMu.Unlock()

		log.Printf("[%s] @%s disconnected", r.Code, username)
		if empty {
			hub.RemoveRoom(r.Code)
			log.Printf("[%s] room removed (empty)", r.Code)
		}
	}()

	for {
		var msg Message
		if err := pc.conn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				return
			}
			log.Printf("[%s] read error @%s: %v", r.Code, username, err)
			return
		}
		r.handleMessage(pc, msg, username, hash)
	}
}

func (r *Room) handleMessage(pc *PlayerConnection, msg Message, username, hash string) {
	switch msg.Type {
	case "ADD_PLAYER":
		playerID, err := r.Game.AddPlayer(username, hash)
		if err != nil {
			r.sendError(pc, err.Error())
			return
		}
		pc.PlayerID = playerID
		r.sendResponse(pc, "PLAYER_ADDED", map[string]any{
			"name":     username,
			"playerID": playerID,
		})
		r.BroadcastState()

	case "START_GAME":
		if err := r.Game.StartGame(); err != nil {
			r.sendError(pc, err.Error())
			return
		}
		// On start, broadcast the real state (no queue yet)
		r.BroadcastState()

	case "PLAY_CARD":
		r.handlePlayCard(pc, msg.Payload)

	case "ROLL_DICE":
		r.handleRollDice(pc, msg.Payload)

	case "FINISH_DICE":
		r.handleFinishDice(pc)

	case "END_TURN":
		r.handleEndTurn(pc)

	case "GET_STATE":
		r.sendState(pc)

	default:
		r.sendError(pc, "unknown message type: "+msg.Type)
	}
}

func (r *Room) handlePlayCard(pc *PlayerConnection, payload any) {
	if pc.PlayerID == -1 {
		r.sendError(pc, "not in game yet")
		return
	}
	var p CardPlayPayload
	if err := decodePayload(payload, &p); err != nil {
		r.sendError(pc, "invalid card payload")
		return
	}

	// Queue the card on the real state
	if err := r.Game.QueueCard(pc.PlayerID, p.CardIndex, p.Inputs); err != nil {
		r.sendReply(pc, false, fmt.Sprintf("failed: %v", err))
		return
	}

	// Generate preview (virtual: all queued cards applied, queue preserved for display)
	// and broadcast to ALL players
	r.BroadcastPreview()

	r.sendReply(pc, true, "card queued")
}

func (r *Room) handleRollDice(pc *PlayerConnection, payload any) {
	if pc.PlayerID == -1 {
		r.sendError(pc, "not in game yet")
		return
	}

	var p DiceRollPayload
	if err := decodePayload(payload, &p); err != nil {
		r.sendError(pc, "invalid dice payload")
		return
	}

	var err error
	if p.Single && len(p.Values) >= 1 {
		_, err = r.Game.ProcessDiceRollSingle(pc.PlayerID, p.Values[0])
	} else {
		_, err = r.Game.ProcessDiceRoll(pc.PlayerID, p.Values)
	}

	if err != nil {
		r.sendReply(pc, false, fmt.Sprintf("dice: %v", err))
		return
	}

	// Dice changes the real state, broadcast preview (which includes dice changes + queue applied)
	r.BroadcastPreview()
}

func (r *Room) handleFinishDice(pc *PlayerConnection) {
	if pc.PlayerID == -1 {
		r.sendError(pc, "not in game yet")
		return
	}
	if _, err := r.Game.FinishDiceRoll(pc.PlayerID); err != nil {
		r.sendReply(pc, false, err.Error())
		return
	}
	// No need to broadcast — dice finished is a local flag
	r.sendReply(pc, true, "dice done")
}

func (r *Room) handleEndTurn(pc *PlayerConnection) {
	if pc.PlayerID == -1 {
		r.sendError(pc, "not in game yet")
		return
	}

	_, err := r.Game.ProcessNextTurn(pc.PlayerID)
	if err != nil {
		r.sendReply(pc, false, fmt.Sprintf("failed: %v", err))
		return
	}

	// After turn end, broadcast the real state.
	// If all players are done, cards have been applied and turn advanced — real state is correct.
	// If not all done yet, we should still show the preview of pending cards.
	r.BroadcastPreview()
}

// --- send helpers ---

func (r *Room) sendResponse(pc *PlayerConnection, t string, data any) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.conn.WriteJSON(Message{Type: t, Payload: data})
}

func (r *Room) sendError(pc *PlayerConnection, msg string) {
	r.sendResponse(pc, "ERROR", map[string]string{"message": msg})
}

func (r *Room) sendReply(pc *PlayerConnection, ok bool, msg string) {
	r.sendResponse(pc, "REPLY", CardPlayReply{Success: ok, Message: msg})
}

func (r *Room) sendState(pc *PlayerConnection) {
	// Send preview if game has queue, otherwise real state
	if r.Game.IsStarted() {
		r.sendResponse(pc, "STATE_UPDATE", r.Game.GeneratePreview())
	} else {
		r.sendResponse(pc, "STATE_UPDATE", r.Game.CopyState())
	}
}

// BroadcastState sends the real (non-preview) state to all connections.
func (r *Room) BroadcastState() {
	r.connMu.RLock()
	defer r.connMu.RUnlock()
	state := r.Game.CopyState()
	msg := Message{Type: "STATE_UPDATE", Payload: state}
	for _, pc := range r.connections {
		pc.mu.Lock()
		pc.conn.WriteJSON(msg)
		pc.mu.Unlock()
	}
}

// BroadcastPreview sends a preview state (queue applied virtually) to all connections.
func (r *Room) BroadcastPreview() {
	r.connMu.RLock()
	defer r.connMu.RUnlock()
	preview := r.Game.GeneratePreview()
	msg := Message{Type: "STATE_UPDATE", Payload: preview}
	for _, pc := range r.connections {
		pc.mu.Lock()
		pc.conn.WriteJSON(msg)
		pc.mu.Unlock()
	}
}
