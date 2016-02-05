(ns alda.lisp.score.util)

(defn merge-instruments
  [score updated-insts]
  (update score :instruments merge updated-insts))

(defn merge-voice-instruments
  [score voice-number updated-insts]
  (update-in score [:voice-instruments voice-number] merge updated-insts))

(defn update-instruments
  [{:keys [current-voice] :as score} update-fn]
  (let [instruments-key (if current-voice
                          [:voice-instruments current-voice]
                          [:instruments])]
    (update-in score instruments-key
               #(into {}
                  (map (fn [[id inst]] [id (update-fn inst)]) %)))))
