package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func repeat(event model.ScoreUpdate, times int32) model.Repeat {
	return model.Repeat{Event: event, Times: times}
}

func onRepetition(n int32) model.RepetitionRange {
	return model.RepetitionRange{First: n, Last: n}
}

func forRepetitionRange(first int32, last int32) model.RepetitionRange {
	return model.RepetitionRange{First: first, Last: last}
}

func eventOnRepetitions(
	event model.ScoreUpdate, ranges ...model.RepetitionRange,
) model.OnRepetitions {
	return model.OnRepetitions{Repetitions: ranges, Event: event}
}

func TestRepeats(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "repeat event seq w/ 3 notes",
			given: "[c d e] *4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.D}},
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.E}},
					),
					4,
				),
			},
		},
		parseTestCase{
			label: "repeat event seq w/ a note and an octave-up",
			given: "[ c > ]*5",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						octaveUp(),
					),
					5,
				),
			},
		},
		parseTestCase{
			label: "repeat event seq w/ a note and an octave-up (more whitespace)",
			given: "[ c > ] * 5",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						octaveUp(),
					),
					5,
				),
			},
		},
		parseTestCase{
			label: "repeat note w/ explicit duration",
			given: "c8*7",
			expect: []model.ScoreUpdate{
				repeat(
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 8},
							},
						},
					},
					7,
				),
			},
		},
		parseTestCase{
			label: "repeat note w/ explicit duration (more whitespace)",
			given: "c8 *7",
			expect: []model.ScoreUpdate{
				repeat(
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 8},
							},
						},
					},
					7,
				),
			},
		},
		parseTestCase{
			label: "repeat note w/ explicit duration (even more whitespace)",
			given: "c8 * 7",
			expect: []model.ScoreUpdate{
				repeat(
					model.Note{
						Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
						Duration: model.Duration{
							Components: []model.DurationComponent{
								model.NoteLength{Denominator: 8},
							},
						},
					},
					7,
				),
			},
		},
		parseTestCase{
			label: "repeated event sequence containing repeated note",
			given: "[c*2]*2",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						repeat(
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
							},
							2,
						),
					),
					2,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ repetitions: [c'1,3]*3",
			given: "[c'1,3]*3",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnRepetitions(
							model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
							onRepetition(1), onRepetition(3),
						),
					),
					3,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ repetitions: [c d'1 e'2]*2",
			given: "[c d'1 e'2]*2",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{Pitch: model.LetterAndAccidentals{NoteLetter: model.C}},
						eventOnRepetitions(
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.D},
							},
							onRepetition(1),
						),
						eventOnRepetitions(
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
							},
							onRepetition(2),
						),
					),
					2,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ repetitions: [c'1-2,4 [d e]'2-3]*4",
			given: "[c'1-2,4 [d e]'2-3]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnRepetitions(
							model.Note{
								Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
							},
							forRepetitionRange(1, 2), onRepetition(4),
						),
						eventOnRepetitions(
							eventSequence(
								model.Note{
									Pitch: model.LetterAndAccidentals{NoteLetter: model.D},
								},
								model.Note{
									Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
								},
							),
							forRepetitionRange(2, 3),
						),
					),
					4,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ repetitions: [{c d e}2'1,3 [f r8 > g]'2-4]*4",
			given: "[{c d e}2'1,3 [f r8 > g]'2-4]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnRepetitions(
							model.Cram{
								Events: []model.ScoreUpdate{
									model.Note{
										Pitch: model.LetterAndAccidentals{NoteLetter: model.C},
									},
									model.Note{
										Pitch: model.LetterAndAccidentals{NoteLetter: model.D},
									},
									model.Note{
										Pitch: model.LetterAndAccidentals{NoteLetter: model.E},
									},
								},
								Duration: model.Duration{
									Components: []model.DurationComponent{
										model.NoteLength{Denominator: 2},
									},
								},
							},
							onRepetition(1), onRepetition(3),
						),
						eventOnRepetitions(
							eventSequence(
								model.Note{
									Pitch: model.LetterAndAccidentals{NoteLetter: model.F},
								},
								model.Rest{
									Duration: model.Duration{
										Components: []model.DurationComponent{
											model.NoteLength{Denominator: 8},
										},
									},
								},
								octaveUp(),
								model.Note{
									Pitch: model.LetterAndAccidentals{NoteLetter: model.G},
								},
							),
							forRepetitionRange(2, 4),
						),
					),
					4,
				),
			},
		},
	)
}
