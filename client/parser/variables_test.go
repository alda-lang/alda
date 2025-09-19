package parser

import (
	"fmt"
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

func variableNameCheck(name string) parseTestCase {
	return parseTestCase{
		label: fmt.Sprintf("variable definition and reference: %s", name),
		given: fmt.Sprintf("%[1]s = r \n%[1]s", name),
		expectUpdates: []model.ScoreUpdate{
			variableDefinition(name, model.Rest{}),
			variableReference(name),
		},
	}
}

func TestVariables(t *testing.T) {
	executeParseTestCases(
		t,
		variableNameCheck("aa"),
		variableNameCheck("aaa"),
		variableNameCheck("HI"),
		variableNameCheck("celloPart2"),
		variableNameCheck("xy42"),
		variableNameCheck("my20cats"),
		variableNameCheck("apple_cider"),
		variableNameCheck("underscores_are_great"),
		parseTestCase{
			label: "variable reference in a part",
			given: "flan = e\nflute: c flan f",
			expectUpdates: []model.ScoreUpdate{
				variableDefinition("flan",
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
					},
				),
				model.PartDeclaration{Names: []string{"flute"}},
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
				variableReference("flan"),
				model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.F}},
			},
		},
		parseTestCase{
			label: "only a variable reference in a part",
			given: "pudding123 = r\nclarinet: pudding123",
			expectUpdates: []model.ScoreUpdate{
				variableDefinition("pudding123", model.Rest{}),
				model.PartDeclaration{Names: []string{"clarinet"}},
				variableReference("pudding123"),
			},
		},
		parseTestCase{
			label: "variable definition containing a cram expression",
			given: "cheesecake = { c/e }2",
			expectUpdates: []model.ScoreUpdate{
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
			expectUpdates: []model.ScoreUpdate{
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
			expectUpdates: []model.ScoreUpdate{
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
			expectUpdates: []model.ScoreUpdate{
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
			expectUpdates: []model.ScoreUpdate{
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
			given: `
			soprano = r
			alto = r
			tenor = r
			bass = r
			satb = V1: soprano V2: alto V3: tenor V4: bass
			`,
			expectUpdates: []model.ScoreUpdate{
				variableDefinition("soprano", model.Rest{}),
				variableDefinition("alto", model.Rest{}),
				variableDefinition("tenor", model.Rest{}),
				variableDefinition("bass", model.Rest{}),
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
			label: "variable definition followed by a newline",
			given: `bar = r
			foo = bar
			`,
			expectUpdates: []model.ScoreUpdate{
				variableDefinition("bar", model.Rest{}),
				variableDefinition("foo", variableReference("bar")),
			},
		},
	)
}
