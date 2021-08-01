package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func eventSequence(events ...model.ScoreUpdate) model.EventSequence {
	if events == nil {
		events = []model.ScoreUpdate{}
	}
	return model.EventSequence{Events: events}
}

func TestEventSequences(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "empty event sequence",
			given: "[]",
			expect: []model.ScoreUpdate{
				eventSequence(),
			},
		},
		parseTestCase{
			label: "empty event sequence with internal whitespace",
			given: "[    ]",
			expect: []model.ScoreUpdate{
				eventSequence(),
			},
		},
		parseTestCase{
			label: "event sequence with some notes and rests",
			given: "[c d c r]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Rest{},
				),
			},
		},
		parseTestCase{
			label: "event sequence with some notes, rests and a little right padding",
			given: "[c d c r ]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Rest{},
				),
			},
		},
		parseTestCase{
			label: "event sequence with some notes and a chord",
			given: "[ c d e f c/e/g ]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
					model.Chord{
						Events: []model.ScoreUpdate{
							model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
							model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
							model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.G}},
						},
					},
				),
			},
		},
		parseTestCase{
			label: "nested event sequence with some notes",
			given: "[c d [e f] g]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
					),
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.G}},
				),
			},
		},
		parseTestCase{
			label: "event sequence containing voices",
			given: "[V1: e b d V2: a c f]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.VoiceMarker{VoiceNumber: 1},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.VoiceMarker{VoiceNumber: 2},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
				),
			},
		},
		parseTestCase{
			label: "event sequence containing a note",
			given: "[c1]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 1},
							},
						},
					},
				),
			},
		},
		parseTestCase{
			label: "event sequence containing a note w/ duration in seconds",
			given: "[c1s]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLengthMs{Quantity: 1000},
							},
						},
					},
				),
			},
		},
	)
}
