package model

import (
	"errors"
)

type OctaveSet struct {
	OctaveNumber int32
}

func (OctaveSet) updateScore(score *Score) error {
	return errors.New("not implemented")
}

type OctaveUp struct{}

func (OctaveUp) updateScore(score *Score) error {
	return errors.New("not implemented")
}

type OctaveDown struct{}

func (OctaveDown) updateScore(score *Score) error {
	return errors.New("not implemented")
}
