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

var replHost string
var replPort int
var startREPLClient bool
var startREPLServer bool

func init() {
	replCmd.Flags().StringVarP(
		&replHost, "host", "H", "localhost", "The hostname of the Alda REPL server",
	)

	replCmd.Flags().IntVarP(
		&replPort, "port", "p", -1, "The port of the Alda REPL server",
	)

	replCmd.Flags().BoolVarP(
		&startREPLClient, "client", "c", false, "Start an Alda REPL client session",
	)

	replCmd.Flags().BoolVarP(
		&startREPLServer, "server", "s", false, "Start an Alda REPL server",
	)
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an Alda REPL client/server",
	Long: `Start an Alda REPL client/server

---

Examples:

  alda repl
    Equivalent to ` + "`alda repl --client --server`" + `.

  alda repl --client --server
    Starts an interactive REPL session where the server is running in the
    background.

  alda repl --client --port 12345
    Starts an interactive REPL session that connects to an existing Alda REPL
    server running on port 12345.

  alda repl --server --port 12345
    Starts an Alda REPL server without an interactive prompt. Clients can then
    connect by running ` + "`alda repl --client --port 12345`" + `.

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		// If both --client and --server are omitted, run them both.
		if !startREPLClient && !startREPLServer {
			startREPLClient = true
			startREPLServer = true
		}

		if startREPLServer {
			// If --port isn't specified, pick an arbitrary port that's available.
			if replPort == -1 {
				port, err := system.FindOpenPort()
				if err != nil {
					return err
				}
				replPort = port
			}

			server, err := repl.RunServer(replPort)
			if err != nil {
				return err
			}

			// Ensure that the server is closed on normal exit.
			defer server.Close()

			// Ensure that the server is closed if the process is interrupted or
			// terminated.
			//
			// Depending on whether we are running in server-only mode or not, we
			// either wait for this interrupt signal in the foreground or the
			// background.
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

			if startREPLClient {
				go func() {
					<-signals
					server.Close()
				}()
			} else {
				<-signals
				server.Close()
			}
		}

		if startREPLClient {
			if replPort == -1 {
				// TODO: user-facing error
				return fmt.Errorf(
					"You must specify the --port of a running Alda REPL server.",
				)
			}

			return repl.RunClient(replHost, replPort)
		}

		return nil
	},
}
