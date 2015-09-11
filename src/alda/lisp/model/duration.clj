(ns alda.lisp.model.duration)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.duration...")

(declare set-duration)

(defn note-length
  "Converts a number, representing a note type, e.g. 4 = quarter, 8 = eighth,
   into a number of beats. Handles dots if present."
  ([number]
   (/ 4 number))
  ([number {:keys [dots]}]
   (* (/ 4 number) (- 2 (Math/pow 2 (- dots))))))

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
     :slurred slurred
     :beats beats}))
