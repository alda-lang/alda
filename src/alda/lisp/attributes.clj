(ns alda.lisp.attributes)
(in-ns 'alda.lisp)

(def ^:dynamic *attribute-vars* #{})

(defmulti set-attribute
  "Top level fn for setting attributes."
  (fn [attr val] attr))

(defmethod set-attribute :default [attr val]
  (throw (Exception. (str attr " is not a valid attribute."))))

(defn set-attributes
  "Convenience fn for setting multiple attributes at once.
   e.g. (set-attributes :tempo 100 :volume 50)"
  [& attrs]
  (doall
    (for [[attr num] (partition 2 attrs)]
      (set-attribute attr num))))

(defrecord AttributeChange [attr val])

(defmacro defattribute
  "Convenience macro for setting up attributes."
  [attr-name & things]
  (let [{:keys [aliases var initial-val
                fn-name transform] :as opts} (if (string? (first things))
                                               (rest things)
                                               things)
        var-name     (or var (symbol (str \* attr-name \*)))
        attr-aliases (vec (cons (keyword attr-name) (or aliases [])))
        transform-fn (or transform #(constantly %))]
    `(do
       (def ~(vary-meta var-name assoc :dynamic true) ~initial-val)
       (alter-var-root (var *attribute-vars*) conj (var ~var-name))
       (doseq [alias# ~attr-aliases]
         (defmethod set-attribute alias# [attr# val#]
           (let [new-value# (alter-var-root (var ~var-name)
                                            (~transform-fn val#))]
             (AttributeChange. ~(keyword attr-name) new-value#))))
       (defn ~(or fn-name attr-name) [x#]
         (set-attribute ~(keyword attr-name) x#)))))

;;;

(defn- percentage [x]
  {:pre [(<= 0 x 100)]}
  (constantly (/ x 100.0)))

(defattribute instruments
  "The instruments for the current part."
  :initial-val []
  :fn-name set-instruments)

(defattribute current-marker
  "The marker that events will be added to."
  :initial-val :start
  :fn-name set-current-marker)

(defattribute events
  "The master map of events, keyed by time marker.
   As the score is evaluated, events are added to the appropriate marker,
   and new markers may be added dynamically."
  :initial-val {:start {:offset 0, :events []}}
  :fn-name add-event
  :transform (fn [event]
               #(update-in % [*current-marker* :events] conj event)))

(declare new-global-attrs)

(defattribute current-offset
  "The offset, in ms, where the next event will occur."
  :initial-val 0
  :fn-name set-current-offset
  :transform (fn [offset]
               (doseq [[attr val] (new-global-attrs offset)]
                 (set-attribute attr val))
               (constantly offset)))

(defattribute last-offset
  "The value of current-offset before it was last changed.
   Can be used in conjunction with current-offset to determine whether an
   event occurred within a certain window."
  :initial-val 0
  :fn-name set-last-offset)

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

(defattribute global-attributes
  "A map of offsets to the global attribute changes that occur at that offset.
   Each time current-offset is updated, it checks here to see if the new offset
   has crossed any points where global attributes occur, and if so, applies
   them.

   Note: this only works moving forward. That is to say, a global attribute
         change event will only affect the current part and any others that
         follow it in the score."
  :initial-val (sorted-map)
  :fn-name global-attribute*
  :transform (fn [[attr val]]
               (set-attribute attr val)
               #(update-in % [*current-offset*] (fnil conj []) [attr val])))

(defn global-attribute
  [attr val]
  (global-attribute* [attr val]))

(defn global-attributes
  "Convenience fn for setting multiple global attributes at once."
  [& attrs]
  (doall
    (for [[attr val] (partition 2 attrs)]
      (global-attribute attr val))))

(defn new-global-attrs
  "Determines any global attributes to be applied at a given moment."
  [now]
  (->> (filter (fn [[offset attrs]]
                 (<= *last-offset* offset now))
               *global-attributes*)
       (mapcat second)))

;;;

(defn snapshot
  "Returns a map of current offset and attribute information."
  []
  (into {} (map (fn [var] [var (deref var)])
                (disj *attribute-vars* #'*events* #'*global-attributes*))))

(defn load-snapshot
  "Alters the current values of vars based on the input snapshot."
  [snapshot]
  (doseq [[var val] snapshot]
    (alter-var-root var (constantly val))))
