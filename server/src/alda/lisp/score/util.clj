(ns alda.lisp.score.util)

(defn merge-instruments
  [score updated-insts]
  (update score :instruments merge updated-insts))

(defn merge-voice-instruments
  [score voice-number updated-insts]
  (update-in score [:voice-instruments voice-number] merge updated-insts))

(defn- instruments-key
  [{:keys [current-voice] :as score}]
  (if (and current-voice (not= 0 current-voice))
    [:voice-instruments current-voice]
    [:instruments]))

(defn update-instruments
  [score update-fn]
  (update-in score
             (instruments-key score)
             #(into {}
                (map (fn [[id inst]]
                       [id (update-fn inst)]) %))))

(defn get-current-instruments
  [{:keys [current-instruments] :as score}]
  (map (get-in score (instruments-key score)) current-instruments))
