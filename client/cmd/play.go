package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"alda.io/client/emitter"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/spf13/cobra"
)

var playerID string
var port int
var file string

func init() {
	playCmd.Flags().StringVarP(
		&playerID, "player-id", "i", "", "The ID of the Alda player process to use",
	)

	playCmd.Flags().IntVarP(
		&port, "port", "p", -1, "The port of the Alda player process to use",
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
		if player.Condition == "new" {
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
		if player.id == id {
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

	fmt.Println("Starting player process...")

	cmd := exec.Command(aldaPlayer, "run")
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func findOrSpawnPlayer() (playerState, error) {
	player, err := findAvailablePlayer()

	// Found an available player. Success!
	if err == nil {
		return player, nil
	}

	// There was an unexpected error while trying to find a player.
	if err != errNoPlayersAvailable {
		return playerState{}, err
	}

	// No players available, so spawn one and check again.
	if err := spawnPlayer(); err != nil {
		return playerState{}, err
	}

	if err := await(
		func() error {
			p, err := findAvailablePlayer()
			if err != nil {
				return err
			}

			player = p
			return nil
		},
		reasonableTimeout,
	); err != nil {
		return playerState{}, err
	}

	return player, nil
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
		if err := score.Update(scoreUpdates...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

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
		// Find an available player process to use, spawning one if necessary.
		default:
			player, err := findOrSpawnPlayer()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			port = player.Port
		}

		fmt.Println("Waiting for player to respond to ping...")
		if _, err := ping(port); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Sending OSC messages to player on port: %d\n", port)
		if err := (emitter.OSCEmitter{Port: port}).EmitScore(score); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
