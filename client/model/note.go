package model

import (
	"alda.io/client/json"
	log "alda.io/client/logging"
)

// A Note represents a single pitch being sustained for a period of time.
type Note struct {
	SourceContext AldaSourceContext
	Pitch         PitchIdentifier
	Duration      Duration
	// When a note is slurred, it means there is minimal space between that note
	// and the next.
	Slurred bool
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (note Note) GetSourceContext() AldaSourceContext {
	return note.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (note Note) JSON() *json.Container {
	value := json.Object("pitch", note.Pitch.JSON())

	if note.Duration.Components != nil {
		value.Set(note.Duration.JSON(), "duration")
	}

	if note.Slurred {
		value.Set(true, "slurred?")
	}

	return json.Object("type", "note", "value", value)
}

// A NoteEvent is a Note expressed in absolute terms with the goal of performing
// the note e.g. on a MIDI sequencer/synthesizer.
type NoteEvent struct {
	Part            *Part
	MidiNote        int32
	Offset          float64
	Duration        float64
	AudibleDuration float64
	Volume          float64
	TrackVolume     float64
	Panning         float64
}

// JSON implements RepresentableAsJSON.JSON.
func (note NoteEvent) JSON() *json.Container {
	return json.Object(
		"part", note.Part.ID(),
		"midi-note", note.MidiNote,
		"offset", note.Offset,
		"duration", note.Duration,
		"audible-duration", note.AudibleDuration,
		"volume", note.Volume,
		"track-volume", note.TrackVolume,
		"panning", note.Panning,
	)
}

// EventOffset implements ScoreEvent.EventOffset by returning the offset of the
// note.
func (note NoteEvent) EventOffset() float64 {
	return note.Offset
}

func effectiveDuration(specifiedDuration Duration, part *Part) Duration {
	// If no duration is specified, use the part's default duration.
	if specifiedDuration.Components == nil {
		return part.Duration
	}

	return specifiedDuration
}

// Note/rest duration is "sticky." Any subsequent notes/rests without a
// specified duration will take on the duration of the part's last note/rest.
func updateDefaultDuration(part *Part, duration Duration) {
	if duration.Components != nil {
		part.Duration = duration
	}
}

func addNoteOrRest(score *Score, noteOrRest ScoreUpdate) {
	var specifiedDuration Duration
	switch noteOrRest.(type) {
	case Note:
		specifiedDuration = noteOrRest.(Note).Duration
	case Rest:
		specifiedDuration = noteOrRest.(Rest).Duration
	}

	for _, part := range score.CurrentParts {
		duration := effectiveDuration(specifiedDuration, part)
		durationMs := duration.Ms(part.Tempo) * part.TimeScale

		switch noteOrRest.(type) {
		case Note:
			note := noteOrRest.(Note)

			audibleDurationMs := durationMs
			if !note.Slurred {
				audibleDurationMs *= part.Quantization
			}

			if audibleDurationMs > 0 {
				midiNote := note.Pitch.CalculateMidiNote(
					part.Octave, part.KeySignature, part.Transposition,
				)

				noteEvent := NoteEvent{
					Part:            part.origin,
					MidiNote:        midiNote,
					Offset:          part.CurrentOffset,
					Duration:        durationMs,
					AudibleDuration: audibleDurationMs,
					Volume:          part.Volume,
					TrackVolume:     part.TrackVolume,
					Panning:         part.Panning,
				}

				log.Debug().
					Int32("MidiNote", noteEvent.MidiNote).
					Float64("Offset", noteEvent.Offset).
					Float64("AudibleDuration", noteEvent.AudibleDuration).
					Float64("Duration", noteEvent.Duration).
					Msg("Adding note.")

				score.Events = append(score.Events, noteEvent)
			}
		}

		if !score.chordMode {
			part.LastOffset = part.CurrentOffset
			part.CurrentOffset += durationMs
		}

		updateDefaultDuration(part, duration)
	}

	score.ApplyGlobalAttributes()
}

// UpdateScore implements ScoreUpdate.UpdateScore by adding a note to the score
// for all current parts and adjusting the parts' CurrentOffset, LastOffset, and
// Duration accordingly.
func (note Note) UpdateScore(score *Score) error {
	addNoteOrRest(score, note)
	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the duration of the
// note (if specified) or the current duration of the part, within the context
// of the part's current tempo.
//
// Also updates the part's default duration, so that it can be correctly
// considered when tallying the duration of subsequent events.
func (note Note) DurationMs(part *Part) float64 {
	durationMs := effectiveDuration(note.Duration, part).Ms(part.Tempo)
	updateDefaultDuration(part, note.Duration)
	return durationMs
}

// VariableValue implements ScoreUpdate.VariableValue.
func (note Note) VariableValue(score *Score) (ScoreUpdate, error) {
	return note, nil
}

// A Rest represents a period of time spent waiting.
//
// The function of a rest is to synchronize the following note so that it starts
// at a particular point in time.
type Rest struct {
	SourceContext AldaSourceContext
	Duration      Duration
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (rest Rest) GetSourceContext() AldaSourceContext {
	return rest.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (rest Rest) JSON() *json.Container {
	value := json.Object()

	if rest.Duration.Components != nil {
		value.Set(rest.Duration.JSON(), "duration")
	}

	return json.Object("type", "rest", "value", value)
}

// UpdateScore implements ScoreUpdate.UpdateScore by adjusting the
// CurrentOffset, LastOffset, and Duration of all current parts.
func (rest Rest) UpdateScore(score *Score) error {
	addNoteOrRest(score, rest)
	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning the duration of the
// rest (if specified) or the current duration of the part, within the context
// of the part's current tempo.
//
// Also updates the part's default duration, so that it can be correctly
// considered when tallying the duration of subsequent events.
func (rest Rest) DurationMs(part *Part) float64 {
	durationMs := effectiveDuration(rest.Duration, part).Ms(part.Tempo)
	updateDefaultDuration(part, rest.Duration)
	return durationMs
}

// VariableValue implements ScoreUpdate.VariableValue.
func (rest Rest) VariableValue(score *Score) (ScoreUpdate, error) {
	return rest, nil
}
