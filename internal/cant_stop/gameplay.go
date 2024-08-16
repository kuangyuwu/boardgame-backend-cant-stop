package cantstop

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	maxTurnCount = 1000
	maxMoveCount = 1000
)

func (g *GameCantStop) nextTurn() {
	if g.turnCount == maxTurnCount {
		g.logErrorAndTerminate("max turn count reached")
	}
	g.turnCount++
	g.playing = -1
	g.broadcast(dataTurnCount(g.turnCount))
	g.nextPlayer()
}

func (g *GameCantStop) nextPlayer() {
	g.playing++
	if int(g.playing) == len(g.players) {
		g.nextTurn()
		return
	}
	g.moveCount = 0
	g.failed = false
	p := g.players[g.playing]
	g.announce(fmt.Sprintf("Player %s's turn", p.username))
	g.broadcast(dataPlayer(p.username, true, p.score()))
	g.nextMove()
}

func (g *GameCantStop) nextMove() {
	if g.moveCount == maxMoveCount {
		g.logErrorAndTerminate("max move count reached")
	}
	g.moveCount++
	g.broadcast(dataMoveCount(g.moveCount))
	g.phase = phaseRoll
	if g.moveCount == 1 {
		g.send(dataRoll())
		g.mu.Unlock()
	} else {
		g.handleRoll()
	}
}

func (g *GameCantStop) handleRoll() {
	defer g.mu.Unlock()
	if g.phase != phaseRoll {
		logError(fmt.Sprintf("unexpected roll message in phase %d", g.phase))
		return
	}
	points := rollDices(g.dices)
	p := g.players[g.playing]
	g.announce(fmt.Sprintf("Player %s rolled %s", p.username, numsToString(points)))
	groupings := pointsToGroupings(points, g.partitions)
	options := []option{}
	failed := true
	for _, grouping := range groupings {
		actions := g.actionGenerator(grouping, g.isValidAction)
		if len(actions) > 0 {
			failed = false
		}
		options = append(options, option{
			Grouping: grouping,
			Actions:  actions,
		})
	}
	if failed {
		g.announce("No valid actions")
		g.failed = true
		g.phase = phaseConfirm
	} else {
		g.phase = phaseAct
	}
	g.send(dataResult(points, options, failed))
}

func (g *GameCantStop) handleAct(body map[string]interface{}) {
	defer g.mu.Unlock()
	if g.phase != phaseAct {
		logError(fmt.Sprintf("unexpected act message in phase %d", g.phase))
		return
	}
	p := g.players[g.playing]
	action := []int8{}
	for _, num := range body["action"].([]interface{}) {
		i, ok := num.(float64)
		if !ok {
			logError("invalid act message")
			return
		}
		p.takeAction(int8(i))
		action = append(action, int8(i))
	}
	g.broadcastGameboard()
	g.announce(fmt.Sprintf("Player %s advanced %s", p.username, numsToString(action)))
	g.phase = phaseConfirm
	g.send(dataConfirm())
}

func (g *GameCantStop) handleConfirm(body map[string]interface{}) {
	if g.phase != phaseConfirm {
		logError(fmt.Sprintf("unexpected confirm message in phase %d", g.phase))
		return
	}
	p := g.players[g.playing]
	if g.failed {
		p.resetTemp()
		p.addMoves(g.moveCount)
		g.broadcastGameboard()
		g.broadcast(dataPlayer(p.username, false, p.score()))
		g.nextPlayer()
		return
	}
	willContinue, ok := body["willContinue"].(bool)
	if !ok {
		logError("invalid confirm message")
		return
	}
	if willContinue {
		g.phase = phaseRoll
		g.nextMove()
	} else {
		p.updateState()
		p.resetTemp()
		p.addMoves(g.moveCount)
		g.broadcastGameboard()
		g.broadcast(dataPlayer(p.username, false, p.score()))
		g.announce(fmt.Sprintf("Player %s ended their turn", p.username))
		if g.isWinner(p) {
			g.broadcast(dataWinner(p.username))
			g.ended = true
			g.mu.Unlock()
			return
		}
		g.nextPlayer()
	}
}

func (g *GameCantStop) handleExit(username string) {
	defer g.mu.Unlock()
	if !g.ended {
		g.logErrorAndTerminate(fmt.Sprintf("player %s exited unexpectedly", username))
	}
	for n, p := range g.players {
		if p.username == username {
			g.players[n].left = true
		}
	}
	g.sendExit(username)
}

func (g GameCantStop) completedBy(i int8) int8 {
	for n, p := range g.players {
		if p.progress[i] == 0 {
			return int8(n)
		}
	}
	return -1
}

func (g GameCantStop) isValidPath(i int8) bool {
	if g.completedBy(i) != -1 {
		return false
	}
	p := g.players[g.playing]
	if k, ok := p.temp[i]; ok {
		return p.progress[i] > k
	}
	return int8(len(p.temp)) < g.numTempPaths
}

func (g GameCantStop) isValidAction(action []int8) bool {
	p := g.players[g.playing]
	for k, i := range action {
		if !g.isValidPath(i) {
			for m := k - 1; m >= 0; m-- {
				p.undoAction(action[m])
			}
			return false
		}
		p.takeAction(i)
	}
	for _, i := range action {
		p.undoAction(i)
	}
	return true
}

func (g GameCantStop) isWinner(p player) bool {
	return p.score() >= g.goal
}

func numsToString(nums []int8) string {
	strSlice := make([]string, len(nums))
	for i, num := range nums {
		strSlice[i] = strconv.Itoa(int(num))
	}
	result := strings.Join(strSlice, ", ")
	return result
}
