(ns alda.lisp.events.voice
  (:require [alda.lisp.model.event :refer (update-score)]
            [alda.lisp.score.util  :as    score]))

(defn- update-instruments
  "As the events for each voice are added to the score, the state of each
   instrument is tracked separately (in :voice-instruments) for each voice.

   After all these events are added to the score, we call this function to
   set each instrument's state to that of the last voice to finish."
  [{:keys [current-instruments voice-instruments] :as score}]
  (score/update-instruments score
    (fn [{:keys [id] :as inst}]
      (if (and (contains? current-instruments id) voice-instruments)
        (apply max-key
               (comp :offset :current-offset)
               (map #(get % id) (vals voice-instruments)))
        inst))))

(defn end-voice-group
  [score]
  (-> score
      (dissoc :current-voice)
      update-instruments
      (dissoc :voice-instruments)))

(defmethod update-score :voice
  [{:keys [instruments] :as score}
   {:keys [number events] :as voice}]
  (let [score (-> score
                  (assoc :current-voice number)
                  (update-in [:voice-instruments number]
                     #(if % % instruments)))]
    (reduce update-score score events)))

(defmethod update-score :voice-group
  [score {:keys [voices] :as voice-group}]
  (let [score (-> score 
                  (end-voice-group) 
                  (assoc :voice-instruments {}))]
    (reduce update-score score voices)))

(defmethod update-score :end-voice-group
  [score _]
  (end-voice-group score))

