(ns alda.lisp.events.variable
  (:require [alda.lisp.model.event :refer (update-score)]))

(defmethod update-score :set-variable
  [{:keys [variables env] :as score}
   {:keys [variable events]}]
  (assoc-in score [:variables variable] {:env    (or env variables)
                                         :events events}))

(defmethod update-score :get-variable
  [{:keys [variables env] :as score}
   {:keys [variable]}]
  (if env
    (update-score score (get-in env [variable :events]))
    (if-let [{:keys [env events]} (get variables variable)]
      (-> score
          (assoc :env env)
          (update-score events)
          (dissoc :env))
      (throw (Exception. (str "Undefined variable: " (name variable)))))))

