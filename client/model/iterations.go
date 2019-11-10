package model


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

// AppliesTo returns true if a particular iteration number belongs to one of the
// specified iteration ranges.
func (oi OnIterations) AppliesTo(iteration int32) bool {
	for _, r := range oi.Iterations {
		if r.First <= iteration && iteration <= r.Last {
			return true
		}
	}

	return false
}

// UpdateScore implements ScoreUpdate.UpdateScore by either updating the score
// with the event or doing nothing, depending on whether or not we are currently
// on a relevant iteration.
func (oi OnIterations) UpdateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		if oi.AppliesTo(part.CurrentIteration) {
			if err := oi.Event.UpdateScore(score); err != nil {
				return err
			}
		}
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the duration of the
// event on the current iteration.
func (oi OnIterations) DurationMs(part *Part) float32 {
	return 0 // TODO
}
