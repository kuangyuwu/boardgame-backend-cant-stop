package cantstop

import "errors"

type RuleSet struct {
	numTempPaths    int8
	goal            int8
	dices           []int8
	pathLengths     []int8
	partitions      [][][]int8
	actionGenerator func([][]int8, func([]int8) bool) [][]int8
}

var ErrRuleSetNotFound = errors.New("rule set not found")

func getRuleSet(i int) (RuleSet, error) {
	switch i {
	case 0:
		return RuleSet{
			numTempPaths:    3,
			goal:            3,
			dices:           []int8{6, 6, 6, 6},
			pathLengths:     []int8{-1, -1, 3, 5, 7, 9, 11, 13, 11, 9, 7, 5, 3},
			partitions:      [][][]int8{{{0, 1}, {2, 3}}, {{0, 2}, {1, 3}}, {{0, 3}, {1, 2}}},
			actionGenerator: actionGenerator2Groups,
		}, nil
	// case 1:
	// 	return RuleSet{
	// 		numTempPaths: 2,
	// 		goal:         2,
	// 		dices:        []int{6, 6},
	// 		pathLengths:  []int{-1, 6, 6, 6, 6, 6, 6},
	// 		partitions:   [][][]int{{{0}, {1}}},
	// 	}, nil
	default:
		return RuleSet{}, ErrRuleSetNotFound
	}
}

func actionGenerator2Groups(grouping [][]int8, isValidAction func([]int8) bool) [][]int8 {
	actions := [][]int8{}
	g0 := sum(grouping[0])
	g1 := sum(grouping[1])
	if isValidAction([]int8{g0, g1}) {
		return [][]int8{{g0, g1}}
	}
	if isValidAction([]int8{g0}) {
		actions = append(actions, []int8{g0})
	}
	if isValidAction([]int8{g1}) {
		actions = append(actions, []int8{g1})
	}
	return actions
}

func sum(slice []int8) int8 {
	result := int8(0)
	for _, x := range slice {
		result += x
	}
	return result
}
