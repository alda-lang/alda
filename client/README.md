# The Alda client

The Alda client is a Go program that parses input (Alda code), turns it into a
data representation of a musical score, and sends instructions to an Alda player
process to perform the score.

## Development

### Requirements

You'll need to have Go installed in order to run the client locally. I'm using
version 1.11.4.

### Setting up your PATH

> If you're already running an Alda player process locally (let's say on port
> 27278), then you can specify that port via the `-p, --port` option and the
> client will send messages to the player you have running, instead:
>
> ```bash
> bin/run play -p 27278 -f ../examples/hello_world.alda
> ```


The client expects to find the executable for the [Alda player](../player)
(`alda-player`) on your PATH. To make this easy when developing, there is a
`bin/player-on-path` script at the root of this repository that will build the
player executable if necessary, and then run its arguments as a command in a
subshell where the directory containing the current build of `alda-player` is
first on the PATH.

For example:

```bash
# NB: These commands are all run from inside the `client` directory.

# This doesn't work unless you have `alda-player` on your PATH:
$ bin/run play -f ../examples/hello_world.alda
Jun 21 19:56:51 WRN cmd/root.go:148 > Failed to fill player pool. error="exec: \"alda-player\": executable file not found in $PATH"
no players available

# This works - a player process is spawned for you, using the current build of
# the player.
$ ../bin/player-on-path bin/run play -f ../examples/hello_world.alda
# ... output elided, building player ...
# ... output elided, building client ...
# 🎵 music 🎶

# Alternatively, you can start an interactive subshell that's set up so that the
# current build of `alda-player` is on the PATH:
$ ../bin/player-on-path
# ... output elided, building player ...

# Now we're in a subshell with a modified PATH:
$ echo $PATH
/home/dave/code/alda/player/target/1.99.0-1e33bd7fcf91c5384d916bb030b918d6e7e20441/non-windows:/home/dave/.local/bin:/home/dave/bin:/home/dave/.bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/games

# Now this works, because the directory with the current build of `alda-player`
# is on the PATH.
$ bin/run play -f ../examples/hello_world.alda
# ... output elided, building client ...
# 🎵 music 🎶
```

### Running the client locally

There are two ways to run the client locally:

* **Basic**: run a script and pass it arguments as if you're running `alda` on
  your PATH. Takes a little bit longer sometimes, but is more convenient most of
  the time and it's exactly like running a release executable.

* **Go toolchain**: use `go run` for faster builds (no need to fully compile the
  executable) and more control over build options. Useful if you're handy with
  Go and you know what you're doing.

#### Basic

To run the client with no arguments, which displays usage information:

```bash
# equivalent to running `alda` or `alda --help`
bin/run
```

To run health checks to determine if Alda can run correctly:

```bash
# equivalent to running `alda doctor`
bin/run doctor
```

To play an Alda source file:

```bash
# equivalent to running `alda play -f ../examples/hello_world.alda`
bin/run play -f ../examples/hello_world.alda
```

#### Go toolchain

There is a tiny bit of code generation going on, which means you need to run
[`go generate`](https://blog.golang.org/generate) first, or else the code won't
compile.

After running `go generate`, you can use all of the usual `go` commands, like
`go build` and `go run`.

Example usage of `go run`:

```bash
# equivalent to running `alda` or `alda --help`
go run main.go

# equivalent to running `alda doctor`
go run main.go doctor

# equivalent to running `alda play -f ../examples/hello_world.alda`
go run main.go play -f ../examples/hello_world.alda
```

### Running tests

There is a comprehensive test suite that you can run with this command:

```bash
go test ./...
```

### OSC API

The client sends instructions to the player in the form of OSC message bundles.

See the [OSC API doc](../player/doc/alda-osc-api.md) and [OSC API
demo](dev/osc/README.md).

## License

Copyright © 2019-2020 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.
