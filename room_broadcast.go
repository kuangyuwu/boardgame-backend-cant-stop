package main

func (r Room) broadcast(data Data) {
	for _, u := range r.users {
		u.send <- data
	}
}

func (r Room) broadcastLog(event string) {
	r.broadcast(Data{
		Type: "log",
		Body: map[string]interface{}{
			"event": event,
		},
	})
}

func (r Room) broadcastPrepUpdate() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.status == statusInPrep {
			u.sendPrepUpdate()
		}
	}
}

func (r Room) broadcastStart() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := Data{
		Type: "start",
		Body: map[string]interface{}{
			"pathLengths": r.game.ruleset.pathLengths,
			"usernames":   r.usernames(),
		},
	}
	r.broadcast(data)
}

func (r Room) broadcastTurnCount() {
	data := Data{
		Type: "turnCount",
		Body: map[string]interface{}{
			"turnCount": r.game.turnCount,
		},
	}
	r.broadcast(data)
}

func (r Room) broadcastPlayer(isPlaying bool) {
	u := r.users[r.game.playing]
	data := Data{
		Type: "player",
		Body: map[string]interface{}{
			"username":  u.username,
			"isPlaying": isPlaying,
			"score":     u.player.score(),
		},
	}
	r.broadcast(data)
}

func (r Room) broadcastMoveCount() {
	data := Data{
		Type: "moveCount",
		Body: map[string]interface{}{
			"moveCount": r.game.turnCount,
		},
	}
	r.broadcast(data)
}

func (r Room) broadcastWinner(username string) {
	data := Data{
		Type: "winner",
		Body: map[string]interface{}{
			"winner": username,
		},
	}
	r.broadcast(data)
}

func (r Room) broadcastGameboard() {
	type Space struct {
		Colors  []int `json:"colors"`
		HasTemp bool  `json:"hasTemp"`
	}
	gameboard := [][]Space{}
	for path, length := range r.game.ruleset.pathLengths {
		gameboard = append(gameboard, []Space{})
		if length == -1 {
			continue
		}
		for j := 0; j < length; j++ {
			gameboard[path] = append(gameboard[path], Space{
				Colors:  []int{},
				HasTemp: false,
			})
		}
	}
	for n, u := range r.users {
		for path, length := range r.game.ruleset.pathLengths {
			if length == -1 {
				continue
			}
			j := length - u.player.state[path] - 1
			if j != -1 {
				gameboard[path][j].Colors = append(gameboard[path][j].Colors, n)
			}
		}
	}
	u := r.users[r.game.playing]
	for path, x := range u.player.temp {
		j := r.game.ruleset.pathLengths[path] - u.player.state[path] - 1
		for k := 1; k <= x; k++ {
			gameboard[path][j+k].Colors = append(gameboard[path][j+k].Colors, r.game.playing)
			gameboard[path][j+k].HasTemp = true
		}
	}
	blocked := []struct {
		Path  int `json:"path"`
		Color int `json:"color"`
	}{}
	for path, length := range r.game.ruleset.pathLengths {
		if length == -1 {
			continue
		}
		n := r.completedBy(path)
		if n != -1 {
			blocked = append(blocked, struct {
				Path  int `json:"path"`
				Color int `json:"color"`
			}{
				Path:  path,
				Color: n,
			})
		}
	}
	data := Data{
		Type: "gameboard",
		Body: map[string]interface{}{
			"gameboard": gameboard,
			"blocked":   blocked,
		},
	}
	r.broadcast(data)
}
