(ns alda.lisp.attributes)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.attributes...")

(def ^:dynamic *events* {:start {:offset (AbsoluteOffset. 0), :events []}})
(def ^:dynamic *instruments* {})
(def ^:dynamic *nicknames* {})
(def ^:dynamic *current-instruments* #{})
(def ^:dynamic *initial-attr-values* {:current-offset (AbsoluteOffset. 0),
                                      :last-offset (AbsoluteOffset. 0)
                                      :current-marker :start})

(defn add-event
  [instrument event]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil conj []) event)))

(defmulti set-attribute
  "Top level fn for setting attributes."
  (fn [attr val] attr))

(defmethod set-attribute :default [attr val]
  (log/error (str attr " is not a valid attribute.")))

(defn set-attributes
  "Convenience fn for setting multiple attributes at once.
   e.g. (set-attributes :tempo 100 :volume 50)"
  [& attrs]
  (doall
    (for [[attr num] (partition 2 attrs)]
      (set-attribute attr num))))

(defrecord AttributeChange [inst attr from to])

(defmacro defattribute
  "Convenience macro for setting up attributes."
  [attr-name & things]
  (let [{:keys [aliases kw initial-val fn-name transform] :as opts}
        (if (string? (first things)) (rest things) things)
        kw-name      (or kw (keyword attr-name))
        attr-aliases (vec (cons (keyword attr-name) (or aliases [])))
        transform-fn (or transform #(constantly %))]
    `(do
       (alter-var-root (var *initial-attr-values*) assoc ~kw-name ~initial-val)
       (doseq [alias# ~attr-aliases]
         (defmethod set-attribute alias# [attr# val#]
           (doall
             (for [instrument# *current-instruments*]
               (let [old-val# (-> (*instruments* instrument#) ~kw-name)
                     new-val# ((~transform-fn val#) old-val#)]
                 (alter-var-root (var *instruments*) assoc-in
                                                     [instrument# ~kw-name]
                                                     new-val#)
                 (if (not= old-val# new-val#)
                   (log/debug (format "%s %s changed from %s to %s."
                                      instrument#
                                      ~(str attr-name)
                                      old-val#
                                      new-val#)))
                 (AttributeChange. instrument# ~(keyword attr-name)
                                   old-val# new-val#))))))
       (defn ~(or fn-name attr-name) [x#]
         (set-attribute ~(keyword attr-name) x#)))))

;;;

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

;;;

(declare apply-global-attributes)

(defn set-current-offset
  "Set the offset, in ms, where the next event will occur."
  [instrument offset]
  (let [old-offset (-> (*instruments* instrument) :current-offset)
        current-marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*instruments* assoc-in [instrument :current-offset] offset)
    (apply-global-attributes instrument offset)
    (AttributeChange. instrument :current-offset old-offset offset)))

(defn set-last-offset
  "Set the :last-offset; this value will generally be the value of
   :current-offset before it was last changed. This value is used in
   conjunction with :current-offset to determine whether an event
   occurred within a given window."
  [instrument offset]
  (let [old-offset (-> (*instruments* instrument) :last-offset)]
    (alter-var-root #'*instruments* assoc-in [instrument :last-offset] offset)
    (AttributeChange. instrument :last-offset old-offset offset)))

(defn instruments-all-at-same-offset
  "If all of the *current-instruments* are at the same absolute offset, returns
   that offset. Returns nil otherwise."
  []
  (let [offsets (for [instrument *current-instruments*
                      :let [get-attr #(-> (*instruments* instrument) %)
                            current-offset (get-attr :current-offset)]]
                  (absolute-offset current-offset))]
    (when (apply == offsets)
      (AbsoluteOffset. (first offsets)))))

(defn at-marker
  "Set the marker that events will be added to."
  [marker]
  (doall
    (for [instrument *current-instruments*]
      (let [old-marker (-> (*instruments* instrument) :current-marker)]
        (set-current-offset instrument (RelativeOffset. marker 0))
        (alter-var-root #'*instruments* assoc-in [instrument :current-marker]
                                                 marker)
        (log/debug instrument "is now at marker" (str marker \.))
        (AttributeChange. instrument :current-marker old-marker marker)))))

(defrecord Marker [name offset])

(defn marker
  "Places a marker at the current absolute offset. Logs an error if there are
   multiple instruments active at different offsets."
  [name]
  (if-let [offset (instruments-all-at-same-offset)]
    (do
      (alter-var-root #'*events* assoc-in [name :offset] offset)
      (log/debug "Set marker" (str \" name \") "at offset"
                 (str (int (absolute-offset offset)) \.))
      (Marker. name offset))
    (log/error "Can't place marker" (str \" name \") "- offset unclear.")))

;;;

(comment
  "*global-attributes* is a map of offsets to the global attribute changes that
   occur (for all instruments) at each offset.

   apply-global-attributes is a fn that looks at the window between an
   instrument's :last-offset and :current-offset and sees if any global attribute
   changes have occurred in that window, and if so, applies them for the
   instrument. This fn is called by set-current-offset each time an instrument's
   :current-offset is changed, so that global attributes can be applied at the
   new offset.

   Note: global attributes only work moving forward. That is to say, a global
   attribute change will only affect the current part and any others that
   follow it in the score.")

(def ^:dynamic *global-attributes* {})

(defrecord GlobalAttribute [offset attr val])

(defn global-attribute
  [attr val]
  (set-attribute attr val)
  (if-let [offset (instruments-all-at-same-offset)]
    (do
      (alter-var-root #'*global-attributes* update-in [(absolute-offset offset)]
                                            (fnil conj []) [attr val])
      (log/debug "Set global attribute" attr val "at offset"
                 (str (int (absolute-offset offset)) \.))
      (GlobalAttribute. offset attr val))
    (throw (Exception. (str "Can't set global attribute " attr " to " val
                            " - offset unclear. There are multiple instruments "
                            "active with different time offsets.")))))

(defn global-attributes
  "Convenience fn for setting multiple global attributes at once."
  [& attrs]
  (doall
    (for [[attr val] (partition 2 attrs)]
      (global-attribute attr val))))

(defn apply-global-attributes
  [instrument now]
  (let [get-attr       (fn [attr]
                         (fn [] (-> (*instruments* instrument) attr)))
        current-marker (get-attr :current-marker)
        now            (absolute-offset now)
        last-offset    (absolute-offset ((get-attr :last-offset)))
        new-global-attrs (->> *global-attributes*
                              (filter (fn [[offset attrs]]
                                        (<= last-offset
                                            offset
                                            now)))
                              (mapcat (fn [[offset attrs]] attrs)))]
    (binding [*current-instruments* #{instrument}]
      (doall
        (for [[attr val] new-global-attrs]
          (set-attribute attr val))))))

;;;

(defn snapshot
  [instrument]
  (*instruments* instrument))

(defn load-snapshot
  [instrument snapshot]
  (alter-var-root #'*instruments* assoc instrument snapshot))
