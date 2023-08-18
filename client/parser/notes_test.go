package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestNotes(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "note with implicit duration",
			given: "c",
			expectUpdates: []model.ScoreUpdate{
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
			},
		},
		parseTestCase{
			label: "note with explicit duration",
			given: "c4",
			expectUpdates: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 4},
						},
					},
				},
			},
		},
		parseTestCase{
			label: "sharp note",
			given: "c+",
			expectUpdates: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{
						NoteLetter:  model.C,
						Accidentals: []model.Accidental{model.Sharp},
					},
				},
			},
		},
		parseTestCase{
			label: "flat note",
			given: "b-",
			expectUpdates: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{
						NoteLetter:  model.B,
						Accidentals: []model.Accidental{model.Flat},
					},
				},
			},
		},
		parseTestCase{
			label: "double sharp note",
			given: "c++",
			expectUpdates: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{
						NoteLetter:  model.C,
						Accidentals: []model.Accidental{model.Sharp, model.Sharp},
					},
				},
			},
		},
		parseTestCase{
			label: "double flat note",
			given: "b--",
			expectUpdates: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{
						NoteLetter:  model.B,
						Accidentals: []model.Accidental{model.Flat, model.Flat},
					},
				},
			},
		},
		parseTestCase{
			label: "rest with implicit duration",
			given: "r",
			expectUpdates: []model.ScoreUpdate{
				model.Rest{},
			},
		},
		parseTestCase{
			label: "rest with explicit duration",
			given: "r1",
			expectUpdates: []model.ScoreUpdate{
				model.Rest{
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 1},
						},
					},
				},
			},
		},
	)
}
