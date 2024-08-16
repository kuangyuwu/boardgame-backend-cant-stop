package cantstop

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type GameCantStop struct {
	mu         *sync.Mutex
	toGame     chan Data
	fromGame   chan Data
	turnCount  int16
	moveCount  int16
	playing    int8
	phase      phase
	failed     bool
	terminated bool
	ended      bool
	players    []player
	RuleSet
}

type phase int8

const (
	phaseRoll    phase = 0
	phaseAct     phase = 1
	phaseConfirm phase = 2
)

func StartGameCantStop(indexRuleSet int, usernames []string) (toGame, fromGame chan Data, err error) {
	ruleSet, err := getRuleSet(indexRuleSet)
	if err != nil {
		return nil, nil, err
	}

	players := make([]player, 0, len(usernames))
	for _, username := range usernames {
		players = append(players, newPlayer(username, ruleSet.pathLengths))
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rd.Shuffle(len(players), func(i, j int) { players[i], players[j] = players[j], players[i] })

	g := GameCantStop{
		mu:        &sync.Mutex{},
		toGame:    make(chan Data),
		fromGame:  make(chan Data),
		turnCount: 0,
		playing:   0,
		moveCount: 0,
		phase:     phaseRoll,
		players:   players,
		RuleSet:   ruleSet,
	}
	go g.run()
	return g.toGame, g.fromGame, nil
}

func (g *GameCantStop) run() {
	g.mu.Lock()
	g.broadcast(dataStart(g.usernames(), g.pathLengths))
	g.announce("Game starts!")
	g.nextTurn()
	for {
		g.mu.Lock()

		if g.terminated || g.allPlayerLeft() {
			g.broadcast(dataTerminate())
			return
		}

		data, ok := <-g.toGame
		if !ok {
			g.logErrorAndTerminate("channel toGame closed unexpectedly")
			return
		}
		if data.Type == "exit" {
			g.handleExit(data.Username)
			continue
		}
		if data.Username != g.players[g.playing].username {
			logError(fmt.Sprintf("received unexpected message from %s", data.Username))
			g.mu.Unlock()
			continue
		}

		switch data.Type {
		case "roll":
			g.handleRoll()
		case "act":
			g.handleAct(data.Body)
		case "confirm":
			g.handleConfirm(data.Body)
		}
	}
}

func logError(errMsg string) {
	log.Println(errMsg)
}

func (g *GameCantStop) logErrorAndTerminate(errMsg string) {
	content := "game terminated: " + errMsg
	log.Println(content)
	g.announce(content)
	g.terminated = true
}

func (g GameCantStop) usernames() []string {
	result := make([]string, len(g.players))
	for n, p := range g.players {
		result[n] = p.username
	}
	return result
}

func (g GameCantStop) allPlayerLeft() bool {
	result := true
	for _, p := range g.players {
		if !p.left {
			result = false
		}
	}
	return result
}
