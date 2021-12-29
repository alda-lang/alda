package cmd

import (
	"fmt"
	"time"

	"alda.io/client/color"
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
	Long: fmt.Sprintf(`Display the result of parsing Alda source code

---

%s

---

The -o / --output parameter determines what output is displayed.
Options include:

ast:

  The AST that results from parsing the source code, represented as a JSON
  object.

ast-human:

  The AST that results from parsing the source code, displayed in a more
  human-readable way.

events:

  A JSON array of objects, each of which represents an "event" parsed from the
  source code.

data (default):

  A JSON object representing the score that is constructed after parsing the
  source code into events and evaluating them in order within the context of a
  new score.

---`,
		sourceCodeInputOptions("parse", false),
	),
	RunE: func(_ *cobra.Command, args []string) error {
		switch outputType {
		case "ast", "ast-human", "events", "data": // OK to proceed
		default:
			return help.UserFacingErrorf(
				`%s is not a supported output type.

Please choose one of:

  %s
  The AST that results from parsing the source code, represented as a JSON
  object.

  %s
  The AST that results from parsing the source code, displayed in a more
  human-readable way.

  %s
  A JSON array of objects, each of which represents an "event" parsed from the
  source code.

  %s (default)
  A JSON object representing the score that is constructed after parsing the
  source code into events and evaluating them in order within the context of a
  new score.`,
				color.Aurora.BrightYellow(outputType),
				color.Aurora.BrightYellow("ast"),
				color.Aurora.BrightYellow("ast-human"),
				color.Aurora.BrightYellow("events"),
				color.Aurora.BrightYellow("data"),
			)
		}

		var ast parser.ASTNode
		var scoreUpdates []model.ScoreUpdate
		var err error

		switch {
		case file != "":
			ast, err = parser.ParseFile(file)

		case code != "":
			ast, err = parser.ParseString(code)

		default:
			ast, err = parseStdin()
		}

		if err == errNoInputSupplied {
			return userFacingNoInputSuppliedError("parse")
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

		if err != nil {
			return err
		}

		if outputType == "ast" {
			fmt.Println(ast.JSON().String())

			return nil
		}

		if outputType == "ast-human" {
			// HACK: Instead of the code below, we should really just be able to have
			// this one line:
			//
			//   fmt.Println(parser.HumanReadableAST(ast.JSON()))
			//
			// However, for some bizarre reason, it only seems to work if we serialize
			// the JSON directly emitted by ast.JSON() as a string, then parse the
			// string to get a new object. I guess something about ast.JSON() is doing
			// something weird with nested gabs.Containers, or something like that.
			jsonObj := ast.JSON()
			jsonStr := jsonObj.String()
			parsedJsonObj, err := json.ParseJSON([]byte(jsonStr))
			if err != nil {
				return err
			}

			fmt.Println(parser.HumanReadableAST(parsedJsonObj))

			return nil
		}

		scoreUpdates, err = ast.Updates()

		// Errors with source context are presented to the user as-is.
		//
		// See TODO notes above.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		if err != nil {
			return err
		}

		if outputType == "events" {
			updates := json.Array()
			for _, update := range scoreUpdates {
				updates.ArrayAppend(update.JSON())
			}

			fmt.Println(updates.String())

			return nil
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

		if err != nil {
			return err
		}

		log.Info().
			Int("updates", len(scoreUpdates)).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Constructed score.")

		fmt.Println(score.JSON().String())

		return nil
	},
}
