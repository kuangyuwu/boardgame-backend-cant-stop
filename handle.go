package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cant_stop"
)

type Data = cantstop.Data

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

		// log.Printf("The server received the following data: %v", data)
		data.Username = u.username

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
			u.room.forwardToGame(data)
		case "act":
			u.room.forwardToGame(data)
		case "confirm":
			u.room.forwardToGame(data)
		case "exit":
			u.room.forwardToGame(data)
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

	// log.Print("Starting game")

	u.room.startGame()
}

// func (u *User) hasCorrectStatus(expected Status) bool {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()

// 	if u.status != expected {
// 		log.Printf("error: expect status %d, get status %d", expected, u.status)
// 		return false
// 	}
// 	return true
// }
