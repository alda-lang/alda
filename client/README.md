# The Alda client

The Alda client is a Go program that parses input (Alda code), turns it into a
data representation of a musical score, and sends instructions to an Alda
[player](../player) process to perform the score.

## OSC API

The client sends instructions to the player in the form of OSC message bundles.

See the [OSC API doc](../player/doc/alda-osc-api.md) and [OSC API
demo](osc_api_demo/README.md).

## Development

### Requirements

* The client is written in Go. I'm using version 1.11.4.

* The player is written in Kotlin and developed via Gradle. There is a `gradlew`
  wrapper script checked into the repo that makes it so that you don't need to
  have Gradle installed in order to run the player.

### Demo

I haven't gotten around to implementing a nice CLI interface for the Alda client
with option parsing, etc., but for development purposes, I've set up a `main.go`
that starts a player subprocess (piping output into the foreground) and uses it
to play an Alda source file.

Example usage:

```bash
go run main.go ../examples/hello_world.alda
```

If you have an Alda player process already running (let's say on port 27278),
you can supply the port as a first argument and `main.go` will skip starting a
player subprocess and send messages to the player you have running, instead:

```bash
go run main.go 27278 ../examples/hello_world.alda
```

## License

Copyright © 2019-2020 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.
