package code_generator

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"testing"
)

func eof() parser.Token {
	return parser.Token{TokenType: parser.EOF, Text: ""}
}

func TestGenerator(t *testing.T) {
	executeGeneratorTestCases(t, generatorTestCase{
		label: "attribute updates",
		updates: []model.ScoreUpdate{
			model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
			model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
			model.AttributeUpdate{PartUpdate: model.OctaveSet{OctaveNumber: 1}},
			model.AttributeUpdate{PartUpdate: model.OctaveSet{OctaveNumber: 3}},
		},
		expected: []parser.Token{
			{TokenType: parser.OctaveUp, Text: ">"},
			{TokenType: parser.OctaveDown, Text: "<"},
			{TokenType: parser.OctaveSet, Text: "o1"},
			{TokenType: parser.OctaveSet, Text: "o3"},
			eof(),
		},
	}, generatorTestCase{
		label: "barlines",
		updates: []model.ScoreUpdate{
			model.Barline{},
			model.Barline{},
		},
		expected: []parser.Token{
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.Barline, Text: "|"},
			eof(),
		},
	})
}
