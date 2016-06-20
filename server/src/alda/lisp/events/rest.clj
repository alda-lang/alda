(ns alda.lisp.events.rest
  (:require [alda.lisp.events      :refer (apply-global-attributes)]
            [alda.lisp.events.note :refer (add-note-or-rest)]
            [alda.lisp.model.event :refer (update-score)]))

(comment
  "Implementation-wise, a rest is just a note without a pitch, so a rest event
   is just a note event that updates all the instruments to have a later offset,
   without adding any events to the score.")

(defmethod update-score :rest
  [score rest-event]
  (-> score
      (add-note-or-rest rest-event)
      (update-score (apply-global-attributes))))

