package parser

import (
	"alda.io/client/model"
	_ "alda.io/client/testing"
	"testing"
)

func TestMarkers(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "marker",
			given: "piano: %chorus",
			expectUpdates: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Marker{Name: "chorus"},
			},
		},
		parseTestCase{
			label: "at marker",
			given: "piano: %verse-1 @verse-1",
			expectUpdates: []model.ScoreUpdate{
				model.PartDeclaration{Names: []string{"piano"}},
				model.Marker{Name: "verse-1"},
				model.AtMarker{Name: "verse-1"},
			},
		},
	)
}
