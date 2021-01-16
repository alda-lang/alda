package cmd

import (
	"fmt"
	"time"

	"alda.io/client/help"
	"alda.io/client/json"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/spf13/cobra"
)

var outputType string

func init() {
	parseCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read Alda source code from a file",
	)

	parseCmd.Flags().StringVarP(
		&code, "code", "c", "", "Supply Alda source code as a string",
	)

	parseCmd.Flags().StringVarP(
		&outputType, "output", "o", "data", "The desired parse output",
	)
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Display the result of parsing Alda source code",
	Long: `Display the result of parsing Alda source code

---

Source code can be provided in one of three ways:

The path to a file (-f, --file):
  alda parse -f path/to/my-score.alda

A string of code (-c, --code):
  alda parse -c "harpsichord: o5 d+8 < b g+ e d+1"

Text piped into the process on stdin:
  echo "glockenspiel: o5 g8 < g > g e4 d4." | alda parse

---

The -o / --output parameter determines what output is displayed. Options include:

events:

  A JSON array of objects, each of which represents an "event" parsed from the
  source code.

data:

  A JSON object representing the score that is constructed after parsing the
  source code into events and evaluating them in order within the context of a
  new score.

---`,
	Run: func(_ *cobra.Command, args []string) {
		switch outputType {
		case "events", "data": // OK to proceed
		default:
			// TODO: user-facing error
			help.ExitOnError(
				fmt.Errorf("Invalid output type: %s\n", outputType),
			)
		}

		var scoreUpdates []model.ScoreUpdate
		var err error

		switch {
		case file != "":
			scoreUpdates, err = parser.ParseFile(file)

		case code != "":
			scoreUpdates, err = parser.ParseString(code)

		default:
			scoreUpdates, err = parseStdin()
		}

		// Errors with source context are presented to the user as-is.
		//
		// TODO: Consider writing better user-facing error messages with suggestions
		// for the users to try, etc. The Elm compiler is a good source of
		// inspiration.
		//
		// TODO: If we do the above, we should also consider providing a flag that
		// keeps the error messages terse/parseable, the way they are now. That way,
		// tooling can be built around the parseable error messages with source
		// context, e.g. an editor plugin can easily parse out the line and column
		// number and display the error message at the relevant position in the
		// file.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		help.ExitOnError(err)

		if outputType == "events" {
			updates := json.Array()
			for _, update := range scoreUpdates {
				updates.ArrayAppend(update.JSON())
			}

			fmt.Println(updates.String())

			return
		}

		score := model.NewScore()
		start := time.Now()
		err = score.Update(scoreUpdates...)

		// Errors with source context are presented to the user as-is.
		//
		// TODO: See TODO comment above about writing better user-facing error
		// messages.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		help.ExitOnError(err)

		log.Info().
			Int("updates", len(scoreUpdates)).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Constructed score.")

		fmt.Println(score.JSON().String())
	},
}
