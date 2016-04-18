# Development Guide

## What You'll Need

* The Alda project uses [Boot](http://boot-clj.com) to build releases from the Java/Clojure source, as well as to perform useful development tasks like running tests. You will need Boot in order to test any changes you make to the source code.

  > The `boot` commands described in this guide need to be run while in the root directory of this project, which contains the project file `build.boot`.

  To test changes to the Alda server, client, or REPL, there is a convenient `boot dev` task that requires no additional setup beyond installing Boot. For more information on `boot dev`, see the sections below on the Alda client, server, and REPL.

* (Optional) The `boot build` task (used to build the `alda`/`alda.exe` executables) requires two additional things:

  * [Launch4j](http://launch4j.sourceforge.net) is needed in order to build the Windows executable.

    If you're on a Mac with [Homebrew](http://brew.sh) installed, you can install Launch4j by running:

        brew install launch4j

  * [Java Development Kit 7](http://www.oracle.com/technetwork/java/javase/downloads/jdk7-downloads-1880260.html) is needed in order to create executables that will run on systems that have Java 7+.

    After installing JDK7, set the environment variable `JDK7_BOOTCLASSPATH` to the path to your JDK7 classpath jar. This location will vary depending on your OS. On my Mac, the path is `/Library/Java/JavaVirtualMachines/jdk1.7.0_71.jdk/Contents/Home/jre/lib/rt.jar`.

## Testing changes

### Manual testing

The Alda client, server, and REPL can each be run in development (including any changes you've made to your copy of the code) via the convenient `boot dev` task. See the sections below on the client, server, and REPL for details.

### Unit tests

You should run `boot test` prior to submitting any Pull Request involving changes to the server (Clojure) code. This will run automated tests that live in the `server/test` directory.

#### Adding tests

It is generally good to add to the existing tests wherever it makes sense, i.e. whenever there is a new test case that Alda needs to consider. [Test-driven development](https://en.wikipedia.org/wiki/Test-driven_development) is a good idea.

If you find yourself adding a new file to the tests, be sure to add its namespace to the `test` task option in `build.boot` so that it will be included when you run the tests via `boot test`.

The automated test battery includes parsing and evaluating all of the example Alda scores in the `examples/` directory. If you add an additional example score, be sure to add it to the list of score files in `server/test/alda/examples_test.clj`.

## Building the Project

The Alda client, server and REPL are packaged together in the same uberjar. You can build the project by running:

    boot build -o /path/to/output-dir/

This will build the `alda` and `alda.exe` executables and place them in the output directory of your choice.

## Client

### Overview

The Alda client is a fairly straightforward Java CLI app that uses [JCommander](http://jcommander.org) to parse command-line arguments.

Interaction with servers is done via simple HTTP requests. Unless specified via the `-H/--host` option, the Alda client assumes the servers are run locally and sends requests to localhost. The default port is 27713.

Running `alda start` forks a new Alda process in the background, passing it the (hidden) `server` command to start the server. Server output is hidden from the user (though the client will report errors).

To see server output for development purposes, you can start a server in the foreground by running `alda server`. You may specify a port via the `-p/--port` option (e.g. `alda -p 2000 server`) -- just make sure you're sending requests from the client to the right port (e.g. `alda -p 2000 play -f my-score.alda`).

To stop a running server, run `alda stop`.

### Development

To run the Alda client locally to test changes you've made to the code, you can run:

    boot dev -a client -x "args here"

For example, to test changes to the way the `alda play` command plays a file, you can run:

    boot dev -a client -x "play --file /path/to/file.alda"

If you'd prefer, you can do a development build, outputting the executables to a directory of your choice (I like to use `/tmp`), and use the outputted executable:

    boot build -o /tmp
    /tmp/alda play --file /path/to/file.alda

Note that the client forks server processes into the background, which continue to run even after the client has exited. If you're testing out changes you've made to the server code, you will need to restart the server by running `alda restart`, which will stop the server and start a new process using your new build.

## Server

### Overview

The Alda server is written in Clojure. It handles a variety of things, including parsing Alda code into executable Clojure code, executing the code to create or modify a score, modeling a musical score as a Clojure map of data, and using the map of data as input to interpret the score (generating sound).

### Development

#### Alda Server

To run an Alda server with any local changes you've made:

    boot dev -a server --port 27713 --alda-fingerprint

The `--port` and `--alda-fingerprint` arguments are strictly optional, but including them will ensure that the Alda client recognizes your development server as an Alda server and includes it in the output of `alda list`.

There is a middleware built into the development server that reloads all of the server namespaces before each request. This makes it so that the server does not need to be restarted each time you make a change to the code.

#### Alda REPL

To run an Alda REPL with any local changes you've made:

    boot dev -a repl

### Namespaces

* [alda.parser](#aldaparser)
* [alda.lisp](#aldalisp)
* [alda.sound](#aldasound)
* [alda.now](alda-now.md)
* [alda.repl](#aldarepl)
* [alda.server](#aldaserver)

#### alda.parser

Parsing begins with the `parse-input` function in the [`alda.parser`](https://github.com/alda-lang/alda/blob/master/server/src/alda/parser.clj) namespace. This function uses a series of parsers built using [Instaparse](https://github.com/Engelberg/instaparse), an excellent parser-generator library for Clojure.
The grammars for each step of the parsing process are composed from [small files written in BNF](https://github.com/alda-lang/alda/blob/master/server/grammar) (with some Instaparse-specific sugar); if you find yourself editing any of these files, it may be helpful to read up on Instaparse. [The tutorial in the Instaparse README](https://github.com/Engelberg/instaparse) is comprehensive and excellent.

The parsers transform Alda code into an intermediate AST, which looks something like this:

```clojure
[:score
  [:part
    [:calls [:name "piano"]]
    [:note
      [:pitch "c"]
      [:duration
        [:note-length [:positive-number "8"]]]]
    [:note
      [:pitch "e"]]
    [:note
      [:pitch "g"]]
    [:chord
      [:note
        [:pitch "c"]
        [:duration [:note-length [:positive-number "1"]]]]
      [:note
        [:pitch "f"]]
      [:note
        [:pitch "a"]]]]]
```

These parse trees are then [transformed](https://github.com/Engelberg/instaparse#transforming-the-tree) into Clojure code which, when run, will produce a data representation of a musical score.

Clojure is a Lisp; in Lisp, code is data and data is code. This powerful concept allows us to represent a morsel of code as a list of elements. The first element in the list is a function, and every subsequent element is an argument to that function. These code morsels can even be nested, just like our parse tree. Alda's parser's transformation phase translates each type of node in the parse tree into a Clojure expression that can be evaluated with the help of the `alda.lisp` namespace.

```clojure
alda.parser=> (parse-input "piano: c8 e g c1/f/a")

(alda.lisp/score
  (alda.lisp/part {:names ["piano"]}
    (alda.lisp/note
      (alda.lisp/pitch :c)
      (alda.lisp/duration (alda.lisp/note-length 8)))
    (alda.lisp/note
      (alda.lisp/pitch :e))
    (alda.lisp/note
      (alda.lisp/pitch :g))
    (alda.lisp/chord
      (alda.lisp/note
        (alda.lisp/pitch :c)
        (alda.lisp/duration (alda.lisp/note-length 1)))
      (alda.lisp/note
        (alda.lisp/pitch :f))
      (alda.lisp/note
        (alda.lisp/pitch :a)))))
```

#### alda.lisp

When you evaluate a score [S-expression](https://en.wikipedia.org/wiki/S-expression) like the one above, the result is a map of score information, which provides all of the data that Alda's audio component needs to make an audible version of your score.

```clojure
{:events
 #{{:offset 750.0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 69,
    :pitch 440.0,
    :duration 1800.0}
   {:offset 500.0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 67,
    :pitch 391.99543598174927,
    :duration 225.0}
   {:offset 750.0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 65,
    :pitch 349.2282314330039,
    :duration 1800.0}
   {:offset 0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 60,
    :pitch 261.6255653005986,
    :duration 225.0}
   {:offset 750.0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 60,
    :pitch 261.6255653005986,
    :duration 1800.0}
   {:offset 250.0,
    :instrument "piano-foyYJ",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :midi-note 64,
    :pitch 329.6275569128699,
    :duration 225.0}},
 :markers {:start 0},
 :instruments
 {"piano-foyYJ"
  {:octave 4,
   :current-offset {:offset 2750.0},
   :config {:type :midi, :patch 1},
   :duration 4,
   :volume 1.0,
   :last-offset {:offset 750.0},
   :id "piano-foyYJ",
   :quantization 0.9,
   :tempo 120,
   :panning 0.5,
   :current-marker :start,
   :stock "midi-acoustic-grand-piano",
   :track-volume 0.7874015748031497}}}
```

There are 3 keys in this map:

* **:events** -- a set of note events
* **:markers** -- a map of marker names to offsets, expressed as milliseconds from the beginning of the score (`:start` is a special marker that is always placed at offset 0)
* **:instruments** -- a map of randomly-generated ids to all of the information that Alda has about an instrument, *at the point where the score ends*.

A note event contains information such as the pitch, MIDI note and duration of a note, which instrument instance is playing the note, and what its offset is relative to the beginning of the score (i.e., where the note is in the score)

Because `alda.lisp` is a Clojure DSL, it's possible to use it to build scores within a Clojure program, as an alternative to using Alda syntax:

```clojure
(ns my-clj-project.core
  (:require [alda.lisp :refer :all]))

(score
  (part "piano"
    (note (pitch :c) (duration (note-length 8)))
    (note (pitch :d))
    (note (pitch :e))
    (note (pitch :f))
    (note (pitch :g))
    (note (pitch :a))
    (note (pitch :b))
    (octave :up)
    (note (pitch :c))))
```

#### alda.sound

The `alda.sound` namespace handles the implementation details of playing the score.

There is an "audio type" abstraction which refers to different ways to generate audio, e.g. MIDI, waveform synthesis, samples, etc. Adding a new audio type is as simple as providing an implementation for each of the multimethods in this namespace, i.e. `set-up-audio-type!`, `refresh-audio-type!`, `tear-down-audio-type!`, `start-event!` and `stop-event!`.

The `play!` function handles playing an entire Alda score. It does this by using a [JSyn](http://www.softsynth.com/jsyn) SynthesisEngine to schedule all of the note events to be played in realtime. The time that each event starts and stops is determined by its `:offset` and `:duration`.

What happens, exactly, at the beginning and end of an event, is determined by the `start-event!`/`stop-event!` implementations for each instrument type. For example, for MIDI instruments, `start-event!` sets parameters such as volume and panning and sends a MIDI note-on message at the beginning of the score plus `:offset` milliseconds, and `stop-event!` sends a note-off message `:duration` milliseconds later.

##### alda.lisp.instruments

Although technically a part of `alda.lisp`, stock instrument configurations are defined here, which are included in an Alda score map, and then used by `alda.sound` to provide details about how to play an instrument's note events. Each instrument's `:config` field is available to the `alda.sound/play-event!` function via the `:instrument` field of the event.

#### alda.repl

`alda.repl` is the codebase for Alda's **R**ead-**E**val-**P**lay **L**oop, which lets you build a score interactively by entering Alda code one line at a time.

There are built-in commands defined in `alda.repl.commands` that are defined using the `defcommand` macro. Defining a command here makes it available from the Alda REPL prompt by typing a colon before the name of the command, i.e. `:score`.

The core logic for what goes on behind the curtain when you use the REPL lives in `alda.repl.core`. A good practice for implementing a REPL command in `alda.repl.commands` is to move implementation details into `alda.repl.core` (or perhaps into a new sub-namespace of `alda.repl`, if appropriate) if the body of the command definition starts to get too long.

#### alda.server

`alda.server/start-server!` is the entrypoint to the Alda server. It starts a Clojure web app using [Ring](https://github.com/ring-clojure/ring) and [Compojure](https://github.com/weavejester/compojure), and serves it via [Jetty](https://en.wikipedia.org/wiki/Jetty_(web_server)).

Requests can be made to the server via any HTTP client, e.g. curl, or via the Alda client, which communicates with the server by making HTTP requests.

