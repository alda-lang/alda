package code_generator

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"testing"
)

func eof() parser.Token {
	return parser.Token{TokenType: parser.EOF, Text: ""}
}

func TestGenerator(t *testing.T) {
	executeGeneratorTestCases(t, generatorTestCase{
		label: "attribute updates",
		updates: []model.ScoreUpdate{
			model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
			model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
			model.AttributeUpdate{PartUpdate: model.OctaveSet{OctaveNumber: 1}},
			model.AttributeUpdate{PartUpdate: model.OctaveSet{OctaveNumber: 3}},
		},
		expected: []parser.Token{
			{TokenType: parser.OctaveUp, Text: ">"},
			{TokenType: parser.OctaveDown, Text: "<"},
			{TokenType: parser.OctaveSet, Text: "o1"},
			{TokenType: parser.OctaveSet, Text: "o3"},
			eof(),
		},
	}, generatorTestCase{
		label: "barlines",
		updates: []model.ScoreUpdate{
			model.Barline{},
		},
		expected: []parser.Token{
			{TokenType: parser.Barline, Text: "|"},
			eof(),
		},
	}, generatorTestCase{
		label: "chords",
		updates: []model.ScoreUpdate{
			model.Chord{Events: []model.ScoreUpdate{
				model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
				model.Barline{},
				model.Rest{},
				model.Barline{},
				model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
				model.Rest{},
				model.Rest{},
				model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
			}},
		},
		expected: []parser.Token{
			{TokenType: parser.OctaveDown, Text: "<"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.OctaveDown, Text: "<"},
			{TokenType: parser.Separator, Text: "/"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.Separator, Text: "/"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.OctaveUp, Text: ">"},
			eof(),
		},
	}, generatorTestCase{
		label: "cram",
		updates: []model.ScoreUpdate{
			model.Cram{Duration: model.Duration{
				Components: []model.DurationComponent{
					model.NoteLength{Denominator: 4},
				},
			}, Events: []model.ScoreUpdate{
				model.Barline{},
			}},
		},
		expected: []parser.Token{
			{TokenType: parser.CramOpen, Text: "{"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.CramClose, Text: "}"},
			{TokenType: parser.NoteLength, Text: "4"},
			eof(),
		},
	}, generatorTestCase{
		label: "event sequence",
		updates: []model.ScoreUpdate{
			model.EventSequence{Events: []model.ScoreUpdate{
				model.Barline{},
			}},
		},
		expected: []parser.Token{
			{TokenType: parser.EventSeqOpen, Text: "["},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.EventSeqClose, Text: "]"},
			eof(),
		},
	}, generatorTestCase{
		label: "markers",
		updates: []model.ScoreUpdate{
			model.Marker{Name: "marker1"},
			model.Marker{Name: "marker2"},
			model.AtMarker{Name: "marker2"},
			model.AtMarker{Name: "marker1"},
		},
		expected: []parser.Token{
			{TokenType: parser.Marker, Text: "%marker1"},
			{TokenType: parser.Marker, Text: "%marker2"},
			{TokenType: parser.AtMarker, Text: "@marker2"},
			{TokenType: parser.AtMarker, Text: "@marker1"},
			eof(),
		},
	}, generatorTestCase{
		label: "note pitches",
		updates: []model.ScoreUpdate{
			model.Note{Pitch: model.LetterAndAccidentals{
				NoteLetter: model.C,
				Accidentals: []model.Accidental{model.Sharp, model.Flat},
			}},
			model.Note{Pitch: model.LetterAndAccidentals{
				NoteLetter: model.G,
				Accidentals: []model.Accidental{model.Natural},
			}},
			model.Note{Pitch: model.LetterAndAccidentals{
				NoteLetter: model.F,
			}},
		},
		expected: []parser.Token{
			{TokenType: parser.NoteLetter, Text: "c"},
			{TokenType: parser.Sharp, Text: "+"},
			{TokenType: parser.Flat, Text: "-"},
			{TokenType: parser.NoteLetter, Text: "g"},
			{TokenType: parser.Natural, Text: "_"},
			{TokenType: parser.NoteLetter, Text: "f"},
			eof(),
		},
	}, generatorTestCase{
		label: "rests and durations",
		updates: []model.ScoreUpdate{
			model.Rest{},
			model.Rest{Duration: model.Duration{
				Components: []model.DurationComponent{
					model.NoteLength{Denominator: 4},
				}},
			},
			model.Rest{Duration: model.Duration{
				Components: []model.DurationComponent{
					model.NoteLengthMs{Quantity: 200},
				}},
			},
			model.Rest{Duration: model.Duration{
				Components: []model.DurationComponent{
					model.NoteLength{Denominator: 2, Dots: 3},
					model.NoteLengthMs{Quantity: 100.12345},
					model.Barline{},
					model.NoteLength{Denominator: 3.123},
					model.Barline{},
				}},
			},
		},
		expected: []parser.Token{
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.NoteLength, Text: "4"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.NoteLengthMs, Text: "200ms"},
			{TokenType: parser.RestLetter, Text: "r"},
			{TokenType: parser.NoteLength, Text: "2..."},
			{TokenType: parser.Tie, Text: "~"},
			{TokenType: parser.NoteLengthMs, Text: "100.12345ms"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.Tie, Text: "~"},
			{TokenType: parser.NoteLength, Text: "3.123"},
			{TokenType: parser.Barline, Text: "|"},
			eof(),
		},
	}, generatorTestCase{
		label: "part declarations",
		updates: []model.ScoreUpdate{
			model.PartDeclaration{
				Names: []string{"name1"},
			},
			model.PartDeclaration{
				Names: []string{"name2", "name3"},
				Alias: "part1",
			},
		},
		expected: []parser.Token{
			{TokenType: parser.Name, Text: "name1"},
			{TokenType: parser.Colon, Text: ":"},
			{TokenType: parser.Name, Text: "name2"},
			{TokenType: parser.Separator, Text: "/"},
			{TokenType: parser.Name, Text: "name3"},
			{TokenType: parser.Alias, Text: "\"part1\""},
			{TokenType: parser.Colon, Text: ":"},
			eof(),
		},
	}, generatorTestCase{
		label: "repeats",
		updates: []model.ScoreUpdate{
			model.Repeat{
				Event: model.EventSequence{Events: []model.ScoreUpdate{
					model.Barline{},
				}},
				Times: 2,
			},
		},
		expected: []parser.Token{
			{TokenType: parser.EventSeqOpen, Text: "["},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.EventSeqClose, Text: "]"},
			{TokenType: parser.Repeat, Text: "*2"},
			eof(),
		},
	}, generatorTestCase{
		label: "on repetitions",
		updates: []model.ScoreUpdate{
			model.OnRepetitions{
				Event: model.EventSequence{Events: []model.ScoreUpdate{
					model.Barline{},
				}},
				Repetitions: []model.RepetitionRange{
					{First: 1, Last: 1},
				},
			},
			model.OnRepetitions{
				Event: model.EventSequence{Events: []model.ScoreUpdate{
					model.Barline{},
				}},
				Repetitions: []model.RepetitionRange{
					{First: 1, Last: 2},
				},
			},
			model.OnRepetitions{
				Event: model.EventSequence{Events: []model.ScoreUpdate{
					model.Barline{},
				}},
				Repetitions: []model.RepetitionRange{
					{First: 1, Last: 2},
					{First: 3, Last: 3},
					{First: 5, Last: 7},
				},
			},
		},
		expected: []parser.Token{
			{TokenType: parser.EventSeqOpen, Text: "["},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.EventSeqClose, Text: "]"},
			{TokenType: parser.Repetitions, Text: "'1"},
			{TokenType: parser.EventSeqOpen, Text: "["},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.EventSeqClose, Text: "]"},
			{TokenType: parser.Repetitions, Text: "'1-2"},
			{TokenType: parser.EventSeqOpen, Text: "["},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.EventSeqClose, Text: "]"},
			{TokenType: parser.Repetitions, Text: "'1-2,3,5-7"},
			eof(),
		},
	}, generatorTestCase{
		label: "variables",
		updates: []model.ScoreUpdate{
			model.VariableDefinition{
				VariableName: "var1",
				Events: []model.ScoreUpdate{model.Barline{}},
			},
			model.VariableReference{VariableName: "var1"},
		},
		expected: []parser.Token{
			{TokenType: parser.Name, Text: "var1"},
			{TokenType: parser.Equals, Text: "="},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.Name, Text: "var1"},
			eof(),
		},
	}, generatorTestCase{
		label: "voices",
		updates: []model.ScoreUpdate{
			model.VoiceMarker{VoiceNumber: 1},
			model.Barline{},
			model.VoiceMarker{VoiceNumber: 2},
			model.Barline{},
			model.VoiceGroupEndMarker{},
			model.Barline{},
		},
		expected: []parser.Token{
			{TokenType: parser.VoiceMarker, Text: "V1:"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.VoiceMarker, Text: "V2:"},
			{TokenType: parser.Barline, Text: "|"},
			{TokenType: parser.VoiceMarker, Text: "V0:"},
			{TokenType: parser.Barline, Text: "|"},
			eof(),
		},
	})
}
