package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func TestBarlines(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "simple use of barlines",
			given: "violin: c d | e f | g a",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"violin"}},
				model.Note{NoteLetter: model.C},
				model.Note{NoteLetter: model.D},
				model.Barline{},
				model.Note{NoteLetter: model.E},
				model.Note{NoteLetter: model.F},
				model.Barline{},
				model.Note{NoteLetter: model.G},
				model.Note{NoteLetter: model.A},
			},
		},
		parseTestCase{
			label: "a note tied over many barlines",
			given: "marimba: c1|~1|~1~|1|~1~|2.",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"marimba"}},
				model.Note{
					NoteLetter: model.C,
					Duration: model.Duration{
						Components: []model.DurationComponent{
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 1},
							model.Barline{},
							model.NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
		},
	)

}
