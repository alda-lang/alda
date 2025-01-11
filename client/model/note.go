package model

import (
	"alda.io/client/help"
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
	MidiChannel     int32
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
		"midi-channel", note.MidiChannel,
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

func addNoteOrRest(score *Score, noteOrRest ScoreUpdate) error {
	// Avoid applying the same global attribute change multiple times if we're in
	// a chord.
	//
	// (To apply the change one time in the case of a chord, we call
	// score.ApplyGlobalAttributes() as part of the chord score update).
	if !score.chordMode {
		if err := score.ApplyGlobalAttributes(); err != nil {
			return err
		}
	}

	var specifiedDuration Duration
	switch noteOrRest := noteOrRest.(type) {
	case Note:
		specifiedDuration = noteOrRest.Duration
	case Rest:
		specifiedDuration = noteOrRest.Duration
	}

	if err := specifiedDuration.Validate(); err != nil {
		return err
	}

	for _, part := range score.CurrentParts {
		duration := effectiveDuration(specifiedDuration, part)
		durationMs := duration.Ms(part.Tempo) * part.TimeScale

		switch noteOrRest := noteOrRest.(type) {
		case Note:
			audibleDurationMs := durationMs
			if !noteOrRest.Slurred {
				audibleDurationMs *= part.Quantization
			}

			if audibleDurationMs > 0 {
				midiNote := noteOrRest.Pitch.CalculateMidiNote(
					part.Octave, part.KeySignature, part.Transposition,
				)

				if midiNote < 0 || midiNote > 127 {
					return help.UserFacingErrorf("MIDI note out of the 0-127 range. Input note: %d", midiNote)
				}

				midiChannel, err := score.assignMidiChannel(part, audibleDurationMs)
				if err != nil {
					return err
				}

				part.MidiChannel = midiChannel
				part.origin.MidiChannel = midiChannel

				noteEvent := NoteEvent{
					Part:            part.origin,
					MidiChannel:     midiChannel,
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

				// TODO: Consider proactively sorting `score.Events` by offset. This
				// might help to speed up MIDI channel availability checks, if we need
				// to.
				//
				// It might also be helpful to proactively sort by MIDI channel.
				//
				// See `complexCheck` in `model/midi.go`.
				score.Events = append(score.Events, noteEvent)
			}
		}

		if !score.chordMode {
			part.LastOffset = part.CurrentOffset
			part.CurrentOffset += durationMs
		}

		updateDefaultDuration(part, duration)
	}

	return nil
}

// UpdateScore implements ScoreUpdate.UpdateScore by adding a note to the score
// for all current parts and adjusting the parts' CurrentOffset, LastOffset, and
// Duration accordingly.
func (note Note) UpdateScore(score *Score) error {
	return addNoteOrRest(score, note)
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
	return addNoteOrRest(score, rest)
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
