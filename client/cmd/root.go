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

// We want to ensure that we fill the player pool in a variety of situations,
// including `alda`, `alda --help`, `alda some-nonexistent-command`, etc.
// Unfortunately, Cobra doesn't give you an easy way to consistently do that.
// PersistentPreRunE _should_ do the trick, but it doesn't run in exceptional
// scenarios like where the user is using flags incorrectly or the user is
// requesting --help.
//
// As a result, we have to hook into "UsageFunc", the function that gets run
// when Cobra prints usage information. But it's possible that both the
// PersistentPreRunE _and_ the UsageFunc will be run, in cases like `alda` being
// run without arguments.
//
// To ensure that we don't double-fill the pool in those scenarios, we use this
// boolean to keep track of whether or not we're already doing it.
var fillingPlayerPool = false

// Fills the player pool, unless we're already doing it.
func fillPlayerPool() {
	if !fillingPlayerPool {
		startBackgroundActivity("fill the player pool", func() {
			if err := system.FillPlayerPool(); err != nil {
				log.Warn().Err(err).Msg("Failed to fill player pool.")
			}
		})
	}

	fillingPlayerPool = true
}

func sendTelemetry(command string) {
	// We don't send telemetry if the user runs `alda` without a subcommand,
	// because that's effectively the same thing as running `alda --help`, and we
	// don't send telemetry when the user is requesting `--help` either.
	if command == "alda" {
		return
	}

	// We will not record telemetry if the user has disabled it by running `alda
	// telemetry --disable`.
	status, err := readTelemetryStatus()

	// If the telemetry status file contains unexpected content, or if we couldn't
	// read the file for some reason, the only reasonable thing to do is not to
	// record telemetry.
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Couldn't determine whether telemetry is enabled.")
		return
	}

	// Don't record telemetry if the user has opted out.
	if status == TelemetryDisabled {
		return
	}

	startBackgroundActivity("send telemetry", func() {
		if err := sendTelemetryRequest(command); err != nil {
			log.Debug().Err(err).Msg("Failed to send telemetry.")
		}
	})
}

func cleanUpRenamedExecutables() {
	startBackgroundActivity("clean up renamed executables", func() {
		renamedExecutables, err := system.FindRenamedExecutables()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to search for renamed executables.")
			return
		}

		for _, filepath := range renamedExecutables {
			if err := os.Remove(filepath); err != nil {
				log.Warn().
					Err(err).
					Str("filepath", filepath).
					Msg("Failed to delete renamed executable")
			} else {
				log.Debug().
					Str("filepath", filepath).
					Msg("Deleted renamed executable.")
			}
		}
	})
}

func init() {
	// Inspired by the approach here:
	// https://github.com/spf13/cobra/issues/914#issuecomment-548411337
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &help.UsageError{Cmd: cmd, Err: err}
	})

	// Cobra doesn't run my PersistentPreRunE function when there is a usage error
	// or `--help` is requested. So, we have to wrap the default usage function
	// here with the behavior that we want.
	defaultUsageFunc := rootCmd.UsageFunc()
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		// handleVerbosity returns a usage error if the supplied verbosity level is
		// invalid. We're choosing to ignore it here because the UsageFunc is called
		// all over the place, e.g. as a side effect of constructing the usage
		// string (see UsageString() in Cobra). In light of this, it feels dangerous
		// to override the default UsageFunc with a function that might return an
		// error for a reason unrelated to printing usage information.
		//
		// When an invalid verbosity level is supplied, we fallback to the default
		// log level, 1 (warn).
		_ = handleVerbosity(cmd)

		informUserOfTelemetryIfNeeded()

		fillPlayerPool()

		return defaultUsageFunc(cmd)
	})

	// In some cases, the UsageFunc runs. In some cases, the HelpFunc runs,
	// followed by the UsageFunc. I've even seen the UsageFunc run twice for some
	// weird reason.
	//
	// The only reason we need a custom HelpFunc is so that in the case of
	// informing the user of telemetry on the very first run of Alda, we can make
	// sure that we print the notice about telemetry first, before the help text
	// (instead of before the usage text, which is further down).
	//
	// Note that `informUserOfTelemetryIfNeeded` is idempotent. After informing
	// the user of telemetry, it writes a `telemetry-status` file which makes it
	// so that if we run `informUserOfTelemetryIfNeeded` again, it won't do
	// anything because we already informed the user of telemetry.
	defaultHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// See the line that looks like this above in SetUsageFunc(...)
		_ = handleVerbosity(cmd)

		informUserOfTelemetryIfNeeded()

		defaultHelpFunc(cmd, args)
	})

	// This is almost identical to Cobra's default usage template found in the
	// source code. I just removed the `{{if .Runnable}}{{.UseLine}}{{end}}` part,
	// i.e. the part that suggests that running `alda` with no subcommand is a way
	// to use Alda.
	//
	// Usually, Cobra gets this right, but in this case, because I implemented a
	// `RunE` for the root command, Cobra (reasonably) thinks that it should tell
	// the user that both `alda` and `alda [command]` are ways to run Alda, when
	// in fact, you can't really do anything without a subcommand.
	//
	// Update: I also had to add an {{else}} branch under Usage: because it was
	// just showing up as Usage: with no content in most cases. I'm not sure if
	// this is my fault or Cobra's, but I fixed it in the template.
	rootCmd.SetUsageTemplate(`Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{else}}
  {{.CommandPath}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

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
		updateCmd,
		versionCmd,
	} {
		rootCmd.AddCommand(cmd)
	}
}

func handleVerbosity(cmd *cobra.Command) error {
	switch verbosity {
	case 0:
		log.SetGlobalLevel("error")
	case 1:
		log.SetGlobalLevel("warn")
	case 2:
		log.SetGlobalLevel("info")
	case 3:
		log.SetGlobalLevel("debug")
	default:
		log.SetGlobalLevel("warn")

		return &help.UsageError{
			Cmd: cmd,
			Err: fmt.Errorf(
				"invalid verbosity level (%d). Valid levels are 0-3",
				verbosity,
			),
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
		if err := handleVerbosity(cmd); err != nil {
			return err
		}

		cleanUpRenamedExecutables()

		informUserOfTelemetryIfNeeded()

		sendTelemetry(cmd.Name())

		// Unless the command is one of the exceptions below, Alda will preemptively
		// spawn player processes in the background, up to a desired amount. This
		// helps to ensure that the application will feel fast, because each time
		// you need a player process, there will probably already be one available.
		//
		// Exceptions:
		// * `alda ps` is designed to be run repeatedly, e.g.
		//   `watch -n 0.25 alda ps`, in order to provide a live-updating view of
		//   current Alda processes.
		//
		// * `alda shutdown` shuts down a player process (or all of them, if no
		//   player ID or port is specified). It's probably fair to assume that if
		//   someone is running `alda shutdown`, they don't want additional player
		//   processes to be spawned.
		//
		// * `alda doctor` spawns its own processes as part of the checks that it
		//   does, and it simplifies our CI setup if we only spawn those explicit
		//   ones without also spawning some implicit ones here.
		switch cmd.Name() {
		case "ps", "shutdown", "doctor":
			// Don't fill the player pool.
		default:
			fillPlayerPool()
		}

		return nil
	},

	// I think this is equivalent to the default behavior that Cobra gives you
	// when you run the root command with no subcommand; it just prints the help
	// text.
	//
	// Why am I doing it explicitly like this? Because if I don't have a RunE
	// function, Cobra won't run my PersistentPreRunE function.
	//
	// ...And because I have a RunE now, I had to also customize the UsageTemplate
	// (see init()) to omit the part that suggests that running `alda` without a
	// subcommand is a way to use Alda. Hacks upon hacks.
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},

	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute parses command-line arguments and runs the Alda command-line client.
func Execute() error {
	err := rootCmd.Execute()

	// Cobra helpfully gives us cmd.SetFlagErrorFunc to allow us to recognize
	// usage errors due to incorrect/unrecognized flags so that we can treat them
	// specially,, but AFAICT, it doesn't give you any way to do the same thing
	// for usage errors due to unrecognized command names. (╯°□°)╯︵ ┻━┻
	//
	// To work around that, we assume that no other types of errors in the
	// application will begin with the string "unknown command" and do the special
	// handling here.
	//
	// Worth noting: Cobra seems to be failing to set the global `verbosity` flag
	// value in this scenario, which means in a case like `alda -v3 bogus-cmd`,
	// the `-v3` part is ignored by Cobra and `rootCmd.Execute()` just returns the
	// "unknown command" error. This is not ideal, but it's sort of an exceptional
	// scenario, and the default log level of 1 (warn) is reasonable; the
	// important part is that the user can see that they've entered an unknown
	// command.
	if err != nil && strings.HasPrefix(err.Error(), "unknown command") {
		return &help.UsageError{Cmd: rootCmd, Err: err}
	}

	return err
}
