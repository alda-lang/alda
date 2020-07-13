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

type replCommand struct {
	helpSummary string
	helpDetails string
	run         func(args string) error
}

// FIXME
var replCommands = map[string]replCommand{
	"help": {
		helpSummary: "Display this help text.",
		helpDetails: `Usage:
  :help help
  :help new`,
		run: func(args string) error {
			return fmt.Errorf("not yet implemented")
		}},
}

func handleCommand(name string, args string) error {
	command, defined := replCommands[name]
	if !defined {
		return fmt.Errorf("unrecognized command: %s", name)
	}

	return command.run(args)
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
			if err := handleCommand(command, args); err != nil {
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
			if err := server.play(input); err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
		}
	}

	return nil
}
