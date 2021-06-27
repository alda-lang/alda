package importer

import (
	log "alda.io/client/logging"
	"alda.io/client/model"
	"github.com/beevik/etree"
	"reflect"
	"strings"
)

var noteType = reflect.TypeOf(model.Note{})
var chordType = reflect.TypeOf(model.Chord{})
var restType = reflect.TypeOf(model.Rest{})
var attributeUpdateType = reflect.TypeOf(model.AttributeUpdate{})
var octaveUpType = reflect.TypeOf(model.OctaveUp{})
var octaveDownType = reflect.TypeOf(model.OctaveDown{})
var repeatType = reflect.TypeOf(model.Repeat{})
var repetitionType = reflect.TypeOf(model.OnRepetitions{})
var barlineType = reflect.TypeOf(model.Barline{})

var noteLetters = []model.NoteLetter{
	model.A, model.B, model.C, model.D, model.E, model.F, model.G,
}

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

// insert is a helper to insert an element in a slice
// insert returns the updated slice with element inserted at the provided index
func insert(
	update model.ScoreUpdate, updates []model.ScoreUpdate, index int,
) []model.ScoreUpdate {
	// Make space
	updates = append(updates, model.Note{})
	// Shift over
	copy(updates[index + 1:], updates[index:])
	// Set inserted element
	updates[index] = update
	return updates
}


// getBeats counts beats for a slice of model.ScoreUpdate
func getBeats(updates ...model.ScoreUpdate) float64 {
	beats := 0.0
	for _, update := range updates {
		switch value := update.(type) {
		case model.Note:
			beats += value.Duration.Beats()
		case model.Rest:
			beats += value.Duration.Beats()
		case model.Chord:
			min := 0.0
			for _, event := range value.Events {
				eventBeats := getBeats(event)
				if eventBeats < min {
					min = eventBeats
				}
			}
			beats += min
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
		last := nested[len(nested) - 1]
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

func percussionPartNameToAlias(name string) string {
	return strings.Join(strings.Split(name, " "), "_")
}
