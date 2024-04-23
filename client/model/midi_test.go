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

func manyInstrumentsScoreUpdates(instruments []string) []ScoreUpdate {
	updates := []ScoreUpdate{}

	for _, instrument := range instruments {
		updates = append(updates, []ScoreUpdate{
			PartDeclaration{
				Names: []string{instrument},
			},
			Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
		}...)
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
		var channel int32
		if instrument == "percussion" {
			channel = 9
		} else if i >= 9 {
			channel = int32(i + 1)
		} else {
			channel = int32(i)
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
	)
}
