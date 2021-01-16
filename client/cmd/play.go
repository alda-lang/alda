package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/transmitter"
	"alda.io/client/util"
	"github.com/spf13/cobra"
)

var playerID string
var playerPort int
var file string
var code string
var optionFrom string
var optionTo string

func init() {
	playCmd.Flags().StringVarP(
		&playerID, "player-id", "i", "", "The ID of the player process to use",
	)

	playCmd.Flags().IntVarP(
		&playerPort, "port", "p", -1, "The port of the player process to use",
	)

	playCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read Alda source code from a file",
	)

	playCmd.Flags().StringVarP(
		&code, "code", "c", "", "Supply Alda source code as a string",
	)

	playCmd.Flags().StringVarP(
		&optionFrom,
		"from",
		"F",
		"",
		"A time marking (e.g. 0:30) or marker from which to start playback",
	)

	playCmd.Flags().StringVarP(
		&optionTo,
		"to",
		"T",
		"",
		"A time marking (e.g. 1:00) or marker at which to end playback",
	)
}

// Returns true if input is being piped into stdin.
//
// Returns an error if something went wrong while trying to determine this.
func isInputBeingPipedIn() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	return stat.Mode()&os.ModeCharDevice == 0, nil
}

var errNoInputSupplied = fmt.Errorf("no input supplied")

// Reads all bytes piped into stdin and returns them.
//
// Returns the error `errNoInputSupplied` if no input is being piped in, or a
// different error if something else went wrong.
func readStdin() ([]byte, error) {
	isInputSupplied, err := isInputBeingPipedIn()
	if err != nil {
		return nil, err
	}

	if !isInputSupplied {
		return nil, errNoInputSupplied
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Parses Alda source code piped into stdin and returns the parsed score
// updates.
//
// Returns `errNoInputSupplied` if no input is being piped into stdin.
//
// Returns a different error if the input couldn't be parsed as valid Alda code,
// or if something else went wrong.
func parseStdin() ([]model.ScoreUpdate, error) {
	bytes, err := readStdin()
	if err != nil {
		return nil, err
	}

	return parser.ParseString(string(bytes))
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Evaluate and play Alda source code",
	Long: `Evaluate and play Alda source code

---

Source code can be provided in one of three ways:

The path to a file (-f, --file):
  alda play -f path/to/my-score.alda

A string of code (-c, --code):
  alda play -c "harpsichord: o5 d+8 < b g+ e d+1"

Text piped into the process on stdin:
  echo "glockenspiel: o5 g8 < g > g e4 d4." | alda play

---`,
	Run: func(_ *cobra.Command, args []string) {
		var scoreUpdates []model.ScoreUpdate
		var err error

		// If no input Alda code is provided, then we treat the `alda play` command
		// as an "unpause" command. We will send a bundle that just contains a
		// "play" message to all player processes, and if any of them are in a
		// "paused" state (i.e. they were playing something and then they got paused
		// via `alda stop`), they will resume playback from where they left off.
		action := "play"

		switch {
		case file != "":
			scoreUpdates, err = parser.ParseFile(file)

		case code != "":
			scoreUpdates, err = parser.ParseString(code)

		default:
			scoreUpdates, err = parseStdin()
			if err == errNoInputSupplied {
				action = "unpause"
				err = nil
			}
		}

		// Errors with source context are presented to the user as-is.
		//
		// TODO: See TODO comment in cmd/parse.go about writing better user-facing
		// error messages.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		help.ExitOnError(err)

		score := model.NewScore()
		start := time.Now()
		err = score.Update(scoreUpdates...)

		// Errors with source context are presented to the user as-is.
		//
		// TODO: See TODO comment in cmd/parse.go about writing better user-facing
		// error messages.
		switch err.(type) {
		case *model.AldaSourceError:
			err = &help.UserFacingError{Err: err}
		}

		help.ExitOnError(err)

		log.Info().
			Int("updates", len(scoreUpdates)).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Constructed score.")

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
			help.ExitOnError(err)
			players = []system.PlayerState{player}

		// We're actually unpausing, not playing, so send the message to all active
		// player processes so that if any of them are paused, they'll resume
		// playing.
		case action == "unpause":
			allPlayers, err := system.ReadPlayerStates()
			help.ExitOnError(err)
			players = []system.PlayerState{}
			for _, player := range allPlayers {
				if player.State == "active" {
					players = append(players, player)
				}
			}

		// Find an available player process to use.
		default:
			system.StartingPlayerProcesses()

			help.ExitOnError(
				util.Await(
					func() error {
						player, err := system.FindAvailablePlayer()
						if err != nil {
							return err
						}

						players = []system.PlayerState{player}
						return nil
					},
					reasonableTimeout,
				),
			)
		}

		transmitOpts := []transmitter.TransmissionOption{
			transmitter.TransmitFrom(optionFrom),
			transmitter.TransmitTo(optionTo),
		}

		if action == "play" {
			transmitOpts = append(transmitOpts, transmitter.OneOff())
		}

		log.Info().
			Interface("players", players).
			Str("action", action).
			Msg("Sending messages to players.")

		for _, player := range players {
			log.Debug().
				Interface("player", player).
				Msg("Waiting for player to respond to ping.")

			_, err := ping(player.Port)
			help.ExitOnError(err)

			transmitter := transmitter.OSCTransmitter{Port: player.Port}

			var transmissionError error
			if action == "unpause" {
				transmissionError = transmitter.TransmitPlayMessage()
			} else {
				transmissionError = transmitter.TransmitScore(score, transmitOpts...)
			}
			help.ExitOnError(transmissionError)

			log.Info().
				Interface("player", player).
				Msg("Sent OSC messages to player.")
		}

		// We don't have to print something here, but it's a good idea because it
		// indicates to the user that we did what they asked. Otherwise, it might
		// not be obvious that we did anything, especially in cases where there is
		// no audible output, e.g. `alda play -c "c d e"` (valid syntax, but no
		// audible output because no part was indicated).
		fmt.Fprintln(os.Stderr, "Playing...")
	},
}
