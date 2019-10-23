package model

import (
	"errors"
)

// IterationRange represents a single, inclusive range of iteration numbers,
// e.g. 1-4.
// An IterationRange can also represent a single ending number, e.g. 1-1.
type IterationRange struct {
	First int32
	Last  int32
}

// OnIterations wraps an event (something that implements ScoreUpdate) in order
// to specify that it should only occur on certain iteration numbers through a
// repeated pattern.
type OnIterations struct {
	Iterations []IterationRange
	Event      ScoreUpdate
}

// UpdateScore implements ScoreUpdate.UpdateScore by either updating the score
// with the event or doing nothing, depending on whether or not we are currently
// on a relevant iteration.
func (OnIterations) UpdateScore(score *Score) error {
	return errors.New("OnIterations.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning the duration of the
// event on the current iteration.
func (oi OnIterations) DurationMs(part *Part) float32 {
	return 0 // TODO
}
