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
			given: "%chorus",
			expect: []model.ScoreUpdate{
				model.Marker{Name: "chorus"},
			},
		},
		parseTestCase{
			label: "at marker",
			given: "@verse-1",
			expect: []model.ScoreUpdate{
				model.AtMarker{Name: "verse-1"},
			},
		},
	)
}
