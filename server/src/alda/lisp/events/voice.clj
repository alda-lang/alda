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
        (max-key (comp :offset :current-offset)
                 (map #(get % id) (vals voice-instruments)))
        inst))))

(defn end-voice-group
  [score]
  (-> score
      update-instruments
      (assoc :current-voice nil)
      (assoc :voice-instruments nil)))

(defmethod update-score :voice
  [{:keys [instruments] :as score}
   {:keys [number events] :as voice}]
  (let [score (-> score
                  (assoc :current-voice number)
                  (assoc-in [:voice-instruments number] instruments))]
    (reduce update-score score events)))

(defmethod update-score :voice-group
  [score {:keys [voices] :as voice-group}]
  (let [score  (assoc score :voice-instruments {})]
    (reduce update-score score voices)))

(defn voice
  "One voice in a voice group."
  [voice-number & events]
  {:event-type :voice
   :number     voice-number
   :events     events})

(defn voices
  "Voices are chronological sequences of events that each start at the same
   time. The resulting :current-offset is at the end of the voice that finishes
   last."
  [& voices]
  {:event-type :voice-group
   :voices     voices})

