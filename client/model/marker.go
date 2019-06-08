package model

import (
	"errors"
)

// StartOfScore is a special marker that represents the start of the score.  All
// parts start at this marker implicitly.
const StartOfScore string = "alda-start-of-score"

type Marker struct {
	Name string
}

func (Marker) updateScore(score *Score) error {
	return errors.New("Marker.updateScore not implemented")
}

type AtMarker struct {
	Name string
}

func (AtMarker) updateScore(score *Score) error {
	return errors.New("AtMarker.updateScore not implemented")
}
