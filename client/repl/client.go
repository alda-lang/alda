// vim: tabstop=2

package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"alda.io/client/generated"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"github.com/chzyer/readline"
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

// Client is a stateful Alda REPL client object.
type Client struct {
	// TODO: The end goal is to have the client send the input over the network to
	// a server (which could potentially be in another process or on another
	// machine) and print feedback based on the response from the server.
	//
	// For now, just to get something working, I'm implementing this all in one
	// process, using stuff in repl/server.go directly.
	server *Server
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
		run: func(client *Client, args string) error {
			// TODO: send a message over the network instead
			//       * I need to think about the parameters that the message allows.
			//         It might look something like the EmissionContext, but it
			//         doesn't necessarily have to be the same.
			//
			//         UPDATE: I was tempted to just use EmissionContext and make sure
			//         that we keep it serializable, because that would be one less
			//         set of options to worry about. But then, thinking about it
			//         more, I decided it would be better to keep the two things
			//         separate, because one is user-facing (the REPL server API) and
			//         the other is an implementation detail (EmissionContext).
			//
			// TODO: parse and handle "from" and "to" options
			return client.server.replay()
		}},
}

func (client *Client) handleCommand(name string, args string) error {
	command, defined := replCommands[name]
	if !defined {
		return fmt.Errorf("unrecognized command: %s", name)
	}

	return command.run(client, args)
}

var replHistoryFilepath = system.CachePath("history", "alda-repl-history")

// RunClient runs an Alda REPL client session in the foreground.
func RunClient() error {
	// TODO: The end goal is to have the client send the input to a server (which
	// could potentially be in another process or on another machine) and print
	// feedback based on the response from the server.
	//
	// For now, just to get something working, I'm implementing this all in one
	// process, using stuff in repl/server.go directly.
	server, err := RunServer()
	if err != nil {
		return err
	}

	client := &Client{server: server}

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
			// FIXME: see comment at the top of RunClient about communicating with a
			// server process instead of doing this all in one client+server process
			if err := server.evalAndPlay(input); err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
		}
	}

	return nil
}
