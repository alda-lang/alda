package model

import (
	"errors"
)

type Cram struct {
	Events   []ScoreUpdate
	Duration Duration
}

func (Cram) updateScore(score *Score) error {
	return errors.New("Cram.updateScore not implemented")
}
