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

// FIXME: just doing the bare minimum for now to support parts tests i'm writing
func (note Note) updateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		for _, component := range note.Duration.Components {
			part.CurrentOffset += component.(NoteLengthMs).Quantity
		}
	}

	return nil
}

type Rest struct {
	Duration Duration
}

func (Rest) updateScore(score *Score) error {
	return errors.New("Rest.updateScore not implemented")
}
