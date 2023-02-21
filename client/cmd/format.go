package cmd

import (
	"alda.io/client/color"
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/parser"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var formatInputFile string
var formatOverwrite bool
var formatConfiguredWrapLen int
var formatConfiguredIndentText string

func init() {
	formatCmd.Flags().StringVarP(
		&formatInputFile, "file", "f", "", "Input Alda file to format",
	)

	formatCmd.Flags().BoolVarP(
		&formatOverwrite, "output", "o", false, "Overwrite input file with formatted output",
	)

	formatCmd.Flags().IntVarP(
		&formatConfiguredWrapLen, "wrap", "w", 0, "Configured line character length to wrap formatted output (default 80)",
	)

	formatCmd.Flags().StringVarP(
		&formatConfiguredIndentText, "indent", "i", "", "Configured indent text (default two spaces)",
	)
}

var formatCmd = &cobra.Command{
	Use:    "format",
	Hidden: true,
	Short:  "Format Alda source code",
	Long: `Format Alda source code

---

Source code must be provided by specifying the path to a file (-f, --file):
  alda format -f path/to/my-score.alda

In this case, the formatted output will be printed to standard output.

When -o / --output is specified, the input file is instead overwritten.
  alda format -f path/to/my-score.alda -o

Formatted output can be configured with the -w / --wrap and -i / --indent flags.
  alda format -f path/to/my-score.alda -w 120 -i "    "

---

Currently, formatting cannot handle comments (i.e. all comments are dropped)

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		// TODO (experimental): remove warning log
		log.Warn().Msg(fmt.Sprintf(
			`The %s command is currently experimental. All comments are dropped during formatting.`,
			color.Aurora.BrightYellow("format"),
		))

		root, err := parser.ParseFile(formatInputFile)
		if err != nil {
			return err
		}

		var out io.Writer
		if formatOverwrite {
			f, err := os.OpenFile(
				formatInputFile,
				os.O_WRONLY|os.O_TRUNC,
				0664, // default rw-rw-r perms
			)
			if err != nil {
				return help.UserFacingErrorf(
					`Issue opening file %s.`,
					color.Aurora.BrightYellow(outputAldaFilename),
				)
			}
			defer f.Close()
			out = f
		} else {
			out = os.Stdout
		}

		if formatConfiguredWrapLen < 0 {
			return help.UserFacingErrorf(
				`Configured line wrap length %d must be positive.`,
				formatConfiguredWrapLen,
			)
		}

		if formatConfiguredWrapLen > 0 && len(formatConfiguredIndentText) > 0 {
			err = parser.FormatASTToCode(
				root,
				out,
				parser.ConfigureSoftWrapLen(formatConfiguredWrapLen),
				parser.ConfigureIndentText(formatConfiguredIndentText),
			)
		} else if formatConfiguredWrapLen > 0 {
			err = parser.FormatASTToCode(
				root,
				out,
				parser.ConfigureSoftWrapLen(formatConfiguredWrapLen),
			)
		} else if len(formatConfiguredIndentText) > 0 {
			err = parser.FormatASTToCode(
				root,
				out,
				parser.ConfigureIndentText(formatConfiguredIndentText),
			)
		} else {
			err = parser.FormatASTToCode(root, out)
		}

		if err != nil {
			return help.UserFacingErrorf(
				`Issue formatting Alda: %s.`,
				err.Error(),
			)
		}

		return nil
	},
}
