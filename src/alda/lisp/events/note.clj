(ns alda.lisp.events.note)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.note...")

(defrecord Note [offset instrument volume track-volume midi-note pitch duration])

(defn note*
  ([instrument pitch-fn]
   {:pre [(fn? pitch-fn)]}
    (note* instrument
           pitch-fn
           (duration ($duration instrument))
           false))
  ([instrument pitch-fn arg3]
    (cond ; arg3 could be a duration or :slur
      (map? arg3)    (note* instrument
                            pitch-fn
                            arg3
                            false)
      (= :slur arg3) (note* instrument
                            pitch-fn
                            (duration ($duration instrument))
                            true)))
  ([instrument pitch-fn {:keys [duration-fn slurred]} slur?]
    (let [quant          (if (or slur? slurred) 1.0 ($quantization instrument))
          note-duration  (duration-fn ($tempo instrument))
          event          (map->Note
                           {:offset       ($current-offset instrument)
                            :instrument   instrument
                            :volume       ($volume instrument)
                            :track-volume ($track-volume instrument)
                            :midi-note    (pitch-fn ($octave instrument) :midi true)
                            :pitch        (pitch-fn ($octave instrument))
                            :duration     (* note-duration quant)})]
      (add-event instrument event)
      (set-last-offset instrument ($current-offset instrument))
      (set-current-offset instrument (offset+ ($current-offset instrument)
                                              note-duration))
      (log/debug (format "%s plays at %s + %s for %s ms, at %.2f Hz."
                         instrument
                         ($current-marker instrument)
                         (int (:offset (:offset event)))
                         (int (:duration event))
                         (:pitch event)))
      event)))

(defmacro note
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (note* instrument# ~@args)))))
