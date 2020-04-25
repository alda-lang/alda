package model

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestCram(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "cram 3 notes into the span of a half note",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Cram{
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}, Slurred: true},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}, Slurred: true},
						Note{Pitch: LetterAndAccidentals{NoteLetter: E}, Slurred: true},
					},
					Duration: Duration{
						Components: []DurationComponent{
							// A half note at 120 BPM = 1000 ms
							NoteLength{Denominator: 2},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1000/3.0, (1000/3.0)*2),
				expectPartCurrentOffset("piano", 1000),
				expectPartLastOffset("piano", (1000/3.0)*2),
			},
		},
		scoreUpdateTestCase{
			label: "cram 2 notes into the span of an (implicit) whole note",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// A whole note at 120 bpm = 2000 ms
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-note-length"},
					LispNumber{Value: 1},
				}},
				Cram{
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}, Slurred: true},
						Note{Pitch: LetterAndAccidentals{NoteLetter: G}, Slurred: true},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1000),
				expectPartCurrentOffset("piano", 2000),
				expectPartLastOffset("piano", 1000),
			},
		},
		scoreUpdateTestCase{
			label: "cram a variety of note lengths into the span of a half note",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Cram{
					Events: []ScoreUpdate{
						// Implicit quarter note, 250 ms
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}, Slurred: true},
						// half note, 500 ms
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: D},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 2},
								},
							},
							Slurred: true,
						},
						// quarter note, 250 ms
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: E},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 4},
								},
							},
							Slurred: true,
						},
					},
					Duration: Duration{
						Components: []DurationComponent{
							// Total duration: 1000 ms
							NoteLength{Denominator: 2},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 250, 750),
				expectPartCurrentOffset("piano", 1000),
				expectPartLastOffset("piano", 750),
			},
		},
		scoreUpdateTestCase{
			label: "nested cram expressions",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Cram{
					Events: []ScoreUpdate{
						// 500 ms
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}, Slurred: true},
						Cram{
							Events: []ScoreUpdate{
								// 250 ms
								Note{Pitch: LetterAndAccidentals{NoteLetter: E}, Slurred: true},
								// 250 ms
								Note{Pitch: LetterAndAccidentals{NoteLetter: G}, Slurred: true},
							},
						},
					},
					Duration: Duration{
						Components: []DurationComponent{
							// Total duration: 1000 ms
							NoteLength{Denominator: 2},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 750),
				expectPartCurrentOffset("piano", 1000),
				expectPartLastOffset("piano", 750),
			},
		},
		scoreUpdateTestCase{
			label: "repeated nested cram expressions",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 2,
					Event: Cram{
						Events: []ScoreUpdate{
							// 500 ms
							Note{Pitch: LetterAndAccidentals{NoteLetter: C}, Slurred: true},
							Cram{
								Events: []ScoreUpdate{
									// 250 ms
									Note{Pitch: LetterAndAccidentals{NoteLetter: E}, Slurred: true},
									// 250 ms
									Note{Pitch: LetterAndAccidentals{NoteLetter: G}, Slurred: true},
								},
							},
						},
						Duration: Duration{
							Components: []DurationComponent{
								// Total duration: 1000 ms
								NoteLength{Denominator: 2},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 750, 1000, 1500, 1750),
				expectPartCurrentOffset("piano", 2000),
				expectPartLastOffset("piano", 1750),
			},
		},
		scoreUpdateTestCase{
			label: "cram expression containing a repeat with OnRepetitions",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Cram{
					Events: []ScoreUpdate{
						Repeat{
							Times: 2,
							Event: EventSequence{
								Events: []ScoreUpdate{
									Note{
										Pitch:   LetterAndAccidentals{NoteLetter: C},
										Slurred: true,
									},
									Note{
										Pitch:   LetterAndAccidentals{NoteLetter: C},
										Slurred: true,
									},
									OnRepetitions{
										Repetitions: []RepetitionRange{
											RepetitionRange{First: 1, Last: 1},
										},
										Event: Note{
											Pitch:   LetterAndAccidentals{NoteLetter: C},
											Slurred: true,
										},
									},
									OnRepetitions{
										Repetitions: []RepetitionRange{
											RepetitionRange{First: 2, Last: 2},
										},
										Event: Note{
											Pitch:   LetterAndAccidentals{NoteLetter: D},
											Slurred: true,
										},
									},
									Note{
										Pitch:   LetterAndAccidentals{NoteLetter: C},
										Slurred: true,
									},
								},
							},
						},
					},
					Duration: Duration{
						Components: []DurationComponent{
							// Total duration: 1000 ms
							NoteLength{Denominator: 2},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 125, 250, 375, 500, 625, 750, 875),
				expectMidiNoteNumbers(60, 60, 60, 60, 60, 60, 62, 60),
			},
		},
		scoreUpdateTestCase{
			label: "cram expression containing variable reference",
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
				Cram{
					Events: []ScoreUpdate{
						VariableReference{VariableName: "foo"},
						VariableReference{VariableName: "foo"},
					},
					Duration: Duration{
						Components: []DurationComponent{
							// Total duration: 1000 ms
							NoteLength{Denominator: 2},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(
					(1000.0/6)*0,
					(1000.0/6)*1,
					(1000.0/6)*2,
					(1000.0/6)*3,
					(1000.0/6)*4,
					(1000.0/6)*5,
				),
				expectMidiNoteNumbers(60, 62, 64, 60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "the specified durations inside a cram expression shouldn't " +
				"affect parts' default duration",
			updates: []ScoreUpdate{
				// default duration is a quarter note
				PartDeclaration{Names: []string{"piano"}},
				// the events inside the cram shouldn't change that
				Cram{
					Events: []ScoreUpdate{
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: C},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 8},
								},
							},
						},
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: C},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 1},
								},
							},
						},
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: C},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 8},
								},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 1),
			},
		},
		scoreUpdateTestCase{
			label: "the specified duration outside a cram expression should affect " +
				"parts' default duration",
			updates: []ScoreUpdate{
				// default duration is a quarter note
				PartDeclaration{Names: []string{"piano"}},
				// the cram's duration is a sixteenth note, which should change the
				// part's default duration to a sixteenth note
				Cram{
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
					},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 16},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 0.25),
			},
		},
	)
}
