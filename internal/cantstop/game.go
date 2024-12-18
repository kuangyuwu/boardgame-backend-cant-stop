package cantstop

import (
	"errors"
	"fmt"
	"sync"
)

const (
	numNeutral = 3
	goal       = 3
)

var (
	diceType = [...]uint8{6, 6, 6, 6}
	pathLens = []int8{-1, -1, 3, 5, 7, 9, 11, 13, 11, 9, 7, 5, 3}
	// partitions = [][][]int{{{0, 1}, {2, 3}}, {{0, 2}, {1, 3}}, {{0, 3}, {1, 2}}},
	// actionGenerator: actionGenerator2Groups,
)

var (
	ErrInvalidNumPlayer = errors.New("numPlayer must be between 1 and 5")

	ErrGameOngoing    = errors.New("the game is still ongoing")
	ErrGameConcluded  = errors.New("the game has already concluded")
	ErrGameTerminated = errors.New("the game is already terminated")

	ErrInvalidMove          = errors.New("the move is invalid")
	ErrNotThisPlayersTurn   = errors.New("it's not this player's turn")
	ErrInvalidDiceValue     = errors.New("the value of dice is invalid")
	ErrInvalidAdvancesValue = errors.New("the value of advances is invalid")
	ErrAdvancesNotMatching  = errors.New("the advances don't match the dice")
	ErrInvalidAdvances      = errors.New("the advances are invalid")
	ErrMustAdvance          = errors.New("must advance if possible")
	ErrCantContinue         = errors.New("can't continue with no possible advances")
	ErrMustUseBothGroups    = errors.New("must use both groups of dice if possible")
)

type Game struct {
	mu      *sync.RWMutex
	err     error
	state   State
	moves   []Move
	winner  uint8
	isEnded bool
}

func NewGame(numPlayer int) (*Game, error) {
	if numPlayer <= 0 || numPlayer > 5 {
		return nil, ErrInvalidNumPlayer
	}

	g := Game{
		mu:      new(sync.RWMutex),
		err:     nil,
		state:   getInitialState(numPlayer),
		moves:   make([]Move, 0),
		winner:  0,
		isEnded: false,
	}

	return &g, nil
}

func (g *Game) State() State {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.state
}

func (g *Game) TurnCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return int(g.state.TurnCount)
}

func (g *Game) Player() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return int(g.state.Player)
}

func (g *Game) MoveCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return int(g.state.MoveCount)
}

func (g *Game) NumPlayer() int {
	return int(g.state.numPlayer())
}

func (g *Game) Scores() []int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	scores := make([]int, g.NumPlayer())
	for i := range scores {
		scores[i] = g.score(i + 1)
	}
	return scores
}

func (g *Game) score(player int) int {
	score := 0
	for i := range g.state.Progress[player] {
		if g.state.Progress[player][i] == 0 {
			score++
		}
	}
	return score
}

func (g *Game) Board() {
	g.mu.RLock()
	defer g.mu.RUnlock()

}

func (g *Game) SubmitMove(m Move) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isTerminated() {
		return fmt.Errorf("%w: %w", ErrGameTerminated, g.err)
	}
	if g.isConcluded() {
		return ErrGameConcluded
	}

	err := g.validateMove(m)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMove, err)
	}

	g.state.makeMove(m)
	g.moves = append(g.moves, m)
	if !m.Continues && g.score(g.Player()) >= goal {
		g.winner = g.state.Player
		g.isEnded = true
	}
	return nil
}

func (g *Game) validateMove(m Move) error {
	if m.Player != g.state.Player {
		return ErrNotThisPlayersTurn
	}
	if !isValidDiceValue(m.Dice) {
		return ErrInvalidDiceValue
	}
	for _, a := range m.Advances {
		if a == 1 || a > 12 {
			return ErrInvalidAdvancesValue
		}
	}
	a1, a2 := m.Advances[0], m.Advances[1]
	if a1 == 0 && a2 == 0 {
		for i := range 6 {
			options := generateOptions(m.Dice, g.state)
			if options[i>>1][i&1] != [2]uint8{0, 0} {
				return ErrMustAdvance
			}
			if m.Continues {
				return ErrCantContinue
			}
		}
		return nil
	}
	if a1 == 0 || a2 == 0 {
		a1, a2 = a1+a2, 0
		groupings := groupDice(m.Dice)
		for i, gr := range groupings {
			if a1 == gr[0] {
				break
			} else if a1 == gr[1] {
				break
			}
			if i == 2 {
				return ErrAdvancesNotMatching
			}
		}
		if !g.state.isValidAdvances([2]uint8{a1, 0}) {
			return ErrInvalidAdvances
		}
		sum := m.Dice[0] + m.Dice[1] + m.Dice[2] + m.Dice[3]
		if g.state.isValidAdvances([2]uint8{a1, sum - a1}) {
			return ErrMustUseBothGroups
		}
		return nil
	}
	groupings := groupDice(m.Dice)
	for i, gr := range groupings {
		if a1 == gr[0] {
			if a2 != gr[1] {
				return ErrAdvancesNotMatching
			}
		} else if a1 == gr[1] {
			if a2 != gr[0] {
				return ErrAdvancesNotMatching
			}
		}
		if i == 2 {
			return ErrAdvancesNotMatching
		}
	}
	if !g.state.isValidAdvances([2]uint8{a1, a2}) {
		return ErrInvalidAdvances
	}
	return nil
}

func (g *Game) IsEnded() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.isEnded
}

func (g *Game) IsConcluded() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.isConcluded()
}

func (g *Game) isConcluded() bool {
	return g.isEnded && g.winner > 0
}

func (g *Game) GetWinner() (uint8, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.IsTerminated() {
		return 0, ErrGameTerminated
	}
	if !g.IsEnded() {
		return 0, ErrGameOngoing
	}

	return g.winner, nil
}

func (g *Game) IsTerminated() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.isTerminated()
}

func (g *Game) isTerminated() bool {
	return g.isEnded && g.winner == 0
}

func (g *Game) Err() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.err
}

func (g *Game) Terminate(reason error) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isTerminated() {
		return fmt.Errorf("%w: %w", ErrGameTerminated, g.err)
	}
	if g.isConcluded() {
		return ErrGameConcluded
	}

	g.isEnded = true
	g.err = reason
	return nil
}
