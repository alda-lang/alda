package model

// RepetitionRange represents a single, inclusive range of repetition numbers,
// e.g. 1-4.
// An RepetitionRange can also represent a single ending number, e.g. 1-1.
type RepetitionRange struct {
	First int32
	Last  int32
}

// OnRepetitions wraps an event (something that implements ScoreUpdate) in order
// to specify that it should only occur on certain repetition numbers through a
// repeated pattern.
type OnRepetitions struct {
	Repetitions []RepetitionRange
	Event       ScoreUpdate
}

// AppliesTo returns true if a particular repetition number belongs to one of the
// specified repetition ranges.
func (or OnRepetitions) AppliesTo(repetition int32) bool {
	for _, r := range or.Repetitions {
		if r.First <= repetition && repetition <= r.Last {
			return true
		}
	}

	return false
}

// UpdateScore implements ScoreUpdate.UpdateScore by either updating the score
// with the event or doing nothing, depending on whether or not we are currently
// on a relevant repetition.
func (or OnRepetitions) UpdateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		if or.AppliesTo(part.CurrentRepetition) {
			if err := or.Event.UpdateScore(score); err != nil {
				return err
			}
		}
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the duration of the
// event on the current repetition.
func (or OnRepetitions) DurationMs(part *Part) float32 {
	return 0 // TODO
}
