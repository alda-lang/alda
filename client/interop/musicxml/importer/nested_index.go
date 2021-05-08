package importer

import (
	"alda.io/client/model"
	"reflect"
)

// nestedIndex identifies a model.ScoreUpdate in a []model.ScoreUpdate
// nestedIndex is designed to work with slices as built by musicXMLImporter
// all parameters must be set to -1 if not found / not used
type nestedIndex struct {
	index      int
	chordIndex int
}

// lastWithState finds the last element in updates that satisfies filter
// lastWithState continuously updates a current state while processing updates
// lastWithState returns the final state
func lastWithState(
	updates []model.ScoreUpdate,
	filter func(update model.ScoreUpdate, state interface{}) bool,
	initial interface{},
	updateState func(update model.ScoreUpdate, state interface{}) interface{},
) (nestedIndex, interface{}) {
	curr := initial
	for i := len(updates) - 1; i >= 0; i-- {
		if filter(updates[i], curr) {
			return nestedIndex{index: i, chordIndex: -1}, curr
		}
		curr = updateState(updates[i], curr)
		switch value := updates[i].(type) {
		case model.Chord:
			ni, state := lastWithState(
				value.Events, filter, curr, updateState,
			)
			curr = state
			if ni.index != -1 {
				return nestedIndex{index: i, chordIndex: ni.index}, curr
			}
		}
	}
	return nestedIndex{index: -1, chordIndex: -1}, curr
}

// last finds the last element in updates that satisfies filter
func last(
	updates []model.ScoreUpdate, filter func(update model.ScoreUpdate) bool,
) nestedIndex {
	for i := len(updates) - 1; i >= 0; i-- {
		if filter(updates[i]) {
			return nestedIndex{index: i, chordIndex: -1}
		}
		switch value := updates[i].(type) {
		case model.Chord:
			ni := last(value.Events, filter)
			if ni.index != -1 {
				return nestedIndex{index: i, chordIndex: ni.index}
			}
		}
	}
	return nestedIndex{index: -1, chordIndex: -1}
}

// filterUpdateWithDuration only takes elements with a duration
func filterUpdateWithDuration(update model.ScoreUpdate) bool {
	return reflect.TypeOf(update) == noteType ||
		reflect.TypeOf(update) == restType ||
		reflect.TypeOf(update) == chordType
}

// apply returns updates with a change applied at the nestedIndex provided
func apply(
	updates []model.ScoreUpdate, ni nestedIndex,
	change func(update model.ScoreUpdate) model.ScoreUpdate,
) []model.ScoreUpdate {
	if ni.index == -1 {
		return updates
	}

	if ni.chordIndex == -1 {
		updates[ni.index] = change(updates[ni.index])
	} else {
		updates[ni.index].(model.Chord).Events[ni.chordIndex] = change(
			updates[ni.index].(model.Chord).Events[ni.chordIndex],
		)
	}
	return updates
}

// insert returns updates with an update inserted at the nestedIndex provided
func insert(
	updates []model.ScoreUpdate, ni nestedIndex, update model.ScoreUpdate,
) []model.ScoreUpdate {
	if ni.chordIndex == -1 {
		// Make space for the new element
		updates = append(updates, model.Note{})
		// Shift over
		copy(updates[ni.index + 1:], updates[ni.index:])
		// Set inserted element
		updates[ni.index] = update
		return updates
	} else {
		chord := updates[ni.index].(model.Chord)
		events := insert(
			chord.Events,
			nestedIndex{index: ni.chordIndex, chordIndex: -1},
			update,
		)
		chord.Events = events
		updates[ni.index] = chord
		return updates
	}
}
