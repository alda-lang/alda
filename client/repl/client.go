package repl

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"alda.io/client/color"
	"alda.io/client/generated"
	"alda.io/client/help"
	"alda.io/client/json"
	log "alda.io/client/logging"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/util"
	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/google/uuid"
	bencode "github.com/jackpal/bencode-go"
)

const aldaASCIILogo = `
 █████╗ ██╗     ██████╗  █████╗
██╔══██╗██║     ██╔══██╗██╔══██╗
███████║██║     ██║  ██║███████║
██╔══██║██║     ██║  ██║██╔══██║
██║  ██║███████╗██████╔╝██║  ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝
`

func aldaVersionText(serverVersion string) string {
	return `
    Client version: ` + generated.ClientVersion + `
    Server version: ` + serverVersion + `
`
}

const serverConnectTimeout = 5 * time.Second

// Client is a stateful Alda REPL client object.
type Client struct {
	// Whether or not the client should continue the REPL session.
	running bool
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
	// The (optional) filepath to a file containing Alda source code. The contents
	// of the file can be loaded into the REPL server via the `:load` command. The
	// `:save` command creates/updates this file.
	inputFilepath string
}

type replCommand struct {
	helpSummary string
	helpDetails string
	run         func(client *Client, args string) error
}

var replCommands map[string]replCommand
var commandsSummmary string

func invalidArgsError(args []string) error {
	return fmt.Errorf("invalid arguments: %#v", args)
}

func init() {
	replCommands = map[string]replCommand{
		"export": {
			helpSummary: "Exports the current score as a MIDI file.",
			helpDetails: `Example usage:

  :export my-score.mid`,
			run: func(client *Client, argsString string) error {
				args, err := shlex.Split(argsString)
				if err != nil {
					return err
				}

				if len(args) != 1 {
					return invalidArgsError(args)
				}

				filename := args[0]

				res, err := client.sendRequest(map[string]interface{}{"op": "export"})
				if err != nil {
					return err
				}

				switch res["binary-data"].(type) {
				// It's a little weird that this is coming through as a string and not a
				// byte array, but it seems to work if we convert the string to a byte
				// array ¯\_(ツ)_/¯
				case string: // OK to proceed
				default:
					return fmt.Errorf(
						"the response from the REPL server did not contain the exported " +
							"score",
					)
				}
				binaryData := []byte(res["binary-data"].(string))

				return ioutil.WriteFile(filename, binaryData, 0644)
			},
		},

		"help": {
			helpSummary: "Displays this help text.",
			helpDetails: `Usage:

  :help help
  :help new`,
			run: func(client *Client, argsString string) error {
				args, err := shlex.Split(argsString)
				if err != nil {
					return err
				}

				if len(args) == 0 {
					fmt.Println(
						`For commands marked with (*), more detailed ` +
							`information about the command is
available via the :help command.

e.g. :help play

Available commands:

` + commandsSummmary)

					return nil
				}

				// Right now, :help only supports looking up documentation for a
				// command, so we only ever expect there to be a single argument like
				// "play" (i.e. `:help play`).
				//
				// In the future, we might want to provide help for a variety of topics,
				// so I'm going ahead and setting it up so that something like
				// `:help key signature` could work.
				subject := strings.Join(args, " ")

				cmd, hit := replCommands[subject]
				if !hit {
					return fmt.Errorf("no documentation available for '%s'", subject)
				}

				fmt.Println(cmd.helpSummary + "\n")
				if cmd.helpDetails != "" {
					fmt.Println(cmd.helpDetails + "\n")
				}

				return nil
			}},

		"instruments": {
			helpSummary: "Displays the list of available instruments.",
			run: func(client *Client, argsString string) error {
				res, err := client.sendRequest(
					map[string]interface{}{"op": "instruments"},
				)
				if err != nil {
					return err
				}

				switch res["instruments"].(type) {
				// For some reason, Go isn't recognizing the list of strings as a list
				// of strings, so I have to treat it like a list of anythings. OK,
				// whatever.
				case []interface{}: // OK to proceed
				default:
					return fmt.Errorf(
						"the response from the REPL server did not contain the list of " +
							"available instruments",
					)
				}
				instruments := res["instruments"].([]interface{})

				for _, instrument := range instruments {
					fmt.Println(instrument)
				}

				return nil
			},
		},

		"load": {
			helpSummary: "Loads a score file (*.alda) into the current REPL session.",
			helpDetails: `Usage:

  :load test/examples/bach_cello_suite_no_1.alda
  :load /Users/rick/Scores/love_is_alright_tonite.alda

After using :load <filename> to load a file, or after using :save to save the
working score to a file, running :load without arguments will reload the score
file into the REPL server.`,
			run: func(client *Client, argsString string) error {
				args, err := shlex.Split(argsString)
				if err != nil {
					return err
				}

				switch len(args) {
				case 0:
					if client.inputFilepath == "" {
						return fmt.Errorf("please specify the path to a file to load")
					}
				case 1:
					client.inputFilepath = args[0]
				default:
					return invalidArgsError(args)
				}

				contents, err := ioutil.ReadFile(client.inputFilepath)
				if err != nil {
					return err
				}

				_, err = client.sendRequest(
					map[string]interface{}{
						"op":   "load",
						"code": string(contents)},
				)
				if err != nil {
					return err
				}

				return nil
			},
		},

		"new": {
			helpSummary: "Resets the REPL server state and initializes a new score.",
			run: func(client *Client, argsString string) error {
				_, err := client.sendRequest(
					map[string]interface{}{"op": "new-score"},
				)
				if err != nil {
					return err
				}

				return nil
			},
		},

		"parts": {
			helpSummary: "Displays instrument currently in use.",
			run: func(client *Client, argsString string) error {
				// Read and parse score information
				scoreData, err := client.scoreData()
					if err != nil {
						return err
					}

				parts := scoreData.Search("parts")
				if parts.Data() == nil {
					return fmt.Errorf("Server response missing information about parts.")
				}
			
				// Print instrument Names and IDs from current score
				if len(parts.ChildrenMap()) == 0 {
					fmt.Println("No instruments in current score.")
				} else {
					fmt.Println("Parts:")
					for id, part := range parts.ChildrenMap() {
						fmt.Printf(
							"  %s (%s)\n",
							id,
							part.Search("stock-instrument").Data(),
						)
					}
				}
				return nil
			},
		},

		"play": {
			helpSummary: "Plays the current score.",
			helpDetails: `Can take optional ` + "`from`" + `and ` + "`to`" +
				`arguments, in the form of markers or mm:ss
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

				req := map[string]interface{}{"op": "replay"}

				for i := 0; i < len(args); i++ {
					// If this is the last argument, that means there are an odd number of
					// arguments, which is invalid because we are expecting an even number
					// of arguments.
					if i == len(args)-1 {
						return invalidArgsError(args)
					}

					switch args[i] {
					case "from":
						_, hit := req["from"]
						if hit {
							return invalidArgsError(args)
						}
						i++
						req["from"] = args[i]
					case "to":
						_, hit := req["to"]
						if hit {
							return invalidArgsError(args)
						}
						i++
						req["to"] = args[i]
					default:
						return invalidArgsError(args)
					}
				}

				_, err = client.sendRequest(req)
				if err != nil {
					return err
				}

				return nil
			},
		},

		"quit": {
			helpSummary: "Exits the Alda REPL session.",
			run: func(client *Client, argsString string) error {
				client.running = false
				return nil
			},
		},

		"save": {
			helpSummary: "Saves the current score into a file (*.alda).",
			helpDetails: `Usage:

  :save test/examples/bach_cello_suite_no_1.alda
  :save /Users/rick/Scores/love_is_alright_tonite.alda

After using :save <filename> to save a file, running :save again without
arguments will save the updated score to the same file.`,
			run: func(client *Client, argsString string) error {
				args, err := shlex.Split(argsString)
				if err != nil {
					return err
				}

				switch len(args) {
				case 0:
					if client.inputFilepath == "" {
						return fmt.Errorf("please specify the path to a file")
					}
				case 1:
					client.inputFilepath = args[0]
				default:
					return invalidArgsError(args)
				}

				scoreText, err := client.scoreText()
				if err != nil {
					return err
				}

				return ioutil.WriteFile(client.inputFilepath, []byte(scoreText), 0644)
			},
		},

		"score": {
			helpSummary: "Prints information about the current score.",
			helpDetails: `Example usage:

  Print the text (Alda code) of the score (both commands are equivalent):
    :score
    :score text

  Print a summary of the score, including information about instruments and
  markers:
    :score info

  Print a data representation of the score (This is the output that you get
  when you run ` + "`alda parse -o data ...`" + ` at the command line.):
    :score data

  Print the parsed events output of the score (This is the output that you get
  when you run ` + "`alda parse -o events ...`" + ` at the command line.):
    :score events

  Print the parsed AST output of the score (This is the output that you get
  when you run ` + "`alda parse -o ast-human ...`" + ` at the command line.):
    :score ast`,
			run: func(client *Client, argsString string) error {
				args, err := shlex.Split(argsString)
				if err != nil {
					return err
				}

				var subcommand string
				switch len(args) {
				case 0:
					subcommand = "text"
				case 1:
					subcommand = args[0]
				default:
					return invalidArgsError(args)
				}

				switch subcommand {
				default:
					return invalidArgsError(args)
				case "text":
					scoreText, err := client.scoreText()
					if err != nil {
						return err
					}

					fmt.Println(scoreText)

				case "data":
					scoreData, err := client.scoreData()
					if err != nil {
						return err
					}

					fmt.Println(scoreData.StringIndent("", "  "))

				case "events":
					scoreEvents, err := client.scoreEvents()
					if err != nil {
						return err
					}

					fmt.Println(scoreEvents.StringIndent("", "  "))

				case "ast":
					scoreAST, err := client.scoreAST()
					if err != nil {
						return err
					}

					fmt.Println(parser.HumanReadableAST(scoreAST))

				case "info":
					scoreData, err := client.scoreData()
					if err != nil {
						return err
					}

					return printScoreInfo(scoreData)
				}

				return nil
			},
		},

		"stop": {
			helpSummary: "Stops playback.",
			run: func(client *Client, argsString string) error {
				_, err := client.sendRequest(map[string]interface{}{"op": "stop"})
				if err != nil {
					return err
				}

				return nil
			},
		},

		"version": {
			helpSummary: "Displays the version numbers of the Alda server and client.",
			run: func(client *Client, argsString string) error {
				fmt.Printf("Client version: %s\n", generated.ClientVersion)

				res, err := client.sendRequest(map[string]interface{}{"op": "describe"})
				if err != nil {
					return err
				}

				serverVersionInfo, err := serverVersion(res)
				if err != nil {
					return err
				}

				switch serverVersionInfo["version-string"].(type) {
				case string: // OK to proceed
				default:
					return fmt.Errorf("server response is missing version string")
				}
				serverVersion := serverVersionInfo["version-string"].(string)

				fmt.Printf("Server version: %s\n", serverVersion)

				return nil
			},
		},
	}

	sortedKeys := []string{}
	maxKeyLength := 0
	for k := range replCommands {
		sortedKeys = append(sortedKeys, k)
		if len(k) > maxKeyLength {
			maxKeyLength = len(k)
		}
	}
	sort.Strings(sortedKeys)

	var builder strings.Builder
	for _, k := range sortedKeys {
		cmd := replCommands[k]

		detailIndicator := ""
		if cmd.helpDetails != "" {
			detailIndicator = " (*)"
		}

		fmt.Fprintf(
			&builder,
			"    :%s%s%s%s\n",
			k,
			strings.Repeat(" ", maxKeyLength-len(k)+2),
			cmd.helpSummary,
			detailIndicator,
		)
	}

	commandsSummmary = builder.String()
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
		return help.UserFacingErrorf(
			`Attempting to connect to %s resulted in this error:

  %s

Is there an Alda REPL server running on that port? If not, you can start one
by running:

  %s`,
			color.Aurora.Bold(client.serverAddr),
			color.Aurora.BgRed(err),
			color.Aurora.BrightYellow(
				fmt.Sprintf("alda repl --server --port %d", client.serverAddr.Port),
			),
		)
	}

	return nil
}

// replClientRequestContext provides context about the behavior around sending
// nREPL requests to the server.
type replClientRequestContext struct {
	suppressErrorPrinting bool
}

// replClientRequestOption is a function that customizes a
// replClientRequestContext instance.
type replClientRequestOption func(*replClientRequestContext)

// suppressErrorPrinting disables the default behavior where we print errors in
// responses from the server.
func suppressErrorPrinting() replClientRequestOption {
	return func(ctx *replClientRequestContext) {
		ctx.suppressErrorPrinting = true
	}
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
	req map[string]interface{}, opts ...replClientRequestOption,
) (map[string]interface{}, error) {
	ctx := &replClientRequestContext{}
	for _, opt := range opts {
		opt(ctx)
	}

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

	// Originally, this wasn't here; instead, we called `printResponseErrors(res)`
	// as needed to handle the response returned by this function.
	//
	// Then I realized that we were _always_ calling `printResponseErrors`, and I
	// also ended up needing to print response errors from outside of the repl
	// package. Instead of making `printResponseErrors` public and continuing to
	// use it everywhere, I decided that it makes more sense to put it here and
	// just always print response errors automatically.
	//
	// This default behavior can be disabled via the `suppressErrorPrinting`
	// option.
	if !ctx.suppressErrorPrinting {
		printResponseErrors(res)
	}

	return res, nil
}

func ResponseErrors(res map[string]interface{}) []string {
	statuses, ok := res["status"].([]interface{})
	if !ok {
		return []string{fmt.Sprintf("%#v", res)}
	}

	errorStatus := false
	for _, status := range statuses {
		if status == "error" {
			errorStatus = true
		}
	}
	if !errorStatus {
		return nil
	}

	problems, ok := res["problems"].([]interface{})
	if !ok {
		return []string{fmt.Sprintf("%#v", res)}
	}

	// We return the entire response as an error in this case, because the error
	// status indicated that _something_ was wrong.
	if len(problems) == 0 {
		return []string{fmt.Sprintf("%#v", res)}
	}

	errors := []string{}
	for _, problem := range problems {
		if str, ok := problem.(string); ok {
			errors = append(errors, str)
		} else {
			errors = append(errors, fmt.Sprintf("%#v", problem))
		}
	}

	return errors
}

func printResponseErrors(res map[string]interface{}) {
	errors := ResponseErrors(res)

	switch len(errors) {
	case 0:
		return
	case 1:
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", errors[0])
	default:
		fmt.Fprintln(os.Stderr, "ERRORS:")
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
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

	client := &Client{serverAddr: addr, running: true}
	if err := client.connect(); err != nil {
		return nil, err
	}

	return client, nil
}

var errNotAnAldaServer = fmt.Errorf(
	"the server does not appear to be an Alda server",
)

var errEvalAndPlayNotSupported = fmt.Errorf(
	"the server does not appear to support the `eval-and-play` op",
)

func serverVersion(res map[string]interface{}) (map[string]interface{}, error) {
	switch res["versions"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return nil, errNotAnAldaServer
	}
	versions := res["versions"].(map[string]interface{})

	switch versions["alda"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return nil, errNotAnAldaServer
	}
	return versions["alda"].(map[string]interface{}), nil
}

func (client *Client) scoreText() (string, error) {
	res, err := client.sendRequest(
		map[string]interface{}{"op": "score-text"},
	)
	if err != nil {
		return "", err
	}

	switch res["text"].(type) {
	case string: // OK to proceed
	default:
		return "", fmt.Errorf(
			"the response from the REPL server did not contain the score " +
				"text",
		)
	}
	return res["text"].(string), nil
}

func (client *Client) scoreData() (*json.Container, error) {
	res, err := client.sendRequest(
		map[string]interface{}{"op": "score-data"},
	)
	if err != nil {
		return nil, err
	}

	switch res["data"].(type) {
	case string: // OK to proceed
	default:
		return nil, fmt.Errorf(
			"the response from the REPL server did not contain the score data",
		)
	}

	return json.ParseJSON([]byte(res["data"].(string)))
}

func (client *Client) scoreEvents() (*json.Container, error) {
	res, err := client.sendRequest(
		map[string]interface{}{"op": "score-events"},
	)
	if err != nil {
		return nil, err
	}

	switch res["events"].(type) {
	case string: // OK to proceed
	default:
		return nil, fmt.Errorf(
			"the response from the REPL server did not contain the score events",
		)
	}

	return json.ParseJSON([]byte(res["events"].(string)))
}

func (client *Client) scoreAST() (*json.Container, error) {
	res, err := client.sendRequest(
		map[string]interface{}{"op": "score-ast"},
	)
	if err != nil {
		return nil, err
	}

	switch res["ast"].(type) {
	case string: // OK to proceed
	default:
		return nil, fmt.Errorf(
			"the response from the REPL server did not contain the score AST",
		)
	}

	return json.ParseJSON([]byte(res["ast"].(string)))
}

func printScoreInfo(scoreData *json.Container) error {
	parts := scoreData.Search("parts")
	if parts.Data() == nil {
		return fmt.Errorf("server response missing information about parts")
	}

	fmt.Println("Parts:")

	if len(parts.ChildrenMap()) == 0 {
		fmt.Println("  (none)")
	} else {
		for id, part := range parts.ChildrenMap() {
			fmt.Printf(
				"  %s (%s)\n",
				id,
				part.Search("stock-instrument").Data(),
			)
		}
	}

	fmt.Println()

	currentParts := scoreData.Search("current-parts")
	if currentParts.Data() == nil {
		return fmt.Errorf(
			"server response missing information about current parts",
		)
	}

	fmt.Println("Current parts:")

	if len(currentParts.Children()) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, id := range currentParts.Children() {
			fmt.Printf(
				"  %s (%s)\n",
				id.Data(),
				parts.Search(id.Data().(string), "stock-instrument").Data(),
			)
		}
	}

	fmt.Println()

	events := scoreData.Search("events")
	if events.Data() == nil {
		return fmt.Errorf(
			"server response missing information about events",
		)
	}

	fmt.Printf("Events:\n  %d\n\n", len(events.Children()))

	markers := scoreData.Search("markers")
	if markers.Data() == nil {
		return fmt.Errorf(
			"server response missing information about markers",
		)
	}

	fmt.Println("Markers:")

	if len(markers.ChildrenMap()) == 0 {
		fmt.Println("  (none)")
	} else {
		type markerEntry struct {
			name   string
			offset float64
		}

		markerEntries := []markerEntry{}
		for name, offset := range markers.ChildrenMap() {
			markerEntries = append(markerEntries, markerEntry{
				name: name,
				// We could do a type switch here instead and return an error if
				// the value isn't a float, but that would be really weird if it
				// wasn't a float. If we've gotten this far and the offset value isn't a
				// float, let's just panic.
				offset: offset.Data().(float64),
			})
		}

		sort.Slice(markerEntries, func(i, j int) bool {
			return markerEntries[i].offset < markerEntries[j].offset
		})

		for _, marker := range markerEntries {
			fmt.Printf("  %s (%f)\n", marker.name, marker.offset)
		}
	}

	fmt.Println()

	return nil
}

// StartSession starts an nREPL session by sending a "clone" request and keeping
// track of the session ID sent by the server in the response.
//
// Also sends a "describe" request and examines the response to ensure that the
// server is an Alda server.
//
// Returns the "describe" response, or an error if something went wrong. (One
// example of something going wrong is that the response doesn't appear to be
// from an Alda server.)
func (client *Client) StartSession() (map[string]interface{}, error) {
	req := map[string]interface{}{"op": "clone"}
	res, err := client.sendRequest(req)
	if err != nil {
		return nil, err
	}

	// We could consider it an error if the "clone" response doesn't contain a new
	// session ID, because the consequence of not including a session ID on
	// requests is that on the server, every request is executed in a one-off
	// session context.
	//
	// At the moment, it's not a problem if we don't have a session ID, because
	// every request is executed in the same global context. If that ever stops
	// being the case, then it would probably make sense to check that there is a
	// session ID here and bomb out if it's absent (i.e. return an error, causing
	// the program to print the message and exit).

	switch newSession := res["new-session"].(type) {
	case string:
		log.Info().Str("sessionID", newSession).Msg("Started nREPL session.")
		client.sessionID = newSession
	}

	req = map[string]interface{}{"op": "describe"}
	res, err = client.sendRequest(req)
	if err != nil {
		return nil, err
	}

	serverVersion, err := serverVersion(res)
	if err != nil {
		return nil, err
	}

	// NOTE: We don't currently need to use the server's Alda version for
	// anything, so we're just logging it for informational purposes.
	log.Info().Interface("version", serverVersion).Msg("Alda REPL server version")

	switch res["ops"].(type) {
	case map[string]interface{}: // OK to proceed
	default:
		return nil, errEvalAndPlayNotSupported
	}
	ops := res["ops"].(map[string]interface{})

	_, supported := ops["eval-and-play"]
	if !supported {
		return nil, errEvalAndPlayNotSupported
	}

	return res, nil
}

// RunClient runs an Alda REPL client session in the foreground.
func RunClient(serverHost string, serverPort int) error {
	client, err := NewClient(serverHost, serverPort)
	if err != nil {
		return err
	}
	defer client.Disconnect()

	describeResponse, err := client.StartSession()
	if err != nil {
		return err
	}

	serverVersionInfo, err := serverVersion(describeResponse)
	if err != nil {
		return err
	}

	var serverVersion string
	switch versionString := serverVersionInfo["version-string"].(type) {
	case string:
		serverVersion = versionString
	default:
		serverVersion = fmt.Sprintf("%#v", serverVersionInfo)
	}

	fmt.Printf(
		"%s\n\n%s\n\n%s\n\n",
		color.Aurora.Blue(strings.Trim(aldaASCIILogo, "\n")),
		color.Aurora.Cyan(strings.Trim(aldaVersionText(serverVersion), "\n")),
		color.Aurora.Bold("Type :help for a list of available commands."),
	)

	err = os.MkdirAll(filepath.Dir(replHistoryFilepath), os.ModePerm)
	if err != nil {
		return err
	}

	console, err := readline.NewEx(&readline.Config{
		Prompt:          "alda> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
		HistoryFile:     replHistoryFilepath,
	})
	if err != nil {
		return err
	}

	switch serverHost {
	case "localhost", "127.0.0.1", "0.0.0.0":
		system.StartingPlayerProcesses()
	}

	defer console.Close()

	log.SetOutput(console.Stderr())

	for client.running {
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
			client.running = false
		default:
			req := map[string]interface{}{"op": "eval-and-play", "code": input}
			_, err := client.sendRequest(req)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				continue
			}
		}
	}

	return nil
}

// SendMessage opens a one-off client session to an Alda REPL server, sends a
// message, and returns the response from the server.
func SendMessage(
	host string, port int, message map[string]interface{},
) (map[string]interface{}, error) {
	client, err := NewClient(host, port)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect()

	_, err = client.StartSession()
	if err != nil {
		return nil, err
	}

	return client.sendRequest(message, suppressErrorPrinting())
}
