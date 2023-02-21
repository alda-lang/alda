package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "alda.io/client/testing"
)

// TestExamples tests each of the example scores in the `examples` directory.
func TestExamples(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	examplesDir := filepath.Join(filepath.Dir(filepath.Dir(dir)), "examples")

	filepath.Walk(examplesDir,
		func(path string, info os.FileInfo, err error) error {
			label := fmt.Sprintf("examples test for %s", path)
			if err != nil {
				t.Error(label)
				t.Errorf("filepath walk error %s", err)
				return err
			}

			if info.IsDir() {
				return nil
			}

			contents, err := ioutil.ReadFile(path)
			if err != nil {
				t.Error(label)
				t.Error("issue reading file contents")
				return err
			}

			executeParseTestCases(t, parseTestCase{
				label: label,
				given: string(contents),
			})

			return nil
		},
	)
}
