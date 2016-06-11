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
  (let [score  (assoc score :voice-instruments {})]
    (reduce update-score score voices)))

(defmethod update-score :end-voice-group
  [score _]
  (end-voice-group score))

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

(defn end-voices
  "By default, the score remains in 'voice mode' until it reaches an end-voices
   event. This is so that if an instrument part ends with a voice group, the
   same voices can be appended later if the part is resumed, e.g. when building
   a score gradually in the Alda REPL or in a Clojure process.

   The end-voices event is emitted by the parser when it parses 'V0:'."
  []
  {:event-type :end-voice-group})
