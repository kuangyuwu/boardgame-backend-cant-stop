package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

type Data struct {
	Type string                 `json:"type"`
	Body map[string]interface{} `json:"body"`
}

func (cfg *Config) handle(u *User) {
	defer u.conn.Close()
	defer cfg.deleteUser(u)

	for {
		_, msg, err := u.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %s\n", err)
			}
			break
		}

		data := Data{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Printf("Error unmarshaling JSON: %s", string(msg))
			continue
		}

		log.Printf("The server received the following data: %v", data)

		switch data.Type {
		case "ready":
			u.handleReady()
		case "prepNew":
			cfg.handlePrepNew(u)
		case "prepJoin":
			cfg.handlePrepJoin(u, data.Body)
		case "prepLeave":
			cfg.handlePrepLeave(u)
		case "prepReady":
			u.handlePrepReady()
		case "prepUnready":
			u.handlePrepUnready()
		case "start":
			u.handleStart()
		case "roll":
			u.handleRoll()
		case "action":
			u.handleAction(data.Body)
		case "continue":
			u.handleContinue()
		case "stop":
			u.handleStop()
		case "fail":
			u.handleFail()
		case "endGame":
			u.handleEndGame()
		default:
			log.Print("unsupported type")
		}
	}
}

func (u User) handleReady() {
	if !u.hasStatus(statusFree) {
		return
	}

	u.sendPrep()
}

func (cfg *Config) handlePrepNew(u *User) {
	if !u.hasStatus(statusFree) {
		return
	}

	r, err := cfg.newRoom(u)
	if err != nil {
		log.Printf("error creating new game: %s\n", err)
		u.sendError("error creating new game")
		u.sendPrep()
		return
	}

	err = r.addUser(u)
	if err != nil {
		log.Printf("error adding user to the game: %s", err)
		u.sendError("error joining the game")
		u.sendPrep()
		return
	}

	u.mu.Lock()
	u.isReady = true
	u.mu.Unlock()

	r.broadcastPrepUpdate()
}

func (cfg *Config) handlePrepJoin(u *User, body map[string]interface{}) {
	if !u.hasStatus(statusFree) {
		return
	}

	roomId, ok := body["roomId"].(string)
	if !ok {
		log.Print("invalid room ID")
		u.sendError("invalid room ID")
		u.sendPrep()
		return
	}

	r := cfg.findRoom(roomId)
	if r == nil {
		log.Print("invalid room ID")
		u.sendError("invalid room ID")
		u.sendPrep()
		return
	}

	err := r.addUser(u)
	if err != nil {
		log.Printf("error adding user to the game: %s", err)
		u.sendError("error joining the game")
		u.sendPrep()
		return
	}

	r.broadcastPrepUpdate()
}

func (cfg *Config) handlePrepLeave(u *User) {
	if !u.hasStatus(statusInPrep) {
		return
	}
	if u.isHost() {
		cfg.deleteRoom(u.room)
	}
	u.leaveRoom()
	u.sendPrep()
}

func (u *User) handlePrepReady() {
	if !u.hasStatus(statusInPrep) {
		return
	}

	u.mu.Lock()
	u.isReady = true
	u.mu.Unlock()

	u.room.broadcastPrepUpdate()
}

func (u *User) handlePrepUnready() {
	if !u.hasStatus(statusInPrep) {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	u.isReady = false
	go u.room.broadcastPrepUpdate()
}

func (u *User) handleStart() {
	if !u.hasStatus(statusInPrep) {
		return
	}

	u.room.mu.RLock()
	for _, user := range u.room.users {
		user.mu.RLock()
	}

	for _, user := range u.room.users {
		if !user.isReady {
			log.Print("error starting game")
			u.sendError("error starting game")
			go u.room.broadcastPrepUpdate()
			return
		}
	}
	u.room.mu.RUnlock()
	for _, user := range u.room.users {
		user.mu.RUnlock()
	}

	log.Print("Starting game")

	u.room.startGame()
	u.room.broadcastStart()
	u.room.broadcastLog("Game starts!")
	go u.room.nextTurn()
}

type Option struct {
	Grouping [][]int `json:"grouping"`
	Actions  [][]int `json:"actions"`
}

func (u *User) handleRoll() {
	if !u.hasStatus(statusInGameRolling) {
		return
	}
	u.mu.Lock()
	u.status = statusInGameChoosing
	u.mu.Unlock()
	points := rollDices(u.room.game.ruleset.dices)
	groupings := pointsToGroupings(points, u.room.game.ruleset.partitions)
	options := []Option{}
	hasOptions := false
	for _, grouping := range groupings {
		actions := u.groupingToActions(grouping)
		if len(actions) > 0 {
			hasOptions = true
		}
		options = append(options, Option{
			Grouping: grouping,
			Actions:  actions,
		})
	}
	u.sendResult(points, options, hasOptions)
}

func rollDices(dices []int) []int {
	result := make([]int, 0, len(dices))
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, d := range dices {
		result = append(result, rd.Intn(d)+1)
	}
	return result
}

func pointsToGroupings(points []int, partitions [][][]int) [][][]int {
	groupings := [][][]int{}
	for _, partition := range partitions {
		grouping := [][]int{}
		for _, part := range partition {
			group := []int{}
			for _, i := range part {
				group = append(group, points[i])
			}
			grouping = append(grouping, group)
		}
		groupings = append(groupings, grouping)
	}
	return groupings
}

func (u *User) groupingToActions(grouping [][]int) [][]int {
	actions := [][]int{}
	g0 := sum(grouping[0])
	g1 := sum(grouping[1])
	if u.room.isValidAction(u, []int{g0, g1}) {
		return [][]int{{g0, g1}}
	}
	if u.room.isValidAction(u, []int{g0}) {
		actions = append(actions, []int{g0})
	}
	if u.room.isValidAction(u, []int{g1}) {
		actions = append(actions, []int{g1})
	}
	return actions
}

func sum(slice []int) int {
	result := 0
	for _, x := range slice {
		result += x
	}
	return result
}

func (u *User) handleAction(body map[string]interface{}) {
	if !u.hasCorrectStatus(statusInGameChoosing) {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	for _, path := range body["action"].([]interface{}) {
		u.player.takeAction(int(path.(float64)))
	}
	u.room.broadcastGameboard()
	u.sendContinue()
}

func (u *User) handleContinue() {
	if !u.hasCorrectStatus(statusInGameChoosing) {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	u.status = statusInGameRolling

	go u.room.nextMove(false)
}

func (u *User) handleStop() {
	if !u.hasCorrectStatus(statusInGameChoosing) {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	u.player.updateState()
	u.player.resetTemp()
	u.player.addMoves(u.room.game.moveCount)
	u.room.broadcastGameboard()
	u.status = statusInGameNotPlaying
	if u.player.isWinner(u.room.game.ruleset.goal) {
		u.room.broadcastWinner(u.username)
		return
	}
	u.room.broadcastPlayer(false)
	go u.room.nextPlayer()
}

func (u *User) handleFail() {
	if !u.hasCorrectStatus(statusInGameChoosing) {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	u.player.resetTemp()
	u.player.addMoves(u.room.game.moveCount)
	u.room.broadcastGameboard()
	u.status = statusInGameNotPlaying
	u.room.broadcastPlayer(false)
	go u.room.nextPlayer()
}

func (u *User) hasCorrectStatus(expected Status) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.status != expected {
		log.Printf("error: expect status %d, get status %d", expected, u.status)
		return false
	}
	return true
}

func (u *User) handleEndGame() {
	u.mu.Lock()
	u.status = statusInPrep
	u.mu.Unlock()

	u.room.broadcastPrepUpdate()
}
