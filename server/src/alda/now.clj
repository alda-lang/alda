(ns alda.now
  (:require [clojure.set :as set]
            [alda.lisp   :as lisp]
            [alda.sound  :as sound :refer (*play-opts*)]
            [alda.util   :as util]))

(defn- prepare-audio-context!
  [score]
  (let [audio-ctx (or (:audio-context @score) (sound/new-audio-context))]
    (swap! score assoc :audio-context audio-ctx)))

(defn new-score
  ([]
   (new-score (lisp/score)))
  ([score]
   (doto (atom score) (prepare-audio-context!))))

(defn set-up!
  "Prepares the audio context of a score (creating the audio context if one
   does not already exist) to play one or more audio types.

   `score` is an atom referencing an Alda score map.

   `audio-type` (optional) is either a keyword representing an audio type, such
   as :midi, or a collection of such keywords. If this option is omitted, the
   audio types to set up will be determined based on the instruments in the
   score."
  [score & [audio-type]]
  (let [audio-types (or audio-type (sound/determine-audio-types @score))]
    (prepare-audio-context! score)
    (sound/set-up! (:audio-context @score) audio-types @score)))

(defn tear-down!
  "Cleans up after a score after you're done using it.

   Closes the MIDI synth, etc."
  [score]
  (sound/tear-down! @score))

(def ^:dynamic *current-score* nil)

(defmacro with-score
  "When `play!` is used within this scope, appends to `score` and plays any new
   notes.

   Returns the score."
  [score & body]
  `(binding [*current-score* ~score]
     ~@body
     ~score))

(defmacro with-new-score
  "Starts a new score and appends to it each time `play!` is used within this
   scope.

   Returns the score."
  [& body]
  `(let [score# (new-score)]
     (binding [*current-score* score#]
       ~@body
       score#)))

(defn play!
  "Evaluates some alda.lisp code and plays only the new events.

   By default, each call to `play!` uses a new score.

   To append to an existing score (represented as an atom reference to a score
   map), wrap multiple calls to `play!` in `(with-score <atom>)`.

   To start a new score and append to it, use `with-new-score`.

   Both `with-score` and `with-new-score` return the score that is being
   appended."
  [& body]
  (sound/with-play-opts {:async?   true
                         :one-off? (or (:one-off? *play-opts*)
                                       (not *current-score*))}
    (let [score-before (if *current-score*
                         @*current-score*
                         @(new-score))
          score-after  (apply lisp/continue score-before body)
          new-events   (set/difference (:events score-after)
                                       (:events score-before))]
      (sound/play! score-after new-events)
      (when *current-score*
        (reset! *current-score* score-after)))))

(defn play-score!
  "Plays an entire Alda score.

   The score may be represented as a map of the form that results from
   evaluating alda.lisp code, e.g. (score (part 'piano' (note (pitch :c)))),
   or an atom referencing such a map."
  [score & [play-opts]]
  (sound/with-play-opts (merge {:async? true} play-opts)
    (sound/play! (if (instance? clojure.lang.Atom score)
                   @score
                   score))))

