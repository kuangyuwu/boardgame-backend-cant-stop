package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func (u *User) sendMessage() {
	for data := range u.send {
		msg, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshalling JSON %s", err)
			return
		}
		u.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (u User) sendError(errMsg string) {
	data := Data{
		Type: "error",
		Body: map[string]interface{}{
			"error": errMsg,
		},
	}
	u.send <- data
}

func (u User) sendPrep() {
	data := Data{
		Type: "prep",
		Body: nil,
	}
	u.send <- data
}

func (u *User) sendPrepUpdate() {
	data := Data{
		Type: "prepUpdate",
		Body: map[string]interface{}{
			"roomId":    u.room.id,
			"isHosting": false,
			"isReady":   u.isReady,
			"usernames": u.room.usernames(),
		},
	}
	if u.isHost() {
		data.Body["isHosting"] = true
		for _, user := range u.room.users {
			if !user.isReady {
				data.Body["isReady"] = false
			}
		}
	}
	u.send <- data
}

func (u User) sendRoll() {
	data := Data{
		Type: "roll",
		Body: nil,
	}
	u.send <- data
}

func (u User) sendResult(points []int, options []Option, hasOptions bool) {
	data := Data{
		Type: "result",
		Body: map[string]interface{}{
			"points":     points,
			"options":    options,
			"hasOptions": hasOptions,
		},
	}
	u.send <- data
}

func (u User) sendContinue() {
	data := Data{
		Type: "continue",
		Body: nil,
	}
	u.send <- data
}
