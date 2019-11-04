package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestCRAM(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "CRAM expression",
			given: "{c d e}",
			expect: []model.ScoreUpdate{
				model.Cram{
					Events: []model.ScoreUpdate{
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
					},
				},
			},
		},
		parseTestCase{
			label: "CRAM expression with specified duration",
			given: "{c d e}2",
			expect: []model.ScoreUpdate{
				model.Cram{
					Events: []model.ScoreUpdate{
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
					},
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 2},
						},
					},
				},
			},
		},
	)
}
