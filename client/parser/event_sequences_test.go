package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
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
					model.Note{NoteLetter: model.C},
					model.Note{NoteLetter: model.D},
					model.Note{NoteLetter: model.C},
					model.Rest{},
				),
			},
		},
		parseTestCase{
			label: "event sequence with some notes, rests and a little right padding",
			given: "[c d c r ]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{NoteLetter: model.C},
					model.Note{NoteLetter: model.D},
					model.Note{NoteLetter: model.C},
					model.Rest{},
				),
			},
		},
		parseTestCase{
			label: "event sequence with some notes and a chord",
			given: "[ c d e f c/e/g ]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.Note{NoteLetter: model.C},
					model.Note{NoteLetter: model.D},
					model.Note{NoteLetter: model.E},
					model.Note{NoteLetter: model.F},
					model.Chord{
						Events: []model.ScoreUpdate{
							model.Note{NoteLetter: model.C},
							model.Note{NoteLetter: model.E},
							model.Note{NoteLetter: model.G},
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
					model.Note{NoteLetter: model.C},
					model.Note{NoteLetter: model.D},
					eventSequence(
						model.Note{NoteLetter: model.E},
						model.Note{NoteLetter: model.F},
					),
					model.Note{NoteLetter: model.G},
				),
			},
		},
		parseTestCase{
			label: "event sequence containing voices",
			given: "[V1: e b d V2: a c f]",
			expect: []model.ScoreUpdate{
				eventSequence(
					model.VoiceMarker{VoiceNumber: 1},
					model.Note{NoteLetter: model.E},
					model.Note{NoteLetter: model.B},
					model.Note{NoteLetter: model.D},
					model.VoiceMarker{VoiceNumber: 2},
					model.Note{NoteLetter: model.A},
					model.Note{NoteLetter: model.C},
					model.Note{NoteLetter: model.F},
				),
			},
		},
	)
}
