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

	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
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

	if numArgs == 1 {
		openPort, err := findOpenPort()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		port = openPort
		file = os.Args[1]

		fmt.Printf("Starting player process on port %d...\n", port)

		cmd = exec.Command("./gradlew", "run", "--args", strconv.Itoa(port))
		cmd.Dir = "../player"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// See below about black magic to try and make sure the subprocess gets
		// killed when the main process exits.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Start()

		fmt.Println("Waiting a bit for the player process to start...")
		time.Sleep(5 * time.Second)
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

	scoreUpdates, err := parser.ParseFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	score := model.NewScore()
	score.Update(scoreUpdates...)
	fmt.Printf("%#v\n", score)

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nSending OSC messages to player on port: %d\n", port)
	fmt.Println("-- Press Ctrl-C to interrupt --")

	select {
	case <-sigChan:
		if cmd != nil {
			fmt.Println("Killing player process...")

			// Black magic to try and kill the subprocess once the main process ends.
			// I have no idea how robust this is. Probably a better approach would be
			// to add a lifetime to all player processes, such that they kill
			// themselves after being idle for a while. Could make this configurable
			// and set a short lifetime for the purposes of testing.
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			syscall.Kill(-pgid, 15)
		}
	}
}
