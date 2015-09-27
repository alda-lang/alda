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

Of note, [`alda.repl`](alda-repl.md) uses `alda.now` to play the score the user is creating during the REPL session, so you could think of `alda.repl` as an `alda.now` "sample project."

Another thing to note is that `alda.now` does not load the FluidR3 MIDI soundfont like the CLI version of Alda does by default. In the near future, Alda may be packaged with FluidR3 and `alda.now` could provide a simple helper method to load FluidR3. Currently, FluidR3 is loaded dynamically by the Alda CLI.

If you are interested in using FluidR3 or other MIDI soundfonts with Alda in a Clojure application, you can use [`midi.soundfont`](https://github.com/daveyarwood/midi.soundfont).

