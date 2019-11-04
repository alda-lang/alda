package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestComments(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "simple comment",
			given: `piano: c
			# d
			e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "comment at the end of a line",
			given: `piano: c # d
			e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "comment without a leading space",
			given: `piano: c #d
			e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
	)
}
