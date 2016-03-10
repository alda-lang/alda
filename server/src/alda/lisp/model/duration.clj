(ns alda.lisp.model.duration)
(in-ns 'alda.lisp)

(declare set-duration)

; used by CRAM to proportionately expand or shrink the duration of a group of
; events; initial value: 1
(declare ^:dynamic *time-scaling*)
; used by CRAM to calculate *time-scaling*; initial value: nil
(declare ^:dynamic *beats-tally*)

(defn ms
  "Represents a duration value specified in milliseconds.

   Wraps it in a map to give the `duration` function context that we're talking
   about milliseconds, not beats."
  [n]
  {:type :milliseconds
   :value n})

(defn note-length
  "Converts a number, representing a note type, e.g. 4 = quarter, 8 = eighth,
   into a number of beats. Handles dots if present.

   Wraps the result in a map to give the `duration` function context that we're
   talking about a number of beats."
  ([number]
    (note-length number {:dots 0}))
  ([number {:keys [dots]}]
   {:pre [(pos? number)]}
    {:type :beats
     :value (* (/ 4 number)
               (- 2 (Math/pow 2 (- dots))))}))

(defn duration
  "Combines a variable number of tied note-lengths into one.

   Note-lengths can be expressed in beats or milliseconds. This distinction is
   made via the :type key in the map. The number of beats/milliseconds is
   provided via the :value key in the map.

   e.g.
   (note-length 1) => {:type :beats, :value 4}
   (ms 4000)       => {:type :milliseconds, :value 4000}

   To preserve backwards compatibility, a note-length expressed as a simple
   number (not wrapped in a map) is interpreted as a number of beats.

   Barlines can be inserted inside of a duration -- these currently serve a
   purpose in the parse tree only, and evaluate to `nil` in alda.lisp. This
   function ignores barlines by removing the nils.

   A slur may appear as the final argument of a duration, making the current
   note legato (effectively slurring it into the next).

   Returns a map containing a duration-fn, which gives the duration in ms when
   provide with a tempo, and whether or not the note is slurred."
  [& components]
  (let [components (remove nil? components)
        [note-lengths slurred] (if (= (last components) :slur)
                                 (conj [(drop-last components)] true)
                                 (conj [components] false))
        note-lengths (map (fn [x] (if (map? x)
                                    x
                                    {:type :beats, :value x}))
                          note-lengths)
        beats (apply + (for [{:keys [type value]} note-lengths
                             :when (= type :beats)]
                         value))
        ms    (apply + (for [{:keys [type value]} note-lengths
                             :when (= type :milliseconds)]
                         value))]
    (when beats (set-duration beats))
    {:duration-fn (fn [tempo]
                    (+ (float (* beats
                                 (/ 60000 tempo)
                                 *time-scaling*))
                       ms))
     :slurred slurred
     :beats beats}))
