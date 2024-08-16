package cantstop

type player struct {
	username   string
	totalMoves int32
	progress   []int8
	temp       map[int8]int8
	left       bool
}

func newPlayer(username string, pathLengths []int8) player {
	p := player{
		username:   username,
		totalMoves: 0,
		progress:   make([]int8, len(pathLengths)),
		temp:       map[int8]int8{},
	}
	copy(p.progress, pathLengths)
	return p
}

func (p *player) resetTemp() {
	for i := range p.temp {
		delete(p.temp, i)
	}
}

func (p *player) takeAction(i int8) {
	if _, ok := p.temp[i]; !ok {
		p.temp[i] = 0
	}
	p.temp[i]++
}

func (p *player) undoAction(i int8) {
	p.temp[i]--
	if p.temp[i] == 0 {
		delete(p.temp, i)
	}
}

func (p *player) updateState() {
	for i, k := range p.temp {
		p.progress[i] -= k
	}
}

func (p *player) addMoves(moveCount int16) {
	p.totalMoves += int32(moveCount)
}

func (p player) score() int8 {
	count := int8(0)
	for _, k := range p.progress {
		if k == 0 {
			count++
		}
	}
	return count
}
