package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func cNoteWithDuration(components ...model.DurationComponent) model.Note {
	return model.Note{
		Pitch:    model.LetterAndAccidentals{NoteLetter: model.C},
		Duration: model.Duration{Components: components},
	}
}

func TestDurations(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "Note with integer note length",
			given: "c2",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLength{Denominator: 2}),
			},
		},
		parseTestCase{
			label: "Note with fractional note length",
			given: "c0.5",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLength{Denominator: 0.5}),
			},
		},
		parseTestCase{
			label: "Note with integer note length and dots",
			given: "c2..",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLength{Denominator: 2, Dots: 2}),
			},
		},
		parseTestCase{
			label: "Note with fractional note length and dots",
			given: "c0.5..",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLength{Denominator: 0.5, Dots: 2}),
			},
		},
		parseTestCase{
			label: "Note with duration in milliseconds",
			given: "c450ms",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLengthMs{Quantity: 450}),
			},
		},
		parseTestCase{
			label: "Note with duration in seconds",
			given: "c2s",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(model.NoteLengthMs{Quantity: 2000}),
			},
		},
		parseTestCase{
			label: "Note with tied integer note lengths",
			given: "c1~2~4",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(
					model.NoteLength{Denominator: 1},
					model.NoteLength{Denominator: 2},
					model.NoteLength{Denominator: 4},
				),
			},
		},
		parseTestCase{
			label: "Note with tied integer/fractional note lengths",
			given: "c1.5~2.5~4",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(
					model.NoteLength{Denominator: 1.5},
					model.NoteLength{Denominator: 2.5},
					model.NoteLength{Denominator: 4},
				),
			},
		},
		parseTestCase{
			label: "Note with tied millisecond note lengths",
			given: "c500ms~350ms",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(
					model.NoteLengthMs{Quantity: 500},
					model.NoteLengthMs{Quantity: 350},
				),
			},
		},
		parseTestCase{
			label: "Note with various tied note lengths",
			given: "c5s~4~350ms~0.5",
			expect: []model.ScoreUpdate{
				cNoteWithDuration(
					model.NoteLengthMs{Quantity: 5000},
					model.NoteLength{Denominator: 4},
					model.NoteLengthMs{Quantity: 350},
					model.NoteLength{Denominator: 0.5},
				),
			},
		},
		parseTestCase{
			label: "Slurred note with implicit duration",
			given: "c~",
			expect: []model.ScoreUpdate{
				model.Note{
					Pitch:   model.LetterAndAccidentals{NoteLetter: model.C},
					Slurred: true,
				},
			},
		},
		parseTestCase{
			label: "Slurred note with integer note length",
			given: "c4~",
			expect: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 4},
						},
					},
					Slurred: true,
				},
			},
		},
		parseTestCase{
			label: "Slurred note with millisecond note length",
			given: "c420ms~",
			expect: []model.ScoreUpdate{
				model.Note{
					Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLengthMs{Quantity: 420},
						},
					},
					Slurred: true,
				},
			},
		},
	)
}
