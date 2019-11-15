package model

import (
	"fmt"
	"testing"

	_ "alda.io/client/testing"
)

func TestVoices(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "voices",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},

				VoiceMarker{VoiceNumber: 2},
				Note{Pitch: LetterAndAccidentals{NoteLetter: B}},

				VoiceMarker{VoiceNumber: 3},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 0, 0),
				expectMidiNoteNumbers(67, 71, 74),
			},
		},
		scoreUpdateTestCase{
			label: "voices: note events have a reference to the original part",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},

				VoiceMarker{VoiceNumber: 2},
				Note{Pitch: LetterAndAccidentals{NoteLetter: B}},

				VoiceMarker{VoiceNumber: 3},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},

				VoiceGroupEndMarker{},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 1000, 1500, 1500, 1500, 2000),
				expectMidiNoteNumbers(60, 62, 64, 67, 71, 74, 76),
				func(score *Score) error {
					part := score.Events[0].(NoteEvent).Part

					for _, event := range score.Events[1:] {
						if event.(NoteEvent).Part != part {
							return fmt.Errorf(
								"Note events from different voices have different part " +
									"references",
							)
						}
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "V0: resumes from the last voice to end, offset-wise",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				// implicit quarter note (= 500ms)
				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},

				// four half notes (1000 * 4 = 4000ms)
				VoiceMarker{VoiceNumber: 2},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: B},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: A}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},

				// implicit quarter note (= 500ms)
				VoiceMarker{VoiceNumber: 3},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},

				// voice 2 should become "the" voice at the end of the voice group, so
				// this note should happen at offset 4000, not 500
				//
				// and it should be in octave 4, not 5
				// (voice 2 didn't go up an octave, voice 3 did)
				VoiceGroupEndMarker{},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(
					// voice 1
					0,
					// voice 2
					0, 1000, 2000, 3000,
					// voice 3
					0,
					// end of voice group (i.e. resume from voice 2)
					4000,
				),
				expectNoteDurations(
					// voice 1
					500,
					// voice 2
					1000, 1000, 1000, 1000,
					// voice 3
					500,
					// end of voice group (i.e. resume from voice 2)
					1000,
				),
				expectMidiNoteNumbers(
					// voice 1 (G4)
					67,
					// voice 2 (B4, A4, G4, F4)
					71, 69, 67, 65,
					// voice 3 (D5)
					74,
					// end of voice group (i.e. resume from voice 2)
					64,
				),
			},
		},
		scoreUpdateTestCase{
			label: "a part declaration implicitly ends a preceding voice group",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				// implicit quarter note (= 500ms)
				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},

				// four half notes (1000 * 4 = 4000ms)
				VoiceMarker{VoiceNumber: 2},
				Note{
					Pitch: LetterAndAccidentals{NoteLetter: B},
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},
				Note{Pitch: LetterAndAccidentals{NoteLetter: A}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},

				// implicit quarter note (= 500ms)
				VoiceMarker{VoiceNumber: 3},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},

				// this should implicitly end the voice group
				PartDeclaration{Names: []string{"piano"}},
				// voice 2 should become "the" voice at the end of the voice group, so
				// this note should happen at offset 4000, not 500
				//
				// and it should be in octave 4, not 5
				// (voice 2 didn't go up an octave, voice 3 did)
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectNoteOffsets(
					// voice 1
					0,
					// voice 2
					0, 1000, 2000, 3000,
					// voice 3
					0,
					// end of voice group (i.e. resume from voice 2)
					4000,
				),
				expectNoteDurations(
					// voice 1
					500,
					// voice 2
					1000, 1000, 1000, 1000,
					// voice 3
					500,
					// end of voice group (i.e. resume from voice 2)
					1000,
				),
				expectMidiNoteNumbers(
					// voice 1 (G4)
					67,
					// voice 2 (B4, A4, G4, F4)
					71, 69, 67, 65,
					// voice 3 (D5)
					74,
					// end of voice group (i.e. resume from voice 2)
					64,
				),
			},
		},
		scoreUpdateTestCase{
			label: "repeated calls to the same voice 1",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 1000, 1500),
				expectNoteDurations(500, 500, 500, 500),
				expectMidiNoteNumbers(60, 62, 64, 65),
			},
		},
		scoreUpdateTestCase{
			label: "repeated calls to the same voice 2",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 2,
					Event: EventSequence{
						Events: []ScoreUpdate{
							VoiceMarker{VoiceNumber: 1},
							Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
							Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 1000, 1500),
				expectNoteDurations(500, 500, 500, 500),
				expectMidiNoteNumbers(60, 62, 60, 62),
			},
		},
		scoreUpdateTestCase{
			label: "repeated calls to the same voice 3",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: D}},

				VoiceMarker{VoiceNumber: 2},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},

				VoiceMarker{VoiceNumber: 1},
				Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: F}},

				VoiceMarker{VoiceNumber: 2},
				Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: A}},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(
					// voice 1
					0, 500,
					// voice 2
					0, 500,
					// voice 1, continued
					1000, 1500,
					// voice 2, continued
					1000, 1500,
				),
				expectNoteDurations(
					// voice 1
					500, 500,
					// voice 2
					500, 500,
					// voice 1, continued
					500, 500,
					// voice 2, continued
					500, 500,
				),
				expectMidiNoteNumbers(
					// voice 1
					60, 62,
					// voice 2
					64, 65,
					// voice 1, continued
					64, 65,
					// voice 2, continued
					67, 69,
				),
			},
		},
		scoreUpdateTestCase{
			label: "voice containing a Cram expression",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},

				VoiceMarker{VoiceNumber: 1},
				Cram{
					Events: []ScoreUpdate{
						Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
						AttributeUpdate{PartUpdate: (OctaveDown{})},
						Note{Pitch: LetterAndAccidentals{NoteLetter: B}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: A}},
						Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectMidiNoteNumbers(60, 59, 57, 55),
				expectNoteOffsets(0, 125, 250, 375),
				expectNoteDurations(125, 125, 125, 125),
			},
		},
	)
}
