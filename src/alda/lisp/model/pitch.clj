(ns alda.lisp.model.pitch)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.pitch...")

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
  (fn [octave & {:keys [midi]}]
    (let [midi-note (reduce (fn [number accidental]
                              (case accidental
                                :flat  (dec number)
                                :sharp (inc number)))
                            (midi-note letter octave)
                            accidentals)]
      (if midi
        midi-note
        (midi->hz midi-note)))))
