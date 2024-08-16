package main

import (
	"log"
)

func (u *User) handleReady() {
	u.sendUsername()
}

func (u *User) handleUsername(body map[string]interface{}) {
	username := body["username"].(string)
	if u.lobby.findUserByUsername(username) != nil {
		u.sendUsername()
		return
	}
	u.username = username
	u.sendPrep()
}

func (u *User) handlePrepNew() {
	if u.room != nil {
		log.Printf("handlePrepNew: %s is already in room %s", u.username, u.room.id)
		u.room.broadcastPrepUpdate()
		return
	}

	r, err := u.lobby.newRoom()
	if err != nil {
		log.Printf("handlePrepNew: error creating new room: %s\n", err)
		u.sendError("error creating new room")
		u.sendPrep()
		return
	}

	u.room = r
	r.addPlayer(u)
}

func (u *User) handlePrepJoin(body map[string]interface{}) {
	if u.room != nil {
		log.Printf("handlePrepJoin: %s is already in room %s", u.username, u.room.id)
		u.room.broadcastPrepUpdate()
		return
	}

	roomId, ok := body["roomId"].(string)
	if !ok {
		log.Print("handlePrepJoin: invalid room ID")
		u.sendError("invalid room ID")
		u.sendPrep()
		return
	}

	r := u.lobby.findRoomById(roomId)
	if r == nil {
		log.Print("handlePrepJoin: room not found")
		u.sendError("room not found")
		u.sendPrep()
		return
	}

	err := r.addPlayer(u)
	if err != nil {
		log.Printf("handlePrepJoin: error adding user to the room: %s", err)
		u.sendError("error joining the room")
		u.sendPrep()
		return
	}
	u.room = r
}

func (u *User) handlePrepLeave() {
	if u.room == nil {
		log.Printf("handlePrepLeave: %s is already not in any room", u.username)
		u.sendPrep()
		return
	}
	u.room.removePlayer(u.username)
	u.room = nil
	u.sendPrep()
}

func (u *User) handleRuleset(body map[string]interface{}) {
	if u.room == nil {
		log.Printf("handlePrepReady: %s is not in any room", u.username)
		u.sendPrep()
		return
	}
	i := int(body["ruleset"].(float64))
	u.room.setIndexRuleset(i)
}

func (u *User) handlePrepReady() {
	if u.room == nil {
		log.Printf("handlePrepReady: %s is not in any room", u.username)
		u.sendPrep()
		return
	}
	u.room.setReady(u.username)
}

func (u *User) handlePrepUnready() {
	if u.room == nil {
		log.Printf("handlePrepUnready: %s is not in any room", u.username)
		u.sendPrep()
		return
	}
	u.room.setUnready(u.username)
}

func (u *User) handleStart() {
	if u.room == nil {
		log.Printf("handlePrepUnready: %s is not in any room", u.username)
		u.sendPrep()
		return
	}
	if u.room.indexPlayer(u.username) != 0 {
		log.Printf("handlePrepUnready: %s is not the host", u.username)
		u.room.broadcastPrepUpdate()
		return
	}
	if !u.room.isAllReady() {
		log.Printf("handlePrepUnready: not everyone is ready")
		u.room.broadcastPrepUpdate()
		return
	}

	u.room.startGame()
}
