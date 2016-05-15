(ns alda.lisp.events.chord
  (:require [alda.lisp.model.duration :refer (max-beats)]
            [alda.lisp.model.event    :refer (update-score add-events)]
            [alda.lisp.model.offset   :refer (offset+ offset=)]
            [alda.lisp.score.util     :refer (update-instruments)]))

(comment
  "To add the note events for chords, we turn on the :chord-mode flag, which
   makes the notes all get added at the same offset.

   Then we call `initialize-min-durations`, which sets the :min-duration of
   every instrument to nil.

   As each note is evaluated, each instrument's :min-duration is updated
   accordingly with the shortest duration evaluated so far.

   Finally, we call `bump-by-min-durations`, which bumps each instrument's
   :current-offset forward by its :min-duration.")

(defn- initialize-min-durations
  [{:keys [current-instruments] :as score}]
  (update-instruments score
    (fn [{:keys [id] :as inst}]
      (if (contains? current-instruments id)
        (assoc inst :min-duration Long/MAX_VALUE)
        inst))))

(defn- bump-by-min-durations
  [{:keys [current-instruments] :as score}]
  (update-instruments score
    (fn [{:keys [id min-duration current-offset] :as inst}]
      (if (contains? current-instruments id)
        (assoc inst
               :last-offset    current-offset
               :current-offset (offset+ current-offset min-duration)
               :min-duration   nil)
        inst))))

(defmethod update-score :chord
  [{:keys [beats-tally current-instruments] :as score}
   {:keys [events] :as chord}]
  (if (and beats-tally (not (empty? current-instruments)))
    (update score :beats-tally + (max-beats events))
    (-> score
        (assoc :chord-mode true)
        initialize-min-durations
        (#(reduce update-score % events))
        bump-by-min-durations
        (assoc :chord-mode false))))

(defn chord
  "Causes every instrument in :current-instruments to play each note in the
   chord simultaneously at the instrument's :current-offset."
  [& events]
  {:event-type :chord
   :events     events})
