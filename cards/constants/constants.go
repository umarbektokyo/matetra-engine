package constants

import (
	"fmt"
	"math"

	"github.com/umarbektokyo/matetra-engine/model"
)

func AddConstant(vgs *model.GameState, player int, value model.Number) error {
	for i := range vgs.Numbers[player] {
		if vgs.Numbers[player][i].Mark == "n" {
			vgs.Numbers[player][i] = value
			return nil
		}
	}

	// Replace smallest value
	minIdx := 0
	for i := 1; i < len(vgs.Numbers[player]); i++ {
		if vgs.Numbers[player][i].Cmp(vgs.Numbers[player][minIdx]) < 0 {
			minIdx = i
		}
	}

	vgs.Numbers[player][minIdx] = value
	return nil
}

func DICEAtSlot(vgs *model.GameState, player int, slotIndex int, diceValue int) error {
	if slotIndex < 0 || slotIndex >= len(vgs.Numbers[player]) {
		return fmt.Errorf("invalid slot index: %d", slotIndex)
	}

	vgs.Numbers[player][slotIndex] = model.Number{
		Value: float64(diceValue),
		Base:  0,
		Mark:  "",
	}
	return nil
}

func CONSTPI(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(math.Pi))
}

func CONSTE(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(math.E))
}

func CONSTN1(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(-1))
}

func CONST73(vgs *model.GameState, card *model.Card) error {
	target := model.NumFromFloat(73)
	exists73 := false

	for _, num := range vgs.Numbers[card.Owner] {
		if !num.IsNull() && num.Cmp(target) == 0 {
			exists73 = true
			break
		}
	}

	if !exists73 {
		return AddConstant(vgs, card.Owner, model.NumFromFloat(12))
	}
	return AddConstant(vgs, card.Owner, target)
}

func CONSTGOOGLE(vgs *model.GameState, card *model.Card) error {
	attackedPlayer := card.Inputs[0]
	attackedIndex := card.Inputs[1]

	if attackedPlayer >= 0 &&
		attackedPlayer < len(vgs.Numbers) &&
		attackedIndex >= 0 &&
		attackedIndex < len(vgs.Numbers[attackedPlayer]) {

		num := &vgs.Numbers[attackedPlayer][attackedIndex]

		if !num.IsNull() && !num.IsZero() {
			f := num.ToFloat64()
			if f > 0 {
				log := math.Log10(f)
				if log == math.Trunc(log) {
					// It's a power of 10 — steal it
					stolen := *num
					num.Value = 0
					num.Base = 0
					num.Mark = "n"
					return AddConstant(vgs, card.Owner, stolen)
				}
			}
		}
	}

	return AddConstant(vgs, card.Owner, model.NumFromFloat(10))
}

func CONST42(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(42))
}

func CONSTPHI(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(math.Phi))
}

func CONSTZERO(vgs *model.GameState, card *model.Card) error {
	n := model.Number{Value: 0, Base: 0, Mark: ""}
	return AddConstant(vgs, card.Owner, n)
}

func CONST7(vgs *model.GameState, card *model.Card) error {
	d1 := card.Inputs[0]
	d2 := card.Inputs[1]

	if err := AddConstant(vgs, card.Owner, model.NumFromFloat(7)); err != nil {
		return err
	}

	if d1+d2 == 7 {
		return AddConstant(vgs, card.Owner, model.NumFromFloat(7))
	}
	return nil
}

func CONST28(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(28))
}

func CONST6(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(6))
}

func CONSTFIBONACCI(vgs *model.GameState, card *model.Card) error {
	n := model.Number{Value: 1, Base: 0, Mark: "F"}
	return AddConstant(vgs, card.Owner, n)
}

func CONST69(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(69))
}

func CONSTTAU(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(math.Pi*2))
}

func CONSTTENPOWER(vgs *model.GameState, card *model.Card) error {
	dice := card.Inputs[0]
	return AddConstant(vgs, card.Owner, model.Number{Value: 1, Base: int64(dice), Mark: ""})
}

func CONSTGRAHAM(vgs *model.GameState, card *model.Card) error {
	return AddConstant(vgs, card.Owner, model.NumFromFloat(9))
}

func CONSTCUPID(vgs *model.GameState, card *model.Card) error {
	d1 := card.Inputs[0]
	d2 := card.Inputs[1]
	if d1 <= 3 && d2 <= 3 {
		return AddConstant(vgs, card.Owner, model.NumFromFloat(29))
	}
	return AddConstant(vgs, card.Owner, model.NumFromFloat(14))
}

func FACTORIAL(vgs *model.GameState, card *model.Card) error {
	dice := card.Inputs[0]
	result := 1
	for i := 2; i <= dice; i++ {
		result *= i
	}
	return AddConstant(vgs, card.Owner, model.NumFromFloat(float64(result)))
}
