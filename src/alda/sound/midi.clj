(ns alda.sound.midi
  (:require [overtone.at-at :refer (mk-pool now at)])
  (:import  (javax.sound.midi MidiSystem Synthesizer)))

; TODO: use different MIDI instruments depending on the instruments defined in
; the Alda score

(defn log [base x]
  (/ (Math/log x) (Math/log base)))

;;; http://en.wikipedia.org/wiki/MIDI_Tuning_Standard#Frequency_values
(defn frequency->note
  [f]
  (int (+ 69 (* 12 (log 2 (/ f 440))))))

(defn ->midi-notes
  [evald]
  (for [e (map #(select-keys % [:offset :pitch :duration]) (:events evald))]
    (-> e
        (assoc :note (frequency->note (:pitch e)))
        (dissoc :pitch))))

(defn play! [alda-events score-length & [{:keys [lead-time] :as opts}]]
  (with-open [synth (doto (MidiSystem/getSynthesizer) .open)]
    (let [channel (aget (.getChannels synth) 0)
          lead-time (or lead-time 1000)
          start   (+ (now) lead-time 1000)
          pool    (mk-pool)]
      (doseq [note (->midi-notes alda-events)]
        (at (+ start (:offset note))
            (fn []
              (.noteOn channel (:note note) 127)
              (Thread/sleep (:duration note))
              (.noteOff channel (:note note)))
            pool))
      (Thread/sleep (+ score-length 5000)))))
