package lobby

import (
	"math/rand"
	"time"

	cantstop "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/cantstop"
	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isAllReady() {
		r.broadcastPrepUpdate()
		return ErrNotReady
	}

	n := len(r.members)
	order := make([]uint8, n)
	for i := range n {
		order[i] = uint8(i + 1)
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rd.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	usernames := make([]string, n)
	for i := range n {
		r.members[i].idx = order[i]
		usernames[order[i]-1] = r.members[i].Username()
	}
	r.Send(dataUsernames(usernames))

	gm, err := cantstop.New(len(r.members), r)
	if err != nil {
		return err
	}

	r.gm = gm
	clog.Infof("game started in room %s", r.Id)

	for i := range r.members {
		r.members[i].isReady = false
		r.members[i].isInGame = true
	}

	return nil
}

func (r *Room) Send(data any) {
	for _, m := range r.members {
		m.Send(data)
	}
}

func (r Room) ForwardToGame(username, dataType string, body any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.indexMember(username)
	if i == -1 {
		return ErrMemberNotExist
	}

	if r.gm == nil {
		clog.Error("there is no game running")
		return ErrNoOngoingGame
	}
	return r.gm.HandleData(int(r.members[i].idx), dataType, body)
}
