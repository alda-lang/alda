package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VERSION is the current version of the `alda` client.
const VERSION = "1.99.0" // FIXME

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Alda version information",
	Run: func(_ *cobra.Command, args []string) {
		fmt.Printf("alda %s\n", VERSION)
	},
}
