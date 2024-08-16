package main

import (
	"errors"
	"log"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MaxLenUsername     = 20
	MaxNumRooms        = 20
	MaxNumUsersTotal   = 10
	MaxNumUsersPerRoom = 5
)

var (
	ErrTooManyRooms       = errors.New("too many rooms")
	ErrTooManyUsers       = errors.New("too many users")
	ErrUsernameUsed       = errors.New("the username is used")
	ErrTooManyUsersInRoom = errors.New("too many users in the room")
	ErrUserNotExist       = errors.New("the user does not exist")
	ErrRoomNotExist       = errors.New("the room does not exist")
)

type Lobby struct {
	mu    *sync.Mutex
	rooms []*Room
	users []*User
}

func initializeLobby() *Lobby {
	return &Lobby{
		mu:    &sync.Mutex{},
		rooms: make([]*Room, 0, MaxNumRooms),
		users: make([]*User, 0, MaxNumUsersTotal),
	}
}

func (l *Lobby) createUser(conn *websocket.Conn) (*User, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.users) >= MaxNumUsersTotal {
		return nil, ErrTooManyUsers
	}

	u := &User{
		conn:     conn,
		lobby:    l,
		room:     nil,
		username: "",
		toUser:   make(chan Data),
	}
	l.users = append(l.users, u)
	return u, nil
}

func (l *Lobby) findUserByUsername(username string) *User {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, u := range l.users {
		if u.username == username {
			return u
		}
	}
	return nil
}

func (l *Lobby) deleteUser(u *User) {
	if u == nil {
		log.Printf("deleteUser: received nil User")
		return
	}

	l.mu.Lock()
	i := slices.Index(l.users, u)
	if i == -1 {
		log.Printf("deleteUser: user does not exist")
		return
	}
	l.users = slices.Delete(l.users, i, i+1)
	l.mu.Unlock()
}

func (l *Lobby) newRoom() (*Room, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.rooms) >= MaxNumRooms {
		return nil, ErrTooManyRooms
	}

	id := randId()
	for slices.IndexFunc(l.rooms, func(r *Room) bool { return r.id == id }) != -1 {
		id = randId()
	}

	r := &Room{
		mu:           &sync.RWMutex{},
		id:           id,
		players:      make([]RoomPlayer, 0, MaxNumUsersPerRoom),
		toGame:       nil,
		fromGame:     nil,
		indexRuleset: 0,
	}
	l.rooms = append(l.rooms, r)

	return r, nil
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

func (l *Lobby) findRoomById(roomId string) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, r := range l.rooms {
		if r.id == roomId {
			return r
		}
	}
	return nil
}

func (l *Lobby) deleteRoom(r *Room) {
	if r == nil {
		log.Printf("deleteRoom: received nil Room")
		return
	}

	l.mu.Lock()
	i := slices.Index(l.rooms, r)
	if i == -1 {
		log.Printf("deleteRoom: room does not exist")
		return
	}
	l.rooms = slices.Delete(l.rooms, i, i+1)
	l.mu.Unlock()

	log.Printf("deleted room %s", r.id)
}
