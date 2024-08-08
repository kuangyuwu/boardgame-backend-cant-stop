package main

import (
	"math/rand"
	"time"
)

type Game struct {
	turnCount int
	playing   int
	moveCount int
	ruleset   RuleSet
}

func (r *Room) startGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ruleset, _ := getRuleSet(0)
	r.game = &Game{
		turnCount: 0,
		playing:   -1,
		moveCount: 0,
		ruleset:   ruleset,
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rd.Shuffle(len(r.users), func(i, j int) { r.users[i], r.users[j] = r.users[j], r.users[i] })

	for _, u := range r.users {
		u.mu.Lock()
		if u != r.host {
			u.isReady = false
		}
		u.initializePlayer(ruleset.pathLengths)
		u.status = statusInGameNotPlaying
		u.mu.Unlock()
	}
}

func (r Room) nextTurn() {
	r.game.turnCount++
	r.game.playing = -1
	r.broadcastTurnCount()
	r.nextPlayer()
}

func (r Room) nextPlayer() {
	r.game.playing++
	if r.game.playing == len(r.users) {
		r.nextTurn()
	}
	r.game.moveCount = 0
	r.broadcastPlayer(true)
	r.nextMove(true)
}

func (r Room) nextMove(isFirstMove bool) {
	r.game.moveCount++
	r.broadcastMoveCount()
	r.users[r.game.playing].status = statusInGameRolling
	if isFirstMove {
		r.users[r.game.playing].sendRoll()
	} else {
		r.users[r.game.playing].handleRoll()
	}

}

func (r Room) completedBy(path int) int {
	for n, u := range r.users {
		if u.player.state[path] == 0 {
			return n
		}
	}
	return -1
}

func (r *Room) isValidPath(u *User, path int) bool {
	if r.completedBy(path) != -1 {
		return false
	}
	if k, ok := u.player.temp[path]; ok {
		return u.player.state[path] > k
	}
	return len(u.player.temp) < r.game.ruleset.numTempPaths
}

func (r *Room) isValidAction(u *User, action []int) bool {
	for i, path := range action {
		if !r.isValidPath(u, path) {
			for j := i - 1; j >= 0; j-- {
				u.player.undoAction(action[j])
			}
			return false
		}
		u.player.takeAction(path)
	}
	for _, path := range action {
		u.player.undoAction(path)
	}
	return true
}
