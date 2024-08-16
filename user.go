package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type User struct {
	conn     *websocket.Conn
	lobby    *Lobby
	room     *Room
	username string
	toUser   chan Data
}

func (u *User) disconnect() {
	if u.lobby != nil {
		u.lobby.deleteUser(u)
	}
	if u.room != nil {
		u.room.removePlayer(u.username)
		if len(u.room.players) == 0 {
			u.lobby.deleteRoom(u.room)
		}
	}
	u.conn.Close()
	log.Printf("User %s disconnected", u.username)
}

func (u *User) handleMessage() {
	defer u.disconnect()
	for {
		_, msg, err := u.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %s\n", err)
			}
			return
		}

		data := Data{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Printf("error unmarshaling JSON: %s", string(msg))
			continue
		}

		log.Printf("The server received the following data from %s: %v", u.username, data)
		data.Username = u.username

		switch data.Type {
		case "ready":
			u.handleReady()
		case "username":
			u.handleUsername(data.Body)
		case "prepNew":
			u.handlePrepNew()
		case "prepJoin":
			u.handlePrepJoin(data.Body)
		case "prepLeave":
			u.handlePrepLeave()
		case "ruleset":
			u.handleRuleset(data.Body)
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

func (u *User) sendMessage() {
	for data := range u.toUser {
		msg, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshalling JSON %s", err)
			return
		}
		u.conn.WriteMessage(websocket.TextMessage, msg)
	}
}
