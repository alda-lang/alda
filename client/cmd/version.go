package cmd

import (
	"fmt"

	"alda.io/client/generated"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Alda version information",
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Printf("alda %s\n", generated.ClientVersion)
		return nil
	},
}
