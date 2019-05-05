package parser

import (
	"alda.io/client/model"
	"github.com/go-test/deep"
	"testing"
)

// A parseTestCase models the score updates that should result from parsing a
// string of Alda code.
type parseTestCase struct {
	label  string
	given  string
	expect []model.ScoreUpdate
}

// executeParseTestCases parses each test case's given string of Alda code and
// compares the result with the expected sequence of score updates.
func executeParseTestCases(t *testing.T, testCases ...parseTestCase) {
	for _, testCase := range testCases {
		actual, err := Parse(testCase.label, testCase.given)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}

		if diff := deep.Equal(testCase.expect, actual); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}
