# The Alda client

The Alda client is a Go program that parses input (Alda code), turns it into a
data representation of a musical score, and sends instructions to an Alda
[player](../player) process to perform the score.

...That's the end goal, anyway. We have a lot of work to do!

In the meantime, for development purposes, the program here just sends example
OSC messages to be handled by the player. Each example has a short identifier so
that you can select it from the command-line.

## OSC API

See [Alda OSC API](../player/doc/alda-osc-api.md).

### Demo

> **TODO**: Turn this into proper tests.

#### Setup

* Go is needed in order to run the client. I'm using version 1.11.4.

* Start the [player](../player) on port 27278 (or another port of your choosing)
  and leave it running.

Now you should be able to use the client to send messages on the same port:

```bash
go run osc_api_demo/main.go 27278 EXAMPLE_NAME
```

The client takes two arguments: the port on which to send messages, and the name
of the example to send.

If the example name is omitted, example `1` is used.

#### Examples

> Unless otherwise specified, examples are played on track 1.

| name | description |
|--|--|
| 1 | A single note |
| 16fast | Sixteen 16th notes on the same (randomly chosen) pitch |
| pat1 | A five-note pattern called `simple` is defined and played once |
| pat2 | A five-note pattern called `simple` is defined and played twice |
| patchange | Redefines the `simple` pattern to four randomly chosen notes |
| patx | Clears the contents of the `simple` pattern
| perc | Turns track 1 into a percussion track |
| play | Sends the system "play" message |
| stop | Sends the system "stop" message |

---

Queue up four single notes in time:

```bash
# send these in rapid succession
go run osc_api_demo/main.go 27278 1
go run osc_api_demo/main.go 27278 1
go run osc_api_demo/main.go 27278 1
go run osc_api_demo/main.go 27278 1
```

Queue up a four measures, each containing one randomly chosen note played as
sixteen fast 16th notes:


```bash
# send these in rapid succession
go run osc_api_demo/main.go 27278 16fast
go run osc_api_demo/main.go 27278 16fast
go run osc_api_demo/main.go 27278 16fast
go run osc_api_demo/main.go 27278 16fast
```

Queue up interspersed single notes, 16th note bars, and instances of the
`simple` pattern:

```bash
# send these in rapid succession
go run osc_api_demo/main.go 27278 1
go run osc_api_demo/main.go 27278 16fast
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 1
go run osc_api_demo/main.go 27278 16fast
go run osc_api_demo/main.go 27278 pat1
```

Queue up the `simple` pattern to be played several times, and change the pattern
while it's playing:

```bash
# send these in rapid succession
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1
go run osc_api_demo/main.go 27278 pat1

# send while the pattern is playing
go run osc_api_demo/main.go 27278 patchange

# send while the new pattern is playing
go run osc_api_demo/main.go 27278 patchange
```

## License

Copyright © 2019 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.
