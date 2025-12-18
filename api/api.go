package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/umarbektokyo/matetra-engine/engine"
	"github.com/umarbektokyo/matetra-engine/model"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PlayerPayload struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type CardPlayPayload struct {
	CardIndex int   `json:"card_index"`
	Inputs    []int `json:"inputs"`
	Permanent bool  `json:"permanent"`
}

type CardPlayReply struct {
	Success      bool             `json:"success"`
	Message      string           `json:"message"`
	NewGameState *model.GameState `json:"newGameState,omitempty"`
}

type PlayerConnection struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	PlayerID int
}

type API struct {
	Game        *engine.Game
	Connections map[int]*PlayerConnection
	nextConnID  int
}

func New(game *engine.Game) *API {
	return &API{
		Game:        game,
		Connections: make(map[int]*PlayerConnection),
	}
}

// websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// starts the server + endpoints
func (a *API) Start() {
	http.HandleFunc("/ws", a.handleWebSocket)

	log.Println("API running on :1729")
	log.Fatal(http.ListenAndServe(":1729", nil))
}

func (a *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}

	playerConn := &PlayerConnection{conn: conn, PlayerID: -1}
	connID := a.nextConnID
	a.Connections[connID] = playerConn
	a.nextConnID++

	log.Printf("client %d connected", connID)

	go a.readMessages(connID, playerConn)
}

func (a *API) readMessages(connID int, pc *PlayerConnection) {
	defer func() {
		pc.conn.Close()
		log.Printf("client %d disconnected", connID)
	}()

	for {
		var incomingMsg Message
		if err := pc.conn.ReadJSON(&incomingMsg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				return
			}
			log.Printf("read error for client %d, %v", connID, err)
			return
		}
		a.handleIncomingMessages(pc, incomingMsg)
	}
}

func (a *API) handleIncomingMessages(pc *PlayerConnection, msg Message) {
	switch msg.Type {
	case "ADD_PLAYER":
		var payload PlayerPayload
		payloadBytes, err := json.Marshal(msg.Payload)
		if err != nil {
			log.Printf("error marshalling payload: %v", err)
		}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			a.sendError(pc, "invalid player payload format")
			return
		}

		playerID, err := a.Game.AddPlayer(payload.Name, payload.Hash)
		if err != nil {
			a.sendError(pc, err.Error())
			return
		}

		pc.PlayerID = playerID

		a.sendResponse(pc, "PLAYER_ADDED", map[string]string{"name": payload.Name})
		a.BroadcastState()
	case "PLAY_CARD":
		a.handlePlayCard(pc, msg.Payload)
	case "PROCESS_NEXT_TURN":
		a.handleNextTurn(pc)
	case "ROLL_DICE":
		a.handleRollDice(pc)
	default:
		a.sendError(pc, "unknown message type: "+msg.Type)
	}
}

func (a *API) handlePlayCard(pc *PlayerConnection, payload interface{}) {
	if pc.PlayerID == -1 {
		a.sendCustomReply(pc, false, "player is not authenticated", nil)
		return
	}

	var cardPayload CardPlayPayload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		a.sendCustomReply(pc, false, "error parsing card play payload", nil)
		return
	}
	if err := json.Unmarshal(payloadBytes, &cardPayload); err != nil {
		a.sendCustomReply(pc, false, "invalid card play payload format", nil)
		return
	}

	resultState, err := a.Game.ProcessMove(
		pc.PlayerID,
		cardPayload.CardIndex,
		cardPayload.Inputs,
		cardPayload.Permanent,
	)

	if err != nil {
		a.sendCustomReply(pc, false, fmt.Sprintf("move failed: %v", err), nil)
		return
	}

	message := "non-pernament move previewed successfully"
	if cardPayload.Permanent {
		playerName := resultState.Players[pc.PlayerID].Name
		cardName := resultState.Cards[cardPayload.CardIndex].Name
		message = fmt.Sprintf("@%s used %s!", playerName, cardName)
		a.BroadcastReply(true, message, resultState)
		message = "permanent move recorded successfully"
	}

	a.sendCustomReply(pc, true, message, resultState)
}

func (a *API) sendCustomReply(pc *PlayerConnection, success bool, message string, state *model.GameState) {
	reply := CardPlayReply{
		Success:      success,
		Message:      message,
		NewGameState: state,
	}

	respMsg := Message{
		Type:    "PLAY_CARD_REPLY",
		Payload: reply,
	}

	pc.mu.Lock()
	if err := pc.conn.WriteJSON(respMsg); err != nil {
		log.Printf("error sending custom play card reply: %v", err)
	}
	pc.mu.Unlock()
}

func (a *API) BroadcastReply(success bool, message string, state *model.GameState) {
	reply := CardPlayReply{
		Success:      success,
		Message:      message,
		NewGameState: state,
	}

	respMsg := Message{
		Type:    "PLAY_CARD_REPLY",
		Payload: reply,
	}

	for _, pc := range a.Connections {
		pc.mu.Lock()
		if err := pc.conn.WriteJSON(respMsg); err != nil {
			log.Printf("error broadcasting reply: %v", err)
		}
		pc.mu.Unlock()
	}
}

func (a *API) BroadcastState() {
	stateMsg := Message{
		Type:    "STATE_UPDATE",
		Payload: a.Game.CopyState(),
	}
	for _, pc := range a.Connections {
		pc.mu.Lock()
		if err := pc.conn.WriteJSON(stateMsg); err != nil {
			log.Printf("error broadcasting state: %v", err)
		}
		pc.mu.Unlock()
	}
}

func (a *API) sendResponse(pc *PlayerConnection, responseType string, data interface{}) {
	respMsg := Message{
		Type:    responseType,
		Payload: data,
	}
	pc.mu.Lock()
	if err := pc.conn.WriteJSON(respMsg); err != nil {
		log.Printf("error sending response: %v", err)
	}
	pc.mu.Unlock()
}

func (a *API) sendError(pc *PlayerConnection, errMsg string) {
	errorMsg := Message{
		Type: "ERROR",
		Payload: map[string]string{
			"message": errMsg,
		},
	}
	pc.mu.Lock()
	if err := pc.conn.WriteJSON(errorMsg); err != nil {
		log.Printf("error sending error: %v", err)
	}
	pc.mu.Unlock()
}

func (a *API) handleNextTurn(pc *PlayerConnection) {
	if pc.PlayerID == -1 {
		a.sendCustomReply(pc, false, "player is not authenticated", nil)
		return
	}

	resultState, err := a.Game.ProcessNextTurn(pc.PlayerID)
	if err != nil {
		a.sendCustomReply(pc, false, fmt.Sprintf("failed to end the turn: %v", err), nil)
		return
	}

	message := fmt.Sprintf("player @%s has ended their turn.", resultState.Players[pc.PlayerID].Name)
	if resultState.Turn != a.Game.State.Turn {
		message = fmt.Sprintf("turn finished! started turn %d. current player is @%s", resultState.Turn, resultState.Players[resultState.Turn%len(resultState.Players)].Name)
	}
	a.BroadcastReply(true, message, resultState)
}

func (a *API) handleRollDice(pc *PlayerConnection) {
	if pc.PlayerID == -1 {
		a.sendError(pc, "not authenticated")
		return
	}

	resultState, err := a.Game.ProcessDiceRoll(pc.PlayerID)
	if err != nil {
		a.sendCustomReply(pc, false, fmt.Sprintf("Dice roll failed: %v", err), nil)
		return
	}

	playerName := resultState.Players[pc.PlayerID].Name
	message := fmt.Sprintf("@%s rolled the dice!", playerName)

	a.BroadcastReply(true, message, resultState)
}
