package cmd

import (
	"fmt"
	"time"

	log "alda.io/client/logging"
	"alda.io/client/system"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List background processes",
	RunE: func(_ *cobra.Command, args []string) error {
		states, err := system.ReadPlayerStates()
		if err != nil {
			return err
		}

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

		return nil
	},
}
