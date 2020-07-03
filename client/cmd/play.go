package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"time"

	"alda.io/client/emitter"
	"alda.io/client/generated"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/spf13/cobra"
)

var playerID string
var port int
var file string
var code string
var playFrom string
var playTo string

func init() {
	playCmd.Flags().StringVarP(
		&playerID, "player-id", "i", "", "The ID of the player process to use",
	)

	playCmd.Flags().IntVarP(
		&port, "port", "p", -1, "The port of the player process to use",
	)

	playCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read Alda source code from a file",
	)

	playCmd.Flags().StringVarP(
		&code, "code", "c", "", "Supply Alda source code as a string",
	)

	playCmd.Flags().StringVarP(
		&playFrom,
		"from",
		"F",
		"",
		"A time marking (e.g. 0:30) or marker from which to start playback",
	)

	playCmd.Flags().StringVarP(
		&playTo,
		"to",
		"T",
		"",
		"A time marking (e.g. 1:00) or marker at which to end playback",
	)
}

var errNoPlayersAvailable = fmt.Errorf("no players available")

func findAvailablePlayer() (playerState, error) {
	players, err := readPlayerStates()
	if err != nil {
		return playerState{}, err
	}

	for _, player := range players {
		if player.State == "ready" {
			return player, nil
		}
	}

	return playerState{}, errNoPlayersAvailable
}

func findPlayerByID(id string) (playerState, error) {
	players, err := readPlayerStates()
	if err != nil {
		return playerState{}, err
	}

	for _, player := range players {
		if player.ID == id {
			return player, nil
		}
	}

	return playerState{}, fmt.Errorf("player not found: %s", id)
}

func spawnPlayer() error {
	aldaPlayer, err := exec.LookPath("alda-player")
	if err != nil {
		return err
	}

	// First, we run `alda-player info` and parse the version number from the
	// output, so that we can confirm that the player is the same version as the
	// client.
	infoCmd := exec.Command(aldaPlayer, "info")
	infoCmd.Stdout = nil
	infoCmd.Stderr = nil
	outputBytes, err := infoCmd.Output()
	if err != nil {
		return err
	}

	output := string(outputBytes)

	re := regexp.MustCompile(`alda-player (.*)`)
	captured := re.FindStringSubmatch(output)
	if len(captured) < 2 {
		return fmt.Errorf(
			"unable to parse player version from output: %s", output,
		)
	}
	// captured[0] is "alda-player X.X.X", captured[1] is "X.X.X"
	playerVersion := captured[1]

	// TODO: If the player version is different from the client version, offer to
	// download and install the correct player version.
	if playerVersion != generated.ClientVersion {
		return fmt.Errorf(
			"client version is %s, but player version is %s",
			generated.ClientVersion, playerVersion,
		)
	}

	// Once we've confirmed that the client and player version are the same, we
	// run `alda-player run` to start the player process.
	runCmd := exec.Command(aldaPlayer, "run")
	if err := runCmd.Start(); err != nil {
		return err
	}

	log.Info().Msg("Spawned player process.")

	return nil
}

func fillPlayerPool() error {
	players, err := readPlayerStates()
	if err != nil {
		return err
	}

	availablePlayers := 0
	for _, player := range players {
		if player.State == "ready" || player.State == "starting" {
			availablePlayers++
		}
	}

	desiredAvailablePlayers := 2
	playersToStart := desiredAvailablePlayers - availablePlayers

	log.Debug().
		Int("availablePlayers", availablePlayers).
		Int("desiredAvailablePlayers", desiredAvailablePlayers).
		Int("playersToStart", playersToStart).
		Msg("Spawning players.")

	results := []<-chan error{}

	for i := 0; i < playersToStart; i++ {
		result := make(chan error)
		results = append(results, result)
		go func() { result <- spawnPlayer() }()
	}

	for _, result := range results {
		err := <-result
		if err != nil {
			return err
		}
	}

	return nil
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

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		score := model.NewScore()
		start := time.Now()
		if err := score.Update(scoreUpdates...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Info().
			Int("updates", len(scoreUpdates)).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Constructed score.")

		var players []playerState

		// Determine the port to use based on the provided CLI options.
		switch {

		// Port is explicitly specified, so use that port.
		case port != -1:
			player := playerState{ID: "unknown", State: "unknown", Port: port}
			players = []playerState{player}

		// Player ID is specified; look up the player by ID and use its port.
		case playerID != "":
			player, err := findPlayerByID(playerID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			players = []playerState{player}

		// We're actually unpausing, not playing, so send the message to all active
		// player processes so that if any of them are paused, they'll resume
		// playing.
		case action == "unpause":
			allPlayers, err := readPlayerStates()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			players = []playerState{}
			for _, player := range allPlayers {
				if player.State == "active" {
					players = append(players, player)
				}
			}

		// Find an available player process to use.
		default:
			if err := await(
				func() error {
					player, err := findAvailablePlayer()
					if err != nil {
						return err
					}

					players = []playerState{player}
					return nil
				},
				reasonableTimeout,
			); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		emitOpts := []emitter.EmissionOption{
			emitter.EmitFrom(playFrom),
			emitter.EmitTo(playTo),
		}

		if action == "play" {
			emitOpts = append(emitOpts, emitter.OneOff())
		}

		log.Info().
			Interface("players", players).
			Str("action", action).
			Msg("Sending messages to players.")

		for _, player := range players {
			log.Debug().
				Interface("player", player).
				Msg("Waiting for player to respond to ping.")

			if _, err := ping(player.Port); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			emitter := emitter.OSCEmitter{Port: player.Port}

			var emissionError error

			if action == "unpause" {
				emissionError = emitter.EmitPlayMessage()
			} else {
				emissionError = emitter.EmitScore(score, emitOpts...)
			}
			if emissionError != nil {
				fmt.Println(emissionError)
				os.Exit(1)
			}

			log.Info().
				Interface("player", player).
				Msg("Sent OSC messages to player.")
		}
	},
}
