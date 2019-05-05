package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
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
						model.Note{NoteLetter: model.C},
						model.Note{NoteLetter: model.D},
						model.Note{NoteLetter: model.E},
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
						model.Note{NoteLetter: model.C},
						model.Note{NoteLetter: model.D},
						model.Note{NoteLetter: model.E},
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
