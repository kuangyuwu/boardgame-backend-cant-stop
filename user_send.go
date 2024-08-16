package main

import cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cant_stop"

type Data = cantstop.Data

func (u User) sendError(errMsg string) {
	data := Data{
		Type: "error",
		Body: map[string]interface{}{
			"error": errMsg,
		},
	}
	u.toUser <- data
}

func (u *User) sendUsername() {
	data := Data{
		Type: "username",
		Body: nil,
	}
	u.toUser <- data
}

func (u User) sendPrep() {
	data := Data{
		Type: "prep",
		Body: nil,
	}
	u.toUser <- data
}
