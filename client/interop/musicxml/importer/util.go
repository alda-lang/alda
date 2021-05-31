package importer

import (
	"alda.io/client/model"
	"reflect"
)

var noteType = reflect.TypeOf(model.Note{})
var chordType = reflect.TypeOf(model.Chord{})
var restType = reflect.TypeOf(model.Rest{})
var attributeUpdateType = reflect.TypeOf(model.AttributeUpdate{})
var octaveUpType = reflect.TypeOf(model.OctaveUp{})
var octaveDownType = reflect.TypeOf(model.OctaveDown{})
var repeatType = reflect.TypeOf(model.Repeat{})
var repetitionType = reflect.TypeOf(model.OnRepetitions{})

func getNestedUpdates(
	update model.ScoreUpdate, toImport bool,
) ([]model.ScoreUpdate, bool) {
	switch value := update.(type) {
	case model.Repeat:
		if toImport && value.Times > 0 {
			// We use a repeat's times to track whether the repeat has ended
			// If times > 0, it has been set at the repeat ending
			// Then we do not continue to import into the repeat
			break
		}
		return value.Event.(model.EventSequence).Events, true
	case model.OnRepetitions:
		// We use repetitions to determine whether an ending is complete
		if toImport && len(value.Repetitions) > 0 {
			break
		}
		return value.Event.(model.EventSequence).Events, true
	case model.Chord:
		// We never import directly into a chord
		if !toImport {
			return value.Events, true
		}
	}
	return nil, false
}

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
	}
	return update, false
}

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
			eventSequence := value.Event.(model.EventSequence)
			beats += getBeats(eventSequence.Events...)
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

// Finds an update which can be imported into
func filterNestedImportableUpdate(update model.ScoreUpdate) bool {
	if nested, ok := getNestedUpdates(update, true); ok {
		if len(nested) == 0 {
			return true
		}
		last := nested[len(nested) - 1]
		_, nestedOk := getNestedUpdates(last, true)

		// We stop if the nested layer does not have further nested layers
		return !nestedOk
	}
	return false
}