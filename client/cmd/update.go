package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var assumeYes bool
var desiredVersion string

func init() {
	updateCmd.Flags().BoolVarP(
		&assumeYes,
		"yes",
		"y",
		false,
		"Do not prompt for confirmation before updating",
	)

	updateCmd.Flags().StringVar(
		&desiredVersion,
		"version",
		"",
		"The version to update to (default: latest)",
	)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to the latest version of Alda",
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Println("TODO: implement `alda update`")
		return nil
	},
}
