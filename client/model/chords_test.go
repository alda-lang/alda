package model

import (
	"fmt"
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func TestChords(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "chord",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// Add 1000ms to the initial offset of the chord in order to eliminate
				// false positives from offset initializing to 0
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},
				Chord{
					Events: []ScoreUpdate{
						Note{
							NoteLetter: E,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 1},
								},
							},
						},
						Note{
							NoteLetter: G,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 4},
								},
							},
						},
						Rest{
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 8},
								},
							},
						},
						AttributeUpdate{PartUpdate: OctaveUp{}},
						Note{
							NoteLetter: G,
							Duration: Duration{
								Components: []DurationComponent{
									NoteLength{Denominator: 2},
								},
							},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				func(s *Score) error {
					// The three notes of the chord, plus the one note we added at the
					// beginning to reduce false positives.
					expectedEvents := 4
					if len(s.Events) != expectedEvents {
						return fmt.Errorf(
							"there are %d events, not %d", len(s.Events), expectedEvents,
						)
					}

					piano, err := getPart(s, "piano")
					if err != nil {
						return err
					}

					// initial 1000ms buffer + 250, which is the shortest note/rest
					// duration in the chord (the 8th note rest == 250 ms)
					expectedCurrentOffset := 1250.0

					if piano.CurrentOffset != expectedCurrentOffset {
						return fmt.Errorf(
							"piano part's current offset is %f, not %f",
							piano.CurrentOffset,
							expectedCurrentOffset,
						)
					}

					expectedLastOffset := 1000.0

					if piano.LastOffset != expectedLastOffset {
						return fmt.Errorf(
							"piano part's last offset is %f, not %f",
							piano.LastOffset,
							expectedLastOffset,
						)
					}

					for i, event := range s.Events {
						expectedOffset := 0.0
						if i != 0 {
							expectedOffset = 1000
						}

						actualOffset := event.(NoteEvent).Offset
						if actualOffset != expectedOffset {
							return fmt.Errorf(
								"expected note #%d to have offset %f, but it was %f",
								i+1, expectedOffset, actualOffset,
							)
						}
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "accessing a part in a group using the dot operator",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "trumpet"},
					Alias: "trumpiano",
				},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 1234},
						},
					},
				},
				PartDeclaration{Names: []string{"trumpiano.trumpet"}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 1000},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "trumpet"),
				expectCurrentParts("trumpet"),
				expectPart("trumpet", func(part *Part) error {
					if part.CurrentOffset != 2234 {
						return fmt.Errorf(
							"expected trumpet part's offset to be 2234, but it was %f",
							part.CurrentOffset,
						)
					}

					return nil
				}),
			},
		},
		scoreUpdateTestCase{
			label: "referring to an alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "trumpet"},
					Alias: "trumpiano",
				},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 1234},
						},
					},
				},
				PartDeclaration{Names: []string{"bassoon"}},
				Note{
					NoteLetter: C,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLengthMs{Quantity: 1234},
						},
					},
				},
				PartDeclaration{Names: []string{"trumpiano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("trumpet", "piano", "bassoon"),
				expectCurrentParts("trumpet", "piano"),
			},
		},
		scoreUpdateTestCase{
			label: "referring to an existing unnamed part",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectCurrentParts("piano"),
			},
		},
		scoreUpdateTestCase{
			label: "referring to an existing named part by its alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "piano-1",
				},
				PartDeclaration{Names: []string{"piano-1"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano"),
				expectCurrentParts("piano"),
			},
		},
		scoreUpdateTestCase{
			label: "referring to an existing named group by its alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "guitar"},
					Alias: "foos",
				},
				PartDeclaration{Names: []string{"foos"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "guitar"),
				expectCurrentParts("piano", "guitar"),
			},
		},
		scoreUpdateTestCase{
			label: "group consisting of multiple parts",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano", "clarinet", "flute"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "flute"),
				expectCurrentParts("piano", "clarinet", "flute"),
			},
		},
		scoreUpdateTestCase{
			label: "group consisting of multiple aliased parts",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"clarinet"},
					Alias: "bar",
				},
				PartDeclaration{
					Names: []string{"flute"},
					Alias: "baz",
				},
				PartDeclaration{Names: []string{"foo", "bar", "baz"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "flute"),
				expectCurrentParts("piano", "clarinet", "flute"),
			},
		},
		scoreUpdateTestCase{
			label: "aliased group consisting of multiple aliased parts",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"clarinet"},
					Alias: "bar",
				},
				PartDeclaration{
					Names: []string{"flute"},
					Alias: "baz",
				},
				PartDeclaration{
					Names: []string{"foo", "bar", "baz"},
					Alias: "quux",
				},
				PartDeclaration{Names: []string{"quux"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "flute"),
				expectCurrentParts("piano", "clarinet", "flute"),
			},
		},
		scoreUpdateTestCase{
			label: "aliased group consisting of multiple aliased parts: dot access",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"clarinet"},
					Alias: "bar",
				},
				PartDeclaration{
					Names: []string{"flute"},
					Alias: "baz",
				},
				PartDeclaration{
					Names: []string{"foo", "bar", "baz"},
					Alias: "quux",
				},
				PartDeclaration{Names: []string{"quux.bar"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "flute"),
				expectCurrentParts("clarinet"),
			},
		},
		scoreUpdateTestCase{
			label: "group of stock instruments where some already exist",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				PartDeclaration{
					Names: []string{"clarinet"},
				},
				PartDeclaration{
					Names: []string{"piano", "clarinet", "flute"},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "flute"),
				expectCurrentParts("piano", "clarinet", "flute"),
			},
		},
		scoreUpdateTestCase{
			label: "named group of stock instruments where some already exist",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				PartDeclaration{
					Names: []string{"clarinet"},
				},
				PartDeclaration{
					Names: []string{"piano", "clarinet", "flute"},
					Alias: "floop",
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "piano", "clarinet", "flute"),
				expectCurrentParts("piano", "clarinet", "flute"),
			},
		},
		scoreUpdateTestCase{
			label: "named group of stock instruments where some already exist: dot " +
				"access",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				PartDeclaration{Names: []string{"clarinet"}},
				PartDeclaration{
					Names: []string{"piano", "clarinet", "flute"},
					Alias: "floop",
				},
				PartDeclaration{Names: []string{"floop.piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "clarinet", "piano", "clarinet", "flute"),
				expectCurrentParts("piano"),
			},
		},
		scoreUpdateTestCase{
			label: "group consisting of two groups",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"clarinet", "flute"},
					Alias: "woodwinds",
				},
				PartDeclaration{
					Names: []string{"trumpet", "trombone"},
					Alias: "brass",
				},
				PartDeclaration{
					Names: []string{"woodwinds", "brass"},
					Alias: "wwab",
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("clarinet", "flute", "trumpet", "trombone"),
				expectCurrentParts("clarinet", "flute", "trumpet", "trombone"),
			},
		},
		scoreUpdateTestCase{
			label: "group consisting of two overlapping groups",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"clarinet"}, Alias: "foo"},
				PartDeclaration{Names: []string{"flute"}, Alias: "bar"},
				PartDeclaration{Names: []string{"trumpet"}, Alias: "baz"},
				PartDeclaration{Names: []string{"foo", "bar"}, Alias: "group1"},
				PartDeclaration{Names: []string{"foo", "baz"}, Alias: "group2"},
				PartDeclaration{
					Names: []string{"group1", "group2"},
					Alias: "groups1and2",
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("clarinet", "flute", "trumpet"),
				expectCurrentParts("clarinet", "flute", "trumpet"),
			},
		},
		scoreUpdateTestCase{
			label: "group containing the member of another group",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano", "guitar"}, Alias: "foo"},
				PartDeclaration{Names: []string{"clarinet"}, Alias: "bob"},
				PartDeclaration{Names: []string{"bob", "foo.piano"}, Alias: "bar"},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("clarinet", "piano", "guitar"),
				expectCurrentParts("clarinet", "piano"),
			},
		},
		scoreUpdateTestCase{
			label: "group containing the member of another group: dot access",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano", "guitar"}, Alias: "foo"},
				PartDeclaration{Names: []string{"clarinet"}, Alias: "bob"},
				PartDeclaration{Names: []string{"bob", "foo.piano"}, Alias: "bar"},
				PartDeclaration{Names: []string{"bar.foo.piano"}},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("clarinet", "piano", "guitar"),
				expectCurrentParts("piano"),
			},
		},
		scoreUpdateTestCase{
			label: "in a score with one part, that part is the tempo master",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
			},
			expectations: []scoreUpdateExpectation{
				func(s *Score) error {
					part, err := getPart(s, "piano")
					if err != nil {
						return err
					}

					if part.TempoRole != TempoRoleMaster {
						return fmt.Errorf(
							"Expected TempoRole to be %d, but it was %d",
							TempoRoleMaster, part.TempoRole,
						)
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "in a score with multiple part, the first part is the tempo " +
				"master",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				PartDeclaration{Names: []string{"guitar", "trombone"}, Alias: "group"},
				PartDeclaration{Names: []string{"clarinet"}, Alias: "something"},
			},
			expectations: []scoreUpdateExpectation{
				func(s *Score) error {
					for i, part := range s.Parts {
						expectedTempoRole := TempoRoleUnspecified
						if i == 0 {
							expectedTempoRole = TempoRoleMaster
						}

						if part.TempoRole != expectedTempoRole {
							return fmt.Errorf(
								"Expected TempoRole to be %d, but it was %d",
								expectedTempoRole, part.TempoRole,
							)
						}
					}

					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "ambiguous instrument reference",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
				},
				PartDeclaration{
					Names: []string{"piano"}, Alias: "foo",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Ambiguous instrument reference") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "ambiguous instrument reference (reverse)",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"}, Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"piano"},
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Ambiguous instrument reference") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "referring to a nonexistent instrument/part",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"quizzledyblarf"}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Unrecognized instrument") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "referring to a nonexistent instrument with an alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"quizzledyblarf"},
					Alias: "norman",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Unrecognized instrument") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "assigning a new alias to a part that already has an alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"foo"},
					Alias: "bar",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Can't assign alias") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "reassigning an alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"clarinet"},
					Alias: "foo",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "has already been assigned") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "reassigning a group alias",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "accordion"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"clarinet"},
					Alias: "foo",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "has already been assigned") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "part declaration with duplicate names",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "piano"},
					Alias: "pianos",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "included multiple times") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "part declaration with duplicate names 2",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano", "piano"}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "included multiple times") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "part declaration with duplicate names 3",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{Names: []string{"foo", "foo"}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "included multiple times") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "part declaration with duplicate names 4",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"foo", "foo"},
					Alias: "foos",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "included multiple times") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "mix of named and unnamed parts",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{Names: []string{"foo", "trumpet"}},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(
						err.Error(),
						"can't use both stock instruments and named parts",
					) {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "mix of named and unnamed parts 2",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano"},
					Alias: "foo",
				},
				PartDeclaration{
					Names: []string{"foo", "trumpet"},
					Alias: "floop",
				},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(
						err.Error(),
						"can't use both stock instruments and named parts",
					) {
						return err
					}
					return nil
				},
			},
		},
	)
}
