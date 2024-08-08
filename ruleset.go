package main

import "errors"

type RuleSet struct {
	numTempPaths int
	goal         int
	dices        []int
	pathLengths  []int
	partitions   [][][]int
}

var ErrRuleSetNotFound = errors.New("rule set not found")

func getRuleSet(i int) (RuleSet, error) {
	switch i {
	case 0:
		return RuleSet{
			numTempPaths: 3,
			goal:         3,
			dices:        []int{6, 6, 6, 6},
			pathLengths:  []int{-1, -1, 3, 5, 7, 9, 11, 13, 11, 9, 7, 5, 3},
			partitions:   [][][]int{{{0, 1}, {2, 3}}, {{0, 2}, {1, 3}}, {{0, 3}, {1, 2}}},
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
