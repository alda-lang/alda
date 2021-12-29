package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

// TestExamples parses and compiles each of the example scores in the `examples`
// directory.
//
// Strictly speaking, this is not testing just the parser. Perhaps these tests
// should live in another package that consumes both the `parser` and `model`
// packages, but I'm not sure yet which package that should be.
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

			fmt.Printf("● %s\n", path)

			ast, err := ParseFile(path)
			if err != nil {
				return err
			}

			scoreUpdates, err := ast.Updates()
			if err != nil {
				return err
			}

			score := model.NewScore()
			return score.Update(scoreUpdates...)
		})
	if err != nil {
		t.Errorf("%v\n", err)
	}
}
