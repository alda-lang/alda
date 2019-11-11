package model

import (
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func TestVariables(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "variable defined before parts",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "variable defined during part",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
					},
				},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "reference to undefined variable (before parts)",
			updates: []ScoreUpdate{
				VariableReference{VariableName: "lolbadvariable"},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "undefined variable") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "reference to undefined variable (during part)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "lolbadvariable"},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "undefined variable") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "a variable that references another variable",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
					},
				},
				VariableDefinition{
					VariableName: "bar",
					Events: []ScoreUpdate{
						VariableReference{VariableName: "foo"},
						Note{Pitch: LetterAndAccidentals{NoteLetter: F}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "bar"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64, 65, 67),
			},
		},
		scoreUpdateTestCase{
			label: "def a var that refs another var, then redef the first var",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
					},
				},
				VariableDefinition{
					VariableName: "bar",
					Events: []ScoreUpdate{
						VariableReference{VariableName: "foo"},
						Note{Pitch: LetterAndAccidentals{NoteLetter: F}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
					},
				},
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "bar"},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64, 65, 67, 60),
			},
		},
		scoreUpdateTestCase{
			label: "a variable definition that references an undefined variable",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						VariableReference{VariableName: "porkchopSandwiches"},
					},
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "undefined variable") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "reusing a variable in its own definition",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
					},
				},
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						VariableReference{VariableName: "foo"},
						Note{Pitch: LetterAndAccidentals{NoteLetter: F}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64, 65),
			},
		},
		scoreUpdateTestCase{
			label: "variable whose definition is an empty event sequence",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						EventSequence{Events: []ScoreUpdate{}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(),
			},
		},
		scoreUpdateTestCase{
			label: "variable whose definition is a non-empty event sequence",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						EventSequence{Events: []ScoreUpdate{
							Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
							Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
							Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
						}},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "variable whose definition is a repeated event sequence",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Repeat{
							Times: 3,
							Event: EventSequence{Events: []ScoreUpdate{
								Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
								Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
								Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
							}},
						},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				VariableReference{VariableName: "foo"},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 64, 60, 62, 64, 60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "repeating a variable",
			updates: []ScoreUpdate{
				VariableDefinition{
					VariableName: "foo",
					Events: []ScoreUpdate{
						Repeat{
							Times: 2,
							Event: EventSequence{Events: []ScoreUpdate{
								Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
								Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
							}},
						},
					},
				},
				PartDeclaration{Names: []string{"piano"}},
				Repeat{Times: 2, Event: VariableReference{VariableName: "foo"}},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 62, 60, 62, 60, 62, 60, 62),
			},
		},
	)
}
