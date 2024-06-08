package model

import (
	"fmt"
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func expectPartMidiChannel(
	instrument string, expectedChannel int32,
) func(s *Score) error {
	return func(s *Score) error {
		part, err := getPart(s, instrument)
		if err != nil {
			return err
		}

		if part.MidiChannel != expectedChannel {
			return fmt.Errorf(
				"expected %s part to have MIDI channel %d, got %d",
				instrument,
				expectedChannel,
				part.MidiChannel,
			)
		}

		return nil
	}
}

func expectNoteMidiChannels(expectedMidiChannels ...int32) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedMidiChannels) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedMidiChannels),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedMidiChannels); i++ {
			expectedMidiChannel := expectedMidiChannels[i]
			actualMidiChannel := s.Events[i].(NoteEvent).MidiChannel
			if expectedMidiChannel != actualMidiChannel {
				return fmt.Errorf(
					"expected note #%d to have MIDI channel %d, but it was %d",
					i+1,
					expectedMidiChannel,
					actualMidiChannel,
				)
			}
		}

		return nil
	}
}

func fifteenInstruments() []string {
	return []string{
		"piano",
		"harpsichord",
		"clavinet",
		"celeste",
		"glockenspiel",
		"music-box",
		"vibraphone",
		"marimba",
		"xylophone",
		"tubular-bells",
		"dulcimer",
		"organ",
		"accordion",
		"harmonica",
		"guitar",
	}
}

func sixteenInstruments() []string {
	return append(fifteenInstruments(), "percussion")
}

func fifteenMoreInstruments() []string {
	return []string{
		"upright-bass",
		"violin",
		"viola",
		"cello",
		"double-bass",
		"harp",
		"timpani",
		"trumpet",
		"trombone",
		"tuba",
		"french-horn",
		"soprano-sax",
		"alto-sax",
		"tenor-sax",
		"bari-sax",
	}
}

func manyInstrumentsScoreUpdates(
	instruments []string, eventsBeforeNote ...ScoreUpdate,
) []ScoreUpdate {
	updates := []ScoreUpdate{}

	for _, instrument := range instruments {
		updates = append(updates, PartDeclaration{Names: []string{instrument}})
		updates = append(updates, eventsBeforeNote...)
		updates = append(updates, Note{Pitch: LetterAndAccidentals{NoteLetter: C}})
	}

	return updates
}

func manyInstrumentsExpectations(
	instruments []string,
) []scoreUpdateExpectation {
	expectations := []scoreUpdateExpectation{
		expectParts(instruments...),
	}

	for i, instrument := range instruments {
		j := i % 16

		var channel int32
		if instrument == "percussion" {
			channel = 9
		} else if j >= 9 {
			channel = int32(j + 1)
		} else {
			channel = int32(j)
		}

		expectations = append(
			expectations, expectPartMidiChannel(instrument, channel),
		)
	}

	return expectations
}

func fifteenInstrumentsScoreUpdates() []ScoreUpdate {
	return manyInstrumentsScoreUpdates(fifteenInstruments())
}

func sixteenInstrumentsScoreUpdates() []ScoreUpdate {
	return manyInstrumentsScoreUpdates(sixteenInstruments())
}

// 17 instruments all playing at the same time. Problematic because there are
// only 16 channels.
func seventeenInstrumentsScoreUpdatesWithConflict() []ScoreUpdate {
	firstSixteen := sixteenInstruments()
	oneMore := fifteenMoreInstruments()[0]

	return append(
		manyInstrumentsScoreUpdates(firstSixteen),
		manyInstrumentsScoreUpdates([]string{oneMore})...,
	)
}

func fifteenInstrumentsExpectations() []scoreUpdateExpectation {
	return manyInstrumentsExpectations(fifteenInstruments())
}

func sixteenInstrumentsExpectations() []scoreUpdateExpectation {
	return manyInstrumentsExpectations(sixteenInstruments())
}

func thirtyOneInstrumentsScoreUpdates() []ScoreUpdate {
	return append(
		// 16 instruments (15 non-percussion, 1 percussion), each playing 1 note at
		// offset 0
		manyInstrumentsScoreUpdates(sixteenInstruments()),
		// 15 more instruments (all non-percussion), each playing 1 note at offset 0
		// + 1 whole note
		//
		// All 16 MIDI channels should be available for this, since the first 16
		// instruments only played 1 note at offset 0 and were done playing before
		// these next instruments come in with their notes.
		manyInstrumentsScoreUpdates(fifteenMoreInstruments(),
			Rest{Duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 1},
				},
			}})...,
	)
}

func thirtyOneInstrumentsExpectations() []scoreUpdateExpectation {
	return manyInstrumentsExpectations(
		append(sixteenInstruments(), fifteenMoreInstruments()...),
	)
}

func TestMidiChannelAssignment(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "initial MIDI channel value",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectPartMidiChannel("piano", -1),
			},
		},
		scoreUpdateTestCase{
			label: "automatic channel assignment - 1 part",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectPartMidiChannel("piano", 0),
				expectNoteMidiChannels(0),
			},
		},
		scoreUpdateTestCase{
			label: "automatic channel assignment - 2 parts (separate)",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				PartDeclaration{
					Names: []string{"bassoon"},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "bassoon"),
				expectPartMidiChannel("piano", 0),
				expectPartMidiChannel("bassoon", 1),
				expectNoteMidiChannels(0, 1),
			},
		},
		scoreUpdateTestCase{
			label: "automatic channel assignment - 2 parts (together)",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "bassoon"},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "bassoon"),
				expectPartMidiChannel("piano", 0),
				expectPartMidiChannel("bassoon", 1),
				expectNoteMidiChannels(0, 1),
			},
		},
		scoreUpdateTestCase{
			label: "automatic channel assignment - 1 part with voices",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
				VoiceMarker{VoiceNumber: 2},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: A}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectPartMidiChannel("piano", 0),
				expectNoteMidiChannels(0, 0, 0, 0, 0, 0),
			},
		},
		scoreUpdateTestCase{
			label:        "automatic channel assignment - 15 instruments",
			updates:      fifteenInstrumentsScoreUpdates(),
			expectations: fifteenInstrumentsExpectations(),
		},
		scoreUpdateTestCase{
			label:        "automatic channel assignment - 16 instruments",
			updates:      sixteenInstrumentsScoreUpdates(),
			expectations: sixteenInstrumentsExpectations(),
		},
		scoreUpdateTestCase{
			label:   "automatic channel assignment - 17 instruments + channel conflict",
			updates: seventeenInstrumentsScoreUpdatesWithConflict(),
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "No MIDI channel available") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label:        "automatic channel assignment - 31 instruments",
			updates:      thirtyOneInstrumentsScoreUpdates(),
			expectations: thirtyOneInstrumentsExpectations(),
		},
	)
}

func TestMidiChannelAttribute(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "midi-channel attribute - 1 part",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 5},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectPartMidiChannel("piano", 5),
				expectNoteMidiChannels(5),
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - 2 parts, different channels",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 2},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				PartDeclaration{
					Names: []string{"guitar"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 5},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "guitar"),
				expectPartMidiChannel("piano", 2),
				expectPartMidiChannel("guitar", 5),
				expectNoteMidiChannels(2, 5),
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - 2 parts sharing same channel",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 2},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				PartDeclaration{
					Names: []string{"guitar"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 2},
				}},
				Rest{Duration: Duration{
					Components: []DurationComponent{
						NoteLength{Denominator: 1},
					},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "guitar"),
				expectPartMidiChannel("piano", 2),
				expectPartMidiChannel("guitar", 2),
				expectNoteMidiChannels(2, 2),
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - 2 parts with channel conflict",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 2},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				PartDeclaration{
					Names: []string{"guitar"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 2},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "not available") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - assign to invalid channel",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 42},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "expected integer in range 0-15") {
						return err
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - assign percussion to channel 9",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"percussion"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 9},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("percussion"),
				expectPartMidiChannel("percussion", 9),
				expectNoteMidiChannels(9),
			},
		},
		scoreUpdateTestCase{
			label: "midi-channel attribute - assign non-percussion to channel 9",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"bassoon"},
				},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "midi-channel"},
					LispNumber{Value: 9},
				}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "channel 9") {
						return err
					}

					return nil
				},
			},
		},
	)
}
