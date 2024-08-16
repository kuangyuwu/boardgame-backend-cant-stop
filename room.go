package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cant_stop"
)

type Room struct {
	mu       *sync.RWMutex
	host     *User
	id       string
	users    []*User
	toGame   chan Data
	fromGame chan Data
}

func (cfg *Config) newRoom(host *User) (*Room, error) {

	if cfg.hasTooManyRooms() {
		return nil, ErrTooManyRooms
	}

	id := ""
	for {
		id = randId()
		if cfg.findRoom(id) == nil {
			break
		}
	}

	newRoom := Room{
		mu:       &sync.RWMutex{},
		host:     host,
		id:       id,
		users:    []*User{},
		toGame:   nil,
		fromGame: nil,
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.rooms[&newRoom] = true

	return &newRoom, nil
}

func (cfg *Config) findRoom(roomId string) *Room {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	for r := range cfg.rooms {
		if r.id == roomId {
			return r
		}
	}
	return nil
}

func (cfg *Config) hasTooManyRooms() bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	result := len(cfg.rooms) >= MaxNumRooms
	return result
}

func randId() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	letters := []byte("ABCDEFGHJKLMNPQRSTUVWXYZ1234567890")
	id := make([]byte, 8)

	for i := range id {
		id[i] = letters[r.Intn(34)]
	}
	return string(id)
}

func (cfg *Config) deleteRoom(r *Room) error {
	if _, ok := cfg.rooms[r]; !ok {
		return ErrRoomNotExist
	}

	for _, u := range r.users {
		if !u.isHost() {
			u.leaveRoom()
			u.sendPrep()
		}
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	delete(cfg.rooms, r)

	return nil
}

func (r *Room) addUser(u *User) error {
	if r.hasTooManyUsers() {
		return ErrTooManyUsersInRoom
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	u.mu.Lock()
	defer u.mu.Unlock()

	r.users = append(r.users, u)
	u.room = r
	u.status = statusInPrep
	return nil
}

func (r Room) hasTooManyUsers() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := len(r.users) >= MaxNumUsersPerRoom
	return result
}

func (r Room) usernames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := []string{}
	for _, u := range r.users {
		result = append(result, u.username)
	}

	return result
}

func (r *Room) startGame() {

	toGame, fromGame, err := cantstop.StartGameCantStop(0, r.usernames())
	if err != nil {
		log.Printf("error starting game: %s", err)
		return
	}
	// log.Printf("Started game")

	r.toGame = toGame
	r.fromGame = fromGame

	for _, u := range r.users {
		u.mu.Lock()
		u.status = statusInGame
		u.isReady = false
		u.mu.Unlock()
	}

	go r.forwardToUsers()
}

func (r Room) forwardToGame(d Data) {
	r.toGame <- d
}

func (r Room) forwardToUsers() {
	for d := range r.fromGame {
		if d.Type == "exit" {
			u := r.findUser(d.Username)
			if u == nil {
				log.Printf("error finding user %s", d.Username)
				return
			}
			u.handleEndGame()
			continue
		}
		if d.Type == "terminate" {
			return
		}
		if d.Username != "" {
			u := r.findUser(d.Username)
			if u == nil {
				log.Printf("error finding user %s", d.Username)
				return
			}
			u.send <- d
		} else {
			for _, u := range r.users {
				u.send <- d
			}
		}
	}
}

func (r Room) findUser(username string) *User {
	for _, u := range r.users {
		if u.username == username {
			return u
		}
	}
	return nil
}

func (u *User) handleEndGame() {
	u.mu.Lock()
	u.status = statusInPrep
	u.mu.Unlock()
	if u.isHost() {
		u.isReady = true
	}
	u.room.broadcastPrepUpdate()
}
