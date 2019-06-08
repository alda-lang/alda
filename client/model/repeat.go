package model

import (
	"errors"
)

type Repeat struct {
	Event ScoreUpdate
	Times int32
}

func (Repeat) updateScore(score *Score) error {
	return errors.New("Repeat.updateScore not implemented")
}
