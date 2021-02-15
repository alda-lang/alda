# `alda-player` development guide

## Table of contents

* [Commands](#commands)
* [Logging](#logging)
* [State management](#state-management)
* [OSC messages](#osc-messages)
* [MIDI](#midi)

## Commands

> Relevant files: `Main.kt`

`alda-player` is a simple command-line program that runs an Alda player process.
We use the [Clikt] library to handle command/option/argument parsing.

Commands (e.g. `alda-player run`, `alda-player info`) are implemented as
subclasses of the `CliktCommand` class.

## Logging

> Relevant files: `log4j2.xml`, `Main.kt`

We use the lightweight [kotlin-logging] library for logging. Before we
initialize the logger, we set a couple of system properties, `logPath` and
`playerId`. These are used in the logging config to control the location of the
log file and the ID of the player process, which is included at the beginning of
every log line. (These are the same player IDs that you see when you run `alda
ps` to list running player processes).

The location of the log files is based on platform-specific conventions (e.g.
the XDG specification for Linux), as codified in the [directories-jvm] library.


> See the `cacheDir` row in [this table][file-locations] for reference about
> where to find the logs on your system. (Logs are also printed to stdout when
> you're running the player process, which is usually good enough!)
>
> On my Ubuntu 16.04 machine, the current log file is
`/home/dave/.cache/alda/logs/alda-player.log`. Older log files are found in the
same directory, with names like `alda-player-2021-02-10.log`.

## State management

> Relevant files: `StateManager.kt`

It is important that the Alda client can easily discover which player processes
are available to play a score. When a player process comes up, it creates a
StateManager instance, which runs a background thread that continuously updates
a **state file** in the cache directory.

> See the `cacheDir` row in [this table][file-locations] if you need help
> locating the player state files on your system.
>
> On my Ubuntu 16.04 machine, the player state files are found in
> `/home/dave/.cache/alda/state/players/1.99.3/` (where `1.99.3` is the current
> Alda version, controlled by the `VERSION` resource file in this repo).

Here is some example state file content:

```json
{"expiry" : 1613251867419, "port" : 36841, "state" : "ready"}
```

In order to discover available player processes, the Alda client looks in the
player state file directory for any state files that show the player's state is
`ready`.

A player process begins in the `starting` state. After initializing, its state
becomes `ready`. Once a player process starts receiving messages from a client,
it becomes `active`.

A player process shuts down after a random length of inactivity between 5 and 10
minutes. This helps to ensure that a bunch of old player processes aren't left
hanging around, running idle in the background. The time at which a player
process is set to expire is represented in the state file as `expiry`, as a Unix
timestamp (i.e. time since the Unix epoch) in milliseconds.

## OSC messages

> Relevant files: `Receiver.kt`, `Parser.kt`

The Alda client sends messages to the Alda player using the [Open Sound
Control][osc] protocol over TCP. The Alda OSC API is [documented
here][alda-osc-api].

Incoming OSC messages/bundles are placed onto a queue and parsed and processed
in the order that they were received.

## MIDI

> Relevant files: `MidiEngine.kt`

Although the [Java MIDI synthesizer][java-midi-synth] is currently Alda's only
method of producing audio, there are plans to (eventually) include other types
of audio instruments, such as waveform synthesizers and audio samplers.

In the future, when we have instruments like those, we will likely still use the
[Java MIDI sequencer][java-midi-sequencer] to manage the scheduling and
sychronization of musical events, the same way we are doing now. This is
possible by using [MetaMessages][java-meta-message] to schedule arbitrary events
in a MIDI sequence.

We are already using MetaMessages for a handful of things. For example, to
support live coding, we regularly schedule "continue" metamessages that we
handle by making sure that the sequencer is still running once it reaches the
end of the current sequence.

Another thing we use MetaMessages for is deferring the scheduling of events in a
pattern until right before the pattern is due to be played. This is useful
because patterns are mutable, which allows composers to do fun things like loop
a pattern and change the notes in the pattern live during a performance.

[clikt]: https://ajalt.github.io/clikt/
[kotlin-logging]: https://github.com/MicroUtils/kotlin-logging
[directories-jvm]: https://github.com/dirs-dev/directories-jvm
[file-locations]: https://github.com/dirs-dev/directories-jvm#projectdirectories
[osc]: https://en.wikipedia.org/wiki/Open_Sound_Control
[alda-osc-api]: ./alda-osc-api.md
[java-midi-synth]: https://docs.oracle.com/javase/7/docs/api/javax/sound/midi/Synthesizer.html
[java-midi-sequencer]: https://docs.oracle.com/javase/7/docs/api/javax/sound/midi/Sequencer.html
[java-meta-message]: https://docs.oracle.com/javase/7/docs/api/javax/sound/midi/MetaMessage.html
