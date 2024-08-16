package cantstop

type Data struct {
	Username string                 `json:"-"`
	Type     string                 `json:"type"`
	Body     map[string]interface{} `json:"body"`
}

func (g GameCantStop) send(d Data) {
	d.Username = g.players[g.playing].username
	g.fromGame <- d
}

func (g GameCantStop) broadcast(d Data) {
	d.Username = ""
	g.fromGame <- d
}

func (g GameCantStop) announce(content string) {
	g.broadcast(dataLogging(content))
}

func (g GameCantStop) sendExit(username string) {
	d := Data{
		Username: username,
		Type:     "exit",
		Body:     nil,
	}
	g.fromGame <- d
}

func dataLogging(content string) Data {
	data := Data{
		Type: "log",
		Body: map[string]interface{}{
			"content": content,
		},
	}
	return data
}

func dataStart(usernames []string, pathLengths []int8) Data {
	data := Data{
		Type: "start",
		Body: map[string]interface{}{
			"usernames":   usernames,
			"pathLengths": pathLengths,
		},
	}
	return data
}

func dataTurnCount(turnCount int16) Data {
	data := Data{
		Type: "turnCount",
		Body: map[string]interface{}{
			"turnCount": turnCount,
		},
	}
	return data
}

func dataPlayer(username string, isPlaying bool, score int8) Data {
	data := Data{
		Type: "player",
		Body: map[string]interface{}{
			"username":  username,
			"isPlaying": isPlaying,
			"score":     score,
		},
	}
	return data
}

func dataMoveCount(moveCount int16) Data {
	data := Data{
		Type: "moveCount",
		Body: map[string]interface{}{
			"moveCount": moveCount,
		},
	}
	return data
}

func dataRoll() Data {
	data := Data{
		Type: "roll",
		Body: nil,
	}
	return data
}

func dataResult(points []int8, options []option, failed bool) Data {
	data := Data{
		Type: "result",
		Body: map[string]interface{}{
			"points":  points,
			"options": options,
			"failed":  failed,
		},
	}
	return data
}

func dataConfirm() Data {
	data := Data{
		Type: "confirm",
		Body: nil,
	}
	return data
}

func dataWinner(username string) Data {
	data := Data{
		Type: "winner",
		Body: map[string]interface{}{
			"winner": username,
		},
	}
	return data
}

type space struct {
	Colors  []int8 `json:"colors"`
	HasTemp bool   `json:"hasTemp"`
}

type blocked struct {
	Path  int8 `json:"path"`
	Color int8 `json:"color"`
}

func dataGameboard(gameboard [][]space, blockedPaths []blocked) Data {
	data := Data{
		Type: "gameboard",
		Body: map[string]interface{}{
			"gameboard":    gameboard,
			"blockedPaths": blockedPaths,
		},
	}
	return data
}

// func dataExit(username string) Data {
// 	d := Data{
// 		Username: username,
// 		Type:     "exit",
// 		Body:     nil,
// 	}
// 	return d
// }

func dataTerminate() Data {
	d := Data{
		Username: "",
		Type:     "terminate",
		Body:     nil,
	}
	return d
}
