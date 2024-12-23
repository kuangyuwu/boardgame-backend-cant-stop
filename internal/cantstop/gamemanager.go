package cantstop

import (
	"errors"
	"fmt"
	"sync"
)

type GameManager struct {
	mu   *sync.RWMutex
	game *Game
	DataSender
	tempMove    Move
	tempOptions [3][2][2]uint8
	phase
}

type DataSender interface {
	Send(data any)
}

type phase int8

const (
	rolling phase = iota
	advancing
	deciding
	gameover
)

var (
	ErrGameStartFailed = errors.New("failed to start a new game")
)

func New(numPlayer int, ds DataSender) (*GameManager, error) {
	g, err := NewGame(numPlayer)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGameStartFailed, err)
	}

	gm := &GameManager{
		mu:         new(sync.RWMutex),
		phase:      rolling,
		game:       g,
		DataSender: ds,
	}

	gm.sendStateUpdate()
	gm.Send(dataRoll())

	return gm, nil
}
