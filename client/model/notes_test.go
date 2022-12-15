package model

import (
	"fmt"
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func expectNoteOffsets(expectedOffsets ...float64) func(*Score) error {
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

func expectNoteFloatValues(
	valueName string, method func(NoteEvent) float64, expectedValues []float64,
) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedValues) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedValues),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedValues); i++ {
			expectedValue := expectedValues[i]
			actualValue := method(s.Events[i].(NoteEvent))
			if !equalish(expectedValue, actualValue) {
				return fmt.Errorf(
					"expected note #%d to have %s %f, but it was %f",
					i+1, valueName, expectedValue, actualValue,
				)
			}
		}

		return nil
	}
}

func expectNoteDurations(expectedDurations ...float64) func(*Score) error {
	return expectNoteFloatValues(
		"audible duration",
		func(note NoteEvent) float64 { return note.Duration },
		expectedDurations,
	)
}

func expectNoteAudibleDurations(
	expectedAudibleDurations ...float64,
) func(*Score) error {
	return expectNoteFloatValues(
		"audible duration",
		func(note NoteEvent) float64 { return note.AudibleDuration },
		expectedAudibleDurations,
	)
}

func expectMidiNoteNumbers(expectedNoteNumbers ...int32) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedNoteNumbers) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedNoteNumbers),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedNoteNumbers); i++ {
			expectedNoteNumber := expectedNoteNumbers[i]
			actualNoteNumber := s.Events[i].(NoteEvent).MidiNote
			if expectedNoteNumber != actualNoteNumber {
				return fmt.Errorf(
					"expected note #%d to be MIDI note %d, but it was MIDI note %d",
					i+1,
					expectedNoteNumber,
					actualNoteNumber,
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
			label: "notes with provided durations",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: D},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 8},
						},
					},
				},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: E},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 2222},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1500, 1750),
				expectNoteDurations(1500, 250, 2222),
				expectMidiNoteNumbers(60, 62, 64),
			},
		},
		scoreUpdateTestCase{
			label: "implicit note duration",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
				Note{
					Pitch: LetterAndAccidentals{
						NoteLetter: D,
					},
				},
				Note{
					Pitch: LetterAndAccidentals{
						NoteLetter: D, Accidentals: []Accidental{Sharp},
					},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 50},
						},
					},
				},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: E},
				},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: F},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 8},
						},
					},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1500, 3000, 3050, 3100, 3350),
				expectNoteDurations(1500, 1500, 50, 50, 250, 250),
				expectMidiNoteNumbers(60, 62, 63, 64, 65, 67),
			},
		},
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
		scoreUpdateTestCase{
			label: "lisp note: pitch, note-length",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				// (note (pitch '(d)) (note-length 1))
				LispList{Elements: []LispForm{
					LispSymbol{Name: "note"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "pitch"},
						LispQuotedForm{Form: LispList{
							Elements: []LispForm{
								LispSymbol{Name: "d"},
							},
						}},
					}},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "note-length"},
						LispNumber{Value: 1},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteDurations(2000),
				expectMidiNoteNumbers(62),
			},
		},
		scoreUpdateTestCase{
			label: "lisp note: pitch w/ accidental, dotted note-length",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				// (note (pitch '(f sharp)) (note-length "1.."))
				LispList{Elements: []LispForm{
					LispSymbol{Name: "note"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "pitch"},
						LispQuotedForm{Form: LispList{
							Elements: []LispForm{
								LispSymbol{Name: "f"},
								LispSymbol{Name: "sharp"},
							},
						}},
					}},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "note-length"},
						LispString{Value: "1.."},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteDurations(3500),
				expectMidiNoteNumbers(66),
			},
		},
		scoreUpdateTestCase{
			label: "lisp note: midi-note, ms",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				// (note (midi-note 42) (ms 1234))
				LispList{Elements: []LispForm{
					LispSymbol{Name: "note"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "midi-note"},
						LispNumber{Value: 42},
					}},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "ms"},
						LispNumber{Value: 1234},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteDurations(1234),
				expectMidiNoteNumbers(42),
			},
		},
		scoreUpdateTestCase{
			label: "lisp note: pitch, multiple duration components",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				// (note (pitch '(c))
				//       (duration (note-length 4)
				//                 (ms 2222)
				//                 (note-length 8)))
				LispList{Elements: []LispForm{
					LispSymbol{Name: "note"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "pitch"},
						LispQuotedForm{Form: LispList{
							Elements: []LispForm{
								LispSymbol{Name: "c"},
							},
						}},
					}},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "duration"},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "note-length"},
							LispNumber{Value: 4},
						}},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "ms"},
							LispNumber{Value: 2222},
						}},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "note-length"},
							LispNumber{Value: 8},
						}},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteDurations(2972),
				expectMidiNoteNumbers(60),
			},
		},
		scoreUpdateTestCase{
			label: "slurred note",
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
				// (slur (note (pitch '(g)) (note-length 4)))
				LispList{Elements: []LispForm{
					LispSymbol{Name: "slur"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "note"},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "pitch"},
							LispQuotedForm{Form: LispList{
								Elements: []LispForm{
									LispSymbol{Name: "g"},
								},
							}},
						}},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "note-length"},
							LispNumber{Value: 4},
						}},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteAudibleDurations(500),
				expectMidiNoteNumbers(67),
			},
		},
		scoreUpdateTestCase{
			// r4. (pause) c2.
			label: "Rest with implicit duration from previous rest",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4, Dots: 1},
						},
					},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "pause"},
				}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(1500),
				expectNoteDurations(1500),
			},
		},
		scoreUpdateTestCase{
			// c4. (pause) c2.
			label: "Rest with implicit duration from previous note",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4, Dots: 1},
						},
					},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "pause"},
				}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 1500),
				expectNoteDurations(750, 1500),
			},
		},
		scoreUpdateTestCase{
			// (pause (note-length 4)) c2.
			label: "Rest with note length 4",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "pause"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "note-length"},
						LispNumber{Value: 4},
					}},
				}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(500),
				expectNoteDurations(1500),
			},
		},
		scoreUpdateTestCase{
			// (pause (duration (ms 12345))) c2.
			label: "Rest with duration of 12345ms",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "pause"},
					LispList{Elements: []LispForm{
						LispSymbol{Name: "duration"},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "ms"},
							LispNumber{Value: 12345},
						}},
					}},
				}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(12345),
				expectNoteDurations(1500),
			},
		},
		scoreUpdateTestCase{
			// alda play -c 'piano: o10 c' - MIDI note out of range
			label: "C note with MIDI value out of range",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 10}},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: C},
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "MIDI note out of the 0-127 range") {
						return err
					}
					return nil
				},
			},
		},
	)
}

func TestNoteValidation(t *testing.T) {
	for _, durationComponent := range []DurationComponent{
		NoteLength{Denominator: -1},
		NoteLength{Denominator: 0},
		NoteLengthBeats{Quantity: -1},
		NoteLengthBeats{Quantity: 0},
		NoteLengthMs{Quantity: -1},
		NoteLengthMs{Quantity: 0},
	} {
		score := NewScore()
		err := score.Update(
			PartDeclaration{Names: []string{"piano"}},
			Note{
				Pitch: LetterAndAccidentals{NoteLetter: C},
				Duration: Duration{
					Components: []DurationComponent{durationComponent},
				},
			},
		)

		if err == nil {
			t.Errorf(
				"unexpected success: duration component %#v\n", durationComponent,
			)
		}
	}
}
