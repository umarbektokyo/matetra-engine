package constants

import (
	"fmt"
	"math"
	"math/big"

	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

func AddConstant(vgs *model.GameState, player int, value *big.Float, mark string) error {
	fmt.Printf("[DEBUG] AddConstant: Adding %s to player %d\n", value.Text('g', 10), player)
	for i := range vgs.Numbers[player] {
		// prefer an empty slot
		if vgs.Numbers[player][i].Mark == "n" {
			fmt.Printf("[DEBUG] AddConstant: Found empty slot at %d\n", i)
			a := &vgs.Numbers[player][i]
			a.Value = value
			a.Mark = mark
			return nil
		}
	}

	// resort to replacing smallest value
	minIdx := 0
	for i := 1; i < len(vgs.Numbers[player]); i++ {
		if vgs.Numbers[player][i].Value.Cmp(vgs.Numbers[player][minIdx].Value) < 0 {
			minIdx = i
		}
	}

	a := &vgs.Numbers[player][minIdx]
	a.Value = value
	a.Mark = mark

	return nil
}

func DICE(vgs *model.GameState, player int) error {
	return AddConstant(vgs, player, big.NewFloat(float64(utils.RollDice(6))), "")
}

func DICEAtSlot(vgs *model.GameState, player int, slotIndex int) error {
	if slotIndex < 0 || slotIndex >= len(vgs.Numbers[player]) {
		return fmt.Errorf("invalid slot index: %d", slotIndex)
	}

	diceValue := big.NewFloat(float64(utils.RollDice(6)))
	fmt.Printf("[DEBUG] DICEAtSlot: Rolling dice for player %d at slot %d, value: %s\n",
		player, slotIndex, diceValue.Text('g', 10))

	vgs.Numbers[player][slotIndex].Value = diceValue
	vgs.Numbers[player][slotIndex].Mark = ""

	return nil
}

func CONSTPI(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(math.Pi), "")
}

func CONSTE(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(math.E), "")
}

func CONSTN1(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(-1), "")
}

func CONST73(vgs *model.GameState, card *model.Card) error {
	value := big.NewFloat(73)
	exists73 := false

	for _, num := range vgs.Numbers[card.Owner] {
		if num.Value != nil && num.Value.Cmp(value) == 0 {
			exists73 = true
			break
		}
	}

	if !exists73 {
		value = big.NewFloat(12)
	}

	return AddConstant(vgs, card.Owner, value, "")
}

func CONSTGOOGLE(vgs *model.GameState, card *model.Card) error {
	attackedPlayer := card.Inputs[0]
	attackedIndex := card.Inputs[1]

	ten := big.NewFloat(10)

	if attackedPlayer >= 0 &&
		attackedPlayer < len(vgs.Numbers) &&
		attackedIndex >= 0 &&
		attackedIndex < len(vgs.Numbers[attackedPlayer]) {

		num := &vgs.Numbers[attackedPlayer][attackedIndex]

		if num.Value != nil {
			q := new(big.Float).Quo(num.Value, ten)
			if q.IsInt() {
				// steal
				_ = AddConstant(vgs, card.Owner, new(big.Float).Set(num.Value), "")
				num.Value = big.NewFloat(0)
				num.Mark = "n"
				return nil
			}
		}
	}

	// default behavior
	return AddConstant(vgs, card.Owner, ten, "")
}

func CONST42(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(42), "")
}

func CONSTPHI(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(math.Phi), "")
}

func CONSTZERO(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(0), "")
}

func CONST7(vgs *model.GameState, card *model.Card) error {
	if err := AddConstant(vgs, card.Owner, big.NewFloat(7), ""); err != nil {
		return err
	}

	r1, r2 := utils.RollDice(6), utils.RollDice(6)
	if r1+r2 == 7 {
		return AddConstant(vgs, card.Owner, big.NewFloat(7), "")
	}
	return nil
}

func CONST26(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(26), "")
}

func CONST6(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(6), "")
}

func CONSTFIBONACCI(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(1), "F")
}

func CONST69(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(69), "")
}

func CONSTTAU(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(math.Pi*2), "")
}

func CONSTTENPOWER(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(math.Pow(10, float64(utils.RollDice(6)))), "")
}

func CONSTGRAHAM(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, big.NewFloat(9), "")
}

func CONSTCUPID(vgs *model.GameState, card *model.Card) error {
	roll1, roll2 := utils.RollDice(6), utils.RollDice(6)
	if roll1 <= 3 && roll2 <= 3 {
		return AddConstant(vgs, card.Owner, big.NewFloat(29), "")
	}
	return AddConstant(vgs, card.Owner, big.NewFloat(14), "")
}

func FACTORIAL(vgs *model.GameState, card *model.Card) error {
	dice := utils.RollDice(6)
	result := big.NewInt(1)
	for i := int64(2); i <= int64(dice); i++ {
		result.Mul(result, big.NewInt(i))
	}

	return AddConstant(vgs, card.Owner, new(big.Float).SetInt(result), "")
}
