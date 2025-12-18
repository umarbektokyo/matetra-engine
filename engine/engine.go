package engine

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"github.com/umarbektokyo/matetra-engine/cards"
	"github.com/umarbektokyo/matetra-engine/cards/constants"
	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

type Game struct {
	State *model.GameState
	mu    sync.RWMutex
}

// Initializes a new empty game
func New(gameID string) *Game {
	return &Game{
		State: &model.GameState{
			GameID:  gameID,
			Players: []model.Player{},
			Cards:   []model.Card{},
			Numbers: make([][5]model.Number, 0),
			Done:    make([]bool, 0),
			Queue:   make([]int, 0),
			Turn:    0,
		},
	}
}

func NewNumber() model.Number {
	return model.Number{
		Value: big.NewFloat(0),
		Mark:  "n",
	}
}

func NewNumberRow() (row [5]model.Number) {
	for i := range row {
		row[i] = NewNumber()
	}
	return
}

// Adds a new player to the game
func (g *Game) AddPlayer(name, hash string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Adds a new player object
	playerID := len(g.State.Players)
	g.State.Players = append(g.State.Players, model.Player{
		Name: name,
		Hash: hash,
	})
	g.State.Numbers = append(g.State.Numbers, NewNumberRow())
	g.State.Done = append(g.State.Done, false)
	return playerID, nil
}

// Return the index of the player whoose turn it is
func (g *Game) CurrentPlayer() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	n := len(g.State.Players)
	if n == 0 {
		return -1
	}
	return g.State.Turn % n
}

// Loads the card deck into game state
func (g *Game) LoadCards() {
	g.mu.Lock()
	defer g.mu.Unlock()

	deck := utils.Must(cards.LoadCardsFromCSV(utils.DECK_PATH))
	g.State.Cards = append(g.State.Cards, deck...)
}

// Check if everyone has finished the turn
func (g *Game) TurnsFinished() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, done := range g.State.Done {
		if !done {
			return false
		}
	}
	return true
}

// Checks how many cards does the player have
func (g *Game) PlayerHandCount(player int) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, card := range g.State.Cards {
		if card.Owner == player {
			count++
		}
	}
	return count
}

// Internal version (no lock)
func (g *Game) restockCards() {
	for p := range g.State.Players {
		handCount := 0
		for _, card := range g.State.Cards {
			if card.Owner == p {
				handCount++
			}
		}

		for handCount < 6 {
			// Build a deck
			deck := []int{}
			for i, c := range g.State.Cards {
				if c.Owner == -1 {
					deck = append(deck, i)
				}
			}

			// If deck is empty, recycle used cards and build a deck
			if len(deck) == 0 {
				for i := range g.State.Cards {
					if g.State.Cards[i].Owner == -2 {
						g.State.Cards[i].Owner = -1
					}
				}

				for i, c := range g.State.Cards {
					if c.Owner == -1 {
						deck = append(deck, i)
					}
				}
			}

			if len(deck) == 0 {
				break
			}
			// Choose a card from a deck
			idx := deck[rand.Intn(len(deck))]
			g.State.Cards[idx].Owner = p
			handCount++
		}
	}
}

// Fills everyone's hands up (6 cards max) (needs optimisation)
func (g *Game) RestockCards() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.restockCards()
}

// Internal version (no lock)
func (g *Game) copyState() *model.GameState {
	virtual := &model.GameState{
		GameID:  g.State.GameID,
		Players: append([]model.Player(nil), g.State.Players...),
		Cards:   append([]model.Card(nil), g.State.Cards...),
		Numbers: make([][5]model.Number, len(g.State.Numbers)),
		Done:    append([]bool(nil), g.State.Done...),
		Queue:   append([]int(nil), g.State.Queue...),
		Turn:    g.State.Turn,
	}

	for i := range g.State.Numbers {
		for j := 0; j < 5; j++ {
			orig := g.State.Numbers[i][j]
			virtual.Numbers[i][j] = model.Number{
				Mark:  orig.Mark,
				Value: new(big.Float).Set(orig.Value),
			}
		}
	}

	return virtual
}

// Makes a virtual deep copy of the game state
func (g *Game) CopyState() *model.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.copyState()
}

// Applies a singular card
func (g *Game) ApplyCard(vgs *model.GameState, cardIndex int) error {
	err := cards.CardFunction(vgs, cardIndex)
	if err != nil {
		return err
	}

	// Remove card after applying
	vgs.Cards[cardIndex].Owner = -2
	vgs.Cards[cardIndex].Inputs = nil

	return nil
}

// Applies all the Cards in Queue
func (g *Game) ApplyCards(vgs *model.GameState) error {
	for _, cardIndex := range vgs.Queue {
		err := g.ApplyCard(vgs, cardIndex)
		if err != nil {
			return err
		}
	}
	vgs.Queue = nil

	return nil
}

// Internal version (no lock)
func (g *Game) nextTurn() error {
	virtual := g.copyState()

	virtual.Queue = nil

	g.State = virtual
	g.restockCards()

	// cleanup card marks
	for i := range g.State.Numbers {
		for j := range g.State.Numbers[i] {
			// Un-immune
			if g.State.Numbers[i][j].Mark == "I" {
				g.State.Numbers[i][j].Mark = ""
			}

			// Fibonacci
			if g.State.Numbers[i][j].Mark == "F" && g.State.Numbers[i][j].Value != nil {
				val, _ := g.State.Numbers[i][j].Value.Float64()
				v := int64(val)

				a, b := int64(1), int64(2)
				if v == 1 {
					g.State.Numbers[i][j].Value = big.NewFloat(2)
				} else {
					for b <= v {
						if b == v {
							g.State.Numbers[i][j].Value = big.NewFloat(float64(a + b))
							break
						}
						a, b = b, a+b
					}
				}
			}
		}
	}

	g.State.Turn++
	for i := range g.State.Done {
		g.State.Done[i] = false
	}

	return nil
}

// checks if the player is moving their own card
func (g *Game) PlayerCanPlayCard(playerID, cardIndex int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if cardIndex < 0 || cardIndex >= len(g.State.Cards) {
		return false
	}

	return g.State.Cards[cardIndex].Owner == playerID
}

// API: Moves
func (g *Game) ProcessMove(playerID int, cardIndex int, inputs []int, permanent bool) (*model.GameState, error) {
	g.mu.RLock()

	// check ownership
	if !g.PlayerCanPlayCard(playerID, cardIndex) {
		g.mu.RUnlock()
		return nil, fmt.Errorf("you do not own this card")
	}

	// validate input
	expected := len(g.State.Cards[cardIndex].InputsReq)
	if len(inputs) != expected {
		g.mu.RUnlock()
		return nil, fmt.Errorf("expected %d inputs but got %d", expected, len(inputs))
	}

	g.mu.RUnlock()

	// Virtual state for preview/calculation
	virtual := g.CopyState()

	// Apply the specific move to the virtual state (queue it)
	vCard := &virtual.Cards[cardIndex]
	vCard.Inputs = append([]int(nil), inputs...)
	virtual.Queue = append(virtual.Queue, cardIndex)

	// Execute ALL queued cards on the virtual state to get the final result
	fmt.Printf("[DEBUG] ProcessMove: Applying queue on virtual state. Queue len: %d\n", len(virtual.Queue))
	if err := g.ApplyCards(virtual); err != nil {
		return nil, fmt.Errorf("calculation failed: %v", err)
	}
	fmt.Printf("[DEBUG] ProcessMove: Applied. Virtual Numbers[0]: %v\n", virtual.Numbers[0])

	if !permanent {
		// Non-permanent: just return the virtual calculated state
		return virtual, nil
	} else {
		// Permanent: Update valid state AND return virtual calculated state
		g.mu.Lock()
		defer g.mu.Unlock()

		if g.State.Done[playerID] {
			return nil, fmt.Errorf("you have already finished your turn")
		}

		// Validation on live state
		originalInputs := g.State.Cards[cardIndex].Inputs
		g.State.Cards[cardIndex].Inputs = append([]int(nil), inputs...)

		if err := utils.ValidateInputs(g.State, &g.State.Cards[cardIndex]); err != nil {
			g.State.Cards[cardIndex].Inputs = originalInputs
			return nil, fmt.Errorf("invalid inputs: %v", err)
		}

		// Queue in real state
		g.State.Queue = append(g.State.Queue, cardIndex)

		// Return the VIRTUAL state (which has the queue applied) for display
		return virtual, nil
	}
}

// API: Turns
func (g *Game) ProcessNextTurn(playerID int) (*model.GameState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State.Done[playerID] {
		return nil, fmt.Errorf("you have already finished your turn")
	}

	g.State.Done[playerID] = true

	finished := true
	for _, done := range g.State.Done {
		if !done {
			finished = false
			break
		}
	}

	if finished {
		// execute all queued cards
		if err := g.ApplyCards(g.State); err != nil {
			// If application fails at this stage, it's problematic because turn is "done".
			// But for now, returning error is all we can do.
			// Ideally we should have validated everything perfectly before.
			return nil, fmt.Errorf("failed to apply queued cards: %v", err)
		}

		if err := g.nextTurn(); err != nil {
			return nil, err
		}
	}

	return g.copyState(), nil
}

// API: Dice
func (g *Game) ProcessDiceRoll(playerID int) (*model.GameState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. check if it is the player's turn
	if len(g.State.Players) == 0 || playerID != (g.State.Turn%len(g.State.Players)) {
		return nil, fmt.Errorf("it is not your turn")
	}

	// 2. Dice can only be rolled if the queue is empty
	if len(g.State.Queue) > 0 {
		return nil, fmt.Errorf("cannot roll dice after playing cards")
	}

	// 3. Find the first empty slot in the ACTUAL state
	firstEmptySlot := -1
	for i, num := range g.State.Numbers[playerID] {
		if num.Mark == "n" {
			firstEmptySlot = i
			break
		}
	}

	if firstEmptySlot == -1 {
		return nil, fmt.Errorf("no empty slots available for dice roll")
	}

	// 4. Apply dice directly to the ACTUAL state (permanent)
	err := constants.DICEAtSlot(g.State, playerID, firstEmptySlot)
	if err != nil {
		return nil, err
	}

	// Return a copy of the updated state
	return g.copyState(), nil
}
