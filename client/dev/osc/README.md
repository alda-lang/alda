# Alda OSC API demo

This demo program sends example OSC messages to be handled by the player. Each
example has a short identifier so that you can select it from the command-line.

## Setup

* Go 1.19 or higher is needed in order to run the demo.

* Start the [player](../../player) on port 27278 (or another port of your
  choosing) and leave it running.

Now you should be able to use the client to send messages on the same port:

```bash
go run dev/osc/main.go 27278 EXAMPLE_NAME
```

The client takes two arguments: the port on which to send messages, and the name
of the example to send.

If the example name is omitted, example `1` is used.

## Examples

> Unless otherwise specified, examples are played on track 1.

| name | description |
|--|--|
| 1 | A single note |
| 16fast | Sixteen 16th notes on the same (randomly chosen) pitch |
| 2infinity | Plays two random patterns concurrently on the same track, indefinitely |
| 2loops | Plays two random patterns concurrently on the same track, 4 times each |
| clear | Sends the system "clear" message |
| clear1 | Sends the track "clear" message |
| drums | Plays a short percussion demo |
| export | Sends the system "export" message to export the contents of the player's MIDI sequencer to `/tmp/alda-test.mid` |
| pat1 | A five-note pattern called `simple` is defined and played once |
| pat2 | A five-note pattern called `simple` is defined and played twice |
| pat3 | A five-note pattern called `simple` is defined and played thrice |
| patchange | Redefines the `simple` pattern to four randomly chosen notes |
| patfin | Stops looping the `simple` pattern |
| patloop | Loops the `simple` pattern indefinitely |
| patx | Clears the contents of the `simple` pattern
| play | Sends the system "play" message |
| shutdown | Sends the system "shutdown" message |
| stop | Sends the system "stop" message |
| tempos | Sends a bunch of arbitrary tempo change messages |

---

Queue up four single notes in time:

```bash
# send these in rapid succession
go run dev/osc/main.go 27278 1
go run dev/osc/main.go 27278 1
go run dev/osc/main.go 27278 1
go run dev/osc/main.go 27278 1
```

Queue up a four measures, each containing one randomly chosen note played as
sixteen fast 16th notes:


```bash
# send these in rapid succession
go run dev/osc/main.go 27278 16fast
go run dev/osc/main.go 27278 16fast
go run dev/osc/main.go 27278 16fast
go run dev/osc/main.go 27278 16fast
```

Queue up interspersed single notes, 16th note bars, and instances of the
`simple` pattern:

```bash
# send these in rapid succession
go run dev/osc/main.go 27278 1
go run dev/osc/main.go 27278 16fast
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 1
go run dev/osc/main.go 27278 16fast
go run dev/osc/main.go 27278 pat1
```

Queue up the `simple` pattern to be played several times, and change the pattern
while it's playing:

```bash
# send these in rapid succession
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1
go run dev/osc/main.go 27278 pat1

# send while the pattern is playing
go run dev/osc/main.go 27278 patchange

# send while the new pattern is playing
go run dev/osc/main.go 27278 patchange
```

## License

Copyright © 2019-2026 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.
