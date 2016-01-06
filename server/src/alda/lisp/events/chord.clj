(ns alda.lisp.events.chord
  (:require [alda.lisp.model.event   :refer (set-current-offset
                                             set-last-offset)]
            [alda.lisp.model.marker  :refer ($current-marker)]
            [alda.lisp.model.offset  :refer ($current-offset offset=)]
            [alda.lisp.model.records :refer (->Chord)]
            [alda.lisp.score.context :refer (*beats-tally*
                                             *current-instruments*
                                             *events*)]))

(defmacro tally-chord-duration
  "Determines the duration of all events in the chord and adds the longest one
   to *beats-tally*."
  [& events]
  (let [start   (gensym "start")
        tallies (gensym "tallies")]
    (list* 'let [start `*beats-tally*
                 tallies (list 'atom [])]
           (concat
             (interleave
               (repeat `(alter-var-root (var *beats-tally*)
                                        (constantly ~start)))
               events
               (repeat `(swap! ~tallies conj *beats-tally*)))
             [`(alter-var-root (var *beats-tally*)
                               (constantly (apply max (deref ~tallies))))]))))

(defmacro chord*
  "Chords contain notes/rests that all start at the same time/offset.
   The resulting *current-offset* is at the end of the shortest note/rest in
   the chord."
  [instrument & events]
  (let [num-of-events  (count (filter #(= (first %) 'note) events))
        start          (gensym "start")
        offsets        (gensym "offsets")]
    (list* 'let [start   (list `$current-offset instrument)
                 offsets (list 'atom [])]
           (concat
             (interleave
               (repeat `(set-current-offset ~instrument ~start))
               events
               (repeat `(swap! ~offsets conj ($current-offset ~instrument))))
             [`(set-last-offset ~instrument ~start)
              `(set-current-offset ~instrument (apply (partial min-key :offset)
                                                      (remove #(offset= % ~start)
                                                              (deref ~offsets))))
              `(let [chord#
                     (->Chord (take-last ~num-of-events
                                        (get-in *events*
                                                [($current-marker ~instrument)
                                                 :events])))]
                 chord#)]))))

(defmacro chord
  [& args]
  `(if (and *beats-tally* (not (empty? *current-instruments*)))
     (tally-chord-duration ~@args)
     (doall
       (for [instrument# *current-instruments*]
         (binding [*current-instruments* #{instrument#}]
           (chord* instrument# ~@args))))))
