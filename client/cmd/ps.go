package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

type playerState struct {
	Condition string
	Port      int
	Expiry    int64
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Display information about running Alda processes",
	Run: func(_ *cobra.Command, args []string) {
		xdg := xdg.New("", "alda")
		cacheDir := xdg.CacheHome()
		playersDir := filepath.Join(cacheDir, "state", "players", VERSION)

		files, err := ioutil.ReadDir(playersDir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("id\tport\tcondition\texpiry")

		for _, file := range files {
			id := strings.Replace(file.Name(), ".json", "", 1)

			filepath := filepath.Join(playersDir, file.Name())
			contents, err := ioutil.ReadFile(filepath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			var state playerState
			err = json.Unmarshal(contents, &state)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			expiry := humanize.Time(time.Unix(state.Expiry/1000, 0))

			fmt.Printf("%s\t%d\t%s\t%s\n", id, state.Port, state.Condition, expiry)
		}
	},
}
