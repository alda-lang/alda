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

## A music programming language for musicians

<p align="center">

<a href="http://clojars.org/alda">
  <img src="http://clojars.org/alda/latest-version.svg" alt="Clojars Project">
</a>
<br>

<b><a href="#installation">Installation</a></b>
|
<b><a href="doc/index.md">Docs</a></b>
|
<b><a href="#contributing">Contributing</a></b>
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
* Inline Clojure code allows you to [hack the Gibson][hackers] and write scores programmatically
* Create MIDI music using any of the instruments in the [General MIDI Sound Set][gm-sound-set]

[hackers]: https://www.youtube.com/watch?v=vYNnPx8fZBs
[gm-sound-set]: http://www.midi.org/techspecs/gm1sound.php

### TODO

* [Define and use waveform synthesis instruments](https://github.com/alda-lang/alda/issues/100) 
* [Import MIDI files](https://github.com/alda-lang/alda/issues/85)
* [Export to MusicXML](https://github.com/alda-lang/alda/issues/44) for inter-operability with other music software
* [A more robust REPL](https://github.com/alda-lang/alda/issues/54), tailor-made for editing scores interactively
* [A plugin system](https://github.com/alda-lang/alda/issues/37) allowing users to define custom/unofficial syntax in Alda scores
* [An "alda daemon"](https://github.com/alda-lang/alda/issues/49) with server/client semantics

If you're a developer and you'd like to help, come on in -- [the water's fine](#contributing)!

## Syntax example

    piano: o3
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4
    << g1/>g/>g/b/>d/g

For more examples, see these [example scores](https://github.com/alda-lang/alda/tree/master/test/examples).

## Quick Start

### Installation

> More information can be found in [the docs](doc/installation.md).

#### Mac OS X / Linux

1. Install [Boot](http://www.boot-clj.com).
2. Run this command to place the `alda` script in your `$PATH`:

        curl https://raw.githubusercontent.com/alda-lang/alda/master/bin/alda -o /usr/local/bin/alda && chmod +x /usr/local/bin/alda

#### Windows

See [the docs](doc/installation.md#windows).

### Demo

> NOTE: The first time you run one of these tasks, you may need to wait a minute for the FluidR3 MIDI soundfont dependency (~141 MB) to download. Alda uses this soundfont in order to make your JVM's MIDI instruments sound a lot nicer. If you'd prefer to skip this step and use your JVM's default soundfont instead, include the `--stock` flag (i.e. `play --stock --file ...`).

To play a file:

    alda play --file test/examples/bach_cello_suite_no_1.alda

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

Sign up to the universe of Clojure chat @ http://clojurians.net/, then join us on #alda

## License

Copyright © 2012-2015 Dave Yarwood et al

Distributed under the Eclipse Public License version 1.0.
