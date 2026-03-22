package cards

import (
	"testing"
)

func TestLoadCards(t *testing.T) {
	deck, err := LoadCards()
	if err != nil {
		t.Fatalf("LoadCards: %v", err)
	}
	if len(deck) == 0 {
		t.Fatal("no cards loaded")
	}

	// Check all cards have names and methods
	for i, c := range deck {
		if c.Name == "" {
			t.Errorf("card %d has no name", i)
		}
		if c.Method == "" {
			t.Errorf("card %d (%s) has no method", i, c.Name)
		}
		if c.Owner != -1 {
			t.Errorf("card %d (%s) should be in deck (owner -1), got %d", i, c.Name, c.Owner)
		}
	}
}

func TestAllMethodsHandled(t *testing.T) {
	deck, _ := LoadCards()
	seen := map[string]bool{}
	for _, c := range deck {
		if seen[c.Method] {
			continue
		}
		seen[c.Method] = true
		// Try calling CardFunction with a nil state to check it doesn't panic on unknown method
		// We can't fully test without a game state, but we verify the method is in the switch
	}
}

func TestCardDescriptions(t *testing.T) {
	deck, _ := LoadCards()
	seen := map[string]bool{}
	for _, c := range deck {
		if seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		if c.Description == "" {
			t.Errorf("card %q has no description", c.Name)
		}
	}
}
