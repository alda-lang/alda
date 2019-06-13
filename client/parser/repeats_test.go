package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func repeat(event model.ScoreUpdate, times int32) model.Repeat {
	return model.Repeat{Event: event, Times: times}
}

func onIteration(n int32) model.IterationRange {
	return model.IterationRange{First: n, Last: n}
}

func forIterationRange(first int32, last int32) model.IterationRange {
	return model.IterationRange{First: first, Last: last}
}

func eventOnIterations(
	event model.ScoreUpdate, ranges ...model.IterationRange,
) model.OnIterations {
	return model.OnIterations{Iterations: ranges, Event: event}
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
						model.Note{NoteLetter: model.C},
						model.Note{NoteLetter: model.D},
						model.Note{NoteLetter: model.E},
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
						model.Note{NoteLetter: model.C},
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
						model.Note{NoteLetter: model.C},
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
						NoteLetter: model.C,
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
						NoteLetter: model.C,
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
						NoteLetter: model.C,
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
			label: "repeat w/ iterations: [c'1,3]*3",
			given: "[c'1,3]*3",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnIterations(
							model.Note{NoteLetter: model.C},
							onIteration(1), onIteration(3),
						),
					),
					3,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ iterations: [c d'1 e'2]*2",
			given: "[c d'1 e'2]*2",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.C},
						eventOnIterations(model.Note{NoteLetter: model.D}, onIteration(1)),
						eventOnIterations(model.Note{NoteLetter: model.E}, onIteration(2)),
					),
					2,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ iterations: [c'1-2,4 [d e]'2-3]*4",
			given: "[c'1-2,4 [d e]'2-3]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnIterations(
							model.Note{NoteLetter: model.C},
							forIterationRange(1, 2), onIteration(4),
						),
						eventOnIterations(
							eventSequence(
								model.Note{NoteLetter: model.D},
								model.Note{NoteLetter: model.E},
							),
							forIterationRange(2, 3),
						),
					),
					4,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ iterations: [{c d e}2'1,3 [f r8 > g]'2-4]*4",
			given: "[{c d e}2'1,3 [f r8 > g]'2-4]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						eventOnIterations(
							model.Cram{
								Events: []model.ScoreUpdate{
									model.Note{NoteLetter: model.C},
									model.Note{NoteLetter: model.D},
									model.Note{NoteLetter: model.E},
								},
								Duration: model.Duration{
									Components: []model.DurationComponent{
										model.NoteLength{Denominator: 2},
									},
								},
							},
							onIteration(1), onIteration(3),
						),
						eventOnIterations(
							eventSequence(
								model.Note{NoteLetter: model.F},
								model.Rest{
									Duration: model.Duration{
										Components: []model.DurationComponent{
											model.NoteLength{Denominator: 8},
										},
									},
								},
								octaveUp(),
								model.Note{NoteLetter: model.G},
							),
							forIterationRange(2, 4),
						),
					),
					4,
				),
			},
		},
	)
}
