package parser

import (
	"testing"

	"alda.io/client/model"
	"github.com/go-test/deep"
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
		// We're suppressing the source context here because we're about to do a
		// deep diff of our expected list of score updates (which are all devoid of
		// source context) and the actual list of score updates that result from
		// parsing the score, and we need them to be considered the same even if the
		// source context differs.
		//
		// By default, the actual list of updates will include source context, which
		// would cause an avalanche of spurious test failures because the diff would
		// show numerous differences in the source contexts.
		actualAST, err :=
			Parse(testCase.label, testCase.given, SuppressSourceContext)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}

		actualUpdates, err := actualAST.Updates()
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}

		if diff := deep.Equal(testCase.expect, actualUpdates); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}
