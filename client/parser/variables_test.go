package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func variableDefinition(name string, events ...model.ScoreUpdate,
) model.VariableDefinition {
	return model.VariableDefinition{VariableName: name, Events: events}
}

func variableReference(name string) model.VariableReference {
	return model.VariableReference{VariableName: name}
}

func TestVariables(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label:  "variable reference: aa",
			given:  "aa",
			expect: []model.ScoreUpdate{variableReference("aa")},
		},
		parseTestCase{
			label:  "variable reference: aaa",
			given:  "aaa",
			expect: []model.ScoreUpdate{variableReference("aaa")},
		},
		parseTestCase{
			label:  "variable reference: HI",
			given:  "HI",
			expect: []model.ScoreUpdate{variableReference("HI")},
		},
		parseTestCase{
			label:  "variable reference: celloPart2",
			given:  "celloPart2",
			expect: []model.ScoreUpdate{variableReference("celloPart2")},
		},
		parseTestCase{
			label:  "variable reference: xy42",
			given:  "xy42",
			expect: []model.ScoreUpdate{variableReference("xy42")},
		},
		parseTestCase{
			label:  "variable reference: my20cats",
			given:  "my20cats",
			expect: []model.ScoreUpdate{variableReference("my20cats")},
		},
		parseTestCase{
			label:  "variable reference: apple_cider",
			given:  "apple_cider",
			expect: []model.ScoreUpdate{variableReference("apple_cider")},
		},
		parseTestCase{
			label:  "variable reference: underscores_are_great",
			given:  "underscores_are_great",
			expect: []model.ScoreUpdate{variableReference("underscores_are_great")},
		},
		parseTestCase{
			label: "variable reference in a part",
			given: "flute: c flan f",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"flute"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				variableReference("flan"),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
			},
		},
		parseTestCase{
			label: "only a variable reference in a part",
			given: "clarinet: pudding123",
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"clarinet"}},
				variableReference("pudding123"),
			},
		},
		parseTestCase{
			label: "variable definition containing a cram expression",
			given: "cheesecake = { c/e }2",
			expect: []model.ScoreUpdate{
				variableDefinition(
					"cheesecake",
					model.Cram{
						Events: []model.ScoreUpdate{
							model.Chord{
								Events: []model.ScoreUpdate{
									model.Note{
										Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
									},
									model.Note{
										Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
									},
								},
							},
						},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 2},
							},
						},
					},
				),
			},
		},
		parseTestCase{
			label: "variable definition within an instrument part",
			given: `harpsichord:
			custard_ = c d e/g`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"harpsichord"}},
				variableDefinition(
					"custard_",
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Chord{
						Events: []model.ScoreUpdate{
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
							},
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.G},
							},
						},
					},
				),
			},
		},
		parseTestCase{
			label: "variable definition within an instrument part (variation)",
			given: `glockenspiel:

			sorbet=c d e/g
			c`,
			expect: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"glockenspiel"}},
				variableDefinition(
					"sorbet",
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Chord{
						Events: []model.ScoreUpdate{
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
							},
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.G},
							},
						},
					},
				),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
			},
		},
		parseTestCase{
			label: "variable definition before an instrument part",
			given: `GELATO=d e

			clavinet: c/f`,
			expect: []model.ScoreUpdate{
				variableDefinition(
					"GELATO",
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
				),
				model.PartDeclaration{Names: []string{"clavinet"}},
				model.Chord{
					Events: []model.ScoreUpdate{
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
					},
				},
			},
		},
		// Regression test for https://github.com/alda-lang/alda-core/issues/52
		// which involved incorrectly parsing a rest `r` followed by a newline
		// within a variable definition.
		parseTestCase{
			label: "var defined and used s/t var ends with a rest",
			given: "foo = c8 d c r\npiano: foo*2",
			expect: []model.ScoreUpdate{
				variableDefinition(
					"foo",
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 8},
							},
						},
					},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
					model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
					model.Rest{},
				),
				model.PartDeclaration{Names: []string{"piano"}},
				model.Repeat{Event: variableReference("foo"), Times: 2},
			},
		},
		// Regression test for https://github.com/alda-lang/alda-core/issues/64
		// NB: the trailing newline was essential to reproducing the issue!
		parseTestCase{
			label: "variable definition ending with a variable reference",
			given: "satb = V1: soprano V2: alto V3: tenor V4: bass\n",
			expect: []model.ScoreUpdate{
				variableDefinition(
					"satb",
					model.VoiceMarker{VoiceNumber: 1},
					variableReference("soprano"),
					model.VoiceMarker{VoiceNumber: 2},
					variableReference("alto"),
					model.VoiceMarker{VoiceNumber: 3},
					variableReference("tenor"),
					model.VoiceMarker{VoiceNumber: 4},
					variableReference("bass"),
				),
			},
		},
		// Regression test for https://github.com/alda-lang/alda-core/issues/64
		// NB: the trailing newline was essential to reproducing the issue!
		parseTestCase{
			label: "variable definition ending with a variable reference",
			given: "foo = bar\n",
			expect: []model.ScoreUpdate{
				variableDefinition("foo", variableReference("bar")),
			},
		},
	)
}
