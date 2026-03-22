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
