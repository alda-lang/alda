package importer

import (
	"alda.io/client/help"
	"alda.io/client/model"
	"github.com/beevik/etree"
	"github.com/logrusorgru/aurora"
	"io"
	"sort"
)

// musicXMLImporter contains global state for importing a MusicXML file
type musicXMLImporter struct {
	parts       map[string]*musicXMLPart
	currentPart *musicXMLPart

	// unsupported stores the tags of unsupported elements that have been
	// encountered, so we don't warn the user multiple times for each tag type
	unsupported []string
}

func newMusicXMLImporter() *musicXMLImporter {
	return &musicXMLImporter{
		parts: make(map[string]*musicXMLPart),
	}
}

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
		updates = append(updates, part.generateScoreUpdates()...)
	}

	return updates
}

// musicXMLPart contains part-specific information necessary for import
type musicXMLPart struct {
	instruments  []string
	voices       map[int32]*musicXMLVoice
	currentVoice *musicXMLVoice
	divisions    int
	beats        float64

	// We maintain the following information for unpitched percussion parts
	// unpitched maps a part's instrument ID to the corresponding MIDI pitch
	unpitched map[string]int32
	alias     string
}

func newMusicXMLPart() *musicXMLPart {
	return &musicXMLPart{
		voices:    make(map[int32]*musicXMLVoice),
		divisions: 1, // Divisions for a part are set in the first measure
		beats:     0,
		unpitched: make(map[string]int32),
		alias:     "",
	}
}

func (part *musicXMLPart) generateScoreUpdates() []model.ScoreUpdate {
	partDeclaration := model.PartDeclaration{
		Names: part.instruments,
		Alias: part.alias,
	}

	updates := []model.ScoreUpdate{partDeclaration}

	if len(part.voices) == 1 {
		// For a single voice, we don't include a voice marker
		updates = append(
			updates, part.currentVoice.generateScoreUpdates()...,
		)
	} else {
		// Process voices in order of voice number
		var voiceNumber int32 = 1
		for voicesLeft := len(part.voices); voicesLeft > 0; voiceNumber++ {
			if voice, ok := part.voices[voiceNumber]; ok {
				voiceMarker := model.VoiceMarker{
					VoiceNumber: voiceNumber,
				}
				updates = append(updates, voiceMarker)
				updates = append(
					updates,
					voice.generateScoreUpdates()...,
				)
				voicesLeft--
			}
		}
	}
	return updates
}

// musicXMLVoice contains voice-specific information necessary for import
type musicXMLVoice struct {
	updates []model.ScoreUpdate
	beats   float64
	octave  int64

	// slurs are represented by start and stop tags
	// slurs can be nested, so we track the nested depth as an integer
	// Then a note is slurred if the depth is strictly greater than 0
	slurs int64

	// Alda notes contain only a pitch, not an octave
	// We must then keep track of octave information to handle repeats
	// See barlineHandler for how this is managed
	sectionStartOctave int64
	endingStartOctave  int64
}

func newMusicXMLVoice() *musicXMLVoice {
	return &musicXMLVoice{
		// We import into an event sequence so we can always use recursion
		updates:     []model.ScoreUpdate{model.EventSequence{}},
		beats:       0,
		octave:      4, // 4 is the default Alda octave
		slurs:       0, // By default a note is not slurred

		// To handle octave setting in repeats
		sectionStartOctave: 4,
		endingStartOctave:  4,
	}
}

func (voice *musicXMLVoice) generateScoreUpdates() []model.ScoreUpdate {
	return voice.updates[0].(model.EventSequence).Events
}

func ImportMusicXML(r io.Reader) ([]model.ScoreUpdate, error) {
	doc := etree.NewDocument()

	_, err := doc.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	scorePartwise := doc.SelectElement("score-partwise")
	scoreTimewise := doc.SelectElement("score-timewise")

	if scorePartwise == nil && scoreTimewise != nil {
		return nil, help.UserFacingErrorf(
			`Issue importing MusicXML file: please convert to %s instead of %s using XSLT before importing`,
			aurora.BrightYellow("score-partwise"),
			aurora.BrightYellow("score-timewise"),
		)
	} else if scorePartwise == nil {
		return nil, help.UserFacingErrorf(
			`Issue importing MusicXML file: could not last %s root tag`,
			aurora.BrightYellow("score-partwise"),
		)
	}

	importer := newMusicXMLImporter()
	handle(scorePartwise, importer)

	// We postprocess the updates for each voice to generate more idiomatic Alda
	for _, part := range importer.parts {
		for _, voice := range part.voices {
			voice.updates = postProcess(voice.updates)
		}
	}
	return importer.generateScoreUpdates(), nil
}

func (importer *musicXMLImporter) part() *musicXMLPart {
	return importer.currentPart
}

func (importer *musicXMLImporter) voice() *musicXMLVoice {
	return importer.currentPart.currentVoice
}

// flatUpdates returns the base level updates of the current voice
func (importer *musicXMLImporter) flatUpdates() []model.ScoreUpdate {
	return importer.voice().updates[0].(model.EventSequence).Events
}

// updates returns the current list of model.ScoreUpdate to import into
func (importer *musicXMLImporter) updates() []model.ScoreUpdate {
	// We can use findLast to recursively determine the current updates
	importInto, _ := importer.findLast(filterNestedImportableUpdate)
	nestedUpdates, _ := getNestedUpdates(importInto, true)
	return nestedUpdates
}

// last returns the last update in the current active list of model.ScoreUpdate
func (importer *musicXMLImporter) last() model.ScoreUpdate {
	updates := importer.updates()
	return updates[len(updates) - 1]
}

// append appends updates to the current slice to import into
func (importer *musicXMLImporter) append(newUpdates ...model.ScoreUpdate) {
	_, ni := importer.findLast(filterNestedImportableUpdate)
	importer.modifyAt(
		ni,
		func(update model.ScoreUpdate) model.ScoreUpdate {
			modified, _ := modifyNestedUpdates(
				update,
				func(updates []model.ScoreUpdate) []model.ScoreUpdate {
					return append(updates, newUpdates...)
				},
			)
			return modified
		},
	)
}

// setAll replaces the updates for the current slice to import into
func (importer *musicXMLImporter) setAll(newUpdates []model.ScoreUpdate) {
	_, ni := importer.findLast(filterNestedImportableUpdate)
	importer.modifyAt(
		ni,
		func(update model.ScoreUpdate) model.ScoreUpdate {
			modified, _ := setNestedUpdates(update, newUpdates)
			return modified
		},
	)
}

func (importer *musicXMLImporter) recountBeats() {
	beats := getBeats(importer.voice().updates...)
	importer.part().beats = beats
	importer.voice().beats = beats
}

// modifyAt modifies an update at a specific nestedIndex
// modifyAt is the central method in the importer for dealing with updating
// []model.ScoreUpdate slices while importing
func (importer *musicXMLImporter) modifyAt(
	ni nestedIndex, modify func(update model.ScoreUpdate) model.ScoreUpdate,
) {
	if len(ni.indices) == 1 {
		// We're modifying the base voice event sequence
		importer.voice().updates[0] = modify(importer.voice().updates[0])
	} else {
		// We recurse, each time going one index deeper in the nestedIndex
		depth := 1
		var modifyAtNested func(updates []model.ScoreUpdate) []model.ScoreUpdate
		modifyAtNested = func(updates []model.ScoreUpdate) []model.ScoreUpdate {
			if depth == len(ni.indices) - 1 {
				// Base level
				updates[ni.indices[depth]] = modify(updates[ni.indices[depth]])
				return updates
			} else {
				// Recurse
				currDepth := depth
				depth++
				modifiedUpdate, _ := modifyNestedUpdates(
					updates[ni.indices[currDepth]],
					modifyAtNested,
				)
				updates[ni.indices[currDepth]] = modifiedUpdate
				return updates
			}
		}

		newEventSequence, _ := modifyNestedUpdates(
			importer.voice().updates[0],
			modifyAtNested,
		)
		importer.voice().updates[0] = newEventSequence
	}
	importer.recountBeats()
}

// getAt returns the update at a nestedIndex
func (importer *musicXMLImporter) getAt(ni nestedIndex) model.ScoreUpdate {
	var foundUpdate model.ScoreUpdate = nil
	importer.modifyAt(ni, func(update model.ScoreUpdate) model.ScoreUpdate {
		foundUpdate = update
		return update
	})
	return foundUpdate
}

// setAt sets the update at a nestedIndex
func (importer *musicXMLImporter) setAt(
	ni nestedIndex, update model.ScoreUpdate,
) {
	importer.modifyAt(ni, func(_ model.ScoreUpdate) model.ScoreUpdate {
		return update
	})
}

// findLast finds the last update that passes filter
func (importer *musicXMLImporter) findLast(
	filter func(update model.ScoreUpdate) bool,
) (model.ScoreUpdate, nestedIndex) {
	updates := importer.voice().updates
	return findLastRecursive(updates, filter)
}

// findLastFrom finds the last update that passes filter, but starting from a
// specific index in flatUpdates
func (importer *musicXMLImporter) findLastFrom(
	filter func(update model.ScoreUpdate) bool,
	startIndex int,
) (model.ScoreUpdate, nestedIndex) {
	found, ni := findLastRecursive(importer.flatUpdates()[:startIndex], filter)
	if len(ni.indices) > 0 {
		return found, nestedIndex{indices: append([]int{0}, ni.indices...)}
	}
	return found, ni
}

func findLastRecursive(
	updates []model.ScoreUpdate,
	filter func(updates model.ScoreUpdate) bool,
) (model.ScoreUpdate, nestedIndex) {
	for i := len(updates) - 1; i >= 0; i-- {
		if filter(updates[i]) {
			return updates[i], nestedIndex{indices: []int{i}}
		}

		if nested, ok := getNestedUpdates(updates[i], false); ok {
			update, ni := findLastRecursive(nested, filter)
			if len(ni.indices) > 0 {
				return update, nestedIndex{
					indices: append([]int{i}, ni.indices...),
				}
			}
		}
	}
	return nil, nestedIndex{indices: nil}
}

// findLastWithState finds the last update to pass filter
// findLastWithState modifies a state as it searches
func (importer *musicXMLImporter) findLastWithState(
	filter func(update model.ScoreUpdate, state interface{}) bool,
	initialState interface{},
	updateState func(update model.ScoreUpdate, state interface{}) interface{},
) (model.ScoreUpdate, nestedIndex, interface{}) {
	updates := importer.voice().updates
	return findLastWithStateRecursive(
		updates, filter, initialState, updateState,
	)
}

func findLastWithStateRecursive(
	updates []model.ScoreUpdate,
	filter func(updates model.ScoreUpdate, state interface{}) bool,
	initialState interface{},
	updateState func(
		update model.ScoreUpdate, state interface{},
	) interface{},
) (model.ScoreUpdate, nestedIndex, interface{}) {
	curr := initialState
	for i := len(updates) - 1; i >= 0; i-- {
		if filter(updates[i], curr) {
			return updates[i], nestedIndex{indices: []int{i}}, curr
		}

		curr = updateState(updates[i], curr)

		if nested, ok := getNestedUpdates(updates[i], false); ok {
			update, ni, state := findLastWithStateRecursive(
				nested, filter, curr, updateState,
			)
			curr = state
			if len(ni.indices) > 0 {
				return update, nestedIndex{
					indices: append([]int{i}, ni.indices...),
				}, curr
			}
		}
	}
	return nil, nestedIndex{indices: nil}, curr
}
