package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"alda.io/client/color"
	"alda.io/client/generated"
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/util"

	"github.com/daveyarwood/go-osc/osc"
)

const reasonableTimeout = 20 * time.Second

func DeletePlayerStateFile(playerID string) error {
	path := CachePath(
		"state", "players", generated.ClientVersion, playerID+".json",
	)

	log.Debug().
		Str("path", path).
		Msg("Deleting player state file.")

	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func cleanUpStaleStateFiles(stateDir string) error {
	if err := filepath.WalkDir(
		stateDir,
		func(path string, info os.DirEntry, err error) error {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}

			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			fileInfo, err := os.Stat(path)
			if errors.Is(err, os.ErrNotExist) {
				return nil
			} else if err != nil {
				return err
			}

			timeSinceModified := time.Since(fileInfo.ModTime())

			if timeSinceModified > time.Minute*2 {
				log.Debug().
					Float64("seconds-since-modified", timeSinceModified.Seconds()).
					Str("path", path).
					Msg("Deleting stale file.")

				err := os.Remove(path)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				return nil
			}

			return nil
		},
	); err != nil {
		return err
	}

	// NOTE: I initially also walked the directory a second time and attempted to
	// delete empty directories, but I found that it was too easy to run into a
	// race condition where the directory, so we delete it, but right before we
	// delete it, a player starts up and puts a file in the directory, and then
	// the directory gets deleted even though it isn't actually empty at that
	// point.
	//
	// Player processes also delete empty directories for old versions, and it
	// seems to behave well, so we don't need to include the deletion of empty
	// directories part here in the client.

	return nil
}

// Each `alda-player` process creates a state file where it tracks its own
// state. The `alda` client uses this information to list player processes
// (`alda ps`) or to find an available player process to play a score.
//
// Each player process deletes its own state file when it exits, but this
// doesn't happen in exceptional scenarios like an out of memory error or the
// process being forcibly terminated (e.g. `kill -9`).
//
// Because it will always be possible for player processes to die before they
// can clean up their own state files, both `alda-player` and `alda` routinely
// check for and clean up stale player state files.
func CleanUpStaleStateFiles() error {
	for _, stateDir := range []string{
		CachePath("state", "players"),
		CachePath("state", "repl-servers"),
	} {
		if err := cleanUpStaleStateFiles(stateDir); err != nil {
			return err
		}
	}

	return nil
}

// PlayerState describes the current state of a player process. These states are
// continously written to files by each player process. (See: StateManager.kt.)
type PlayerState struct {
	State  string
	Port   int
	Expiry int64
	ID     string
}

// REPLServerState describes the current state of an Alda REPL server process.
// These states are continously written to files by each Alda REPL process.
// (See: repl/server.go.)
type REPLServerState struct {
	Port int    `json:"port"`
	ID   string `json:"id"`
}

func processFiles(
	directory string,
	process func(filename string, contents []byte, readError error),
) error {
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return err
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		filepath := filepath.Join(directory, filename)
		contents, readError := os.ReadFile(filepath)
		process(filename, contents, readError)
	}

	return nil
}

// FIXME: There is a lot of duplication between ReadPlayerStates and
// ReadREPLServerStates because prior to Go 1.18, Go didn't have generics.
//
// TODO: Refactor these 2 functions into a generic function parameterized on the
// kind of state (PlayerState vs. REPLServerState).

// ReadPlayerStates reads all of the player state files in the Alda cache
// directory and returns a list of player state structs describing the current
// state of each player process.
//
// It's common for these files to be unreadable/empty, e.g. if a player process
// is busy writing to the file. In the event that the file is unreadable or
// cannot be parsed as JSON, we skip that file. The goal is that we end up with
// a list of only known, valid player states.
//
// Returns an error if something goes horribly wrong, e.g. we cannot list the
// files in the directory for some reason.
func ReadPlayerStates() ([]PlayerState, error) {
	if err := CleanUpStaleStateFiles(); err != nil {
		log.Warn().Err(err).Msg("Failed to clean up stale state files.")
	}

	states := []PlayerState{}

	if err := processFiles(
		CachePath("state", "players", generated.ClientVersion),
		func(filename string, contents []byte, readError error) {
			var state PlayerState

			// `readError` is initially a possible error reading the file.
			if readError == nil {
				readError = json.Unmarshal(contents, &state)
			}

			// Now, `readError` is either a possible error reading the file or a
			// possible error parsing its contents as JSON.
			//
			// If either of these scenarios, we log a warning and move on to the next
			// file.
			if readError != nil {
				log.Warn().
					Err(readError).
					Str("filename", filename).
					Msg("Failed to read player state")
				return
			}

			state.ID = strings.Replace(filename, ".json", "", 1)

			states = append(states, state)
		},
	); err != nil {
		return nil, err
	}

	return states, nil
}

// ReadREPLServerStates reads all of the REPL server state files in the Alda
// cache directory and returns a list of REPLServerState structs describing the
// current state of each REPL server process.
//
// It's common for these files to be unreadable/empty, e.g. if a REPL server
// process is busy writing to the file. In the event that the file is unreadable
// or cannot be parsed as JSON, we skip that file. The goal is that we end up
// with a list of only known, valid REPL server states.
//
// Returns an error if something goes horribly wrong, e.g. we cannot list the
// files in the directory for some reason.
func ReadREPLServerStates() ([]REPLServerState, error) {
	states := []REPLServerState{}

	processFiles(
		CachePath("state", "repl-servers"),
		func(filename string, contents []byte, readError error) {
			var state REPLServerState

			// `readError` is initially a possible error reading the file.
			if readError == nil {
				readError = json.Unmarshal(contents, &state)
			}

			// Now, `readError` is either a possible error reading the file or a
			// possible error parsing its contents as JSON.
			//
			// If either of these scenarios, we log a warning and move on to the next
			// file.
			if readError != nil {
				log.Warn().
					Err(readError).
					Str("filename", filename).
					Msg("Failed to read REPL server state")
				return
			}

			state.ID = strings.Replace(filename, ".json", "", 1)

			states = append(states, state)
		},
	)

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
		color.Aurora.Bold("alda-player"),
		color.Aurora.Bold("alda"),
		color.Aurora.Bold("alda-player"),
		color.Aurora.Bold("alda-player"),
		color.Aurora.BrightYellow("alda doctor"),
		color.Aurora.Bold("y"),
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
		color.Aurora.BrightYellow("alda doctor"),
		color.Aurora.BrightYellow("alda ps"),
		color.Aurora.BrightYellow(playerLogFile),
		color.Aurora.BrightYellow("alda-player -v run -p 27278"),
		color.Aurora.BrightYellow("alda -v2 play -p 27278 -c \"piano: c12 e g > c4\""),
	)
}

// PingPlayer sends a ping message to the specified port number, where a player
// process is expected to be listening.
//
// If the first ping doesn't go through, it is tried repeatedly for the
// `reasonableTimeout` duration, so as to avoid concluding too hastily that a
// player process is not reachable (e.g. it might be in the process of coming up
// and will be available shortly).
//
// If the player is reachable, the OSC client that we created in order to ping
// the player process is returned, so that it can be reused if desired.
//
// Returns an error if the player isn't reachable after the timeout duration.
func PingPlayer(port int) (*osc.Client, error) {
	log.Debug().
		Int("port", port).
		Msg("Waiting for player to respond to ping.")

	client := osc.NewClient("localhost", port, osc.ClientProtocol(osc.TCP))

	err := util.Await(
		func() error {
			return client.Send(osc.NewMessage("/ping"))
		},
		reasonableTimeout,
	)

	return client, err
}

// FindAvailablePlayer finds a player that is in an available state, confirms
// that it can be reached by sending a ping, and returns current information
// about the player's state.
//
// In the case where a player appears to exist in the available state, but is
// not reachable (e.g. if the player suddenly died and wasn't able to clean up
// its own state file), FindAvailablePlayer will delete that player's state file
// and try again with other players that appear to be in the available state,
// and will also fill the player pool to help ensure that we don't run out of
// players.
//
// Returns `ErrNoPlayersAvailable` if no player is currently in an available
// state.
func FindAvailablePlayer() (PlayerState, error) {
	players, err := ReadPlayerStates()
	if err != nil {
		return PlayerState{}, err
	}

	for _, player := range players {
		if player.State != "ready" {
			continue
		}

		if _, err := PingPlayer(player.Port); err != nil {
			log.Warn().
				Interface("player", player).
				Err(err).
				Msg("Failed to reach player process. Will try another one.")

			if err := DeletePlayerStateFile(player.ID); err != nil {
				return PlayerState{}, err
			}

			if err := FillPlayerPool(); err != nil {
				return PlayerState{}, err
			}

			return FindAvailablePlayer()
		}

		return player, nil
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

	// FIXME: repl/player_management.go does a check whether the error message
	// returned by FindPlayerByID starts with "No player was found". If we ever
	// change the verbiage here, we'll need to adjust that other code accordingly.
	//
	// TODO: Consider adding an optional error code to UserFacingErrors that we
	// could check for instead.
	return PlayerState{}, help.UserFacingErrorf(
		`No player was found with the ID %s.

To list the current player processes, you can run %s.

You can also omit the %s / %s option, and Alda will find a player process
for you automatically.`,
		color.Aurora.BrightYellow(id),
		color.Aurora.BrightYellow("alda ps"),
		color.Aurora.BrightYellow("-i"),
		color.Aurora.BrightYellow("--player-id"),
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
