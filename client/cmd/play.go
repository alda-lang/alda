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

var port int
var file string

func init() {
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
	states, err := readPlayerStates()
	if err != nil {
		return playerState{}, err
	}

	for _, state := range states {
		if state.Condition == "new" {
			return state, nil
		}
	}

	return playerState{}, errNoPlayersAvailable
}

func spawnPlayerProcess() error {
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

		// If no port is specified, the client will find an available player process
		// to use, spawning one if necessary.
		if port == -1 {
			player, err := findAvailablePlayer()

			if err == errNoPlayersAvailable {
				if err := spawnPlayerProcess(); err != nil {
					fmt.Println(err)
					os.Exit(1)
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
					fmt.Println(err)
					os.Exit(1)
				}
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
