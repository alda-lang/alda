package model

import (
	"errors"
)

// A Chord is a collection of notes and rests starting at the same point in
// time.
//
// Certain other types of events are allowed to occur between the notes and
// rests, e.g. octave and other attribute changes.
type Chord struct {
	Events []ScoreUpdate
}

func (Chord) updateScore(score *Score) error {
	return errors.New("not implemented")
}
