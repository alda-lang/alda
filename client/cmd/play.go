package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"alda.io/client/emitter"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/spf13/cobra"
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

var port int
var file string

func init() {
	playCmd.Flags().IntVarP(
		&port, "port", "p", -1, "The port of the Alda player process to use",
	)

	playCmd.Flags().StringVarP(
		&file, "file", "f", "", "Read Alda source code from a file",
	)
	// TODO: Make this flag optional. Instead, allow input to be provided either
	// as a file (-f / --file), as a string (-c / --code), or via STDIN.
	playCmd.MarkFlagRequired("file")
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Evaluate and play Alda source code",
	Run: func(_ *cobra.Command, args []string) {
		var cmd *exec.Cmd
		startPlayerProcess := false

		if port == -1 {
			openPort, err := findOpenPort()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			startPlayerProcess = true
			port = openPort
		}

		scoreUpdates, err := parser.ParseFile(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		score := model.NewScore()
		if err := score.Update(scoreUpdates...); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if startPlayerProcess {
			aldaPlayer, err := exec.LookPath("alda-player")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Starting player process on port %d...\n", port)

			cmd = exec.Command(aldaPlayer, "run", "-p", strconv.Itoa(port))
			if err := cmd.Start(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// This is a hacky way to make sure that the player process is ready before
			// we start sending it messages. When we do this for real, it would be
			// better for the client to actually confirm that the player is up somehow.
			fmt.Println("Waiting a bit for the player process to start...")
			time.Sleep(5 * time.Second)
		}

		fmt.Printf("\nSending OSC messages to player on port: %d\n", port)
		if err := (emitter.OSCEmitter{Port: port}).EmitScore(score); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
