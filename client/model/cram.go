package model

import (
	"alda.io/client/json"
	"github.com/mohae/deepcopy"
)

// A Cram expression fits a variable number of events into a fixed duration.
//
// The relative durations of the events are preserved, and time-stretched so
// that the total duration is equal to the "outer" duration of the Cram
// expression.
type Cram struct {
	SourceContext AldaSourceContext
	Events        []ScoreUpdate
	Duration      Duration
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (cram Cram) GetSourceContext() AldaSourceContext {
	return cram.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (cram Cram) JSON() *json.Container {
	events := json.Array()
	for _, event := range cram.Events {
		events.ArrayAppend(event.JSON())
	}

	value := json.Object("events", events)

	if cram.Duration.Components != nil {
		value.Set(cram.Duration.JSON(), "duration")
	}

	return json.Object("type", "cram", "value", value)
}

// Within the context of a part, the "inner duration" of the events in a cram
// expression is the sum of the result of calling DurationMs on each event,
// passing in the part for context.
//
// DurationMs can mutate the part, so we have to make a copy of the part and use
// that for context.
func innerDurationMs(cram Cram, part *Part) (float64, error) {
	partCopy := part.Clone()

	totalDurationMs := 0.0

	for _, event := range cram.Events {
		totalDurationMs += event.DurationMs(partCopy)
	}

	return totalDurationMs, nil
}

func timeScale(cram Cram, part *Part) (float64, error) {
	duration := effectiveDuration(cram.Duration, part)

	// The "outer duration" is the total duration for the entire cram
	// expression.
	//
	// For example, in the expression {c c c c c c c}8, the "outer duration" is
	// an eighth note (half a beat).
	outerDurationMs := duration.Ms(part.Tempo)

	// The "inner duration" is the sum of the durations of the events inside of
	// the cram expression, ignoring the context of time scale.
	//
	// For example, in the expression {c c c c c c c}8, the "inner duration" is
	// seven quarter notes (seven beats).
	innerDurationMs, err := innerDurationMs(cram, part)
	if err != nil {
		return -1, err
	}

	return (part.TimeScale / innerDurationMs) * outerDurationMs, nil
}

// UpdateScore implements ScoreUpdate.UpdateScore by doing the following for
// each current part:
//
//   - Calculate the time scaling factor for the notes within the Cram expression.
//     This depends on the part's current default duration, the "outer duration"
//     of the Cram expression (i.e. either the specified duration of the Cram
//     expression, or part's default duration), and the total "inner duration" of
//     the events within the Cram expression.
//
// * Set the part's TimeScale value.
//
//   - Within the new time scaling context, use the events within the Cram
//     expression to update the score.
//
// * Restore the part's previous TimeScale value.
func (cram Cram) UpdateScore(score *Score) error {
	previousDurations := map[*Part]Duration{}
	previousTimeScales := map[*Part]float64{}
	for _, part := range score.CurrentParts {
		previousDurations[part] = part.Duration
		previousTimeScales[part] = part.TimeScale
	}

	for _, part := range score.CurrentParts {
		timeScale, err := timeScale(cram, part)
		if err != nil {
			return err
		}

		part.TimeScale = timeScale
	}

	for _, event := range cram.Events {
		if err := event.UpdateScore(score); err != nil {
			return err
		}
	}

	for _, part := range score.CurrentParts {
		// If the cram expression has a specified duration, then that becomes the
		// new duration of each part.
		//
		// Otherwise, it remains whatever it was before the cram expression.
		if cram.Duration.Components != nil {
			part.Duration = cram.Duration
		} else {
			part.Duration = previousDurations[part]
		}
		part.TimeScale = previousTimeScales[part]
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the effective
// duration of the Cram expression, i.e. either the specified duration of the
// Cram expression or the part's default duration.
func (cram Cram) DurationMs(part *Part) float64 {
	return effectiveDuration(cram.Duration, part).Ms(part.Tempo)
}

// VariableValue implements ScoreUpdate.VariableValue by returning a version of
// the Cram expression where each event is the captured value of that event.
func (cram Cram) VariableValue(score *Score) (ScoreUpdate, error) {
	result := deepcopy.Copy(cram).(Cram)
	result.Events = []ScoreUpdate{}

	for _, event := range cram.Events {
		eventValue, err := event.VariableValue(score)
		if err != nil {
			return nil, err
		}

		result.Events = append(result.Events, eventValue)
	}

	return result, nil
}
