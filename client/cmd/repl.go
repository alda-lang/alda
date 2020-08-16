package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"alda.io/client/repl"
	"alda.io/client/system"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an Alda REPL client/server",
	Run: func(_ *cobra.Command, args []string) {
		// TODO:
		// * Add --client, --server, --host and --port options
		// * Only start a server if it makes sense to do so based on the provided
		//   options.
		// * Only find an open port if we're starting a server and the user didn't
		//   specify what port to serve on.
		port, err := system.FindOpenPort()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		server, err := repl.RunServer(port)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Ensure that the server is closed on normal exit.
		defer server.Close()

		// Ensure that the server is closed if the process is interrupted or
		// terminated.
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			server.Close()
		}()

		if err := repl.RunClient("localhost", port); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
