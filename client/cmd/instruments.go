package cmd

import (
	"fmt"
	"strings"

	"alda.io/client/model"
	"github.com/spf13/cobra"
)

var instrumentsCmd = &cobra.Command{
	Use:   "instruments",
	Short: "Display the list of available instruments",
	RunE: func(_ *cobra.Command, args []string) error {
		fmt.Println(strings.Join(model.InstrumentsList(), "\n"))
		return nil
	},
}
