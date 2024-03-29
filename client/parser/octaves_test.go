package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func octaveUp() model.AttributeUpdate {
	return model.AttributeUpdate{PartUpdate: model.OctaveUp{}}
}

func octaveDown() model.AttributeUpdate {
	return model.AttributeUpdate{PartUpdate: model.OctaveDown{}}
}

func octaveSet(octaveNumber int32) model.AttributeUpdate {
	return model.AttributeUpdate{
		PartUpdate: model.OctaveSet{OctaveNumber: octaveNumber},
	}
}

func TestOctaves(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label:         "octave up",
			given:         ">",
			expectUpdates: []model.ScoreUpdate{octaveUp()},
		},
		parseTestCase{
			label:         "octave down",
			given:         "<",
			expectUpdates: []model.ScoreUpdate{octaveDown()},
		},
		parseTestCase{
			label:         "octave set",
			given:         "o5",
			expectUpdates: []model.ScoreUpdate{octaveSet(5)},
		},
		parseTestCase{
			label: "multi-octave up",
			given: ">>>",
			expectUpdates: []model.ScoreUpdate{
				octaveUp(),
				octaveUp(),
				octaveUp(),
			},
		},
		parseTestCase{
			label: "multi-octave down",
			given: "<<<",
			expectUpdates: []model.ScoreUpdate{
				octaveDown(),
				octaveDown(),
				octaveDown(),
			},
		},
		parseTestCase{
			label: "octave fish ><>",
			given: "><>",
			expectUpdates: []model.ScoreUpdate{
				octaveUp(),
				octaveDown(),
				octaveUp(),
			},
		},
		parseTestCase{
			label: "octave up immediately followed by note",
			given: ">c",
			expectUpdates: []model.ScoreUpdate{
				octaveUp(),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
			},
		},
		parseTestCase{
			label: "octave down immediately followed by note",
			given: "<c",
			expectUpdates: []model.ScoreUpdate{
				octaveDown(),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
			},
		},
		parseTestCase{
			label: "note immediately followed by octave up",
			given: "c>",
			expectUpdates: []model.ScoreUpdate{
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				octaveUp(),
			},
		},
		parseTestCase{
			label: "note immediately followed by octave down",
			given: "c<",
			expectUpdates: []model.ScoreUpdate{
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				octaveDown(),
			},
		},
		parseTestCase{
			label: "note sandwiched between octave up/down",
			given: ">c<",
			expectUpdates: []model.ScoreUpdate{
				octaveUp(),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				octaveDown(),
			},
		},
	)
}
