# alda.now

`alda.now`, when coupled with [`alda.lisp`](alda-lisp.md), provides a way to work with Alda scores and play music programmatically within a Clojure application.

`alda.now` provides a `play!` macro, which evaluates the body, finds any new note events that were added to the score, and plays them.

Example usage of `alda.now` in a Clojure application:

```clojure
(require '[alda.lisp :refer :all])
(require '[alda.now  :refer (set-up! play!)])

(score*)
(part* "upright-bass")

; This is optional. If left out, Alda will set up the MIDI synth the first
; time you tell it to play something.
(set-up! :midi)

(play!
  (octave 2)
  (note (pitch :c) (duration (note-length 8)))
  (note (pitch :d))
  (note (pitch :e))
  (note (pitch :f))
  (note (pitch :g) (duration (note-length 4))))
```

Of note, [`alda.repl`](alda-repl.md) uses `alda.now` to play the score the user is creating during the REPL session, so you could think of `alda.repl` as an `alda.now` sample project.

## MIDI soundfonts

The default JVM soundfont sounds pretty bad. If you're using Alda in a Clojure application and you want to have nice MIDI sounds, you can use [`midi.soundfont`](https://github.com/daveyarwood/midi.soundfont) to load FluidR3 or another MIDI soundfont into the JVM at runtime.

