package cmd

import (
	"alda.io/client/parser"
	"fmt"
	"os"
	"strings"

	"alda.io/client/color"
	"alda.io/client/help"
	"alda.io/client/interop/musicxml/importer"
	"alda.io/client/model"
	"alda.io/client/system"
	"github.com/spf13/cobra"
)

var outputAldaFilename string
var importFormat string

func init() {
	importCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read data from a file to convert to Alda",
	)

	importCmd.Flags().StringVarP(
		&code, "code", "c", "", "Read data from a string to convert to Alda",
	)

	importCmd.Flags().StringVarP(
		&outputAldaFilename, "output", "o", "", "The output Alda code filename",
	)

	importCmd.Flags().StringVarP(
		&importFormat, "import-format", "i", "", "The format of the imported data",
	)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Alda source code from other formats",
	Long: `Import Alda source code from other formats

---

Currently, the only supported import format is MusicXML. Most popular software applications support exporting scores to MusicXML.

---

Source code can be provided in one of three ways:

The path to a file (-f, --file):
  alda import -i musicxml -f path/to/my-score.musicxml -o my-score.alda

A string of code (-c, --code):
  alda import -i musicxml -c "...some musicxml data..." -o my-score.alda

Text piped into the process on stdin:
  echo "...some musicxml data..." | alda import -i musicxml -o my-score.alda

---

When -o / --output FILENAME is provided, the results are written into that file.

  alda import -i musicxml -c "...some musicxml data..." -o my-score.alda

Otherwise, the results are written to stdout, which is convenient for
redirecting into other files or processes.

  alda import -i musicxml -f path/to/my-score.musicxml > my-score.alda
  alda import -i musicxml -f path/to/my-score.musicxml | some-process > my-score.alda

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		if importFormat != "musicxml" {
			return help.UserFacingErrorf(
				`Provided %s is not a supported input format.

Currently, the only supported input format is %s.`,
				color.Aurora.BrightYellow(importFormat),
				color.Aurora.BrightYellow("musicxml"),
			)
		}

		var scoreUpdates []model.ScoreUpdate
		var err error

		switch {
		case file != "":
			inputFile, err := os.Open(file)
			if err != nil {
				return help.UserFacingErrorf(
					`Failed to open file %s: %s.`,
					color.Aurora.BrightYellow(file),
					err.Error(),
				)
			}

			scoreUpdates, err = importer.ImportMusicXML(inputFile)
			if err != nil {
				return err
			}
		case code != "":
			reader := strings.NewReader(code)
			scoreUpdates, err = importer.ImportMusicXML(reader)
			if err != nil {
				return err
			}

		default:
			bytes, err := system.ReadStdin()
			if err != nil {
				return help.UserFacingErrorf(
					`Failed to read from stdin: %s.`,
					err.Error(),
				)
			}

			reader := strings.NewReader(string(bytes))
			scoreUpdates, err = importer.ImportMusicXML(reader)
			if err != nil {
				return err
			}
		}

		root, err := parser.GenerateASTFromScoreUpdates(scoreUpdates)
		if err != nil {
			return help.UserFacingErrorf(
				`Issue generating Alda AST: %s.`,
				err.Error(),
			)
		}

		if outputAldaFilename == "" {
			// When no output filename is specified, we write directly to stdout
			err = parser.FormatASTToCode(root, os.Stdout)
			if err != nil {
				return help.UserFacingErrorf(
					`Issue formatting imported Alda: %s.`,
					err.Error(),
				)
			}
		} else {
			out, err := os.OpenFile(
				outputAldaFilename,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				0664, // default rw-rw-r perms
			)
			if err != nil {
				return help.UserFacingErrorf(
					`Issue opening output file %s.`,
					color.Aurora.BrightYellow(outputAldaFilename),
				)
			}
			defer out.Close()

			err = parser.FormatASTToCode(root, out)
			if err != nil {
				return help.UserFacingErrorf(
					`Issue formatting imported Alda: %s.`,
					err.Error(),
				)
			}

			fmt.Fprintf(
				os.Stderr,
				"Imported score to %s\n.",
				outputAldaFilename,
			)
		}

		return nil
	},
}
