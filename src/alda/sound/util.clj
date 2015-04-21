(ns alda.sound.util)

(defn score-length
  "Calculates the length of a score in ms."
  [{:keys [events] :as score}]
  (when (and events (not (empty? events)))
    (letfn [(note-end [{:keys [offset duration] :as note}] (+ offset duration))]
      (apply max (map note-end events)))))
