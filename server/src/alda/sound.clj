(ns alda.sound
  (:require [alda.sound.midi :as    midi]
            [alda.util       :refer (parse-time
                                     pdoseq-block
                                     parse-position)]
            [taoensso.timbre :as    log])
  (:import [com.softsynth.shared.time TimeStamp ScheduledCommand]
           [com.jsyn.engine SynthesisEngine]))

(def ^:dynamic *synthesis-engine* nil)

(defn new-synthesis-engine
  []
  (doto (SynthesisEngine.) .start))

(defn start-synthesis-engine!
  []
  (alter-var-root #'*synthesis-engine* (constantly (new-synthesis-engine))))

(defn new-audio-context
  []
  (atom
    {:audio-types      #{}
     :synthesis-engine (or *synthesis-engine* (new-synthesis-engine))}))

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
  (log/debug "Setting up MIDI...")
  (midi/get-midi-synth! audio-ctx))

(declare determine-audio-types)

(defn set-up!
  "Does any necessary setup for one or more audio types.
   e.g. for MIDI, create and open a MIDI synth."
  ([{:keys [audio-context] :as score}]
   (let [audio-types (determine-audio-types score)]
     (set-up! audio-context audio-types score)))
  ([audio-ctx audio-type & [score]]
   (if (coll? audio-type)
     (pdoseq-block [a-t audio-type]
                   (set-up! audio-ctx a-t score))
     (when-not (set-up? audio-ctx audio-type)
       (set-up-audio-type! audio-ctx audio-type score)
       (swap! audio-ctx update :audio-types conj audio-type)))))

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
  ([{:keys [audio-context] :as score}]
   (let [audio-types (determine-audio-types score)]
     (tear-down! audio-context audio-types score)))
  ([audio-ctx audio-type & [score]]
   (if (coll? audio-type)
     (pdoseq-block [a-t audio-type]
                   (tear-down! audio-ctx a-t score))
     (when (set-up? audio-ctx audio-type)
       (tear-down-audio-type! audio-ctx audio-type score)
       (swap! audio-ctx update :audio-types disj audio-type)))))

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
    (->> (sequence (comp (map #(update-in % [:offset] - offset))
                         (filter (comp keep? :offset)))
                   events)
         (sort-by :offset))))

(defn schedule-event!
  [^SynthesisEngine engine offset f]
  (let [ts  (TimeStamp. offset)
        cmd (proxy [ScheduledCommand] [] (run [] (f)))]
    (.scheduleCommand engine ts cmd)))

(defn schedule-events!
  [events score audio-ctx playing? wait]
  (let [{:keys [instruments]} score
        engine (:synthesis-engine @audio-ctx)
        begin  (.getCurrentTime ^SynthesisEngine engine)
        ; bug? this could cause infinite blocking if @playing? is false
        end!   #(when @playing? (deliver wait :done))]
    (pdoseq-block [{:keys [offset instrument duration] :as event} events]
      (let [inst   (-> instrument instruments)
            start! #(when @playing?
                      (if-let [f (:function event)]
                        (future (f))
                        (start-event! audio-ctx event inst)))
            stop!  #(when-not (:function event)
                      (stop-event! audio-ctx event inst))]
        (schedule-event! engine (+ begin
                                   (/ offset 1000.0)) start!)
        (when-not (:function event)
          (schedule-event! engine (+ begin
                                     (/ offset 1000.0)
                                     (/ duration 1000.0)) stop!))))
    (schedule-event! engine (+ begin
                               (/ (score-length events) 1000.0)
                               1) end!)))

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
  [score & [event-set]]
  (let [{:keys [one-off? async?]} *play-opts*
        _           (log/debug "Determining audio types...")
        audio-types (determine-audio-types score)
        audio-ctx   (or (:audio-context score) (new-audio-context))
        _           (log/debug "Setting up audio types...")
        _           (set-up! audio-ctx audio-types score)
        _           (refresh! audio-ctx audio-types score)
        playing?    (atom true)
        wait        (promise)
        _           (log/debug "Determining events to schedule...")
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
        clean-up    #(tear-down! audio-ctx audio-types score)]
    (log/debug "Scheduling events...")
    (schedule-events! events score audio-ctx playing? wait)
    (cond
      (and one-off? async?)       (future @wait (clean-up))
      (and one-off? (not async?)) (do @wait (clean-up))
      (not async?)                @wait)
    #(reset! playing? false)))

