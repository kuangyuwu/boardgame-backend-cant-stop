package main

import (
	"errors"
	"log"
	"slices"
	"sync"
)

type Room struct {
	mu           *sync.RWMutex
	id           string
	players      []RoomPlayer
	toGame       chan Data
	fromGame     chan Data
	indexRuleset int
}

type RoomPlayer struct {
	username string
	toUser   chan Data
	isReady  bool
	isInGame bool
}

func (r *Room) addPlayer(u *User) error {
	if u == nil {
		log.Printf("addPlayer: received nil User")
		return errors.New("received nil User")
	}
	if len(r.players) >= MaxNumUsersPerRoom {
		return ErrTooManyUsersInRoom
	}

	r.mu.Lock()
	r.players = append(r.players, RoomPlayer{
		username: u.username,
		toUser:   u.toUser,
		isReady:  false,
		isInGame: false,
	})
	r.mu.Unlock()

	r.broadcastPrepUpdate()
	return nil
}

func (r *Room) removePlayer(username string) {
	r.mu.Lock()
	i := r.indexPlayer(username)
	if i == -1 {
		log.Printf("removePlayer: %s is already not in the room", username)
		return
	}
	r.players = slices.Delete(r.players, i, i+1)
	r.mu.Unlock()

	r.broadcastPrepUpdate()
}

func (r *Room) setIndexRuleset(i int) {
	r.mu.Lock()
	r.indexRuleset = i
	r.mu.Unlock()
	r.broadcastPrepUpdate()
}

func (r *Room) setReady(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isReady = true
		}
	}
	r.mu.Unlock()
	r.broadcastPrepUpdate()
}

func (r *Room) setUnready(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isReady = false
		}
	}
	r.mu.Unlock()
	r.broadcastPrepUpdate()
}

func (r Room) broadcastPrepUpdate() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i, p := range r.players {
		if p.isInGame {
			continue
		}
		data := Data{
			Type: "prepUpdate",
			Body: map[string]interface{}{
				"roomId":    r.id,
				"isHosting": false,
				"isReady":   p.isReady,
				"usernames": r.usernames(),
				"ruleset":   r.indexRuleset,
			},
		}
		if i == 0 {
			data.Body["isHosting"] = true
			data.Body["isReady"] = r.isAllReady()
		}
		p.toUser <- data
	}
}

func (r Room) usernames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.players))
	for i, u := range r.players {
		result[i] = u.username
	}
	return result
}

func (r Room) isAllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i, p := range r.players {
		if i != 0 && !p.isReady {
			return false
		}
	}
	return true
}

func (r Room) indexPlayer(username string) int {
	return slices.IndexFunc(r.players, func(p RoomPlayer) bool { return p.username == username })
}

func (r *Room) exitGame(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isInGame = false
		}
	}
	r.mu.Unlock()
	r.broadcastPrepUpdate()
}

// func (cfg *Config) hasTooManyRooms() bool {
// 	cfg.mu.RLock()
// 	defer cfg.mu.RUnlock()

// 	result := len(cfg.rooms) >= MaxNumRooms
// 	return result
// }

// func (r Room) hasTooManyUsers() bool {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	result := len(r.users) >= MaxNumUsersPerRoom
// 	return result
// }
