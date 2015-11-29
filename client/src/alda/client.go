package main

import (
	// go standard library
	"errors"
	"fmt"
	"os"

	// within this project
	"server"
	"util"

	// third party
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func main() {
	app := cli.NewApp()
	app.Name = "alda"
	app.Version = "0.14.2"
	app.Usage = "a music language for musicians ♬ ♪"

	// global options
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "V, verbose",
			Usage: "Show verbose output",
		},
		cli.StringFlag{
			Name:  "H, host",
			Value: "localhost",
			Usage: "Specify the host of the Alda server to use",
		},
		cli.IntFlag{
			Name:  "p, port",
			Value: 27713,
			Usage: "Specify the port of the Alda server to use",
		},
		cli.IntFlag{
			Name:  "b, pre-buffer",
			Value: 0,
			Usage: "The number of milliseconds of lead time for buffering",
		},
		cli.IntFlag{
			Name:  "B, post-buffer",
			Value: 1000,
			Usage: "The number of milliseconds to wait after playing the score, before exiting",
		},
		cli.BoolFlag{
			Name:  "s, stock",
			Usage: "Use the default MIDI soundfont of your JVM, instead of FluidR3",
		},
	}

	// alda commands
	app.Commands = []cli.Command{
		{
			Name:  "server",
			Usage: "Manage Alda servers",
			Subcommands: []cli.Command{
				{
					Name:    "start",
					Aliases: []string{"up"},
					Usage:   "Start an Alda server",
					Action:  AldaStart,
				},
				{
					Name:    "stop",
					Aliases: []string{"down"},
					Usage:   "Stop the Alda server",
					Action:  AldaStop,
				},
				{
					Name:    "restart",
					Aliases: []string{"downup"},
					Usage:   "Stop and restart the Alda server",
					Action:  AldaRestart,
				},
			},
		},
		// alias for `alda server start`
		{
			Name:    "start",
			Aliases: []string{"up"},
			Usage:   "Start an Alda server",
			Action:  AldaStart,
		},
		// alias for `alda server stop`
		{
			Name:    "stop",
			Aliases: []string{"down"},
			Usage:   "Stop the Alda server",
			Action:  AldaStop,
		},
		// alias for `alda server restart`
		{
			Name:    "restart",
			Aliases: []string{"downup"},
			Usage:   "Stop and restart the Alda server",
			Action:  AldaRestart,
		},
		{
			Name:  "status",
			Usage: "Returns whether the server is up",
			Action: func(c *cli.Context) {
				host, port := getHostAndPort(c)

				if err := server.CheckForConnection(host, port); err != nil {
					util.Msg(host, port, "Server down %s", color.RedString("✗"))
					os.Exit(1)
				}

				util.Msg(host, port, "Server up %s", color.GreenString("✓"))
			},
		},
		{
			Name:  "version",
			Usage: "Display the version of the Alda server",
			Action: func(c *cli.Context) {
				host, port := getHostAndPort(c)

				statusCode, body, err := server.Get(host, port, "/version", nil)

				if err != nil {
					util.Error(host, port, util.SanitizeError(err).Error())
					os.Exit(1)
				}

				if statusCode != 200 {
					util.Error(host, port, body)
					os.Exit(1)
				}

				util.Msg(host, port, body)
			},
		},
		{
			Name:  "play",
			Usage: "Evaluate and play Alda code",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "f, file",
					Usage: "Read Alda code from a file",
				},
				cli.StringFlag{
					Name:  "c, code",
					Usage: "Supply Alda code as a string",
				},
				cli.StringFlag{
					Name:  "F, from",
					Usage: "A time marking or marker from which to start playback",
				},
				cli.StringFlag{
					Name:  "T, to",
					Usage: "A time marking or marker at which to end playback",
				},
				cli.BoolFlag{
					Name:  "r, replace",
					Usage: "Replace the existing score with new code",
				},
			},
			Action: func(c *cli.Context) {
				host, port := getHostAndPort(c)

				file := c.String("file")
				code := c.String("code")

				codeOrFilename, codeType, err := handleFileAndCodeArgs(file, code)
				if err != nil {
					util.Error(host, port, err.Error())
					os.Exit(1)
				}

				replaceScore := c.Bool("replace")

				err = server.Play(host, port, codeOrFilename, codeType, replaceScore)
				if err != nil {
					util.Error(host, port, err.Error())
					os.Exit(1)
				}

				util.Msg(host, port, "Playing %s.", codeType)
			},
		},
		{
			Name:  "score",
			Usage: "Display the score in progress",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "t, text",
					Usage: "Display the score text",
				},
				cli.BoolFlag{
					Name:  "l, lisp",
					Usage: "Display the score in the form of alda.lisp (Clojure) code",
				},
				cli.BoolFlag{
					Name:  "m, map",
					Usage: "Display the map of score data",
				},
			},
			Action: func(c *cli.Context) {
				host, port := getHostAndPort(c)

				asText := c.Bool("text")
				asLisp := c.Bool("lisp")
				asMap := c.Bool("map")

				endpoint, err := getScoreEndpoint(asText, asLisp, asMap)
				if err != nil {
					util.Error(host, port, err.Error())
					os.Exit(1)
				}

				statusCode, body, err := server.Get(host, port, endpoint, nil)
				if err != nil {
					util.Error(host, port, util.SanitizeError(err).Error())
					os.Exit(1)
				}

				if statusCode != 200 {
					util.Error(host, port, body)
					os.Exit(1)
				}

				fmt.Println(body)
			},
		},
		{
			Name:    "new",
			Aliases: []string{"delete"},
			Usage:   "Delete the score and start a new one",
			Action: func(c *cli.Context) {
				// TODO: ask for confirmation before deleting score (if it's non-empty)
				host, port := getHostAndPort(c)

				statusCode, body, err := server.Delete(host, port, "/score", nil)
				if err != nil {
					util.Error(host, port, util.SanitizeError(err).Error())
					os.Exit(1)
				}

				if statusCode != 200 {
					util.Error(host, port, body)
					os.Exit(1)
				}

				fmt.Println(body)
			},
		},
		{
			Name:  "edit",
			Usage: "Edit the score in progress using $EDITOR",
			Action: func(c *cli.Context) {
				fmt.Println("editing score on port ", c.GlobalInt("port"))
			},
		},
		{
			Name:  "parse",
			Usage: "Display the result of parsing Alda code",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "f, file",
					Usage: "Read Alda code from a file",
				},
				cli.StringFlag{
					Name:  "c, code",
					Usage: "Supply Alda code as a string",
				},
				cli.BoolFlag{
					Name:  "l, lisp",
					Usage: "Return the resulting Clojure (alda.lisp) code",
				},
				cli.BoolFlag{
					Name:  "m, map",
					Usage: "Evaluate the score and return the resulting score map",
				},
			},
			Action: func(c *cli.Context) {
				host, port := getHostAndPort(c)

				file := c.String("file")
				code := c.String("code")

				codeOrFilename, codeType, err := handleFileAndCodeArgs(file, code)
				if err != nil {
					util.Error(host, port, err.Error())
					os.Exit(1)
				}

				lispcode := c.Bool("lisp")
				scoremap := c.Bool("map")

				endpoint, err := getParseEndpoint(lispcode, scoremap)
				if err != nil {
					util.Error(host, port, err.Error())
					os.Exit(1)
				}

				fmt.Println("parsing code:", codeOrFilename)
				fmt.Println("code type:", codeType)
				fmt.Println("endpoint:", endpoint)
			},
		},
	}

	app.Run(os.Args)
	os.Exit(0)
}

func AldaStart(c *cli.Context) {
	host, port := getHostAndPort(c)

	if host != "http://localhost" {
		util.Error(host, port, "Alda servers cannot be started remotely.")
		os.Exit(1)
	}

	if err := server.CheckForConnection(host, port); err == nil {
		util.Msg(host, port, "Alda server already up.")
		os.Exit(0)
	}

	preBuffer := c.GlobalInt("pre-buffer")
	postBuffer := c.GlobalInt("post-buffer")
	stock := c.GlobalBool("stock")

	util.Msg(host, port, "Starting Alda server...")
	if err := server.Start(port, preBuffer, postBuffer, stock); err != nil {
		util.Error(host, port, err.Error())
		os.Exit(1)
	}

	if err := server.WaitForConnection(host, port); err != nil {
		util.Error(host, port, err.Error())
		os.Exit(1)
	}

	util.Msg(host, port, "Server up %s", color.GreenString("✓"))
}

func AldaStop(c *cli.Context) {
	host, port := getHostAndPort(c)

	if err := server.CheckForConnection(host, port); err != nil {
		util.Msg(host, port, "No Alda server running.")
		os.Exit(0)
	}

	util.Msg(host, port, "Stopping Alda server...")

	if err := server.Stop(host, port); err != nil {
		util.Error(host, port, err.Error())
		os.Exit(1)
	}

	if err := server.WaitForLackOfConnection(host, port); err != nil {
		util.Error(host, port, err.Error())
	}

	util.Msg(host, port, "Server down %s", color.GreenString("✓"))
}

func AldaRestart(c *cli.Context) {
	AldaStop(c)
	println()
	AldaStart(c)
}

func getHostAndPort(c *cli.Context) (string, int) {
	host := util.NormalizeHostString(c.GlobalString("host"))
	port := c.GlobalInt("port")
	return host, port
}

func handleFileAndCodeArgs(file string, code string) (string, string, error) {
	if file == "" && code == "" {
		return "", "", errors.New("You must supply a string or a file containing Alda code.")
	}

	if file != "" && code != "" {
		return "", "", errors.New("You must either supply a --code or a --file argument, not both.")
	}

	if file != "" {
		return file, "file", nil
	} else {
		return code, "code", nil
	}
}

func getParseEndpoint(lispcode bool, scoremap bool) (string, error) {
	if lispcode && scoremap {
		return "", errors.New("You must include either the --lisp or --map flag, not both.")
	}

	if lispcode {
		return "/parse/lisp", nil
	}

	if scoremap {
		return "/parse/map", nil
	}

	return "/parse", nil
}

func getScoreEndpoint(asText bool, asLisp bool, asMap bool) (string, error) {
	activeFlags := 0
	for _, flag := range []bool{asText, asLisp, asMap} {
		if flag {
			activeFlags++
		}
	}

	if activeFlags > 1 {
		return "", errors.New("You may include only one score format: --text, --lisp, or --map.")
	}

	if asText {
		return "/score/text", nil
	}

	if asLisp {
		return "/score/lisp", nil
	}

	if asMap {
		return "/score/map", nil
	}

	return "/score", nil
}
