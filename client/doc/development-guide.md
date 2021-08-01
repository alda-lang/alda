# `alda` CLI development guide

## Table of contents

* [Commands](#commands)
* [Logging](#logging)
* [Errors](#errors)
* [Player management](#player-management)
* [Parser](#parser)
* [Score construction](#score-construction)
* [OSC](#osc)

## Commands

> Relevant files: `cmd/root.go`

`alda` is a user-friendly command-line client that performs various Alda-related
functions, including parsing and playing scores and running an Alda REPL server
and/or client session.

We use the [Cobra] library to handle command/option/argument parsing.

Commands (e.g. `alda play`, `alda doctor`) are implemented as instances of
`*cobra.Command`.

## Logging

> Relevant files: `logging/log.go`

We use the [Zerolog][zerolog] library for logging. `logging/log.go` is an
abstraction layer to simplify logging calls throughout the code base.

**`logging/log.go` is the only namespace that should import and use Zerolog
directly.** Everywhere else, we use the `logging/log.go` abstraction layer like
this:

```go
import (
  log "alda.io/client/logging"
)

func something() {
  log.Debug().
    Int("some-value", 42).
    Msg("Doing a thing")
}
```

Note that Zerolog is a structured logging library. Each log line is a set of
key/value pairs accompanied by a concise message.

When running the Alda CLI, the user can specify the logging level via the `-v /
--verbosity` option. The default logging level is 1 (`WARN`), so users shouldn't
typically see log output unless something is wrong or they've specified a higher
verbosity level.

## Errors

> Relevant files: `help/errors.go`

When something goes wrong in a Go function, the convention is to return an
error, typically constructed with `errors.New` or `fmt.Errorf`. This is a normal
thing to do, and we do it all over the Alda code base.

One thing worth calling out, though, is that we differentiate between low-level,
"unfriendly" errors (e.g. the kind that happen unexpectedly when you're doing
some kind of I/O) and "managed" errors that we capture and take care to explain
to the user in a way that's helpful, offering suggestions about how to proceed
whenever possible.

To return a helpful error, use `help.UserFacingErrorf` and try to spend a few
minutes writing a good user-facing error message. Search the code base for
instances of `UserFacingErrorf` for examples.

It's often fine to use `fmt.Errorf` to return an error from a function, in cases
where that error is wrapped higher up the call stack in a `UserFacingError` that
presents the error to the user in a helpful way.

If a non-user-friendly error manages to bubble all the way up without being
wrapped as a `UserFacingError`, we print a message along the lines of "Oops!
Something went wrong. This might be a bug." accompanied by the "unfriendly"
error message and instructions about where to report the bug and seek help.

Our goal is to avoid presenting the user with the "Oops! Something went wrong."
error messages, and to instead display good, user-friendly error messages as
much as we possibly can.

## Player management

> Relevant files: `system/process_management.go`, `cmd/root.go`

When you play a score using the Alda CLI (i.e. `alda play -f some-file.alda`),
the playback occurs asynchronously in a background `alda-player` process. These
player processes take a few seconds or so to initialize, so to make playback
more immediate, the Alda client automatically spawns player processes when it
starts. (See `system.FillPlayerPool`.) The Alda client will automatically find
an available player process to play a score, so that the user never needs to
worry about managing player processes explicitly.

Player processes automatically expire after a period of inactivity, in order to
avoid stale player processes from accumulating over time.

When the Alda client "fills the player pool", it first checks for any existing
player processes that are in the `ready` state. If there are already a
sufficient number of player processes available, no new player processes are
spawned.

The Alda REPL server is a special case. An Alda REPL server will regularly
refill the player pool in the background over time as it runs. (See
`managePlayers` in `repl/player_management.go.`)

## Parser

> Relevant files: `parser/scanner.go`, `parser/parser.go`

Alda uses a simple, hand-rolled parser to parse input (Alda code) into a list of
`model.ScoreUpdate`s. A `ScoreUpdate` is any kind of event that modifies the
score that we are constructing. For example, a `PartDeclaration` event sets the
instrument parts that are currently active; and a `Note` event adds a note to
the score.

The phases of the parser are:

1. **Scanning**: We scan through the input one character at a time and emit a
   list of tokens. (See `Scan` in `parser/scanner.go`.)

2. **Parsing**: We iterate through the list of tokens and emit a list of
   `ScoreUpdate`s. (See `Parse` in `parser/parser.go`.)

## Score construction

> Relevant files: `model/score.go`, `model/*.go`

Once we have the list of `ScoreUpdate`s emitted by the parser, we apply them in
order to construct the score. We start with an empty `*Score` instance, obtained
by calling `model.NewScore()`, and then we use the `Update` function to apply
the list of score updates in order.

### alda-lisp

> Relevant files: `model/lisp.go`

Alda includes a minimal Lisp implementation as a subset of the language, in
order to facilitate adding new features to the language without accumulating
syntax.

Attributes, for example, are implemented as Lisp function calls like `(volume
50)` that return `ScoreUpdate` values. (See `defattribute` in `model/lisp.go`.)

## OSC

> Relevant files: `transmitter/osc.go`

After the input is parsed and the score is constructed, the Alda client sends
instructions to a player process in the form of OSC message bundles.

For information about the types of messages that the player processes accept,
see the [OSC API doc][alda-osc-api] and [OSC API demo][alda-osc-api-demo].

[cobra]: https://github.com/spf13/cobra
[zerolog]: https://github.com/rs/zerolog
[alda-osc-api]: ../../player/doc/alda-osc-api.md
[alda-osc-api-demo]: ../dev/osc/README.md
