package model

import (
	"testing"

	_ "alda.io/client/testing"
)

type chordDurationMsTestCase struct {
	chord              Chord
	expectedDurationMs float64
}

func executeChordDurationMsTestCases(
	t *testing.T, testCases ...chordDurationMsTestCase,
) {
	// Boilerplate: create a new score and a new part (with a default tempo of 120
	// BPM)
	score := NewScore()
	part, err := score.NewPart("piano")
	if err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range testCases {
		actualDurationMs := testCase.chord.DurationMs(part)
		if testCase.expectedDurationMs != actualDurationMs {
			t.Errorf(
				"expected chord duration (ms) to be %f, but got %f\n",
				testCase.expectedDurationMs,
				actualDurationMs,
			)
		}
	}
}

func TestChords(t *testing.T) {
	// For the purposes of calculating duration in a cram expression, the duration
	// of a chord is the shortest duration of its events, not including octave
	// change events (which have an arbitrary duration of 0).
	executeChordDurationMsTestCases(
		t,
		chordDurationMsTestCase{
			chord: Chord{
				Events: []ScoreUpdate{
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
				},
			},
			expectedDurationMs: 500,
		},
		chordDurationMsTestCase{
			chord: Chord{
				Events: []ScoreUpdate{
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: E},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
				},
			},
			expectedDurationMs: 500,
		},
		chordDurationMsTestCase{
			chord: Chord{
				Events: []ScoreUpdate{
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 2},
							},
						},
					},
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: E},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
				},
			},
			expectedDurationMs: 500,
		},
		chordDurationMsTestCase{
			chord: Chord{
				Events: []ScoreUpdate{
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 8},
							},
						},
					},
					AttributeUpdate{
						PartUpdate: OctaveUp{},
					},
					Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
				},
			},
			expectedDurationMs: 250,
		},
		chordDurationMsTestCase{
			chord: Chord{
				Events: []ScoreUpdate{
					Note{
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 1},
							},
						},
					},
					Rest{
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 8, Dots: 1},
							},
						},
					},
					Note{
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 4},
							},
						},
					},
				},
			},
			expectedDurationMs: 375,
		},
	)

	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "chord",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// Add 1000ms to the initial offset of the chord in order to eliminate
				// false positives from offset initializing to 0
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},
				Chord{
					Events: []ScoreUpdate{
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: E},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 1},
								},
							},
						},
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: G},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 4},
								},
							},
						},
						Rest{
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 8},
								},
							},
						},
						AttributeUpdate{PartUpdate: OctaveUp{}},
						Note{
							Pitch: LetterAndAccidentals{NoteLetter: G},
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 2},
								},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				// The three notes of the chord, plus the one note we added at the
				// beginning to reduce false positives.
				expectNoteOffsets(0, 1000, 1000, 1000),
				// initial 1000ms buffer + 250, which is the shortest note/rest duration
				// in the chord (the 8th note rest == 250 ms)
				expectPartCurrentOffset("piano", 1250),
				expectPartLastOffset("piano", 1000),
			},
		},
	)
}
