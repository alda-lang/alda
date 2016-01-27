(ns alda.lisp.score
  (:require [alda.lisp.model.offset  :refer (absolute-offset)]
            [alda.lisp.model.records :refer (->AbsoluteOffset)]
            [alda.lisp.score.context :refer :all]
            [taoensso.timbre         :as    log]))

;; for alda.repl use ;;

(defn score-text<< [s]
  (if (empty? *score-text*)
    (alter-var-root #'*score-text* str s)
    (alter-var-root #'*score-text* str \newline s)))

;;;;;;;;;;;;;;;;;;;;;;;

(defn score*
  []
  (load-score-context new-score-context))

(defn event-set
  "Takes *events* in its typical form (organized by markers with relative
   offsets) and transforms it into a single set of events with absolute
   offsets."
  [events-map]
  (into #{}
        (mapcat (fn [[_ {:keys [offset events]}]]
                  (for [event events]
                    (update-in event [:offset] absolute-offset))))
        events-map))

(defn markers [events-map]
  (into {}
        (map (fn [[marker-name {marker-offset :offset}]]
               [marker-name (absolute-offset marker-offset)]))
        events-map))

(defn score-map
  []
  (if (bound? #'*events*)
    {:events (event-set *events*)
     :markers (markers *events*)
     :instruments *instruments*}
    (log/error "A score must be initialized with (score*) before you can use (score-map).")))

(defmacro score
  "Initializes a new score, evaluates body, and returns the map containing the
   set of events resulting from evaluating the score, and information about the
   instrument instances, including their states at the end of the score.

   Restores the score evaluation context back to the way it was afterwards, so
   that score evaluation is a self-contained, pseudo-immutable process."
  [& body]
  `(let [previous-ctx# (score-context)
         score-map# (do
                      (score*)
                      ~@body
                      (score-map))]
     (load-score-context previous-ctx#)
     score-map))
