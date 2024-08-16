package cantstop

import (
	"math/rand"
	"time"
)

type option struct {
	Grouping [][]int8 `json:"grouping"`
	Actions  [][]int8 `json:"actions"`
}

func rollDices(dices []int8) []int8 {
	result := make([]int8, 0, len(dices))
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, d := range dices {
		result = append(result, int8(rd.Intn(int(d)))+1)
	}
	return result
}

func pointsToGroupings(points []int8, partitions [][][]int8) [][][]int8 {
	groupings := [][][]int8{}
	for _, partition := range partitions {
		grouping := [][]int8{}
		for _, part := range partition {
			group := []int8{}
			for _, i := range part {
				group = append(group, points[i])
			}
			grouping = append(grouping, group)
		}
		groupings = append(groupings, grouping)
	}
	return groupings
}
