(ns alda.lisp.events.sequence
  (:require [alda.lisp.model.event     :refer (update-score)]
            [alda.lisp.model.attribute :refer (apply-attributes)]))

(comment
  "Rather than have an `event-sequence` function that generates an event
   sequence 'event', an event sequence is simply represented in alda.lisp as a
   sequential collection of events. You can use a list, vector, etc.

   The `update-score` multimethod's dispatch function contains logic that if
   its argument is sequential, it dispatches to the :event-sequence
   implementation below.")

(defmethod update-score :event-sequence
  [score events]
  (let [events (concat (interpose (apply-attributes) events)
                       [(apply-attributes)])]
    (reduce update-score score events)))

