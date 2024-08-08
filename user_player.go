package main

type Player struct {
	totalMoves int
	state      []int
	temp       map[int]int
}

func (u *User) initializePlayer(pathLengths []int) {
	u.player = &Player{
		totalMoves: 0,
		state:      make([]int, len(pathLengths)),
		temp:       map[int]int{},
	}
	copy(u.player.state, pathLengths)
}

func (p *Player) resetTemp() {
	for path := range p.temp {
		delete(p.temp, path)
	}
}

func (p *Player) takeAction(path int) {
	if _, ok := p.temp[path]; !ok {
		p.temp[path] = 0
	}
	p.temp[path]++
}

func (p *Player) undoAction(path int) {
	p.temp[path]--
	if p.temp[path] == 0 {
		delete(p.temp, path)
	}
}

func (p *Player) updateState() {
	for path, progress := range p.temp {
		p.state[path] -= progress
	}
}

func (p *Player) addMoves(move int) {
	p.totalMoves += move
}

func (p Player) score() int {
	count := 0
	for _, remaining := range p.state {
		if remaining == 0 {
			count++
		}
	}
	return count
}

func (p Player) isWinner(goal int) bool {
	return p.score() >= goal
}
