# Inline Clojure Code

Alda allows score-writers to program in [Clojure](http://www.clojure.org) by writing inline Clojure expressions alongside Alda code.

From the perspective of Alda's parser, anything between parentheses is considered a Clojure expression. You can write Clojure expressions anywhere within in Alda score, alongside Alda syntax.

Clojure expressions are evaluated in the context of the `alda.lisp` namespace, which gives you first-class access to the [alda.lisp](alda-lisp.md) DSL. For example, out of the box you can do things like:

```clojure
piano:
  (note (pitch :c))

  (do (volume 50) (octave 6))

  (chord (note (pitch :c))
         (note (pitch :e))
         (note (pitch :g)))

  (for [letter [:c :d :e :f :g]]
    (note (pitch letter)
          (duration (note-length 8))))
```

## Evaluating strings of Alda code

The `alda-code` function provides a convenient way to parse and evaluate a string of Alda code, in the context of inline Clojure code. This gives you the power to construct strings of Alda code using Clojure, and then splice that Alda code into your score.

Here is an example where we repeat a 3-note phrase 7 times by building the string `"e f g e f g e f g e f g e f g e f g e f g "` and evaluating it:

```
cello:
  o3
  (alda-code (apply str (repeat 7 "e8 f g ")))
```

Here is another example, where we play 5 random notes out of the C major scale:

```
bassoon:
  o3
  (alda-code (apply str (repeatedly 5 #(str (rand-nth "abcdefg") \space))))
```

## Scheduling custom events

You might initially think that each Clojure expression is not evaluated until the point in time where it is situated in the score, but this is actually not the case. Clojure expressions *are* evaluated in score-order, but this happens very quickly during the brief period of time before the score is played. When the score is evaluated, it queues up a bunch of events (mostly notes being played), and it is during this very brief period that your Clojure code will run.

It is still possible, however to schedule custom events to occur at a specific time in the score, thanks to the `schedule` function.

```
piano:
  (schedule #(println "playing c")) c8
  (schedule #(println "playing d")) d
  (schedule #(println "playing e")) e
  (schedule #(println "playing f")) f
  (schedule #(println "playing g")) g2
```

`schedule` takes a function as its argument. The function that you give it can be any Clojure function that takes zero arguments.
