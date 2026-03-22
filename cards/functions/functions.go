package functions

import (
	"fmt"
	"math"

	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

func nullOut(num *model.Number) {
	num.Value = 0
	num.Base = 0
	num.Mark = "n"
}

func applyResult(dst *model.Number, r model.Number) error {
	if err := r.Sanitize(); err != nil {
		return err
	}
	dst.Value = r.Value
	dst.Base = r.Base
	return nil
}

func getNum(vgs *model.GameState, p, i int) (model.Number, error) {
	if p < 0 || p >= len(vgs.Numbers) {
		return model.Number{}, fmt.Errorf("player %d out of range", p)
	}
	if i < 0 || i >= len(vgs.Numbers[p]) {
		return model.Number{}, fmt.Errorf("slot %d out of range", i)
	}
	if err := utils.CheckCardMark(vgs, p, i); err != nil {
		return model.Number{}, err
	}
	n := vgs.Numbers[p][i]
	return model.Number{Value: n.Value, Base: n.Base}, nil
}

// Input: AnUn
func ADD(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	if err := applyResult(&vgs.Numbers[aP][aI], model.NumAdd(a, b)); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	return nil
}

// Input: AnUn
func SUBTRACT(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	if err := applyResult(&vgs.Numbers[aP][aI], model.NumSub(a, b)); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	return nil
}

// Input: AnUn
func MULTIPLY(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	if err := applyResult(&vgs.Numbers[aP][aI], model.NumMul(a, b)); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	return nil
}

// Input: AnUn
func DIVIDE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	if b.IsZero() { return fmt.Errorf("cannot divide by zero") }
	r, err := model.NumDiv(a, b); if err != nil { return err }
	if err := applyResult(&vgs.Numbers[aP][aI], r); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	return nil
}

// Input: An
func ABSOLUTEVALUE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	if _, err := getNum(vgs, aP, aI); err != nil { return err }
	if vgs.Numbers[aP][aI].Value < 0 { vgs.Numbers[aP][aI].Value = -vgs.Numbers[aP][aI].Value }
	return nil
}

// Input: An
func INVERSE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.IsZero() { return fmt.Errorf("cannot invert zero") }
	r, err := model.NumDiv(model.NumFromFloat(1), a); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: An
func NEGATIVE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	if _, err := getNum(vgs, aP, aI); err != nil { return err }
	vgs.Numbers[aP][aI].Value = -vgs.Numbers[aP][aI].Value
	return nil
}

// Input: An
func POSITIVE(vgs *model.GameState, card *model.Card) error {
	return nil
}

// Input: An
func SQRT(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.Value < 0 { return fmt.Errorf("cannot sqrt negative number (%s)", a.Display()) }
	r, err := model.NumSqrt(a); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: An
func SQUARE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], model.NumSquare(a))
}

// Input: And
func COSMOD(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	r := model.NumMul(a, model.NumFromFloat(math.Cos(float64(dice))))
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: And
func SINMOD(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	r := model.NumMul(a, model.NumFromFloat(math.Sin(float64(dice))))
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: And
func TANMOD(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	tanVal := math.Tan(float64(dice))
	if math.IsInf(tanVal, 0) || math.IsNaN(tanVal) {
		return fmt.Errorf("tan(%d) is undefined", dice)
	}
	r := model.NumMul(a, model.NumFromFloat(tanVal))
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: An
func LOG10(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.Value <= 0 { return fmt.Errorf("log10 requires positive number, got %s", a.Display()) }
	r, err := model.NumLog10(a); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: An
func EXPONENTIAL(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	r := model.NumExp(a)
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: An
func NATLOG(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.Value <= 0 { return fmt.Errorf("ln requires positive number, got %s", a.Display()) }
	r, err := model.NumLn(a); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: And
func LOGORHYTHM(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.Value <= 0 { return fmt.Errorf("log requires positive number, got %s", a.Display()) }
	if dice == 1 { return fmt.Errorf("log base 1 is undefined") }
	r, err := model.NumLogBase(a, float64(dice)); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: And
func ROOTBASE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	if a.Value < 0 && dice%2 == 0 { return fmt.Errorf("even root of negative number") }
	r, err := model.NumPow(a, 1.0/float64(dice)); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: And
func EXPONENTBASE(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	dice := card.Inputs[2]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	r, err := model.NumPow(a, float64(dice)); if err != nil { return err }
	return applyResult(&vgs.Numbers[aP][aI], r)
}

// Input: AnUnd
func POLYNOMIAL1(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	dice := card.Inputs[4]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	d := model.NumFromFloat(float64(dice))
	r := model.NumAdd(model.NumMul(a, d), b)
	if err := applyResult(&vgs.Numbers[aP][aI], r); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	return nil
}

// Input: AnUnUnd
func POLYNOMIAL2(vgs *model.GameState, card *model.Card) error {
	aP, aI := card.Inputs[0], card.Inputs[1]
	uP, uI := card.Inputs[2], card.Inputs[3]
	u2P, u2I := card.Inputs[4], card.Inputs[5]
	dice := card.Inputs[6]
	a, err := getNum(vgs, aP, aI); if err != nil { return err }
	b, err := getNum(vgs, uP, uI); if err != nil { return err }
	c, err := getNum(vgs, u2P, u2I); if err != nil { return err }
	d := model.NumFromFloat(float64(dice))
	d2 := model.NumSquare(d)
	r := model.NumAdd(model.NumAdd(model.NumMul(a, d2), model.NumMul(b, d)), c)
	if err := applyResult(&vgs.Numbers[aP][aI], r); err != nil { return err }
	nullOut(&vgs.Numbers[uP][uI])
	nullOut(&vgs.Numbers[u2P][u2I])
	return nil
}

// Input: A
func SIGMANOTATION(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	if player < 0 || player >= len(vgs.Numbers) { return fmt.Errorf("player out of range") }
	numbers := vgs.Numbers[player]

	dest := -1
	for i := range numbers {
		if numbers[i].Mark != "n" { dest = i; break }
	}
	if dest == -1 { return fmt.Errorf("no numbers to sum") }

	sum := model.Number{Value: 0, Base: 0}
	for i := range numbers {
		if numbers[i].Mark != "n" {
			sum = model.NumAdd(sum, model.Number{Value: numbers[i].Value, Base: numbers[i].Base})
			if i != dest { nullOut(&numbers[i]) }
		}
	}
	if err := sum.Sanitize(); err != nil { return err }
	numbers[dest].Value = sum.Value
	numbers[dest].Base = sum.Base
	vgs.Numbers[player] = numbers
	return nil
}

// Input: A
func PRODUCTNOTATION(vgs *model.GameState, card *model.Card) error {
	player := card.Inputs[0]
	if player < 0 || player >= len(vgs.Numbers) { return fmt.Errorf("player out of range") }
	numbers := vgs.Numbers[player]

	dest := -1
	for i := range numbers {
		if numbers[i].Mark != "n" { dest = i; break }
	}
	if dest == -1 { return fmt.Errorf("no numbers to multiply") }

	product := model.NumFromFloat(1)
	for i := range numbers {
		if numbers[i].Mark != "n" {
			product = model.NumMul(product, model.Number{Value: numbers[i].Value, Base: numbers[i].Base})
			if i != dest { nullOut(&numbers[i]) }
		}
	}
	if err := product.Sanitize(); err != nil { return err }
	numbers[dest].Value = product.Value
	numbers[dest].Base = product.Base
	vgs.Numbers[player] = numbers
	return nil
}
