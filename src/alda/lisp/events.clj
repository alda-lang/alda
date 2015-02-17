(ns alda.lisp.events)
(in-ns 'alda.lisp)

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

   Returns a map containing the duration in ms (within the context of the
   current tempo) and whether or not the note is slurred."
  [& components]
  (let [[note-lengths slurred] (if (= (last components) :slur)
                                 (conj [(drop-last components)] true)
                                 (conj [components] false))
        beats (apply + note-lengths)]
    (set-duration beats)
    {:duration (* beats (/ 60000 *tempo*))
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
  "Determines the frequency in Hz, within the context of the current
   octave."
  [letter & accidentals]
  (let [midi-note (reduce (fn [number accidental]
                            (case accidental
                              :flat  (dec number)
                              :sharp (inc number)))
                          (midi-note letter *octave*)
                          accidentals)]
    (midi->hz midi-note)))

(defrecord Note [offset instruments volume pitch duration])

(defn note
  ([pitch]
   {:pre [(number? pitch)]}
    (note pitch (duration *duration*) false))
  ([pitch arg2] ; arg2 could be a duration or :slur
    (cond
      (map? arg2)    (note pitch arg2 false)
      (= :slur arg2) (note pitch (duration *duration*) true)))
  ([pitch {:keys [duration slurred]} slur?]
    (binding [*quant* (if (or slur? slurred)
                        1.0
                        *quant*)]
      (let [event (map->Note {:offset *current-offset*
                              :instruments *instruments*
                              :volume *volume*
                              :pitch pitch
                              :duration (* duration *quant*)})]
        (add-event event)
        (set-last-offset *current-offset*)
        (set-current-offset (+ *current-offset* duration))
        event))))

(defrecord Rest [offset duration])

(defn pause
  ([]
    (pause (duration *duration*)))
  ([{:keys [duration] :as dur}]
    {:pre [(map? dur)]}
    (set-last-offset *current-offset*)
    (set-current-offset (+ *current-offset* duration))
    (Rest. *last-offset* duration)))

(defrecord Chord [events])

(defmacro chord
  "Chords contain notes/rests that all start at the same time/offset.
   The resulting *current-offset* is at the end of the shortest note/rest in
   the chord."
  [& events]
  (let [num-of-events (count (filter #(= (first %) 'note) events))
        offsets (gensym "offsets")]
    (list* 'let ['start '*current-offset*
                 offsets (list 'atom [])]
           (concat
             (interleave
               (repeat `(set-current-offset ~'start))
               events
               (repeat `(swap! ~offsets conj *current-offset*)))
             [`(set-last-offset ~'start)
              `(set-current-offset (apply min (remove #(= % ~'start)
                                                      (deref ~offsets))))
              `(Chord. (take-last ~num-of-events
                                  (get-in *events* [*current-marker* :events])))]))))

(defn voice
  "Returns a list of the events, executing them in the process."
  [& events]
  (remove #(not (contains? #{alda.lisp.Note alda.lisp.Chord} (type %))) events))

(defmacro voices
  "Voices are chronological sequences of events that each start at the same time.
   The resulting *current-offset* is at the end of the voice that finishes last."
  [& voices]
  (let [voice-snapshots (gensym "voice-snapshots")
        voice-events    (gensym "voice-events")]
    (list* 'let ['start-snapshot `(snapshot)
                 voice-snapshots (list 'atom {})
                 voice-events    (list 'atom {})]
           (concat
             (for [[_ num# & events# :as voice#] voices]
               (list 'let ['voice-name (keyword (str \v num#))]
                     `(load-snapshot (get (deref ~voice-snapshots) ~'voice-name
                                                                   ~'start-snapshot))
                      (list 'swap! voice-events
                        (list 'fn ['m]
                          (list 'merge-with 'concat 'm {'voice-name (vec events#)})))
                     `(swap! ~voice-snapshots assoc ~'voice-name (snapshot))))
             [`(let [last-voice# (apply max-key #(get % (var *current-offset*))
                                                (vals (deref ~voice-snapshots)))]
                 (load-snapshot last-voice#))
              `(deref ~voice-events)]))))
