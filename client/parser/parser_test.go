package parser

import (
	_ "alda.io/client/testing"
	"os"
	"path/filepath"
	"testing"
)

func TestExamples(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("%v\n", err)
		return
	}

	examplesDir := filepath.Join(filepath.Dir(filepath.Dir(dir)), "examples")

	err = filepath.Walk(examplesDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			return ParseFile(path)
		})
	if err != nil {
		t.Errorf("%v\n", err)
	}
}
