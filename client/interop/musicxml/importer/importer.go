package importer

import (
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"github.com/beevik/etree"
	"github.com/logrusorgru/aurora"
	"io"
	"reflect"
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

// musicXMLPart contains part-specific information necessary for import
type musicXMLPart struct {
	instruments  []string
	voices       map[int32]*musicXMLVoice
	currentVoice *musicXMLVoice
	divisions    int
	beats        float64
}

func newMusicXMLPart() *musicXMLPart {
	return &musicXMLPart{
		voices:    make(map[int32]*musicXMLVoice),
		divisions: 1, // Divisions for a part are set in the first measure
		beats:     0,
	}
}

// musicXMLVoice contains voice-specific information necessary for import
type musicXMLVoice struct {
	updates     []model.ScoreUpdate
	beats       float64
	octave      int64
	// Alda notes contain only a pitch, not an octave
	// We must keep track of the current octave of a voice while importing
	// We also need to maintain the startOctave to facilitate building repeats
	startOctave int64
	// slurs are represented by start and stop tags
	// slurs can be nested, so we track the nested depth as an integer
	// Then a note is slurred if the depth is strictly greater than 0
	slurs       int64
}

func newMusicXMLVoice() *musicXMLVoice {
	return &musicXMLVoice{
		beats:       0,
		octave:      4, // 4 is the default Alda octave
		startOctave: 4,
		slurs:       0, // By default a note is not slurred
	}
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
	voiceUpdates := importer.currentPart.currentVoice.updates

	if eventSequence, ok := getLastRepeatEventSequence(voiceUpdates); ok {
		return eventSequence.Events
	} else {
		return voiceUpdates
	}
}

// append appends an update to the current voice
// append increments both part-level and voice-level beats
func (importer *musicXMLImporter) append(newUpdates ...model.ScoreUpdate) {
	voiceUpdates := importer.currentPart.currentVoice.updates

	beats := getBeats(newUpdates...)
	importer.currentPart.beats += beats
	importer.currentPart.currentVoice.beats += beats

	if eventSequence, ok := getLastRepeatEventSequence(voiceUpdates); ok {
		eventSequence.Events = append(
			eventSequence.Events,
			newUpdates...
		)
	} else {
		importer.currentPart.currentVoice.updates = append(
			importer.currentPart.currentVoice.updates,
			newUpdates...,
		)
	}
}

// set sets the updates for the current voice
// set recomputes both part-level and voice-level beats
func (importer *musicXMLImporter) set(newUpdates []model.ScoreUpdate) {
	voiceUpdates := importer.currentPart.currentVoice.updates

	beats := getBeats(newUpdates...)
	importer.currentPart.beats = beats
	importer.currentPart.currentVoice.beats = beats

	if eventSequence, ok := getLastRepeatEventSequence(voiceUpdates); ok {
		eventSequence.Events = newUpdates
	} else {
		importer.currentPart.currentVoice.updates = newUpdates
	}
}

// Identifies if the last update is a repeat, and returns its event sequence
func getLastRepeatEventSequence(
	updates []model.ScoreUpdate,
) (*model.EventSequence, bool) {
	if len(updates) > 0 {
		switch value := updates[len(updates) - 1].(type) {
		case model.Repeat:
			// We can cast to an event sequence because in our importer we only
			// ever repeat sequences, never individual notes
			eventSequence := value.Event.(model.EventSequence)
			return &eventSequence, true
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

var noteType = reflect.TypeOf(model.Note{})
var chordType = reflect.TypeOf(model.Chord{})
var restType = reflect.TypeOf(model.Rest{})
var attributeUpdateType = reflect.TypeOf(model.AttributeUpdate{})
var octaveUpType = reflect.TypeOf(model.OctaveUp{})
var octaveDownType = reflect.TypeOf(model.OctaveDown{})

func warnWhileParsing(element *etree.Element, message string) {
	log.Warn().
		Str("tag", element.Tag).
		Str("issue", message).
		Msg("Issue parsing tag")
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
 	postProcess(importer)

	return importer.generateScoreUpdates(), nil
}
