```
                                       ________________________________
                                      /    o   oooo ooo oooo   o o o  /\
                                     /    oo  ooo  oo  oooo   o o o  / /
                                    /    _________________________  / /
                                   / // / // /// // /// // /// / / / /
                                  /___ //////////////////////////_/ /
                                  \____\________________________\_\/

                                               ~ alda ~
```

<h2 align=center>a music programming language for musicians</h2>

<p align="center">

<a href="http://clojars.org/alda">
  <img src="http://clojars.org/alda/latest-version.svg" alt="Clojars Project">
</a>
<br>


<b><a href="#installation">Installation</a></b>
|
<b><a href="doc/index.md">Docs</a></b>
|
<b><a href="CHANGELOG.md">Changelog</a></b>
|
<b><a href="#contributing">Contributing</a></b>

(
<a href="https://waffle.io/alda-lang/alda" alt="Features in Progress">
  <img src="https://badge.waffle.io/alda-lang/alda.png?label=in%20progress&title=In%20Progress:">
</a>
)

<br>

</p>

*New to Alda? You may be interested in reading [this blog post][alda-blog-post] as an introduction.*

Inspired by other music/audio programming languages such as [PPMCK][ppmck],
[LilyPond][lilypond] and [ChucK][chuck], Alda aims to be a
powerful and flexible programming language for the musician who wants to easily
compose and generate music on the fly, using naught but a text editor.
Alda is designed in a way that equally favors aesthetics, flexibility and
ease of use, with (eventual) support for the text-based creation of all manner
of music: classical, popular, chiptune, electroacoustic, and more!

[alda-blog-post]: http://daveyarwood.github.io/alda/2015/09/05/alda-a-manifesto-and-gentle-introduction
[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu

## Features

* Easy to understand, markup-like syntax
* Perfect for musicians who don't know how to program and programmers who don't know how to music
* Represent scores as text files and play them back with the `alda` command-line tool
* [Interactive REPL](doc/alda-repl.md) lets you type Alda code and hear the results in real time
* [Underlying Clojure DSL](doc/alda-lisp.md) allows you to [use Alda directly in your Clojure project](doc/alda-now.md).
* [Inline Clojure code](doc/inline-clojure-code.md) allows you to [hack the Gibson][hackers] and write scores programmatically
* Create MIDI music using any of the instruments in the [General MIDI Sound Set][gm-sound-set]

[hackers]: https://www.youtube.com/watch?v=vYNnPx8fZBs
[gm-sound-set]: http://www.midi.org/techspecs/gm1sound.php

### TODO

* [Define and use waveform synthesis instruments](https://github.com/alda-lang/alda/issues/100)
* [Import MIDI files](https://github.com/alda-lang/alda/issues/85)
* [Export to MusicXML](https://github.com/alda-lang/alda/issues/44) for inter-operability with other music software
* [A more robust REPL](https://github.com/alda-lang/alda/issues/54), tailor-made for editing scores interactively
* [A plugin system](https://github.com/alda-lang/alda/issues/37) allowing users to define custom/unofficial syntax in Alda scores

If you're a developer and you'd like to help, come on in -- [the water's fine](#contributing)!

## Syntax example

    piano: o3
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4
    << g1/>g/>g/b/>d/g

For more examples, see these [example scores](https://github.com/alda-lang/alda/tree/master/examples).

## Installation

> You must have [Java](https://www.java.com/en/download) 7+ installed on your system in order to run Alda.
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
  *  Go to the Windows `System` control panel option, select `Advanced System Settings` and then click on `Environment Variables`, then edit the `PATH` variable (either specifically for your user account or for the system in general) and add `;C:\Program Files (x86)\Alda` to the end. Save this edit. Note that if you placed `alda.exe` in a different folder, you will need to use that folder's full path name in your edit, instead.

You will now be able to run Alda from anywhere in the command prompt by typing `alda`, but note that command prompts that were already open will need to be restarted before they will pick up on the new PATH value.

### Updating Alda

Once you have Alda installed, you can update to the latest version at any time by running:

```
alda update
```

### MIDI soundfonts

Default JVM soundfonts usually are of low quality. We recommend installing a good freeware soundfont like FluidR3 to make your MIDI instruments sound a lot nicer. For your convenience, there is a script in this repo that will install the FluidR3 soundfont for Mac and Linux users.

> If you're a Windows user and you know how to install a MIDI soundfont to the Java Virtual Machine, please let us know!

To install FluidR3 on your Mac or Linux system, clone this repo and run:

    scripts/install-fluidr3

This will download FluidR3 and replace `~/.gervill/soundbank-emg.sf2` (your JVM's default soundfont) with it.

### Editor Plugins

For the best experience when editing Alda score files, install the Alda file-type plugin for your editor of choice.

> Don't see a plugin for your favorite editor? Write your own and open a Pull Request to add it here! :)

- [Sublime Text](https://github.com/archimedespi/sublime-alda)
- [Atom](https://github.com/MadcapJake/language-alda)
- [Vim](https://github.com/daveyarwood/vim-alda)
- [Emacs](https://github.com/jgkamat/alda-mode)

## Demo

To play a file:

    alda play --file examples/bach_cello_suite_no_1.alda

To play arbitrary code:

    alda play --code "piano: c6 d12 e6 g12~4"

To start an [Alda REPL](doc/alda-repl.md):

    alda repl

## Documentation

Alda's documentation can be found [here](doc/index.md).

## Contributing

PRs welcome! See: [CONTRIBUTING.md](CONTRIBUTING.md)

:clap: :clap: :clap: A big shout-out to our [contributors](https://github.com/alda-lang/alda/graphs/contributors)! :clap: :clap: :clap:

## Support, Discussion, Comaraderie

**Slack**: Sign up to the universe of Clojure chat @ http://clojurians.net/, then join us on #alda

**Reddit**: Come join us in [/r/alda](https://www.reddit.com/r/alda/), where you can discuss all things Alda and share your Alda scores!

## License

Copyright © 2012-2016 Dave Yarwood et al

Distributed under the Eclipse Public License version 1.0.
