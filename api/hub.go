package api

import (
	"math/rand"
	"sync"
)

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex
	Users *UserStore
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
		Users: NewUserStore(),
	}
}

func (h *Hub) CreateRoom(settings CreateRoomPayload) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	code := h.generateCode()
	room := newRoom(code, settings)
	h.rooms[code] = room
	return room
}

func (h *Hub) GetRoom(code string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[code]
}

func (h *Hub) RemoveRoom(code string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, code)
}

// FindRoomForPlayer searches all rooms for a player with the given name+hash.
// Returns the room and player ID if found, nil otherwise.
func (h *Hub) FindRoomForPlayer(name, hash string) (*Room, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, room := range h.rooms {
		for i, p := range room.Game.State.Players {
			if p.Name == name && p.Hash == hash {
				return room, i
			}
		}
	}
	return nil, -1
}

func (h *Hub) generateCode() string {
	for {
		b := make([]byte, 6)
		for i := range b {
			b[i] = codeChars[rand.Intn(len(codeChars))]
		}
		code := string(b)
		if _, exists := h.rooms[code]; !exists {
			return code
		}
	}
}
