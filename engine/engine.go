package engine

import (
	"fmt"
	"math/rand"
	"sort"
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

func New(gameID string, settings model.GameSettings) *Game {
	if settings.HandSize <= 0 {
		settings.HandSize = 6
	}
	if settings.WinCondition == model.WinLimit && settings.Limit == 0 {
		settings.Limit = 1000
	}
	return &Game{
		State: &model.GameState{
			GameID:   gameID,
			Players:  []model.Player{},
			Cards:    []model.Card{},
			Numbers:  make([][5]model.Number, 0),
			Done:     make([]bool, 0),
			DiceUsed: make([]bool, 0),
			Queue:    make([]int, 0),
			Turn:     0,
			Started:  false,
			Finished: false,
			Winner:   -1,
			Settings: settings,
			Log:      make([]string, 0),
		},
	}
}

func NewNullNumber() model.Number {
	return model.Number{Value: 0, Base: 0, Mark: "n"}
}

func NewNumberRow() (row [5]model.Number) {
	for i := range row {
		row[i] = NewNullNumber()
	}
	return
}

func (g *Game) AddPlayer(name, hash string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State.Started {
		return -1, fmt.Errorf("game already started")
	}

	for _, p := range g.State.Players {
		if p.Name == name {
			return -1, fmt.Errorf("player name '%s' already taken", name)
		}
	}

	playerID := len(g.State.Players)
	g.State.Players = append(g.State.Players, model.Player{
		Name: name, Hash: hash, Online: true, Points: 0,
	})
	g.State.Numbers = append(g.State.Numbers, NewNumberRow())
	g.State.Done = append(g.State.Done, false)
	g.State.DiceUsed = append(g.State.DiceUsed, false)
	return playerID, nil
}

func (g *Game) Reconnect(name, hash string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, p := range g.State.Players {
		if p.Name == name && p.Hash == hash {
			g.State.Players[i].Online = true
			return i, nil
		}
	}
	return -1, fmt.Errorf("no matching player found")
}

func (g *Game) SetPlayerOffline(playerID int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if playerID >= 0 && playerID < len(g.State.Players) {
		g.State.Players[playerID].Online = false
	}
}

func (g *Game) LoadCards() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	deck, err := cards.LoadCards()
	if err != nil { return err }
	g.State.Cards = append(g.State.Cards, deck...)
	return nil
}

func (g *Game) StartGame() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State.Started {
		return fmt.Errorf("game already started")
	}
	if len(g.State.Players) < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}

	g.State.Started = true

	if len(g.State.Cards) == 0 {
		deck, err := cards.LoadCards()
		if err != nil { return fmt.Errorf("failed to load cards: %v", err) }
		g.State.Cards = append(g.State.Cards, deck...)
	}

	g.restockCards()

	for p := range g.State.Players {
		for s := range g.State.Numbers[p] {
			g.State.Numbers[p][s] = model.Number{
				Value: float64(rand.Intn(6) + 1), Base: 0, Mark: "",
			}
		}
	}

	g.addLog("Game started! It's @%s's turn.", g.State.Players[0].Name)
	return nil
}

func (g *Game) IsStarted() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.State.Started
}

func (g *Game) restockCards() {
	handSize := g.State.Settings.HandSize
	for p := range g.State.Players {
		handCount := 0
		for _, card := range g.State.Cards {
			if card.Owner == p {
				handCount++
			}
		}
		for handCount < handSize {
			deck := []int{}
			for i, c := range g.State.Cards {
				if c.Owner == -1 {
					deck = append(deck, i)
				}
			}
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
			idx := deck[rand.Intn(len(deck))]
			g.State.Cards[idx].Owner = p
			handCount++
		}
	}
}

func (g *Game) copyState() *model.GameState {
	virtual := &model.GameState{
		GameID: g.State.GameID,
		Players: make([]model.Player, len(g.State.Players)),
		Cards: make([]model.Card, len(g.State.Cards)),
		Numbers: make([][5]model.Number, len(g.State.Numbers)),
		Done: make([]bool, len(g.State.Done)),
		DiceUsed: make([]bool, len(g.State.DiceUsed)),
		Queue: make([]int, len(g.State.Queue)),
		Turn: g.State.Turn,
		Started: g.State.Started,
		Finished: g.State.Finished,
		Winner: g.State.Winner,
		Settings: g.State.Settings,
		Log: make([]string, len(g.State.Log)),
	}
	copy(virtual.Players, g.State.Players)
	copy(virtual.Cards, g.State.Cards)
	copy(virtual.Done, g.State.Done)
	copy(virtual.DiceUsed, g.State.DiceUsed)
	copy(virtual.Queue, g.State.Queue)
	copy(virtual.Log, g.State.Log)
	for i := range g.State.Numbers {
		virtual.Numbers[i] = g.State.Numbers[i]
	}
	return virtual
}

func (g *Game) CopyState() *model.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.copyState()
}

// GeneratePreview creates a virtual state where all queued cards have been applied,
// but preserves the queue list so players can see what's pending.
// This is what everyone sees during the turn.
func (g *Game) GeneratePreview() *model.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.generatePreview()
}

func (g *Game) generatePreview() *model.GameState {
	preview := g.copyState()
	savedQueue := make([]int, len(preview.Queue))
	copy(savedQueue, preview.Queue)

	// Apply all queued cards on the preview copy
	g.applyCards(preview)

	// Restore the queue so players can see what cards were played
	preview.Queue = savedQueue

	return preview
}

func (g *Game) applyCard(vgs *model.GameState, cardIndex int) error {
	err := cards.CardFunction(vgs, cardIndex)
	if err != nil {
		return err
	}
	vgs.Cards[cardIndex].Owner = -2
	vgs.Cards[cardIndex].Inputs = nil
	return nil
}

func (g *Game) applyCards(vgs *model.GameState) error {
	if len(vgs.Queue) == 0 {
		return nil
	}

	type queueEntry struct {
		cardIndex int
		order     int
	}
	entries := make([]queueEntry, len(vgs.Queue))
	for i, ci := range vgs.Queue {
		entries[i] = queueEntry{cardIndex: ci, order: i}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		pi := vgs.Cards[entries[i].cardIndex].Precedence
		pj := vgs.Cards[entries[j].cardIndex].Precedence
		if pi != pj {
			return pi < pj
		}
		return entries[i].order < entries[j].order
	})

	for _, entry := range entries {
		ci := entry.cardIndex
		owner := vgs.Cards[ci].Owner
		if err := g.applyCard(vgs, ci); err != nil {
			// Card failed at apply time (e.g. divide by zero, non-integer for FTA,
			// number became null or immune due to an earlier card).
			// Return the card to the player's hand instead of consuming it.
			vgs.Cards[ci].Owner = owner
			vgs.Cards[ci].Inputs = nil
			continue
		}
	}
	vgs.Queue = nil
	return nil
}

// QueueCard: validates and queues a card on the REAL state.
// The card stays in the player's hand on the real state until turn end.
// Returns error if invalid.
func (g *Game) QueueCard(playerID int, cardIndex int, inputs []int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.State.Started {
		return fmt.Errorf("game has not started yet")
	}
	if g.State.Finished {
		return fmt.Errorf("game is finished")
	}
	if g.State.Done[playerID] {
		return fmt.Errorf("you have already finished your turn")
	}

	// Check ownership
	if cardIndex < 0 || cardIndex >= len(g.State.Cards) || g.State.Cards[cardIndex].Owner != playerID {
		return fmt.Errorf("you do not own this card")
	}

	// Check card isn't already queued
	for _, qi := range g.State.Queue {
		if qi == cardIndex {
			return fmt.Errorf("card already queued")
		}
	}

	// Validate input count
	expected := len(g.State.Cards[cardIndex].InputsReq)
	if len(inputs) != expected {
		return fmt.Errorf("expected %d inputs but got %d", expected, len(inputs))
	}

	// Validate inputs
	g.State.Cards[cardIndex].Inputs = append([]int(nil), inputs...)
	if err := utils.ValidateInputs(g.State, &g.State.Cards[cardIndex]); err != nil {
		g.State.Cards[cardIndex].Inputs = nil
		return fmt.Errorf("invalid inputs: %v", err)
	}

	// Queue the card (inputs stay stored on the card, card stays owned by player)
	g.State.Queue = append(g.State.Queue, cardIndex)

	// Track stats
	g.State.Players[playerID].Stats.CardsPlayed++
	// Check if this card targets another player's number (attack)
	card := g.State.Cards[cardIndex]
	cp := g.State.Turn % len(g.State.Players)
	if playerID != cp {
		// This player is attacking the current player
		for i, c := range card.InputsReq {
			if c == 'A' && i+1 < len(card.InputsReq) && card.InputsReq[i+1] == 'n' {
				g.State.Players[playerID].Stats.AttacksDealt++
				break
			}
		}
	}

	g.addLog("@%s queued %s", g.State.Players[playerID].Name, card.Name)
	return nil
}

// ProcessNextTurn: player declares they're done. When ALL are done, apply cards for real.
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
		// NOW actually apply all queued cards on the real state.
		// applyCards skips individual card failures gracefully.
		g.applyCards(g.State)

		g.trackBiggestNumbers()

		if g.checkWinCondition() {
			return g.copyState(), nil
		}

		g.nextTurn()
	}

	return g.copyState(), nil
}

func (g *Game) ProcessDiceRoll(playerID int, diceValues []int) (*model.GameState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.State.Started {
		return nil, fmt.Errorf("game has not started yet")
	}
	if g.State.DiceUsed[playerID] {
		return nil, fmt.Errorf("you have already rolled dice this turn")
	}
	if g.State.Done[playerID] {
		return nil, fmt.Errorf("you have already finished your turn")
	}

	for _, d := range diceValues {
		if d < 1 || d > 6 {
			return nil, fmt.Errorf("invalid dice value: %d (must be 1-6)", d)
		}
	}

	emptySlots := []int{}
	for i, num := range g.State.Numbers[playerID] {
		if num.Mark == "n" {
			emptySlots = append(emptySlots, i)
		}
	}
	if len(emptySlots) == 0 {
		return nil, fmt.Errorf("no empty slots available")
	}

	filled := 0
	for i, slot := range emptySlots {
		if i >= len(diceValues) {
			break
		}
		constants.DICEAtSlot(g.State, playerID, slot, diceValues[i])
		filled++
	}

	g.State.DiceUsed[playerID] = true
	g.State.Players[playerID].Stats.DiceRolls++
	for i := 0; i < filled; i++ { g.State.Players[playerID].Stats.DiceTotal += diceValues[i] }
	g.addLog("@%s rolled dice (%d slots)", g.State.Players[playerID].Name, filled)
	return g.copyState(), nil
}

func (g *Game) ProcessDiceRollSingle(playerID int, diceValue int) (*model.GameState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.State.Started {
		return nil, fmt.Errorf("game has not started yet")
	}
	if g.State.DiceUsed[playerID] {
		return nil, fmt.Errorf("you have already rolled dice this turn")
	}
	if g.State.Done[playerID] {
		return nil, fmt.Errorf("you have already finished your turn")
	}
	if diceValue < 1 || diceValue > 6 {
		return nil, fmt.Errorf("invalid dice value: %d", diceValue)
	}

	slot := -1
	for i, num := range g.State.Numbers[playerID] {
		if num.Mark == "n" {
			slot = i
			break
		}
	}
	if slot == -1 {
		return nil, fmt.Errorf("no empty slots")
	}

	constants.DICEAtSlot(g.State, playerID, slot, diceValue)
	g.State.Players[playerID].Stats.DiceRolls++
	g.State.Players[playerID].Stats.DiceTotal += diceValue

	hasEmpty := false
	for _, num := range g.State.Numbers[playerID] {
		if num.Mark == "n" {
			hasEmpty = true
			break
		}
	}
	if !hasEmpty {
		g.State.DiceUsed[playerID] = true
	}

	return g.copyState(), nil
}

func (g *Game) FinishDiceRoll(playerID int) (*model.GameState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.State.DiceUsed[playerID] = true
	return g.copyState(), nil
}

func (g *Game) trackBiggestNumbers() {
	for p := range g.State.Players {
		for _, num := range g.State.Numbers[p] {
			if num.IsNull() { continue }
			n := model.Number{Value: num.Value, Base: num.Base}
			cur := model.Number{Value: g.State.Players[p].Stats.BiggestNumber, Base: g.State.Players[p].Stats.BiggestBase}
			if n.Cmp(cur) > 0 {
				g.State.Players[p].Stats.BiggestNumber = n.Value
				g.State.Players[p].Stats.BiggestBase = n.Base
			}
		}
	}
}

func (g *Game) nextTurn() {
	g.trackBiggestNumbers()
	g.restockCards()

	for i := range g.State.Numbers {
		for j := range g.State.Numbers[i] {
			mark := g.State.Numbers[i][j].Mark

			// Remove immunity, preserving Fibonacci if combined
			if mark == "I" {
				g.State.Numbers[i][j].Mark = ""
			} else if mark == "FI" {
				g.State.Numbers[i][j].Mark = "F"
			}

			// Grow Fibonacci numbers (including those just restored from "FI")
			if g.State.Numbers[i][j].Mark == "F" {
				val := g.State.Numbers[i][j].ToFloat64()
				next := utils.NextFibonacci(val)
				n := model.NumFromFloat(next)
				g.State.Numbers[i][j].Value = n.Value
				g.State.Numbers[i][j].Base = n.Base
			}
		}
	}

	g.State.Turn++
	for i := range g.State.Done {
		g.State.Done[i] = false
	}
	for i := range g.State.DiceUsed {
		g.State.DiceUsed[i] = false
	}
	g.State.Queue = nil

	// Auto-fill empty slots with random dice rolls
	for p := range g.State.Players {
		for s := range g.State.Numbers[p] {
			if g.State.Numbers[p][s].Mark == "n" {
				g.State.Numbers[p][s] = model.Number{
					Value: float64(rand.Intn(6) + 1),
					Base:  0,
					Mark:  "",
				}
			}
		}
	}

	cp := g.State.Turn % len(g.State.Players)
	g.addLog("Turn %d: @%s's turn", g.State.Turn, g.State.Players[cp].Name)
}

func (g *Game) checkWinCondition() bool {
	switch g.State.Settings.WinCondition {
	case model.WinLimit:
		return g.checkLimitWin()
	case model.WinLargest:
		return g.checkTurnLimitWin(true)
	case model.WinSmallest:
		return g.checkTurnLimitWin(false)
	}
	return false
}

func (g *Game) checkLimitWin() bool {
	limit := g.State.Settings.Limit
	if limit == 0 {
		limit = 1000
	}
	limitNum := model.NumFromFloat(limit)
	scored := false

	for p := range g.State.Players {
		for s := range g.State.Numbers[p] {
			num := &g.State.Numbers[p][s]
			if num.IsNull() {
				continue
			}
			n := model.Number{Value: num.Value, Base: num.Base}
			if n.IsPositive() && n.Cmp(limitNum) > 0 {
				g.State.Players[p].Points++
				g.addLog("@%s scored! %s > %s (%d pts)",
					g.State.Players[p].Name, n.Display(), limitNum.Display(), g.State.Players[p].Points)
				num.Value = 0
				num.Base = 0
				num.Mark = "n"
				scored = true
			}
		}
		if g.State.Players[p].Points >= 3 {
			g.State.Finished = true
			g.State.Winner = p
			g.addLog("@%s wins with %d points!", g.State.Players[p].Name, g.State.Players[p].Points)
			return true
		}
	}
	if scored {
		g.State.Settings.Limit = limit * 10
		g.addLog("Limit -> %s", model.NumFromFloat(g.State.Settings.Limit).Display())
	}
	return false
}

func (g *Game) checkTurnLimitWin(largest bool) bool {
	if g.State.Turn < g.State.Settings.TurnLimit-1 {
		return false
	}
	bestPlayer := -1
	var bestNum model.Number
	for p := range g.State.Players {
		for _, num := range g.State.Numbers[p] {
			if num.IsNull() {
				continue
			}
			n := model.Number{Value: num.Value, Base: num.Base}
			if bestPlayer == -1 {
				bestPlayer = p
				bestNum = n
				continue
			}
			cmp := n.Cmp(bestNum)
			if (largest && cmp > 0) || (!largest && cmp < 0) {
				bestPlayer = p
				bestNum = n
			}
		}
	}
	if bestPlayer >= 0 {
		g.State.Finished = true
		g.State.Winner = bestPlayer
		mode := "largest"
		if !largest {
			mode = "smallest"
		}
		g.addLog("Game over! @%s wins (%s: %s)", g.State.Players[bestPlayer].Name, mode, bestNum.Display())
	}
	return g.State.Finished
}

func (g *Game) addLog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	g.State.Log = append(g.State.Log, msg)
	if len(g.State.Log) > 50 {
		g.State.Log = g.State.Log[len(g.State.Log)-50:]
	}
}

