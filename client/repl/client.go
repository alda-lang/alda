// vim: tabstop=2

package repl

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"alda.io/client/generated"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"alda.io/client/util"
	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/google/uuid"
	bencode "github.com/jackpal/bencode-go"
	"github.com/logrusorgru/aurora"
)

const aldaASCIILogo = `
 █████╗ ██╗     ██████╗  █████╗
██╔══██╗██║     ██╔══██╗██╔══██╗
███████║██║     ██║  ██║███████║
██╔══██║██║     ██║  ██║██╔══██║
██║  ██║███████╗██████╔╝██║  ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝
`

const aldaVersionText = `
             ` + generated.ClientVersion + `
         repl session
`

const serverConnectTimeout = 5 * time.Second

// Client is a stateful Alda REPL client object.
type Client struct {
	// The TCP address of the server with which the client will communicate.
	serverAddr *net.TCPAddr
	// The current connection to the server with which the client is
	// communicating.
	serverConn net.Conn
	// The current session ID that is sent to the server on every request.
	// Following the nREPL protocol, the client makes an initial "clone" request
	// to the server and the response from the server contains the session ID that
	// the client will use for the rest of the session.
	sessionID string
}

type replCommand struct {
	helpSummary string
	helpDetails string
	run         func(client *Client, args string) error
}

// TODO: implement the rest for parity with v1
var replCommands = map[string]replCommand{
	"help": {
		helpSummary: "Display this help text.",
		helpDetails: `Usage:
  :help help
  :help new`,
		run: func(client *Client, args string) error {
			return fmt.Errorf("not yet implemented")
		}},

	"play": {
		helpSummary: "Plays the current score.",
		helpDetails: `Can take optional ` + "`from`" + `and ` + "`to`" + `arguments, in the form of markers or mm:ss
times.

Without arguments, will play the entire score from beginning to end.

Example usage:

  :play
  :play from 0:05
  :play to 0:10
  :play from 0:05 to 0:10
  :play from guitarIn
  :play to verse
  :play from verse to bridge`,
		run: func(client *Client, argsString string) error {
			args, err := shlex.Split(argsString)
			if err != nil {
				return err
			}

			errInvalidArgs := fmt.Errorf("invalid arguments: %#v", args)

			req := map[string]interface{}{"op": "replay"}

			for i := 0; i < len(args); i++ {
				// If this is the last argument, that means there are an odd number of
				// arguments, which is invalid because we are expecting an even number
				// of arguments.
				if i == len(args)-1 {
					return errInvalidArgs
				}

				switch args[i] {
				case "from":
					_, hit := req["from"]
					if hit {
						return errInvalidArgs
					}
					i++
					req["from"] = args[i]
				case "to":
					_, hit := req["to"]
					if hit {
						return errInvalidArgs
					}
					i++
					req["to"] = args[i]
				default:
					return errInvalidArgs
				}
			}

			res, err := client.sendRequest(req)
			if err != nil {
				return err
			}

			printResponseErrors(res)

			return nil
		}},
}

func (client *Client) handleCommand(name string, args string) error {
	command, defined := replCommands[name]
	if !defined {
		return fmt.Errorf("unrecognized command: %s", name)
	}

	return command.run(client, args)
}

// Disconnect closes the client's connection with the server.
func (client *Client) Disconnect() {
	if client.serverConn != nil {
		client.serverConn.Close()
	}
}

func (client *Client) connect() error {
	// First, close any existing connection to avoid a memory leak.
	client.Disconnect()

	if err := util.Await(
		func() error {
			conn, err := net.DialTCP("tcp", nil, client.serverAddr)
			if err != nil {
				return err
			}

			client.serverConn = conn
			return nil
		},
		serverConnectTimeout,
	); err != nil {
		return err
	}

	return nil
}

// Sends a request as a bencoded payload to the server, awaits a response from
// the server, and returns the bdecoded response.
//
// Returns an error if there is some networking problem, if the request can't be
// bencoded, if the response can't be bdecoded, or if the response CAN be
// bdecoded but the resulting data structure is not of the type we expect.
//
// NOTE: The nREPL design actually specifies that a server may send more than
// one response. The last response's "status" value (a list) typically includes
// "done", but even then, the design specifies that a server may continue to
// send additional responses after that.
//
// This will need to be refactored if/when we want the server to send multiple
// responses to a request. Perhaps this function could return a `chan
// map[string]interface{}` and it could continuously receive responses from the
// server and put them onto the channel until it encounters a response whose
// "status" value includes "done". At the moment, we don't have that need, so
// we're keeping it simple.
func (client *Client) sendRequest(
	req map[string]interface{},
) (map[string]interface{}, error) {
	if client.sessionID != "" {
		req["session"] = client.sessionID
	}

	messageID := uuid.New().String()
	req["id"] = messageID

	log.Debug().Interface("request", req).Msg("Sending request.")

	if err := bencode.Marshal(client.serverConn, req); err != nil {
		return nil, err
	}

	// Avoid hanging forever if the server doesn't respond.
	client.serverConn.SetReadDeadline(time.Now().Add(30 * time.Second))

	response, err := bencode.Decode(client.serverConn)
	if err != nil {
		return nil, err
	}

	switch response.(type) {
	case map[string]interface{}: // OK to continue
	default:
		return nil,
			fmt.Errorf("response could not be decoded into the expected type")
	}
	res := response.(map[string]interface{})

	log.Debug().Interface("response", res).Msg("Received response.")

	if res["id"] != messageID {
		return nil,
			fmt.Errorf(
				"unexpected \"id\" in response. Expected %s, got %s",
				messageID,
				res["id"],
			)
	}

	return res, nil
}

func printResponseErrors(res map[string]interface{}) {
	switch res["status"].(type) {
	case []interface{}: // OK to proceed
	default:
		fmt.Printf("ERROR: %#v\n", res)
		return
	}
	statuses := res["status"].([]interface{})

	errorStatus := false
	for _, status := range statuses {
		if status == "error" {
			errorStatus = true
		}
	}
	if !errorStatus {
		return
	}

	switch res["problems"].(type) {
	case []interface{}: // OK to proceed
	default:
		fmt.Printf("ERROR: %#v\n", res)
		return
	}
	problems := res["problems"].([]interface{})

	switch len(problems) {
	case 0:
		fmt.Printf("ERROR: %#v\n", res)
	case 1:
		switch problem := problems[0].(type) {
		case string:
			fmt.Printf("ERROR: %s\n", problem)
		default:
			fmt.Printf("ERROR: %#v\n", problem)
		}
	default:
		fmt.Println("ERRORS:")
		for _, problem := range problems {
			switch p := problem.(type) {
			case string:
				fmt.Println(p)
			default:
				fmt.Printf("%#v\n", p)
			}
		}
	}
}

var replHistoryFilepath = system.CachePath("history", "alda-repl-history")

// NewClient returns an initialized instance of an Alda REPL client.
func NewClient(host string, port int) (*Client, error) {
	addr, err := net.ResolveTCPAddr(
		"tcp",
		fmt.Sprintf("%s:%d", host, port),
	)
	if err != nil {
		return nil, err
	}

	client := &Client{serverAddr: addr}
	if err := client.connect(); err != nil {
		return nil, err
	}

	return client, nil
}

// StartSession starts an nREPL session by sending a "clone" request and keeping
// track of the session ID sent by the server in the response.
//
// Also sends a "describe" request and examines the response to ensure that the
// server is an Alda server.
func (client *Client) StartSession() error {
	req := map[string]interface{}{"op": "clone"}
	res, err := client.sendRequest(req)
	if err != nil {
		return err
	}

	// We could consider it an error if the "clone" response doesn't contain a new
	// session ID, because the consequence of not including a session ID on
	// requests is that on the server, every request is executed in a one-off
	// session context.
	//
	// At the moment, it's not a problem if we don't have a session ID, because
	// every request is executed in the same global context. If that ever stops
	// being the case, then it would probably make sense to bomb out here (i.e.
	// return an error, causing the program to print the message and exit)
	printResponseErrors(res)
	switch newSession := res["new-session"].(type) {
	case string:
		log.Info().Str("sessionID", newSession).Msg("Started nREPL session.")
		client.sessionID = newSession
	}

	req = map[string]interface{}{"op": "describe"}
	res, err = client.sendRequest(req)
	if err != nil {
		return err
	}
	printResponseErrors(res)

	errNotAnAldaServer := fmt.Errorf(
		"the server does not appear to be an Alda server",
	)

	switch res["versions"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return errNotAnAldaServer
	}
	versions := res["versions"].(map[string]interface{})

	switch versions["alda"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return errNotAnAldaServer
	}
	aldaVersion := versions["alda"].(map[string]interface{})

	// NOTE: We don't currently need to use the server's Alda version for
	// anything, so we're just logging it for informational purposes.
	log.Info().Interface("version", aldaVersion).Msg("Alda REPL server version")

	errEvalAndPlayNotSupported := fmt.Errorf(
		"the server does not appear to support the `eval-and-play` op",
	)

	switch res["ops"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return errEvalAndPlayNotSupported
	}
	ops := res["ops"].(map[string]interface{})

	_, supported := ops["eval-and-play"]
	if !supported {
		return errEvalAndPlayNotSupported
	}

	return nil
}

// RunClient runs an Alda REPL client session in the foreground.
func RunClient(serverHost string, serverPort int) error {
	client, err := NewClient(serverHost, serverPort)
	if err != nil {
		return err
	}
	defer client.Disconnect()

	if err := client.StartSession(); err != nil {
		return err
	}

	fmt.Printf(
		"%s\n\n%s\n\n%s\n\n",
		aurora.Blue(strings.Trim(aldaASCIILogo, "\n")),
		aurora.Cyan(strings.Trim(aldaVersionText, "\n")),
		aurora.Bold("Type :help for a list of available commands."),
	)

	os.MkdirAll(filepath.Dir(replHistoryFilepath), os.ModePerm)

	console, err := readline.NewEx(&readline.Config{
		Prompt:          "alda> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
		HistoryFile:     replHistoryFilepath,
	})
	if err != nil {
		return err
	}

	defer console.Close()

	log.SetOutput(console.Stderr())

ReplLoop:
	for {
		line, err := console.Readline()

		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		}

		if err == io.EOF {
			break
		}

		if len(line) == 0 {
			continue
		}

		input := strings.TrimSpace(line)

		if strings.HasPrefix(input, ":") && len(input) > 1 {
			re := regexp.MustCompile(`^:([^ ]+) ?(.*)$`)
			captured := re.FindStringSubmatch(input)

			// This shouldn't happen, but just in case...
			if len(captured) < 3 {
				fmt.Println("ERROR: failed to parse command.")
				continue
			}

			// captured[0] is the full string, e.g. ":foo bar baz"
			command := captured[1]
			args := captured[2]
			if err := client.handleCommand(command, args); err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}

			continue
		}

		switch input {
		case "quit", "exit", "bye":
			break ReplLoop
		default:
			req := map[string]interface{}{"op": "eval-and-play", "code": input}
			res, err := client.sendRequest(req)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				continue
			}

			printResponseErrors(res)
		}
	}

	return nil
}
