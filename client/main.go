package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"alda.io/client/emitter"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/daveyarwood/go-osc/osc"
)

func findOpenPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	defer listener.Close()

	if err != nil {
		return 0, err
	}

	address := listener.Addr().String()
	portStr := address[strings.LastIndex(address, ":")+1 : len(address)]
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		fmt.Printf("Failed to find open port. Address: %s\n", address)
		return 0, err
	}

	return int(port), nil
}

func printUsage() {
	fmt.Printf(
		"Usage:\n"+
			"  %s FILE         Start an Alda player and play FILE.\n"+
			"  %s PORT FILE    Use Alda player on port PORT to play FILE.\n",
		os.Args[0],
		os.Args[0])
}

// TODO:
// * Command/arg parsing via cobra or similar
// * Top-level -v / --verbose flag that sets the log level via
//   log.SetGlobalLevel

func main() {
	log.SetGlobalLevel("info")

	numArgs := len(os.Args[1:])

	if numArgs < 1 || numArgs > 2 {
		printUsage()
		os.Exit(1)
	}

	var port int
	var file string
	var cmd *exec.Cmd
	startPlayerProcess := false

	if numArgs == 1 {
		openPort, err := findOpenPort()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		startPlayerProcess = true
		port = openPort
		file = os.Args[1]
	} else {
		specifiedPort, err := strconv.ParseInt(os.Args[1], 10, 32)
		if err != nil {
			fmt.Println(err)
			printUsage()
			os.Exit(1)
		}

		port = int(specifiedPort)
		file = os.Args[2]
	}

	var score *model.Score

	if scoreUpdates, err := parser.ParseFile(file); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		score = model.NewScore()
		if err := score.Update(scoreUpdates...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if startPlayerProcess {
		fmt.Printf("Starting player process on port %d...\n", port)

		cmd = exec.Command("./gradlew", "run", "--args", strconv.Itoa(port))
		cmd.Dir = "../player"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()

		// This is a hacky way to make sure that the player process is ready before
		// we start sending it messages. When we do this for real, it would be
		// better for the client to actually confirm that the player is up somehow.
		fmt.Println("Waiting a bit for the player process to start...")
		time.Sleep(5 * time.Second)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nSending OSC messages to player on port: %d\n", port)
	emitter.OSCEmitter{Port: port}.EmitScore(score)

	if startPlayerProcess {
		fmt.Println("-- Press Ctrl-C to interrupt --")

		select {
		case <-sigChan:
			if cmd != nil {
				fmt.Println("Sending /system/shutdown message to player process...")

				client := osc.NewClient("localhost", int(port))
				client.SetNetworkProtocol(osc.TCP)
				if err := client.Send(osc.NewMessage("/system/shutdown")); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}
}
