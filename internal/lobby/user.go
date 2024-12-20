package lobby

import (
	"encoding/json"
	"log"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

type User struct {
	lobby    *Lobby
	room     *Room
	username string
	out      chan<- []byte
}

func (u *User) disconnect() {
	if u.lobby != nil {
		u.lobby.deleteUser(u)
	}
	if u.room != nil {
		u.room.RemovePlayer(u.username)
		if u.room.IsEmpty() {
			u.lobby.deleteRoom(u.room)
		}
	}
	close(u.out)
	log.Printf("User %s disconnected", u.username)
}

func (u *User) handleMessage(msg []byte) {
	data := Data{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Printf("error unmarshaling JSON: %s", string(msg))
		return
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
		u.room.ForwardToGame(data)
	case "act":
		u.room.ForwardToGame(data)
	case "confirm":
		u.room.ForwardToGame(data)
	case "exit":
		u.room.ForwardToGame(data)
	default:
		log.Print("unsupported type")
	}
}

func (u *User) send(data any) {
	msg, err := json.Marshal(data)
	if err != nil {
		clog.Errorf("Error marshalling data %v: %v", data, err)
		return
	}
	u.out <- msg
}
