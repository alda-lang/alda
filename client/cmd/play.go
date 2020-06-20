package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"alda.io/client/emitter"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/spf13/cobra"
)

var playerID string
var port int
var file string

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
	// TODO: Make this flag optional. Instead, allow input to be provided either
	// as a file (-f / --file), as a string (-c / --code), or via STDIN.
	playCmd.MarkFlagRequired("file")
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
	if playerVersion != VERSION {
		return fmt.Errorf(
			"client version is %s, but player version is %s",
			VERSION, playerVersion,
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

	for i := 0; i < playersToStart; i++ {
		err := spawnPlayer()
		if err != nil {
			return err
		}
	}

	return nil
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Evaluate and play Alda source code",
	Run: func(_ *cobra.Command, args []string) {
		scoreUpdates, err := parser.ParseFile(file)
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

		// Determine the port to use based on the provided CLI options.
		switch {
		// Nothing to do; port is explicitly specified.
		case port != -1:
		// Player ID is specified; look up the player by ID and use its port.
		case playerID != "":
			player, err := findPlayerByID(playerID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			port = player.Port
		// Find an available player process to use.
		default:
			if err := await(
				func() error {
					player, err := findAvailablePlayer()
					if err != nil {
						return err
					}

					port = player.Port
					return nil
				},
				reasonableTimeout,
			); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		log.Debug().
			Int("port", port).
			Msg("Waiting for player to respond to ping.")

		if _, err := ping(port); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := (emitter.OSCEmitter{Port: port}).EmitScore(score); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.Info().
			Int("port", port).
			Msg("Sent OSC messages to player.")
	},
}
