(ns alda.lisp.events.rest)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.rest...")

(defrecord Rest [offset instrument duration])

(defn pause*
  ([instrument]
    (pause* instrument (duration (-> (*instruments* instrument) :duration))))
  ([instrument {:keys [duration-fn] :as dur}]
    {:pre [(map? dur)]}
    (let [get-attribute (fn [attr]
                          (fn []
                            (-> (*instruments* instrument) attr)))
          current-offset (get-attribute :current-offset)
          current-marker (get-attribute :current-marker)
          last-offset    (get-attribute :last-offset)
          tempo          (get-attribute :tempo)
          rest-duration  (duration-fn (tempo))]
      (set-last-offset instrument (current-offset))
      (set-current-offset instrument (offset+ (current-offset) rest-duration))
      (let [rest (Rest. (last-offset) instrument rest-duration)]
        (log/debug (format "%s rests at %s + %s for %s ms."
                           instrument
                           (current-marker)
                           (int (:offset (last-offset)))
                           (int rest-duration)))
        rest))))

(defmacro pause
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (pause* instrument# ~@args)))))
