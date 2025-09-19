package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"alda.io/client/color"
	"alda.io/client/help"
	"alda.io/client/json"
	"alda.io/client/repl"
	"alda.io/client/system"
	"github.com/spf13/cobra"
)

var replHost string
var replPort int
var startREPLClient bool
var startREPLServer bool
var replMessage string

func init() {
	replCmd.Flags().StringVarP(
		&replHost, "host", "H", "127.0.0.1", "The hostname of the Alda REPL server",
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

	replCmd.Flags().StringVarP(
		&replMessage,
		"message",
		"m",
		"",
		"A JSON nREPL message to send to the server",
	)
}

func errInvalidNREPLMessage(message string) error {
	return help.UserFacingErrorf(
		`Invalid nREPL message:

  %s

Here is an example of a valid nREPL message:

  %s`,
		color.Aurora.BgRed(message),
		color.Aurora.Bold(`{"op": "eval-and-play", "code": "banjo: c"}`),
	)
}

func sendREPLMessage(host string, port int, message string) error {
	parsed, err := json.ParseJSON([]byte(message))
	if err != nil {
		return errInvalidNREPLMessage(message)
	}

	msg, ok := parsed.Data().(map[string]interface{})
	if !ok {
		return errInvalidNREPLMessage(message)
	}

	// This is just a basic check that the provided JSON is an object that
	// contains an "op" entry. An invalid message will fail below anyway, but it's
	// nice if we can recognize earlier that the input is bad and tell the user in
	// a more useful way.
	if _, hasOp := msg["op"]; !hasOp {
		return errInvalidNREPLMessage(message)
	}

	res, err := repl.SendMessage(host, port, msg)
	if err != nil {
		return err
	}

	fmt.Println(json.ToJSON(res))

	if len(repl.ResponseErrors(res)) > 0 {
		return help.UserFacingErrorf(
			`The Alda REPL server indicated that your request was unsuccessful.

See the response output for details.`,
		)
	}

	return nil
}

var errREPLServerPortUnspecified = help.UserFacingErrorf(
	`You must specify the port of the Alda REPL server that you wish to connect to.

For example, if you've started an Alda REPL server on port 12345, you can
connect to it by running:

  %s

See %s for more information about starting Alda REPL servers and
clients.`,
	color.Aurora.BrightYellow("alda repl --client --port 12345"),
	color.Aurora.BrightYellow("alda repl --help"),
)

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

  alda repl --port 12345 --message '{"op": "eval-and-play", "code": "banjo: c"}'
    Sends an nREPL message to the Alda REPL server running on port 12345.
    This is mainly useful for writing scripts and tools for working with Alda.

---`,
	RunE: func(_ *cobra.Command, args []string) error {
		if replMessage != "" {
			if replPort == -1 {
				return errREPLServerPortUnspecified
			}

			return sendREPLMessage(replHost, replPort, replMessage)
		}

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
				return errREPLServerPortUnspecified
			}

			return repl.RunClient(replHost, replPort)
		}

		return nil
	},
}
