package model

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestChords(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "chord",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// Add 1000ms to the initial offset of the chord in order to eliminate
				// false positives from offset initializing to 0
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},
				Chord{
					Events: []ScoreUpdate{
						Note{
							NoteLetter: E,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 1},
								},
							},
						},
						Note{
							NoteLetter: G,
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
							NoteLetter: G,
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
