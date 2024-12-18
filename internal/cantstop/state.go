package cantstop

type State struct {
	TurnCount uint8
	Player    uint8
	MoveCount uint8
	Progress  [][]int8 // progress[0] represents the neutral markers
}

type Move struct {
	Dice      [4]uint8
	Advances  [2]uint8
	Player    uint8
	Continues bool
}

func getInitialState(numPlayer int) State {
	progress := make([][]int8, numPlayer+1)
	progress[0] = make([]int8, len(pathLens))
	for i := 1; i < len(progress); i++ {
		progress[i] = make([]int8, len(pathLens))
		copy(progress[i], pathLens)
	}

	s := State{
		TurnCount: 1,
		Player:    1,
		MoveCount: 1,
		Progress:  progress,
	}
	return s
}

func (s State) numPlayer() uint8 {
	return uint8(len(s.Progress) - 1)
}

func (s *State) makeMove(m Move) {
	if m.Advances[0] == 0 {
		for i := range s.Progress[0] {
			s.Progress[0][i] = 0
		}
		if s.Player != s.numPlayer() {
			s.Player++
		} else {
			s.Player = 1
			s.TurnCount++
		}
	} else {
		s.Progress[0][m.Advances[0]]++
		if m.Advances[1] != 0 {
			s.Progress[0][m.Advances[1]]++
		}

		if m.Continues {
			s.MoveCount++
		} else {
			for i := range s.Progress[0] {
				s.Progress[s.Player][i] += s.Progress[0][i]
				s.Progress[0][i] = 0
			}
			if s.Player != s.numPlayer() {
				s.Player++
			} else {
				s.Player = 1
				s.TurnCount++
			}
		}
	}
}

func (s State) ownerOfPath(path uint8) uint8 {
	for i := 1; i <= int(s.numPlayer()); i++ {
		if s.Progress[i][path] == 0 {
			return uint8(i)
		}
	}
	return 0
}

func (s State) isOwnedPath(path uint8) bool {
	if path == 0 {
		return false
	}
	return s.ownerOfPath(path) > 0
}

func (s State) hasSpaceLeft(path uint8, numSpace int8) bool {
	if path == 0 {
		return true
	}
	return s.Progress[s.Player][path]-s.Progress[0][path] > numSpace
}

func (s State) isValidAdvances(a [2]uint8) bool {
	a1, a2 := a[0], a[1]

	if s.isOwnedPath(a1) || s.isOwnedPath(a2) {
		return false
	}

	if a1 == a2 {
		return s.hasSpaceLeft(a1, 2)
	}
	return s.hasSpaceLeft(a1, 1) && s.hasSpaceLeft(a2, 1)
}
