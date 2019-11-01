package model

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	_ "alda.io/client/testing"
)

func expectMarker(name string, expectedOffset OffsetMs) func(s *Score) error {
	return func(s *Score) error {
		actualOffset, hit := s.Markers[name]
		if !hit {
			return fmt.Errorf("marker \"%s\" undefined", name)
		}

		if !equalish(expectedOffset, actualOffset) {
			return fmt.Errorf(
				"expected marker \"%s\" to be at offset %f, but it was at offset %f",
				name, expectedOffset, actualOffset,
			)
		}

		return nil
	}
}

func TestMarkers(t *testing.T) {
	executeScoreUpdateTestCases(
		t,
		scoreUpdateTestCase{
			label: "marker placed at offset 0",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Marker{Name: "test-marker"},
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
							NoteLength{Denominator: 1},
						},
					},
				},
				AtMarker{Name: "test-marker"},
				Note{
					NoteLetter: D,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectMarker("test-marker", 0),
				expectPartLastOffset("piano", 0),
				expectPartCurrentOffset("piano", 500),
				expectNoteOffsets(0),
			},
		},
		scoreUpdateTestCase{
			label: "@marker referring to a marker placed by another part",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				// Wait for 2 measures and place a marker.
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
							NoteLength{Denominator: 1},
						},
					},
				},
				Marker{Name: "test-marker"},

				// Begin at the marker placed in the piano part.
				PartDeclaration{Names: []string{"bassoon"}},
				AtMarker{Name: "test-marker"},
				Note{
					NoteLetter: D,
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 4},
						},
					},
				},
			},
			expectations: []scoreUpdateExpectation{
				expectMarker("test-marker", 4000),
				expectPartLastOffset("bassoon", 4000),
				expectPartCurrentOffset("bassoon", 4500),
				expectNoteOffsets(4000),
			},
		},
		scoreUpdateTestCase{
			label: "reference to undefined marker",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				AtMarker{Name: "nonexistent-marker"},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !strings.Contains(err.Error(), "Marker undefined") {
						return err
					}
					return nil
				},
			},
		},
		scoreUpdateTestCase{
			label: "marker placed in multiple parts with differing current offsets",
			updates: []ScoreUpdate{
				PartDeclaration{Names: []string{"piano"}},
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 1},
							NoteLength{Denominator: 1},
						},
					},
				},

				PartDeclaration{Names: []string{"celeste"}},
				Rest{
					Duration: Duration{
						Components: []DurationComponent{
							NoteLength{Denominator: 2},
						},
					},
				},

				PartDeclaration{Names: []string{"piano", "celeste"}},
				Marker{Name: "ambiguous-offset-marker"},
			},
			errorExpectations: []scoreUpdateErrorExpectation{
				func(err error) error {
					if !regexp.MustCompile("offset.*unclear").MatchString(err.Error()) {
						return err
					}
					return nil
				},
			},
		},
	)
}
