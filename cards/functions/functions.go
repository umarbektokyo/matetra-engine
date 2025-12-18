package functions

import (
	"fmt"
	"math"
	"math/big"

	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

// Input: AnUn
func ADD(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]

	a.Value.Add(a.Value, b.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: AnUn
func SUBTRACT(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]

	a.Value.Sub(a.Value, b.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: AnUn
func MULTIPLY(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]

	a.Value.Mul(a.Value, b.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: AnUn
func DIVIDE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]

	if b.Value.Cmp(big.NewFloat(0)) == 0 {
		return fmt.Errorf("Cannot divide by zero")
	}

	a.Value.Quo(a.Value, b.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: An
func ABSOLUTEVALUE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	a.Value.Abs(a.Value)

	return nil
}

// Input: AnUn
func INVERSE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	if a.Value.Cmp(big.NewFloat(0)) == 0 {
		return fmt.Errorf("Cannot divide by zero")
	}

	a.Value.Quo(big.NewFloat(1), a.Value)

	return nil
}

// Input: An
func NEGATIVE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	a.Value.Mul(a.Value, big.NewFloat(-1))

	return nil
}

// Input: An
func POSITIVE(vgs *model.GameState, card *model.Card) error {
	return nil
}

// Input: An
func SQRT(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	if a.Value.Sign() < 0 {
		return fmt.Errorf("cannot take a square root a negative number")
	}

	a.Value.Sqrt(a.Value)

	return nil
}

// Input: An
func SQUARE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	a.Value.Mul(a.Value, a.Value)

	return nil
}

// Input: An
func COSMOD(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	dice := utils.RollDice(6)
	cosVal := math.Cos(float64(dice))
	cosBig := new(big.Float).SetPrec(a.Value.Prec()).SetFloat64(cosVal)

	a.Value.Mul(a.Value, cosBig)

	return nil
}

// Input: An
func SINMOD(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	dice := utils.RollDice(6)
	sinVal := math.Sin(float64(dice))
	sinBig := new(big.Float).SetPrec(a.Value.Prec()).SetFloat64(sinVal)

	a.Value.Mul(a.Value, sinBig)

	return nil
}

// Input: An
func TANMOD(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	dice := utils.RollDice(6)
	tanVal := math.Tan(float64(dice))
	tanBig := new(big.Float).SetPrec(a.Value.Prec()).SetFloat64(tanVal)

	a.Value.Mul(a.Value, tanBig)

	return nil
}

// Input: An
func LOG10(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	if a.Value.Sign() < 0 {
		return fmt.Errorf("cannot take a logarithm a negative number")
	}

	val, _ := a.Value.Float64()
	logVal := math.Log10(val)
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(logVal)

	return nil
}

// Input: An
func EXPONENTIAL(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	val, _ := a.Value.Float64()
	expVal := math.Exp(val)
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(expVal)

	return nil
}

// Input: An
func NATLOG(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]

	if a.Value.Sign() < 0 {
		return fmt.Errorf("cannot take a logarithm a negative number")
	}

	val, _ := a.Value.Float64()
	lnVal := math.Log(val)
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(lnVal)

	return nil
}

// Input: An
func LOGORHYTHM(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	dice := utils.RollDice(6)

	if a.Value.Sign() < 0 {
		return fmt.Errorf("cannot take a logarithm a negative number")
	}

	val, _ := a.Value.Float64()
	logVal := math.Log(val) / math.Log(float64(dice))
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(logVal)

	return nil
}

// Input: An
func ROOTBASE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	dice := utils.RollDice(6)

	if a.Value.Sign() < 0 {
		return fmt.Errorf("cannot take a logarithm a negative number")
	}

	val, _ := a.Value.Float64()
	result := math.Pow(val, 1.0/float64(dice))
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(result)

	return nil
}

// Input: An
func EXPONENTBASE(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	dice := utils.RollDice(6)

	val, _ := a.Value.Float64()
	result := math.Pow(val, float64(dice))
	a.Value.SetPrec(a.Value.Prec()).SetFloat64(result)

	return nil
}

// Input: AnUn
func POLYNOMIAL1(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]
	prec := a.Value.Prec()
	d := new(big.Float).SetPrec(prec).SetFloat64(float64(utils.RollDice(6)))

	term1 := new(big.Float).SetPrec(prec).Mul(a.Value, d)

	a.Value.Add(term1, b.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	return nil
}

// Input: AnUnUn
func POLYNOMIAL2(vgs *model.GameState, card *model.Card) error {
	attackerPlayer := card.Inputs[0]
	attackerIndex := card.Inputs[1]
	userPlayer := card.Inputs[2]
	userIndex := card.Inputs[3]
	userPlayer2 := card.Inputs[4]
	userIndex2 := card.Inputs[5]

	utils.CheckCardMark(vgs, attackerPlayer, attackerIndex)
	utils.CheckCardMark(vgs, userPlayer, userIndex)
	utils.CheckCardMark(vgs, userPlayer2, userIndex2)

	a := &vgs.Numbers[attackerPlayer][attackerIndex]
	b := &vgs.Numbers[userPlayer][userIndex]
	c := &vgs.Numbers[userPlayer2][userIndex2]
	prec := a.Value.Prec()
	d := new(big.Float).SetPrec(prec).SetFloat64(float64(utils.RollDice(6)))
	d2 := new(big.Float).SetPrec(prec).Mul(d, d)

	term1 := new(big.Float).SetPrec(prec).Mul(a.Value, d2)
	term2 := new(big.Float).SetPrec(prec).Mul(b.Value, d)

	a.Value.Add(term1, term2)
	a.Value.Add(a.Value, c.Value)

	b.Value = big.NewFloat(0)
	b.Mark = "n"

	c.Value = big.NewFloat(0)
	c.Mark = "n"

	return nil
}

// Input: A
func SIGMANOTATION(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	numbers := vgs.Numbers[player]

	dest := -1
	for i, num := range numbers {
		if num.Mark != "n" {
			dest = i
			break
		}
	}
	if dest == -1 {
		return fmt.Errorf("no numbers to sum")
	}

	prec := numbers[dest].Value.Prec()
	sum := new(big.Float).SetPrec(prec).SetFloat64(0)

	for i := range numbers {
		if numbers[i].Mark != "n" {
			sum.Add(sum, numbers[i].Value)

			if i != dest {
				numbers[i].Value = big.NewFloat(0)
				numbers[i].Mark = "n"
			}
		}
	}

	numbers[dest].Value = sum

	// Write back
	vgs.Numbers[player] = numbers

	return nil
}

// Input: A
func PRODUCTNOTATION(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	numbers := vgs.Numbers[player]

	dest := -1
	for i, num := range numbers {
		if num.Mark != "n" {
			dest = i
			break
		}
	}
	if dest == -1 {
		return fmt.Errorf("no numbers to multiply")
	}

	prec := numbers[dest].Value.Prec()
	product := new(big.Float).SetPrec(prec).SetFloat64(1)

	for i := range numbers {
		if numbers[i].Mark != "n" {
			product.Mul(product, numbers[i].Value)

			if i != dest {
				numbers[i].Value = big.NewFloat(0)
				numbers[i].Mark = "n"
			}
		}
	}

	numbers[dest].Value = product

	// Write back
	vgs.Numbers[player] = numbers

	return nil
}
