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
	log "alda.io/client/logging"
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
	os.MkdirAll(playersDir, os.ModePerm)

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

// ErrNoPlayersAvailable is the error message that is returned when no player
// process is in an available state.
var ErrNoPlayersAvailable = fmt.Errorf("no players available")

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

	return PlayerState{}, fmt.Errorf("player not found: %s", id)
}

func spawnPlayer() error {
	aldaPlayer, err := exec.LookPath("alda-player")
	if err != nil {
		return err
	}

	// First, we run `alda-player info` and parse the version number from the
	// output, so that we can confirm that the player is the same version as the
	// client.
	infoCmd := exec.Command(aldaPlayer, "info")
	infoCmd.Stdout = nil
	infoCmd.Stderr = nil
	outputBytes, err := infoCmd.Output()
	if err != nil {
		return err
	}

	output := string(outputBytes)

	re := regexp.MustCompile(`alda-player ([^\r\n]+)`)
	captured := re.FindStringSubmatch(output)
	if len(captured) < 2 {
		return fmt.Errorf(
			"unable to parse player version from output: %s", output,
		)
	}
	// captured[0] is "alda-player X.X.X", captured[1] is "X.X.X"
	playerVersion := captured[1]

	// TODO: If the player version is different from the client version, offer to
	// download and install the correct player version.
	if playerVersion != generated.ClientVersion {
		return fmt.Errorf(
			"client version is %s, but player version is %s",
			generated.ClientVersion, playerVersion,
		)
	}

	// Once we've confirmed that the client and player version are the same, we
	// run `alda-player run` to start the player process.
	runCmd := exec.Command(aldaPlayer, "run")
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
		go func() { result <- spawnPlayer() }()
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
	_, err := FindAvailablePlayer()
	if err == ErrNoPlayersAvailable {
		fmt.Fprintln(os.Stderr, "Starting player processes...")
	}
}
