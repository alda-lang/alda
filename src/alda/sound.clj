(ns alda.sound
  (:require [alda.sound.midi :as    midi]
            [overtone.at-at  :refer (mk-pool now at)]
            [taoensso.timbre :as    log]
            [alda.util       :refer [check-for parse-time pdoseq-block parse-position]]))

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
    (pdoseq-block [a-t audio-type]
      (set-up! a-t score))
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
    (pdoseq-block [a-t audio-type]
      (refresh! a-t score))
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
    (pdoseq-block [a-t audio-type]
      (tear-down! a-t score))
    (when (set-up? audio-type)
      (tear-down-audio-type! audio-type score)
      (alter-var-root #'*active-audio-types* disj audio-type))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn shift-events
  [events offset cut-off]
  (let [offset  (or offset 0)
        cut-off (when cut-off (- cut-off offset))
        keep?   (if cut-off
                  #(and (<= 0 %) (> cut-off %))
                  #(<= 0 %))]
    (sequence (comp (map #(update-in % [:offset] - offset))
                    (filter (comp keep? :offset)))
              events)))

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

(defmacro with-play-opts
  "Apply `opts` as overrides to *play-opts* when executing `body`"
  [opts & body]
  `(binding [*play-opts* (merge *play-opts* ~opts)]
     ~@body))

(defn- lookup-time [markers pos]
  (let [pos (if (string? pos)
              (parse-position pos)
              pos)]
    (cond (keyword? pos)
          (or (markers (name pos))
              (throw (Exception. (str "Marker " pos " not found."))))

          (or (number? pos) (nil? pos))
          pos

          :else
          (throw (Exception. (str "Do not support " (type pos) " as a play time."))))))

(defn start-finish-times [{:keys [from to]} markers]
  (map (partial lookup-time markers) [from to]))

(defn play!
  "Plays an Alda score, optionally from given start/end marks.

   Returns a function that, when called mid-playback, will stop any further
   events from playing."
  [{:keys [events markers instruments] :as score}]
  (let [{:keys [pre-buffer post-buffer one-off? async?]} *play-opts*
        audio-types (determine-audio-types score)
        _           (set-up! audio-types score)
        _           (refresh! audio-types score)
        pool        (mk-pool)
        playing?    (atom true)
        begin       (+ (now) (or pre-buffer 0))
        [start end] (start-finish-times *play-opts* markers)
        events      (shift-events events start end)
        duration    (- (or end (score-length score))
                       (or start 0))]
    (pdoseq-block [{:keys [offset instrument] :as event} events
                   :let [inst (-> instrument instruments)]]
      (at (+ begin offset)
          #(when @playing?
             (play-event! event inst))
          pool))

    (when-not async?
      ; block until the score is done playing
      (Thread/sleep (+ duration
                       (or pre-buffer 0)
                       (or post-buffer 0))))
    (when one-off? (tear-down! audio-types score))
    #(reset! playing? false)))

(defn make-wav!
  "Parses an input file and saves the resulting sound data as a wav file,
   using the specified options."
  [input-file output-file {:keys [start end]}]
  (let [target-file (check-for output-file)]
    ;; TODO
    nil))
