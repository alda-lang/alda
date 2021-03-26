package system

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"alda.io/client/generated"
	"alda.io/client/help"
	log "alda.io/client/logging"
	"github.com/logrusorgru/aurora"
)

// PlayerState describes the current state of a player process. These states are
// continously written to files by each player process. (See: StateManager.kt.)
type PlayerState struct {
	State     string
	Port      int
	Expiry    int64
	ID        string
	ReadError error
}

// ReadPlayerStates reads all of the player state files in the Alda cache
// directory and returns a list of player state structs describing the current
// state of each player process.
//
// Returns an error if something goes wrong.
func ReadPlayerStates() ([]PlayerState, error) {
	playersDir := CachePath("state", "players", generated.ClientVersion)
	if err := os.MkdirAll(playersDir, os.ModePerm); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(playersDir)
	if err != nil {
		return nil, err
	}

	states := []PlayerState{}

	for _, file := range files {
		filepath := filepath.Join(playersDir, file.Name())

		var readError error
		var state PlayerState

		contents, err := ioutil.ReadFile(filepath)
		if err != nil {
			readError = err
		} else {
			err := json.Unmarshal(contents, &state)
			if err != nil {
				readError = err
			}
		}

		state.ID = strings.Replace(file.Name(), ".json", "", 1)
		state.ReadError = readError

		states = append(states, state)
	}

	return states, nil
}

// ErrAldaPlayerNotFoundOnPath is the error message that is returned when
// `alda-player` is not found on the PATH.
var ErrAldaPlayerNotFoundOnPath error

// ErrNoPlayersAvailable is the error message that is returned when no player
// process is in an available state.
var ErrNoPlayersAvailable error

func init() {
	ErrAldaPlayerNotFoundOnPath = help.UserFacingErrorf(
		`%s does not appear to be installed.

The %s command-line client needs to spawn %s processes in order to play audio
in the background.

To install %s, run %s and answer %s when prompted.`,
		aurora.Bold("alda-player"),
		aurora.Bold("alda"),
		aurora.Bold("alda-player"),
		aurora.Bold("alda-player"),
		aurora.BrightYellow("alda doctor"),
		aurora.Bold("y"),
	)

	playerLogFile := CachePath("logs", "alda-player.log")

	ErrNoPlayersAvailable = help.UserFacingErrorf(
		`It looks like Alda is having trouble starting player processes in the
background. This could happen for a number of reasons.

To troubleshoot:

  • Run %s and see if any of the health checks fail.

  • Run %s to see the state of any currently running player
    processes.

  • Look for error messages in:
      %s

  • Try running a player process in the foreground:
      %s

  • Try to make it play something:
      %s`,
		aurora.BrightYellow("alda doctor"),
		aurora.BrightYellow("alda ps"),
		aurora.BrightYellow(playerLogFile),
		aurora.BrightYellow("alda-player -v run -p 27278"),
		aurora.BrightYellow("alda -v2 play -p 27278 -c \"piano: c12 e g > c4\""),
	)
}

// FindAvailablePlayer finds a player that is in an available state and returns
// current information about its state.
//
// Returns `ErrNoPlayersAvailable` if no player is currently in an available
// state.
func FindAvailablePlayer() (PlayerState, error) {
	players, err := ReadPlayerStates()
	if err != nil {
		return PlayerState{}, err
	}

	for _, player := range players {
		if player.State == "ready" {
			return player, nil
		}
	}

	return PlayerState{}, ErrNoPlayersAvailable
}

// FindPlayerByID returns the current state of the player process with the
// provided ID.
//
// Returns an error if no player is found with that ID, or if something goes
// wrong in the process of reading the player states.
func FindPlayerByID(id string) (PlayerState, error) {
	players, err := ReadPlayerStates()
	if err != nil {
		return PlayerState{}, err
	}

	for _, player := range players {
		if player.ID == id {
			return player, nil
		}
	}

	return PlayerState{}, help.UserFacingErrorf(
		`No player was found with the ID %s.

To list the current player processes, you can run %s.

You can also omit the %s / %s option, and Alda will find a player process
for you automatically.`,
		aurora.BrightYellow(id),
		aurora.BrightYellow("alda ps"),
		aurora.BrightYellow("-i"),
		aurora.BrightYellow("--player-id"),
	)
}

// AldaPlayerPath returns the absolute path to the `alda-player` executable and
// whether the player is the same version as the client.
//
// To determine whether the player is the same version of the client, we run
// `alda-player info` and parse the version number from the output. If the
// versions don't match, a warning is logged that includes instructions about
// how to fix this situation.
//
// Returns `ErrAldaPlayerNotFoundOnPath` if `alda-player` is not found on the
// PATH.
//
// Returns an unspecified error if something else goes wrong.
func AldaPlayerPath() (playerPath string, sameVersion bool, err error) {
	aldaPlayer, err := exec.LookPath("alda-player")
	if err != nil {
		return "", false, ErrAldaPlayerNotFoundOnPath
	}

	infoCmd := exec.Command(aldaPlayer, "info")
	infoCmd.Stdout = nil
	infoCmd.Stderr = nil
	outputBytes, err := infoCmd.Output()
	if err != nil {
		return "", false, err
	}

	output := string(outputBytes)

	re := regexp.MustCompile(`alda-player ([^\r\n]+)`)
	captured := re.FindStringSubmatch(output)
	if len(captured) < 2 {
		return "", false, fmt.Errorf(
			"unable to parse player version from output: %s", output,
		)
	}
	// captured[0] is "alda-player X.X.X", captured[1] is "X.X.X"
	playerVersion := captured[1]

	if playerVersion != generated.ClientVersion {
		log.Warn().
			Str("clientVersion", generated.ClientVersion).
			Str("playerVersion", playerVersion).
			Msg("`alda` and `alda-player` are different versions. " +
				"Run `alda doctor` to install the correct version of `alda-player`.")
	}

	return aldaPlayer, playerVersion == generated.ClientVersion, nil
}

// spawnPlayer spawns an Alda player process, using the provided absolute path
// to `alda-player`.
//
// Note that this path can be obtained by calling `AldaPlayerPath()`, which also
// checks that the client and player versions are the same and logs a warning if
// they aren't.
func spawnPlayer(playerPath string) error {
	runCmd := exec.Command(playerPath, "run")
	if err := runCmd.Start(); err != nil {
		return err
	}

	log.Info().Msg("Spawned player process.")

	return nil
}

// FillPlayerPool ensures that a minimum desired number of player processes is
// available. Spawns as many player processes as it takes to make that happen.
//
// Returns an error if something goes wrong.
func FillPlayerPool() error {
	// If ALDA_DISABLE_SPAWNING is set to true, we do nothing.
	//
	// This is useful for CI/CD purposes. (See .circleci/config.yml.)
	if os.Getenv("ALDA_DISABLE_SPAWNING") == "yes" {
		log.Info().
			Str("ALDA_DISABLE_SPAWNING", "yes").
			Msg("Skipping filling the player pool.")

		return nil
	}

	playerPath, _, err := AldaPlayerPath()
	if err != nil {
		return err
	}

	players, err := ReadPlayerStates()
	if err != nil {
		return err
	}

	availablePlayers := 0
	for _, player := range players {
		if player.State == "ready" || player.State == "starting" {
			availablePlayers++
		}
	}

	desiredAvailablePlayers := 3
	playersToStart := desiredAvailablePlayers - availablePlayers

	log.Debug().
		Int("availablePlayers", availablePlayers).
		Int("desiredAvailablePlayers", desiredAvailablePlayers).
		Int("playersToStart", playersToStart).
		Msg("Spawning players.")

	results := []<-chan error{}

	for i := 0; i < playersToStart; i++ {
		result := make(chan error)
		results = append(results, result)
		go func() { result <- spawnPlayer(playerPath) }()
	}

	for _, result := range results {
		err := <-result
		if err != nil {
			return err
		}
	}

	return nil
}

// Alda starts player processes in the background as needed when running
// (almost) any command. Most of the time, this is totally transparent to the
// user, as when we get to this point, there is already a player process
// available, so playback is immediate.
//
// However, on the very first run (or the first run after a period of
// inactivity), it's easy for there not to be any Alda player processes
// available yet, so there is a brief (but noticeable) pause before playback
// starts.
//
// To avoid making it look like Alda is "hanging" here while we wait for player
// processes come up, we print a message to make it clear what we're waiting
// for.
func StartingPlayerProcesses() {
	if _, err := FindAvailablePlayer(); err == ErrNoPlayersAvailable {
		fmt.Fprintln(os.Stderr, "Starting player processes...")
	}
}
