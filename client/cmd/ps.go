package cmd

import (
	"fmt"
	"time"

	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List background processes",
	Run: func(_ *cobra.Command, args []string) {
		states, err := system.ReadPlayerStates()
		help.ExitOnError(err)

		fmt.Println("id\tport\tstate\texpiry")

		for _, state := range states {
			if state.ReadError != nil {
				log.Warn().Err(state.ReadError).Msg("Failed to read player state")
				continue
			}

			expiry := humanize.Time(time.Unix(state.Expiry/1000, 0))

			fmt.Printf(
				"%s\t%d\t%s\t%s\n", state.ID, state.Port, state.State, expiry,
			)
		}
	},
}
