package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func TestOctaves(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label:  "octave up",
			given:  ">",
			expect: []model.ScoreUpdate{model.OctaveUp{}},
		},
		parseTestCase{
			label:  "octave down",
			given:  "<",
			expect: []model.ScoreUpdate{model.OctaveDown{}},
		},
		parseTestCase{
			label:  "octave set",
			given:  "o5",
			expect: []model.ScoreUpdate{model.OctaveSet{OctaveNumber: 5}},
		},
		parseTestCase{
			label: "multi-octave up",
			given: ">>>",
			expect: []model.ScoreUpdate{
				model.OctaveUp{},
				model.OctaveUp{},
				model.OctaveUp{},
			},
		},
		parseTestCase{
			label: "multi-octave down",
			given: "<<<",
			expect: []model.ScoreUpdate{
				model.OctaveDown{},
				model.OctaveDown{},
				model.OctaveDown{},
			},
		},
		parseTestCase{
			label: "octave fish ><>",
			given: "><>",
			expect: []model.ScoreUpdate{
				model.OctaveUp{},
				model.OctaveDown{},
				model.OctaveUp{},
			},
		},
		parseTestCase{
			label: "octave up immediately followed by note",
			given: ">c",
			expect: []model.ScoreUpdate{
				model.OctaveUp{},
				model.Note{NoteLetter: model.C},
			},
		},
		parseTestCase{
			label: "octave down immediately followed by note",
			given: "<c",
			expect: []model.ScoreUpdate{
				model.OctaveDown{},
				model.Note{NoteLetter: model.C},
			},
		},
		parseTestCase{
			label: "note immediately followed by octave up",
			given: "c>",
			expect: []model.ScoreUpdate{
				model.Note{NoteLetter: model.C},
				model.OctaveUp{},
			},
		},
		parseTestCase{
			label: "note immediately followed by octave down",
			given: "c<",
			expect: []model.ScoreUpdate{
				model.Note{NoteLetter: model.C},
				model.OctaveDown{},
			},
		},
		parseTestCase{
			label: "note sandwiched between octave up/down",
			given: ">c<",
			expect: []model.ScoreUpdate{
				model.OctaveUp{},
				model.Note{NoteLetter: model.C},
				model.OctaveDown{},
			},
		},
	)
}
