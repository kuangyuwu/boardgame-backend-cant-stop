package room

import (
	"errors"
	"log"
	"slices"
	"sync"

	cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cant_stop"
)

type Data = cantstop.Data

type Room struct {
	mu           *sync.RWMutex
	Id           string
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

const (
	MaxNumUsersPerRoom = 5
)

var (
	ErrTooManyUsersInRoom = errors.New("too many users in the room")
)

func New(id string) *Room {
	return &Room{
		mu:           &sync.RWMutex{},
		Id:           id,
		players:      make([]RoomPlayer, 0, MaxNumUsersPerRoom),
		toGame:       nil,
		fromGame:     nil,
		indexRuleset: 0,
	}
}

func (r *Room) AddPlayer(username string, toUser chan Data) error {
	if len(r.players) >= MaxNumUsersPerRoom {
		return ErrTooManyUsersInRoom
	}

	r.mu.Lock()
	r.players = append(r.players, RoomPlayer{
		username: username,
		toUser:   toUser,
		isReady:  false,
		isInGame: false,
	})
	r.mu.Unlock()

	r.BroadcastPrepUpdate()
	return nil
}

func (r *Room) RemovePlayer(username string) {
	r.mu.Lock()
	i := r.IndexPlayer(username)
	if i == -1 {
		log.Printf("removePlayer: %s is already not in the room", username)
		return
	}
	r.players = slices.Delete(r.players, i, i+1)
	r.mu.Unlock()

	r.BroadcastPrepUpdate()
}

func (r *Room) SetIndexRuleset(i int) {
	r.mu.Lock()
	r.indexRuleset = i
	r.mu.Unlock()
	r.BroadcastPrepUpdate()
}

func (r *Room) SetReady(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isReady = true
		}
	}
	r.mu.Unlock()
	r.BroadcastPrepUpdate()
}

func (r *Room) SetUnready(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isReady = false
		}
	}
	r.mu.Unlock()
	r.BroadcastPrepUpdate()
}

func (r Room) BroadcastPrepUpdate() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i, p := range r.players {
		if p.isInGame {
			continue
		}
		data := Data{
			Type: "prepUpdate",
			Body: map[string]interface{}{
				"roomId":    r.Id,
				"isHosting": false,
				"isReady":   p.isReady,
				"usernames": r.usernames(),
				"ruleset":   r.indexRuleset,
			},
		}
		if i == 0 {
			data.Body["isHosting"] = true
			data.Body["isReady"] = r.IsAllReady()
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

func (r Room) IsAllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i, p := range r.players {
		if i != 0 && !p.isReady {
			return false
		}
	}
	return true
}

func (r Room) IndexPlayer(username string) int {
	return slices.IndexFunc(r.players, func(p RoomPlayer) bool { return p.username == username })
}

func (r Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.players) == 0
}

func (r *Room) ExitGame(username string) {
	r.mu.Lock()
	for i, p := range r.players {
		if p.username == username {
			r.players[i].isInGame = false
		}
	}
	r.mu.Unlock()
	r.BroadcastPrepUpdate()
}
