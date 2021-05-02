package interop

import (
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"fmt"
	"github.com/beevik/etree"
	"github.com/logrusorgru/aurora"
	"io"
	"math"
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
type elementHandler = func(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate

// handlers stores the elementHandler functions for most imported elements
// Some elements are not stored here, and are used directly in the handling of
// an ancestor tag (e.x. non-traditional-key, duration, pitch)
var handlers map[string]elementHandler

func init() {
	handlers = map[string]elementHandler{
		// Category (1)
		"score-partwise": aggregateHandler,
		"part-list":      aggregateHandler,
		"score-part":     scorePartHandler,
		"part":           partHandler,
		"measure":        aggregateHandler,
		"attributes":     aggregateHandler,
		"key":            keyHandler,
		"divisions":      divisionsHandler,
		"note":           noteHandler,

		// Category (2)
		// We can handle cue notes by just using a rest with the same duration
		"cue": unsupportedHandler("cue notes", true),
		// Handling grace notes by stealing time may be unnecessarily difficult
		"grace": unsupportedHandler("grace notes", false),

		// Category (3) and uncategorized elements are automatically skipped
	}
}

// handle imports an element by finding and calling its handler
func handle(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	if handler, ok := handlers[element.Tag]; ok {
		return handler(element, importer)
	} else {
		return skipHandler(element, importer)
	}
}

// musicXMLImporter contains global state for importing a MusicXML file
type musicXMLImporter struct {
	// parts maps part ID to part information
	parts map[string]musicXMLPart

	// currentPart in the score that is being processed
	currentPart musicXMLPart

	// unsupported stores the tags of unsupported elements that have been
	// encountered, so we don't warn the user multiple times for each tag type
	unsupported []string
}

// musicXMLPart contains part-specific information for importing MusicXML
type musicXMLPart struct {
	instrument string
	divisions  int
	octave     int
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
			`Issue importing MusicXML file: could not find %s root tag`,
			aurora.BrightYellow("score-partwise"),
		)
	}

	importer := musicXMLImporter{
		parts: make(map[string]musicXMLPart),
	}

	return handle(scorePartwise, &importer), nil
}

// unsupportedHandler logs an error to the user for category (2) elements
func unsupportedHandler(
	elementDisplayName string, isPlanned bool,
) elementHandler {
	return func(
		element *etree.Element, importer *musicXMLImporter,
	) []model.ScoreUpdate {
		// Only mention that a tag is unsupported once
		for _, value := range importer.unsupported {
			if value == element.Tag {
				return skipHandler(element, importer)
			}
		}

		importer.unsupported = append(
			[]string{element.Tag}, importer.unsupported...,
		)
		if isPlanned {
			log.Warn().Msg(fmt.Sprintf(
				`%s with the <%s> tag are currently not supported for MusicXML import.

We plan to add support for importing %s in the future.`,
				aurora.BrightYellow(elementDisplayName),
				aurora.BrightYellow(element.Tag),
				aurora.BrightYellow(elementDisplayName),
			))
		} else {
			log.Warn().Msg(fmt.Sprintf(
				`%s with the <%s> tag are not supported for MusicXML import.`,
				aurora.BrightYellow(elementDisplayName),
				aurora.BrightYellow(element.Tag),
			))
		}
		return skipHandler(element, importer)
	}
}

// skipHandler skips the entire element and returns an empty slice
func skipHandler(_ *etree.Element, _ *musicXMLImporter, ) []model.ScoreUpdate {
	return []model.ScoreUpdate{}
}

// aggregateHandler merges recursive calls to handle for all direct children
func aggregateHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	var scoreUpdates []model.ScoreUpdate
	for _, child := range element.ChildElements() {
		scoreUpdates = append(scoreUpdates, handle(child, importer)...)
	}
	return scoreUpdates
}

func scorePartHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	// Create the part
	id := element.SelectAttr("id")

	instrumentName := element.FindElement("score-instrument/instrument-name")

	importer.parts[id.Value] = musicXMLPart{
		instrument: strings.ToLower(instrumentName.Text()),
		divisions:  1, // Divisions for a part are set in the first measure
		octave:     4,
	}

	return skipHandler(element, importer)
}

func partHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	// Set current part in importer
	id := element.SelectAttr("id")
	importer.currentPart = importer.parts[id.Value]

	partDeclaration := model.PartDeclaration{
		Names: []string{importer.currentPart.instrument},
	}

	// TODO: add barlines between each measure?
	return append(
		[]model.ScoreUpdate{partDeclaration},
		aggregateHandler(element, importer)...,
	)
}

func keyHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	if element.FindElement("key-step") != nil ||
		element.FindElement("key-alter") != nil ||
		element.FindElement("key-accidental") != nil {
		// non-traditional-key
		return unsupportedHandler(
			"non-traditional key signatures", false,
		)(element, importer)
	} else {
		// traditional-key
		fifths, _ := strconv.ParseInt(
			element.FindElement("fifths").Text(), 10, 8,
		)
		keySignature := model.KeySignatureFromCircleOfFifths(int(fifths))
		scoreUpdate := model.AttributeUpdate{
			PartUpdate: model.KeySignatureSet{KeySignature: keySignature},
		}
		return []model.ScoreUpdate{scoreUpdate}
	}
	// key-octave tags are purely for appearance, and are thus category (3)
}

func divisionsHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	value, _ := strconv.ParseInt(element.Text(), 10, 8)
	importer.currentPart.divisions = int(value)
	return skipHandler(element, importer)
}

func noteHandler(
	element *etree.Element, importer *musicXMLImporter,
) []model.ScoreUpdate {
	noteType := element.FindElement("type")

	if cue := element.FindElement("cue"); cue != nil ||
		noteType != nil && noteType.SelectAttrValue("type", "full") == "cue" {
		// Cue notes are category (3) as they do not change music sound directly
		// That being said, we cannot ignore cue notes as they have duration
		return handle(cue, importer)
	}

	if grace := element.FindElement("grace"); grace != nil ||
		noteType != nil && noteType.SelectAttrValue("type", "full") == "grace" {
		// Grace notes are category (2), as they can technically be imported
		// into Alda by "stealing time"
		return handle(grace, importer)
	}

	// We should be dealing with a full note now
	durationInt, _ := strconv.ParseInt(
		element.FindElement("duration").Text(), 10, 8,
	)

	duration := TranslateDuration(
		importer.currentPart.divisions, int(durationInt),
	)

	// A full note can be pitched, unpitched, or a rest
	// It can also be a chord
	// for now, we only handle rests and pitched notes
	if pitch := element.FindElement("pitch"); pitch != nil {
		step := pitch.FindElement("step")

		noteLetter, _ := model.NewNoteLetter(
			[]rune(strings.ToLower(step.Text()))[0],
		)

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

		note := model.Note{
			Pitch: model.LetterAndAccidentals{
				NoteLetter:  noteLetter,
				Accidentals: accidentals,
			},
			Duration: duration,
		}

		// To manage octaves, we use OctaveUp/OctaveDown instead of OctaveSet
		octave := pitch.FindElement("octave")
		octaveInt, _ := strconv.ParseInt(octave.Text(), 10, 8)
		octaveDifference := octaveInt - int64(importer.currentPart.octave)
		importer.currentPart.octave = int(octaveInt)

		var octaveUpdate model.PartUpdate
		if octaveDifference > 0 {
			octaveUpdate = model.OctaveUp{}
		} else if octaveDifference < 0 {
			octaveUpdate = model.OctaveDown{}
		}

		var scoreUpdates []model.ScoreUpdate
		for i := 0; i < int(math.Abs(float64(octaveDifference))); i++ {
			scoreUpdates = append(
				scoreUpdates, model.AttributeUpdate{
					PartUpdate: octaveUpdate,
				},
			)
		}
		return append(scoreUpdates, note)
	} else if rest := element.FindElement("rest"); rest != nil {
		restScoreUpdate := model.Rest{
			Duration: duration,
		}
		return []model.ScoreUpdate{restScoreUpdate}
	} else {
		return unsupportedHandler("unpitched notes", true)(element, importer)
	}
}
