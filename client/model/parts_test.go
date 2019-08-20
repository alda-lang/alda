package model

import (
	"fmt"
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func getParts(score *Score, stockInstrumentName string) ([]*Part, error) {
	stockInstrument, err := stockInstrument(stockInstrumentName)
	if err != nil {
		return nil, err
	}

	parts := []*Part{}

	for _, part := range score.Parts {
		if part.StockInstrument == stockInstrument {
			parts = append(parts, part)
		}
	}

	return parts, nil
}

func getPart(score *Score, stockInstrumentName string) (*Part, error) {
	parts, err := getParts(score, stockInstrumentName)
	if err != nil {
		return nil, err
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf(
			"score doesn't include a %s part", stockInstrumentName,
		)
	}

	if len(parts) > 1 {
		return nil, fmt.Errorf(
			"score includes more than one %s part", stockInstrumentName,
		)
	}

	return parts[0], nil
}

func expectParts(instruments ...string) func(s *Score) error {
	return func(s *Score) error {
		if len(s.Parts) != len(instruments) {
			return fmt.Errorf(
				"expected %d parts, got %d",
				len(instruments),
				len(s.Parts))
		}

		accountedFor := map[*Part]bool{}

		for _, instrument := range instruments {
			parts, err := getParts(s, instrument)
			if err != nil {
				return err
			}

			if len(parts) == 0 {
				return fmt.Errorf("score doesn't include a %s part", instrument)
			}

			instrumentAccountedFor := false
			for _, part := range parts {
				if !accountedFor[part] {
					accountedFor[part] = true
					instrumentAccountedFor = true
					break
				}
			}
			if !instrumentAccountedFor {
				return fmt.Errorf("score missing a %s part", instrument)
			}
		}

		return nil
	}
}

func expectPart(
	instrument string, expectations ...func(*Part) error,
) func(s *Score) error {
	return func(s *Score) error {
		part, err := getPart(s, instrument)
		if err != nil {
			return err
		}

		for _, expectation := range expectations {
			if err := expectation(part); err != nil {
				return err
			}
		}

		return nil
	}
}

func expectCurrentParts(instruments ...string) func(s *Score) error {
	return func(s *Score) error {
		if len(s.CurrentParts) != len(instruments) {
			return fmt.Errorf(
				"expected %d current parts, got %d",
				len(instruments),
				len(s.CurrentParts))
		}

		for _, instrument := range instruments {
			parts, err := getParts(s, instrument)
			if err != nil {
				return err
			}

			includedInCurrentParts := false
			for _, currentPart := range s.CurrentParts {
				for _, part := range parts {
					if currentPart == part {
						includedInCurrentParts = true
					}
				}
			}
			if !includedInCurrentParts {
				return fmt.Errorf(
					"%s part not included in current parts", instrument,
				)
			}
		}

		return nil
	}
}

func expectAliasParts(
	alias string, instruments ...string,
) func(s *Score) error {
	return func(s *Score) error {
		aliasParts := s.Aliases[alias]

		if len(aliasParts) != len(instruments) {
			return fmt.Errorf(
				"expected alias %q to refer to %d parts, not %d parts",
				alias,
				len(instruments),
				len(aliasParts),
			)
		}

		for _, instrument := range instruments {
			part, err := getPart(s, instrument)
			if err != nil {
				return err
			}

			includedInAlias := false
			for _, aliasPart := range aliasParts {
				if aliasPart == part {
					includedInAlias = true
				}
			}
			if !includedInAlias {
				return fmt.Errorf("%s part not included in alias %q", instrument, alias)
			}

			dotAlias := alias + "." + instrument
			dotAliasParts := s.Aliases[dotAlias]
			if len(dotAliasParts) != 1 {
				return fmt.Errorf(
					"expected %q alias to refer to 1 part, not %d parts",
					dotAlias,
					len(dotAliasParts),
				)
			}
			if dotAliasParts[0] != part {
				return fmt.Errorf("%q alias not set up", dotAlias)
			}
		}

		return nil
	}
}

func TestParts(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "part initialization",
			updates: []ScoreUpdate{
				PartDeclaration{
					Names: []string{"piano", "trumpet"},
					Alias: "trumpiano",
				},
			},
			expectations: []scoreUpdateExpectation{
				expectParts("piano", "trumpet"),
				expectCurrentParts("piano", "trumpet"),
				expectAliasParts("trumpiano", "piano", "trumpet"),
				func(s *Score) error {
					for _, instrument := range []string{"piano", "trumpet"} {
						part, err := getPart(s, instrument)
						if err != nil {
							return err
						}

						if part.CurrentOffset != 0 {
							return fmt.Errorf(
								"%s part's offset is %f, not 0",
								instrument,
								part.CurrentOffset,
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
