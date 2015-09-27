# alda.lisp

Under the hood, Alda transforms input (i.e. Alda code) into Clojure code which, when evaluated, produces a map of score information, which the audio component of Alda can then use to make sound. This Clojure code is written in a DSL called **alda.lisp**. See below for an example of alda.lisp code and the result of evaluating it.

## Parsing demo

You can use the `parse` task to parse Alda code into alda.lisp (`-l`/`--lisp`) and/or evaluate it to produce a map (`-m`/`--map`) of score information.

    $ alda parse --lisp --map -f test/examples/hello_world.alda

    (alda.lisp/score
     (alda.lisp/part
      {:names ["piano"]}
      (alda.lisp/note
       (alda.lisp/pitch :c)
       (alda.lisp/duration (alda.lisp/note-length 8)))
      (alda.lisp/note (alda.lisp/pitch :d))
      (alda.lisp/note (alda.lisp/pitch :e))
      (alda.lisp/note (alda.lisp/pitch :f))
      (alda.lisp/note (alda.lisp/pitch :g))
      (alda.lisp/note (alda.lisp/pitch :f))
      (alda.lisp/note (alda.lisp/pitch :e))
      (alda.lisp/note (alda.lisp/pitch :d))
      (alda.lisp/note
       (alda.lisp/pitch :c)
       (alda.lisp/duration (alda.lisp/note-length 2 {:dots 1})))))
    
    {:events
     #{{:offset 500.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 64,
        :pitch 329.6275569128699,
        :duration 225.0}
       {:offset 2000.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 60,
        :pitch 261.6255653005986,
        :duration 1350.0}
       {:offset 1750.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 62,
        :pitch 293.6647679174076,
        :duration 225.0}
       {:offset 1000.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 67,
        :pitch 391.99543598174927,
        :duration 225.0}
       {:offset 0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 60,
        :pitch 261.6255653005986,
        :duration 225.0}
       {:offset 1250.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 65,
        :pitch 349.2282314330039,
        :duration 225.0}
       {:offset 750.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 65,
        :pitch 349.2282314330039,
        :duration 225.0}
       {:offset 250.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 62,
        :pitch 293.6647679174076,
        :duration 225.0}
       {:offset 1500.0,
        :instrument "piano-y50tv",
        :volume 1.0,
        :track-volume 0.7874015748031497,
        :midi-note 64,
        :pitch 329.6275569128699,
        :duration 225.0}},
     :markers {:start 0},
     :instruments
     {"piano-y50tv"
      {:octave 4,
       :current-offset {:offset 3500.0},
       :config {:type :midi, :patch 1},
       :duration 3.0,
       :volume 1.0,
       :last-offset {:offset 2000.0},
       :id "piano-y50tv",
       :quantization 0.9,
       :tempo 120,
       :panning 0.5,
       :current-marker :start,
       :stock "midi-acoustic-grand-piano",
       :track-volume 0.7874015748031497}}}

    $ alda parse --lisp -c 'cello: c+'

    (alda.lisp/score
      (alda.lisp/part {:names ["cello"]}
        (alda.lisp/note (alda.lisp/pitch :c :sharp))))

