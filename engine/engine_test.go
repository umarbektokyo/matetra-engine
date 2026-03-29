package engine

import (
	"testing"

	"github.com/umarbektokyo/matetra-engine/model"
)

func newTestGame() *Game {
	g := New("test", model.GameSettings{HandSize: 6, WinCondition: model.WinLimit, Limit: 1000})
	g.LoadCards()
	return g
}

// startedTestGame returns a 2-player game that's already started.
func startedTestGame() *Game {
	g := newTestGame()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	g.StartGame()
	return g
}

// findCardByMethod returns the index of the first card owned by the given player
// with the given method, or -1 if not found.
func findCardByMethod(g *Game, owner int, method string) int {
	for i, c := range g.State.Cards {
		if c.Owner == owner && c.Method == method {
			return i
		}
	}
	return -1
}

// giveCard assigns a card with the given method to a player. Returns the card index.
// If the player already has one, returns that. Otherwise reassigns from deck/used/other.
func giveCard(g *Game, playerID int, method string) int {
	// Already owned by this player
	for i, c := range g.State.Cards {
		if c.Owner == playerID && c.Method == method {
			return i
		}
	}
	// Try unowned deck
	for i, c := range g.State.Cards {
		if c.Owner == -1 && c.Method == method {
			g.State.Cards[i].Owner = playerID
			return i
		}
	}
	// Try used pile
	for i, c := range g.State.Cards {
		if c.Owner == -2 && c.Method == method {
			g.State.Cards[i].Owner = playerID
			return i
		}
	}
	// Steal from another player
	for i, c := range g.State.Cards {
		if c.Owner != playerID && c.Method == method {
			g.State.Cards[i].Owner = playerID
			return i
		}
	}
	return -1
}

func TestAddPlayer(t *testing.T) {
	g := newTestGame()
	id, err := g.AddPlayer("alice", "hash1")
	if err != nil || id != 0 {
		t.Fatalf("AddPlayer: id=%d err=%v", id, err)
	}
	id2, err := g.AddPlayer("bob", "hash2")
	if err != nil || id2 != 1 {
		t.Fatalf("AddPlayer bob: id=%d err=%v", id2, err)
	}
}

func TestAddDuplicatePlayer(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "hash1")
	_, err := g.AddPlayer("alice", "hash2")
	if err == nil {
		t.Error("expected duplicate name error")
	}
}

func TestStartGameNeedsTwoPlayers(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h")
	err := g.StartGame()
	if err == nil {
		t.Error("expected error starting with 1 player")
	}
}

func TestStartGame(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	err := g.StartGame()
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if !g.State.Started {
		t.Error("game should be started")
	}
	// Check cards were dealt
	aliceCards := 0
	for _, c := range g.State.Cards {
		if c.Owner == 0 { aliceCards++ }
	}
	if aliceCards != 6 {
		t.Errorf("alice has %d cards, want 6", aliceCards)
	}
	// Check numbers populated
	for _, num := range g.State.Numbers[0] {
		if num.IsNull() {
			t.Error("alice should have all number slots filled")
		}
	}
}

func TestReconnect(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "hash1")
	g.SetPlayerOffline(0)
	if g.State.Players[0].Online {
		t.Error("should be offline")
	}
	id, err := g.Reconnect("alice", "hash1")
	if err != nil || id != 0 {
		t.Fatalf("Reconnect: id=%d err=%v", id, err)
	}
	if !g.State.Players[0].Online {
		t.Error("should be online after reconnect")
	}
}

func TestReconnectBadHash(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "hash1")
	_, err := g.Reconnect("alice", "wrong")
	if err == nil {
		t.Error("expected error for wrong hash")
	}
}

func TestQueueCardNotStarted(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h")
	g.AddPlayer("bob", "h")
	err := g.QueueCard(0, 0, nil)
	if err == nil {
		t.Error("expected error queuing card before start")
	}
}

func TestProcessNextTurn(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	g.StartGame()

	// Both end turn
	g.ProcessNextTurn(0)
	gs, err := g.ProcessNextTurn(1)
	if err != nil {
		t.Fatalf("ProcessNextTurn: %v", err)
	}
	if gs.Turn != 1 {
		t.Errorf("turn should be 1, got %d", gs.Turn)
	}
}

func TestDiceRoll(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	g.StartGame()

	// Clear alice's slots to make room for dice
	for i := range g.State.Numbers[0] {
		g.State.Numbers[0][i] = NewNullNumber()
	}

	_, err := g.ProcessDiceRoll(0, []int{3, 4, 5, 2, 1})
	if err != nil {
		t.Fatalf("DiceRoll: %v", err)
	}
	// Check slots filled
	for i, num := range g.State.Numbers[0] {
		if num.IsNull() {
			t.Errorf("slot %d should be filled", i)
		}
	}
}

func TestDiceRollAlreadyUsed(t *testing.T) {
	g := newTestGame()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	g.StartGame()
	for i := range g.State.Numbers[0] { g.State.Numbers[0][i] = NewNullNumber() }

	g.ProcessDiceRoll(0, []int{1, 2, 3, 4, 5})
	_, err := g.ProcessDiceRoll(0, []int{6})
	if err == nil {
		t.Error("expected error on second dice roll")
	}
}

func TestWinConditionLimit(t *testing.T) {
	g := New("test", model.GameSettings{HandSize: 6, WinCondition: model.WinLimit, Limit: 10})
	g.LoadCards()
	g.AddPlayer("alice", "h1")
	g.AddPlayer("bob", "h2")
	g.StartGame()

	// Give alice a huge number to cross the limit
	g.State.Numbers[0][0] = model.Number{Value: 5, Base: 1, Mark: ""} // 50 > 10

	// End turn to trigger win check
	g.State.Done[0] = true
	g.ProcessNextTurn(1)

	if g.State.Players[0].Points != 1 {
		t.Errorf("alice should have 1 point, got %d", g.State.Players[0].Points)
	}
}

// --- Immunity tests ---

func TestImmunityBlocksCardsAtApplyTime(t *testing.T) {
	g := startedTestGame()

	// Turn 0: alice (player 0) is the defender.
	// Give alice an ELEMENTCLOSURE card and bob a NEGATIVE card.
	closureIdx := giveCard(g, 0, "ELEMENTCLOSURE")
	negIdx := giveCard(g, 1, "NEGATIVE")
	if closureIdx == -1 || negIdx == -1 {
		t.Fatal("could not find required cards")
	}

	// Alice grants immunity to her own slot 0
	// ELEMENTCLOSURE InputsReq = "An" → [defenderPlayerIdx, numberIdx]
	err := g.QueueCard(0, closureIdx, []int{0, 0})
	if err != nil {
		t.Fatalf("queue immunity: %v", err)
	}

	// Bob targets alice's slot 0 with NEGATIVE
	// NEGATIVE InputsReq = "An" → [defenderPlayerIdx, numberIdx]
	err = g.QueueCard(1, negIdx, []int{0, 0})
	if err != nil {
		t.Fatalf("queue negative: %v", err)
	}

	originalVal := g.State.Numbers[0][0].Value

	// Both end turn — this must NOT deadlock
	g.ProcessNextTurn(0)
	gs, err := g.ProcessNextTurn(1)
	if err != nil {
		t.Fatalf("ProcessNextTurn should not error: %v", err)
	}

	// Turn should have advanced
	if gs.Turn != 1 {
		t.Errorf("turn should be 1, got %d", gs.Turn)
	}

	// The NEGATIVE card should have been skipped (immunity protected the number),
	// so value should remain positive (same sign as original).
	// Note: auto-fill may have replaced the number if it became null, but immunity
	// should have prevented NEGATIVE from running at all.
	if g.State.Numbers[0][0].Value == -originalVal {
		t.Error("immunity should have blocked NEGATIVE from flipping the sign")
	}
}

func TestImmunityDoesntDeadlockGame(t *testing.T) {
	g := startedTestGame()

	closureIdx := giveCard(g, 0, "ELEMENTCLOSURE")
	addIdx := giveCard(g, 1, "ADD")
	if closureIdx == -1 || addIdx == -1 {
		t.Fatal("could not find required cards")
	}

	// Alice immunizes slot 0
	if err := g.QueueCard(0, closureIdx, []int{0, 0}); err != nil {
		t.Fatalf("queue immunity: %v", err)
	}

	// Bob tries to ADD into alice's slot 0 using his own slot 0
	// ADD InputsReq = "AnUn" → [targetPlayer, targetSlot, sourcePlayer, sourceSlot]
	if err := g.QueueCard(1, addIdx, []int{0, 0, 1, 0}); err != nil {
		t.Fatalf("queue add: %v", err)
	}

	// End both turns — must not return error
	if _, err := g.ProcessNextTurn(0); err != nil {
		t.Fatalf("alice end turn: %v", err)
	}
	gs, err := g.ProcessNextTurn(1)
	if err != nil {
		t.Fatalf("bob end turn should succeed: %v", err)
	}

	if gs.Turn != 1 {
		t.Errorf("turn should advance to 1, got %d", gs.Turn)
	}

	// Both players should be able to act next turn
	if gs.Done[0] || gs.Done[1] {
		t.Error("Done flags should be reset for new turn")
	}
}

func TestImmunityExpires(t *testing.T) {
	g := startedTestGame()

	// Manually set immunity on slot 0
	g.State.Numbers[0][0].Mark = "I"

	// Advance turn (both players end)
	g.ProcessNextTurn(0)
	g.ProcessNextTurn(1)

	// Immunity should be cleared
	if g.State.Numbers[0][0].Mark == "I" {
		t.Error("immunity should expire after one turn")
	}
}

// --- Fibonacci tests ---

func TestFibonacciGrows(t *testing.T) {
	g := startedTestGame()

	// Place a Fibonacci number (value=1) in alice's slot 0
	g.State.Numbers[0][0] = model.Number{Value: 1, Base: 0, Mark: "F"}

	// Advance turn
	g.ProcessNextTurn(0)
	g.ProcessNextTurn(1)

	num := g.State.Numbers[0][0]
	if num.Mark != "F" {
		t.Errorf("fibonacci mark should be preserved, got %q", num.Mark)
	}
	// Fibonacci: 1 → 2
	if num.Value != 2 || num.Base != 0 {
		t.Errorf("fibonacci should grow from 1 to 2, got %s", num.Display())
	}
}

func TestFibonacciGrowsMultipleTurns(t *testing.T) {
	g := startedTestGame()

	// Fibonacci sequence: 1, 2, 3, 5, 8, 13, ...
	g.State.Numbers[0][0] = model.Number{Value: 1, Base: 0, Mark: "F"}

	expected := []float64{2, 3, 5, 8, 13}
	for turn, exp := range expected {
		g.ProcessNextTurn(0)
		g.ProcessNextTurn(1)

		num := g.State.Numbers[0][0]
		got := num.ToFloat64()
		if got != exp {
			t.Errorf("turn %d: fibonacci should be %.0f, got %.0f", turn+1, exp, got)
		}
		if num.Mark != "F" {
			t.Errorf("turn %d: fibonacci mark lost, got %q", turn+1, num.Mark)
		}
	}
}

func TestFibonacciPreservedWithImmunity(t *testing.T) {
	g := startedTestGame()

	// Start with a Fibonacci number (value=2)
	g.State.Numbers[0][0] = model.Number{Value: 2, Base: 0, Mark: "F"}

	closureIdx := giveCard(g, 0, "ELEMENTCLOSURE")
	if closureIdx == -1 {
		t.Fatal("could not find ELEMENTCLOSURE card")
	}

	// Alice grants immunity to her Fibonacci number
	if err := g.QueueCard(0, closureIdx, []int{0, 0}); err != nil {
		t.Fatalf("queue immunity on fibonacci: %v", err)
	}

	// End turn to apply cards
	g.ProcessNextTurn(0)
	g.ProcessNextTurn(1)

	// After turn: immunity should be removed, but Fibonacci mark should be restored.
	// Fibonacci should also have grown: 2 → 3
	num := g.State.Numbers[0][0]
	if num.Mark != "F" {
		t.Errorf("fibonacci mark should be restored after immunity expires, got %q", num.Mark)
	}
	if num.ToFloat64() != 3 {
		t.Errorf("fibonacci should grow from 2 to 3, got %s", num.Display())
	}
}

func TestImmunityOnFibonacciSetsFIMark(t *testing.T) {
	g := startedTestGame()

	g.State.Numbers[0][0] = model.Number{Value: 1, Base: 0, Mark: "F"}

	closureIdx := giveCard(g, 0, "ELEMENTCLOSURE")
	if closureIdx == -1 {
		t.Fatal("could not find ELEMENTCLOSURE card")
	}

	if err := g.QueueCard(0, closureIdx, []int{0, 0}); err != nil {
		t.Fatalf("queue immunity: %v", err)
	}

	// Generate preview — immunity should be applied, mark should be "FI"
	preview := g.GeneratePreview()
	if preview.Numbers[0][0].Mark != "FI" {
		t.Errorf("preview should show FI mark, got %q", preview.Numbers[0][0].Mark)
	}
}

// --- Auto-fill empty slots tests ---

func TestAutoFillEmptySlots(t *testing.T) {
	g := startedTestGame()

	// Null out some of alice's slots
	g.State.Numbers[0][2] = NewNullNumber()
	g.State.Numbers[0][4] = NewNullNumber()

	// Advance turn
	g.ProcessNextTurn(0)
	g.ProcessNextTurn(1)

	// All slots should be filled
	for i, num := range g.State.Numbers[0] {
		if num.IsNull() {
			t.Errorf("slot %d should be auto-filled, still null", i)
		}
	}
}

func TestAutoFillDiceRange(t *testing.T) {
	g := startedTestGame()

	// Null out all of alice's slots
	for i := range g.State.Numbers[0] {
		g.State.Numbers[0][i] = NewNullNumber()
	}

	// Advance turn
	g.ProcessNextTurn(0)
	g.ProcessNextTurn(1)

	// All slots should be filled with values 1-6
	for i, num := range g.State.Numbers[0] {
		if num.IsNull() {
			t.Errorf("slot %d should be auto-filled", i)
		}
		v := num.ToFloat64()
		if v < 1 || v > 6 {
			t.Errorf("slot %d auto-fill value %.0f should be in range 1-6", i, v)
		}
	}
}

// --- Queue tests ---

func TestQueueNoSizeLimit(t *testing.T) {
	g := startedTestGame()

	// Give alice many cards (use POSITIVE which is a no-op, InputsReq = "An")
	queued := 0
	for i := range g.State.Cards {
		if g.State.Cards[i].Method == "POSITIVE" {
			g.State.Cards[i].Owner = 0
			err := g.QueueCard(0, i, []int{0, 0})
			if err != nil {
				continue
			}
			queued++
		}
	}

	if queued < 2 {
		t.Fatalf("expected to queue at least 2 POSITIVE cards, got %d", queued)
	}

	if len(g.State.Queue) != queued {
		t.Errorf("queue length %d != queued count %d", len(g.State.Queue), queued)
	}
}

func TestNoDeadlockWhenCardTargetsConsumedNumber(t *testing.T) {
	g := startedTestGame()

	// Turn 0: alice (player 0) is defender.
	// Set up known values so we can track what happens.
	g.State.Numbers[0][0] = model.Number{Value: 5, Base: 0, Mark: ""}
	g.State.Numbers[0][1] = model.Number{Value: 3, Base: 0, Mark: ""}

	// Alice queues ADD: target=slot0, source=slot1 → slot1 becomes null
	addIdx := giveCard(g, 0, "ADD")
	if addIdx == -1 {
		t.Fatal("could not find ADD card")
	}
	if err := g.QueueCard(0, addIdx, []int{0, 0, 0, 1}); err != nil {
		t.Fatalf("queue ADD: %v", err)
	}

	// Bob queues NEGATIVE on alice's slot 1 (which ADD will consume first)
	negIdx := giveCard(g, 1, "NEGATIVE")
	if negIdx == -1 {
		t.Fatal("could not find NEGATIVE card")
	}
	if err := g.QueueCard(1, negIdx, []int{0, 1}); err != nil {
		t.Fatalf("queue NEGATIVE: %v", err)
	}

	// Both end turn — must NOT deadlock
	if _, err := g.ProcessNextTurn(0); err != nil {
		t.Fatalf("alice end turn: %v", err)
	}
	gs, err := g.ProcessNextTurn(1)
	if err != nil {
		t.Fatalf("bob end turn should not deadlock: %v", err)
	}

	if gs.Turn != 1 {
		t.Errorf("turn should advance to 1, got %d", gs.Turn)
	}
	if gs.Done[0] || gs.Done[1] {
		t.Error("Done flags should be reset for new turn")
	}
}

func TestPlayerCanActAfterOtherEndsTurn(t *testing.T) {
	g := startedTestGame()

	// Alice ends her turn
	_, err := g.ProcessNextTurn(0)
	if err != nil {
		t.Fatalf("alice end turn: %v", err)
	}

	// Bob should still be able to queue cards
	cardIdx := giveCard(g, 1, "NEGATIVE")
	if cardIdx == -1 {
		t.Fatal("could not find card for bob")
	}
	// Target the defending player's (alice, player 0) slot 0
	err = g.QueueCard(1, cardIdx, []int{0, 0})
	if err != nil {
		t.Fatalf("bob should still be able to queue cards after alice ended turn: %v", err)
	}

	// Bob should still be able to roll dice
	for i := range g.State.Numbers[1] {
		g.State.Numbers[1][i] = NewNullNumber()
	}
	_, err = g.ProcessDiceRoll(1, []int{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("bob should still be able to roll dice: %v", err)
	}

	// Now bob ends turn — turn should advance
	gs, err := g.ProcessNextTurn(1)
	if err != nil {
		t.Fatalf("bob end turn: %v", err)
	}
	if gs.Turn != 1 {
		t.Errorf("turn should advance to 1, got %d", gs.Turn)
	}
}

func TestQueueCardAlreadyQueued(t *testing.T) {
	g := startedTestGame()

	cardIdx := findCardByMethod(g, 0, "")
	// Find any card owned by alice
	for i, c := range g.State.Cards {
		if c.Owner == 0 && c.Method == "POSITIVE" {
			cardIdx = i
			break
		}
	}
	if cardIdx == -1 {
		// Give alice a POSITIVE card
		cardIdx = giveCard(g, 0, "POSITIVE")
	}
	if cardIdx == -1 {
		t.Skip("no POSITIVE card available")
	}

	err := g.QueueCard(0, cardIdx, []int{0, 0})
	if err != nil {
		t.Fatalf("first queue: %v", err)
	}

	// Queuing the same card again should fail
	err = g.QueueCard(0, cardIdx, []int{0, 0})
	if err == nil {
		t.Error("expected error when queuing same card twice")
	}
}
