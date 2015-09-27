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
* [Interactive REPL](#aldarepl) lets you type Alda code and hear the results in real time
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

## Quick demo

Assuming you have [Boot](http://www.boot-clj.com) installed, try this on for size:

    git clone git@github.com:alda-lang/alda.git
    cd alda
    bin/alda play --file test/examples/awobmolg.alda

> NOTE: Default JVM soundfonts usually are of low quality. You can install FluidR3 soundfont by executing `bin/install-fluidr3`. You may need to wait a minute for the FluidR3 MIDI soundfont dependency (~141 MB) to download. Alda uses this soundfont in order to make your JVM's MIDI instruments sound a lot nicer.

You can also execute arbitrary Alda code, like this:

    bin/alda play --code "piano: c6 d12 e6 g12~4"

## Installation

### Mac OS X / Linux

The executable file `alda` in the `bin` directory of this repository is a standalone executable script that can be run from anywhere. It will retrieve the latest release version of Alda and run it, passing along any command-line arguments you give it.

To install Alda, simply copy the `alda` script from this repo into any directory in your `$PATH`, e.g. `/bin` or `/usr/local/bin`:

    curl https://raw.githubusercontent.com/alda-lang/alda/master/bin/alda -o /usr/local/bin/alda && chmod +x /usr/local/bin/alda

This script requires the Clojure build tool [Boot](http://www.boot-clj.com), so you will need to have that installed as well. Mac OS X users with [Homebrew](https://github.com/homebrew/homebrew) can run `brew install boot-clj` to install Boot. Otherwise, [see here](https://github.com/boot-clj/boot#install) for more details about installing Boot.

Once you've completed the steps above, you'll be able to run `alda` from any working directory. Running the command `alda` by itself will display the help text.

### Windows

The `alda` script doesn't currently work for Windows users. If you're a Windows power user, [please feel free to weigh in on this issue](https://github.com/alda-lang/alda/issues/48). Until we have that sorted out, there is a workaround:

1. Clone this repo and `cd` into it.
2. You can now run `boot alda -x <cmd> <args>` while you are in this directory.

Examples:

* `boot alda -x repl` to start the Alda REPL
* `boot alda -x "play --code 'piano: c d e f g'"`

Caveats:

* It's more typing.
* It only works if you're in the Alda repo folder.
* Unlike the `alda` script, running the `boot alda` task will not automatically update Alda; you will have to do so manually by running `git pull`.
* If the command you're running is longer than one word, you must wrap it in double quotes -- see the examples above.

## alda.repl

Alda comes with an interactive REPL (**R**ead-**E**val-**P**lay **L**oop) that you can use to play around with its syntax. After each line of code that you enter into the REPL prompt, you will hear the result.

To start the Alda REPL, run:

    alda repl

## alda.lisp

Under the hood, Alda transforms input (i.e. Alda code) into Clojure code which, when evaluated, produces a map of score information, which the audio component of Alda can then use to make sound. This Clojure code is written in a DSL called **alda.lisp**. See below for an example of alda.lisp code and the result of evaluating it.

### Parsing demo

You can use the `parse` task to parse Alda code into alda.lisp (`-l`/`--lisp`) and/or evaluate it to produce a map (`-m`/`--map`) of score information.

    $ alda parse --lisp --map -f test/examples/hello_world.alda
    (alda.lisp/score
      (alda.lisp/part {:names ["piano"]}
        (alda.lisp/note (alda.lisp/pitch :c)
                        (alda.lisp/duration (alda.lisp/note-length 8)))
        (alda.lisp/note (alda.lisp/pitch :d))
        (alda.lisp/note (alda.lisp/pitch :e))
        (alda.lisp/note (alda.lisp/pitch :f))
        (alda.lisp/note (alda.lisp/pitch :g))
        (alda.lisp/note (alda.lisp/pitch :f))
        (alda.lisp/note (alda.lisp/pitch :e))
        (alda.lisp/note (alda.lisp/pitch :d))
        (alda.lisp/note (alda.lisp/pitch :c)
                        (alda.lisp/duration (alda.lisp/note-length 2 {:dots 1})))))

    {:events #{#alda.lisp.Note{:offset 2000.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 261.6255653005986, :duration 1350.0} #alda.lisp.Note{:offset 0, :instrument "piano-VoUlp", :volume 1.0, :pitch 261.6255653005986, :duration 225.0} #alda.lisp.Note{:offset 250.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 293.6647679174076, :duration 225.0} #alda.lisp.Note{:offset 1250.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 349.2282314330039, :duration 225.0} #alda.lisp.Note{:offset 750.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 349.2282314330039, :duration 225.0} #alda.lisp.Note{:offset 1000.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 391.99543598174927, :duration 225.0} #alda.lisp.Note{:offset 1750.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 293.6647679174076, :duration 225.0} #alda.lisp.Note{:offset 1500.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 329.6275569128699, :duration 225.0} #alda.lisp.Note{:offset 500.0, :instrument "piano-VoUlp", :volume 1.0, :pitch 329.6275569128699, :duration 225.0}},
     :instruments {"piano-VoUlp" {:octave 4, :current-offset #alda.lisp.AbsoluteOffset{:offset 3500.0}, :config {:type :midi}, :duration 3N, :volume 1.0, :last-offset #alda.lisp.AbsoluteOffset{:offset 2000.0}, :id "piano-VoUlp", :quantization 0.9, :tempo 120, :panning 0.5, :current-marker :start, :stock "piano"}}}

    $ alda parse --lisp -c 'cello: c+'
    (alda.lisp/score
      (alda.lisp/part {:names ["cello"]}
        (alda.lisp/note (alda.lisp/pitch :c :sharp))))

## alda.now

Alda can also be used as a Clojure library.

For more details on how this works, see our [Developer Guide](CONTRIBUTING.md#aldanow).

## Log levels

Alda uses [timbre](https://github.com/ptaoussanis/timbre) for logging. Every note event, attribute change, etc. is logged at the DEBUG level, which can be useful for debugging purposes.

The default logging level is WARN, so by default, you will not see these debug-level logs; you will only see warnings and errors.

To override this setting (e.g. for development and debugging), you can set the `TIMBRE_LEVEL` environment variable.

To see debug logs, for example, you can do this:

    export TIMBRE_LEVEL=debug

When running tests via `boot test`, the log level will default to `debug` unless `TIMBRE_LEVEL` is set to something else.

## Documentation

See: [doc/index.md](doc/index.md)

## Contributing

PRs welcome! See: [CONTRIBUTING.md](CONTRIBUTING.md)

:clap: :clap: :clap: A big shout-out to our [contributors](https://github.com/alda-lang/alda/graphs/contributors)! :clap: :clap: :clap:

## Support, Discussion, Comaraderie

Sign up to the universe of Clojure chat @ http://clojurians.net/, then join us on #alda

## License

Copyright © 2012-2015 Dave Yarwood et al

Distributed under the Eclipse Public License version 1.0.
