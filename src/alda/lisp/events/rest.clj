(ns alda.lisp.events.rest)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.rest...")

(defrecord Rest [offset instrument duration])

(defn pause*
  ([instrument]
    (pause* instrument (duration ($duration instrument))))
  ([instrument {:keys [duration-fn] :as dur}]
    {:pre [(map? dur)]}
    (let [rest-duration  (duration-fn ($tempo instrument))]
      (set-last-offset instrument ($current-offset instrument))
      (set-current-offset instrument (offset+ ($current-offset instrument)
                                              rest-duration))
      (let [rest (Rest. ($last-offset instrument) instrument rest-duration)]
        (log/debug (format "%s rests at %s + %s for %s ms."
                           instrument
                           ($current-marker instrument)
                           (int (:offset ($last-offset instrument)))
                           (int rest-duration)))
        rest))))

(defmacro pause
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (pause* instrument# ~@args)))))
