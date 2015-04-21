(ns alda.sound.midi
  (:require [alda.sound.util :refer (score-length)]
            [overtone.at-at  :refer (mk-pool now at)]
            [taoensso.timbre :as    log])
  (:import  (javax.sound.midi MidiSystem Synthesizer)))

; TODO: do something with the volume values (convert to MIDI velocity? volume?)
; TODO: work around the limitation of 16 MIDI channels?
; TODO: enable percussion

; note: there are 16 channels (1-16), channel 10 is reserved for percussion,
;       channel 11 can be used for percussion too

(defn log [base x]
  (/ (Math/log x) (Math/log base)))

;;; http://en.wikipedia.org/wiki/MIDI_Tuning_Standard#Frequency_values
(defn frequency->note
  [f]
  (int (+ 69 (* 12 (log 2 (/ f 440))))))

(defn midi-event
  [event]
  (-> event
      (assoc :note (frequency->note (:pitch event)))
      (dissoc :pitch)))

(defn midi-channels
  "Returns a map of patch (instrument) numbers to their respective events."
  [{:keys [events instruments] :as score}]
  (into {}
    (for [[id events] (group-by :instrument (map midi-event events))
          :let [patch-number (-> id instruments :config :patch)]]
      [patch-number events])))

(defn load-instrument! [patch-number synth channel]
  (let [instruments (.. synth getDefaultSoundbank getInstruments)
        instrument  (nth instruments (dec patch-number))]
    (.loadInstrument synth instrument)
    (.programChange channel (dec patch-number))
    instrument))

(defn new-midi-synth []
  (log/debug "Loading MIDI synth...")
  (let [synth (doto (MidiSystem/getSynthesizer) .open)]
    (log/debug "Done loading MIDI synth.")
    synth))

(defn play! [score & [{:keys [pre-buffer post-buffer synth one-off?] :as opts}]]
  (let [synth    (or synth (new-midi-synth))
        channels (atom -1)
        start    (+ (now) (or pre-buffer 0))
        pool     (mk-pool)]
      (doseq [[patch events] (midi-channels score)]
        (let [channel (try
                        (->> (swap! channels inc)
                             ; skip channel 10 (percussion)
                             (#(if (= 10 %) (swap! channels inc) %))
                             (aget (.getChannels synth)))
                        (catch ArrayIndexOutOfBoundsException e
                          (throw (Exception. "Ran out of MIDI channels! :("))))]
          (load-instrument! patch synth channel)
          (doseq [{:keys [offset note duration]} events]
            (at (+ start offset)
                (fn []
                  (.noteOn channel note 127)
                  (Thread/sleep duration)
                  (.noteOff channel note))
                pool))))
      (Thread/sleep (try
                      (+ (score-length score) (or post-buffer 0))
                      (catch NullPointerException e 0)))
      (when one-off? (.close synth))))
