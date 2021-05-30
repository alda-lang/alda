package importer

import (
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"github.com/beevik/etree"
	"github.com/logrusorgru/aurora"
	"io"
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
