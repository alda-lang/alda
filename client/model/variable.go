package model

import (
	"errors"
)

// A VariableDefinition stores a sequence of ScoreUpdates, using the provided
// variable name as a lookup key.
type VariableDefinition struct {
	VariableName string
	Events       []ScoreUpdate
}

// UpdateScore implements ScoreUpdate.UpdateScore by defining a variable.
func (VariableDefinition) UpdateScore(score *Score) error {
	return errors.New("VariableDefinition.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a
// variable definition is conceptually instantaneous.
func (VariableDefinition) DurationMs(part *Part) float32 {
	return 0
}

// A VariableReference dereferences a stored variable. A variable with the
// provided name is looked up, and assuming that it was previously defined, the
// corresponding sequence of events defined is used to update the score.
type VariableReference struct {
	VariableName string
}

// UpdateScore implements ScoreUpdate.UpdateScore by looking up a variable and
// (assuming it was previously defined) using the corresponding sequence of
// events to update the score.
func (VariableReference) UpdateScore(score *Score) error {
	return errors.New("VariableReference.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by looking up the sequence of
// events corresponding to the given variable name, and returning the sum of
// the durations of the events.
func (VariableReference) DurationMs(part *Part) float32 {
	// TODO: look up the variable, sum the durations of the corresponding events
	//
	// If the variable is undefined, an error will be thrown when we come back
	// through and look it up again for UpdateScore. So, we can safely ignore the
	// fact that the variable is undefined here and simply return 0.
	return 0
}
