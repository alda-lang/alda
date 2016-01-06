(ns alda.lisp.events.note
  (:require [alda.lisp.attributes     :refer :all]
            [alda.lisp.model.duration :refer (duration)]
            [alda.lisp.model.event    :refer (add-event
                                              set-current-offset
                                              set-last-offset)]
            [alda.lisp.model.marker   :refer ($current-marker)]
            [alda.lisp.model.offset   :refer ($current-offset
                                              offset+)]
            [alda.lisp.model.records  :refer (map->Note)]
            [alda.lisp.score.context  :refer (*beats-tally*
                                              *current-instruments*)]
            [taoensso.timbre          :as    log]))

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
  ([instrument pitch-fn {:keys [duration-fn beats slurred]} slur?]
    (let [quant          (if (or slur? slurred) 1.0 ($quantization instrument))
          note-duration  (duration-fn ($tempo instrument))
          quant-duration (* note-duration quant)
          event          (when-not *beats-tally*
                           (map->Note
                             {:offset       ($current-offset instrument)
                              :instrument   instrument
                              :volume       ($volume instrument)
                              :track-volume ($track-volume instrument)
                              :panning      ($panning instrument)
                              :midi-note    (pitch-fn ($octave instrument)
                                                      ($key-signature instrument)
                                                      :midi true)
                              :pitch        (pitch-fn ($octave instrument)
                                                      ($key-signature instrument))
                              :duration     quant-duration}))]
      (if event
        (do
          (when (pos? quant-duration) (add-event instrument event))
          (set-last-offset instrument ($current-offset instrument))
          (set-current-offset instrument (offset+ ($current-offset instrument)
                                                  note-duration))
          (log/debug (format "%s plays at %s + %s for %s ms, at %.2f Hz."
                             instrument
                             ($current-marker instrument)
                             (int (:offset (:offset event)))
                             (int (:duration event))
                             (:pitch event)))
          event)
        (alter-var-root #'*beats-tally* + beats)))))

(defmacro note
  [& args]
  `(doall
     (for [instrument# (if (and *beats-tally*
                                (not (empty? *current-instruments*)))
                         [(first *current-instruments*)]
                         *current-instruments*)]
       (binding [*current-instruments* #{instrument#}]
         (note* instrument# ~@args)))))
