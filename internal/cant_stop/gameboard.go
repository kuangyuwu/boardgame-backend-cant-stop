package cantstop

func (g GameCantStop) broadcastGameboard() {
	g.broadcast(dataGameboard(g.gameboard(), g.blockedPaths()))
}

func (g GameCantStop) gameboard() [][]space {
	gameboard := [][]space{}
	for i, length := range g.pathLengths {
		gameboard = append(gameboard, []space{})
		if length == -1 {
			continue
		}
		for j := int8(0); j < length; j++ {
			gameboard[i] = append(gameboard[i], space{
				Colors:  []int8{},
				HasTemp: false,
			})
		}
	}
	for n, p := range g.players {
		for i, length := range g.pathLengths {
			if length == -1 {
				continue
			}
			j := length - p.progress[i] - 1
			if j != -1 {
				gameboard[i][j].Colors = append(gameboard[i][j].Colors, int8(n))
			}
		}
	}
	p := g.players[g.playing]
	for i, x := range p.temp {
		j := g.pathLengths[i] - p.progress[i] - 1
		for k := int8(1); k <= x; k++ {
			gameboard[i][j+k].Colors = append(gameboard[i][j+k].Colors, g.playing)
			gameboard[i][j+k].HasTemp = true
		}
	}
	return gameboard
}

func (g GameCantStop) blockedPaths() []blocked {
	blockedPaths := []blocked{}
	for i, length := range g.pathLengths {
		if length == -1 {
			continue
		}
		n := g.completedBy(int8(i))
		if n != -1 {
			blockedPaths = append(blockedPaths, blocked{
				Path:  int8(i),
				Color: n,
			})
		}
	}
	return blockedPaths
}
