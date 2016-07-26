(ns alda.lisp.events.variable
  (:require [alda.lisp.model.event :refer (update-score)]))

(defn- undefined-variable-error!
  [variable]
  (throw (Exception. (str "Undefined variable: " (name variable)))))

(defn- get-variable
  "Given an `env` reflecting the current state of variables defined in a score
   and a `variable` name, returns the stored value of the variable in the env,
   or throws an undefined variable error if the variable is undefined."
  [env variable]
  (or (get env variable) (undefined-variable-error! variable)))

(defn- replace-variables
  "Given a sequence of `events` and an `env` reflecting the current state of
   variables defined in a score, replaces all :get-variable events in `events`
   with their values from the `env`.

   Throws an undefined variable error if the variable doesn't exist in the env."
  [events env]
  (doall (for [{:keys [event-type variable] :as event} events
               :when (not (empty? event))]
           (if (= event-type :get-variable)
             (get-variable env variable)
             event))))

(defmethod update-score :set-variable
  [{:keys [env] :as score}
   {:keys [variable events]}]
  (let [events (replace-variables events env)]
    (assoc-in score [:env variable] events)))

(defmethod update-score :get-variable
  [{:keys [env] :as score}
   {:keys [variable]}]
  (update-score score (get-variable env variable)))
