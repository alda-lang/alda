package model

// An EventSequence is an ordered sequence of events.
type EventSequence struct {
	Events []ScoreUpdate
}

// UpdateScore implements ScoreUpdate.UpdateScore by updating the score with
// each event in the sequence, in order.
func (es EventSequence) UpdateScore(score *Score) error {
	return score.Update(es.Events...)
}

// DurationMs implements ScoreUpdate.DurationMs by returning the total duration
// of the events in the sequence.
func (es EventSequence) DurationMs(part *Part) float32 {
	durationMs := float32(0)

	for _, event := range es.Events {
		durationMs += event.DurationMs(part)
	}

	return durationMs
}
