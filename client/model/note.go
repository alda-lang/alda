package model

import (
	"errors"
)

type Note struct {
	NoteLetter  NoteLetter
	Accidentals []Accidental
	Duration    Duration
	// When a note is slurred, it means there is minimal space between that note
	// and the next.
	Slurred bool
}

func (Note) updateScore(score *Score) error {
	return errors.New("not implemented")
}

type Rest struct {
	Duration Duration
}

func (Rest) updateScore(score *Score) error {
	return errors.New("not implemented")
}
