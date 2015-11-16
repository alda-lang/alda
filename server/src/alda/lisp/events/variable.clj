(ns alda.lisp.events.variable
  (:require [alda.lisp.model.event :refer (update-score)]))

(defmethod update-score :set-variable
  [score {:keys [variable events]}]
  (assoc-in score [:variables variable] events))

(defmethod update-score :get-variable
  [{:keys [variables] :as score} {:keys [variable]}]
  (update-score score (get variables variable)))

