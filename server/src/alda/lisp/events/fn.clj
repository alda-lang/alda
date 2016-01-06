(ns alda.lisp.events.fn
  (:require [alda.lisp.model.event   :refer (add-event)]
            [alda.lisp.model.offset  :refer ($current-offset)]
            [alda.lisp.model.records :refer (->Function)]
            [alda.lisp.score.context :refer (*beats-tally*
                                             *current-instruments*)]))

(defn schedule
  "Schedules an arbitrary function to be called at the current point in the
   score (determined by the current instrument's marker and offset).

   If there are multiple current instruments, the function will be executed
   once for each instrument, at the marker + offset of that instrument."
  [f]
  (when-not *beats-tally*
    (doseq [instrument *current-instruments*]
      (let [event (->Function ($current-offset instrument) f)]
        (add-event instrument event)))))
