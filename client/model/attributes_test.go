package model

import (
	"fmt"
	"reflect"
	"testing"

	_ "alda.io/client/testing"
	"github.com/go-test/deep"
)

func expectPartIntValue(
	instrument string, valueName string, method func(p *Part) int32,
	expected int32) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if actual != expected {
			return fmt.Errorf(
				"%s %s is %d, not %d", instrument, valueName, actual, expected,
			)
		}

		return nil
	})
}

func expectPartFloatValue(
	instrument string, valueName string, method func(p *Part) float64,
	expected float64) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if !equalish(actual, expected) {
			return fmt.Errorf(
				"%s %s is %f, not %f", instrument, valueName, actual, expected,
			)
		}

		return nil
	})
}

func expectPartOffsetMsValue(
	instrument string, valueName string, method func(p *Part) float64,
	expected float64) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if !equalish(actual, expected) {
			return fmt.Errorf(
				"%s %s is %f, not %f", instrument, valueName, actual, expected,
			)
		}

		return nil
	})
}

func expectPartValueDeepEquals(
	instrument string, valueName string, method func(p *Part) interface{},
	expected interface{}) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := method(part)

		if !reflect.DeepEqual(actual, expected) {
			return fmt.Errorf(
				"%s %s is %#v, not %#v", instrument, valueName, actual, expected,
			)
		}

		return nil
	})
}

func expectPartOctave(instrument string, octave int32) func(s *Score) error {
	return expectPartIntValue(
		instrument, "octave", func(part *Part) int32 { return part.Octave }, octave,
	)
}

func expectPartVolume(instrument string, volume float64) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "volume", func(part *Part) float64 { return part.Volume },
		volume,
	)
}

func expectPartTrackVolume(
	instrument string, trackVolume float64,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "track volume",
		func(part *Part) float64 { return part.TrackVolume }, trackVolume,
	)
}

func expectPartPanning(
	instrument string, panning float64,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "panning", func(part *Part) float64 { return part.Panning },
		panning,
	)
}

func expectPartQuantization(
	instrument string, quantization float64,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "quantization",
		func(part *Part) float64 { return part.Quantization }, quantization,
	)
}

func expectPartTempo(instrument string, tempo float64) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "tempo", func(part *Part) float64 { return part.Tempo }, tempo,
	)
}

func expectPartKeySignature(
	instrument string, keySignature KeySignature,
) func(s *Score) error {
	return expectPartValueDeepEquals(
		instrument, "key signature",
		func(part *Part) interface{} { return part.KeySignature }, keySignature,
	)
}

func expectPartTransposition(
	instrument string, transposition int32,
) func(s *Score) error {
	return expectPartIntValue(
		instrument, "transposition",
		func(part *Part) int32 { return part.Transposition }, transposition,
	)
}

func expectPartReferencePitch(
	instrument string, frequency float64,
) func(s *Score) error {
	return expectPartFloatValue(
		instrument, "reference pitch",
		func(part *Part) float64 { return part.ReferencePitch }, frequency,
	)
}

func expectPartCurrentOffset(
	instrument string, expected float64,
) func(s *Score) error {
	return expectPartOffsetMsValue(
		instrument, "current offset",
		func(part *Part) float64 { return part.CurrentOffset }, expected,
	)
}

func expectPartLastOffset(
	instrument string, expected float64,
) func(s *Score) error {
	return expectPartOffsetMsValue(
		instrument, "last offset",
		func(part *Part) float64 { return part.LastOffset }, expected,
	)
}

func expectPartDurationBeats(
	instrument string, expected float64,
) func(s *Score) error {
	return expectPart(instrument, func(part *Part) error {
		actual := part.Duration.Beats()

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
	instrument string, expected float64,
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

func expectGlobalAttributeUpdates(
	offset float64, updates []PartUpdate,
) func(s *Score) error {
	return func(s *Score) error {
		updatesAtOffset, hit := s.GlobalAttributes.itinerary[offset]
		if !hit {
			return fmt.Errorf(
				"no global attribute updates recorded at offset %f", offset,
			)
		}

		if diff := deep.Equal(updates, updatesAtOffset); diff != nil {
			errStr := fmt.Sprintf(
				"global attributes at offset %f are not what we expected",
				offset,
			)

			for _, diffItem := range diff {
				errStr += fmt.Sprintf("\n%v", diffItem)
			}

			return fmt.Errorf(errStr)
		}

		return nil
	}
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
				expectPartTrackVolume("piano", 100.0/127),
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
					Pitch: LetterAndAccidentals{NoteLetter: C},
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
		scoreUpdateTestCase{
			label: "initial key signature",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				// The default key signature is empty, i.e. no NoteLetters have any
				// Accidentals. (i.e. C major / A minor)
				expectPartKeySignature("piano", map[NoteLetter][]Accidental{}),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: KeySignatureSet{
					KeySignature: KeySignature{F: {Sharp}, C: {Sharp}, G: {Sharp}}},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{F: {Sharp}, C: {Sharp}, G: {Sharp}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (string shorthand 1)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispString{Value: "b- e- a- d-"},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}, E: {Flat}, A: {Flat}, D: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (string shorthand 2)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispString{Value: "f+ c+ g+"},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{F: {Sharp}, C: {Sharp}, G: {Sharp}},
				),
			},
		},
		scoreUpdateTestCase{
			// (key-signature '(g major))
			label: "set key signature via lisp (name of scale 1a)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "g"}, LispSymbol{Name: "major"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{F: {Sharp}},
				),
			},
		},
		scoreUpdateTestCase{
			// (key-signature (quote (g major)))
			label: "set key signature via lisp (name of scale 1b)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispList{
						Elements: []LispForm{
							LispSymbol{Name: "quote"},
							LispList{
								Elements: []LispForm{
									LispSymbol{Name: "g"}, LispSymbol{Name: "major"},
								},
							},
						},
					},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{F: {Sharp}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 2)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "b"},
						LispSymbol{Name: "flat"},
						LispSymbol{Name: "major"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}, E: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 3)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "ionian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 4)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "dorian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}, E: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 5)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "phrygian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}, E: {Flat}, A: {Flat}, D: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 6)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "lydian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{F: {Sharp}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 7)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "mixolydian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 8)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "aeolian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat}, E: {Flat}, A: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (name of scale 9)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "c"},
						LispSymbol{Name: "locrian"},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano",
					KeySignature{B: {Flat}, E: {Flat}, A: {Flat}, D: {Flat}, G: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "set key signature via lisp (list of letter/accidentals pairs)",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "key-signature"},
					LispQuotedForm{Form: LispList{Elements: []LispForm{
						LispSymbol{Name: "e"},
						LispList{Elements: []LispForm{LispSymbol{Name: "flat"}}},
						LispSymbol{Name: "b"},
						LispList{Elements: []LispForm{
							LispSymbol{Name: "flat"}, LispSymbol{Name: "flat"},
						}},
					},
					}},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartKeySignature(
					"piano", KeySignature{B: {Flat, Flat}, E: {Flat}},
				),
			},
		},
		scoreUpdateTestCase{
			label: "initial transposition",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTransposition("piano", 0),
			},
		},
		scoreUpdateTestCase{
			label: "set transposition",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: TranspositionSet{Semitones: 8}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTransposition("piano", 8),
			},
		},
		scoreUpdateTestCase{
			label: "set transposition using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "transpose"},
					LispNumber{Value: 82},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTransposition("piano", 82),
			},
		},
		scoreUpdateTestCase{
			label: "initial reference pitch",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartReferencePitch("piano", 440.0),
			},
		},
		scoreUpdateTestCase{
			label: "set reference pitch",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AttributeUpdate{PartUpdate: ReferencePitchSet{Frequency: 432.1}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartReferencePitch("piano", 432.1),
			},
		},
		scoreUpdateTestCase{
			label: "set reference pitch using lisp",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "reference-pitch"},
					LispNumber{Value: 550.0},
				}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartReferencePitch("piano", 550.0),
			},
		},
		scoreUpdateTestCase{
			label: "part-specific attribute updates aren't interpreted as global",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-duration"},
					LispNumber{Value: 4},
				}},

				PartDeclaration{Names: []string{"harp"}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "set-duration"},
					LispNumber{Value: 2},
				}},
				LispList{Elements: []LispForm{
					LispSymbol{Name: "octave"},
					LispNumber{Value: 3},
				}},

				PartDeclaration{Names: []string{"piano", "harp"}},
				Note{Pitch: LetterAndAccidentals{NoteLetter: C}},
			},
			expectations: []scoreUpdateExpectation{
				expectPartOctave("piano", 4),
				expectPartDurationBeats("piano", 4),
				expectPartOctave("harp", 3),
				expectPartDurationBeats("harp", 2),
			},
		},
		scoreUpdateTestCase{
			label: "a global tempo attribute update",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// A whole note rest at 120 bpm = 2000 ms
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
						},
					},
				},
				// Globally set tempo to 60 bpm, 2000 ms into the score.
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo!"},
					LispNumber{Value: 60},
				}},

				// The viola part should start off having a tempo of 120, not 60
				PartDeclaration{Names: []string{"viola"}},
				// After resting for 3 beats (still before the global tempo update), the
				// viola's tempo should still be 120 bpm.
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectGlobalAttributeUpdates(2000, []PartUpdate{TempoSet{Tempo: 60}}),
				expectPartTempo("piano", 60),
				expectPartTempo("viola", 120),
			},
		},
		scoreUpdateTestCase{
			label: "global updates are applied once you cross the offset",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// A whole note rest at 120 bpm = 2000 ms
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
						},
					},
				},
				// Globally set tempo to 60 bpm, 2000 ms into the score.
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo!"},
					LispNumber{Value: 60},
				}},

				// The viola part should start off having a tempo of 120, not 60
				PartDeclaration{Names: []string{"viola"}},
				// After resting for 3 beats (still before the global tempo update), the
				// viola's tempo should still be 120 bpm.
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2, Dots: 1},
						},
					},
				},
				// After resting for 1 more beat, we've arrived at the 2000ms mark, so
				// the viola's tempo should now be 60 bpm.
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("viola", 60),
			},
		},
		scoreUpdateTestCase{
			label: "global updates are applied once you cross the offset via @marker",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// A whole note rest at 120 bpm = 2000 ms
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
						},
					},
				},
				Marker{Name: "two-thousand"},
				// Globally set tempo to 60 bpm, 2000 ms into the score.
				LispList{Elements: []LispForm{
					LispSymbol{Name: "tempo!"},
					LispNumber{Value: 60},
				}},

				// The viola part should start off having a tempo of 120, not 60
				PartDeclaration{Names: []string{"viola"}},
				// After jumping to the marker, we've passed the 2000ms mark, so the
				// viola's tempo should now be 60 bpm.
				AtMarker{Name: "two-thousand"},
			},
			expectations: []scoreUpdateExpectation{
				expectPartTempo("viola", 60),
			},
		},
	)
}
