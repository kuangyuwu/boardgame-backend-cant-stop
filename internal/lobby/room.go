package lobby

import (
	"slices"
	"sync"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

const (
	maxNumMembers = 5
)

type Room struct {
	mu       *sync.RWMutex
	Id       string
	members  []Member
	toGame   chan Data
	fromGame chan Data
	// indexRuleset int
}

type Member struct {
	user
	isReady  bool
	isInGame bool
}

type user interface {
	Username() string
	Send(data any)
}

func NewRoom(id string) *Room {
	return &Room{
		mu:       &sync.RWMutex{},
		Id:       id,
		members:  make([]Member, 0, maxNumMembers),
		toGame:   nil,
		fromGame: nil,
		// indexRuleset: 0,
	}
}

func (r *Room) AddMember(u user) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.members) >= maxNumMembers {
		return ErrTooManyUsersInRoom
	}

	r.members = append(r.members, Member{
		user:     u,
		isReady:  false,
		isInGame: false,
	})

	r.broadcastPrepUpdate()

	return nil
}

func (r *Room) RemoveMember(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexMember(username)
	if i == -1 {
		clog.Errorf("%s is already not in the room", username)
		return
	}
	r.members = slices.Delete(r.members, i, i+1)

	r.broadcastPrepUpdate()
}

// func (r *Room) SetIndexRuleset(i int) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	r.indexRuleset = i
// 	r.broadcastPrepUpdate()
// }

func (r *Room) SetReady(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexMember(username)
	if i == -1 {
		clog.Errorf("%s is already not in the room", username)
		return
	}
	r.members[i].isReady = true

	r.broadcastPrepUpdate()
}

func (r *Room) SetUnready(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexMember(username)
	if i == -1 {
		clog.Errorf("%s is already not in the room", username)
		return
	}
	r.members[i].isReady = false

	r.broadcastPrepUpdate()
}

func (r *Room) BroadcastPrepUpdate() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.broadcastPrepUpdate()
}

func (r *Room) broadcastPrepUpdate() {

	for i, m := range r.members {
		if m.isInGame {
			continue
		}
		data := Data{
			Type: "prepUpdate",
			Body: map[string]any{
				"roomId":    r.Id,
				"isHosting": false,
				"isReady":   m.isReady,
				"usernames": r.usernames(),
				// "ruleset":   r.indexRuleset,
			},
		}
		if i == 0 {
			data.Body.(map[string]any)["isHosting"] = true
			data.Body.(map[string]any)["isReady"] = r.isAllReady()
		}
		m.Send(data)
	}
}

func (r Room) usernames() []string {
	result := make([]string, len(r.members))
	for i, m := range r.members {
		result[i] = m.Username()
	}
	return result
}

func (r Room) isAllReady() bool {
	for i, p := range r.members {
		if i != 0 && !p.isReady {
			return false
		}
	}
	return true
}

func (r Room) indexMember(username string) int {
	return slices.IndexFunc(r.members, func(m Member) bool { return m.Username() == username })
}

func (r Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.members) == 0
}

func (r *Room) ExitGame(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexMember(username)
	if i == -1 {
		clog.Errorf("%s is already not in the room", username)
		return
	}
	r.members[i].isInGame = false

	r.mu.Unlock()
	r.BroadcastPrepUpdate()
}
