package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func TestParts(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "part with single name",
			given: "piano: c d e",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "part with single name and an alias",
			given: `harmonica "bob": c d e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"harmonica"}, Alias: "bob"},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "part with multiple names",
			given: "violin/viola: c d e",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"violin", "viola"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "part with multiple names and an alias",
			given: `trumpet/trombone/tuba "brass": c d e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{
					Names: []string{"trumpet", "trombone", "tuba"},
					Alias: "brass",
				},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
		parseTestCase{
			label: "multiple parts",
			given: `guitar: e
			electric-bass: e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"guitar"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				model.PartDeclaration{Names: []string{"electric-bass"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
			},
		},
	)
}
