(ns alda.lisp.events)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events...")

(defn note-length
  "Converts a number, representing a note type, e.g. 4 = quarter, 8 = eighth,
   into a number of beats. Handles dots if present."
  ([number]
    (/ 4 number))
  ([number {:keys [dots]}]
    (let [value (/ 4 number)]
      (loop [total value, factor 1/2, dots dots]
        (if (pos? dots)
          (recur (+ total (* value factor)) (* factor 1/2) (dec dots))
          total)))))

(defn duration
  "Combines a variable number of tied note-lengths into one.

   A slur may appear as the final argument of a duration, making the current
   note legato (effectively slurring it into the next).

   Returns a map containing a duration-fn, which gives the duration in ms when
   provide with a tempo, and whether or not the note is slurred."
  [& components]
  (let [[note-lengths slurred] (if (= (last components) :slur)
                                 (conj [(drop-last components)] true)
                                 (conj [components] false))
        beats (apply + note-lengths)]
    (set-duration beats)
    {:duration-fn (fn [tempo] (float (* beats (/ 60000 tempo))))
     :slurred slurred}))

(def ^:private intervals
  {:c 0, :d 2, :e 4, :f 5, :g 7, :a 9, :b 11})

(defn- midi-note
  "Given a letter and an octave, returns the MIDI note number.
   e.g. :c, 4  =>  60"
  [letter octave]
  (+ (intervals letter) (* octave 12) 12))

(defn- midi->hz
  "Converts a MIDI note number to the note's frequency in Hz."
  [note]
  (* 440.0 (Math/pow 2.0 (/ (- note 69.0) 12.0))))

(defn pitch
  "Returns a fn that will calculate the frequency in Hz, within the context
   of the octave that an instrument is in."
  [letter & accidentals]
  (fn [octave]
    (let [midi-note (reduce (fn [number accidental]
                              (case accidental
                                :flat  (dec number)
                                :sharp (inc number)))
                            (midi-note letter octave)
                            accidentals)]
      (midi->hz midi-note))))

(defrecord Note [offset instrument volume pitch duration])

(defn note*
  ([instrument pitch-fn]
   {:pre [(fn? pitch-fn)]}
    (note* instrument
           pitch-fn
           (duration (-> (*instruments* instrument) :duration))
           false))
  ([instrument pitch-fn arg3]
    (cond ; arg3 could be a duration or :slur
      (map? arg3)    (note* instrument
                            pitch-fn
                            arg3
                            false)
      (= :slur arg3) (note* instrument
                            pitch-fn
                            (duration (-> (*instruments* instrument)
                                          :duration))
                            true)))
  ([instrument pitch-fn {:keys [duration-fn slurred]} slur?]
    (let [get-attribute (fn [attr]
                          (fn []
                            (-> (*instruments* instrument) attr)))
          tempo          (get-attribute :tempo)
          inst-name      (get-attribute :name)
          volume         (get-attribute :volume)
          octave         (get-attribute :octave)
          current-offset (get-attribute :current-offset)
          current-marker (get-attribute :current-marker)
          quant          (if (or slur? slurred) 1.0 ((get-attribute :quantization)))
          note-duration  (duration-fn (tempo))
          event          (map->Note {:offset (current-offset)
                                     :instrument (inst-name)
                                     :volume (volume)
                                     :pitch (pitch-fn (octave))
                                     :duration (* note-duration quant)})]
      (add-event instrument event)
      (set-last-offset instrument (current-offset))
      (set-current-offset instrument (offset+ (current-offset) note-duration))
      (log/debug (format "%s plays at %s + %s for %s ms, at %.2f Hz."
                         instrument
                         (current-marker)
                         (int (:offset (:offset event)))
                         (int (:duration event))
                         (:pitch event)))
      event)))

(defmacro note
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (note* instrument# ~@args)))))

(defrecord Rest [offset instrument duration])

(defn pause*
  ([instrument]
    (pause* instrument (duration (-> (*instruments* instrument) :duration))))
  ([instrument {:keys [duration-fn] :as dur}]
    {:pre [(map? dur)]}
    (let [get-attribute (fn [attr]
                          (fn []
                            (-> (*instruments* instrument) attr)))
          current-offset (get-attribute :current-offset)
          current-marker (get-attribute :current-marker)
          last-offset    (get-attribute :last-offset)
          tempo          (get-attribute :tempo)
          rest-duration  (duration-fn (tempo))]
      (set-last-offset instrument (current-offset))
      (set-current-offset instrument (offset+ (current-offset) rest-duration))
      (let [rest (Rest. (last-offset) instrument rest-duration)]
        (log/debug (format "%s rests at %s + %s for %s ms."
                           instrument
                           (current-marker)
                           (int (:offset (last-offset)))
                           (int rest-duration)))
        rest))))

(defmacro pause
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (pause* instrument# ~@args)))))

(defrecord Chord [events])

(defmacro chord*
  "Chords contain notes/rests that all start at the same time/offset.
   The resulting *current-offset* is at the end of the shortest note/rest in
   the chord."
  [instrument & events]
  (let [num-of-events  (count (filter #(= (first %) 'note) events))
        start          (gensym "start")
        offsets        (gensym "offsets")
        current-offset (gensym "current-offset")
        current-marker (gensym "current-marker")]
    (list* 'let [current-offset `(fn [] (-> (*instruments* ~instrument)
                                            :current-offset))
                 current-marker `(fn [] (-> (*instruments* ~instrument)
                                            :current-marker))
                 start   (list current-offset)
                 offsets (list 'atom [])]
           (concat
             (interleave
               (repeat `(set-current-offset ~instrument ~start))
               events
               (repeat `(swap! ~offsets conj (~current-offset))))
             [`(set-last-offset ~instrument ~start)
              `(set-current-offset ~instrument (apply (partial min-key :offset)
                                                  (remove #(offset= % ~start)
                                                          (deref ~offsets))))
              `(let [chord#
                     (Chord. (take-last ~num-of-events
                                        (get-in *events*
                                                [(~current-marker) :events])))]
                 chord#)]))))

(defmacro chord
  [& args]
  `(doall
     (for [instrument# *current-instruments*]
       (binding [*current-instruments* #{instrument#}]
         (chord* instrument# ~@args)))))

(defn voice
  "Returns a list of the events, executing them in the process."
  [& events]
  (remove (fn [event]
            (not (every? #(contains? #{alda.lisp.Note alda.lisp.Chord} (type %))
                         event)))
          events))

(defn voice
  "Returns a list of the events, executing them in the process."
  [& events]
  (remove #(not (contains? #{alda.lisp.Note alda.lisp.Chord} (type %)))
          (flatten events)))

(defmacro voices*
  "Voices are chronological sequences of events that each start at the same
   time. The resulting :current-offset is at the end of the voice that finishes
   last."
  [instrument & voices]
  (let [voice-snapshots (gensym "voice-snapshots")
        voice-events    (gensym "voice-events")]
    (list* 'let ['start-snapshot `(snapshot ~instrument)
                 voice-snapshots (list 'atom {})
                 voice-events    (list 'atom {})]
           (concat
             (for [[_ num# & events# :as voice#] voices]
               (list 'let ['voice-name (keyword (str \v num#))]
                     `(load-snapshot ~instrument
                                     (get (deref ~voice-snapshots)
                                          ~'voice-name ~'start-snapshot))
                      (list 'swap! voice-events
                        (list 'fn ['m]
                          (list 'merge-with 'concat 'm
                                {'voice-name
                                 (list 'vec
                                       (list 'map 'first (vec events#)))})))
                     `(swap! ~voice-snapshots assoc ~'voice-name
                                              (snapshot ~instrument))))
             [`(let [last-voice#
                     (apply (partial max-key #(-> % :current-offset :offset))
                            (vals (deref ~voice-snapshots)))]
                 (load-snapshot ~instrument last-voice#))
              `(deref ~voice-events)]))))

(defmacro voices
  [& args]
  `(let [voices-per-instrument#
         (for [instrument# *current-instruments*]
           (binding [*current-instruments* #{instrument#}]
             (voices* instrument# ~@args)))]
     (apply merge-with (comp vec concat) voices-per-instrument#)))
