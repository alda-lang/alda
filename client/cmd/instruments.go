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
	Run: func(_ *cobra.Command, args []string) {
		fmt.Println(strings.Join(model.InstrumentsList(), "\n"))
	},
}
