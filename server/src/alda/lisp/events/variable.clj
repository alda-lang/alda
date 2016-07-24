(ns alda.lisp.events.variable
  (:require [alda.lisp.model.event :refer (update-score)]))

(defn- undefined-variable-error!
  [variable]
  (throw (Exception. (str "Undefined variable: " (name variable)))))

(defmethod update-score :set-variable
  [{:keys [variables env] :as score}
   {:keys [variable events]}]
  (assoc-in score [:variables variable] {:env    (or env variables {})
                                         :events events}))

(defmethod update-score :get-variable
  [{:keys [variables env] :as score}
   {:keys [variable]}]
  (try
    (if env
      (if-let [{:keys [events]} (get env variable)]
        (update-score score events)
        (undefined-variable-error! variable))
      (if-let [{:keys [env events]} (get variables variable)]
        (-> score
            (assoc :env env)
            (update-score events)
            (dissoc :env))
        (undefined-variable-error! variable)))
    (catch StackOverflowError e
      (throw (Exception.
               (str "Stack overflow trying to get variable "
                    (name variable)
                    " -- is it used in its own definition?"))))))

