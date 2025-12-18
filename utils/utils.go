package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"

	"github.com/umarbektokyo/matetra-engine/model"
)

var VERSION = "0.1"
var PORT int = 1729
var DECK_PATH string = "cards/cards.csv"
var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func MatetraSplash() {
	content, err := os.ReadFile("ascii.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(content))
}

func ValidateInputs(vgs *model.GameState, card *model.Card) error {
	// Check the length
	if len(card.Inputs) != len(card.InputsReq) {
		return fmt.Errorf(
			"%s expects %d inputs, got %d",
			card.Method, len(card.InputsReq), len(card.Inputs),
		)
	}

	// Check input values
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
			fmt.Println("You forgot to implement this!!")

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

func RollDice(sides int) int {
	roll := r.Intn(sides) + 1
	return roll
}

func CheckCardMark(vgs *model.GameState, playerIndex int, numberIndex int) error {
	if vgs.Numbers[playerIndex][numberIndex].Mark == "n" {
		return fmt.Errorf("cannot use null card")
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

	// expand left
	L := index
	for L > 0 && nums[L-1].Mark != "n" {
		L--
	}

	// expand right
	R := index
	for R < len(nums)-1 && nums[R+1].Mark != "n" {
		R++
	}

	return L, R, nil
}

func PrimeFactors(n *big.Int) []*big.Int {
	factors := []*big.Int{}
	x := new(big.Int).Set(n)

	two := big.NewInt(2)
	zero := big.NewInt(0)

	for new(big.Int).Mod(x, two).Cmp(zero) == 0 {
		factors = append(factors, big.NewInt(2))
		x.Div(x, two)
	}

	d := big.NewInt(3)
	for d.Mul(d, d).Cmp(x) <= 0 {
		for new(big.Int).Mod(x, d).Cmp(zero) == 0 {
			factors = append(factors, new(big.Int).Set(d))
			x.Div(x, d)
		}
		d.Add(d, two)
	}

	if x.Cmp(big.NewInt(1)) > 0 {
		factors = append(factors, x)
	}

	return factors
}

func FloatToIntExact(f *big.Float) (*big.Int, bool) {
	i := new(big.Int)
	if f.IsInt() {
		f.Int(i)
		return i, true
	}
	return nil, false
}
