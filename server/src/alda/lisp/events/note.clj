(ns alda.lisp.events.note
  (:require [alda.lisp.model.duration :refer (calculate-duration)]
            [alda.lisp.model.event    :refer (update-score add-events)]
            [alda.lisp.model.offset   :refer (offset+)]
            [alda.lisp.model.records  :refer (map->Note)]
            [alda.lisp.score.util     :refer (merge-instruments
                                              merge-voice-instruments)]
            [taoensso.timbre          :as    log]))

(defn- event-updates
  "Given a score and a (note/rest) event, returns a list of updates for each
   currently active instrument.

   Each update is a map containing:

   :events -- a list of note events to be added to the score

   :state -- any number of keys with updated values. This will be merged into
   the existing state of the instrument."
  [{:keys [instruments voice-instruments current-instruments chord-mode
           cram-level current-voice] :as score}
   {:keys [event-type pitch-fn beats ms slur?] :as event}]
  (for [{:keys [id duration duration-inside-cram time-scaling tempo
                current-offset last-offset current-marker quantization volume
                track-volume panning octave key-signature min-duration]
         :as inst}
        (map (if voice-instruments
               (voice-instruments current-voice)
               instruments)
             current-instruments)]
    (let [[beats ms]      (if (or beats ms)
                            [beats ms]
                            [(or duration-inside-cram duration) nil])
          quant           (if slur? 1.0 quantization)
          full-duration   (calculate-duration beats
                                              tempo
                                              time-scaling
                                              ms)
          quant-duration  (* full-duration quant)
          pitch           (if (= event-type :note)
                            (pitch-fn octave key-signature))
          midi-note       (if (= event-type :note)
                            (pitch-fn octave key-signature :midi true))
          note            (if (= event-type :note)
                            (map->Note
                              {:offset       current-offset
                               :instrument   id
                               :volume       volume
                               :track-volume track-volume
                               :panning      panning
                               :midi-note    midi-note
                               :pitch        pitch
                               :duration     quant-duration
                               :voice        current-voice}))
          min-duration    (when min-duration
                            (min full-duration min-duration))]
      (log/debug (case event-type
                   :note
                   (format "%s plays at %s + %s for %s ms, at %.2f Hz."
                           id
                           current-marker
                           (int (:offset current-offset))
                           (int quant-duration)
                           pitch)
                   :rest
                   (format "%s rests at %s + %s for %s ms."
                           id
                           current-marker
                           (int (:offset current-offset))
                           (int full-duration))))
      {:instrument id
       :events     (case event-type
                     :note (if (pos? quant-duration) [note] [])
                     :rest [])
       :state      {:duration             (if (pos? cram-level)
                                            duration
                                            beats)
                    :duration-inside-cram (when (pos? cram-level)
                                            beats)
                    :last-offset          (if chord-mode
                                            last-offset
                                            current-offset)
                    :current-offset       (if chord-mode
                                            current-offset
                                            (offset+ current-offset
                                                     full-duration))
                    :min-duration         min-duration}})))

(defn add-note-or-rest
  [{:keys [beats-tally beats-tally-default instruments voice-instruments
           current-voice] :as score}
   {:keys [beats] :as event}]
  (if beats-tally
    (let [beats (or beats beats-tally-default)]
      (-> score
          (update :beats-tally + beats)
          (assoc :beats-tally-default beats)))
    (let [updates             (event-updates score event)
          events              (mapcat :events updates)
          inst-updates        (into {}
                                (map (fn [{:keys [instrument state]}]
                                       [instrument state])
                                     updates))
          instruments         (if current-voice
                                (voice-instruments current-voice)
                                instruments)
          updated-instruments (merge-with merge instruments inst-updates)]
      (-> score
          (#(if current-voice
              (merge-voice-instruments % current-voice updated-instruments)
              (merge-instruments % updated-instruments)))
          (add-events events)))))

(defmethod update-score :note
  [score note]
  (add-note-or-rest score note))

(defn note
  "Causes every instrument in :current-instruments to play a note at its
   :current-offset for the specified duration.

   If no duration is specified, the note is played for the instrument's own
   internal duration, which will be the duration last specified on a note or
   rest in that instrument's part."
  ([pitch-fn]
    (note pitch-fn nil false))
  ([pitch-fn x]
    ; x could be a duration or :slur
    (let [duration (when (map? x) x)
          slur?    (= x :slur)]
      (note pitch-fn duration slur?)))
  ([pitch-fn {:keys [beats ms slurred]} slur?]
     {:event-type :note
      :pitch-fn   pitch-fn
      :beats      beats
      :ms         ms
      :slur?      (or slur? slurred)}))

