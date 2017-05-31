# alda.lisp

Under the hood, Alda parses input (i.e. Alda code) into a sequence of score
events that do things like change instruments, play notes, etc. The sequence of
events is then used to build a Clojure map of score information, which the audio component of Alda can then use to make sound.

The process of building a score from events is done via a Clojure [DSL](https://en.wikipedia.org/wiki/Domain-specific_language) called **alda.lisp**.

## Example

As an alternative to writing scores in Alda syntax, the alda.lisp DSL can also
be used directly within a Clojure program.

```clojure
(require '[alda.lisp :refer :all])

(score
  (part {:names ["piano"]}
    (note
      (pitch :c)
      (duration (note-length 8)))
    (note (pitch :d))
    (note (pitch :e))
    (note (pitch :f))
    (note (pitch :g))
    (note (pitch :f))
    (note (pitch :e))
    (note (pitch :d))
    (note
      (pitch :c)
      (duration (note-length 2 {:dots 1})))))
```

The result of evaluating the above is a score data map that is exactly like something the Alda parser would produce:

```clojure
{:chord-mode false,
 :current-instruments #{"piano-Id8yG"},
 :events
 #{{:offset 250.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 62,
    :pitch 293.6647679174076,
    :duration 225.0,
    :voice nil}
   {:offset 2000.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 60,
    :pitch 261.6255653005986,
    :duration 1350.0,
    :voice nil}
   {:offset 0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 60,
    :pitch 261.6255653005986,
    :duration 225.0,
    :voice nil}
   {:offset 500.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 64,
    :pitch 329.6275569128699,
    :duration 225.0,
    :voice nil}
   {:offset 750.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 65,
    :pitch 349.2282314330039,
    :duration 225.0,
    :voice nil}
   {:offset 1000.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 67,
    :pitch 391.99543598174927,
    :duration 225.0,
    :voice nil}
   {:offset 1250.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 65,
    :pitch 349.2282314330039,
    :duration 225.0,
    :voice nil}
   {:offset 1500.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 64,
    :pitch 329.6275569128699,
    :duration 225.0,
    :voice nil}
   {:offset 1750.0,
    :instrument "piano-Id8yG",
    :volume 1.0,
    :track-volume 0.7874015748031497,
    :panning 0.5,
    :midi-note 62,
    :pitch 293.6647679174076,
    :duration 225.0,
    :voice nil}},
 :beats-tally nil,
 :instruments
 {"piano-Id8yG"
  {:octave 4,
   :current-offset {:offset 3500.0},
   :key-signature {},
   :config {:type :midi, :patch 1},
   :duration 3.0,
   :min-duration nil,
   :volume 1.0,
   :last-offset {:offset 2000.0},
   :id "piano-Id8yG",
   :quantization 0.9,
   :duration-inside-cram nil,
   :tempo 120,
   :panning 0.5,
   :current-marker :start,
   :time-scaling 1,
   :stock "midi-acoustic-grand-piano",
   :track-volume 0.7874015748031497}},
 :markers {:start 0},
 :cram-level 0,
 :global-attributes {},
 :nicknames {},
 ...}
```

## Inline Clojure Code

alda.lisp can also be used [inside of an Alda score](inline-clojure-code.md),
providing a way to generate score events dynamically by programming in Clojure.
