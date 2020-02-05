package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Alda version information",
	Run: func(_ *cobra.Command, args []string) {
		fmt.Println("TODO: print version information")
	},
}
