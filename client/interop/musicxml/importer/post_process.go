package importer

import (
	"alda.io/client/model"
	"github.com/go-test/deep"
	"reflect"
)

// postProcess applies various modifier methods to generate more idiomatic Alda
func postProcess(updates []model.ScoreUpdate) []model.ScoreUpdate {
	processor := newPostProcessor()
	return processor.processAll(updates)
}

// modifier is a function applied to an individual update in postprocessing
type modifier = func(
	update model.ScoreUpdate, processor *postProcessor,
) model.ScoreUpdate

var modifiers = []modifier{removeRepeatedDurations, removeRedundantAccidentals}

type postProcessor struct {
	// currentDuration stores the last duration encountered in a note or rest
	// This lets us remove repeated durations which Alda tracks automatically
	// Note that we do not remove repeated durations for ties such as "c4~4"
	currentDuration float64

	// currentKeySignature stores the last key signature set
	// This lets us remove unnecessary accidentals that are covered by the key
	currentKeySignature model.KeySignature

	// But we do not want to remove all accidentals covered by the key
	// Accidentals that appear in the same measure after a different accidental
	// should be kept to state that we are returning to normal
	// For example, if there is a natural, then we "undo" this natural later in
	// the bar, we want this "undo" accidental to remain
	// This is just to follow musical convention, and is not necessary in Alda

	// currentNoteState will track for each note whether it has diverged from
	// the key and so the next accidental that returns to the key will be kept
	currentNoteState map[model.NoteLetter]bool
}

func newPostProcessor() postProcessor {
	processor := postProcessor{
		currentDuration:     4,
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

func (processor *postProcessor) processAll(
	updates []model.ScoreUpdate,
) []model.ScoreUpdate {
	for i, update := range updates {
		// Apply modifiers
		for _, modifier := range modifiers {
			update = modifier(update, processor)
		}

		// Recurse process through nested score updates
		if modified, ok := modifyNestedUpdates(update, processor.processAll); ok {
			update = modified
		}

		updates[i] = update
	}
	return updates
}

func removeRedundantAccidentals(
	update model.ScoreUpdate, processor *postProcessor,
) model.ScoreUpdate {
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
			noteLetter := typedPitchIdentifier.NoteLetter
			noteAccidentals := typedPitchIdentifier.Accidentals
			keySignatureAccidentals := processor.currentKeySignature[noteLetter]

			different := len(deep.Equal(
				noteAccidentals, keySignatureAccidentals,
			)) > 0

			if different {
				// When accidentals are different than the key, we just leave
				// them and update our current note state
				processor.currentNoteState[noteLetter] = true
			} else {
				if !processor.currentNoteState[noteLetter] {
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

func removeRepeatedDurations(
	update model.ScoreUpdate, processor *postProcessor,
) model.ScoreUpdate {
	getDurationComponents := func(
		update model.ScoreUpdate,
	) []model.DurationComponent {
		if reflect.TypeOf(update) == noteType {
			return update.(model.Note).Duration.Components
		} else if reflect.TypeOf(update) == restType {
			return update.(model.Rest).Duration.Components
		} else {
			return nil
		}
	}

	setDurationComponents := func(
		update model.ScoreUpdate, components []model.DurationComponent,
	) model.ScoreUpdate {
		switch typedUpdate := update.(type) {
		case model.Note:
			typedUpdate.Duration.Components = components
			return typedUpdate
		case model.Rest:
			typedUpdate.Duration.Components = components
			return typedUpdate
		default:
			return update
		}
	}

	durationComponents := getDurationComponents(update)

	// We do preprocessing to obtain some necessary information
	var lastNoteLength model.NoteLength
	var lastNoteLengthIndex int
	noteLengthCount := 0

	for i, duration := range durationComponents {
		if reflect.TypeOf(duration) == noteLengthType {
			lastNoteLength = duration.(model.NoteLength)
			lastNoteLengthIndex = i
			noteLengthCount++
		}
	}

	if noteLengthCount == 1 &&
		lastNoteLength.Denominator == processor.currentDuration {
		// Only in this specific case do we have a repeated duration to remove
		// If noteLengthCount > 1, then we are dealing with ties
		durationComponents = append(
			durationComponents[:lastNoteLengthIndex],
			durationComponents[lastNoteLengthIndex + 1:]...
		)
	}

	processor.currentDuration = lastNoteLength.Denominator
	update = setDurationComponents(update, durationComponents)
	return update
}