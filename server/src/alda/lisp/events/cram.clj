(ns alda.lisp.events.cram
  (:require [alda.lisp.model.event :refer (update-score add-events)]
            [alda.lisp.score.util  :refer (update-instruments)]))

(defn- tally-beats
  [score events]
  (-> score
      (assoc :beats-tally 0
             :beats-tally-default 1)
      (#(reduce update-score % events))
      :beats-tally))

(defn- calculate-time-scaling
  "Given an existing time-scaling value (the default is 1, when not already
   inside of a cram), the 'inner' length of a cram in beats, and the 'outer'
   length of the cram in beats, calculates the effective time-scaling value."
  [time-scaling inner-beats outer-beats]
  (* (/ time-scaling inner-beats) outer-beats))

(defn- set-time-scaling
  "Sets the time-scaling value of each instrument, based on its existing
   time-scaling and duration values and the inner and (optional) outer beats
   tally of the CRAM expression.

   Stashes each instrument's previous time-scaling value in the instrument as
   :previous-time-scaling, so it can be restored after the CRAM expression."
  [score {:keys [duration events] :as cram}]
  (update-instruments score
    (fn [{:keys [time-scaling] :as inst}]
      (let [inner-beats      (tally-beats score events)
            outer-beats      (or (:beats duration)
                                 (:duration-inside-cram inst)
                                 (:duration inst))
            new-time-scaling (calculate-time-scaling time-scaling
                                                     inner-beats
                                                     outer-beats)]
        (-> inst
            (update :previous-time-scaling (fnil conj []) time-scaling)
            (assoc :time-scaling new-time-scaling))))))

(defn- reset-time-scaling
  [score]
  (update-instruments score
    (fn [{:keys [previous-time-scaling] :as inst}]
      (-> inst
          (assoc :time-scaling (peek previous-time-scaling))
          (update :previous-time-scaling pop)))))

(defn- set-initial-duration-inside-cram
  "Sets the initial :duration-inside-cram value for each instrument to 1 beat.
   As events inside the cram are added, :duration-inside-cram is updated instead
   of :duration."
  [score]
  (update-instruments score
    (fn [{:keys [duration-inside-cram] :as inst}]
      (-> inst
          (update :previous-duration-inside-cram
                  (fnil conj []) duration-inside-cram)
          (assoc  :duration-inside-cram 1)))))

(defn- reset-duration-inside-cram
  "Removes the :duration-inside-cram value for each instrument."
  [score]
  (update-instruments score
    (fn [{:keys [previous-duration-inside-cram] :as inst}]
      (-> inst
          (assoc :duration-inside-cram (peek previous-duration-inside-cram))
          (update :previous-duration-inside-cram pop)))))

(defmethod update-score :cram
  [{:keys [current-instruments beats-tally beats-tally-default] :as score}
   {:keys [events duration] :as cram}]
  (if beats-tally
    (let [beats (or (:beats duration) beats-tally-default)]
      (-> score
          (update :beats-tally + beats)
          (assoc :beats-tally-default beats)))
    (-> score
        (update :cram-level inc)
        (set-time-scaling cram)
        set-initial-duration-inside-cram
        (#(reduce update-score % events))
        reset-duration-inside-cram
        reset-time-scaling
        (#(if-let [beats (:beats duration)]
            (update-instruments % (fn [{:keys [id] :as inst}]
                                    (if (contains? current-instruments id)
                                      (assoc inst :duration beats)
                                      inst)))
            %))
        (update :cram-level dec))))

