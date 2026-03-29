package cards

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/umarbektokyo/matetra-engine/cards/constants"
	"github.com/umarbektokyo/matetra-engine/cards/functions"
	"github.com/umarbektokyo/matetra-engine/cards/theorems"
	"github.com/umarbektokyo/matetra-engine/model"
	"github.com/umarbektokyo/matetra-engine/utils"
)

//go:embed cards.csv
var CardsCSV []byte

func LoadCards() ([]model.Card, error) {
	reader := csv.NewReader(bytes.NewReader(CardsCSV))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	records = records[1:] // skip header

	var cards []model.Card

	for _, row := range records {
		count, err := strconv.Atoi(row[6])
		if err != nil {
			return nil, fmt.Errorf("bad count for card %s: %v", row[0], err)
		}
		precedence, _ := strconv.Atoi(row[8])

		for i := 0; i < count; i++ {
			card := model.Card{
				Name:        row[0],
				Description: row[2],
				Type:        row[3],
				Method:      row[4],
				InputsReq:   row[5],
				Owner:       -1,
				Inputs:      []int{},
				Precedence:  precedence,
			}
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func CardFunction(vgs *model.GameState, cardIndex int) error {
	card := &vgs.Cards[cardIndex]

	// Re-validate at apply time: game state may have changed since queue time
	// (e.g. an earlier card in the queue granted immunity). If this fails,
	// applyCards will skip this card gracefully instead of deadlocking.
	if err := utils.ValidateInputs(vgs, card); err != nil {
		return err
	}

	switch card.Method {
	// Functions
	case "ADD":
		return functions.ADD(vgs, card)
	case "SUBTRACT":
		return functions.SUBTRACT(vgs, card)
	case "MULTIPLY":
		return functions.MULTIPLY(vgs, card)
	case "DIVIDE":
		return functions.DIVIDE(vgs, card)
	case "ABSOLUTEVALUE":
		return functions.ABSOLUTEVALUE(vgs, card)
	case "INVERSE":
		return functions.INVERSE(vgs, card)
	case "NEGATIVE":
		return functions.NEGATIVE(vgs, card)
	case "POSITIVE":
		return functions.POSITIVE(vgs, card)
	case "SQRT":
		return functions.SQRT(vgs, card)
	case "SQUARE":
		return functions.SQUARE(vgs, card)
	case "COSMOD":
		return functions.COSMOD(vgs, card)
	case "LOG10":
		return functions.LOG10(vgs, card)
	case "EXPONENTIAL":
		return functions.EXPONENTIAL(vgs, card)
	case "NATLOG":
		return functions.NATLOG(vgs, card)
	case "SINMOD":
		return functions.SINMOD(vgs, card)
	case "TANMOD":
		return functions.TANMOD(vgs, card)
	case "LOGORHYTHM":
		return functions.LOGORHYTHM(vgs, card)
	case "ROOTBASE":
		return functions.ROOTBASE(vgs, card)
	case "EXPONENTBASE":
		return functions.EXPONENTBASE(vgs, card)
	case "SIGMANOTATION":
		return functions.SIGMANOTATION(vgs, card)
	case "PRODUCTNOTATION":
		return functions.PRODUCTNOTATION(vgs, card)
	case "POLYNOMIAL2":
		return functions.POLYNOMIAL2(vgs, card)
	case "POLYNOMIAL1":
		return functions.POLYNOMIAL1(vgs, card)
	// Theorems
	case "ELEMENTIDENTITY":
		return theorems.ELEMENTIDENTITY(vgs, card)
	case "ELEMENTCLOSURE":
		return theorems.ELEMENTCLOSURE(vgs, card)
	case "ELEMENTDISTRIBUTIVE":
		return theorems.ELEMENTDISTRIBUTIVE(vgs, card)
	case "ELEMENTCOMMUTATIVE":
		return theorems.ELEMENTCOMMUTATIVE(vgs, card)
	case "PYTHAGOREANTHEOREM":
		return theorems.PYTHAGOREANTHEOREM(vgs, card)
	case "PASCALTRIANGLE":
		return theorems.PASCALTRIANGLE(vgs, card)
	case "FUNDAMENTALTHEOREMOFARITHMETIC":
		return theorems.FUNDAMENTALTHEOREMOFARITHMETIC(vgs, card)
	// Constants
	case "CONSTE":
		return constants.CONSTE(vgs, card)
	case "CONSTN1":
		return constants.CONSTN1(vgs, card)
	case "CONST73":
		return constants.CONST73(vgs, card)
	case "CONSTGOOGLE":
		return constants.CONSTGOOGLE(vgs, card)
	case "CONST42":
		return constants.CONST42(vgs, card)
	case "CONSTPHI":
		return constants.CONSTPHI(vgs, card)
	case "CONSTZERO":
		return constants.CONSTZERO(vgs, card)
	case "CONSTPI":
		return constants.CONSTPI(vgs, card)
	case "CONST7":
		return constants.CONST7(vgs, card)
	case "CONST28":
		return constants.CONST28(vgs, card)
	case "CONST6":
		return constants.CONST6(vgs, card)
	case "CONSTFIBONACCI":
		return constants.CONSTFIBONACCI(vgs, card)
	case "CONST69":
		return constants.CONST69(vgs, card)
	case "CONSTTAU":
		return constants.CONSTTAU(vgs, card)
	case "CONSTTENPOWER":
		return constants.CONSTTENPOWER(vgs, card)
	case "CONSTGRAHAM":
		return constants.CONSTGRAHAM(vgs, card)
	case "CONSTCUPID":
		return constants.CONSTCUPID(vgs, card)
	case "FACTORIAL":
		return constants.FACTORIAL(vgs, card)
	default:
		return fmt.Errorf("unknown card method %s", card.Method)
	}
}
