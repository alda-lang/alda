(ns alda.lisp.model.event
  (:require [alda.lisp.model.offset :refer (absolute-offset)]
            [alda.lisp.model.records])
  (:import  [alda.lisp.model.records RelativeOffset]))

(defmulti update-score
  "Events in Alda are represented as maps containing, at the minimum, a value
   for :event-type to serve as a unique identifier (by convention, a keyword)
   to be used as a dispatch value.

   An Alda score S-expression simply reduces `update-score` over all of the
   score's events, with the initial score state as the initial value to be
   reduced.

   Lists/vectors are a special case -- they are reduced internally and treated
   as a single 'event sequence'."
  (fn [score event]
    (cond
      (or (nil? event) (var? event)) :nil
      (sequential? event)            :event-sequence
      :else                          (:event-type event))))

(defmethod update-score :default
  [_ event]
  (throw (Exception. (str "Invalid event: " (pr-str event)))))

(defmethod update-score :nil
  [score _]
  ; pass score through unchanged
  ; e.g. for side-effecting inline Clojure code
  score)

; utility fns

(defn add-event
  [{:keys [instruments events markers] :as score}
   {:keys [instrument offset] :as event}]
  (update score :events conj (update event :offset #(absolute-offset % score))))

(defn add-events
  [score events]
  (reduce add-event score events))

