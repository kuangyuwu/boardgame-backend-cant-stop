package cantstop

import (
	"errors"
)

var (
	ErrUnsupportedType = errors.New("received data of unspported type")

	ErrNotYourTurn         = errors.New("it is not this player's turn")
	ErrNotRolling          = errors.New("not in rolling phase")
	ErrNotAdvancing        = errors.New("not in advancing phase")
	ErrInvalidAdvanceData  = errors.New("invalid advance data")
	ErrNotDeciding         = errors.New("not in deciding phase")
	ErrInvalidDecisionData = errors.New("invalid decision data")
)

func (gm *GameManager) HandleData(playerIdx int, dataType string, body any) error {
	switch dataType {
	case "roll":
		return gm.handleRoll(playerIdx)
	case "advance":
		return gm.handleAdvance(playerIdx, body)
	case "decision":
		return gm.handleDecision(playerIdx, body)
		// case "exit":
		// 	return gm.handleExit(playerIdx)
	}
	return ErrUnsupportedType
}

func (gm *GameManager) handleRoll(playerIdx int) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if playerIdx != gm.game.PlayerNow() {
		return ErrNotYourTurn
	}
	if gm.phase != rolling {
		return ErrNotRolling
	}

	dice := RollDice()
	options, failed, _ := GenerateOptions(dice, gm.game.State())
	gm.Send(dataDice(dice, options, failed))

	gm.tempMove.Player = uint8(playerIdx)
	gm.tempMove.Dice = dice
	gm.tempOptions = options
	if failed {
		gm.tempMove.Advances = [2]uint8{0, 0}
		gm.phase = deciding
	} else {
		gm.phase = advancing
	}

	return nil
}

func (gm *GameManager) handleAdvance(playerIdx int, body any) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if playerIdx != gm.game.PlayerNow() {
		return ErrNotYourTurn
	}
	if gm.phase != advancing {
		return ErrNotAdvancing
	}

	slice, ok := body.([]any)
	if !ok || len(slice) != 2 {
		return ErrInvalidAdvanceData
	}
	a1, ok1 := slice[0].(float64)
	a2, ok2 := slice[1].(float64)
	if !ok1 || !ok2 || a1 < 0 || a1 > 12 || a2 < 0 || a2 > 12 {
		return ErrInvalidAdvanceData
	}

	advances := [2]uint8{uint8(a1), uint8(a2)}
	exists := false
	for i := range 6 {
		if gm.tempOptions[i>>1][i&1] == advances {
			exists = true
		}
	}
	if !exists {
		return ErrInvalidAdvanceData
	}

	board, _ := gm.game.BoardAfterAdvances(advances)
	gm.Send(dataAdvance(advances))
	gm.Send(dataBoard(board))

	gm.tempMove.Advances = advances
	gm.phase = deciding

	return nil
}

func (gm *GameManager) handleDecision(playerIdx int, body any) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if playerIdx != gm.game.PlayerNow() {
		return ErrNotYourTurn
	}
	if gm.phase != deciding {
		return ErrNotDeciding
	}

	continues, ok := body.(bool)
	if !ok {
		return ErrInvalidDecisionData
	}

	gm.tempMove.Continues = continues
	err := gm.game.SubmitMove(gm.tempMove)
	if err != nil {
		return err
	}

	gm.sendStateUpdate()
	if !continues {
		gm.Send(dataStop())
		gm.Send(dataScores(gm.game.Scores()))
		gm.Send(dataRoll())
		if gm.game.IsConcluded() {
			w, _ := gm.game.Winner()
			gm.Send(dataWinner(w))
			gm.phase = rolling
			return nil
		}
	}

	gm.phase = rolling
	return nil
}

// func (gm *GameManager) handleExit(playerIdx int) error {
// 	defer g.mu.Unlock()
// 	if !g.ended {
// 		g.logErrorAndTerminate(fmt.Sprintf("player %s exited unexpectedly", username))
// 	}
// 	for n, p := range g.players {
// 		if p.username == username {
// 			g.players[n].left = true
// 		}
// 	}
// 	g.sendExit(username)
// }

func (gm GameManager) sendStateUpdate() {
	gm.Send(dataTurn(gm.game.TurnCount()))
	gm.Send(dataPlayer(gm.game.PlayerNow()))
	gm.Send(dataMove(gm.game.MoveCount()))
	gm.Send(dataBoard(gm.game.Board()))
}
