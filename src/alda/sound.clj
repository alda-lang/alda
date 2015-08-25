(ns alda.sound
  (:require [alda.sound.midi :as    midi]
            [overtone.at-at  :refer (mk-pool now at)]
            [taoensso.timbre :as    log]
            [alda.util       :refer (check-for)]
            [defun           :refer (defun)]))

(defun set-up!
  "Does any necessary setup for a given audio type.
   e.g. for MIDI, create a MIDI synth and load the instruments."
  ([(audio-types :guard coll?) score]
    (doseq [audio-type audio-types]
      (set-up! audio-type score)))
  ([audio-type score]
    (case audio-type
      :midi (do
              (midi/open-midi-synth!)
              (midi/load-instruments! score))
      :catch-all---no-action-needed)))

(defn tear-down!
  "Does any necessary clean-up at the end.
   e.g. for MIDI, close the MIDI synth."
  [audio-type score]
  (case audio-type
    :midi (midi/close-midi-synth!)
    :catch-all---no-action-needed))

(defmulti play-event!
  "Plays a note/event, using the appropriate method based on the type of the
   instrument."
  (fn [event instrument]
    (-> instrument :config :type)))

(defmethod play-event! :default
  [event instrument]
  (log/errorf "No implementation of play-event! defined for type %s"
              (-> instrument :config :type)))

(defmethod play-event! nil
  [event instrument]
  :do-nothing)

(defmethod play-event! :midi
  [note instrument]
  (midi/play-note! note))

(defn- score-length
  "Calculates the length of a score in ms."
  [{:keys [events] :as score}]
  (if (and events (not (empty? events)))
    (letfn [(note-end [{:keys [offset duration] :as note}] (+ offset duration))]
      (apply max (map note-end events)))
    0))

(defn determine-audio-types
  [score]
  ; TODO: actually determine audio types
  [:midi])

; TODO: control where to start and stop playing using the start & end keys
(defn play!
  "Plays an Alda score, optionally from given start/end marks."
  [{:keys [events instruments] :as score}
   & [{:keys [start end pre-buffer post-buffer one-off?] :as opts}]]
  (let [audio-types (determine-audio-types score)]
    (when one-off? (set-up! audio-types score))
    (let [start (+ (now) (or pre-buffer 0))
          pool  (mk-pool)]
      (doall (pmap (fn [{:keys [offset instrument] :as event}]
                     (let [instrument (-> instrument instruments)]
                       (at (+ start offset) #(play-event! event instrument) pool)))
                   events))
      (Thread/sleep (+ (score-length score) (or post-buffer 0))))
    (when one-off? (tear-down! audio-types score))))

(defn make-wav!
  "Parses an input file and saves the resulting sound data as a wav file, using the
   specified options."
  [input-file output-file {:keys [start end]}]
  (let [target-file (check-for output-file)]
    (comment "To do.")))
