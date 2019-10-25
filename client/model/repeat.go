package model

// A Repeat expression repeats an event a number of times.
type Repeat struct {
	Event ScoreUpdate
	Times int32
}

// UpdateScore implements ScoreUpdate.UpdateScore by repeatedly updating the
// score with an event a specified number of times.
func (repeat Repeat) UpdateScore(score *Score) error {
	for iteration := int32(1); iteration <= repeat.Times; iteration++ {
		for _, part := range score.CurrentParts {
			part.CurrentIteration = iteration
		}

		if err := repeat.Event.UpdateScore(score); err != nil {
			return err
		}
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the total duration
// of the event being repeated the specified number of times.
func (repeat Repeat) DurationMs(part *Part) float32 {
	durationMs := float32(0)

	for iteration := int32(1); iteration <= repeat.Times; iteration++ {
		part.CurrentIteration = iteration
		durationMs += repeat.Event.DurationMs(part)
	}

	return durationMs
}
