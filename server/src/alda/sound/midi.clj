(ns alda.sound.midi
  (:require [taoensso.timbre :as log])
  (:import  (javax.sound.midi MidiSystem)))

; note: there are 16 channels (1-16), and channel 10 is reserved for percussion

(declare ^:dynamic *midi-synth*)
(declare ^:dynamic *midi-channels*)

(defn- next-available
  "Given a set of available MIDI channels, returns the next available one,
   bearing in mind that channel 10 can only be used for percussion.

   Returns nil if no channels available."
  [channels & {:keys [percussion?]}]
  (first (filter (partial (if percussion? = not=) 9) channels)))

(defn ids->channels
  "Inspects a score and generates a map of instrument IDs to MIDI channels.
   The channel values are maps with keys :channel (the channel number) and
   :patch (the General MIDI patch number)."
  [{:keys [instruments] :as score}]
  (let [channels (atom (apply sorted-set (range 16)))]
    (reduce (fn [result id]
              (let [{:keys [patch percussion?]} (-> id instruments :config)
                    channel (if-let [channel
                                     (next-available @channels
                                                     :percussion? percussion?)]
                              (do
                                (swap! channels disj channel)
                                channel)
                              (throw
                                (Exception. "Ran out of MIDI channels! :(")))]
                (assoc result id {:channel channel
                                  :patch patch
                                  :percussion? percussion?})))
            {}
            (for [[id {:keys [config]}] instruments
                  :when (= :midi (:type config))]
              id))))

(defn- load-instrument! [patch-number channel]
  (.programChange channel (dec patch-number)))

(defn load-instruments! [score]
  (alter-var-root #'*midi-channels* (constantly (ids->channels score)))
  (doseq [{:keys [channel patch]} (set (vals *midi-channels*))
          :when patch]
    (load-instrument! patch (aget (.getChannels *midi-synth*) channel))))

(defn open-midi-synth!
  "Loads a new MIDI synth into *midi-synth*, and opens it."
  []
  (log/debug "Loading MIDI synth...")
  (alter-var-root #'*midi-synth*
                  (constantly (doto (MidiSystem/getSynthesizer) .open)))
  (log/debug "Done loading MIDI synth."))

(defn close-midi-synth!
  "Closes the MIDI synth referred to by *midi-synth*."
  []
  (.close *midi-synth*))

(defn play-note!
  [{:keys [midi-note instrument volume track-volume panning]}]
  (let [channel-number (-> instrument *midi-channels* :channel)
        channel (aget (.getChannels *midi-synth*) channel-number)]
    (.controlChange channel 7 (* 127 track-volume))
    (.controlChange channel 10 (* 127 panning))
    (log/debugf "Playing note %s on channel %s." midi-note channel-number)
    (.noteOn channel midi-note (* 127 volume))))

(defn stop-note!
  [{:keys [midi-note instrument]}]
  (let [channel-number (-> instrument *midi-channels* :channel)
        channel (aget (.getChannels *midi-synth*) channel-number)]
    (log/debug "MIDI note off:" midi-note)
    (.noteOff channel midi-note)))
