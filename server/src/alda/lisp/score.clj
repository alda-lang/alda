(ns alda.lisp.score
  (:require [alda.lisp.model.event     :refer (update-score)]
            [alda.lisp.model.attribute :refer (apply-attributes)]
            [alda.lisp.model.offset    :refer (absolute-offset)]
            [alda.lisp.model.records   :refer (->AbsoluteOffset)]
            [taoensso.timbre           :as    log]))

(defn new-score
  []
  (log/debug "Starting new score.")
  {
   :events              {:start {:offset (->AbsoluteOffset 0), :events []}}

   ; a map of offsets to the global attribute changes that occur (for all
   ; instruments) at each offset
   :global-attributes   (sorted-map)

   :current-instruments #{}
   :instruments         {}
   :nicknames           {}

   ; used when tallying beats in a CRAM expression
   :beats-tally         nil
   ; used when tallying beats in a CRAM expression
   :beats-tally-default nil

   ; used when adding events in a chord
   :chord-mode          false
   ; used when adding events in a voice
   :current-voice       nil
   ; this number gets incremented each time a cram event starts, and
   ; decremented when the cram event ends
   :cram-level          0

   ; used when inside of a voice group; this is a mapping of voice numbers to
   ; the current state of :instruments within that voice
   ; e.g. {1 {"guitar-abc123" {:current-offset (->AbsoluteOffset 1000.0) ...}}
   ;       2 {"guitar-abc123" {:current-offset (->AbsoluteOffset 2000.0) ...}}}
   :voice-instruments   nil
   })

(defn continue
  "Continues the score represented by the score map `score`, evaluating the
   events in `body` and returning the completed score."
  [score & body]
  (let [events (concat (interpose (apply-attributes) body)
                       [(apply-attributes)])]
    (reduce update-score score events)))

(defn continue!
  "Convenience function for dealing with Alda scores stored in atoms.

   (continue! my-score
     (part 'bassoon'
       (note (pitch :c))))

   is short for:

   (apply swap! my-score continue
     (part 'bassoon'
       (note (pitch :c))))"
  [score-atom & body]
  (apply swap! score-atom continue body))

(defn score
  "Initializes a new score, evaluates the events contained in `body` (updating
   the score accordingly) and returns the completed score.

   A score and its evaluation context are effectively the same thing. This
   means that an evaluated score can be used as an input to `continue-score`"
  [& body]
  (apply continue (new-score) body))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn event-set
  "Given a score, takes its :events map in its typical form (organized by
   markers with relative offsets) and transforms it into a single set of
   events with absolute offsets."
  [{:keys [events] :as score}]
  (into #{}
    (mapcat (fn [[_ {:keys [offset events]}]]
              (for [event events]
                (update-in event [:offset] #(absolute-offset % score))))
            events)))

(defn markers
  "Returns a map of each marker in a score to its absolute offset."
  [{:keys [events] :as score}]
  (into {}
    (map (fn [[marker-name {marker-offset :offset}]]
           [marker-name (absolute-offset marker-offset score)])
         events)))

