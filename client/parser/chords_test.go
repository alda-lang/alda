package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestChords(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "simple chord",
			given: "c/e/g",
			expect: []model.ScoreUpdate{
				model.Chord{
					Events: []model.ScoreUpdate{
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.G}},
					},
				},
			},
		},
		parseTestCase{
			label: "chord that includes a rest",
			given: "c1/>e2/g4/r8",
			expect: []model.ScoreUpdate{
				model.Chord{
					Events: []model.ScoreUpdate{
						model.Note{
							Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 1},
								},
							},
						},
						model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
						model.Note{
							Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 2},
								},
							},
						},
						model.Note{
							Pitch: model.LetterAndAccidentals{NoteLetter: model.G},
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 4},
								},
							},
						},
						model.Rest{
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 8},
								},
							},
						},
					},
				},
			},
		},
		parseTestCase{
			label: "chord that includes a dotted note",
			given: "b>/d/f2.",
			expect: []model.ScoreUpdate{
				model.Chord{
					Events: []model.ScoreUpdate{
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
						model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
						model.Note{
							Pitch: model.LetterAndAccidentals{NoteLetter: model.F},
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 2, Dots: 1},
								},
							},
						},
					},
				},
			},
		},
	)
}
