(ns alda.sound
  (:require [alda.sound.midi :as    midi]
            [overtone.at-at  :refer (mk-pool now at)]
            [taoensso.timbre :as    log]
            [alda.util       :refer (check-for)]))

(def ^:dynamic *active-audio-types* #{})

(defn set-up?
  [x]
  (contains? *active-audio-types* x))

(defmulti set-up-audio-type!
  (fn [audio-type & [score]] audio-type))

(defmethod set-up-audio-type! :default
  [audio-type & [score]]
  (log/errorf "No implementation of set-up-audio-type! defined for type %s" 
              audio-type))

(defmethod set-up-audio-type! :midi
  [_ & [score]]
  (midi/open-midi-synth!))

(defn set-up!
  "Does any necessary setup for one or more audio types.
   e.g. for MIDI, create and open a MIDI synth."
  [audio-type & [score]]
  (if (coll? audio-type)
    (doall 
      (pmap #(set-up! % score) audio-type))
    (when-not (set-up? audio-type) 
      (set-up-audio-type! audio-type score)
      (alter-var-root #'*active-audio-types* conj audio-type))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti refresh-audio-type!
  (fn [audio-type & [score]] audio-type))

(defmethod refresh-audio-type! :default
  [audio-type & [score]]
  (log/errorf "No implementation of refresh-audio-type! defined for type %s" 
              audio-type))

(defmethod refresh-audio-type! :midi
  [_ & [score]]
  (midi/load-instruments! score))

(defn refresh!
  "Performs any actions that may be needed each time the `play!` function is
   called. e.g. for MIDI, load instruments into channels (this needs to be
   done every time `play!` is called because new instruments may have been
   added to the score between calls to `play!`, when using Alda live.)"
  [audio-type & [score]]
  (if (coll? audio-type)
    (doall 
      (pmap #(refresh! % score) audio-type))
    (when (set-up? audio-type) 
      (refresh-audio-type! audio-type score))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti tear-down-audio-type!
  (fn [audio-type & [score]] audio-type))

(defmethod tear-down-audio-type! :default
  [audio-type & [score]]
  (log/errorf "No implementation of tear-down! defined for type %s" audio-type))

(defmethod tear-down-audio-type! :midi
  [_ & [score]]
  (midi/close-midi-synth!))

(defn tear-down!
  "Does any necessary clean-up at the end.
   e.g. for MIDI, close the MIDI synth."
  [audio-type & [score]]
  (if (coll? audio-type)
    (doall 
      (pmap #(tear-down! % score) audio-type))
    (when (set-up? audio-type) 
      (tear-down-audio-type! audio-type score)
      (alter-var-root #'*active-audio-types* disj audio-type))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti play-event!
  "Plays a note/event, using the appropriate method based on the type of the
   instrument."
  (fn [event instrument]
    (-> instrument :config :type)))

(defmethod play-event! :default
  [_ instrument]
  (log/errorf "No implementation of play-event! defined for type %s"
              (-> instrument :config :type)))

(defmethod play-event! nil
  [event instrument]
  :do-nothing)

(defmethod play-event! :midi
  [note instrument]
  (midi/play-note! note))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- score-length
  "Calculates the length of a score in ms."
  [{:keys [events] :as score}]
  (if (and events (not (empty? events)))
    (letfn [(note-end [{:keys [offset duration] :as note}] (+ offset duration))]
      (apply max (map note-end events)))
    0))

(defn determine-audio-types
  [{:keys [instruments] :as score}]
  (set (for [[id {:keys [config]}] instruments]
         (:type config))))

(def ^:dynamic *play-opts* {})

; TODO: control where to start and stop playing using the start & end keys
(defn play!
  "Plays an Alda score, optionally from given start/end marks.
   
   Returns a function that, when called mid-playback, will stop any further
   events from playing."
  [{:keys [events instruments] :as score}]
  (let [{:keys [start end pre-buffer post-buffer one-off? async?]} *play-opts*
        audio-types (determine-audio-types score)
        _           (set-up! audio-types score)
        _           (refresh! audio-types score)
        pool        (mk-pool)
        playing?    (atom true)
        start       (+ (now) (or pre-buffer 0))]
    (doall (pmap (fn [{:keys [offset instrument] :as event}]
                   (let [instrument (-> instrument instruments)]
                     (at (+ start offset) 
                         #(when @playing? 
                            (play-event! event instrument)) 
                         pool)))
                 events))
    (when-not async?
      ; block until the score is done playing
      (Thread/sleep (+ (score-length score) (or post-buffer 0))))
    (when one-off? (tear-down! audio-types score))
    #(reset! playing? false)))

(defn make-wav!
  "Parses an input file and saves the resulting sound data as a wav file, 
   using the specified options."
  [input-file output-file {:keys [start end]}]
  (let [target-file (check-for output-file)]
    (comment "To do.")))
