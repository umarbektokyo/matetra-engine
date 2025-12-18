package model

import "math/big"

type Card struct {
	// ID          string // Unique Identifier for every card, even if a dublicate id is different
	Name        string
	Description string
	Type        string
	Method      string // Defines what method in code will be taken
	Owner       int    // -1: deck, -2: used, User.ID: owner
	Inputs      []int  // length depends on the card
	InputsReq   string // string with each character signifying input number type.
	// InputsReq explained:
	// d: dice (int)
	// p: player (int)
	// n: number (int)
	// c: card, id (string) -> doesn't work yet (we have to figure out something as we can't accept strings anymore)
	// U: makes next digit user's
	// A: Makes next digit attacked one
	// X: minimum for the input (int)
	// Y: maximum for the input (int)
	// i: allow input for a user (int)
}

// Only for authentication
type Player struct {
	Name string
	Hash string
}

type Number struct {
	Value *big.Float
	Mark  string
	// n: null
	// F: fibonacci
	// I: immune
}

// Main Game Object
type GameState struct {
	GameID  string
	Players []Player
	Cards   []Card
	Numbers [][5]Number
	Done    []bool
	Queue   []int // stores cardIndex and every time the move is finished, we apply all the cards and cleane the data in them, marking them as used.
	Turn    int   // total turns elapsed; current player = Turn % len(Players)
}
