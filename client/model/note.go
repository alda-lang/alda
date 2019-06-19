package model

import (
	"errors"
)

// A Note represents a single pitch being sustained for a period of time.
type Note struct {
	NoteLetter  NoteLetter
	Accidentals []Accidental
	Duration    Duration
	// When a note is slurred, it means there is minimal space between that note
	// and the next.
	Slurred bool
}

func (note Note) updateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		part.CurrentOffset += note.Duration.Ms(part.Tempo)

		// Note duration is "sticky." Any subsequent notes without a specified
		// duration will take on the duration of the part's last note.
		if note.Duration.Components != nil {
			part.Duration = note.Duration
		}
	}

	return nil
}

// A Rest represents a period of time spent waiting.
//
// The function of a rest is to synchronize the following note so that it starts
// at a particular point in time.
type Rest struct {
	Duration Duration
}

func (Rest) updateScore(score *Score) error {
	return errors.New("Rest.updateScore not implemented")
}
