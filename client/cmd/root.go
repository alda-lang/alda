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
		fillPlayerPool()
		return defaultUsageFunc(cmd)
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
	rootCmd.SetUsageTemplate(`Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

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

		return handleVerbosity(cmd, verbosity)
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
