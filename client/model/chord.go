package model

import (
	"math"

	"alda.io/client/json"
	"github.com/mohae/deepcopy"
)

// A Chord is a collection of notes and rests starting at the same point in
// time.
//
// Certain other types of events are allowed to occur between the notes and
// rests, e.g. octave and other attribute changes.
type Chord struct {
	SourceContext AldaSourceContext
	Events        []ScoreUpdate
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (chord Chord) GetSourceContext() AldaSourceContext {
	return chord.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (chord Chord) JSON() *json.Container {
	events := json.Array()
	for _, event := range chord.Events {
		events.ArrayAppend(event.JSON())
	}

	return json.Object(
		"type", "chord",
		"value", json.Object("events", events),
	)
}

// UpdateScore implements ScoreUpdate.UpdateScore by adding multiple notes to
// all active parts, and updating each part's CurrentOffset, LastOffset, and
// Duration accordingly.
func (chord Chord) UpdateScore(score *Score) error {
	score.ApplyGlobalAttributes()

	shortestDurationMs := map[*Part]float64{}
	for _, part := range score.CurrentParts {
		shortestDurationMs[part] = math.MaxFloat64
	}

	score.chordMode = true
	for _, event := range chord.Events {
		// Notes/rests in a chord can have different durations. Following a chord, the
		// next note/rest is placed after the shortest note/rest in the chord.
		//
		// Here, we take note of the event's duration so that we can keep track of
		// which one is the shortest.
		var specifiedDuration Duration
		switch event.(type) {
		case Note:
			specifiedDuration = event.(Note).Duration
		case Rest:
			specifiedDuration = event.(Rest).Duration
		}

		for _, part := range score.CurrentParts {
			duration := effectiveDuration(specifiedDuration, part)
			durationMs := duration.Ms(part.Tempo) * part.TimeScale
			shortestDurationMs[part] = math.Min(shortestDurationMs[part], durationMs)
		}

		// Now, we update the score with the event, in "chord mode," which means
		// that notes all start at the same offset.
		if err := event.UpdateScore(score); err != nil {
			return err
		}
	}
	score.chordMode = false

	for _, part := range score.CurrentParts {
		part.LastOffset = part.CurrentOffset
		part.CurrentOffset += shortestDurationMs[part]
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the shortest
// note/rest duration in the chord, within the context of the part's current
// tempo.
func (chord Chord) DurationMs(part *Part) float64 {
	shortestDurationMs := math.MaxFloat64

	for _, event := range chord.Events {
		durationMs := event.DurationMs(part)
		shortestDurationMs = math.Min(shortestDurationMs, durationMs)
	}

	return shortestDurationMs
}

// VariableValue implements ScoreUpdate.VariableValue by returning a version of
// the chord where each event is the captured value of that event.
func (chord Chord) VariableValue(score *Score) (ScoreUpdate, error) {
	result := deepcopy.Copy(chord).(Chord)
	result.Events = []ScoreUpdate{}

	for _, event := range chord.Events {
		eventValue, err := event.VariableValue(score)
		if err != nil {
			return nil, err
		}

		result.Events = append(result.Events, eventValue)
	}

	return result, nil
}
