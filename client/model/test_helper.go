package model

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type scoreUpdateExpectation func(*Score) error
type scoreUpdateErrorExpectation func(error) error

// A scoreUpdateTestCase models the expected outcome of applying a list of
// ScoreUpdates to a new score.
type scoreUpdateTestCase struct {
	label             string
	updates           []ScoreUpdate
	expectations      []scoreUpdateExpectation
	errorExpectations []scoreUpdateErrorExpectation
}

// executeScoreUpdateTestCases applies each test case's list of updates to a new
// score, asserts that this was done successfully (or unsuccessfully, if that's
// what we expect to happen), and checks the test case's list of expectations
// against the updated score (or the error).
func executeScoreUpdateTestCases(
	t *testing.T, testCases ...scoreUpdateTestCase,
) {
	for _, testCase := range testCases {
		score := NewScore()
		err := score.Update(testCase.updates...)

		if len(testCase.errorExpectations) == 0 {
			// happy path test
			if err != nil {
				t.Error(testCase.label)
				t.Error(err)
				return
			}

			for _, expectation := range testCase.expectations {
				if err := expectation(score); err != nil {
					t.Error(spew.Sdump(score))
					t.Error(testCase.label)
					t.Error(err)
					return
				}
			}
		} else {
			// sad path test
			if err == nil {
				t.Error(spew.Sdump(score))
				t.Error(testCase.label)
				t.Error("Unexpected success applying score updates.")
				return
			}
			for _, expectation := range testCase.errorExpectations {
				if err := expectation(err); err != nil {
					t.Error(spew.Sdump(score))
					t.Error(testCase.label)
					t.Error(err)
					return
				}
			}
		}
	}
}

// Floating point equality gets weird, so we consider two floating point numbers
// to be equal-ish if they are within a small threshold of one another.

// Without this, we get errors like "expected note #2 to have offset 333.333333,
// but it was 333.333344"
func equalish(float1 float64, float2 float64) bool {
	threshold := 0.001

	var lower, upper float64
	if float1 <= float2 {
		lower = float1
		upper = float2
	} else {
		lower = float2
		upper = float1
	}

	return (lower + threshold) > upper
}

// The float32 variation of equalish.
func equalish32(float1 float32, float2 float32) bool {
	return equalish(float64(float1), float64(float2))
}
