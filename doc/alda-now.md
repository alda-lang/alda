# alda.now

`alda.now`, when coupled with [`alda.lisp`](alda-lisp.md), provides a way to work with Alda scores and play music programmatically within a Clojure application.

Of note, [`alda.repl`](alda-repl.md) and [`alda.server`](alda-server.md) both use `alda.now` as a means for managing and playing the user's scores. For advanced `alda.now` usage, you may find it useful to refer to the code in the `alda.repl` and `alda.server` namespaces, as they are essentially `alda.now` sample projects.

## tl;dr

```clojure
(require '[alda.lisp :refer :all])
(require '[alda.now  :as    now])

(now/play!
  (part "accordion"
    (note (pitch :c) (duration (note-length 8)))
    (note (pitch :d))
    (note (pitch :e :flat))
    (note (pitch :f))
    (note (pitch :g))
    (note (pitch :a :flat))
    (note (pitch :b))
    (octave :up)
    (note (pitch :c))))
```

## MIDI soundfonts

The default JVM soundfont sounds pretty bad. If you're using Alda in a Clojure application and you want to have nice MIDI sounds, you can use [`midi.soundfont`](https://github.com/daveyarwood/midi.soundfont) to load FluidR3 or another MIDI soundfont into the JVM at runtime.

## Audio Context

Different instruments available in Alda scores belong to different categories based on their *audio type*. Each audio type carries different semantics when it comes to how the audio engine plays notes. MIDI instruments, for example, play notes using the Java Virtual Machine's built-in MIDI synthesizer.

### `set-up!`

Each audio type has its own required setup steps before it is ready to play notes. The MIDI system, for example, has to acquire and initialize a Java MIDI Synthesizer.

When playing a score, Alda will first make sure that all necessary audio types are set up. If this wasn't done in advance, there may be a delay before playback begins.

To avoid this, you may want to pre-initialize the audio types that you know you will want to use. For example, you can initialize the MIDI system by running `(set-up! your-score :midi)`, where `your-score` is a reference to your score (see "Scores" below). When using `with-score` or `with-new-score`, `alda.now/*current-score*` can be used as a reference to the current score.

## Scores

`alda.now` allows you to maintain multiple Alda scores separately in the same Clojure process. In the context of `alda.lisp`, a "score" is a Clojure map containing a number of data points describing the current state of a musical score. For the purposes of `alda.now`, a "score" is an atom referencing an enhanced `alda.lisp` score map that also contains audio context information. As the score develops over time (instrument parts are added, notes are played, etc.), the atom is updated to reflect the current state of the score.

By default, calling `alda.now/play!` without any sort of context about what score you're playing (like in the "tl;dr" example above) will create and use an anonymous score, discarding it once playback is done. This may be all you need, depending on what you'd like to do with `alda.now`. On the other hand, you may want to gradually add on to the *same* score, across multiple calls to `play!`.

### `new-score`

A new score may be created with `alda.now/new-score`. You can assign the score to a symbol via `def`, `let`, etc.

```clojure
(require '[alda.now :as now])

(def my-score (now/new-score))
```

Then you can execute Alda events (parts, notes, chords, etc.) within the context of your score. Each time you call `play!` with some Alda events as arguments, you will hear the result of executing the events, and the events will be added to your score. Just like when writing a score in Alda syntax, stateful events like attribute changes (including octave changes) will remain in effect across separate score snippets, when evaluated in the context of the same score.

### `with-score`

To tell `alda.now` which score you're using, use `alda.now/with-score`:

```clojure
(require '[alda.now :refer :all])

(def my-score (new-score))

(with-score my-score
  (play!
    (part "saw-wave"
      (note (pitch :c))
      (octave :up))))

(Thread/sleep 2000)

(with-score my-score
  (play!
    (note (pitch :c))))
```

> Playback is handled asynchronously; the execution of your program will not block while your score is playing.

The body of the `with-score` macro may contain any Clojure forms, not just calls to `play!`. You might prefer to use `with-score` only once, and include all of your program's code inside of its scope until you're completely done with your score:

```clojure
(require '[alda.now :refer :all])

(def my-score (new-score))

(with-score my-score
  (println "Playing C4")
  (play!
    (part "saw-wave"
      (note (pitch :c))
      (octave :up)))

  (Thread/sleep 2000)

  (println "Playing C5")
  (play!
    (note (pitch :c))))
```

### `with-new-score`

`with-new-score` is a variation of `with-score` that creates a new score and uses it as the context for all calls to `play!` within its scope. This may be useful if you're dealing with a one-off score that you do not intend to re-use after the `with-new-score` scope ends.

```clojure
(require '[alda.now :refer :all])

(with-new-score
  (println "Playing C4")
  (play!
    (part "saw-wave"
      (note (pitch :c))
      (octave :up)))

  (Thread/sleep 2000)

  (println "Playing C5")
  (play!
    (note (pitch :c))))
```

> The return value of `with-score` and `with-new-score` is the score atom, which can then be used again subseqently with `with-score`.

