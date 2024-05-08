package model

import (
	"fmt"
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
			label:        "automatic channel assignment - 31 instruments",
			updates:      thirtyOneInstrumentsScoreUpdates(),
			expectations: thirtyOneInstrumentsExpectations(),
		},
	)
}
