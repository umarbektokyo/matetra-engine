package theorems

import (
	"fmt"

	"github.com/umarbektokyo/matetra-engine/cards/constants"
	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

// Input: An
func ELEMENTIDENTITY(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	if err := utils.CheckCardMark(vgs, aP, aI); err != nil {
		return err
	}
	return nil
}

// Input: An
func ELEMENTCLOSURE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	if err := utils.CheckCardMark(vgs, aP, aI); err != nil {
		return err
	}

	// Preserve Fibonacci mark: "F" becomes "FI" (both Fibonacci and Immune)
	if vgs.Numbers[aP][aI].Mark == "F" {
		vgs.Numbers[aP][aI].Mark = "FI"
	} else {
		vgs.Numbers[aP][aI].Mark = "I"
	}
	return nil
}

// Input: pn
func ELEMENTDISTRIBUTIVE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	if err := utils.CheckCardMark(vgs, aP, aI); err != nil {
		return err
	}

	a := vgs.Numbers[aP][aI]

	for i := range vgs.Players {
		if i == aP {
			continue
		}
		dup := model.Number{Value: a.Value, Base: a.Base, Mark: ""}
		constants.AddConstant(vgs, i, dup)
	}

	return nil
}

// Input: pnpn
func ELEMENTCOMMUTATIVE(vgs *model.GameState, card *model.Card) error {
	p1, i1 := card.Inputs[0], card.Inputs[1]
	p2, i2 := card.Inputs[2], card.Inputs[3]

	if err := utils.CheckCardMark(vgs, p1, i1); err != nil {
		return err
	}
	if err := utils.CheckCardMark(vgs, p2, i2); err != nil {
		return err
	}

	a := &vgs.Numbers[p1][i1]
	b := &vgs.Numbers[p2][i2]

	a.Value, b.Value = b.Value, a.Value
	a.Base, b.Base = b.Base, a.Base
	a.Mark, b.Mark = b.Mark, a.Mark

	return nil
}

// Input: AnUn
func PYTHAGOREANTHEOREM(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]

	if err := utils.CheckCardMark(vgs, aP, aI); err != nil {
		return err
	}
	if err := utils.CheckCardMark(vgs, uP, uI); err != nil {
		return err
	}

	a := &vgs.Numbers[aP][aI]
	b := &vgs.Numbers[uP][uI]

	aNum := model.Number{Value: a.Value, Base: a.Base}
	bNum := model.Number{Value: b.Value, Base: b.Base}

	a2 := model.NumSquare(aNum)
	b2 := model.NumSquare(bNum)
	sum := model.NumAdd(a2, b2)
	result, err := model.NumSqrt(sum)
	if err != nil {
		return err
	}

	a.Value = result.Value
	a.Base = result.Base

	b.Value = 0
	b.Base = 0
	b.Mark = "n"

	return nil
}

// Input: An
func PASCALTRIANGLE(vgs *model.GameState, card *model.Card) error {
	player, index := card.Inputs[0], card.Inputs[1]
	if err := utils.CheckCardMark(vgs, player, index); err != nil {
		return err
	}

	L, R, err := utils.FindIsland(vgs, player, index)
	if err != nil {
		return err
	}

	nums := vgs.Numbers[player]

	// Collapse island: sum all non-immune numbers into the leftmost non-immune slot.
	// Immune numbers ("I" or "FI") are left untouched.
	sum := model.Number{Value: 0, Base: 0}
	destIdx := -1
	collapsible := 0
	for i := L; i <= R; i++ {
		mark := nums[i].Mark
		if mark == "I" || mark == "FI" {
			continue // skip immune numbers
		}
		sum = model.NumAdd(sum, model.Number{Value: nums[i].Value, Base: nums[i].Base})
		if destIdx == -1 {
			destIdx = i
		}
		collapsible++
	}

	if destIdx == -1 || collapsible < 2 {
		return fmt.Errorf("not enough non-immune numbers to collapse")
	}

	nums[destIdx].Value = sum.Value
	nums[destIdx].Base = sum.Base

	for i := L; i <= R; i++ {
		mark := nums[i].Mark
		if mark == "I" || mark == "FI" || i == destIdx {
			continue
		}
		nums[i].Value = 0
		nums[i].Base = 0
		nums[i].Mark = "n"
	}

	vgs.Numbers[player] = nums
	return nil
}

// Input: An
func FUNDAMENTALTHEOREMOFARITHMETIC(vgs *model.GameState, card *model.Card) error {
	player, index := card.Inputs[0], card.Inputs[1]
	if err := utils.CheckCardMark(vgs, player, index); err != nil {
		return err
	}

	num := &vgs.Numbers[player][index]
	n := model.Number{Value: num.Value, Base: num.Base}

	intVal, ok := n.ToInt64()
	if !ok || intVal <= 1 {
		return fmt.Errorf("number must be integer > 1")
	}

	factors := utils.PrimeFactors(intVal)

	for _, f := range factors {
		err := constants.AddConstant(vgs, player, model.NumFromInt(f))
		if err != nil {
			return err
		}
	}

	// Consume original
	num.Value = 0
	num.Base = 0
	num.Mark = "n"

	return nil
}
