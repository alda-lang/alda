package cmd

import (
	"alda.io/client/color"
	log "alda.io/client/logging"
	"alda.io/client/parser"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)
var formatInputFile string
var formatOutputFile string

func init() {
	formatCmd.Flags().StringVarP(
		&formatInputFile, "file", "f", "", "Format an Alda file",
	)

	formatCmd.Flags().StringVarP(
		&formatOutputFile, "output", "o", "", "A separate file to output formatted Alda to",
	)
}

var formatCmd = &cobra.Command{
	Use: "format",
	Hidden: true,
	Short: "Format Alda source code",
	Long: `Format Alda source code

---

Source code must be provided by specifying the path to a file (-f, --file):
  alda format -f path/to/my-score.alda

In this case, the score will be formatted in-place (i.e. overwritten).

When -o / --output FILENAME is provided, the results are instead written into that file.
  alda format -f path/to/my-score.alda -o path/to/formatted-score.alda

---

Currently, formatting does not support comments (i.e. all comments are dropped)

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

		if formatOutputFile == "" {
			formatOutputFile = formatInputFile
		}

		out, err := os.OpenFile(
			formatOutputFile,
			os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
			0664,	// default rw-rw-r perms
		)
		if err != nil {
			return err
		}

		err = parser.FormatASTToCode(root, out)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Formatted score to %s\n", formatOutputFile)
		if err := out.Close(); err != nil {
			return err
		}

		return nil
	},
}
