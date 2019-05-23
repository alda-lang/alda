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

func (OnIterations) updateScore(score *Score) error {
	return errors.New("not implemented")
}
