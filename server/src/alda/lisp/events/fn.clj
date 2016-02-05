(ns alda.lisp.events.fn
  (:require [alda.lisp.model.event   :refer (update-score add-events)]
            [alda.lisp.model.records :refer (->Function)]))

(defmethod update-score :function
  [{:keys [beats-tally current-instruments] :as score}
   {:keys [function] :as fn-event}]
  (if beats-tally
    score
    (add-events score (map (fn [{:keys [current-offset id]}]
                             (->Function current-offset id function))
                           current-instruments))))

(defn schedule
  "Schedules an arbitrary function to be called at the current point in the
   score (determined by the current instrument's marker and offset).

   If there are multiple current instruments, the function will be executed
   once for each instrument, at the marker + offset of that instrument."
  [f]
  {:event-type :function
   :function   f})
