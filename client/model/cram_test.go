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
						Note{NoteLetter: C, Slurred: true},
						Note{NoteLetter: D, Slurred: true},
						Note{NoteLetter: E, Slurred: true},
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
						Note{NoteLetter: C, Slurred: true},
						Note{NoteLetter: G, Slurred: true},
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
						Note{NoteLetter: C, Slurred: true},
						// half note, 500 ms
						Note{
							NoteLetter: D,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 2},
								},
								Slurred: true,
							},
						},
						// quarter note, 250 ms
						Note{
							NoteLetter: E,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 4},
								},
								Slurred: true,
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
						Note{NoteLetter: C, Slurred: true},
						Cram{
							Events: []ScoreUpdate{
								// 250 ms
								Note{NoteLetter: E, Slurred: true},
								// 250 ms
								Note{NoteLetter: G, Slurred: true},
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
							Note{NoteLetter: C, Slurred: true},
							Cram{
								Events: []ScoreUpdate{
									// 250 ms
									Note{NoteLetter: E, Slurred: true},
									// 250 ms
									Note{NoteLetter: G, Slurred: true},
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
				expectPartLastOffset("piano", 0),
			},
		},
	)
}
