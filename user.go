package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Status int

const (
	statusFree             Status = 0
	statusInPrep           Status = 1
	statusInGameNotPlaying Status = 2
	statusInGameRolling    Status = 3
	statusInGameChoosing   Status = 4
)

type User struct {
	mu       *sync.RWMutex
	conn     *websocket.Conn
	room     *Room
	player   *Player
	isReady  bool
	status   Status
	username string
	send     chan Data
}

func (cfg *Config) createUser(username string, conn *websocket.Conn) (*User, error) {

	if cfg.hasTooManyUsers() {
		return nil, ErrTooManyUsers
	}
	if cfg.findUser(username) != nil {
		return nil, ErrUsernameUsed
	}

	newUser := User{
		conn:     conn,
		mu:       &sync.RWMutex{},
		room:     nil,
		isReady:  false,
		status:   0,
		username: username,
		send:     make(chan Data),
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.users[&newUser] = true

	return &newUser, nil
}

func (cfg *Config) deleteUser(u *User) error {
	if _, ok := cfg.users[u]; !ok {
		return ErrUserNotExist
	}

	log.Printf("deleting user %s", u.username)

	if u.room != nil {
		if u.isHost() {
			cfg.deleteRoom(u.room)
		}
		u.leaveRoom()
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	delete(cfg.users, u)

	return nil
}

func (cfg *Config) findUser(username string) *User {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	for u := range cfg.users {
		if u.username == username {
			return u
		}
	}
	return nil
}

func (cfg *Config) hasTooManyUsers() bool {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	result := len(cfg.users) >= MaxNumUsersTotal
	return result
}

func (u *User) hasStatus(status Status) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	result := u.status == status
	return result
}

func (u *User) isHost() bool {
	u.mu.RLock()
	u.room.mu.RLock()
	defer u.mu.RUnlock()
	defer u.room.mu.RUnlock()

	result := u == u.room.host
	return result
}

func (u *User) leaveRoom() {
	u.mu.Lock()
	u.room.mu.Lock()
	defer u.mu.Unlock()
	defer u.room.mu.Unlock()

	newUsers := []*User{}
	for _, user := range u.room.users {
		if user == u {
			continue
		}
		newUsers = append(newUsers, user)
	}
	u.room.users = newUsers

	u.room = nil
	u.isReady = false
	u.status = statusFree
}
