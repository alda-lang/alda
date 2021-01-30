<p align="center">
  <a href="http://alda.io">
    <img src="alda-logo-horizontal.svg"
         alt="alda logo"
         width=360
         height=128>
  </a>

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

> New to Alda? You may be interested in reading [this blog post][alda-blog-post] as an introduction.


Alda is a text-based programming language for music composition. It allows you
to write and play back music using only a text editor and the command line.

```alda
piano:
  o3
  g8 a b > c d e f+ g | a b > c d e f+ g4
  g8 f+ e d c < b a g | f+ e d c < b a g4
  << g1/>g/>g/b/>d/g
```

> For more examples, see these [example scores](./examples/).

The language's design equally favors aesthetics, flexibility and ease of use.

[alda-blog-post]: https://blog.djy.io/alda-a-manifesto-and-gentle-introduction/

## Features

* Easy to understand, markup-like syntax.
* Designed for musicians who don't know how to program, as well as programmers
  who don't know how to music.
* A score is a text file that can be played using the `alda` command-line tool.
* [Interactive REPL](doc/alda-repl.md) lets you enter Alda code and hear the
  results in real time.
* Supports [writing music
  programmatically](doc/writing-music-programmatically.md) (for algorithmic
  composition, live coding, etc.)
* Create MIDI music using any of the instruments in the [General MIDI Sound
  Set][gm-sound-set]

[gm-sound-set]: http://www.midi.org/techspecs/gm1sound.php

### Planned

> If you'd like to help, come on in -- [the water's fine](#contributing)!

* [Define and use waveform synthesis instruments](https://github.com/alda-lang/alda/issues/100)
* [Import MIDI files](https://github.com/alda-lang/alda-core/issues/25)
* [Export to MusicXML](https://github.com/alda-lang/alda-core/issues/3) for inter-operability with other music software
* [A more robust REPL](https://github.com/alda-lang/alda-client-java/issues/2), tailor-made for editing scores interactively

## Installation

> **TODO**: Develop an installation process for Alda 2.0.

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

- [Ace](https://ace.c9.io/)
- [Atom](https://github.com/MadcapJake/language-alda)
- [Eclipse](https://github.com/VishwaasHegde/Alda-Eclipse-Plugin)
- [Emacs](https://github.com/jgkamat/alda-mode)
- [Sublime Text](https://github.com/archimedespi/sublime-alda)
- [Vim](https://github.com/daveyarwood/vim-alda)
- [VS Code](https://github.com/abhi18av/Alda-VsCode)

## Demo

For an overview of available commands and options:

    alda --help

To play a file containing Alda code:

    alda play --file examples/bach_cello_suite_no_1.alda

To play arbitrary code at the command line:

    alda play --code "piano: c6 d12 e6 g12~4"

To start an interactive [Alda REPL](doc/alda-repl.md) session:

    alda repl

## Documentation

Alda's documentation can be found [here](doc/index.md).

## Contributing

We'd love your help -- Pull Requests welcome!

For a top-level overview of things we're talking about and working on, check out
the [Alda GitHub Project board][gh-project].

For more details on how you can contribute to Alda, see [CONTRIBUTING.md](CONTRIBUTING.md).

> Another way you can contribute is by [sponsoring Dave][gh-sponsor] in the
> future development of Alda.

[gh-org]: https://github.com/alda-lang
[gh-project]: https://github.com/orgs/alda-lang/projects/1
[gh-sponsor]: https://github.com/sponsors/daveyarwood

## Support, Discussion, Comaraderie

**Slack**: Joining the [Alda Slack group](https://slack.alda.io) is quick and
painless. Come say hi!

**Reddit**: Subscribe to the [/r/alda](https://www.reddit.com/r/alda/)
subreddit, where you can discuss all things Alda and share your Alda scores!

## License

Copyright © 2012-2021 Dave Yarwood et al

Distributed under the Eclipse Public License version 2.0.
