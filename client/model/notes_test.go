package model

import (
	"fmt"
	"testing"

	_ "alda.io/client/testing"
)

func expectNoteOffsets(expectedOffsets ...OffsetMs) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedOffsets) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedOffsets),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedOffsets); i++ {
			expectedOffset := expectedOffsets[i]
			actualOffset := s.Events[i].(NoteEvent).Offset
			if !equalish(expectedOffset, actualOffset) {
				return fmt.Errorf(
					"expected note #%d to have offset %f, but it was %f",
					i+1,
					expectedOffset,
					actualOffset,
				)
			}
		}

		return nil
	}
}

func expectNoteAudibleDurations(
	expectedAudibleDurations ...float32,
) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedAudibleDurations) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedAudibleDurations),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedAudibleDurations); i++ {
			expectedAudibleDuration := expectedAudibleDurations[i]
			actualAudibleDuration := s.Events[i].(NoteEvent).AudibleDuration
			if !equalish32(expectedAudibleDuration, actualAudibleDuration) {
				return fmt.Errorf(
					"expected note #%d to have audible duration %f, but it was %f",
					i+1,
					expectedAudibleDuration,
					actualAudibleDuration,
				)
			}
		}

		return nil
	}
}

func TestNotes(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "note with 100% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 100},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(500),
			},
		},
		scoreUpdateTestCase{
			label: "note with 90% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(450),
			},
		},
		scoreUpdateTestCase{
			label: "note with 0% quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 0},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(),
			},
		},
		scoreUpdateTestCase{
			label: "slurred notes ignore quantization #1",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
					Slurred: true,
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(500),
			},
		},
		scoreUpdateTestCase{
			label: "slurred notes ignore quantization #2",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 90},
				}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
					Slurred: true,
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(1000),
			},
		},
	)
}
