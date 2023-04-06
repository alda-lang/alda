package importer

import (
	"os"
	"testing"

	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/go-test/deep"
)

type importerTestCase struct {
	label       string
	file        string
	expected    string
	postprocess func(updates []model.ScoreUpdate) []model.ScoreUpdate
}

func (testCase importerTestCase) evaluate() ([]model.ScoreUpdate, error) {
	expectedAST, err := parser.Parse(
		testCase.label, testCase.expected, parser.SuppressSourceContext,
	)
	if err != nil {
		return nil, err
	}

	expectedUpdates, err := expectedAST.Updates()
	if err != nil {
		return nil, err
	}

	// Evaluate all LispList elements and unpacked ScoreUpdates
	expectedUpdates = evaluateLisp(expectedUpdates)

	if testCase.postprocess != nil {
		expectedUpdates = testCase.postprocess(expectedUpdates)
	}

	return expectedUpdates, nil
}

func executeImporterTestCases(
	t *testing.T, testCases ...importerTestCase,
) {
	for _, testCase := range testCases {
		b, err := os.ReadFile(testCase.file)
		if err != nil {
			t.Error(testCase.label)
			t.Error(err)
			return
		}

		actual, err := ImportMusicXML(b)
		if err != nil {
			t.Error(testCase.label)
			t.Error(err)
			return
		}

		expected, err := testCase.evaluate()
		if err != nil {
			t.Error(testCase.label)
			t.Error(err)
			return
		}
		expected = standardizeBarlines(expected)

		if diff := deep.Equal(expected, actual); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}
