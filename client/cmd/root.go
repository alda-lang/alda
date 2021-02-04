package cmd

import (
	"fmt"
	"os"
	"strings"

	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"github.com/spf13/cobra"
)

var verbosity int

// There are certain activities that the Alda CLI performs in the background,
// like sending telemetry and filling the player pool.
//
// We want to avoid prematurely exiting before these activities have completed.
//
// Each activity has a "done" channel, on which a single `true` value is placed
// when the activity completes.
type backgroundActivity struct {
	description string
	done        chan bool
}

// A list of channels, each of which represents an activity happening in the
// background that we want to make sure that we wait for to complete before we
// exit.
//
// Completion of each activity is signaled by placing a single `true` value on
// the channel.
var backgroundActivities []backgroundActivity

func startBackgroundActivity(description string, thunk func()) {
	done := make(chan bool)

	activity := backgroundActivity{description: description, done: done}

	backgroundActivities = append(backgroundActivities, activity)

	go func() {
		thunk()
		done <- true
	}()
}

func AwaitBackgroundActivities() {
	for _, activity := range backgroundActivities {
		log.Debug().
			Str("activity", activity.description).
			Msg("Waiting for background activity to complete.")

		<-activity.done
	}
}

func init() {
	// Inspired by the approach here:
	// https://github.com/spf13/cobra/issues/914#issuecomment-548411337
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &help.UsageError{Cmd: cmd, Err: err}
	})

	rootCmd.PersistentFlags().IntVarP(
		&verbosity, "verbosity", "v", 1, "verbosity level (0-3)",
	)

	for _, cmd := range []*cobra.Command{
		doctorCmd,
		exportCmd,
		instrumentsCmd,
		parseCmd,
		playCmd,
		psCmd,
		replCmd,
		shutdownCmd,
		stopCmd,
		telemetryCmd,
		versionCmd,
	} {
		rootCmd.AddCommand(cmd)
	}
}

func handleVerbosity(cmd *cobra.Command, level int) error {
	switch level {
	case 0:
		log.SetGlobalLevel("error")
	case 1:
		log.SetGlobalLevel("warn")
	case 2:
		log.SetGlobalLevel("info")
	case 3:
		log.SetGlobalLevel("debug")
	default:
		return &help.UsageError{
			Cmd: cmd,
			Err: fmt.Errorf("invalid verbosity level. Valid values are 0-3"),
		}
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "alda",
	Short: "alda: a text-based language for music composition",
	Long: `alda: a text-based language for music composition

Website: https://alda.io
GitHub: https://github.com/alda-lang/alda
Slack: https://slack.alda.io`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return handleVerbosity(cmd, verbosity)
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute parses command-line arguments and runs the Alda command-line client.
func Execute() error {
	// The default logging level is WARN.
	level := 1

	// Cobra is a little bit janky in that there doesn't seem to be a way to add a
	// hook that will _always_ run regardless of the command. PersistentPreRun
	// almost works, but it doesn't run in cases like:
	//
	//   alda
	//   alda --help
	//   alda some-bogus-command
	//
	// We always want to do two things:
	//
	//   1. Set the logging level based on the -v / --verbosity option.
	//   2. Spawn player processes.
	//
	// It's especially important that we do this when the user first installs Alda
	// and they run `alda` or `alda --help` to explore the commands and options.
	// By the time they figure out how to craft an `alda play ...` command, a
	// player process will have fully spawned in the background and they'll hear
	// output immediately, which is a great user experience.
	//
	// As a hacky workaround, we scan through the command line arguments here,
	// make an attempt to set the log level, and spawn player processes, before we
	// invoke Cobra and handle CLI arguments properly.
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-v0", "-v=0", "--verbosity=0":
			level = 0
		case "-v1", "-v=1", "--verbosity=1":
			level = 1
		case "-v2", "-v=2", "--verbosity=2":
			level = 2
		case "-v3", "-v=3", "--verbosity=3":
			level = 3
		}

		if (arg == "-v" || arg == "--verbosity") && (i+1) < len(os.Args) {
			i++
			switch os.Args[i] {
			case "0":
				level = 0
			case "1":
				level = 1
			case "2":
				level = 2
			case "3":
				level = 3
			}
		}
	}

	// Ideally, we could pass in the subcommand here, but we haven't properly
	// parsed it yet (that's done by Cobra when we call rootCmd.Execute() below).
	// This is suboptimal, but I'm working around Cobra's jankiness and I had to
	// make a trade-off.
	//
	// This will only happen if the user uses the correct syntax for specifying
	// verbosity, but specifies a number that isn't in the range 0-3.
	if err := handleVerbosity(rootCmd, level); err != nil {
		return err
	}

	informUserOfTelemetryIfNeeded()

	// Unless the command is one of the exceptions below, Alda will preemptively
	// spawn player processes in the background, up to a desired amount. This
	// helps to ensure that the application will feel fast, because each time you
	// need a player process, there will probably already be one available.
	//
	// Exceptions:
	// * `alda ps` is designed to be run repeatedly, e.g. `watch -n 0.25 alda ps`,
	//   in order to provide a live-updating view of current Alda processes.
	//
	// * `alda shutdown` shuts down a player process (or all of them, if no player
	//   ID or port is specified). It's probably fair to assume that if someone is
	//   running `alda shutdown`, they don't want additional player processes to
	//   be spawned.
	//
	// * `alda doctor` spawns its own processes as part of the checks that it
	//    does, and it simplifies our CI setup if we only spawn those explicit
	//    ones without also spawning some implicit ones here.
	commandIsAnException := false

	// NB: This isn't scientific. If _any_ of the arguments are one of these
	// strings (even if it's an argument other than the command), then that will
	// cause a false positive. But I think that scenario is pretty unlikely, and
	// the consequence is just that it won't fill the player pool on that one run
	// of `alda`. Any other command (e.g. `alda --help`) _will_ fill the player
	// pool, so with typical usage, the odds are high that there will be at least
	// one player process available when you need it.
	for _, arg := range os.Args {
		if arg == "ps" || arg == "shutdown" || arg == "doctor" {
			commandIsAnException = true
		}
	}

	if !commandIsAnException {
		startBackgroundActivity("fill the player pool", func() {
			if err := system.FillPlayerPool(); err != nil {
				log.Warn().Err(err).Msg("Failed to fill player pool.")
			}
		})
	}

	err := rootCmd.Execute()

	// Cobra helpfully gives us cmd.SetFlagErrorFunc to allow us to recognize
	// usage errors due to incorrect/unrecognized flags so that we can treat them
	// specially,, but AFAICT, it doesn't give you any way to do the same thing
	// for usage errors due to unrecognized command names. (╯°□°)╯︵ ┻━┻
	//
	// To work around that, we assume that no other types of errors in the
	// application will begin with the string "unknown command" and do the special
	// handling here.
	if err != nil && strings.HasPrefix(err.Error(), "unknown command") {
		return &help.UsageError{Cmd: rootCmd, Err: err}
	}

	return err
}
