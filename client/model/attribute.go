package model

import (
	"errors"
)

type OctaveSet struct {
	OctaveNumber int32
}

func (OctaveSet) updateScore(score *Score) error {
	return errors.New("OctaveSet.updateScore not implemented")
}

type OctaveUp struct{}

func (OctaveUp) updateScore(score *Score) error {
	return errors.New("OctaveUp.updateScore not implemented")
}

type OctaveDown struct{}

func (OctaveDown) updateScore(score *Score) error {
	return errors.New("OctaveDown.updateScore not implemented")
}
