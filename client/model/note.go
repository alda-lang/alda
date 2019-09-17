package model

// A Note represents a single pitch being sustained for a period of time.
type Note struct {
	NoteLetter  NoteLetter
	Accidentals []Accidental
	Duration    Duration
	// When a note is slurred, it means there is minimal space between that note
	// and the next.
	Slurred bool
}

// A NoteEvent is a Note expressed in absolute terms with the goal of performing
// the note e.g. on a MIDI sequencer/synthesizer.
type NoteEvent struct {
	MidiNote        int32
	Offset          float32
	Duration        float32
	AudibleDuration float32
}

func addNoteOrRest(score *Score, duration Duration, midiNote int32) {
	for _, part := range score.CurrentParts {
		durationMs := duration.Ms(part.Tempo)

		if midiNote != 0 {
			noteEvent := NoteEvent{
				MidiNote:        midiNote,
				Offset:          part.CurrentOffset,
				Duration:        durationMs,
				AudibleDuration: durationMs * part.Quantization,
			}

			score.Events = append(score.Events, noteEvent)
		}

		if !score.chordMode {
			part.LastOffset = part.CurrentOffset
			part.CurrentOffset += durationMs
		}

		// Note/rest duration is "sticky." Any subsequent notes/rests without a
		// specified duration will take on the duration of the part's last
		// note/rest.
		if duration.Components != nil {
			part.Duration = duration
		}
	}
}

func (note Note) updateScore(score *Score) error {
	midiNote := int32(42) // FIXME
	addNoteOrRest(score, note.Duration, midiNote)
	return nil
}

// A Rest represents a period of time spent waiting.
//
// The function of a rest is to synchronize the following note so that it starts
// at a particular point in time.
type Rest struct {
	Duration Duration
}

func (rest Rest) updateScore(score *Score) error {
	addNoteOrRest(score, rest.Duration, 0)
	return nil
}
