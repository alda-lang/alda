(ns alda.lisp.events.rest
  (:require [alda.lisp.attributes     :refer :all]
            [alda.lisp.model.duration :refer (duration)]
            [alda.lisp.model.event    :refer (set-current-offset
                                              set-last-offset)]
            [alda.lisp.model.marker   :refer ($current-marker)]
            [alda.lisp.model.offset   :refer ($current-offset
                                              $last-offset
                                              offset+)]
            [alda.lisp.model.records  :refer (->Rest)]
            [alda.lisp.score.context  :refer (*beats-tally*
                                               *current-instruments*)]
            [taoensso.timbre          :as    log]))

(defn pause*
  ([instrument]
    (pause* instrument (duration ($duration instrument))))
  ([instrument {:keys [duration-fn beats] :as dur}]
    {:pre [(map? dur)]}
    (if *beats-tally*
      (alter-var-root #'*beats-tally* + beats)
      (let [rest-duration (duration-fn ($tempo instrument))]
        (set-last-offset instrument ($current-offset instrument))
        (set-current-offset instrument (offset+ ($current-offset instrument)
                                                rest-duration))
        (let [rest (->Rest ($last-offset instrument) instrument rest-duration)]
          (log/debug (format "%s rests at %s + %s for %s ms."
                             instrument
                             ($current-marker instrument)
                             (int (:offset ($last-offset instrument)))
                             (int rest-duration)))
          rest)))))

(defmacro pause
  [& args]
  `(doall
     (for [instrument# (if (and *beats-tally*
                                (not (empty? *current-instruments*)))
                         [(first *current-instruments*)]
                         *current-instruments*)]
       (binding [*current-instruments* #{instrument#}]
         (pause* instrument# ~@args)))))
