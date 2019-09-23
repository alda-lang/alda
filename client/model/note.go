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

// MidiNote returns the MIDI note number of a note, given contextual information
// about the part playing the note (e.g. octave, key signature, transposition).
func (note Note) MidiNote(part *Part) int32 {
	intervals := map[NoteLetter]int32{
		C: 0, D: 2, E: 4, F: 5, G: 7, A: 9, B: 11,
	}

	baseMidiNoteNumber := ((part.Octave + 1) * 12) + intervals[note.NoteLetter]

	var accidentals []Accidental
	if note.Accidentals == nil {
		accidentals = part.KeySignature[note.NoteLetter]
	} else {
		accidentals = note.Accidentals
	}

	for _, accidental := range accidentals {
		switch accidental {
		case Flat:
			baseMidiNoteNumber--
		case Sharp:
			baseMidiNoteNumber++
		}
	}

	return baseMidiNoteNumber + part.Transposition
}

// A NoteEvent is a Note expressed in absolute terms with the goal of performing
// the note e.g. on a MIDI sequencer/synthesizer.
type NoteEvent struct {
	Part            *Part
	MidiNote        int32
	Offset          OffsetMs
	Duration        float32
	AudibleDuration float32
	Volume          float32
	TrackVolume     float32
	Panning         float32
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
		var duration Duration

		// If no duration is specified, use the part's default duration.
		if specifiedDuration.Components == nil {
			duration = part.Duration
		} else {
			duration = specifiedDuration
		}

		durationMs := duration.Ms(part.Tempo)

		switch noteOrRest.(type) {
		case Note:
			midiNote := noteOrRest.(Note).MidiNote(part)

			noteEvent := NoteEvent{
				Part:            part,
				MidiNote:        midiNote,
				Offset:          part.CurrentOffset,
				Duration:        durationMs,
				AudibleDuration: durationMs * part.Quantization,
				Volume:          part.Volume,
				TrackVolume:     part.TrackVolume,
				Panning:         part.Panning,
			}

			score.Events = append(score.Events, noteEvent)
		}

		if !score.chordMode {
			part.LastOffset = part.CurrentOffset
			part.CurrentOffset += float64(durationMs)
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
	addNoteOrRest(score, note)
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
	addNoteOrRest(score, rest)
	return nil
}
