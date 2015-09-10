(ns alda.lisp.attributes)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.attributes...")

(def ^:dynamic *initial-attr-values* {:current-offset (AbsoluteOffset. 0)
                                      :last-offset (AbsoluteOffset. 0)
                                      :current-marker :start})

(defn- percentage [x]
  {:pre [(<= 0 x 100)]}
  (constantly (/ x 100.0)))

(defn- unbound-percentage [x]
  {:pre [(<= 0 x)]}
  (constantly (/ x 100.0)))

(defattribute tempo
  "Current tempo. Used to calculate the duration of notes."
  :initial-val 120)

(defattribute duration
  "Default note duration in beats."
  :initial-val 1
  :fn-name set-duration)

(defmethod set-attribute :note-length [attr val]
  "The value for this has to be expressed as (duration (note-length ...)).

   Using (duration (note-length ...)) to express a note duration happens to
   already set duration via (set-attribute :duration ...), which leaves nothing
   for (set-attribute :note-length ...) to actually do. Implementing this as a
   placeholder so that the multimethod doesn't complain that :note-length isn't
   a valid attribute :)"

  (for [instrument *current-instruments*]
    (AttributeChange. instrument :duration :??? (:beats val))))

(defattribute octave
  "Current octave. Used to calculate the pitch of notes."
  :initial-val 4
  :transform (fn [val]
               {:pre [(or (number? val)
                          (contains? #{:down :up} val))]}
               (case val
                :down dec
                :up inc
                (constantly val))))

(defattribute quantization
  "The percentage of a note that is heard.
   Used to put a little space between notes.

   e.g. with a quantization value of 90%, a note that would otherwise last
   500 ms will be quantized to last 450 ms. The resulting note event will
   have a duration of 450 ms, and the next event will be set to occur in 500 ms."
  :aliases [:quant :quantize]
  :initial-val 0.9
  :fn-name quant
  :transform unbound-percentage)

(defattribute volume
  "Current volume. For MIDI purposes, the velocity of individual notes."
  :aliases [:vol]
  :initial-val 1.0
  :transform percentage)

(defattribute track-volume
  "More general volume for the track as a whole. Although this can be changed
   just as often as volume, to do so is not idiomatic. For MIDI purposes, this
   corresponds to the volume of a channel."
  :aliases [:track-vol]
  :initial-val (/ 100.0 127.0)
  :transform percentage)

(defattribute panning
  "Current panning."
  :aliases [:pan]
  :initial-val 0.5
  :transform percentage)
