package cantstop

import (
	"math/rand"
	"time"
)

func RollDice() [4]uint8 {
	result := [4]uint8{}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = uint8(rd.Intn(6)) + 1
	}
	return result
}

func GenerateOptions(dice [4]uint8, s State) ([3][2][2]uint8, error) {
	if !isValidDiceValue(dice) {
		return [3][2][2]uint8{}, ErrInvalidDiceValue
	}

	return generateOptions(dice, s), nil
}

func isValidDiceValue(dice [4]uint8) bool {
	for i := range 4 {
		if dice[i] == 0 || dice[i] > diceType[i] {
			return false
		}
	}
	return true
}

func generateOptions(dice [4]uint8, s State) [3][2][2]uint8 {
	groupings := groupDice(dice)
	options := [3][2][2]uint8{}
	for i, gr := range groupings {
		a1, a2 := gr[0], gr[1]
		op := [2][2]uint8{}
		if s.isValidAdvances([2]uint8{a1, a2}) {
			op[0] = [2]uint8{a1, a2}
		} else if s.isValidAdvances([2]uint8{a1, 0}) {
			op[0] = [2]uint8{a1, 0}
			if s.isValidAdvances([2]uint8{a2, 0}) {
				op[1] = [2]uint8{a2, 0}
			}
		} else if s.isValidAdvances([2]uint8{a2, 0}) {
			op[0] = [2]uint8{a2, 0}
		}
		options[i] = op
	}
	return options
}

func groupDice(dice [4]uint8) [3][2]uint8 {
	return [3][2]uint8{
		{dice[0] + dice[1], dice[2] + dice[3]},
		{dice[0] + dice[2], dice[1] + dice[3]},
		{dice[0] + dice[3], dice[1] + dice[2]},
	}
}
