package model

import (
	"errors"
)

// A Marker gives a name to a point in time in a score.
type Marker struct {
	Name string
}

func (Marker) updateScore(score *Score) error {
	return errors.New("Marker.updateScore not implemented")
}

// AtMarker is an action where a part's offset gets set to a point in time
// denoted previously by a Marker.
type AtMarker struct {
	Name string
}

func (AtMarker) updateScore(score *Score) error {
	return errors.New("AtMarker.updateScore not implemented")
}
