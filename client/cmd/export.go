package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/transmitter"
	"alda.io/client/util"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

const midiExportTimeout = 60 * time.Second

var outputFilename string
var outputFormat string

func init() {
	exportCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read Alda source code from a file",
	)

	exportCmd.Flags().StringVarP(
		&code, "code", "c", "", "Supply Alda source code as a string",
	)

	exportCmd.Flags().StringVarP(
		&optionFrom,
		"from",
		"F",
		"",
		"A time marking (e.g. 0:30) or marker from which to start",
	)

	exportCmd.Flags().StringVarP(
		&optionTo,
		"to",
		"T",
		"",
		"A time marking (e.g. 1:00) or marker at which to end",
	)

	exportCmd.Flags().StringVarP(
		&outputFilename, "output", "o", "", "The output filename",
	)

	exportCmd.Flags().StringVarP(
		&outputFormat, "output-format", "O", "midi", "The output format",
	)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Evaluate Alda source code and export to another format",
	Long: `Evaluate Alda source code and export to another format

---

Source code can be provided in one of three ways:

The path to a file (-f, --file):
  alda export -f path/to/my-score.alda -o my-score.mid

A string of code (-c, --code):
  alda export -c "harpsichord: o5 d+8 < b g+ e d+1" -o harpsy.mid

Text piped into the process on stdin:
  echo "glockenspiel: o5 g8 < g > g e4 d4." | alda export -o glock.mid

---

When -o / --output FILENAME is provided, the results are written into that file.

  alda export -c "piano: c d e" -o three-notes.mid

Otherwise, the results are written to stdout, which is convenient for
redirecting into other files or processes.

  alda export -c "piano: c d e" > three-notes.mid
  alda export -c "piano: c d e" | some-process > three-notes-processed.mid

---

Currently, the only output format is MIDI. At some point, there will be other
output formats like MusicXML.

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		if outputFormat != "midi" {
			return help.UserFacingErrorf(
				`%s is not a supported output format.

Currently, the only supported output format is %s.`,
				aurora.Yellow(outputFormat),
				aurora.Yellow("midi"),
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

		if err != nil {
			return err
		}

		score := model.NewScore()
		start := time.Now()
		if err := score.Update(scoreUpdates...); err != nil {
			return err
		}

		log.Info().
			Int("updates", len(scoreUpdates)).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Constructed score.")

		var player system.PlayerState

		// Find an available player process to use.
		system.StartingPlayerProcesses()
		if err := util.Await(
			func() error {
				foundPlayer, err := system.FindAvailablePlayer()
				if err != nil {
					return err
				}

				player = foundPlayer
				return nil
			},
			reasonableTimeout,
		); err != nil {
			return err
		}

		transmitOpts := []transmitter.TransmissionOption{
			transmitter.TransmitFrom(optionFrom),
			transmitter.TransmitTo(optionTo),
			transmitter.LoadOnly(),
		}

		log.Debug().
			Interface("player", player).
			Msg("Waiting for player to respond to ping.")

		if _, err = ping(player.Port); err != nil {
			return err
		}

		transmitter := transmitter.OSCTransmitter{Port: player.Port}

		if err := transmitter.TransmitScore(score, transmitOpts...); err != nil {
			return err
		}

		log.Info().
			Interface("player", player).
			Msg("Transmitted score to player.")

		// When no output filename is specified, we write the result to stdout. But
		// first, we need to ask the player to export to a file, so we use a
		// temporary file.
		tmpFilename := ""
		if outputFilename == "" {
			tmpdir, err := ioutil.TempDir("", "alda-export")
			if err != nil {
				return err
			}

			tmpFilename = filepath.Join(
				tmpdir, fmt.Sprintf(
					"export-%d-%d.mid",
					time.Now().Unix(),
					rand.Intn(10000),
				),
			)
		}

		var targetFilename string
		if outputFilename != "" {
			targetFilename = outputFilename
		} else {
			targetFilename = tmpFilename
		}

		err = transmitter.TransmitMidiExportMessage(targetFilename)
		if err != nil {
			return err
		}

		log.Info().
			Interface("player", player).
			Msg("Transmitted \"export\" message.")

		log.Debug().
			Str("targetFilename", targetFilename).
			Msg("Waiting for target file to be written.")

		// There can be a noticeable delay while we wait for the player process to
		// finish writing the target file. So, we display a message here to give the
		// user some incremental feedback and avoid making it look like Alda is
		// "hanging."
		fmt.Fprintln(os.Stderr, "Exporting...")

		var midiFile *os.File

		if err := util.Await(
			func() error {
				mf, err := os.Open(targetFilename)
				if err != nil {
					return err
				}

				midiFile = mf
				return nil
			},
			midiExportTimeout,
		); err != nil {
			return nil
		}

		if outputFilename != "" {
			fmt.Fprintf(os.Stderr, "Exported score to %s\n", outputFilename)
		} else {
			if _, err := io.Copy(os.Stdout, midiFile); err != nil {
				return err
			}
		}

		return nil
	},
}
