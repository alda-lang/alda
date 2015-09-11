(ns alda.sound.midi
  (:require [taoensso.timbre :as log]
            [midi.soundfont  :refer (load-all-instruments!)])
  (:import  (javax.sound.midi MidiSystem Synthesizer)))

; TODO: work around the limitation of 16 MIDI channels?
; TODO: enable percussion

; note: there are 16 channels (1-16), channel 10 is reserved for percussion,
;       channel 11 can be used for percussion too (or non-percussion)

(declare ^:dynamic *midi-synth*)
(declare ^:dynamic *midi-channels*)
(def ^:dynamic *midi-soundfont* nil)

(defn- next-available
  "Given a set of available MIDI channels, returns the next available one,
   bearing in mind that channel 10 can only be used for percussion, and
   channel 11 can be used for either percussion or non-percussion.

   Returns nil if no channels available."
  [channels & {:keys [percussion?]}]
  (if percussion?
    (first (filter #((set 10 11) %) channels))
    (first (filter (partial not= 10) channels))))

(defn ids->channels
  "Inspects a score and generates a map of instrument IDs to MIDI channels.
   The channel values are maps with keys :channel (the channel number) and
   :patch (the General MIDI patch number)."
  [{:keys [instruments] :as score}]
  (let [channels (atom (apply sorted-set (concat (range 0 9) (range 10 16))))]
    (reduce (fn [result id]
              (let [patch   (-> id instruments :config :patch)
                    ; TODO: pass ":percussion? true" if percussion 
                    channel (if-let [channel (next-available @channels)]
                              (do
                                (swap! channels disj channel)
                                channel)
                              (throw 
                                (Exception. "Ran out of MIDI channels! :(")))]
                (assoc result id {:channel channel
                                  :patch patch})))
            {}
            (for [[id {:keys [config]}] instruments
                  :when (= :midi (:type config))]
              id))))

(defn- load-instrument! [patch-number channel]
  (.programChange channel (dec patch-number)))

(defn load-instruments! [score]
  (alter-var-root #'*midi-channels* (constantly (ids->channels score)))
  (doseq [{:keys [channel patch]} (set (vals *midi-channels*))]
    (load-instrument! patch (aget (.getChannels *midi-synth*) channel))))

(defn open-midi-synth!
  "Loads a new MIDI synth into *midi-synth*, and opens it."
  []
  (log/debug "Loading MIDI synth...")
  (alter-var-root #'*midi-synth*
                  (constantly (doto (MidiSystem/getSynthesizer) .open)))
  (when *midi-soundfont* 
    (log/debug "Loading MIDI soundfont...")
    (load-all-instruments! *midi-synth* *midi-soundfont*)
    (log/debug "Done loading MIDI soundfont."))
  (log/debug "Done loading MIDI synth."))

(defn close-midi-synth!
  "Closes the MIDI synth referred to by *midi-synth*."
  []
  (.close *midi-synth*))

(defn- fraction->bend
  "Convert fractional offset from note to pitch-wheel bend, assuming that it is
   centred at 8192 and set to 2 semitones."
  [fraction]
  (let [range (if (pos? fraction) 8192 8191)]
    (int (+ 8192 (* range fraction 0.5)))))

(defn play-note! [{:keys [midi-note instrument duration volume track-volume]}]
  (let [channel-number (-> instrument *midi-channels* :channel)
        channel (aget (.getChannels *midi-synth*) channel-number)
        pure-note (if (integer? midi-note) midi-note (Math/round midi-note))
        bend (fraction->bend (- midi-note pure-note))]
    (.controlChange channel 7 (* 127 track-volume))
    (log/debugf "Playing note %s on channel %s." midi-note channel-number)
    (.setPitchBend channel bend)
    (future (Thread/sleep 10)
            (.setPitchBend channel bend))
    (.noteOn channel pure-note (* 127 volume))
    (Thread/sleep duration)
    (log/debug "MIDI note off:" midi-note)
    (.noteOff channel pure-note)
    (.setPitchBend channel 8192)))

(comment
  (defn- test-note! [note]
    (play-note!
     {:midi-note    note
      :instrument   "piano-YUbfY"
      :duration     300.0
      :volume       1.0
      :track-volume 0.75}))

  (defn- test-notes! [notes]
    (doseq [note notes]
      (future (test-note! note))
      (Thread/sleep 150)))

  (test-notes! (range 80 82 0.05))

  (test-notes! (shuffle (range 80 82 0.05))))
