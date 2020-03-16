package cmd

import (
	"fmt"
	"os"

	log "alda.io/client/logging"
	"github.com/spf13/cobra"
)

var verbosity int

func init() {
	rootCmd.PersistentFlags().IntVarP(
		&verbosity, "verbosity", "v", 1, "verbosity level (0-3)",
	)

	for _, cmd := range []*cobra.Command{
		doctorCmd,
		playCmd,
		psCmd,
		versionCmd,
	} {
		rootCmd.AddCommand(cmd)
	}
}

var rootCmd = &cobra.Command{
	Use:   "alda",
	Short: "alda: a text-based language for music composition",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch verbosity {
		case 0:
			log.SetGlobalLevel("error")
		case 1:
			log.SetGlobalLevel("warn")
		case 2:
			log.SetGlobalLevel("info")
		case 3:
			log.SetGlobalLevel("debug")
		default:
			fmt.Println("Invalid verbosity level. Valid values are 0-3.")
			os.Exit(1)
		}
	},
}

// Execute parses command-line arguments and runs the Alda command-line client.
func Execute() error {
	// Regardless of the command being run*, Alda will preemptively spawn player
	// processes in the background, up to a desired amount. This helps to ensure
	// that the application will feel fast, because each time you need a player
	// process, there will probably already be one available.
	//
	// *`alda ps` is an exception because it is designed to be run repeatedly,
	// e.g. `watch -n 0.25 alda ps`, in order to provide a live-updating view of
	// current Alda processes.
	commandIsProbablyPs := false

	for _, arg := range os.Args {
		if arg == "ps" {
			commandIsProbablyPs = true
		}
	}

	if !commandIsProbablyPs {
		if err := fillPlayerPool(); err != nil {
			log.Warn().Err(err).Msg("Failed to fill player pool.")
		}
	}

	return rootCmd.Execute()
}
