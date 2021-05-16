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

		expected, err := parser.Parse(
			testCase.label, testCase.expected, parser.SuppressSourceContext,
		)
		if err != nil {
			t.Error(testCase.label)
			t.Error(err)
			return
		}

		// Evaluate all LispList elements and unpacked ScoreUpdates
		for i, value := range expected {
			if lispList, ok := value.(model.LispList); ok {
				lispForm, err := lispList.Eval()
				if err != nil {
					t.Error(testCase.label)
					t.Error(err)
					return
				}
				expected[i] = lispForm.(model.LispScoreUpdate).ScoreUpdate
			}
		}

		if testCase.postprocess != nil {
			expected = testCase.postprocess(expected)
		}

		if diff := deep.Equal(expected, actual); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}
