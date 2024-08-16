package main

import (
	"log"

	cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cant_stop"
)

func (r *Room) startGame() {
	toGame, fromGame, err := cantstop.StartGameCantStop(r.indexRuleset, r.usernames())
	if err != nil {
		log.Printf("error starting game: %s", err)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.toGame = toGame
	r.fromGame = fromGame
	log.Printf("Started")

	for i := range r.players {
		r.players[i].isReady = false
		r.players[i].isInGame = true
	}

	go r.forwardToUsers()
}

func (r Room) forwardToGame(d Data) {
	r.toGame <- d
}

func (r Room) forwardToUsers() {
	for d := range r.fromGame {
		if d.Type == "exit" {
			r.exitGame(d.Username)
			continue
		}
		if d.Type == "terminate" {
			return
		}
		if d.Username != "" {
			for _, p := range r.players {
				if p.username == d.Username {
					p.toUser <- d
				}
			}
		} else {
			for _, p := range r.players {
				p.toUser <- d
			}
		}
	}
}
