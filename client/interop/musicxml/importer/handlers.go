package importer

import (
	log "alda.io/client/logging"
	"alda.io/client/model"
	"fmt"
	"github.com/beevik/etree"
	"github.com/go-test/deep"
	"github.com/logrusorgru/aurora"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// elementHandler is a function that can handle the import of a MusicXML element
// All MusicXML elements are grouped into 1 of 3 categories
//
// (1) Elements that are supported (either fully or partially)
//
// (2) Elements that are not (yet) supported, dealing primarily with music sound
// For these elements, a user can logically expect them to be imported into Alda
// Either they are not yet supported (and plan to be), or they have no
// equivalent Alda representation
// We log a warning message to the user when importing these elements
//
// (3) Elements that can't be supported, dealing primarily with music appearance
// Elements in this category have no representation in Alda
// These elements are skipped automatically, alongside elements with no category
type elementHandler = func(element *etree.Element, importer *musicXMLImporter)

// handlers stores the elementHandler functions for most imported elements
// Elements in category (3) or those used directly in the handling of an
// ancestor tag (e.x. duration, pitch, cue) are excluded from this map
var handlers map[string]elementHandler

func init() {
	handlers = map[string]elementHandler{
		"score-partwise": recurseHandler,
		"part-list":      recurseHandler,
		"score-part":     scorePartHandler,
		"part":           partHandler,
		"measure":        measureHandler,
		"attributes":     recurseHandler,
		"key":            keyHandler,
		"divisions":      divisionsHandler,
		"transpose":      transposeHandler,
		"note":           noteHandler,
		"forward":        forwardHandler,
		"backup":         backupHandler,
	}
}

// handle imports an element by finding and calling its handler
func handle(element *etree.Element, importer *musicXMLImporter) {
	if handler, ok := handlers[element.Tag]; ok {
		handler(element, importer)
	}
}

// unsupportedHandler logs an error to the user for category (2) elements
func unsupportedHandler(elementName string, isPlanned bool) elementHandler {
	return func(element *etree.Element, importer *musicXMLImporter) {
		// Only mention that a tag is unsupported once
		for _, value := range importer.unsupported {
			if value == element.Tag {
				return
			}
		}

		importer.unsupported = append(
			[]string{element.Tag}, importer.unsupported...,
		)
		if isPlanned {
			log.Warn().Msg(fmt.Sprintf(
				`%s with the <%s> tag are currently not supported for MusicXML import.
We plan to add support for importing %s in the future.`,
				aurora.BrightYellow(elementName),
				aurora.BrightYellow(element.Tag),
				aurora.BrightYellow(elementName),
			))
		} else {
			log.Warn().Msg(fmt.Sprintf(
				`%s with the <%s> tag are not supported for MusicXML import.`,
				aurora.BrightYellow(elementName),
				aurora.BrightYellow(element.Tag),
			))
		}
	}
}

// recurseHandler recursively calls handle for all direct children
// This method represents tags we explicitly search instead of skipping
func recurseHandler(element *etree.Element, importer *musicXMLImporter) {
	for _, child := range element.ChildElements() {
		handle(child, importer)
	}
}

func scorePartHandler(element *etree.Element, importer *musicXMLImporter) {
	// Create the part
	id := element.SelectAttr("id")

	var instruments []string

	midiInstruments := element.FindElements("midi-instrument")
	if len(midiInstruments) == 0 {
		// If no instruments are listed, we default to a piano
		instruments = []string{"piano"}
	} else {
		// If instruments are listed, we find each MIDI ID and obtain their name
		var instrumentIDs []int64
		for _, midiInstrument := range midiInstruments {
			instrumentID, _ := strconv.ParseInt(
				midiInstrument.FindElement("midi-program").Text(), 10, 8,
			)
			instrumentIDs = append(instrumentIDs, instrumentID)
		}

		instrumentsList := model.InstrumentsList()
		for _, instrumentID := range instrumentIDs {
			instruments = append(instruments, instrumentsList[instrumentID - 1])
		}
	}

	part := musicXMLPart{
		voices:      make(map[int32]*musicXMLVoice),
		instruments: instruments,
		divisions:   1, // Divisions for a part are set in the first measure
	}
	importer.parts[id.Value] = &part
}

func addBarline(importer *musicXMLImporter) {
	if len(importer.updates()) == 0 {
		importer.set([]model.ScoreUpdate{model.Barline{}})
		return
	}
	// Alda parses barlines into note/rest durations to account for ties
	// across measures. We try to do the same thing when importing

	// tryAddBarline attempts to add a barline and returns success/failure
	var tryAddBarline func(update model.ScoreUpdate) (model.ScoreUpdate, bool)
	tryAddBarline = func(
		update model.ScoreUpdate,
	) (model.ScoreUpdate, bool) {
		switch value := update.(type) {
		case model.Note:
			// If a note is slurred, Alda does not parse the next barline
			// So we do not insert either
			if !value.Slurred {
				value.Duration.Components = append(
					value.Duration.Components, model.Barline{},
				)
				return value, true
			}
		case model.Rest:
			value.Duration.Components = append(
				value.Duration.Components, model.Barline{},
			)
			return value, true
		case model.Chord:
			// For chords, we add the barline by recursing on the last event
			if len(value.Events) > 0 {
				if last, success := tryAddBarline(
					value.Events[len(value.Events) - 1],
				); success {
					value.Events[len(value.Events) - 1] = last
					return value, true
				}
			}
		}
		return update, false
	}

	if newUpdate, success := tryAddBarline(
		importer.updates()[len(importer.updates()) - 1],
	); success {
		importer.updates()[len(importer.updates()) - 1] = newUpdate
	} else {
		importer.append(model.Barline{})
	}
}

func partHandler(element *etree.Element, importer *musicXMLImporter) {
	// Set current part in importer
	id := element.SelectAttr("id")
	part := importer.parts[id.Value]
	importer.currentPart = part

	// Do a pass through all content and create all voices
	// This is so we can fill these voices with rests and barlines as we import
	var recursivelyCreateVoices func(element *etree.Element)
	recursivelyCreateVoices = func(element *etree.Element) {
		if element.Tag == "voice" {
			voiceNumber, _ := strconv.ParseInt(element.Text(), 10, 32)
			voice := musicXMLVoice{
				octave: 4, // 4 is the default Alda octave
				slurs:  0, // By default a note is not slurred
			}
			importer.currentPart.voices[int32(voiceNumber)] = &voice
		}

		for _, child := range element.ChildElements() {
			recursivelyCreateVoices(child)
		}
	}

	recursivelyCreateVoices(element)
	importer.currentPart.currentVoice = importer.currentPart.voices[1]

	for i, child := range element.ChildElements() {
		handle(child, importer)
		// Between each measure, we do cleanup
		for _, voice := range importer.currentPart.voices {
			// We pad each voice with rests
			// This is because we can never backup to before a measure
			importer.currentPart.currentVoice = voice
			padVoiceToPresent(child, importer)

			// We add barlines to create more idiomatic Alda
			if i < len(element.ChildElements()) - 1 {
				addBarline(importer)
			}
		}
	}
}

// padVoiceToPresent fills the current voice with rests to catch up to the
// current beats counter
func padVoiceToPresent(element *etree.Element, importer *musicXMLImporter) {
	beatDifference := importer.currentPart.beats -
		importer.currentPart.currentVoice.beats

	if beatDifference < 0 {
		// We make an assumption that this cannot happen
		// If we must handle this case, importing would have to support changes
		// by beats / duration. This would be non-trivial
		warnWhileParsing(element, `voice is behind in beats, output will be incorrect.`)
	} else if beatDifference > 0 {
		// We update without using append and moving the part-level beats here
		// This is because we pad the voice to "catch up", not move forwards
		importer.currentPart.currentVoice.updates = append(
			importer.currentPart.currentVoice.updates,
			model.Rest{
				Duration: model.Duration{
					Components: []model.DurationComponent{
						model.NoteLength{
							Denominator: 4 / beatDifference,
						},
					},
				},
			},
		)
		importer.currentPart.currentVoice.beats += beatDifference
	}
}

func measureHandler(element *etree.Element, importer *musicXMLImporter) {
	for _, child := range element.ChildElements() {
		// We handle voice switching at the measure level
		if voice := child.FindElement("voice"); voice != nil {
			voiceNumber, _ := strconv.ParseInt(voice.Text(), 10, 8)
			importer.currentPart.currentVoice =
				importer.currentPart.voices[int32(voiceNumber)]

			// We pad the voice with rests as necessary to align beats
			padVoiceToPresent(child, importer)
		}
		handle(child, importer)
	}
}

func backupHandler(element *etree.Element, importer *musicXMLImporter) {
	// We can handle backup by moving the part-level beats counter back
	// We assume that we cannot backup to before an existing note in a voice
	// This is not enforced by MusicXML, but exporters should have no reason to
	// do this
	duration, _ := translateDuration(element, importer)
	importer.currentPart.beats -= getBeats(model.Rest{Duration: duration})
}

func forwardHandler(element *etree.Element, importer *musicXMLImporter) {
	// We can handle forward by moving the part-level beats counter forward
	// This is because when switching voices, we pad to present
	duration, _ := translateDuration(element, importer)
	importer.currentPart.beats += getBeats(model.Rest{Duration: duration})
}

func keyHandler(element *etree.Element, importer *musicXMLImporter) {
	if element.FindElement("key-step") != nil ||
		element.FindElement("key-alter") != nil ||
		element.FindElement("key-accidental") != nil {
		// non-traditional-key
		unsupportedHandler(
			"non-traditional key signatures", false,
		)(element, importer)
	} else {
		// traditional-key
		fifths, _ := strconv.ParseInt(
			element.FindElement("fifths").Text(), 10, 8,
		)
		keySignature := model.KeySignatureFromCircleOfFifths(int(fifths))
		keySignatureSet := model.AttributeUpdate{
			PartUpdate: model.KeySignatureSet{KeySignature: keySignature},
		}
		importer.append(keySignatureSet)
	}
	// key-octave tags are purely for appearance, and are thus category (3)
}

func divisionsHandler(element *etree.Element, importer *musicXMLImporter) {
	// Due to our management of Alda durations and beats, we do not need to
	// maintain history for division changes
	value, _ := strconv.ParseInt(element.Text(), 10, 8)
	importer.currentPart.divisions = int(value)
}

func transposeHandler(element *etree.Element, importer *musicXMLImporter) {
	semitones, _ := strconv.ParseInt(
		element.FindElement("chromatic").Text(), 10, 8,
	)

	transposeSet := model.AttributeUpdate{
		PartUpdate: model.TranspositionSet{Semitones: int32(semitones)},
	}

	importer.append(transposeSet)
}

// translateDuration translates a MusicXML element with a duration to an Alda
// duration using the tracked divisions parameter
func translateDuration(
	element *etree.Element, importer *musicXMLImporter,
) (model.Duration, float64) {
	dots := int32(len(element.FindElements("dot")))

	// We retrieve the real duration without dots using some math
	// For each dot, it adds "half" the value of the original duration
	// 0 dot - total duration = 1x
	// 1 dot - total duration = 1.5x
	// 2 dot - total duration = 1.75x
	// This can be represented by sum from i = 0 to i = dots of (1/2)^dots
	// This is a geometric series with sum (1 - r^(n + 1)) / (1 - r), r = 1/2
	duration, _ := strconv.ParseFloat(element.FindElement("duration").Text(), 8)
	if dots > 0 {
		divisor := (1 - math.Pow(0.5, float64(dots + 1))) / 0.5
		duration = duration / divisor
	}

	aldaDuration := 4 * float64(importer.currentPart.divisions) / duration

	return model.Duration{
		Components: []model.DurationComponent{model.NoteLength{
			Denominator: aldaDuration,
			Dots:        dots,
		}},
	}, aldaDuration
}

// getBeats counts beats for score updates
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
		}
	}
	return beats
}

// translateNote translates a MusicXML note into Alda represented by:
// - A list of score updates ([]model.ScoreUpdate)
// - Octave (int64)
// - Duration (float64)
// - Change in slur count (int64)
// translateNote does not deal with chords or ties (handled in noteHandler)
// translateNote does not directly modify importer state, but state must be
// updated if the note were included directly (see noteHandler)
func translateNote(
	element *etree.Element, importer *musicXMLImporter,
) ([]model.ScoreUpdate, int64, float64, int64) {
	noteType := element.FindElement("type")

	if cue := element.FindElement("cue"); cue != nil ||
		noteType != nil && noteType.SelectAttrValue("type", "full") == "cue" {
		// Cue notes are category (3) as they do not change music sound directly
		// That being said, we cannot ignore cue notes as they have duration
		// So we can handle cue notes by using a rest with the same duration
		unsupportedHandler("cue notes", true)(cue, importer)
		return nil, 0, 0, 0
	} else if grace := element.FindElement("grace"); grace != nil ||
		noteType != nil && noteType.SelectAttrValue("type", "full") == "grace" {
		// Grace notes are category (2), as they can technically be imported
		// into Alda by "stealing time"
		// This is unnecessarily difficult, and is not planned to be handled
		unsupportedHandler("grace notes", false)(grace, importer)
		return nil, 0, 0, 0
	}

	// We should be dealing with a full note now, with duration
	duration, durationValue := translateDuration(element, importer)

	// A full note can be pitched, unpitched, or a rest
	if pitch := element.FindElement("pitch"); pitch != nil {
		// Note letter
		step := pitch.FindElement("step")

		noteLetter, _ := model.NewNoteLetter(
			[]rune(strings.ToLower(step.Text()))[0],
		)

		// Slurs
		var slurChange int64 = 0
		for _, slur := range element.FindElements("notations/slur") {
			switch slurType := slur.SelectAttrValue("type", "none"); slurType {
			case "start":
				slurChange += 1
			case "stop":
				slurChange -= 1
			}
		}

		slurred := (importer.currentPart.currentVoice.slurs + slurChange) > 0

		// Accidentals
		var accidentals []model.Accidental
		if alter := pitch.FindElement("alter"); alter != nil {
			alterAmount, _ := strconv.ParseInt(alter.Text(), 10, 8)

			var accidental model.Accidental
			if alterAmount > 0 {
				accidental = model.Sharp
			} else if alterAmount < 0 {
				accidental = model.Flat
			}

			for i := 0; i < int(math.Abs(float64(alterAmount))); i++ {
				accidentals = append(accidentals, accidental)
			}
		}

		// Octaves
		octave := pitch.FindElement("octave")
		octaveVal, _ := strconv.ParseInt(octave.Text(), 10, 8)
		octaveDifference := octaveVal - importer.currentPart.currentVoice.octave

		var octaveUpdate model.PartUpdate
		if octaveDifference > 0 {
			octaveUpdate = model.OctaveUp{}
		} else if octaveDifference < 0 {
			octaveUpdate = model.OctaveDown{}
		}

		var octaveUpdates []model.ScoreUpdate
		for i := 0; i < int(math.Abs(float64(octaveDifference))); i++ {
			octaveUpdates = append(
				octaveUpdates, model.AttributeUpdate{
					PartUpdate: octaveUpdate,
				},
			)
		}

		note := model.Note{
			Pitch: model.LetterAndAccidentals{
				NoteLetter:  noteLetter,
				Accidentals: accidentals,
			},
			Duration: duration,
			Slurred:  slurred,
		}

		return append(octaveUpdates, note), octaveVal, durationValue, slurChange
	} else if rest := element.FindElement("rest"); rest != nil {
		restScoreUpdate := model.Rest{
			Duration: duration,
		}

		return []model.ScoreUpdate{restScoreUpdate},
			importer.currentPart.currentVoice.octave,
			durationValue,
			0
	} else {
		unsupportedHandler("unpitched notes", true)(element, importer)
		return nil, 0, 0, 0
	}
}

func noteHandler(element *etree.Element, importer *musicXMLImporter) {
	noteUpdates, newOctave, _, slurChange := translateNote(
		element, importer,
	)
	if len(noteUpdates) == 0 {
		return
	}

	updateImporterState := func() {
		importer.currentPart.currentVoice.octave = newOctave
		importer.currentPart.currentVoice.slurs += slurChange
	}

	// We obtain part-level attributes that must be handled separately
	isChord := element.FindElement("chord") != nil

	isTieStop := false
	for _, tie := range element.FindElements("tie") {
		if tie.SelectAttrValue("type", "") == "stop" {
			isTieStop = true
		}
	}
	if reflect.TypeOf(noteUpdates[len(noteUpdates) - 1]) != noteType {
		// rests cannot be tied
		isTieStop = false
	}

	if isTieStop {
		// In MusicXML, a tied note is represented by tie start & stop tags
		// A note with a tie stop tag is tied to the previous note with the
		// same pitch. So we start by finding this previous note
		note := noteUpdates[len(noteUpdates) - 1].(model.Note)
		notePitch := note.Pitch.(model.LetterAndAccidentals)

		getOctaveChange := func(update model.ScoreUpdate) int64 {
			if reflect.TypeOf(update) == attributeUpdateType {
				partUpdate := update.(model.AttributeUpdate).PartUpdate
				if reflect.TypeOf(partUpdate) == octaveUpType {
					return -1
				} else if reflect.TypeOf(partUpdate) == octaveDownType {
					return 1
				} else {
					return 0
				}
			} else {
				return 0
			}
		}

		ni, _ := lastWithState(
			importer.updates(),
			func(update model.ScoreUpdate, state interface{}) bool {
				switch value := update.(type) {
				case model.Note:
					pitch := value.Pitch.(model.LetterAndAccidentals)
					octave := state.(int64)
					return octave == newOctave &&
						deep.Equal(pitch, notePitch) == nil
				default:
					return false
				}
			},
			importer.currentPart.currentVoice.octave,
			func(update model.ScoreUpdate, state interface{}) interface{} {
				return state.(int64) + getOctaveChange(update)
			},
		)

		if ni.index == -1 {
			// We ignore a note if we can't find the start of the tie
			warnWhileParsing(element, `cannot find start of tie`)
			return
		}

		// A note can be tied to a previous note even if there is a gap
		// between them. In this case, the second note is not logical
		// so we will ignore it when importing to Alda
		// To detect this gap, we use the fact that MusicXML restricts notes
		// in a chord to all have the same duration
		// Then we can find the last update with a duration, and compare its
		// index with our tie start index
		lastIndex := last(importer.updates(), filterUpdateWithDuration).index
		indexDiff := lastIndex - ni.index

		var isAdjacentTie bool
		if isChord {
			// If this note is part of a chord, we can allow one element
			// This element represents the start of the chord
			if indexDiff <= 0 {
				isAdjacentTie = true
			} else {
				// We can allow a single element, no more
				lastLastIndex := last(
					importer.updates()[:lastIndex], filterUpdateWithDuration,
				).index

				if lastLastIndex == -1 {
					isAdjacentTie = true
				} else {
					indexDiff = lastLastIndex - ni.index
					isAdjacentTie = indexDiff <= 0
				}
			}
		} else {
			// If a note is not part of a chord, we can allow no elements
			isAdjacentTie = indexDiff <= 0
		}

		if isAdjacentTie {
			// We have passed all checks, so we add the duration
			importer.set(apply(
				importer.updates(),
				ni,
				func(update model.ScoreUpdate) model.ScoreUpdate {
					tieStart := update.(model.Note)
					tieStart.Duration.Components = append(
						tieStart.Duration.Components,
						note.Duration.Components...,
					)
					return tieStart
				},
			))
		} else {
			// We ignore a note that is in a nonadjacent tie
			warnWhileParsing(element, `found nonadjacent tied note`)
			return
		}
	} else if isChord {
		// We find the last score update that is a note, rest, or chord
		lastIndex := last(importer.updates(), filterUpdateWithDuration).index

		if lastIndex == -1 {
			// We continue by considering the note as not part of a chord
			warnWhileParsing(
				element, `note found in chord with no starting note. The note will be ignored.`,
			)
			return
		}

		switch value := importer.updates()[lastIndex].(type) {
		case model.Chord:
			// Add the note updates to the existing chord
			value.Events = append(value.Events, noteUpdates...)
			updateImporterState()
			importer.updates()[lastIndex] = value
		case model.Note, model.Rest:
			// Create a chord from the previous updates and new note updates
			chord := model.Chord{Events: make(
				[]model.ScoreUpdate,
				len(importer.updates()) - lastIndex,
			)}
			copy(chord.Events, importer.updates()[lastIndex:])
			chord.Events = append(chord.Events, noteUpdates...)

			updateImporterState()
			importer.set(append(importer.updates()[:lastIndex], chord))
		}
	} else {
		updateImporterState()
		importer.append(noteUpdates...)
	}
}
