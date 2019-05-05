package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func TestParts(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "part with single name",
			given: "theremin: c d e",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"theremin"}},
				model.Note{NoteLetter: model.C},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
			},
		},
		parseTestCase{
			label: "part with single name and a nickname",
			given: `harmonica "bob": c d e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"harmonica"}, Nickname: "bob"},
				model.Note{NoteLetter: model.C},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
			},
		},
		parseTestCase{
			label: "part with multiple names",
			given: "violin/viola: c d e",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"violin", "viola"}},
				model.Note{NoteLetter: model.C},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
			},
		},
		parseTestCase{
			label: "part with multiple names and a nickname",
			given: `trumpet/trombone/tuba "brass": c d e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{
					Names:    []string{"trumpet", "trombone", "tuba"},
					Nickname: "brass",
				},
				model.Note{NoteLetter: model.C},
				model.Note{NoteLetter: model.D},
				model.Note{NoteLetter: model.E},
			},
		},
		parseTestCase{
			label: "multiple parts",
			given: `guitar: e
			bass: e`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"guitar"}},
				model.Note{NoteLetter: model.E},
				model.PartDeclaration{Names: []string{"bass"}},
				model.Note{NoteLetter: model.E},
			},
		},
	)
}
