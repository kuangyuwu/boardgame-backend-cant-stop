package cantstop

func dataRoll() map[string]any {
	return map[string]any{
		"type": "roll",
		"body": nil,
	}
}

func dataDice(dice [4]uint8, options [3][2][2]uint8, failed bool) map[string]any {
	return map[string]any{
		"type": "dice",
		"body": map[string]any{
			"dice":    dice,
			"options": options,
			"failed":  failed,
		},
	}
}

func dataAdvance(advances [2]uint8) map[string]any {
	return map[string]any{
		"type": "advance",
		"body": advances,
	}
}

func dataStop() map[string]any {
	return map[string]any{
		"type": "stop",
		"body": nil,
	}
}

func dataTurn(turn int) map[string]any {
	return map[string]any{
		"type": "turn",
		"body": turn,
	}
}

func dataPlayer(player int) map[string]any {
	return map[string]any{
		"type": "player",
		"body": player,
	}
}

func dataMove(move int) map[string]any {
	return map[string]any{
		"type": "move",
		"body": move,
	}
}

func dataBoard(board [][]int) map[string]any {
	return map[string]any{
		"type": "board",
		"body": board,
	}
}

func dataScores(scores []int) map[string]any {
	return map[string]any{
		"type": "scores",
		"body": scores,
	}
}

func dataWinner(winner int) map[string]any {
	return map[string]any{
		"type": "winner",
		"body": winner,
	}
}
