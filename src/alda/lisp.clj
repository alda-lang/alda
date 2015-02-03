(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace.")

(defmacro alda-eval [code]
  "Uses eval within the alda.lisp namespace to evaluate code written
   (or generated) in Alda's Lisp DSL."
  (prn :code code)
  (eval code))

;;; score-builder utils ;;;

(defn- add-globals
  "If initial global attributes are set, add them to the first instrument's
   music-data."
  [global-attrs instruments]
  (letfn [(with-global-attrs [[tag call data :as instrument]]
            `(~tag
               ~call
               (music-data ~global-attrs ~@(rest data))))]
    (if global-attrs
      (cons (with-global-attrs (first instruments)) (rest instruments))
      instruments)))

(defn part [& args]
  (identity args))

; FIXME
(defn build-parts
  "Walks through a variable number of instrument calls, building a score
   from scratch. Handles initial global attributes, if present."
  [& components]
  (let [[global-attrs & instrument-calls] (if (= (ffirst components)
                                                 'global-attributes)
                                            components
                                            (cons nil components))
        instrument-calls (add-globals global-attrs instrument-calls)
        ]
    `(for [[[name# number#] music-data#] (-> {:parts {} :name-table {} :nickname-table {}}
                                             ((apply comp ~instrument-calls))
                                             :parts)]
       (part name# number# music-data#))))

(defn- assign-new
  "Assigns a new instance of x, given the table of existing instances."
  [x name-table]
  (let [existing-numbers (for [[name num] (apply concat (vals name-table))
                               :when (= name x)]
                           num)]
    (if (seq existing-numbers)
      [[x (inc (apply max existing-numbers))]]
      [[x 1]])))

;;; score-builder ;;;

; FIXME?
(defmacro alda-score
  "Returns a new version of the code involving consolidated instrument parts
   instead of overlapping instrument calls."
  [& components]
  (let [parts (apply build-parts components)]
  `(score ~parts)))

(defn instrument-call
  "Returns a function which, given the context of the score-builder in
   progress, adds the music data to the appropriate instrument part(s)."
  [& components]
  (let [[_ & music-data] (last components)
        names-and-nicks (drop-last components)
        names (for [{:keys [name]} names-and-nicks :when name] name)
        nickname (some :nickname names-and-nicks)]
    (fn [working-data]
      (reduce (fn [{:keys [parts name-table nickname-table]} name]
                (let [name-table (or (and (map? name-table) name-table) {})
                      instance
                      (if nickname
                        (nickname-table name (assign-new name name-table))
                        (name-table name [[name 1]]))]
                  {:parts (merge-with concat parts {instance music-data})
                   :name-table (assoc name-table name instance)
                   :nickname-table (if nickname
                                     (merge-with concat nickname-table
                                                        {nickname instance})
                                     nickname-table)}))
              working-data
              names))))

(defn music-data
  "to do"
  [& events]
  `(music-data ~@events))

;;; score evaluator utils ;;;

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

;;; score evaluator ;;;

(defn score
  "to do"
  [& args]
  (prn "yo"))

(defn voices
  "to do"
  [& args]
  `(voices ~@args))

(defn voice
  "to do"
  [& args]
  `(voice ~@args))

(defn chord
  "to do"
  [& args]
  `(chord ~@args))

(defn marker
  "to do"
  [& args]
  `(marker ~@args))

(defn at-marker
  "to do"
  [& args]
  `(at-marker ~@args))

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
    (fn [tempo]
      {:duration (* beats (/ 60000 tempo))
       :slurred slurred})))

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
  ([pitch-fn]
   `(note ~pitch-fn))
  ([pitch-fn & more]
   `(note ~pitch-fn ~@more))) ; more = duration, slur, or both

(defn pause
  "to do (pause = rest)"
  ([]
   `(pause))
  ([duration-fn]
   `(pause ~duration-fn)))

(defn octave
  "Sets the current octave. Like pitch, this is also dependent on the current
   octave, and is implemented as a higher-order function."
  [arg]
  (fn [current-octave]
    (case arg
      "<" (dec current-octave)
      ">" (inc current-octave)
      arg)))
