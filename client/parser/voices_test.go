package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func TestVoices(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "part with voice",
			given: "piano: V1: a b c",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				model.Note{NoteLetter: model.A},
				model.Note{NoteLetter: model.B},
				model.Note{NoteLetter: model.C},
			},
		},
		parseTestCase{
			label: "part with two voices",
			given: `piano:
			V1: a b c
			V2: d e f`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				model.Note{NoteLetter: model.A},
				model.Note{NoteLetter: model.B},
				model.Note{NoteLetter: model.C},
				model.VoiceMarker{VoiceNumber: 2},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
				model.Note{NoteLetter: model.F},
			},
		},
		parseTestCase{
			label: "part with two voices separated by a barline",
			given: `piano:
			V1: a b c | V2: d e f`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				model.Note{NoteLetter: model.A},
				model.Note{NoteLetter: model.B},
				model.Note{NoteLetter: model.C},
				model.Barline{},
				model.VoiceMarker{VoiceNumber: 2},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
				model.Note{NoteLetter: model.F},
			},
		},
		parseTestCase{
			label: "part with two slightly more complex voices",
			given: `piano:
			V1: [a b c] *8
			V2: [d e f] *8`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.A},
						model.Note{NoteLetter: model.B},
						model.Note{NoteLetter: model.C},
					),
					8,
				),
				model.VoiceMarker{VoiceNumber: 2},
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.D},
						model.Note{NoteLetter: model.E},
						model.Note{NoteLetter: model.F},
					),
					8,
				),
			},
		},
	)
}
