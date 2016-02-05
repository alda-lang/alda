(ns alda.lisp.events.rest
  (:require [alda.lisp.events.note :refer (add-note-or-rest)]
            [alda.lisp.model.event :refer (update-score)]))

(comment
  "Implementation-wise, a rest is just a note without a pitch, so a rest event
   is just a note event that updates all the instruments to have a later offset,
   without adding any events to the score.")

(defmethod update-score :rest
  [score rest-event]
  (add-note-or-rest score rest-event))

(defn pause
  "Causes every instrument in :current-instruments to rest (not play) for the
   specified duration.

   If no duration is specified, each instrument will rest for its own internal
   duration, which will be the duration last specified on a note or rest in
   that instrument's part."
  [& [{:keys [beats ms] :as dur}]]
   {:event-type :rest
    :beats      beats
    :ms         ms})

