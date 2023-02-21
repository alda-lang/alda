package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestBarlines(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "simple use of barlines",
			given: "violin: c d | e f | g a",
			expectUpdates: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"violin"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Barline{},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
				model.Barline{},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.G}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
			},
		},
		parseTestCase{
			label: "a note tied over many barlines",
			given: "marimba: c1|~1|~1~|1|~1~|2.",
			expectUpdates: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"marimba"}},
				model.Note{
					Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
		},
	)

}
