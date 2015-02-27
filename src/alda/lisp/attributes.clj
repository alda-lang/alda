(ns alda.lisp.attributes)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.attributes...")

(def ^:dynamic *initial-attr-values* {:current-offset (AbsoluteOffset. 0)
                                      :last-offset (AbsoluteOffset. 0)
                                      :current-marker :start})

(defn- percentage [x]
  {:pre [(<= 0 x 100)]}
  (constantly (/ x 100.0)))

(defattribute tempo
  "Current tempo. Used to calculate the duration of notes."
  :initial-val 120)

(defattribute duration
  "Default note duration in beats."
  :initial-val 1
  :fn-name set-duration)

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
  :var *quant*
  :aliases [:quant :quantize]
  :initial-val 0.9
  :fn-name quant
  :transform percentage)

(defattribute volume
  "Current volume."
  :aliases [:vol]
  :initial-val 1.0
  :transform percentage)

(defattribute panning
  "Current panning."
  :aliases [:pan]
  :initial-val 0.5
  :transform percentage)
