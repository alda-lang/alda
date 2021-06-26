package importer

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/go-test/deep"
	"os"
	"testing"
)

type importerTestCase interface {
	testCaseLabel()    string
	testCaseFile()     string
	evaluate() ([]model.ScoreUpdate, error)
}

func executeImporterTestCases(
	t *testing.T, testCases ...importerTestCase,
) {
	for _, testCase := range testCases {
		file, _ := os.Open(testCase.testCaseFile())
		actual, err := ImportMusicXML(file)
		if err != nil {
			t.Error(testCase.testCaseLabel())
			t.Error(err)
			return
		}

		expected, err := testCase.evaluate()
		if err != nil {
			t.Error(testCase.testCaseLabel())
			t.Error(err)
			return
		}

		if diff := deep.Equal(expected, actual); diff != nil {
			t.Error(testCase.testCaseLabel())
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}

type testCaseWithAlda struct {
	label       string
	file        string
	expected    string
	postprocess func(updates []model.ScoreUpdate) []model.ScoreUpdate
}

func (testCase testCaseWithAlda) testCaseLabel() string {
	return testCase.label
}

func (testCase testCaseWithAlda) testCaseFile() string {
	return testCase.file
}

func (testCase testCaseWithAlda) evaluate() ([]model.ScoreUpdate, error) {
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

type testCaseWithUpdates struct {
	label    string
	file     string
	expected []model.ScoreUpdate
}

func (testCase testCaseWithUpdates) testCaseLabel() string {
	return testCase.label
}

func (testCase testCaseWithUpdates) testCaseFile() string {
	return testCase.file
}

func (testCase testCaseWithUpdates) evaluate() ([]model.ScoreUpdate, error) {
	return testCase.expected, nil
}

func aldaPercussionNote(number int32, duration float64) model.ScoreUpdate {
	return model.Note{
		Pitch:    model.MidiNoteNumber{MidiNote: number},
		Duration: model.Duration{Components: []model.DurationComponent{
			model.NoteLength{Denominator: duration},
		}},
	}
}

func aldaPercussionNoteWithBarline(
	number int32, duration float64,
) model.ScoreUpdate {
	return model.Note{
		Pitch:    model.MidiNoteNumber{MidiNote: number},
		Duration: model.Duration{Components: []model.DurationComponent{
			model.NoteLength{Denominator: duration},
			model.Barline{},
		}},
	}
}

func aldaRest(duration float64) model.ScoreUpdate {
	return model.Rest{
		Duration: model.Duration{Components: []model.DurationComponent{
			model.NoteLength{Denominator: duration},
		}},
	}
}

func aldaRestWithBarline(duration float64) model.ScoreUpdate {
	return model.Rest{
		Duration: model.Duration{Components: []model.DurationComponent{
			model.NoteLength{Denominator: duration},
			model.Barline{},
		}},
	}
}
