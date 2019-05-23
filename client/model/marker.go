package model

import (
	"errors"
)

type Marker struct {
	Name string
}

func (Marker) updateScore(score *Score) error {
	return errors.New("not implemented")
}

type AtMarker struct {
	Name string
}

func (AtMarker) updateScore(score *Score) error {
	return errors.New("not implemented")
}
