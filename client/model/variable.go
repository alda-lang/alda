package model

import (
	"errors"
)

type VariableDefinition struct {
	VariableName string
	Events       []ScoreUpdate
}

func (VariableDefinition) updateScore(score *Score) error {
	return errors.New("not implemented")
}

type VariableReference struct {
	VariableName string
}

func (VariableReference) updateScore(score *Score) error {
	return errors.New("not implemented")
}
