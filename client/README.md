# The Alda client

The Alda client is a Go program that parses input (Alda code), turns it into a
data representation of a musical score, and sends instructions to an Alda player
process to perform the score.

## Development

> Once you're all set up to run the Alda client locally, have a look at the
> [development guide](./doc/development-guide.md) for a high-level overview of
> the code base.

### Requirements

You'll need to have Go installed in order to build and run the client locally.

Go 1.16 or newer is required.

### tl;dr

Use `bin/run` to run both the Alda client and player locally.

Examples:

```bash
# equivalent to running `alda --help`
bin/run --help

# equivalent to running `alda doctor`
bin/run doctor

# equivalent to running `alda play -f ../examples/hello_world.alda`
# (Alda player processes are started in the background)
bin/run play -f ../examples/hello_world.alda
```

### `alda-player` and your PATH

> If you're already running an Alda player process locally (let's say on port
> 27278), then you can specify that port via the `-p, --port` option and the
> client will send messages to the player you have running, instead:
>
> ```bash
> bin/run play -p 27278 -f ../examples/hello_world.alda
> ```
>
> If you're developing this way, then you don't need to do anything special with
> your PATH.

The client expects to find the executable for the [Alda player](../player)
(`alda-player`) on your PATH. To make this easy when developing, there is a
`bin/player-on-path` script at the root of this repository that will build the
player executable if necessary, and then run its arguments as a command in a
subshell where the directory containing the current build of `alda-player` is
first on the PATH.

To make this _even easier_, the `bin/run` convenience script already includes an
invocation of `bin/player-on-path`, so you don't even need to remember to use
`bin/player-on-path`. **You can just use `bin/run` and the current build of the
player will be available on your PATH automatically.**

For example:

```bash
# NB: These commands are all run from inside the `client` directory.

# This doesn't work unless you have `alda-player` on your PATH:
$ go run main.go play -f ../examples/hello_world.alda
Starting player processes...
Dec 27 21:41:17 WRN cmd/root.go:159 > Failed to fill player pool. error="exec: \"alda-player\": executable file not found in $PATH"
no players available
exit status 1

# This works - a player process is spawned for you, using the current build of
# the player.
$ bin/run play -f ../examples/hello_world.alda
# ... output elided, building player ...
# ... output elided, building client ...
Starting player processes...
Playing...
# 🎵 music 🎶

# Alternatively, you can start an interactive subshell that's set up so that the
# current build of `alda-player` is on the PATH:
$ ../bin/player-on-path
# ... output elided, building player ...

# Now we're in a subshell with a modified PATH:
$ echo $PATH
/home/dave/code/alda/player/target/1.99.2-d31454955bdee26b844224b7a090d3a06d744090/non-windows:/home/dave/.local/bin:/home/dave/bin:/home/dave/.bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/games

# Now this works, because the directory with the current build of `alda-player`
# is on the PATH.
$ go run main.go play -f ../examples/hello_world.alda
Playing...
# 🎵 music 🎶
```

### `bin/run` vs. `go run`

There are two ways to run the client locally:

* **Basic**: run the `bin/run` script and pass it arguments as if you're running
  `alda` on your PATH. This takes a little bit longer sometimes, but it's more
  convenient most of the time and it's exactly like running a release
  executable.

* **Go toolchain**: use `go run` for faster builds (no need to fully compile the
  executable) and more control over build options. This is useful if you're more
  comfortable with the Go CLI, or if you'd prefer not to wait for a full build
  every time you want to run the client locally after making changes.

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

## License

Copyright © 2019-2021 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.
