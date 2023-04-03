package importer

import (
	"reflect"

	"alda.io/client/model"
	"github.com/go-test/deep"
)

type optimizer struct {
	// currentNoteState is true if the last occurrence of a note had accidentals
	// different from the key signature
	// This signifies that the next redundant accidental (same as key signature)
	// will be kept to re-iterate this return to the key signature
	currentNoteState    map[model.NoteLetter]bool

	// currentDuration and currentOctave maintain the last encountered duration
	// and octave values
	// Both attributes are reset in various situations such as nested structures
	// and barlines to ensure integrity across complex Alda structures
	currentDuration model.Duration
	currentOctave   int32

	// currentKeySignature maintains the current active key signature
	// This is the only state variable that is not soft reset because MusicXML
	// import will never change key signatures within single voices
	currentKeySignature model.KeySignature
}

func newOptimizer() optimizer {
	opt := optimizer{
		currentKeySignature: model.KeySignatureFromCircleOfFifths(0),
		currentNoteState:    make(map[model.NoteLetter]bool),
	}

	opt.softReset()
	return opt
}

func (opt *optimizer) resetNoteState() {
	for noteLetter := range model.NoteLetterIntervals {
		opt.currentNoteState[noteLetter] = false
	}
}

func (opt *optimizer) resetDuration() {
	opt.currentDuration = model.Duration{}
}

func (opt *optimizer) resetOctave() {
	opt.currentOctave = -1
}

func (opt *optimizer) softReset() {
	opt.resetNoteState()
	opt.resetDuration()
	opt.resetOctave()
}

// removeRedundantAccidentals will remove all unnecessary accidentals covered by
// the key signature, but keep redundant accidentals that represent a return to
// the key signature
// While having an accidental to return to a key signature is not necessary in
// Alda, it exists in all sheet music, so makes sense to have here
func (opt *optimizer) removeRedundantAccidentals(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	modify := func(update model.ScoreUpdate) model.ScoreUpdate {
		switch typedUpdate := update.(type) {
		// Key Signature
		case model.AttributeUpdate:
			switch typedPartUpdate := typedUpdate.PartUpdate.(type) {
			case model.KeySignatureSet:
				different := len(deep.Equal(
					opt.currentKeySignature, typedPartUpdate.KeySignature,
				)) > 0

				if different {
					opt.resetNoteState()
					opt.currentKeySignature = typedPartUpdate.KeySignature
				}
			}
		// Accidentals
		case model.Note:
			switch typedPitchIdentifier := typedUpdate.Pitch.(type) {
			case model.LetterAndAccidentals:
				letter := typedPitchIdentifier.NoteLetter
				accidentals := typedPitchIdentifier.Accidentals
				keySignatureAccidentals := opt.currentKeySignature[letter]

				different := len(deep.Equal(
					accidentals, keySignatureAccidentals,
				)) > 0

				if different {
					// When accidentals are different than the key, we just
					// update our current note state
					opt.currentNoteState[letter] = true
				} else {
					if !opt.currentNoteState[letter] {
						// Clear accidentals (redundant)
						typedPitchIdentifier.Accidentals = nil
						typedUpdate.Pitch = typedPitchIdentifier
						update = typedUpdate
					}
				}
			}
		}

		if isOrContainsBarline(update) {
			opt.resetNoteState()
		}

		return update
	}

	for i, update := range updates {
		update = modify(update)

		// Recursively optimize through nested score updates
		if modified, ok := modifyNestedUpdates(
			update, opt.removeRedundantAccidentals,
		); ok {
			update = modified
		}

		updates[i] = update
	}

	return updates
}

// removeRedundantDurations will remove repeated durations within a measure
// The last tracked duration will reset in various situations such as nested
// structures and new measures
// This is so we only remove durations that are truly unnecessary, but keep
// those that are visually important for the Alda code
func (opt *optimizer) removeRedundantDurations(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	modify := func(update model.ScoreUpdate) model.ScoreUpdate {
		if isOrContainsBarline(update) {
			opt.resetDuration()
			return update
		}

		duration := getNoteOrRestDuration(update)

		// If there are no duration components, we just skip
		if len(duration.Components) == 0 {
			return update
		}

		if len(opt.currentDuration.Components) != 0 {
			difference := deep.Equal(opt.currentDuration, duration)
			if len(difference) == 0 {
				// This is a repeated duration, we set it to nil
				update = setNoteOrRestDuration(update, model.Duration{})
			}
		}

		opt.currentDuration = duration
		return update
	}

	for i, update := range updates {
		update = modify(update)

		// Recursively optimize through nested score updates
		if _, ok := getNestedUpdates(update, false); ok {
			opt.resetDuration()

			modified, _ := modifyNestedUpdates(
				update, opt.removeRedundantDurations,
			)
			update = modified

			opt.resetDuration()
		}

		updates[i] = update
	}

	return updates
}

// translateMidiNotePitches takes all imported notes with the
// model.MidiNoteNumber pitch identifier and translates these to
// model.LetterAndAccidentals following standard pitched notes
func (opt *optimizer) translateMidiNotePitches(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	// We track octave sets that need to be inserted
	var octaveSetIndices []int
	var octaveSetOctaves []int32

	for i, update := range updates {
		// Translate note pitch
		switch typedUpdate := update.(type) {
		case model.Note:
			if reflect.TypeOf(typedUpdate.Pitch) == midiNoteNumberType {
				midiNoteNumber := typedUpdate.Pitch.(model.MidiNoteNumber)
				laa, octave := toLetterAndAccidentalsAndOctave(midiNoteNumber)

				typedUpdate.Pitch = laa
				update = typedUpdate

				if octave != opt.currentOctave {
					octaveSetIndices = append(octaveSetIndices, i)
					octaveSetOctaves = append(octaveSetOctaves, octave)
					opt.currentOctave = octave
				}
			}
		}

		// Reset
		if isOrContainsBarline(update) {
			opt.resetOctave()
		}

		// Recursively optimize through nested score updates
		if _, ok := getNestedUpdates(update, false); ok {
			opt.resetOctave()

			modified, _ := modifyNestedUpdates(
				update, opt.translateMidiNotePitches,
			)
			update = modified

			opt.resetOctave()
		}

		updates[i] = update
	}

	// Insert all octave sets
	for i := len(octaveSetIndices) - 1; i >= 0; i-- {
		octaveSet := model.AttributeUpdate{PartUpdate: model.OctaveSet{
			OctaveNumber: octaveSetOctaves[i],
		}}

		updates = insert(octaveSet, updates, octaveSetIndices[i])
	}

	return updates
}

// optimize applies various modifications to generate more idiomatic Alda
// optimize is called on updates without knowledge of parts or voices
func (opt *optimizer) optimize(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	// Required: standardizeBarlines < removeRedundantDurations
	// So we can remove durations for the last note in a bar that originally has
	// the barline imported as the last duration component
	updates = standardizeBarlines(updates)
	updates = opt.removeRedundantAccidentals(updates)

	// Required: translateMidiNotePitches < removeRedundantDurations
	// So unpitched percussion notes can have redundant durations removed too
	updates = opt.translateMidiNotePitches(updates)
	updates = opt.removeRedundantDurations(updates)

	// Reset after single optimize calls so subsequent voices have fresh starts
	opt.softReset()
	return updates
}
