package cmd

import (
	"fmt"
	"os"
	"time"

	"alda.io/client/system"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List background processes",
	Run: func(_ *cobra.Command, args []string) {
		states, err := system.ReadPlayerStates()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("id\tport\tstate\texpiry")

		for _, state := range states {
			if state.ReadError != nil {
				fmt.Fprintln(os.Stderr, state.ReadError)
				continue
			}

			expiry := humanize.Time(time.Unix(state.Expiry/1000, 0))

			fmt.Printf(
				"%s\t%d\t%s\t%s\n", state.ID, state.Port, state.State, expiry,
			)
		}
	},
}
