package model

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestRepeats(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "repeated note",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 8,
					Event: Note{
						Pitch: LetterAndAccidentals{NoteLetter: C},
						Duration: Duration{
							Components: []DurationComponent{
								NoteLength{Denominator: 8},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 250, 500, 750, 1000, 1250, 1500, 1750),
				expectMidiNoteNumbers(60, 60, 60, 60, 60, 60, 60, 60),
			},
		},
		scoreUpdateTestCase{
			label: "repeated chord",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 3,
					Event: Chord{
						Events: []ScoreUpdate{
							Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
							Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
							Note{Pitch: LetterAndAccidentals{NoteLetter: G}},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 0, 0, 500, 500, 500, 1000, 1000, 1000),
				expectMidiNoteNumbers(60, 64, 67, 60, 64, 67, 60, 64, 67),
			},
		},
		scoreUpdateTestCase{
			label: "repeated event sequence",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 3,
					Event: EventSequence{
						Events: []ScoreUpdate{
							Chord{
								Events: []ScoreUpdate{
									Note{
										Pitch: LetterAndAccidentals{NoteLetter: C},
										Duration: Duration{
											Components: []DurationComponent{
												NoteLength{Denominator: 4},
											},
										},
									},
									Note{
										Pitch: LetterAndAccidentals{NoteLetter: E},
									},
								},
							},
							Note{
								Pitch: LetterAndAccidentals{NoteLetter: F},
								Duration: Duration{
									Components: []DurationComponent{
										NoteLength{Denominator: 8},
									},
								},
							},
							Note{
								Pitch: LetterAndAccidentals{NoteLetter: G},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(
					0, 0, 500, 750, 1000, 1000, 1500, 1750, 2000, 2000, 2500, 2750,
				),
				expectMidiNoteNumbers(
					60, 64, 65, 67, 60, 64, 65, 67, 60, 64, 65, 67,
				),
			},
		},
		scoreUpdateTestCase{
			label: "OnRepetitions 1",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 2,
					Event: EventSequence{
						Events: []ScoreUpdate{
							Note{
								Pitch: LetterAndAccidentals{NoteLetter: C},
								Duration: Duration{
									Components: []DurationComponent{
										NoteLength{Denominator: 4},
									},
								},
							},
							OnRepetitions{
								Repetitions: []RepetitionRange{
									RepetitionRange{First: 1, Last: 1},
								},
								Event: Note{
									Pitch: LetterAndAccidentals{NoteLetter: D},
								},
							},
							OnRepetitions{
								Repetitions: []RepetitionRange{
									RepetitionRange{First: 2, Last: 2},
								},
								Event: Note{
									Pitch: LetterAndAccidentals{NoteLetter: E},
								},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 1000, 1500),
				expectMidiNoteNumbers(60, 62, 60, 64),
			},
		},
		scoreUpdateTestCase{
			label: "OnRepetitions 2",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Repeat{
					Times: 4,
					Event: EventSequence{
						Events: []ScoreUpdate{
							OnRepetitions{
								Repetitions: []RepetitionRange{
									RepetitionRange{First: 1, Last: 2},
									RepetitionRange{First: 4, Last: 4},
								},
								Event: Note{
									Pitch: LetterAndAccidentals{NoteLetter: C},
									Duration: Duration{
										Components: []DurationComponent{
											NoteLength{Denominator: 4},
										},
									},
								},
							},
							OnRepetitions{
								Repetitions: []RepetitionRange{
									RepetitionRange{First: 2, Last: 3},
								},
								Event: EventSequence{
									Events: []ScoreUpdate{
										Note{Pitch: LetterAndAccidentals{NoteLetter: D}},
										Note{Pitch: LetterAndAccidentals{NoteLetter: E}},
									},
								},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectNoteOffsets(0, 500, 1000, 1500, 2000, 2500, 3000),
				expectMidiNoteNumbers(60, 60, 62, 64, 62, 64, 60),
			},
		},
	)
}
