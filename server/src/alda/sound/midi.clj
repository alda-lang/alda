(ns alda.sound.midi
  (:require [taoensso.timbre :as log])
  (:import  (javax.sound.midi MidiSystem Synthesizer MidiChannel)))

(comment
  "There are 16 channels per MIDI synth (1-16);
   channel 10 is reserved for percussion.")

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

(defn- load-instrument! [patch-number ^MidiChannel channel]
  (.programChange channel (dec patch-number)))

(defn load-instruments!
  [audio-ctx score]
  (let [midi-channels (ids->channels score)]
    (swap! audio-ctx assoc :midi-channels midi-channels)
    (doseq [{:keys [channel patch]} (set (vals midi-channels))
            :when patch
            :let [synth    (:midi-synth @audio-ctx)
                  channels (.getChannels ^Synthesizer synth)]]
      (load-instrument! patch (aget channels channel)))))

(defn open-midi-synth!
  "Loads a new MIDI synth into :midi-synth, and opens it."
  [audio-ctx]
  (log/debug "Loading MIDI synth...")
  (let [synth (doto ^Synthesizer (MidiSystem/getSynthesizer) .open)]
    (swap! audio-ctx assoc :midi-synth synth))
  (log/debug "Done loading MIDI synth."))

(defn close-midi-synth!
  "Closes the MIDI synth referred to by *midi-synth*."
  [audio-ctx]
  (.close ^Synthesizer (:midi-synth @audio-ctx)))

(defn play-note!
  [audio-ctx {:keys [midi-note instrument volume track-volume panning]}]
  (let [{:keys [midi-synth midi-channels]} @audio-ctx
        channels       (.getChannels ^Synthesizer midi-synth)
        channel-number (-> instrument midi-channels :channel)
        channel        (aget channels channel-number)]
    (.controlChange ^MidiChannel channel 7 (* 127 track-volume))
    (.controlChange ^MidiChannel channel 10 (* 127 panning))
    (log/debugf "Playing note %s on channel %s." midi-note channel-number)
    (.noteOn ^MidiChannel channel midi-note (* 127 volume))))

(defn stop-note!
  [audio-ctx {:keys [midi-note instrument]}]
  (let [{:keys [midi-synth midi-channels]} @audio-ctx
        channels       (.getChannels ^Synthesizer midi-synth)
        channel-number (-> instrument midi-channels :channel)
        channel        (aget channels channel-number)]
    (log/debug "MIDI note off:" midi-note)
    (.noteOff ^MidiChannel channel midi-note)))
