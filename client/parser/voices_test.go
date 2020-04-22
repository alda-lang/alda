package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
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
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
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
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.VoiceMarker{VoiceNumber: 2},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
			},
		},
		parseTestCase{
			label: "part with two voices separated by a barline",
			given: `piano:
			V1: a b c | V2: d e f`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Barline{},
				model.VoiceMarker{VoiceNumber: 2},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
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
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.A}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.B}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					),
					8,
				),
				model.VoiceMarker{VoiceNumber: 2},
				repeat(
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
					),
					8,
				),
			},
		},
		parseTestCase{
			label: "voice group end marker",
			given: `piano:
			V1: c
			V2: e
			V0: g`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.VoiceMarker{VoiceNumber: 1},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.VoiceMarker{VoiceNumber: 2},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				model.VoiceGroupEndMarker{},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.G}},
			},
		},
	)
}
