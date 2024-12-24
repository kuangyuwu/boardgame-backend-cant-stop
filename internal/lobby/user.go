package lobby

import (
	"encoding/json"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

type User struct {
	lobby    *Lobby
	room     *Room
	id       string
	username string
	out      chan<- []byte
}

func NewUser(l *Lobby, id string, out chan<- []byte) *User {
	return &User{
		lobby:    l,
		room:     nil,
		id:       id,
		username: "",
		out:      out,
	}
}

func (u User) Username() string {
	return u.username
}

func (u *User) Send(data any) {
	msg, err := json.Marshal(data)
	if err != nil {
		clog.Errorf("Error marshalling data %v: %v", data, err)
		return
	}
	u.out <- msg
}

func (u User) ListenAndHandleMsg(in <-chan []byte) {
	for {
		msg, ok := <-in
		if !ok {
			clog.Infof("in channel closed for user %s (username: %s)", u.id, u.username)
			break
		}
		clog.Infof("received message from user %s (username: %s): %s", u.id, u.username, string(msg))

		data := Data{}
		err := json.Unmarshal(msg, &data)
		if err != nil {
			clog.Warnf("error unmarshaling message: %v", err)
			continue
		}

		switch data.Type {
		case "start":
			u.handleStart()
		case "username":
			u.handleUsername(data.Body)
		case "newRoom":
			u.handleNewRoom()
		case "joinRoom":
			u.handleJoinRoom(data.Body)
		case "leaveRoom":
			u.handleLeaveRoom()
		// case "ruleset":
		// 	u.handleRuleset(body)
		case "ready":
			u.handleReady()
		case "unready":
			u.handleUnready()
		// case "start":
		// 	u.handleStart()
		// case "roll":
		// 	u.room.ForwardToGame(data)
		// case "act":
		// 	u.room.ForwardToGame(data)
		// case "confirm":
		// 	u.room.ForwardToGame(data)
		// case "exit":
		// 	u.room.ForwardToGame(data)
		default:
			clog.Warn("unsupported type")
		}
	}
	u.disconnect()
}

func (u *User) disconnect() {
	if u.room != nil {
		u.room.RemoveMember(u.username)
		if u.room.IsEmpty() {
			u.lobby.deleteRoom(u.room)
		}
	}
	if u.lobby != nil {
		u.lobby.deleteUser(u)
	}
	clog.Infof("user %s disconnected", u.username)
}

func (u *User) handleStart() {
	u.Send(dataUsername())
}

func (u *User) handleUsername(body any) {
	username, ok := body.(string)
	if !ok {
		clog.Error("invalid username data")
		return
	}
	if u.lobby.findUserByUsername(username) != nil {
		clog.Warn("username used")
		u.Send(dataWarn("username used"))
		u.Send(dataUsername())
		return
	}
	u.username = username
	u.Send(dataPrep())
}

func (u *User) handleNewRoom() {
	if u.room != nil {
		clog.Errorf("%s is already in room %s", u.username, u.room.Id)
		u.room.BroadcastPrepUpdate()
		return
	}

	r, err := u.lobby.CreateRoom()
	if err != nil {
		clog.Errorf("error creating new room: %s\n", err)
		u.Send(dataError("error creating new room"))
		u.Send(dataPrep())
		return
	}

	u.room = r
	r.AddMember(u)
}

func (u *User) handleJoinRoom(body any) {
	if u.room != nil {
		clog.Errorf("%s is already in room %s", u.username, u.room.Id)
		u.room.BroadcastPrepUpdate()
		return
	}

	roomId, ok := body.(string)
	if !ok {
		clog.Error("invalid prepJoin data")
		u.Send(dataError("invalid room ID"))
		u.Send(dataPrep())
		return
	}

	r := u.lobby.findRoomById(roomId)
	if r == nil {
		clog.Warn("room not found")
		u.Send(dataWarn("room not found"))
		u.Send(dataPrep())
		return
	}

	err := r.AddMember(u)
	if err != nil {
		clog.Warnf("error adding user to the room: %s", err)
		u.Send(dataWarn("error joining the room"))
		u.Send(dataPrep())
		return
	}
	u.room = r
}

func (u *User) handleLeaveRoom() {
	if u.room == nil {
		clog.Errorf("%s is already not in any room", u.username)
		u.Send(dataPrep())
		return
	}
	u.room.RemoveMember(u.username)
	u.room = nil
	u.Send(dataPrep())
}

// func (u *User) handleRuleset(body map[string]interface{}) {
// 	if u.room == nil {
// 		log.Printf("handlePrepReady: %s is not in any room", u.username)
// 		u.send(dataPrep())
// 		return
// 	}
// 	i := int(body["ruleset"].(float64))
// 	u.room.SetIndexRuleset(i)
// }

func (u *User) handleReady() {
	if u.room == nil {
		clog.Errorf("%s is not in any room", u.username)
		u.Send(dataPrep())
		return
	}
	u.room.SetReady(u.username)
}

func (u *User) handleUnready() {
	if u.room == nil {
		clog.Errorf("%s is not in any room", u.username)
		u.Send(dataPrep())
		return
	}
	u.room.SetUnready(u.username)
}

// func (u *User) handleStart() {
// 	if u.room == nil {
// 		clog.Errorf("%s is not in any room", u.username)
// 		u.Send(dataPrep())
// 		return
// 	}
// 	if u.room.IndexPlayer(u.username) != 0 {
// 		clog.Errorf("%s is not the host", u.username)
// 		u.room.BroadcastPrepUpdate()
// 		return
// 	}
// 	if !u.room.IsAllReady() {
// 		clog.Warnf("not everyone is ready")
// 		u.room.BroadcastPrepUpdate()
// 		return
// 	}

// 	// u.room.StartGame()
// }
