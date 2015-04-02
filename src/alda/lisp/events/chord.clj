(ns alda.lisp.events.chord)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.chord...")

(defrecord Chord [events])

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
                     (Chord. (take-last ~num-of-events
                                        (get-in *events*
                                                [($current-marker ~instrument)
                                                 :events])))]
                 chord#)]))))

(defmacro chord
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (chord* instrument# ~@args)))))
