package model

import (
	"alda.io/client/json"

	"github.com/mohae/deepcopy"
)

// An EventSequence is an ordered sequence of events.
type EventSequence struct {
	Events []ScoreUpdate
}

// JSON implements RepresentableAsJSON.JSON.
func (es EventSequence) JSON() *json.Container {
	events := json.Array()
	for _, event := range es.Events {
		events.ArrayAppend(event.JSON())
	}

	return json.Object(
		"type", "event-sequence",
		"value", json.Object("events", events),
	)
}

// UpdateScore implements ScoreUpdate.UpdateScore by updating the score with
// each event in the sequence, in order.
func (es EventSequence) UpdateScore(score *Score) error {
	return score.Update(es.Events...)
}

// DurationMs implements ScoreUpdate.DurationMs by returning the total duration
// of the events in the sequence.
func (es EventSequence) DurationMs(part *Part) float64 {
	durationMs := 0.0

	for _, event := range es.Events {
		durationMs += event.DurationMs(part)
	}

	return durationMs
}

// VariableValue implements ScoreUpdate.VariableValue by returning a version of
// the event sequence where each event is the captured value of that event.
func (es EventSequence) VariableValue(score *Score) (ScoreUpdate, error) {
	result := deepcopy.Copy(es).(EventSequence)
	result.Events = []ScoreUpdate{}

	for _, event := range es.Events {
		eventValue, err := event.VariableValue(score)
		if err != nil {
			return nil, err
		}

		result.Events = append(result.Events, eventValue)
	}

	return result, nil
}
