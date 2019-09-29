package model

import (
	"fmt"
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
				func(s *Score) error {
					// The three notes of the chord, plus the one note we added at the
					// beginning to reduce false positives.
					expectedEvents := 4
					if len(s.Events) != expectedEvents {
						return fmt.Errorf(
							"there are %d events, not %d", len(s.Events), expectedEvents,
						)
					}

					piano, err := getPart(s, "piano")
					if err != nil {
						return err
					}

					// initial 1000ms buffer + 250, which is the shortest note/rest
					// duration in the chord (the 8th note rest == 250 ms)
					expectedCurrentOffset := 1250.0

					if piano.CurrentOffset != expectedCurrentOffset {
						return fmt.Errorf(
							"piano part's current offset is %f, not %f",
							piano.CurrentOffset,
							expectedCurrentOffset,
						)
					}

					expectedLastOffset := 1000.0

					if piano.LastOffset != expectedLastOffset {
						return fmt.Errorf(
							"piano part's last offset is %f, not %f",
							piano.LastOffset,
							expectedLastOffset,
						)
					}

					for i, event := range s.Events {
						expectedOffset := 0.0
						if i != 0 {
							expectedOffset = 1000
						}

						actualOffset := event.(NoteEvent).Offset
						if actualOffset != expectedOffset {
							return fmt.Errorf(
								"expected note #%d to have offset %f, but it was %f",
								i+1, expectedOffset, actualOffset,
							)
						}
					}

					return nil
				},
			},
		},
	)
}
