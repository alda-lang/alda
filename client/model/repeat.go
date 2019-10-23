package model

import (
	"errors"
)

// A Repeat expression repeats an event a number of times.
type Repeat struct {
	Event ScoreUpdate
	Times int32
}

// UpdateScore implements ScoreUpdate.UpdateScore by repeatedly updating the
// score with an event a specified number of times.
func (Repeat) UpdateScore(score *Score) error {
	return errors.New("Repeat.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning the total duration
// of the event being repeated the specified number of times.
func (repeat Repeat) DurationMs(part *Part) float32 {
	// FIXME: This probably doesn't work, in light of OnIterations, which makes it
	// so that an event sequence can be different on different iterations.
	return repeat.Event.DurationMs(part) * float32(repeat.Times)
}
