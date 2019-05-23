package model

import (
	"errors"
)

// A PartDeclaration sets the current instruments of the score, creating them if
// necessary.
type PartDeclaration struct {
	Names []string
	Alias string
}

func (pd PartDeclaration) updateScore(score *Score) error {
	return errors.New("not implemented")
}
