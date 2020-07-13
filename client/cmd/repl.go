package cmd

import (
	"fmt"
	"os"

	"alda.io/client/repl"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an Alda REPL client/server",
	Run: func(_ *cobra.Command, args []string) {
		if err := repl.RunClient(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
