# [FIXME] README 1: alda-lang/alda

<p align="center">
  <a href="http://alda.io">
    <img src="alda-logo.png" alt="alda logo">
  </a>

  <h2 align=center>a music programming language for musicians</h2>

  <p align="center">
  <b><a href="#installation">Installation</a></b>
  |
  <b><a href="doc/index.md">Docs</a></b>
  |
  <b><a href="CHANGELOG.md">Changelog</a></b>
  |
  <b><a href="#contributing">Contributing</a></b>

  <br>
  <br>

  <a href="http://slack.alda.io">
  <img src="http://slack.alda.io/badge.svg" alt="Join us on Slack!">
  </a>
  <i>composers chatting</i>
  </p>
</p>

*New to Alda? You may be interested in reading [this blog post][alda-blog-post] as an introduction.*

Inspired by other music/audio programming languages such as [PPMCK][ppmck],
[LilyPond][lilypond] and [ChucK][chuck], Alda aims to be a powerful and flexible
programming language for the musician who wants to easily compose and generate
music on the fly, using only a text editor.  Alda is designed in a way that
equally favors aesthetics, flexibility and ease of use, with (eventual) support
for the text-based creation of all manner of music: classical, popular,
  chiptune, electroacoustic, and more!

[alda-blog-post]: https://blog.djy.io/alda-a-manifesto-and-gentle-introduction/
[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu

## Features

* Easy to understand, markup-like syntax
* Designed for musicians who don't know how to program, and programmers who
  don't know how to music
* A score is a text file that can be played using the `alda` command-line tool
* [Interactive REPL](doc/alda-repl.md) lets you enter Alda code and hear the
  results in real time
* Supports [writing music
  programmatically](doc/writing-music-programmatically.md) (for algorithmic
  composition, live coding, etc.)
* Create MIDI music using any of the instruments in the [General MIDI Sound
  Set][gm-sound-set]

[gm-sound-set]: http://www.midi.org/techspecs/gm1sound.php

### TODO

* [Define and use waveform synthesis instruments](https://github.com/alda-lang/alda/issues/100)
* [Import MIDI files](https://github.com/alda-lang/alda-core/issues/25)
* [Export to MusicXML](https://github.com/alda-lang/alda-core/issues/3) for inter-operability with other music software
* [A more robust REPL](https://github.com/alda-lang/alda-client-java/issues/2), tailor-made for editing scores interactively

If you're a developer and you'd like to help, come on in -- [the water's fine](#contributing)!

## Syntax example

    piano: o3
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4
    << g1/>g/>g/b/>d/g

For more examples, see these [example scores](https://github.com/alda-lang/alda-core/tree/master/examples).

## Installation

> You must have [Java](https://www.java.com/en/download) 8+ installed on your system in order to run Alda.
>
> (Chances are, you already have a recent enough version of Java installed.)

### Mac OS X / Linux

* Go to the [latest release](https://github.com/alda-lang/alda/releases/latest) page and download `alda`.

* Make the file executable:

        chmod +x alda

* Make `alda` available on your `$PATH`:

  > Using `/usr/local/bin` here as an example;
  > you can use any directory on your `$PATH`.

      mv alda /usr/local/bin

### Windows

* Go to the [latest release](https://github.com/alda-lang/alda/releases/latest) page and download `alda.exe`.

* Make the file executable:
  * Go to your downloads folder, right click `alda.exe` to open up its file properties, and click `unblock`

* Copy `alda.exe` to a location that makes sense for you. If you follow standard Windows conventions, this means creating a folder called `Alda` in your `Program Files (x86)` folder, and then moving the `alda.exe` file into it.

* Make `alda` available on your `PATH`:
  *  Go to the Windows `System` Control Panel option, select `Advanced System Settings` and then click on `Environment Variables`, then edit the `PATH` variable (either specifically for your user account or for the system in general) and add `;C:\Program Files (x86)\Alda` to the end. Save this edit. Note that if you placed `alda.exe` in a different folder, you will need to use that folder's full path name in your edit, instead.

You will now be able to run Alda from anywhere in the command prompt by typing `alda`, but note that command prompts that were already open will need to be restarted before they will pick up on the new PATH value.

### Updating Alda

Once you have Alda installed, you can update to the latest version at any time by running:

```
alda update
```

### MIDI soundfonts

Default JVM soundfonts usually are of low quality. We recommend installing a good freeware soundfont like FluidR3 to make your MIDI instruments sound a lot nicer.

#### Mac OS X / Linux

For your convenience, there is a script in this repo that will install the FluidR3 soundfont for Mac and Linux users.

To install FluidR3 on your Mac or Linux system, clone this repo and run:

    scripts/install-fluidr3

This will download FluidR3 and replace `~/.gervill/soundbank-emg.sf2` (your JVM's default soundfont) with it.

#### Windows

<img src="doc/windows_jre_soundfont.png" alt="Replacing the JVM soundfont on Windows">

To replace the default soundfont on a Windows OS:

1. Locate your Java Runtime (JRE) folder and navigate into the `lib` folder. 
   * If you have JDK 8 or earlier installed, locate your JDK folder instead and navigate into the `jre\lib` folder. 
2. Make a new folder named `audio`.
3. Copy any `.sf2` file into this folder.

A variety of popular freeware soundfonts, including FluidR3, are available for download [here](https://musescore.org/en/handbook/soundfonts#list).

### Editor Plugins

For the best experience when editing Alda score files, install the Alda file-type plugin for your editor of choice.

> Don't see a plugin for your favorite editor? Write your own and open a pull request to add it here! :)

- [Sublime Text](https://github.com/archimedespi/sublime-alda)
- [Atom](https://github.com/MadcapJake/language-alda)
- [Eclipse](https://github.com/VishwaasHegde/Alda-Eclipse-Plugin)
- [Vim](https://github.com/daveyarwood/vim-alda)
- [Emacs](https://github.com/jgkamat/alda-mode)

## Demo

First start the Alda server (this may take a minute):

    alda up

To play a file containing Alda code:

    alda play --file examples/bach_cello_suite_no_1.alda

To play arbitrary code at the command line:

    alda play --code "piano: c6 d12 e6 g12~4"

To start an [Alda REPL](doc/alda-repl.md):

    alda repl

## Documentation

Alda's documentation can be found [here](doc/index.md).

## Contributing

We'd love your help -- Pull Requests welcome!

The Alda project is composed of a number of subprojects, each of which has its
own GitHub repository within the [alda-lang][gh-org] organization.

For a top-level overview of things we're talking about and working on across all
of the subprojects, check out the [Alda GitHub Project board][gh-project].

[gh-org]: https://github.com/alda-lang
[gh-project]: https://github.com/orgs/alda-lang/projects/1

For more details on how you can contribute to Alda, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Support, Discussion, Comaraderie

**Slack**: Joining the [Alda Slack group](http://slack.alda.io) is quick and painless. Come say hi!

**Reddit**: Subscribe to the [/r/alda](https://www.reddit.com/r/alda/) subreddit, where you can discuss all things Alda and share your Alda scores!

## License

Copyright © 2012-2019 Dave Yarwood et al

Distributed under the Eclipse Public License version 1.0.

# [FIXME] README 2: daveyarwood/osc-spike

# Alda OSC spike

An experiment for [Alda](https://github.com/alda-lang/alda) 2.0.

## Client

One aspect of Alda 2.0 is that most of the work currently done by the Alda 1.x
worker process will now be done by the client, and the client will be written in
a language closer to the metal like Go. This spike is not concerned with
replicating all of the functionality of Alda in Go; rather, the `client/` in
this repo is a simple Go program that sends example OSC messages that would be
generated by parsing Alda scores.

## Player

Alda still needs to have an asynchronous playback process apart from the client;
a spike of this is the `player/` in this repo. Another aspect of Alda 2.0 is
that the asynchronous process ("worker" in Alda 1.x, "player" in Alda 2.0) be
essentially agnostic of the Alda language. Instead, the player is a process that
receives and handles OSC messages that contain lower-level instructions for
audio playback.

A limitation of Alda 1.x is that it does not support live coding. One important
thing needed to support live coding is the ability to loop patterns and modify
them while they're playing. So, the instructions that the player understands
include some that are about defining, playing, and looping patterns.

## OSC API

See [Alda via OSC](doc/alda-via-osc.md).

## Demo

### Prerequisites

* Go is needed in order to run the client. I'm using version 1.11.4.

> Gradle is used to build/run the player, however there is a `gradlew` wrapper
> checked into the repo that makes it so that you don't need to have Gradle
> installed. You can replace `gradle` below with `./gradlew` (run from the
> `player/` directory) and it should work the same as if you had Gradle
> installed.

### Player

```bash
cd player/
gradle run --args 27278
```

Leave the player running, and use the client to send messages.

### Client

The client is a Go program that sends example OSC messages to be handled by the
player. Each example has a short identifier so that you can select it from the
command-line.

Example usage:

```bash
cd client
go run main.go 27278
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
go run main.go 27278 1
go run main.go 27278 1
go run main.go 27278 1
go run main.go 27278 1
```

Queue up a four measures, each containing one randomly chosen note played as
sixteen fast 16th notes:


```bash
# send these in rapid succession
go run main.go 27278 16fast
go run main.go 27278 16fast
go run main.go 27278 16fast
go run main.go 27278 16fast
```

Queue up interspersed single notes, 16th note bars, and instances of the
`simple` pattern:

```bash
# send these in rapid succession
go run main.go 27278 1
go run main.go 27278 16fast
go run main.go 27278 pat1
go run main.go 27278 1
go run main.go 27278 16fast
go run main.go 27278 pat1
```

Queue up the `simple` pattern to be played several times, and change the pattern
while it's playing:

```bash
# send these in rapid succession
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1
go run main.go 27278 pat1

# send while the pattern is playing
go run main.go 27278 patchange

# send while the new pattern is playing
go run main.go 27278 patchange
```

## License

Copyright © 2019 Dave Yarwood

Distributed under the Eclipse Public License version 2.0.
