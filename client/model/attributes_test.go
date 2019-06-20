package model

import (
	"fmt"
	"testing"

	_ "alda.io/client/testing"
)

func expectPartIntValue(
	instrument string, valueName string, method func(p *Part) int32,
	expected int32) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if actual != expected {
			return fmt.Errorf("%s is %d, not %d", valueName, actual, expected)
		}

		return nil
	})
}

func expectPartFloatValue(
	instrument string, valueName string, method func(p *Part) float32,
	expected float32) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if actual != expected {
			return fmt.Errorf("%s is %f, not %f", valueName, actual, expected)
		}

		return nil
	})
}

func expectPartOctave(instrument string, octave int32) func(s *Score) error {
	return expectPartIntValue(
		instrument, "octave", func(part *Part) int32 { return part.Octave }, octave,
	)
}

func expectPartVolume(instrument string, volume float32) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "volume", func(part *Part) float32 { return part.Volume },
		volume,
	)
}

func expectPartTrackVolume(
	instrument string, trackVolume float32,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "track volume",
		func(part *Part) float32 { return part.TrackVolume }, trackVolume,
	)
}

func expectPartPanning(
	instrument string, panning float32,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "panning", func(part *Part) float32 { return part.Panning },
		panning,
	)
}

func expectPartQuantization(
	instrument string, quantization float32,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "quantization",
		func(part *Part) float32 { return part.Quantization }, quantization,
	)
}

func expectPartTempo(instrument string, tempo float32) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "tempo", func(part *Part) float32 { return part.Tempo }, tempo,
	)
}

func expectPartDurationBeats(
	instrument string, expected float32,
) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual, err := part.Duration.Beats()
		if err != nil {
			return err
		}

		if actual != expected {
			return fmt.Errorf(
				"expected duration to be %f beat(s), got %f beats (%#v)",
				expected, actual, part.Duration,
			)
		}

		return nil
	})
}

func expectPartDurationMs(
	instrument string, expected float32,
) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := part.Duration.Ms(part.Tempo)

		if actual != expected {
			return fmt.Errorf(
				"expected duration to be %f ms, got %f ms (%#v)",
				expected, actual, part.Duration,
			)
		}

		return nil
	})
}

func TestAttributes(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "initial octave",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 4),
			},
		},
		scoreUpdateTestCase{
			label: "set octave",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 2}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 2),
			},
		},
		scoreUpdateTestCase{
			label: "set octave using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "octave"},
					LispNumber{Value: 5},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 5),
			},
		},
		scoreUpdateTestCase{
			label: "decrement octave",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 2}},
				AttributeUpdate{PartUpdate: OctaveDown{}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 1),
			},
		},
		scoreUpdateTestCase{
			label: "decrement octave using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 2}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "octave"},
					LispQuotedForm{Form: LispSymbol{Name: "down"}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 1),
			},
		},
		scoreUpdateTestCase{
			label: "increment octave",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 2}},
				AttributeUpdate{PartUpdate: OctaveUp{}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 3),
			},
		},
		scoreUpdateTestCase{
			label: "increment octave using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 2}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "octave"},
					LispQuotedForm{Form: LispSymbol{Name: "up"}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 3),
			},
		},
		scoreUpdateTestCase{
			label: "several octave operations in a row",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: OctaveSet{OctaveNumber: 4}},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				AttributeUpdate{PartUpdate: OctaveUp{}},
				AttributeUpdate{PartUpdate: OctaveDown{}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 6),
			},
		},
		scoreUpdateTestCase{
			label: "initial volume",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartVolume("piano", 1.0),
			},
		},
		scoreUpdateTestCase{
			label: "set volume",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: VolumeSet{Volume: 0.85}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartVolume("piano", 0.85),
			},
		},
		scoreUpdateTestCase{
			label: "set volume using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "volume"},
					LispNumber{Value: 82},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartVolume("piano", 0.82),
			},
		},
		scoreUpdateTestCase{
			label: "initial track volume",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTrackVolume("piano", float32(100.0/127)),
			},
		},
		scoreUpdateTestCase{
			label: "set track volume",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: TrackVolumeSet{TrackVolume: 0.85}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTrackVolume("piano", 0.85),
			},
		},
		scoreUpdateTestCase{
			label: "set track volume using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "track-volume"},
					LispNumber{Value: 82},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTrackVolume("piano", 0.82),
			},
		},
		scoreUpdateTestCase{
			label: "initial panning",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartPanning("piano", 0.5),
			},
		},
		scoreUpdateTestCase{
			label: "set panning",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: PanningSet{Panning: 0.85}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartPanning("piano", 0.85),
			},
		},
		scoreUpdateTestCase{
			label: "set panning using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "panning"},
					LispNumber{Value: 82},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartPanning("piano", 0.82),
			},
		},
		scoreUpdateTestCase{
			label: "initial quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartQuantization("piano", 0.9),
			},
		},
		scoreUpdateTestCase{
			label: "set quantization",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: QuantizationSet{Quantization: 0.85}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartQuantization("piano", 0.85),
			},
		},
		scoreUpdateTestCase{
			label: "set quantization using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 82},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartQuantization("piano", 0.82),
			},
		},
		scoreUpdateTestCase{
			label: "set quantization using lisp: value > 100",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "quant"},
					LispNumber{Value: 9001},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartQuantization("piano", 90.01),
			},
		},
		scoreUpdateTestCase{
			label: "initial duration",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				// Default note length is a quarter note (1 beat).
				expectPartDurationBeats("piano", 1),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-duration`)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-duration"},
					LispNumber{Value: 3.7},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 3.7),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-duration-ms`)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-duration-ms"},
					LispNumber{Value: 2345},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationMs("piano", 2345),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-note-length`, number)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-note-length"},
					LispNumber{Value: 1},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 4),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-note-length`, string 1)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-note-length"},
					LispString{Value: "2.."},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 3.5),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-note-length`, string 2)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-note-length"},
					LispString{Value: "0.5.."},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 14),
			},
		},
		scoreUpdateTestCase{
			label: "set duration via lisp (`set-note-length`, string 3)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-note-length"},
					LispString{Value: "1~1"},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 8),
			},
		},
		scoreUpdateTestCase{
			label: "a note's duration implicitly changes the part's duration",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
							NoteLength{Denominator: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectPartDurationBeats("piano", 8),
			},
		},
		scoreUpdateTestCase{
			label: "initial tempo",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 120),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: TempoSet{Tempo: 60}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 60},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp: half note = 30",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 2},
					LispNumber{Value: 30},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp: dotted quarter note = 40",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispString{Value: "4."},
					LispNumber{Value: 40},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp: (complicated way to say a half note) = 30",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispString{Value: "8.~16~4"},
					LispNumber{Value: 30},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp: whole note = 15",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispString{Value: "1"},
					LispNumber{Value: 15},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "set tempo via lisp: breve = 7.5",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispString{Value: "0.5"},
					LispNumber{Value: 7.5},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 60),
			},
		},
		scoreUpdateTestCase{
			label: "metric modulation: dotted quarter = half",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 120},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "metric-modulation"},
					LispString{Value: "4."},
					LispNumber{Value: 2},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 160),
			},
		},
		scoreUpdateTestCase{
			label: "metric modulation: half = dotted quarter",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 160},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "metric-modulation"},
					LispNumber{Value: 2},
					LispString{Value: "4."},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 120),
			},
		},
		scoreUpdateTestCase{
			label: "metric modulation: half = dotted quarter (both strings)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 160},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "metric-modulation"},
					LispString{Value: "2"},
					LispString{Value: "4."},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 120),
			},
		},
		scoreUpdateTestCase{
			label: "metric modulation: quarter = eighth",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo"},
					LispNumber{Value: 60},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "metric-modulation"},
					LispNumber{Value: 4},
					LispNumber{Value: 8},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("piano", 30),
			},
		},
	)
}
