package cmd

import (
	log "alda.io/client/logging"
	"github.com/spf13/cobra"
)

var verbose bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false, "verbose output",
	)

	for _, cmd := range []*cobra.Command{
		doctorCmd,
		playCmd,
		versionCmd,
	} {
		rootCmd.AddCommand(cmd)
	}
}

var rootCmd = &cobra.Command{
	Use:   "alda",
	Short: "alda: a text-based language for music composition",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO: add more control over log levels. maybe the default log level
		// should be warn, -v should be info, and -vv should be debug?
		if verbose {
			log.SetGlobalLevel("debug")
		} else {
			log.SetGlobalLevel("info")
		}
	},
}

// Execute parses command-line arguments and runs the Alda command-line client.
func Execute() error {
	return rootCmd.Execute()
}
