package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type playerState struct {
	Condition string
	Port      int
	Expiry    int64
	id        string
	readError error
}

func readPlayerStates() ([]playerState, error) {
	playersDir := cachePath("state", "players", VERSION)
	os.MkdirAll(playersDir, os.ModePerm)

	files, err := ioutil.ReadDir(playersDir)
	if err != nil {
		return nil, err
	}

	states := []playerState{}

	for _, file := range files {
		filepath := filepath.Join(playersDir, file.Name())

		var readError error
		var state playerState

		contents, err := ioutil.ReadFile(filepath)
		if err != nil {
			readError = err
		} else {
			err := json.Unmarshal(contents, &state)
			if err != nil {
				readError = err
			}
		}

		state.id = strings.Replace(file.Name(), ".json", "", 1)
		state.readError = readError

		states = append(states, state)
	}

	return states, nil
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Display information about running Alda processes",
	Run: func(_ *cobra.Command, args []string) {
		states, err := readPlayerStates()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("id\tport\tcondition\texpiry")

		for _, state := range states {
			if state.readError != nil {
				fmt.Fprintln(os.Stderr, state.readError)
				continue
			}

			expiry := humanize.Time(time.Unix(state.Expiry/1000, 0))

			fmt.Printf(
				"%s\t%d\t%s\t%s\n", state.id, state.Port, state.Condition, expiry,
			)
		}
	},
}
