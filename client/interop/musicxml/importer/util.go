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

func getNestedUpdates(
	update model.ScoreUpdate, toImport bool,
) ([]model.ScoreUpdate, bool) {
	switch value := update.(type) {
	case model.Repeat:
		return value.Event.(model.EventSequence).Events, true
	case model.OnRepetitions:
		return value.Event.(model.EventSequence).Events, true
	case model.Chord:
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
