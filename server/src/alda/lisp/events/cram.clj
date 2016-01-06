(ns alda.lisp.events.cram
  (:require [alda.lisp.attributes    :refer ($duration set-duration)]
            [alda.lisp.score.context :refer (*beats-tally*
                                             *current-instruments*
                                             *instruments*
                                             *time-scaling*)]
            [alda.util               :refer (resetting)]))

(defmacro tally-beats [& body]
  `(resetting [*beats-tally* *time-scaling* *instruments*]
     (alter-var-root (var *beats-tally*) (constantly 0))
     (set-duration 1)
     ~@body
     *beats-tally*))

(defn calculate-time-scaling
  "Given a *time-scaling* value, the 'outer' length of a cram in beats, and the
   'inner' length of the cram in beats, calculates the effective time-scaling
   value."
  [time-scaling outer-beats inner-beats]
  (* (/ time-scaling inner-beats) outer-beats))

(defmacro cram [& body]
  (let [lst (last body)
        dur (and (coll? lst)
                 (contains? #{'duration 'alda.lisp/duration} (first lst))
                 lst)
        body (if dur (butlast body) body)]
    `(if *beats-tally*
       (alter-var-root #'*beats-tally* + (or ~dur ($duration)))
       (let [dur#   (:beats ~dur)
             tally# (tally-beats ~@body)
             is#    *current-instruments*
             ts#    *time-scaling*
             beats# (zipmap is# (for [i# is#]
                                  (or dur# ($duration i#))))
             events#
             (mapcat (fn [i#]
                       (binding [~'alda.lisp/*current-instruments* #{i#}
                                 ~'alda.lisp/*time-scaling*
                                 (calculate-time-scaling ts#
                                                         (or dur# ($duration i#))
                                                         tally#)]
                         (set-duration 1)
                         (let [es# [~@body]]
                           (set-duration (beats# i#))
                           es#)))
                     is#)]
         (alter-var-root #'*time-scaling* (constantly ts#))
         events#))))

