# Contributing to Alda

Pull requests are warmly welcomed. Please feel free to take on whatever [issue](https://github.com/alda-lang/alda/issues) interests you. 

## Instructions

- Fork this repository and make changes on your fork.
- Submit a Pull Request.
- Your Pull Request should get the Dave Yarwood Seal of Approval™ before being merged. (Don't worry, he's not hard to win over.)

If you're confused about how some aspect of the code works (Clojure questions, "what does this piece of code do," etc.), don't hesitate to ask questions on the issue you're working on -- we'll be more than happy to help.

## Development Guide

* [alda.parser](#aldaparser)
* [alda.lisp](#aldalisp)
* [alda.sound](#aldasound)
* [alda.now](doc/alda-now.md)
* [alda.repl](#aldarepl)

Alda is a program that takes a string of code written in Alda syntax, parses it into executable Clojure code that will create a score, and then plays the score.

### alda.parser

Parsing begins with the `parse-input` function in the [`alda.parser`](https://github.com/alda-lang/alda/blob/master/src/alda/parser.clj) namespace. This function uses a parser built using [Instaparse](https://github.com/Engelberg/instaparse), an excellent parser-generator library for Clojure. The grammar for Alda is [a single file written in BNF](https://github.com/alda-lang/alda/blob/master/grammar/alda.bnf) (with some Instaparse-specific sugar); if you find
yourself editing this file, it may be helpful to read up on Instaparse. [The tutorial in the README](https://github.com/Engelberg/instaparse) is comprehensive and excellent.

Code is given to the parser, resulting in a parse tree:

```clojure
alda.parser=> (alda-parser "piano: c8 e g c1/f/a")

[:score 
  [:part 
    [:calls [:name "piano"]] 
    [:note 
      [:pitch "c"] 
      [:duration 
        [:note-length [:number "8"]]]] 
    [:note 
      [:pitch "e"]] 
    [:note 
      [:pitch "g"]] 
    [:chord 
      [:note 
        [:pitch "c"] 
        [:duration [:note-length [:number "1"]]]] 
      [:note 
        [:pitch "f"]] 
      [:note 
        [:pitch "a"]]]]]
```

The parse tree is then [transformed](https://github.com/Engelberg/instaparse#transforming-the-tree) into Clojure code which, when run, will produce a data representation of a musical score.

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

### alda.lisp

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

* **events** -- a set of note events
* **markers** -- a map of marker names to offsets, expressed as milliseconds from the beginning of the score (`:start` is a special marker that is always placed at offset 0)
* **instruments** -- a map of randomly-generated ids to all of the information that Alda has about an instrument, *at the point where the score ends*.

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

### alda.sound

The `alda.sound` namespace handles the implementation details of playing the score. 

There is an "audio type" abstraction which refers to different ways to generate audio, e.g. MIDI, waveform synthesis, samples, etc. Adding a new audio type is as simple as providing an implementation for each of the multimethods in this namespace, i.e. `set-up-audio-type!`, `refresh-audio-type!`, `tear-down-audio-type!` and `play-event!`.

The `play!` function handles playing an entire Alda score. It does this by using [overtone.at-at](https://github.com/overtone/at-at) to schedule all of the note events to be played via `play-event!`, based on the `:offset` of each event.

#### alda.lisp.instruments

Although technically a part of `alda.lisp`, stock instrument configurations are defined here, which are included in an Alda score map, and then used by `alda.sound` to provide details about how to play an instrument's note events. Each instrument's `:config` field is available to the `alda.sound/play-event!` function via the `:instrument` field of the event.

### alda.repl

`alda.repl` is the codebase for Alda's **R**ead-**E**val-**P**lay **L**oop, which lets you build a score interactively by entering Alda code one line at a time.

There are built-in commands defined in `alda.repl.commands` that are defined using the `defcommand` macro. Defining a command here makes it available from the Alda REPL prompt by typing a colon before the name of the command, i.e. `:score`.

The core logic for what goes on behind the curtain when you use the REPL lives in `alda.repl.core`. A good practice for implementing a REPL command in `alda.repl.commands` is to move implementation details into `alda.repl.core` (or perhaps into a new sub-namespace of `alda.repl`, if appropriate) if the body of the command definition starts to get too long.

## Testing changes

There are a couple of [Boot](http://boot-clj.com) tasks provided to help test changes.

### `boot test`

You should run `boot test` prior to submitting a Pull Request. This will run automated tests that live in the `test` directory.

#### Adding tests

It is a good idea in general to add to the existing tests wherever it makes sense, i.e. if there is a new test case that Alda needs to consider. [Test-driven development](https://en.wikipedia.org/wiki/Test-driven_development) is a good idea.

If you find yourself adding a new file to the tests, be sure to add its namespace to the `test` task option in `build.boot` so that it will be included when you run the tests via `boot test`.

### `boot alda`

When you run the `alda` executable, it uses the most recent *released* version of Alda. So, if you make any changes locally, they will not be included when you run `alda repl`, `alda play`, etc.

For testing local changes, you can use the `boot alda` task, which uses the current state of the repository, including any local changes you have made.

#### Example usage

    boot alda -x repl

    boot alda -x "play --code 'piano: c d e f g'"
