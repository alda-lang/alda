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

(defn- check-key
  "Modifies the accidentals on notes to fit the key signature.
   If there are no accidentals and this letter is in the signature, 
   return the letter's signature accidentals, otherwise return 
   existing accidentals"
  [signature letter accidentals]
   (if (empty? accidentals)
       (get signature letter nil) 
       (identity accidentals)))

(defn pitch
  "Returns a fn that will calculate the frequency in Hz, within the context
   of the octave that an instrument is in."
  [parsed-pitch]
  (let [letter (first parsed-pitch) accidentals (rest parsed-pitch)]
  (fn [octave key-signature & {:keys [midi]}]
    (let [midi-note (reduce (fn [number accidental]
                              (log/debug (format "number %s accidental %s" number accidental))
                              (case accidental
                                :flat    (dec number)
                                :sharp   (inc number)
                                :natural (identity number)))
                            (midi-note letter octave)
                            (check-key key-signature letter accidentals) )]
      (if midi
        midi-note
        (midi->hz midi-note))))))

(defn parse-pitch 
  "Converts the tokenized representation of a pitch into a keyword sequence.
   For example, ('c' '+' '+') will become (:c :sharp :sharp)"
  [letter & accidentals]
  (list* (keyword (str letter))
        (map {\+ :sharp, \= :natural, \- :flat} accidentals))
)
