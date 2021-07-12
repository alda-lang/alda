package cmd

import (
	"alda.io/client/code_formatter"
	"alda.io/client/code_generator"
	"alda.io/client/color"
	"alda.io/client/help"
	"alda.io/client/interop/musicxml/importer"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/system"
	"alda.io/client/transmitter"
	"alda.io/client/util"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
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
	Use:    "import",
	Hidden: true,
	Short:  "Evaluate external format and import as Alda source code",
	Long: `Evaluate external format and import as Alda source code

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

---

Currently, the only import format is MusicXML.  

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		if importFormat != "musicxml" {
			return help.UserFacingErrorf(
				`%s is not a supported input format.

Currently, the only supported output format is %s.`,
				color.Aurora.BrightYellow(importFormat),
				color.Aurora.BrightYellow("musicxml"),
			)
		}

		// TODO (experimental): remove warning log
		log.Warn().Msg(fmt.Sprintf(
			`The %s command is currently experimental. Imported scores may be incorrect and lack information.`,
			color.Aurora.BrightYellow("import"),
		))

		var scoreUpdates []model.ScoreUpdate
		var err error

		// TODO: add XML validation
		// TODO: add XML conversion to ensure we get score-partwise pieces as input
		switch {
		case file != "":
			inputFile, err := os.Open(file)
			if err != nil {
				return err
			}

			scoreUpdates, err = importer.ImportMusicXML(inputFile)
		case code != "":
			reader := strings.NewReader(code)
			scoreUpdates, err = importer.ImportMusicXML(reader)

		default:
			bytes, err := readStdin()
			if err != nil {
				return err
			}

			reader := strings.NewReader(string(bytes))
			scoreUpdates, err = importer.ImportMusicXML(reader)
		}

		if err != nil {
			return err
		}

		tokens := code_generator.Generate(scoreUpdates)

		if outputAldaFilename == "" {
			// When no output filename is specified, we write directly to stdout
			code_formatter.Format(tokens, os.Stdout)
		} else {
			file, err := os.Create(outputAldaFilename)
			if err != nil {
				return err
			}

			code_formatter.Format(tokens, file)

			fmt.Fprintf(os.Stderr, "Imported score to %s\n", outputAldaFilename)
			if err := file.Close(); err != nil {
				return err
			}
		}

		// Bootstrapping the play command to apply the scoreUpdates
		// TODO (experimental): Remove this temporary play-back
		score := model.NewScore()
		err = score.Update(scoreUpdates...)
		action := "play"

		// Errors with source context are presented to the user as-is.
		//
		// TODO: See TODO comment in cmd/parse.go about writing better user-facing
		// error messages.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		if err != nil {
			return err
		}

		var players []system.PlayerState

		// Determine the port to use based on the provided CLI options.
		switch {

		// Port is explicitly specified, so use that port.
		case playerPort != -1:
			player := system.PlayerState{
				ID:    "unknown",
				State: "unknown",
				Port:  playerPort,
			}
			players = []system.PlayerState{player}

		// Player ID is specified; look up the player by ID and use its port.
		case playerID != "":
			player, err := system.FindPlayerByID(playerID)
			if err != nil {
				return err
			}
			players = []system.PlayerState{player}

		// We're actually unpausing, not playing, so send the message to all active
		// player processes so that if any of them are paused, they'll resume
		// playing.
		case action == "unpause":
			allPlayers, err := system.ReadPlayerStates()
			if err != nil {
				return err
			}
			players = []system.PlayerState{}
			for _, player := range allPlayers {
				if player.State == "active" {
					players = append(players, player)
				}
			}

		// Find an available player process to use.
		default:
			system.StartingPlayerProcesses()

			if err := util.Await(
				func() error {
					player, err := system.FindAvailablePlayer()
					if err != nil {
						return err
					}

					players = []system.PlayerState{player}
					return nil
				},
				reasonableTimeout,
			); err != nil {
				return err
			}
		}

		transmitOpts := []transmitter.TransmissionOption{
			transmitter.TransmitFrom(optionFrom),
			transmitter.TransmitTo(optionTo),
		}

		if action == "play" {
			transmitOpts = append(transmitOpts, transmitter.OneOff())
		}

		for _, player := range players {
			if _, err := ping(player.Port); err != nil {
				return err
			}

			transmitter := transmitter.OSCTransmitter{Port: player.Port}

			var transmissionError error
			if action == "unpause" {
				transmissionError = transmitter.TransmitPlayMessage()
			} else {
				transmissionError = transmitter.TransmitScore(score, transmitOpts...)
			}
			if transmissionError != nil {
				return transmissionError
			}
		}

		// We don't have to print something here, but it's a good idea because it
		// indicates to the user that we did what they asked. Otherwise, it might
		// not be obvious that we did anything, especially in cases where there is
		// no audible output, e.g. `alda play -c "c d e"` (valid syntax, but no
		// audible output because no part was indicated).
		fmt.Fprintln(os.Stderr, "Playing...")

		return nil
	},
}
