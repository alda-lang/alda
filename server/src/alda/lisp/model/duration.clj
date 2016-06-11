(ns alda.lisp.model.duration
  (:require [alda.lisp.model.event :refer (add-events)]))

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

(defn max-beats
  "Returns the duration in beats of the longest event in `events`.

   Ignores events which have no duration in beats.

   Returns 0 if none of the events have a duration in beats."
  [events]
  (apply max (for [{:keys [beats]} (cons {:beats 0} events)
                   :when beats]
               beats)))

(defn calculate-duration
  "Given a number of beats, a tempo, and a time-scaling factor, calculates the
   duration in milliseconds.

   Takes as an optional final argument a number of milliseconds to add to the
   total."
  [beats tempo time-scaling & [ms]]
  (+ (float (* beats
               (/ 60000 tempo)
               time-scaling))
     (or ms 0)))

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

   Returns a map containing the total number of beats (counting only those
   note-lengths that are expressed in standard musical notation), the total
   number of milliseconds (counting only those note-lengths expressed in
   milliseconds), and whether or not the note is slurred.

   This information is used by events (like notes and rests) to calculate the
   total duration in milliseconds (as this depends on the score's time-scaling
   factor and the tempo of the instrument the event belongs to)."
  [& components]
  (let [components (remove nil? components)
        [note-lengths slurred] (if (= (last components) :slur)
                                 (conj [(drop-last components)] true)
                                 (conj [components] false))
        note-lengths (map (fn [x] (if (map? x)
                                    x
                                    {:type :beats, :value x}))
                          note-lengths)
        beats-components (for [{:keys [type value]} note-lengths
                               :when (= type :beats)]
                           value)
        beats (apply + beats-components)
        ms    (apply + (for [{:keys [type value]} note-lengths
                             :when (= type :milliseconds)]
                         value))]
    {:beats     beats
     :ms        ms
     :slurred   slurred
     :duration? true} ; identify this as a duration map
    ))
