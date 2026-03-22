package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/umarbektokyo/matetra-engine/model"
)

func Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func ValidateInputs(vgs *model.GameState, card *model.Card) error {
	if len(card.Inputs) != len(card.InputsReq) {
		return fmt.Errorf(
			"%s expects %d inputs, got %d",
			card.Method, len(card.InputsReq), len(card.Inputs),
		)
	}

	for i := range card.InputsReq {
		val := card.Inputs[i]
		switch card.InputsReq[i] {
		case 'd':
			if val < 1 || val > 6 {
				return fmt.Errorf("input %d must be dice (1..6), got %v", i, val)
			}

		case 'p':
			if val < 0 || val >= len(vgs.Players) {
				return fmt.Errorf("input %d must be player index, got %v", i, val)
			}

		case 'U':
			if val < 0 || val >= len(vgs.Players) {
				return fmt.Errorf("input %d must be player index, got %v", i, val)
			}
			if val != card.Owner {
				return fmt.Errorf("input %d must be your own index (%v), got %v", i, card.Owner, val)
			}

		case 'A':
			if val < 0 || val >= len(vgs.Players) {
				return fmt.Errorf("input %d must be player index, got %v", i, val)
			}
			if val != (vgs.Turn % len(vgs.Players)) {
				return fmt.Errorf("input %d must be index of defending player, got %v", i, val)
			}

		case 'n':
			if val < 0 || val > 4 {
				return fmt.Errorf("input %d must be number index 0..4, got %v", i, val)
			}

		case 'c':
			// Card index — validated elsewhere
		case 'X', 'Y':
			// Range bounds — just integers
		case 'i':
			X := card.Inputs[i-2]
			Y := card.Inputs[i-1]
			if val < X || val > Y {
				return fmt.Errorf("input %d must be in range of %d..%d, got %d", i, X, Y, val)
			}
		}
	}

	// Check for immunity
	for i := 0; i < len(card.InputsReq); i++ {
		if card.InputsReq[i] == 'n' {
			player := card.Inputs[i-1]
			index := card.Inputs[i]
			if vgs.Numbers[player][index].Mark == "I" {
				return fmt.Errorf("number %d of player %d is immune this turn", index, player)
			}
		}
	}

	return nil
}

func CheckCardMark(vgs *model.GameState, playerIndex int, numberIndex int) error {
	if vgs.Numbers[playerIndex][numberIndex].Mark == "n" {
		return fmt.Errorf("cannot use null number")
	}
	return nil
}

func FindIsland(vgs *model.GameState, player int, index int) (int, int, error) {
	if index < 0 || index >= len(vgs.Numbers[player]) {
		return 0, 0, fmt.Errorf("index out of range")
	}

	if vgs.Numbers[player][index].Mark == "n" {
		return 0, 0, fmt.Errorf("selected number is null")
	}

	nums := vgs.Numbers[player]

	L := index
	for L > 0 && nums[L-1].Mark != "n" {
		L--
	}

	R := index
	for R < len(nums)-1 && nums[R+1].Mark != "n" {
		R++
	}

	return L, R, nil
}

func PrimeFactors(n int64) []int64 {
	if n < 2 {
		return nil
	}
	var factors []int64
	x := n

	for x%2 == 0 {
		factors = append(factors, 2)
		x /= 2
	}

	for d := int64(3); d*d <= x; d += 2 {
		for x%d == 0 {
			factors = append(factors, d)
			x /= d
		}
	}

	if x > 1 {
		factors = append(factors, x)
	}

	return factors
}

func NextFibonacci(v float64) float64 {
	if v <= 0 {
		return 1
	}
	if v == 1 {
		return 2
	}
	a, b := 1.0, 2.0
	for b <= v {
		if b == v {
			return a + b
		}
		a, b = b, a+b
	}
	return v + 1
}

