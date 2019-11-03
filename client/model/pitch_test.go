package model

import (
	"testing"

	_ "alda.io/client/testing"
)

type noteToMidiNoteNumberTestCase struct {
	note                   Note
	octave                 int32
	keySignature           KeySignature
	transposition          int32
	expectedMidiNoteNumber int32
}

func TestPitch(t *testing.T) {
	for _, testCase := range []noteToMidiNoteNumberTestCase{
		{
			note: Note{
				NoteLetter: A,
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 69,
		},
		{
			note: Note{
				NoteLetter: A,
			},
			octave:                 5,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 81,
		},
		{
			note: Note{
				NoteLetter: C,
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 60,
		},
		{
			note: Note{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 61,
		},
		{
			note: Note{
				NoteLetter:  D,
				Accidentals: []Accidental{Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 61,
		},
		{
			note: Note{
				NoteLetter:  B,
				Accidentals: []Accidental{Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 70,
		},
		{
			note: Note{
				NoteLetter: B,
			},
			octave:                 4,
			keySignature:           KeySignature{B: {Flat}},
			transposition:          0,
			expectedMidiNoteNumber: 70,
		},
		{
			note: Note{
				NoteLetter:  B,
				Accidentals: []Accidental{Natural},
			},
			octave:                 4,
			keySignature:           KeySignature{B: {Flat}},
			transposition:          0,
			expectedMidiNoteNumber: 71,
		},
		{
			note: Note{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp, Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 62,
		},
		{
			note: Note{
				NoteLetter:  A,
				Accidentals: []Accidental{Flat, Flat},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 67,
		},
		{
			note: Note{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp, Flat, Flat, Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          0,
			expectedMidiNoteNumber: 60,
		},
		{
			note: Note{
				NoteLetter: G,
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          2,
			expectedMidiNoteNumber: 69,
		},
		{
			note: Note{
				NoteLetter: C,
			},
			octave:                 5,
			keySignature:           KeySignature{},
			transposition:          -3,
			expectedMidiNoteNumber: 69,
		},
		{
			note: Note{
				NoteLetter:  C,
				Accidentals: []Accidental{Sharp},
			},
			octave:                 4,
			keySignature:           KeySignature{},
			transposition:          -1,
			expectedMidiNoteNumber: 60,
		},
		{
			note: Note{
				NoteLetter: C,
			},
			octave:                 4,
			keySignature:           KeySignature{C: {Sharp}},
			transposition:          -1,
			expectedMidiNoteNumber: 60,
		},
	} {
		label := "Note => MIDI note number conversion"

		actualMidiNoteNumber := CalculateMidiNote(
			testCase.note,
			testCase.octave,
			testCase.keySignature,
			testCase.transposition,
		)

		if actualMidiNoteNumber != testCase.expectedMidiNoteNumber {
			t.Error(label)
			t.Errorf(
				"Expected note %#v (octave %d, %#v, transposition %d) "+
					"to be MIDI note %d, but it was MIDI note %d",
				testCase.note,
				testCase.octave,
				testCase.keySignature,
				testCase.transposition,
				testCase.expectedMidiNoteNumber,
				actualMidiNoteNumber,
			)
		}
	}
}
