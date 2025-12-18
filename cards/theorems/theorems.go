package theorems

import (
	"fmt"
	"math/big"

	"github.com/umarbektokyo/matetra-engine/cards/constants"
	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

// Input: An
func ELEMENTIDENTITY(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	return nil
}

// Input: An
func ELEMENTCLOSURE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	a.Mark = "I"

	return nil
}

// Input: An
func ELEMENTDISTRIBUTIVE(vgs *model.GameState, card *model.Card) error {
	attackedPlayer := card.Inputs[0]
	attackedIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackedPlayer, attackedIndex)

	a := &vgs.Numbers[attackedPlayer][attackedIndex]

	for i := range vgs.Players {
		if i == attackedPlayer {
			continue
		}
		val := new(big.Float).SetPrec(a.Value.Prec()).Set(a.Value)
		constants.AddConstant(vgs, i, val, "")
	}

	return nil
}

// Input: AnAn
func ELEMENTCOMMUTATIVE(vgs *model.GameState, card *model.Card) error {
	player1 := card.Inputs[0]
	index1 := card.Inputs[1]
	player2 := card.Inputs[2]
	index2 := card.Inputs[3]

	utils.CheckCardMark(vgs, player1, index1)
	utils.CheckCardMark(vgs, player2, index2)

	a := &vgs.Numbers[player1][index1]
	b := &vgs.Numbers[player2][index2]

	tmpVal := new(big.Float).SetPrec(a.Value.Prec()).Set(a.Value)
	tmpMark := a.Mark

	a.Value = new(big.Float).SetPrec(a.Value.Prec()).Set(b.Value)
	a.Mark = b.Mark

	b.Value = tmpVal
	b.Mark = tmpMark

	return nil
}

// Input: AnUn
func PYTHAGOREANTHEOREM(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]
	prec := a.Value.Prec()

	a2 := new(big.Float).SetPrec(prec).Mul(a.Value, a.Value)
	b2 := new(big.Float).SetPrec(prec).Mul(b.Value, b.Value)
	sum := new(big.Float).SetPrec(prec).Add(a2, b2)
	a.Value.SetPrec(prec).Sqrt(sum)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: An
func PASCALTRIANGLE(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	index := card.Inputs[1]

	utils.CheckCardMark(vgs, player, index)

	// find island
	L, R, err := utils.FindIsland(vgs, player, index)
	if err != nil {
		return err
	}

	nums := vgs.Numbers[player]

	// collapse island using Pascal rule
	for i := L + 1; i <= R; i++ {
		prec := nums[L].Value.Prec()
		nums[L].Value = new(big.Float).
			SetPrec(prec).
			Add(nums[L].Value, nums[i].Value)

		nums[i].Value = big.NewFloat(0)
		nums[i].Mark = "n"
	}

	vgs.Numbers[player] = nums

	return nil
}

// Input: An
func FUNDAMENTALTHEOREMOFARITHMETIC(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	index := card.Inputs[1]

	utils.CheckCardMark(vgs, player, index)

	num := &vgs.Numbers[player][index]

	// must be integer
	intVal, ok := utils.FloatToIntExact(num.Value)
	if !ok || intVal.Cmp(big.NewInt(1)) <= 0 {
		return fmt.Errorf("number must be integer > 1")
	}

	factors := utils.PrimeFactors(intVal)

	// add factors
	for _, f := range factors {
		err := constants.AddConstant(
			vgs,
			player,
			new(big.Float).SetInt(f),
			"",
		)
		if err != nil {
			return err
		}
	}

	// consume original
	num.Value = big.NewFloat(0)
	num.Mark = "n"

	return nil
}
