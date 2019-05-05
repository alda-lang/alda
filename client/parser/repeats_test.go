package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func repeat(event model.ScoreUpdate, times int32) model.Repeat {
	return model.Repeat{Event: event, Times: times}
}

func endings(ranges ...model.EndingRange) model.Endings {
	return model.Endings{Ranges: ranges}
}

func onEnding(n int32) model.EndingRange {
	return model.EndingRange{First: n, Last: n}
}

func forEndingRange(first int32, last int32) model.EndingRange {
	return model.EndingRange{First: first, Last: last}
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
						model.OctaveUp{},
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
						model.OctaveUp{},
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
			label: "repeat w/ endings: [c'1,3]*3",
			given: "[c'1,3]*3",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.C},
						endings(onEnding(1), onEnding(3)),
					),
					3,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ endings: [c d'1 e'2]*2",
			given: "[c d'1 e'2]*2",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.C},
						model.Note{NoteLetter: model.D},
						endings(onEnding(1)),
						model.Note{NoteLetter: model.E},
						endings(onEnding(2)),
					),
					2,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ endings: [c'1-2,4 [d e]'2-3]*4",
			given: "[c'1-2,4 [d e]'2-3]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
						model.Note{NoteLetter: model.C},
						endings(forEndingRange(1, 2), onEnding(4)),
						eventSequence(
							model.Note{NoteLetter: model.D},
							model.Note{NoteLetter: model.E},
						),
						endings(forEndingRange(2, 3)),
					),
					4,
				),
			},
		},
		parseTestCase{
			label: "repeat w/ endings: [{c d e}2'1,3 [f r8 > g]'2-4]*4",
			given: "[{c d e}2'1,3 [f r8 > g]'2-4]*4",
			expect: []model.ScoreUpdate{
				repeat(
					eventSequence(
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
						endings(onEnding(1), onEnding(3)),
						eventSequence(
							model.Note{NoteLetter: model.F},
							model.Rest{
								Duration: model.Duration{
									Components: []model.DurationComponent{
										model.NoteLength{Denominator: 8},
									},
								},
							},
							model.OctaveUp{},
							model.Note{NoteLetter: model.G},
						),
						endings(forEndingRange(2, 4)),
					),
					4,
				),
			},
		},
	)
}
