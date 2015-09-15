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

[![Clojars Project](http://clojars.org/alda/latest-version.svg)](http://clojars.org/alda)

*New to Alda? You may be interested in reading [this blog post][alda-blog-post] as an introduction.*

Inspired by other music/audio programming languages such as [PPMCK][ppmck],
[LilyPond][lilypond] and [ChucK][chuck], Alda aims to be a
powerful and flexible programming language for the musician who wants to easily
compose and generate music on the fly, using naught but a text editor.
Alda is designed in a way that equally favors aesthetics, flexibility and
ease of use, with (eventual) support for the text-based creation of all manner
of music: classical, popular, chiptune, electroacoustic, and more!

## Etymology

Alda was originally named after [Yggdrasil][yggdrasil], the venerated tree of Norse legend which held aloft the mythical nine worlds. I thought it to be a fitting name, imagining the realm of sound/music to be an immense tree bearing numerous branches which could represent genres, tonalities, paradigms, etc.

By incredible coincidence, [the company I work for][adzerk] uses Norse mythology as a theme for naming our software projects, and there will soon be a major one in production called Yggdrasil. To be completely honest, I was never totally happy with Yggdrasil as the name for my music programming language (it's a mouthful), so I took this as an opportunity to rename it. *Alda* is [Quenya][quenya] for "tree."

There is a plethora of music software out there, but most of these
programs tend to specialize or "reside" in at most one or two different realms
-- [FamiTracker][famitracker] and MCK are specifically for the creation of NES
music; [puredata][pd], [Csound][csound] and ChucK are mostly useful for
experimental electronic music; Lilypond, [Rosegarden][rosegarden], and
[MuseScore][musescore] can be used for more than just classical music, but
their standard notation interface suggests a preference for classical music;
[Guitar Pro][guitarpro] is targeted toward the creation of guitar-based music. Why
not have one piece of software that can serve as the Great Tree that supports
all of these existing worlds?

[alda-blog-post]: http://daveyarwood.github.io/alda/2015/09/05/alda-a-manifesto-and-gentle-introduction
[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu
[yggdrasil]: http://en.wikipedia.org/wiki/Yggdrasil
[adzerk]: http://www.adzerk.com
[quenya]: http://en.wikipedia.org/wiki/Quenya
[famitracker]: http://famitracker.com
[pd]: http://puredata.info
[csound]: http://www.csounds.com
[rosegarden]: http://www.rosegardenmusic.com
[musescore]: http://musescore.org
[guitarpro]: http://www.guitar-pro.com

## Syntax example

    piano: o3
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4
    << g1/>g/>g/b/>d/g

## Quick demo

Assuming you have [Boot](http://www.boot-clj.com) installed, try this on for size:

    git clone git@github.com:alda-lang/alda.git
    cd alda
    bin/alda play --file test/examples/awobmolg.alda

> NOTE: The first time you run the `play` task, you may need to wait a minute for the FluidR3 MIDI soundfont dependency (~141 MB) to download. Alda uses this soundfont in order to make your JVM's MIDI instruments sound a lot nicer. If you'd prefer to skip this step and use your JVM's default soundfont instead, include the `--stock` flag (i.e. `play --stock --file ...`).

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

> (TODO: documentation on how this works)

## Log levels

Alda uses [timbre](https://github.com/ptaoussanis/timbre) for logging. Every note event, attribute change, etc. is logged at the DEBUG level, which can be useful for debugging purposes.

The default logging level is WARN, so by default, you will not see these debug-level logs; you will only see warnings and errors.

To override this setting (e.g. for development and debugging), you can set the `TIMBRE_LEVEL` environment variable.

To see debug logs, for example, you can do this:

    export TIMBRE_LEVEL=debug

When running tests via `boot test`, the log level will default to `debug` unless `TIMBRE_LEVEL` is set to something else.

## Contributing

See: [CONTRIBUTING.md](CONTRIBUTING.md)

:clap: :clap: :clap: A big shout-out to our [contributors](https://github.com/alda-lang/alda/graphs/contributors)! :clap: :clap: :clap:

## Support, Discussion, Comaraderie

Sign up to the universe of Clojure chat @ http://clojurians.net/, then join us on #alda

## License

Copyright © 2012-2015 Dave Yarwood et al

Distributed under the Eclipse Public License version 1.0.
