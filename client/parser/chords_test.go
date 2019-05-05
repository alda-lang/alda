package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
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
						model.Note{NoteLetter: model.C},
						model.Note{NoteLetter: model.E},
						model.Note{NoteLetter: model.G},
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
							NoteLetter: model.C,
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 1},
								},
							},
						},
						model.OctaveUp{},
						model.Note{
							NoteLetter: model.E,
							Duration: model.Duration{
								Components: []model.DurationComponent{
									model.NoteLength{Denominator: 2},
								},
							},
						},
						model.Note{
							NoteLetter: model.G,
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
						model.Note{NoteLetter: model.B},
						model.OctaveUp{},
						model.Note{NoteLetter: model.D},
						model.Note{
							NoteLetter: model.F,
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
