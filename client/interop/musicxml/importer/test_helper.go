package importer

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/go-test/deep"
	"os"
	"testing"
)

type importerTestCase struct {
	label       string
	file        string
	expected    string
	postprocess func(updates []model.ScoreUpdate) []model.ScoreUpdate
}

func (testCase importerTestCase) evaluate() ([]model.ScoreUpdate, error) {
	expected, err := parser.Parse(
		testCase.label, testCase.expected, parser.SuppressSourceContext,
	)

	if err != nil {
		return nil, err
	}

	// Evaluate all LispList elements and unpacked ScoreUpdates
	expected = evaluateLisp(expected)

	if testCase.postprocess != nil {
		expected = testCase.postprocess(expected)
	}

	return expected, nil
}

func executeImporterTestCases(
	t *testing.T, testCases ...importerTestCase,
) {
	for _, testCase := range testCases {
		file, _ := os.Open(testCase.file)
		actual, err := ImportMusicXML(file)
		if err != nil {
			t.Error(testCase.label)
			t.Error(err)
			return
		}
		actual = standardizeBarlines(actual)

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
