package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload any `json:"payload"`
}

type AuthPayload struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type CreateRoomPayload struct {
	WinCondition int `json:"winCondition"` // 0=limit, 1=largest, 2=smallest
	TurnLimit    int `json:"turnLimit"`
	HandSize     int `json:"handSize"`
}

type JoinPayload struct {
	Code string `json:"code"`
}

type CardPlayPayload struct {
	CardIndex int   `json:"cardIndex"`
	Inputs    []int `json:"inputs"`
	Permanent bool  `json:"permanent"`
}

type DiceRollPayload struct {
	Values []int `json:"values"` // dice values from client
	Single bool  `json:"single"` // true = fill one slot, false = fill all
}

type CardPlayReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Start(hub *Hub, port string) {
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		handleAuth(w, r, hub)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(w, r, hub)
	})

	log.Printf("matetra server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleAuth(w http.ResponseWriter, r *http.Request, hub *Hub) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AuthPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.Name == "" || payload.Password == "" {
		http.Error(w, "name and password required", http.StatusBadRequest)
		return
	}

	token, err := hub.Users.Register(payload.Name, payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Name: payload.Name})
}

func handleWS(w http.ResponseWriter, r *http.Request, hub *Hub) {
	// Extract JWT from query param or Authorization header
	token := r.URL.Query().Get("token")
	if token == "" {
		auth := r.Header.Get("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}

	if token == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	username, err := validateJWT(token)
	if err != nil {
		http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Get user hash for game operations
	hash, ok := hub.Users.GetHash(username)
	if !ok {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v", err)
		return
	}

	log.Printf("WebSocket connected: @%s", username)

	// Check if user was already in a room (auto-rejoin on reconnect)
	if room, pid := hub.FindRoomForPlayer(username, hash); room != nil {
		log.Printf("@%s auto-rejoining room %s as player %d", username, room.Code, pid)
		conn.WriteJSON(Message{
			Type: "AUTO_REJOINED",
			Payload: map[string]any{
				"code":     room.Code,
				"playerID": pid,
			},
		})
		room.AddConnection(conn, hub, username, hash)
		return
	}

	// Send welcome message
	conn.WriteJSON(Message{
		Type: "WELCOME",
		Payload: map[string]string{
			"name":    username,
			"message": "Connected! Send CREATE_ROOM or JOIN_ROOM.",
		},
	})

	// Enter room selection loop
	handleRoomSelection(conn, hub, username, hash)
}

func handleRoomSelection(conn *websocket.Conn, hub *Hub, username, hash string) {
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			conn.Close()
			return
		}

		switch msg.Type {
		case "CREATE_ROOM":
			var p CreateRoomPayload
			if err := decodePayload(msg.Payload, &p); err != nil {
				writeError(conn, "invalid create room payload")
				continue
			}
			room := hub.CreateRoom(p)
			room.hub = hub
			conn.WriteJSON(Message{
				Type:    "ROOM_CREATED",
				Payload: map[string]string{"code": room.Code},
			})
			room.AddConnection(conn, hub, username, hash)
			return

		case "JOIN_ROOM":
			var p JoinPayload
			if err := decodePayload(msg.Payload, &p); err != nil {
				writeError(conn, "invalid join payload")
				continue
			}
			room := hub.GetRoom(p.Code)
			if room == nil {
				writeError(conn, "room not found: "+p.Code)
				continue
			}
			conn.WriteJSON(Message{
				Type:    "JOINED",
				Payload: map[string]string{"code": room.Code},
			})
			room.AddConnection(conn, hub, username, hash)
			return

		default:
			writeError(conn, "expected CREATE_ROOM or JOIN_ROOM")
		}
	}
}

func writeError(conn *websocket.Conn, msg string) {
	conn.WriteJSON(Message{Type: "ERROR", Payload: map[string]string{"message": msg}})
}

func decodePayload(payload any, target any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}
