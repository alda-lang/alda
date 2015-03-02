(ns alda.lisp.events.note)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.note...")

(defrecord Note [offset instrument volume pitch duration])

(defn note*
  ([instrument pitch-fn]
   {:pre [(fn? pitch-fn)]}
    (note* instrument
           pitch-fn
           (duration (-> (*instruments* instrument) :duration))
           false))
  ([instrument pitch-fn arg3]
    (cond ; arg3 could be a duration or :slur
      (map? arg3)    (note* instrument
                            pitch-fn
                            arg3
                            false)
      (= :slur arg3) (note* instrument
                            pitch-fn
                            (duration (-> (*instruments* instrument)
                                          :duration))
                            true)))
  ([instrument pitch-fn {:keys [duration-fn slurred]} slur?]
    (let [get-attribute (fn [attr]
                          (fn []
                            (-> (*instruments* instrument) attr)))
          tempo          (get-attribute :tempo)
          volume         (get-attribute :volume)
          octave         (get-attribute :octave)
          current-offset (get-attribute :current-offset)
          current-marker (get-attribute :current-marker)
          quant          (if (or slur? slurred) 1.0 ((get-attribute :quantization)))
          note-duration  (duration-fn (tempo))
          event          (map->Note {:offset (current-offset)
                                     :instrument instrument
                                     :volume (volume)
                                     :pitch (pitch-fn (octave))
                                     :duration (* note-duration quant)})]
      (add-event instrument event)
      (set-last-offset instrument (current-offset))
      (set-current-offset instrument (offset+ (current-offset) note-duration))
      (log/debug (format "%s plays at %s + %s for %s ms, at %.2f Hz."
                         instrument
                         (current-marker)
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
