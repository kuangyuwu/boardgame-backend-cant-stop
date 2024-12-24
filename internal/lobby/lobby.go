package lobby

import (
	"errors"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

const (
	maxLenUsername = 20
	maxNumRooms    = 20
	maxNumUsers    = 10

	lenUserID = 6
	lenRoomID = 5
	letters   = "ABCDEFGHJKLMNPQRSTUVWXYZ1234567890"
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

func New() *Lobby {
	return &Lobby{
		mu:    &sync.Mutex{},
		rooms: make([]*Room, 0, maxNumRooms),
		users: make([]*User, 0, maxNumUsers),
	}
}

func (l *Lobby) Connect(in <-chan []byte, out chan<- []byte, dc <-chan struct{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.users) >= maxNumUsers {
		return ErrTooManyUsers
	}

	id := l.generateUserID()
	u := NewUser(l, id, out)
	go u.ListenAndHandleMsg(in)
	l.users = append(l.users, u)

	return nil
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
		clog.Errorf("received nil User")
		return
	}

	l.mu.Lock()
	i := slices.Index(l.users, u)
	if i == -1 {
		clog.Errorf("user does not exist")
		return
	}
	l.users = slices.Delete(l.users, i, i+1)
	l.mu.Unlock()
}

func (l *Lobby) CreateRoom() (*Room, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.rooms) >= maxNumRooms {
		return nil, ErrTooManyRooms
	}

	id := l.generateRoomID()
	r := NewRoom(id)
	l.rooms = append(l.rooms, r)

	return r, nil
}

func (l *Lobby) findRoomById(roomId string) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, r := range l.rooms {
		if r.Id == roomId {
			return r
		}
	}
	return nil
}

func (l *Lobby) deleteRoom(r *Room) {
	if r == nil {
		clog.Error("received nil Room")
		return
	}

	l.mu.Lock()
	i := slices.Index(l.rooms, r)
	if i == -1 {
		clog.Error("room does not exist")
		return
	}
	l.rooms = slices.Delete(l.rooms, i, i+1)
	l.mu.Unlock()

	clog.Infof("deleted room %s", r.Id)
}

func (l *Lobby) generateUserID() string {
	id := randId(6)
	for slices.IndexFunc(l.users, func(u *User) bool { return u.id == id }) != -1 {
		id = randId(6)
	}
	return id
}

func (l *Lobby) generateRoomID() string {
	id := randId(lenRoomID)
	for slices.IndexFunc(l.rooms, func(r *Room) bool { return r.Id == id }) != -1 {
		id = randId(lenRoomID)
	}
	return id
}

func randId(L int) string {
	id := make([]byte, L)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range id {
		id[i] = letters[r.Intn(34)]
	}
	return string(id)
}
