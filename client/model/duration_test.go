package model

import (
	"testing"

	_ "alda.io/client/testing"
)

type noteLengthToBeatsTestCase struct {
	noteLength    NoteLength
	expectedBeats float64
}

type durationToMsTestCase struct {
	duration   Duration
	tempo      float64
	expectedMs float64
}

type noteDurationTestCase struct {
	note         Note
	tempo        float64
	quantization float64
	expectedMs   float64
}

func TestDuration(t *testing.T) {
	for _, testCase := range []noteLengthToBeatsTestCase{
		{noteLength: NoteLength{Denominator: 4}, expectedBeats: 1},
		{noteLength: NoteLength{Denominator: 4, Dots: 1}, expectedBeats: 1.5},
		{noteLength: NoteLength{Denominator: 1}, expectedBeats: 4},
		{noteLength: NoteLength{Denominator: 1, Dots: 1}, expectedBeats: 6},
		{noteLength: NoteLength{Denominator: 1, Dots: 2}, expectedBeats: 7},
	} {
		label := "Note length => beats conversion"

		actualBeats, err := testCase.noteLength.Beats()

		if err != nil {
			t.Error(label)
			t.Error(err)
			return
		}

		if !equalish(actualBeats, testCase.expectedBeats) {
			t.Error(label)
			t.Errorf(
				"Expected %#v to equal %f beats, got %f beats",
				testCase.noteLength, testCase.expectedBeats, actualBeats,
			)
		}
	}

	for _, testCase := range []durationToMsTestCase{
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 4},
				},
			},
			tempo:      60,
			expectedMs: 1000,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 2},
					NoteLength{Denominator: 2},
					NoteLength{Denominator: 2, Dots: 2},
				},
			},
			tempo:      60,
			expectedMs: 7500,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 4},
				},
			},
			tempo:      120,
			expectedMs: 500,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 4, Dots: 1},
				},
			},
			tempo:      120,
			expectedMs: 750,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLengthMs{Quantity: 1000},
				},
			},
			tempo:      42,
			expectedMs: 1000,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLengthMs{Quantity: 2000},
					NoteLengthMs{Quantity: 2000},
					NoteLengthMs{Quantity: 3500},
				},
			},
			tempo:      123,
			expectedMs: 7500,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLengthMs{Quantity: 2000},
					NoteLength{Denominator: 2},
					NoteLengthMs{Quantity: 45},
				},
			},
			tempo:      60,
			expectedMs: 4045,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 1, Dots: 1},
					NoteLengthMs{Quantity: 333},
				},
			},
			tempo:      120,
			expectedMs: 3333,
		},
		{
			duration: Duration{
				Components: []DurationComponent{
					NoteLength{Denominator: 4},
					Barline{},
					NoteLength{Denominator: 4},
				},
			},
			tempo:      60,
			expectedMs: 2000,
		},
	} {
		label := "Duration => milliseconds conversion"

		actualMs := testCase.duration.Ms(testCase.tempo)

		if !equalish(actualMs, testCase.expectedMs) {
			t.Error(label)
			t.Errorf(
				"Expected %#v at tempo %f to equal %f ms, got %f ms",
				testCase.duration, testCase.tempo, testCase.expectedMs, actualMs,
			)
		}
	}

	for _, testCase := range []noteDurationTestCase{
		{
			note: Note{
				Pitch: LetterAndAccidentals{NoteLetter: C},
				Duration: Duration{
					Components: []DurationComponent{NoteLength{Denominator: 4}},
				},
				Slurred: true,
			},
			tempo:        126,
			quantization: 42, // ignored because the note is slurred
			// The formula for beats => ms is:
			//
			//   beats * (60,000 / tempo in bpm)
			//
			// So in this case:
			//
			//   1 beat * (60,000 / 126 bpm) = 476.19ms
			expectedMs: 476.19,
		},
	} {
		label := "Note => audible duration"

		// This duplicates some of the business logic in model/note.go. I thought
		// about pulling it out into a helper function and unit test that function
		// here, but I couldn't quite figure out how to do that cleanly.
		durationMs := testCase.note.Duration.Ms(testCase.tempo)
		audibleDurationMs := durationMs
		if !testCase.note.Slurred {
			audibleDurationMs *= testCase.quantization
		}

		if !equalish(audibleDurationMs, testCase.expectedMs) {
			t.Error(label)
			t.Errorf(
				"Expected %#v at tempo %f to have an audible duration of %f ms, "+
					"got %f ms",
				testCase.note, testCase.tempo, testCase.expectedMs, audibleDurationMs,
			)
		}
	}
}
