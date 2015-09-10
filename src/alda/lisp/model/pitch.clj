(ns alda.lisp.model.pitch)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.pitch...")

(def ^:private intervals
  {:c 0, :d 2, :e 4, :f 5, :g 7, :a 9, :b 11})

(defn- midi-note
  "Given a letter and an octave, returns the MIDI note number.
   e.g. :c, 4  =>  60"
  [letter octave]
  (+ (intervals letter) (* (inc octave) 12)))

(defn- midi->hz
  "Converts a MIDI note number to the note's frequency in Hz."
  [note]
  (* 440.0 (Math/pow 2.0 (/ (- note 69.0) 12.0))))

(defn- hz->midi
  "Converts a note's frequency in Hz to its MIDI note number."
  [hz]
  (+ 69.0
     (* 12.0
        (/ (Math/log (/ hz 440.0))
           (Math/log 2.0)))))

(defn- accidentals->offset [accidentals]
  (->> accidentals
       (map {:flat -1, :sharp 1})
       (remove nil?)
       (reduce +)))

(defn mean-tempered
  [letter octave accidentals midi?]
  (let [offset (accidentals->offset accidentals)
        midi-note (+ (midi-note letter octave) offset)]
    (if midi? midi-note (midi->hz midi-note))))

(def ^:private werckmeister-ii
  [1
   (/ 256 243)
   (* (/ 64 81) (Math/sqrt 2))
   (/ 32 27)
   (* (/ 256 243) (Math/pow 2 0.25))
   (/ 4 3)
   (/ 1024 729)
   (* (/ 8 9) (Math/pow 8 0.25))
   (/ 128 81)
   (* (/ 1024 729) (Math/pow 2 0.25))
   (/ 16 9)
   (* (/ 128 81) (Math/pow 2 0.25))])

(defn well-tempered
  [letter octave accidentals midi?]
  (let [base-hz (midi->hz (midi-note :c octave))
        offset (accidentals->offset accidentals)
        semitones (+ (intervals letter) offset)
        ratio (nth werckmeister-ii semitones)
        hz (* ratio base-hz)]
    (if midi? (hz->midi hz) hz)))

(def ^:private tunings
  {:well well-tempered
   :mean mean-tempered})

(defn pitch
  "Returns a fn that will calculate the frequency in Hz, within the context
   of the octave that an instrument is in."
  [letter & accidentals]
  (fn [octave & [tuning & {:keys [midi]}]]
    (let [tuning-fn (get tunings tuning mean-tempered)]
      (tuning-fn letter octave accidentals midi))))