package cmd

import (
	"alda.io/client/help"
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
	Run: func(_ *cobra.Command, args []string) {
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
			help.ExitOnError(err)
			players = append(players, player)
		// Send a "stop" message to all known players.
		default:
			knownPlayers, err := system.ReadPlayerStates()
			help.ExitOnError(err)
			players = append(players, knownPlayers...)
		}

		for _, player := range players {
			transmitter := transmitter.OSCTransmitter{Port: player.Port}
			help.ExitOnError(transmitter.TransmitStopMessage())

			log.Info().
				Interface("player", player).
				Msg("Sent \"stop\" message to player process.")
		}
	},
}
