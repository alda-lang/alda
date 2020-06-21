package cmd

import (
	"fmt"

	"alda.io/client/generated"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Alda version information",
	Run: func(_ *cobra.Command, args []string) {
		fmt.Printf("alda %s\n", generated.ClientVersion)
	},
}
