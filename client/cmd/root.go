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
	return rootCmd.Execute()
}
