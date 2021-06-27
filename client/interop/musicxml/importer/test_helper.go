package importer

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/go-test/deep"
	"os"
	"reflect"
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
	expected, err = evaluateLisp(expected)

	if err != nil {
		return nil, err
	}

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

func standardizeBarlines(updates []model.ScoreUpdate) []model.ScoreUpdate {
	// Alda has to parse barlines into note and rest duration components to
	// handle ties
	// This causes various issues while importing
	// So in our tests, we will "standardize" the location of barlines
	// Any barline that is the last duration component will be moved outside
	for i := len(updates) - 1; i >= 0; i-- {
		barlineAfter := false

		update := updates[i]
		switch typedUpdate := update.(type) {
		case model.Note:
			durations := typedUpdate.Duration.Components
			if len(durations) > 0 &&
				reflect.TypeOf(durations[len(durations) - 1]) == barlineType {
				durations = durations[:len(durations) - 1]
				if len(durations) == 0 {
					durations = nil
				}
				typedUpdate.Duration.Components = durations
				update = typedUpdate
				barlineAfter = true
			}
		case model.Rest:
			durations := typedUpdate.Duration.Components
			if len(durations) > 0 &&
				reflect.TypeOf(durations[len(durations) - 1]) == barlineType {
				durations = durations[:len(durations) - 1]
				if len(durations) == 0 {
					durations = nil
				}
				typedUpdate.Duration.Components = durations
				update = typedUpdate
				barlineAfter = true
			}
		}

		updates[i] = update
		if barlineAfter {
			updates = insert(model.Barline{}, updates, i + 1)
		}

		// Recursively standardize barlines
		if modified, ok := modifyNestedUpdates(
			update, standardizeBarlines,
		); ok {
			updates[i] = modified
		}
	}

	return updates
}

func evaluateLisp(updates []model.ScoreUpdate) ([]model.ScoreUpdate, error) {
	for i, element := range updates {
		switch value := element.(type) {
		case model.Repeat:
			eventSequence := value.Event.(model.EventSequence)
			evaluateLisp(eventSequence.Events)
			value.Event = eventSequence
			updates[i] = value
		case model.OnRepetitions:
			eventSequence := value.Event.(model.EventSequence)
			evaluateLisp(eventSequence.Events)
			value.Event = eventSequence
			updates[i] = value
		default:
			if lispList, ok := value.(model.LispList); ok {
				lispForm, err := lispList.Eval()
				if err != nil {
					return nil, err
				}
				updates[i] = lispForm.(model.LispScoreUpdate).ScoreUpdate
			}
		}
	}
	return updates, nil
}
