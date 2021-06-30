package importer

import (
	"alda.io/client/model"
	"github.com/go-test/deep"
	"reflect"
)

// postProcess applies various modifications to generate more idiomatic Alda
func postProcess(updates []model.ScoreUpdate) []model.ScoreUpdate {
	processor := newPostProcessor()
	return processor.processAll(updates)
}

type postProcessor struct {
	// currentNoteState is true if the last occurrence of a note had accidentals
	// different from the key signature
	// This signifies that the next redundant accidental (same as key signature)
	// will be kept to re-iterate this return to the key signature
	currentNoteState    map[model.NoteLetter]bool
	currentKeySignature model.KeySignature

	// currentDuration and currentOctave maintain the last encountered duration
	// and octave values
	// Both attributes are reset in various situations such as nested structures
	// and barlines to ensure integrity across complex Alda structures
	currentDuration model.Duration
	currentOctave   int32
}

func newPostProcessor() postProcessor {
	processor := postProcessor{
		currentKeySignature: model.KeySignatureFromCircleOfFifths(0),
		currentNoteState:    make(map[model.NoteLetter]bool),
	}

	processor.resetNoteState()
	return processor
}

func (processor *postProcessor) resetNoteState() {
	for noteLetter, _ := range model.NoteLetterIntervals {
		processor.currentNoteState[noteLetter] = false
	}
}

func (processor *postProcessor) resetDuration() {
	processor.currentDuration = model.Duration{}
}

func (processor *postProcessor) resetOctave() {
	processor.currentOctave = -1
}

func (processor *postProcessor) hasDuration() bool {
	return len(processor.currentDuration.Components) != 0
}

func (processor *postProcessor) processAll(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	// Required: standardizeBarlines < removeRedundantDurations
	// So we can remove durations for the last note in a bar that originally has
	// the barline imported as the last duration component
	updates = standardizeBarlines(updates)
	updates = processor.removeRedundantAccidentals(updates)

	// Required: translateMidiNotePitches < removeRedundantDurations
	// So unpitched percussion notes can have redundant durations removed too
	updates = processor.translateMidiNotePitches(updates)
	updates = processor.removeRedundantDurations(updates)
	return updates
}

// removeRedundantAccidentals will remove all unnecessary accidentals covered by
// the key signature, but keep redundant accidentals that represent a return to
// the key signature
// While having an accidental to return to a key signature is not necessary in
// Alda, it exists in all sheet music, so makes sense to have here
func (processor *postProcessor) removeRedundantAccidentals(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	modify := func(update model.ScoreUpdate) model.ScoreUpdate {
		switch typedUpdate := update.(type) {
		// Key Signature
		case model.AttributeUpdate:
			switch typedPartUpdate := typedUpdate.PartUpdate.(type) {
			case model.KeySignatureSet:
				different := len(deep.Equal(
					processor.currentKeySignature, typedPartUpdate.KeySignature,
				)) > 0

				if different {
					processor.resetNoteState()
					processor.currentKeySignature = typedPartUpdate.KeySignature
				}
			}
		// Accidentals
		case model.Note:
			switch typedPitchIdentifier := typedUpdate.Pitch.(type) {
			case model.LetterAndAccidentals:
				letter := typedPitchIdentifier.NoteLetter
				accidentals := typedPitchIdentifier.Accidentals
				keySignatureAccidentals := processor.currentKeySignature[letter]

				different := len(deep.Equal(
					accidentals, keySignatureAccidentals,
				)) > 0

				if different {
					// When accidentals are different than the key, we just
					// update our current note state
					processor.currentNoteState[letter] = true
				} else {
					if !processor.currentNoteState[letter] {
						// Clear accidentals (redundant)
						typedPitchIdentifier.Accidentals = nil
						typedUpdate.Pitch = typedPitchIdentifier
						update = typedUpdate
					}
				}
			}
		}

		if isOrContainsBarline(update) {
			processor.resetNoteState()
		}

		return update
	}

	for i, update := range updates {
		update = modify(update)

		// Recurse process through nested score updates
		if modified, ok := modifyNestedUpdates(
			update, processor.removeRedundantAccidentals,
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
func (processor *postProcessor) removeRedundantDurations(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	modify := func(update model.ScoreUpdate) model.ScoreUpdate {
		if isOrContainsBarline(update) {
			processor.resetDuration()
			return update
		}

		duration := getNoteOrRestDuration(update)

		// If there are no duration components, we just skip
		if len(duration.Components) == 0 {
			return update
		}

		if processor.hasDuration() {
			difference := deep.Equal(processor.currentDuration, duration)
			if len(difference) == 0 {
				// This is a repeated duration, we set it to nil
				update = setNoteOrRestDuration(update, model.Duration{})
			}
		}

		processor.currentDuration = duration
		return update
	}

	for i, update := range updates {
		update = modify(update)

		// Recurse process through nested score updates
		if _, ok := getNestedUpdates(update, false); ok {
			processor.resetDuration()

			modified, _ := modifyNestedUpdates(
				update, processor.removeRedundantDurations,
			)
			update = modified

			processor.resetDuration()
		}

		updates[i] = update
	}

	return updates
}

// translateMidiNotePitches takes all imported notes with the
// model.MidiNoteNumber pitch identifier and translates these to
// model.LetterAndAccidentals following standard pitched notes
func (processor *postProcessor) translateMidiNotePitches(
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
				pitch, octave := midiNoteNumber.ToNoteAndOctave()

				typedUpdate.Pitch = pitch
				update = typedUpdate

				if octave != processor.currentOctave {
					octaveSetIndices = append(octaveSetIndices, i)
					octaveSetOctaves = append(octaveSetOctaves, octave)
					processor.currentOctave = octave
				}
			}
		}

		// Reset
		if isOrContainsBarline(update) {
			processor.resetOctave()
		}

		// Recurse process through nested score updates
		if _, ok := getNestedUpdates(update, false); ok {
			processor.resetOctave()

			modified, _ := modifyNestedUpdates(
				update, processor.translateMidiNotePitches,
			)
			update = modified

			processor.resetOctave()
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
