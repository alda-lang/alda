(ns alda.now
  (:require [clojure.set :as set]
            [alda.lisp   :as lisp]
            [alda.sound  :as sound]
            [alda.util   :as util]))

(comment
  "TODO: move these explanations to docs

   - set-up! requires one argument, a score atom, and sets it up ahead of time
     (like before, this will happen automatically once play! is called, but
     set-up! can be used to do this in advance if desired)

     There is also an optional second argument, which can be either a keyword
     representing an audio type (e.g. :midi) to set up or a collection of such
     keywords.

     (set-up! my-score)
     (set-up! my-score :midi)
     (set-up! my-score [:midi])

   - By default, play! is just a shortcut for creating a one-off score and
     playing it via alda.sound/play!
     - Each time you use it, you're creating a new score from scratch.

   - You can also use with-score, which will append to an existing score and
     play any new notes. The existing score is a score map wrapped in an atom.

     This might look something like:

       (def my-score (atom (score)))

       (with-score my-score
         (play!
           (note (pitch :c))
           (note (pitch :d))
           (note (pitch :e))))

       (with-score my-score
         (play!
           (note (pitch :f))
           (note (pitch :g))
           (note (pitch :a))))

   - You can also use with-new-score, which is equivalent to defining a new
     score atom and using it with with-score:

     (with-new-score
       (play!
         (note (pitch :c))
         (note (pitch :e))))

   - Both with-score and with-new-score will return the score atom when done.")

(defn- prepare-audio-context!
  [score]
  (let [audio-ctx (or (:audio-context @score) (sound/new-audio-context))]
    (swap! score assoc :audio-context audio-ctx)))

(defn new-score
  []
  (doto (atom (lisp/score)) (prepare-audio-context!)))

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
  (sound/with-play-opts {:async? true}
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
  [score]
  (sound/with-play-opts {:async? true}
    (sound/play! (if (instance? clojure.lang.Atom score)
                   @score
                   score))))

