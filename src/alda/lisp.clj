(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated using the functions in this namespace.

   A number of these functions depend on the context of the piece of music
   being evaluated. This context includes things like the current octave,
   current default note length, and other attributes.")

;;; utils ;;;

(def ^:private intervals
  {"c" 0, "d" 2, "e" 4, "f" 5, "g" 7, "a" 9, "b" 11})

(defn- midi-note
  "Given a letter and an octave, returns the MIDI note number.
   e.g. 'c', 4  =>  60"
  [letter octave]
  (+ (intervals letter) (* octave 12) 12))

(defn- midi->hz
  "Converts a MIDI note number to the note's frequency in Hz."
  [note]
  (* 440.0 (Math/pow 2.0 (/ (- note 69.0) 12.0))))

;;;;;;;;;;;;;

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

; TO DO: change this so that it's a higher order function that works within
; the context of the current tempo, and returns the duration in milliseconds
(defn duration
  "Combines a variable number of tied note-lengths into one.

   A slur may appear as the final argument of a duration, making the current
   note legato (effectively slurring it into the next)."
  [& components]
  (if (= (last components) :slur)
    {:beats (apply + (drop-last 1 components)), :slur true}
    {:beats (apply + components)}))

(defn pitch
  "Determines the frequency in Hz, within the context of the current
   octave."
  [letter & accidentals]
  (fn [octave]
    (let [midi-note (reduce (fn [number accidental]
                              (case accidental
                                :flat  (dec number)
                                :sharp (inc number)))
                            (midi-note letter octave)
                            accidentals)]
      (midi->hz midi-note))))

(defn note
  "to do"
  ([pitch-fn])
  ([pitch-fn & more])) ; more = duration, slur, or both

(defn pause
  "to do (pause = rest)"
  ([])
  ([duration-fn]))

(defn octave
  "Sets the current octave. Like pitch, this is also dependent on the current
   octave, and is implemented as a higher-order function."
  [arg]
  (fn [current-octave]
    (case arg
      "<" (dec current-octave)
      ">" (inc current-octave)
      arg)))