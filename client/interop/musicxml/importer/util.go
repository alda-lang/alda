package importer

import (
	"math"
	"reflect"
	"strings"

	log "alda.io/client/logging"
	"alda.io/client/model"
	"github.com/beevik/etree"
)

var noteType = reflect.TypeOf(model.Note{})
var chordType = reflect.TypeOf(model.Chord{})
var restType = reflect.TypeOf(model.Rest{})
var attributeUpdateType = reflect.TypeOf(model.AttributeUpdate{})

// Commenting these out because staticcheck is complaining that they're unused:
//
// var octaveUpType = reflect.TypeOf(model.OctaveUp{})
// var octaveDownType = reflect.TypeOf(model.OctaveDown{})

var repeatType = reflect.TypeOf(model.Repeat{})
var repetitionType = reflect.TypeOf(model.OnRepetitions{})
var barlineType = reflect.TypeOf(model.Barline{})
var lispListType = reflect.TypeOf(model.LispList{})
var midiNoteNumberType = reflect.TypeOf(model.MidiNoteNumber{})

// warnWhileParsing displays a standard importing warning to the user
func warnWhileParsing(element *etree.Element, message string) {
	log.Warn().
		Str("tag", element.Tag).
		Str("issue", message).
		Msg("Issue parsing tag")
}

// getNestedUpdates facilitates recursing on nested slices of updates in Alda IR
// getNestedUpdates abstracts recursive update properties from the importer
// getNestedUpdates returns the nested updates, and whether the update is nested
func getNestedUpdates(
	update model.ScoreUpdate, toImport bool,
) ([]model.ScoreUpdate, bool) {
	switch value := update.(type) {
	case model.Repeat:
		if toImport && value.Times > 0 {
			// We use a repeat's times to track whether the repeat has ended
			break
		}
		return value.Event.(model.EventSequence).Events, true
	case model.OnRepetitions:
		if toImport && len(value.Repetitions) > 0 {
			// We use a repetition's repetitions to track whether it has ended
			break
		}
		return value.Event.(model.EventSequence).Events, true
	case model.Chord:
		// We never import directly into a chord
		if !toImport {
			return value.Events, true
		}
	case model.EventSequence:
		return value.Events, true
	}
	return nil, false
}

// modifyNestedUpdates facilitates recursively modifying Alda IR
// modifyNestedUpdates returns the modified update, and success
func modifyNestedUpdates(
	update model.ScoreUpdate,
	modify func(updates []model.ScoreUpdate) []model.ScoreUpdate,
) (model.ScoreUpdate, bool) {
	switch value := update.(type) {
	case model.Repeat:
		eventSequence := value.Event.(model.EventSequence)
		eventSequence.Events = modify(eventSequence.Events)
		value.Event = eventSequence
		return value, true
	case model.OnRepetitions:
		eventSequence := value.Event.(model.EventSequence)
		eventSequence.Events = modify(eventSequence.Events)
		value.Event = eventSequence
		return value, true
	case model.Chord:
		value.Events = modify(value.Events)
		return value, true
	case model.EventSequence:
		value.Events = modify(value.Events)
		return value, true
	}
	return update, false
}

// setNestedUpdates is short-hand to modifyNestedUpdates but setting a new slice
func setNestedUpdates(
	update model.ScoreUpdate,
	updates []model.ScoreUpdate,
) (model.ScoreUpdate, bool) {
	return modifyNestedUpdates(
		update,
		func(_ []model.ScoreUpdate) []model.ScoreUpdate {
			return updates
		},
	)
}

// getBeats counts beats for a slice of model.ScoreUpdate
// Updates must be un-optimized so redundant durations are not yet removed
func getBeats(updates ...model.ScoreUpdate) float64 {
	beats := 0.0
	for _, update := range updates {
		switch value := update.(type) {
		case model.Note:
			beats += value.Duration.Beats()
		case model.Rest:
			beats += value.Duration.Beats()
		case model.Chord:
			min := math.MaxFloat64
			for _, event := range value.Events {
				eventBeats := getBeats(event)
				if eventBeats > 0 && eventBeats < min {
					// We take the minimum non-zero duration
					// i.e. a real note/rest, not an attribute update
					min = eventBeats
				}
			}
			if min < math.MaxFloat64 {
				beats += min
			}
		case model.Repeat:
			beats += getBeats(value.Event)
		case model.OnRepetitions:
			beats += getBeats(value.Event)
		case model.EventSequence:
			beats += getBeats(value.Events...)
		}
	}
	return beats
}

// insert is a helper to insert an element in a slice
// insert returns the updated slice with element inserted at the provided index
func insert(
	update model.ScoreUpdate, updates []model.ScoreUpdate, index int,
) []model.ScoreUpdate {
	// Make space
	updates = append(updates, model.Note{})
	// Shift over
	copy(updates[index+1:], updates[index:])
	// Set inserted element
	updates[index] = update
	return updates
}

// standardizeBarlines extracts any barlines that are the last duration
// component in a note or rest and inserts them directly after
// standardizeBarlines produces equivalent Alda, but makes the score updates
// easier for the importer to manipulate (both for tests and postprocessing)
func standardizeBarlines(updates []model.ScoreUpdate) []model.ScoreUpdate {
	for i := len(updates) - 1; i >= 0; i-- {
		barlineAfter := false

		removeBarline := func(
			durations []model.DurationComponent,
		) ([]model.DurationComponent, bool) {
			if len(durations) > 0 &&
				reflect.TypeOf(durations[len(durations)-1]) == barlineType {
				durations = durations[:len(durations)-1]
				if len(durations) == 0 {
					durations = nil
				}
				return durations, true
			}
			return nil, false
		}

		update := updates[i]
		switch typedUpdate := update.(type) {
		case model.Note:
			durations := typedUpdate.Duration.Components
			if updatedDurations, ok := removeBarline(durations); ok {
				typedUpdate.Duration.Components = updatedDurations
				update = typedUpdate
				barlineAfter = true
			}
		case model.Rest:
			durations := typedUpdate.Duration.Components
			if updatedDurations, ok := removeBarline(durations); ok {
				typedUpdate.Duration.Components = updatedDurations
				update = typedUpdate
				barlineAfter = true
			}
		}

		updates[i] = update
		if barlineAfter {
			updates = insert(model.Barline{}, updates, i+1)
		}

		// Recursively standardize barlines
		if modified, ok := modifyNestedUpdates(
			update, standardizeBarlines,
		); ok {
			updates[i] = modified
		}
	}

	return updates
}

// evaluateLisp evaluates all lisp expressions into plain score updates
func evaluateLisp(updates []model.ScoreUpdate) []model.ScoreUpdate {
	for i, update := range updates {
		if reflect.TypeOf(update) == lispListType {
			lispList := update.(model.LispList)
			lispForm, err := lispList.Eval()
			if err != nil {
				panic(err)
			}
			updates[i] = lispForm.(model.LispScoreUpdate).ScoreUpdate
		}

		if modified, ok := modifyNestedUpdates(update, evaluateLisp); ok {
			updates[i] = modified
		}
	}

	return updates
}

// filterUpdateWithDuration only takes elements with a duration
func filterUpdateWithDuration(update model.ScoreUpdate) bool {
	return reflect.TypeOf(update) == noteType ||
		reflect.TypeOf(update) == restType ||
		reflect.TypeOf(update) == chordType
}

// filterNestedImportableUpdate finds the update to import into
func filterNestedImportableUpdate(update model.ScoreUpdate) bool {
	if nested, ok := getNestedUpdates(update, true); ok {
		if len(nested) == 0 {
			return true
		}
		last := nested[len(nested)-1]
		_, nestedOk := getNestedUpdates(last, true)

		// We stop if the nested layer does not have further nested layers
		// In which we can then "import" into that nested layer
		return !nestedOk
	}
	return false
}

// filterType finds the last update of a specific reflect.Type
func filterType(requiredType reflect.Type) func(update model.ScoreUpdate) bool {
	return func(update model.ScoreUpdate) bool {
		return reflect.TypeOf(update) == requiredType
	}
}

// percussionPartNameToAlias translates a MusicXML instrument part name into a
// format acceptable for Alda instrument alias
func percussionPartNameToAlias(name string) string {
	return strings.Join(strings.Split(name, " "), "_")
}

// toLetterAndAccidentalsAndOctave translates a MidiNoteNumber to an equivalent
// pitch representation consisting of a LetterAndAccidentals and an octave
// toLetterAndAccidentalsAndOctave naively generates a note letter with either
// no accidentals or a sharp, and does not consider key signature or scale type
func toLetterAndAccidentalsAndOctave(
	number model.MidiNoteNumber,
) (model.LetterAndAccidentals, int32) {
	quotient := (number.MidiNote - 24) / 12
	remainder := (number.MidiNote - 24) % 12

	octave := quotient + 1

	// Find the closest letter than is above the current remainder
	noteLetter := model.C
	for letter, interval := range model.NoteLetterIntervals {
		if interval <= remainder &&
			interval > model.NoteLetterIntervals[noteLetter] {
			noteLetter = letter
		}
	}

	// Add sharps as necessary
	var sharps []model.Accidental

	for i := model.NoteLetterIntervals[noteLetter]; i < remainder; i++ {
		sharps = append(sharps, model.Sharp)
	}

	return model.LetterAndAccidentals{
		NoteLetter:  noteLetter,
		Accidentals: sharps,
	}, octave
}

// getNoteOrRestDuration gets the duration of a note or rest
func getNoteOrRestDuration(update model.ScoreUpdate) model.Duration {
	if reflect.TypeOf(update) == noteType {
		return update.(model.Note).Duration
	} else if reflect.TypeOf(update) == restType {
		return update.(model.Rest).Duration
	} else {
		return model.Duration{}
	}
}

// setNoteOrRestDuration sets the duration of a note or rest
func setNoteOrRestDuration(
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

// isOrContainsBarline returns whether an update is a barline or directly
// contains one in its duration (note / rest)
func isOrContainsBarline(update model.ScoreUpdate) bool {
	if reflect.TypeOf(update) == barlineType {
		return true
	}

	duration := getNoteOrRestDuration(update)
	for _, component := range duration.Components {
		if reflect.TypeOf(component) == barlineType {
			return true
		}
	}

	return false
}

// idiomaticDuration applies simple heuristics to make durations more idiomatic
func idiomaticDuration(aldaDuration float64, dots int32) model.Duration {
	aldaDuration = roundIfCloseEnough(aldaDuration)

	// We preserve actual dotted notes to not cause any confusion
	if dots == 0 {
		// Otherwise, we try to decompose the duration into "normal" durations
		normal := []float64{1, 2, 4, 8, 16}

		remaining := aldaDuration
		decomp := make(map[float64]int)
		for i := 0; i < len(normal); {
			if math.Abs(remaining-normal[i]) < 0.01 {
				decomp[normal[i]] = decomp[normal[i]] + 1
				remaining = 0
				break
			} else if remaining < normal[i] {
				decomp[normal[i]] = decomp[normal[i]] + 1
				// e.g. remaining = 1.3333
				// 1/1.3333 = 3/4, so idiomatically this should be 2~4 or 2.
				// 1/((1/1.3333)-(1/2)) = 1/4, we remove the half note duration
				// We decompose into a half note, then leave a quarter remaining
				remaining = 1 / ((1 / remaining) - (1 / normal[i]))
			} else {
				i++
			}
		}

		if remaining == 0 {
			// We try to represent a decomposition as a single dotted duration
			if len(decomp) == 2 {
				for i := 0; i+1 < len(normal); i++ {
					if decomp[normal[i]] == 1 && decomp[normal[i+1]] == 1 {
						return model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{
									Denominator: normal[i],
									Dots:        1,
								},
							},
						}
					}
				}
			}

			// Otherwise, we just tie together each decomposed normal duration
			components := []model.DurationComponent{}
			for _, d := range normal {
				for decomp[d] > 0 {
					components = append(components, model.NoteLength{
						Denominator: d,
					})
					decomp[d] = decomp[d] - 1
				}
			}
			return model.Duration{Components: components}
		}
	}

	return model.Duration{
		Components: []model.DurationComponent{model.NoteLength{
			Denominator: aldaDuration,
			Dots:        dots,
		}},
	}
}

// roundIfCloseEnough rounds to deal with floating point arithmetic issues
// For long scores, arithmetic can otherwise lead to values like r1407374883...
func roundIfCloseEnough(val float64) float64 {
	if math.Abs(val-math.Round(val)) < 0.0001 {
		return math.Round(val)
	} else {
		return val
	}
}
