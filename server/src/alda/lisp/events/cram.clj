(ns alda.lisp.events.cram)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.cram...")

(require '[alda.util :refer (resetting)])

(defmacro tally-beats [& body]
  `(resetting [~'alda.lisp/*beats-tally*
               ~'alda.lisp/*time-scaling*
               ~'alda.lisp/*instruments*]
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
             is#    alda.lisp/*current-instruments*
             ts#    alda.lisp/*time-scaling*
             beats# (zipmap is# (for [i# is#]
                                  (or dur# ($duration i#))))]
         (doseq [i# is#]
           (binding [~'alda.lisp/*current-instruments* #{i#}
                     ~'alda.lisp/*time-scaling* (calculate-time-scaling
                                                  ts#
                                                  (or dur# ($duration i#))
                                                  tally#)]
             (set-duration 1)
             ~@body
             (set-duration (beats# i#))))
         (alter-var-root #'*time-scaling* (constantly ts#))))))
