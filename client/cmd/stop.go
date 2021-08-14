package cmd

import (
	"fmt"
	"os"

	log "alda.io/client/logging"
	"alda.io/client/system"
	"alda.io/client/transmitter"
	"github.com/spf13/cobra"
)

func init() {
	stopCmd.Flags().StringVarP(
		&playerID,
		"player-id",
		"i",
		"",
		"The ID of the player process to tell to stop playback",
	)

	stopCmd.Flags().IntVarP(
		&playerPort,
		"port",
		"p",
		-1,
		"The port of the player process to tell to stop playback",
	)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop playback",
	RunE: func(_ *cobra.Command, args []string) error {
		// It's a good idea to print something here to indicate to the user that we
		// did what they asked. Otherwise, it might not be obvious that we did
		// anything.
		fmt.Fprintln(os.Stderr, "Stopping playback.")

		players := []system.PlayerState{}

		// Determine the players to which to send a "stop" message based on the
		// provided CLI options.
		switch {
		// Port is explicitly specified, so use that port.
		case playerPort != -1:
			players = append(players, system.PlayerState{
				ID: "unknown", State: "unknown", Port: playerPort,
			})
		// Player ID is specified; look up the player by ID and use its port.
		case playerID != "":
			player, err := system.FindPlayerByID(playerID)
			if err != nil {
				return err
			}
			players = append(players, player)
		// Send a "stop" message to all known players.
		default:
			knownPlayers, err := system.ReadPlayerStates()
			if err != nil {
				return err
			}
			players = append(players, knownPlayers...)
		}

		for _, player := range players {
			transmitter := transmitter.OSCTransmitter{Port: player.Port}
			if err := transmitter.TransmitStopMessage(); err != nil {
				log.Warn().
					Interface("player", player).
					Err(err).
					Msg("Failed to send \"stop\" message to player process.")
			} else {
				log.Info().
					Interface("player", player).
					Msg("Sent \"stop\" message to player process.")
			}
		}
		return nil
	},
}
