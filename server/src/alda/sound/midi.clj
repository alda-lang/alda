(ns alda.sound.midi
  (:require [taoensso.timbre :as log])
  (:import (java.util.concurrent LinkedBlockingQueue)
           (javax.sound.midi MidiSystem Synthesizer MidiChannel)))

(comment
  "There are 16 channels per MIDI synth (1-16);
   channel 10 is reserved for percussion.")

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(comment
  "It takes a second to initialize a MIDI synth. To avoid hiccups and make
   playback more immediate, we maintain a handful of pre-initialized MIDI
   synths, ready for immediate use.")

(defn new-midi-synth
  []
  (doto ^Synthesizer (MidiSystem/getSynthesizer) .open))

(def ^:dynamic *midi-synth-pool* (LinkedBlockingQueue.))

(def ^:const MIDI-SYNTH-POOL-SIZE 4)

(defn fill-midi-synth-pool!
  []
  (dotimes [_ (- MIDI-SYNTH-POOL-SIZE (count *midi-synth-pool*))]
    (future (.add *midi-synth-pool* (new-midi-synth)))))

(defn drain-excess-midi-synths!
  []
  (dotimes [_ (- (count *midi-synth-pool*) MIDI-SYNTH-POOL-SIZE)]
    (future (.close (.take *midi-synth-pool*)))))

(defn midi-synth-available?
  []
  (pos? (count *midi-synth-pool*)))

(defn get-midi-synth
  []
  (fill-midi-synth-pool!)
  (drain-excess-midi-synths!)
  (.take *midi-synth-pool*))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

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

(defn get-midi-synth!
  "If there isn't already a :midi-synth in the audio context, grabs one from
   the pool."
  [audio-ctx]
  (when-not (:midi-synth @audio-ctx)
    (swap! audio-ctx assoc :midi-synth (get-midi-synth))))

(defn close-midi-synth!
  "Closes the MIDI synth in the audio context."
  [audio-ctx]
  (.close ^Synthesizer (:midi-synth @audio-ctx)))

(defn protection-key-for
  [{:keys [offset duration midi-note]}]
  [midi-note (+ offset duration)])

(defn protect-note!
  "Makes a note in the audio context that this note is playing.

   This prevents other notes that have the same MIDI note number from stopping
   this note."
  [audio-ctx note]
  (let [[midi-note offset] (protection-key-for note)]
    (swap! audio-ctx
           update-in [:protected-notes midi-note]
           (fnil conj #{}) offset)))

(defn unprotect-note!
  "Removes protection from this note so that it can be stopped."
  [audio-ctx note]
  (let [[midi-note offset] (protection-key-for note)]
    (swap! audio-ctx
           update-in [:protected-notes midi-note]
           disj offset)))

(defn note-reserved?
  "Returns true if there is ANOTHER note with the same MIDI note number that is
   currently playing. If this is the case, then we will NOT stop the note, and
   instead wait for the other note to stop it."
  [audio-ctx note]
  (let [{:keys [protected-notes]} @audio-ctx
        [midi-note offset]        (protection-key-for note)]
    (boolean (some (partial not= offset) (get protected-notes midi-note)))))

(defn play-note!
  [audio-ctx {:keys [midi-note instrument volume track-volume panning]
              :as note}]
  (protect-note! audio-ctx note)
  (let [{:keys [midi-synth midi-channels]} @audio-ctx
        channels       (.getChannels ^Synthesizer midi-synth)
        channel-number (-> instrument midi-channels :channel)
        channel        (aget channels channel-number)]
    (.controlChange ^MidiChannel channel 7 (* 127 track-volume))
    (.controlChange ^MidiChannel channel 10 (* 127 panning))
    (log/debugf "Playing note %s on channel %s." midi-note channel-number)
    (.noteOn ^MidiChannel channel midi-note (* 127 volume))))

(defn stop-note!
  [audio-ctx {:keys [midi-note instrument] :as note}]
  (unprotect-note! audio-ctx note)
  (when-not (note-reserved? audio-ctx note)
    (let [{:keys [midi-synth midi-channels]} @audio-ctx
          channels       (.getChannels ^Synthesizer midi-synth)
          channel-number (-> instrument midi-channels :channel)
          channel        (aget channels channel-number)]
      (log/debug "MIDI note off:" midi-note)
      (.noteOff ^MidiChannel channel midi-note))))
