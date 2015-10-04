(ns alda.now)

(require '[alda.sound  :as sound]
         '[alda.util   :as util]
         '[clojure.set :as set])

; sets log level to TIMBRE_LEVEL (if set) or :warn
(util/set-timbre-level!)

(require '[alda.lisp :as lisp])

(def set-up! sound/set-up!)

(defn play-new-events!
  [score new-events]
  (let [events        (lisp/event-set {:start
                                       {:offset (lisp/->AbsoluteOffset 0)
                                        :events new-events}})
        earliest      (->> (map :offset events)
                           (apply min Long/MAX_VALUE)
                           (max 0))
        shifted       (alda.sound/shift-events events earliest nil)
        one-off-score (assoc score
                             :events shifted
                             :markers (into {}
                                            (map (juxt first #(- (last %) earliest)))
                                            (:markers score)))]
    (sound/play! one-off-score)))

(defmacro play!
  "Evaluates some alda.lisp code and plays only the new events."
  [& body]
  `(let [old-score# (lisp/score-map)
         new-score# (do ~@body (lisp/score-map))
         new-events# (set/difference
                       (:events new-score#)
                       (:events old-score#))]
     (play-new-events! new-score# new-events#)))

(defn refresh!
  "Clears all events and resets the current-offset of each instrument to 0.

   Useful for playing a new set of notes with multiple instrument parts,
   ensuring that both parts start at the same time, regardless of any prior
   difference in current-offset between the instrument parts.

   When a truthy argument is provided, also resets all the other attributes
   (e.g. volume, track-volume, octave) to their default values."
  [& [all?]]
  (alter-var-root #'alda.lisp/*instruments*
    #(into {}
       (map (fn [[instrument attrs]]
              [instrument
               (merge attrs
                      (if all?
                        lisp/*initial-attr-values*
                        (select-keys lisp/*initial-attr-values*
                                     [:current-offset :last-offset])))])
            %)))
  (alter-var-root #'alda.lisp/*events*
    (constantly {:start {:offset (lisp/->AbsoluteOffset 0), :events []}})))
