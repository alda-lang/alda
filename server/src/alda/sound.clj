(ns alda.sound
  (:require [alda.sound.midi :as    midi]
            [taoensso.timbre :as    log]
            [alda.util       :refer (check-for
                                     parse-time
                                     pdoseq-block
                                     parse-position)]
            [alda.lisp.model.records])
  (:import [com.softsynth.shared.time TimeStamp ScheduledCommand]
           [com.jsyn.engine SynthesisEngine]
           [alda.lisp.model.records Function]))

(defn new-audio-context
  []
  (atom
    {:audio-types      #{}
     :synthesis-engine (doto (SynthesisEngine.) .start)}))

(defn set-up?
  [audio-ctx audio-type]
  (contains? (:audio-types @audio-ctx) audio-type))

(defmulti set-up-audio-type!
  (fn [audio-ctx audio-type & [score]] audio-type))

(defmethod set-up-audio-type! :default
  [_ audio-type & [score]]
  (log/errorf "No implementation of set-up-audio-type! defined for type %s"
              audio-type))

(defmethod set-up-audio-type! :midi
  [audio-ctx _ & [score]]
  (midi/open-midi-synth! audio-ctx))

(defn set-up!
  "Does any necessary setup for one or more audio types.
   e.g. for MIDI, create and open a MIDI synth."
  [audio-ctx audio-type & [score]]
  (if (coll? audio-type)
    (pdoseq-block [a-t audio-type]
      (set-up! audio-ctx a-t score))
    (when-not (set-up? audio-ctx audio-type)
      (set-up-audio-type! audio-ctx audio-type score)
      (swap! audio-ctx update :audio-types conj audio-type))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti refresh-audio-type!
  (fn [audio-ctx audio-type & [score]] audio-type))

(defmethod refresh-audio-type! :default
  [_ audio-type & [score]]
  (log/errorf "No implementation of refresh-audio-type! defined for type %s"
              audio-type))

(defmethod refresh-audio-type! :midi
  [audio-ctx _ & [score]]
  (midi/load-instruments! audio-ctx score))

(defn refresh!
  "Performs any actions that may be needed each time the `play!` function is
   called. e.g. for MIDI, load instruments into channels (this needs to be
   done every time `play!` is called because new instruments may have been
   added to the score between calls to `play!`, when using Alda live.)"
  [audio-ctx audio-type & [score]]
  (if (coll? audio-type)
    (pdoseq-block [a-t audio-type]
      (refresh! audio-ctx a-t score))
    (when (set-up? audio-ctx audio-type)
      (refresh-audio-type! audio-ctx audio-type score))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti tear-down-audio-type!
  (fn [audio-ctx audio-type & [score]] audio-type))

(defmethod tear-down-audio-type! :default
  [_ audio-type & [score]]
  (log/errorf "No implementation of tear-down! defined for type %s" audio-type))

(defmethod tear-down-audio-type! :midi
  [audio-ctx _ & [score]]
  (midi/close-midi-synth! audio-ctx))

(defn tear-down!
  "Does any necessary clean-up at the end.
   e.g. for MIDI, close the MIDI synth."
  [audio-ctx audio-type & [score]]
  (if (coll? audio-type)
    (pdoseq-block [a-t audio-type]
      (tear-down! audio-ctx a-t score))
    (when (set-up? audio-ctx audio-type)
      (tear-down-audio-type! audio-ctx audio-type score)
      (swap! audio-ctx update :audio-types disj audio-type))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti start-event!
  "Kicks off a note/event, using the appropriate method based on the type of the
   instrument."
  (fn [audio-ctx event instrument]
    (-> instrument :config :type)))

(defmethod start-event! :default
  [_ _ instrument]
  (log/errorf "No implementation of start-event! defined for type %s"
              (-> instrument :config :type)))

(defmethod start-event! nil
  [_ _ _]
  :do-nothing)

(defmethod start-event! :midi
  [audio-ctx note _]
  (midi/play-note! audio-ctx note))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti stop-event!
  "Ends a note/event, using the appropriate method based on the type of the
   instrument."
  (fn [audio-ctx event instrument]
    (-> instrument :config :type)))

(defmethod stop-event! :default
  [_ _ instrument]
  (log/errorf "No implementation of start-event! defined for type %s"
              (-> instrument :config :type)))

(defmethod stop-event! nil
  [_ _ _]
  :do-nothing)

(defmethod stop-event! :midi
  [audio-ctx note _]
  (midi/stop-note! audio-ctx note))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- score-length
  "Given an event set from a score, calculates the length of the score in ms."
  [event-set]
  (let [events   (filter :duration event-set)
        note-end (fn [{:keys [offset duration] :as note}]
                   (+ offset duration))]
    (if (and events (not (empty? events)))
      (apply max (map note-end events))
      0)))

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
          (throw (Exception.
                   (str "Do not support " (type pos) " as a play time."))))))

(defn start-finish-times [{:keys [from to]} markers]
  (map (partial lookup-time markers) [from to]))

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

(defn schedule-event!
  [^SynthesisEngine engine offset f]
  (let [ts  (TimeStamp. offset)
        cmd (proxy [ScheduledCommand] [] (run [] (f)))]
    (.scheduleCommand engine ts cmd)))

(defn play!
  "Plays an Alda score, optionally from given start/end marks determined by
   *play-opts*.

   Optionally takes as a second argument a set of events to play (which could
   be pre-filtered, e.g. for playing only a portion of the score).

   In either case, the offsets of the events to be played are shifted back such
   that the earliest event's offset is 0 -- this is so that playback will start
   immediately.

   Returns a function that, when called mid-playback, will stop any further
   events from playing."
  [{:keys [instruments audio-context] :as score} & [event-set]]
  (let [{:keys [pre-buffer post-buffer one-off? async?]} *play-opts*
        audio-types (determine-audio-types score)
        audio-ctx   (or audio-context (new-audio-context))
        _           (set-up! audio-ctx audio-types score)
        _           (refresh! audio-ctx audio-types score)
        engine      (:synthesis-engine @audio-ctx)
        playing?    (atom true)
        events      (if event-set
                      (let [earliest (->> (map :offset event-set)
                                          (apply min Long/MAX_VALUE)
                                          (max 0))]
                        (shift-events event-set earliest nil))
                      (let [event-set   (:events score)
                            markers     (:markers score)
                            [start end] (start-finish-times *play-opts*
                                                            markers)]
                        (shift-events event-set start end)))
        begin       (+ (.getCurrentTime ^SynthesisEngine engine)
                       (or pre-buffer 0))]
    (pdoseq-block [{:keys [offset instrument duration] :as event} events]
      (let [inst   (-> instrument instruments)
            start! #(when @playing?
                      (if (instance? Function event)
                        ((:function event))
                        (start-event! audio-ctx event inst)))
            stop!  #(when-not (instance? Function event)
                      (stop-event! audio-ctx event inst))]
        (schedule-event! engine (+ begin
                                   (/ offset 1000.0)) start!)
        (schedule-event! engine (+ begin
                                   (/ offset 1000.0)
                                   (/ duration 1000.0)) stop!)))
    (when-not async?
      ; block until the score is done playing
      ; TODO: find a way to handle this that doesn't involve Thread/sleep
      (Thread/sleep (+ (score-length events)
                       (or pre-buffer 0)
                       (or post-buffer 0))))
    (when one-off? (tear-down! audio-ctx audio-types score))
    #(reset! playing? false)))

(defn make-wav!
  "Parses an input file and saves the resulting sound data as a wav file,
   using the specified options."
  [input-file output-file {:keys [start end]}]
  (let [target-file (check-for output-file)]
    ;; TODO
    nil))
