package cmd

import (
	"fmt"
	"time"

	"alda.io/client/system"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List background processes",
	RunE: func(_ *cobra.Command, _ []string) error {
		playerStates, err := system.ReadPlayerStates()
		if err != nil {
			return err
		}

		replServerStates, err := system.ReadREPLServerStates()
		if err != nil {
			return err
		}

		fmt.Println("id\tport\tstate\texpiry\ttype\tpid")

		for _, state := range playerStates {
			expiry := humanize.Time(time.Unix(state.Expiry/1000, 0))

			fmt.Printf(
				"%s\t%d\t%s\t%s\t%s\t%d\n",
				state.ID, state.Port, state.State, expiry, "player", state.PID,
			)
		}

		for _, state := range replServerStates {
			fmt.Printf(
				"%s\t%d\t%s\t%s\t%s\t%d\n",
				state.ID, state.Port, "-", "-", "repl-server", state.PID,
			)
		}

		return nil
	},
}
