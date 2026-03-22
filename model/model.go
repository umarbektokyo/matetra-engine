package model

import (
	"fmt"
	"math"
)

// Number represents a value in scientific notation: Value × 10^Base
// e.g., 3140 = {Value: 3.14, Base: 3}
// Zero is {Value: 0, Base: 0}
type Number struct {
	Value float64 `json:"value"`
	Base  int64   `json:"base"`
	Mark  string  `json:"mark"`
	// Marks:
	// "n" = null/empty slot
	// "F" = fibonacci (grows each turn)
	// "I" = immune (protected for one turn)
	// ""  = normal number
}

// Normalize adjusts Value to be in [1, 10) range and updates Base accordingly.
// Special case: zero stays as {0, 0}.
func (n *Number) Normalize() {
	if n.Value == 0 {
		n.Base = 0
		return
	}

	negative := n.Value < 0
	if negative {
		n.Value = -n.Value
	}

	for n.Value >= 10 {
		n.Value /= 10
		n.Base++
	}
	for n.Value < 1 && n.Value > 0 {
		n.Value *= 10
		n.Base--
	}

	if negative {
		n.Value = -n.Value
	}
}

// ToFloat64 converts the scientific notation back to a float64 (may lose precision for huge numbers).
func (n Number) ToFloat64() float64 {
	return n.Value * math.Pow(10, float64(n.Base))
}

// IsZero checks if the number is zero.
func (n Number) IsZero() bool {
	return n.Value == 0
}

// IsNull checks if the slot is empty.
func (n Number) IsNull() bool {
	return n.Mark == "n"
}

// IsPositive checks if the number is positive (not zero).
func (n Number) IsPositive() bool {
	return n.Value > 0
}

// Cmp compares two Numbers. Returns -1, 0, or 1.
func (n Number) Cmp(other Number) int {
	// Handle zeros
	if n.IsZero() && other.IsZero() {
		return 0
	}
	if n.IsZero() {
		if other.Value > 0 {
			return -1
		}
		return 1
	}
	if other.IsZero() {
		if n.Value > 0 {
			return 1
		}
		return -1
	}

	// Different signs
	nNeg := n.Value < 0
	oNeg := other.Value < 0
	if nNeg && !oNeg {
		return -1
	}
	if !nNeg && oNeg {
		return 1
	}

	// Same sign — compare by base first, then value
	if n.Base != other.Base {
		if n.Base > other.Base {
			if nNeg {
				return -1
			}
			return 1
		}
		if nNeg {
			return 1
		}
		return -1
	}

	// Same base
	if n.Value < other.Value {
		return -1
	}
	if n.Value > other.Value {
		return 1
	}
	return 0
}

// Display returns a human-readable string for the number.
func (n Number) Display() string {
	if n.Mark == "n" {
		return "[ ]"
	}
	if n.IsZero() {
		return "0"
	}
	if n.Base == 0 {
		return fmt.Sprintf("%.4g", n.Value)
	}
	if n.Base > -4 && n.Base < 7 {
		// Show as normal number if reasonable size
		f := n.ToFloat64()
		if f == math.Trunc(f) && math.Abs(f) < 1e9 {
			return fmt.Sprintf("%.0f", f)
		}
		return fmt.Sprintf("%.4g", f)
	}
	return fmt.Sprintf("%.4ge%d", n.Value, n.Base)
}

// --- Arithmetic operations (return new Number) ---

func NumAdd(a, b Number) Number {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}

	// Align to the larger base
	if a.Base > b.Base {
		diff := a.Base - b.Base
		if diff > 300 {
			return a // b is negligible
		}
		bVal := b.Value * math.Pow(10, -float64(diff))
		result := Number{Value: a.Value + bVal, Base: a.Base}
		result.Normalize()
		return result
	}
	diff := b.Base - a.Base
	if diff > 300 {
		return b // a is negligible
	}
	aVal := a.Value * math.Pow(10, -float64(diff))
	result := Number{Value: aVal + b.Value, Base: b.Base}
	result.Normalize()
	return result
}

func NumSub(a, b Number) Number {
	neg := b
	neg.Value = -neg.Value
	return NumAdd(a, neg)
}

func NumMul(a, b Number) Number {
	if a.IsZero() || b.IsZero() {
		return Number{Value: 0, Base: 0}
	}
	result := Number{
		Value: a.Value * b.Value,
		Base:  a.Base + b.Base,
	}
	result.Normalize()
	return result
}

func NumDiv(a, b Number) (Number, error) {
	if b.IsZero() {
		return Number{}, fmt.Errorf("cannot divide by zero")
	}
	if a.IsZero() {
		return Number{Value: 0, Base: 0}, nil
	}
	result := Number{
		Value: a.Value / b.Value,
		Base:  a.Base - b.Base,
	}
	result.Normalize()
	return result, nil
}

func NumAbs(a Number) Number {
	result := a
	if result.Value < 0 {
		result.Value = -result.Value
	}
	return result
}

func NumNeg(a Number) Number {
	result := a
	result.Value = -result.Value
	return result
}

func NumSqrt(a Number) (Number, error) {
	if a.Value < 0 {
		return Number{}, fmt.Errorf("cannot take square root of negative number")
	}
	if a.IsZero() {
		return Number{Value: 0, Base: 0}, nil
	}

	val := a.Value
	base := a.Base

	// Make base even for clean sqrt
	if base%2 != 0 {
		val *= 10
		base--
	}

	result := Number{
		Value: math.Sqrt(val),
		Base:  base / 2,
	}
	result.Normalize()
	return result, nil
}

func NumSquare(a Number) Number {
	if a.IsZero() {
		return Number{Value: 0, Base: 0}
	}
	result := Number{
		Value: a.Value * a.Value,
		Base:  a.Base * 2,
	}
	result.Normalize()
	return result
}

func NumPow(a Number, exp float64) (Number, error) {
	if a.IsZero() {
		if exp <= 0 {
			return Number{}, fmt.Errorf("0 raised to non-positive power is undefined")
		}
		return Number{Value: 0, Base: 0}, nil
	}

	// (v × 10^b)^e = v^e × 10^(b*e)
	negative := a.Value < 0
	absVal := math.Abs(a.Value)

	logVal := math.Log10(absVal) + float64(a.Base)
	logResult := logVal * exp

	resultBase := int64(math.Floor(logResult))
	resultVal := math.Pow(10, logResult-float64(resultBase))

	if negative && math.Mod(exp, 2) != 0 {
		resultVal = -resultVal
	}

	result := Number{Value: resultVal, Base: resultBase}
	result.Normalize()
	return result, nil
}

func NumLog10(a Number) (Number, error) {
	if a.Value <= 0 {
		return Number{}, fmt.Errorf("cannot take log of non-positive number")
	}
	// log10(v × 10^b) = log10(v) + b
	result := math.Log10(a.Value) + float64(a.Base)
	r := Number{Value: result, Base: 0}
	r.Normalize()
	return r, nil
}

func NumLn(a Number) (Number, error) {
	if a.Value <= 0 {
		return Number{}, fmt.Errorf("cannot take ln of non-positive number")
	}
	// ln(v × 10^b) = ln(v) + b*ln(10)
	result := math.Log(a.Value) + float64(a.Base)*math.Ln10
	r := Number{Value: result, Base: 0}
	r.Normalize()
	return r, nil
}

func NumLogBase(a Number, base float64) (Number, error) {
	if a.Value <= 0 {
		return Number{}, fmt.Errorf("cannot take log of non-positive number")
	}
	if base <= 0 || base == 1 {
		return Number{}, fmt.Errorf("invalid log base: %v", base)
	}
	// log_b(x) = ln(x) / ln(b)
	lnA, err := NumLn(a)
	if err != nil {
		return Number{}, err
	}
	lnBase := math.Log(base)
	result, err := NumDiv(lnA, Number{Value: lnBase, Base: 0})
	if err != nil {
		return Number{}, err
	}
	result.Normalize()
	return result, nil
}

func NumExp(a Number) Number {
	// e^(v × 10^b)
	// For large exponents this could overflow, but that's fine with our representation
	exponent := a.ToFloat64()
	if exponent > 709 {
		// Too large for float64 exp, use log approach
		// e^x = 10^(x / ln(10))
		log10val := exponent / math.Ln10
		base := int64(math.Floor(log10val))
		val := math.Pow(10, log10val-float64(base))
		result := Number{Value: val, Base: base}
		result.Normalize()
		return result
	}
	if exponent < -709 {
		return Number{Value: 0, Base: 0}
	}
	r := math.Exp(exponent)
	result := Number{Value: r, Base: 0}
	result.Normalize()
	return result
}

func NumFromFloat(f float64) Number {
	if f == 0 {
		return Number{Value: 0, Base: 0}
	}
	n := Number{Value: f, Base: 0}
	n.Normalize()
	return n
}

func NumFromInt(i int64) Number {
	return NumFromFloat(float64(i))
}

// NumIsInt checks if a Number represents an integer value.
func (n Number) IsInt() bool {
	if n.IsZero() {
		return true
	}
	if n.Base < 0 {
		return false
	}
	f := n.ToFloat64()
	return f == math.Trunc(f) && !math.IsInf(f, 0)
}

// NumToInt64 attempts to convert to int64. Returns (value, ok).
func (n Number) ToInt64() (int64, bool) {
	if !n.IsInt() {
		return 0, false
	}
	f := n.ToFloat64()
	if f > math.MaxInt64 || f < math.MinInt64 || math.IsInf(f, 0) {
		return 0, false
	}
	return int64(f), true
}

// Sanitize checks if Value is NaN or Inf and returns an error if so.
// Call after any math operation that could produce bad floats.
func (n Number) Sanitize() error {
	if math.IsNaN(n.Value) {
		return fmt.Errorf("result is NaN (undefined)")
	}
	if math.IsInf(n.Value, 0) {
		return fmt.Errorf("result is infinite")
	}
	return nil
}

// --- Game Model ---

type WinCondition int

const (
	WinLimit    WinCondition = iota // Number crosses limit → point. 3 points = win.
	WinLargest                      // Largest number after X turns wins.
	WinSmallest                     // Smallest number after X turns wins.
)

type GameSettings struct {
	WinCondition WinCondition `json:"winCondition"`
	TurnLimit    int          `json:"turnLimit,omitempty"` // For WinLargest/WinSmallest
	HandSize     int          `json:"handSize"`            // Number of cards in hand
	Limit        float64      `json:"limit,omitempty"`     // For WinLimit mode, starts at 1000
}

type Card struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Method      string `json:"method"`
	Owner       int    `json:"owner"`    // -1: deck, -2: used, >= 0: player
	Inputs      []int  `json:"inputs"`   // filled by player
	InputsReq   string `json:"inputsReq"`
	Precedence  int    `json:"precedence"`
}

type PlayerStats struct {
	CardsPlayed   int     `json:"cardsPlayed"`
	AttacksDealt  int     `json:"attacksDealt"`  // cards targeting other players' numbers
	DiceRolls     int     `json:"diceRolls"`
	DiceTotal     int     `json:"diceTotal"`
	BiggestNumber float64 `json:"biggestNumber"` // largest number ever held (as float for display)
	BiggestBase   int64   `json:"biggestBase"`
}

type Player struct {
	Name   string      `json:"name"`
	Hash   string      `json:"-"`
	Online bool        `json:"online"`
	Points int         `json:"points"`
	Stats  PlayerStats `json:"stats"`
}

type GameState struct {
	GameID   string       `json:"gameId"`
	Players  []Player     `json:"players"`
	Cards    []Card       `json:"cards"`
	Numbers  [][5]Number  `json:"numbers"`
	Done     []bool       `json:"done"`
	DiceUsed []bool       `json:"diceUsed"`
	Queue    []int        `json:"queue"`
	Turn     int          `json:"turn"`
	Started  bool         `json:"started"`
	Finished bool         `json:"finished"`
	Winner   int          `json:"winner"`
	Settings GameSettings `json:"settings"`
	Log      []string     `json:"log"`
}
