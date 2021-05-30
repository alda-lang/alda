package importer

import (
	"alda.io/client/model"
	"reflect"
	"sort"
)

var noteType = reflect.TypeOf(model.Note{})
var chordType = reflect.TypeOf(model.Chord{})
var restType = reflect.TypeOf(model.Rest{})
var attributeUpdateType = reflect.TypeOf(model.AttributeUpdate{})
var octaveUpType = reflect.TypeOf(model.OctaveUp{})
var octaveDownType = reflect.TypeOf(model.OctaveDown{})
var repeatType = reflect.TypeOf(model.Repeat{})

func (importer *musicXMLImporter) part() *musicXMLPart {
	return importer.currentPart
}

func (importer *musicXMLImporter) voice() *musicXMLVoice {
	return importer.currentPart.currentVoice
}

// generateScoreUpdates flattens score updates in an importer to a single slice
// This represents the overall output of the importer
func (importer *musicXMLImporter) generateScoreUpdates() []model.ScoreUpdate {
	var updates []model.ScoreUpdate

	// We process parts in order of ID
	var partIDs []string
	for id := range importer.parts {
		partIDs = append(partIDs, id)
	}

	sort.Sort(sort.StringSlice(partIDs))

	for _, id := range partIDs {
		part := importer.parts[id]
		partDeclaration := model.PartDeclaration{
			Names: part.instruments,
		}
		updates = append(updates, partDeclaration)

		if len(part.voices) == 1 {
			// For a single voice, we don't include a voice marker
			updates = append(updates, part.currentVoice.updates...)
		} else {
			// Process voices in order of voice number
			var voiceNumber int32 = 1
			for voicesLeft := len(part.voices); voicesLeft > 0; voiceNumber++ {
				if voice, ok := part.voices[voiceNumber]; ok {
					voiceMarker := model.VoiceMarker{
						VoiceNumber: voiceNumber,
					}
					updates = append(updates, voiceMarker)
					updates = append(updates, voice.updates...)
					voicesLeft--
				}
			}
		}
	}

	return updates
}

// updates returns the current list of model.ScoreUpdate to import into
func (importer *musicXMLImporter) updates() []model.ScoreUpdate {
	if repeat, ok := getLastRepeat(importer.voice().updates); ok {
		eventSequence := repeat.Event.(model.EventSequence)
		if repetition, ok := isRepeatCurrentlyMultipleEndings(repeat); ok {
			return repetition.Event.(model.EventSequence).Events
		} else {
			return eventSequence.Events
		}
	} else {
		return importer.voice().updates
	}
}

// append appends an update to the current voice
// append increments both part-level and voice-level beats
func (importer *musicXMLImporter) append(newUpdates ...model.ScoreUpdate) {
	beats := getBeats(newUpdates...)
	importer.currentPart.beats += beats
	importer.currentPart.currentVoice.beats += beats

	if repeat, ok := getLastRepeat(importer.voice().updates); ok {
		if repetition, ok := isRepeatCurrentlyMultipleEndings(repeat); ok {



		} else {
			repeat
			switch eventSequence := repeat.Event.(type) {
			case model.EventSequence:
				eventSequence.Events = append(
					eventSequence.Events,
					newUpdates...
				)
			}
			voiceUpdates[len(voiceUpdates) - 1] = repeat
		}
		importer.voice().updates[len(importer.voice().updates) - 1] = repeat
	} else {
		importer.voice().updates = append(
			importer.voice().updates,
			newUpdates...,
		)
	}
}


// To propagate updates up the chain
func recursiveAppend(
	updates []model.ScoreUpdate, newUpdates ...model.ScoreUpdate,
) []model.ScoreUpdate {
	if len(updates) == 0 {
		return newUpdates
	} else {
		switch value := updates[len(updates) - 1].(type) {
		case model.Repeat:
			eventSequence := value.Event.(model.EventSequence)
			eventSequence.Events = recursiveAppend(updates, newUpdates...)

			value.Event = eventSequence
			updates[len(updates) - 1] = value
			return updates
		case model.OnRepetitions:
			eventSequence := value.Event.(model.EventSequence)
			eventSequence.Events = recursiveAppend(updates, newUpdates...)

			value.Event = eventSequence
			updates[len(updates) - 1] = value
			return updates
		default:
			return append(updates, newUpdates...)
		}
	}
}

func recursiveSet(
	updates []model.ScoreUpdate, newUpdates []model.ScoreUpdate,
) []model.ScoreUpdate {
	if len(updates) == 0 {
		return newUpdates
	} else {
		switch value := updates[len(updates) - 1].(type) {
		case model.Repeat:
			eventSequence := value.Event.(model.EventSequence)
			eventSequence.Events = recursiveSet(updates, newUpdates)
		}
	}

}

// set sets the updates for the current voice
// set recomputes both part-level and voice-level beats
func (importer *musicXMLImporter) set(newUpdates []model.ScoreUpdate) {
	voiceUpdates := importer.currentPart.currentVoice.updates

	beats := getBeats(newUpdates...)
	importer.part().beats = beats
	importer.voice().beats = beats

	if repeat, ok := getLastRepeat(voiceUpdates); ok {
		eventSequence := repeat.Event.(model.EventSequence)
		eventSequence.Events = newUpdates
	} else {
		importer.currentPart.currentVoice.updates = newUpdates
	}
}

// getLastRepeat identifies if the last update is an ongoing repeat
// getLastRepeat returns the repeat if it exists
func getLastRepeat(
	updates []model.ScoreUpdate,
) (*model.Repeat, bool) {
	if len(updates) > 0 {
		switch value := updates[len(updates) - 1].(type) {
		case model.Repeat:
			// We assume if times has been set, the repeat is finished
			if value.Times == 0 {
				// We can cast to an event sequence because in our importer we
				// only ever repeat sequences, never individual notes
				return &value, true
			}
		}
	}
	return nil, false
}

// isRepeatCurrentlyMultipleEndings identifies if a repeat has multiple endings
// isRepeatCurrentlyMultipleEndings returns the last repetition if it exists
func isRepeatCurrentlyMultipleEndings(
	repeat *model.Repeat,
) (*model.OnRepetitions, bool) {
	eventSequence := repeat.Event.(model.EventSequence)
	if len(eventSequence.Events) > 0 {
		switch value := eventSequence.Events[len(eventSequence.Events) - 1].(type) {
		case model.OnRepetitions:
			return &value, true
		}
	}
	return nil, false
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
