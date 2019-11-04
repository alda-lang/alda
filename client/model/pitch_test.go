package model

import (
	"testing"

	_ "alda.io/client/testing"
)

type midiNoteNumberTestCase struct {
	pitch                  PitchIdentifier
	octave                 int32
	keySignature           KeySignature
	transposition          int32
	expectedMidiNoteNumber int32
}

func TestPitch(t *testing.T) {
	for _, testCase := range []midiNoteNumberTestCase{
		{
			pitch:                  LetterAndAccidentals{NoteLetter: A},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 69,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: A},
			octave:                 5,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 81,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: C},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 60,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 61,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  D,
				Accidentals: []Accidental{Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 61,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  B,
				Accidentals: []Accidental{Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 70,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: B},
			octave:                 4,
			keySignature:           KeySignature{B: {Flat}},
			transposition:          0,
			expectedMidiNoteNumber: 70,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  B,
				Accidentals: []Accidental{Natural},
			},
			octave:                 4,
			keySignature:           KeySignature{B: {Flat}},
			transposition:          0,
			expectedMidiNoteNumber: 71,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp, Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 62,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  A,
				Accidentals: []Accidental{Flat, Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 67,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp, Flat, Flat, Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 60,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: G},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          2,
			expectedMidiNoteNumber: 69,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: C},
			octave:                 5,
			keySignature:           KeySignature{},
			transposition:          -3,
			expectedMidiNoteNumber: 69,
		},
		{
			pitch: LetterAndAccidentals{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          -1,
			expectedMidiNoteNumber: 60,
		},
		{
			pitch:                  LetterAndAccidentals{NoteLetter: C},
			octave:                 4,
			keySignature:           KeySignature{C: {Sharp}},
			transposition:          -1,
			expectedMidiNoteNumber: 60,
		},
		{
			pitch:                  MidiNoteNumber{MidiNote: 42},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 42,
		},
		{
			pitch:                  MidiNoteNumber{MidiNote: 42},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          10,
			expectedMidiNoteNumber: 52,
		},
	} {
		label := "Note => MIDI note number conversion"

		actualMidiNoteNumber := testCase.pitch.CalculateMidiNote(
			testCase.octave,
			testCase.keySignature,
			testCase.transposition,
		)

		if actualMidiNoteNumber != testCase.expectedMidiNoteNumber {
			t.Error(label)
			t.Errorf(
				"Expected pitch %#v (octave %d, %#v, transposition %d) "+
					"to be MIDI note %d, but it was MIDI note %d",
				testCase.pitch,
				testCase.octave,
				testCase.keySignature,
				testCase.transposition,
				testCase.expectedMidiNoteNumber,
				actualMidiNoteNumber,
			)
		}
	}
}
