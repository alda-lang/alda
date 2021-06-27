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

	// currentDuration stores the last duration encountered in a rest or note
	// currentDuration is reset in various situations such as nested structures
	// and barlines
	currentDuration     model.Duration
}

func newPostProcessor() postProcessor {
	processor := postProcessor{
		currentDuration:     model.Duration{},
		currentKeySignature: model.KeySignatureFromCircleOfFifths(0),
		currentNoteState:    make(map[model.NoteLetter]bool),
	}

	processor.resetNoteState()
	return processor
}

func (processor *postProcessor) resetNoteState() {
	for _, letter := range noteLetters {
		processor.currentNoteState[letter] = false
	}
}

func (processor *postProcessor) resetDuration() {
	processor.currentDuration = model.Duration{}
}

func (processor *postProcessor) hasDuration() bool {
	return len(processor.currentDuration.Components) != 0
}

func (processor *postProcessor) processAll(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	// We standardize barlines in post processing to improve duration removal
	updates = standardizeBarlines(updates)
	updates = processor.removeRedundantAccidentals(updates)
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
		case model.Barline:
			processor.resetNoteState()
		}

		// Process a note or rest's duration components to detect barlines
		var durationComponents []model.DurationComponent
		if reflect.TypeOf(update) == noteType {
			durationComponents = update.(model.Note).Duration.Components
		} else if reflect.TypeOf(update) == restType {
			durationComponents = update.(model.Rest).Duration.Components
		}

		for _, duration := range durationComponents {
			if reflect.TypeOf(duration) == barlineType {
				processor.resetNoteState()
			}
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
	getDuration := func(
		update model.ScoreUpdate,
	) model.Duration {
		if reflect.TypeOf(update) == noteType {
			return update.(model.Note).Duration
		} else if reflect.TypeOf(update) == restType {
			return update.(model.Rest).Duration
		} else {
			return model.Duration{}
		}
	}

	setDuration := func(
		update model.ScoreUpdate, duration model.Duration,
	) model.ScoreUpdate {
		switch typedUpdate := update.(type) {
		case model.Note:
			typedUpdate.Duration = duration
			update = typedUpdate
		case model.Rest:
			typedUpdate.Duration = duration
			update = typedUpdate
		}
		return update
	}

	modify := func(update model.ScoreUpdate) model.ScoreUpdate {
		// We reset duration at every barline
		if reflect.TypeOf(update) == barlineType {
			processor.resetDuration()
			return update
		}

		duration := getDuration(update)

		// If there are no duration components, we just skip
		if len(duration.Components) == 0 {
			return update
		}

		// If the duration contains a barline, we reset
		for _, component := range duration.Components {
			if reflect.TypeOf(component) == barlineType {
				processor.resetDuration()
				return update
			}
		}

		if processor.hasDuration() {
			difference := deep.Equal(processor.currentDuration, duration)
			if len(difference) == 0 {
				// This is a repeated duration, we set it to nil
				update = setDuration(update, model.Duration{})
			}
		}

		processor.currentDuration = duration
		return update
	}

	for i, update := range updates {
		update = modify(update)

		// Recurse process through nested score updates
		if _, ok := getNestedUpdates(update, false); ok {
			// We reset upon entering and returning from a nested layer
			processor.resetDuration()

			modified, _ := modifyNestedUpdates(update, processor.processAll)
			update = modified

			processor.resetDuration()
		}

		updates[i] = update
	}

	return updates
}
