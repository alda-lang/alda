(ns alda.lisp.model.attribute
  (:require [taoensso.timbre        :as    log]
            [alda.lisp.model.event  :refer (update-score)]
            [alda.lisp.model.offset :refer (absolute-offset)]
            [alda.lisp.score.util   :refer (update-instruments)]))

; This is a map of keywords to Attribute records.
; When an attribute change event occurs in a score, we look up its attribute
; (a keyword) in this map and use the data about this name to update the
; appropriate attribute of the current instruments.
;
; This map gets filled in with Alda's various attributes via the `defattribute`
; macro in alda.lisp.attributes.
(def ^:dynamic *attribute-table* {})

(defn get-attr
  "Given a keyword representing an attribute (which could be an alias, e.g.
   :quant for :quantization), returns the attribute map, which includes the
   attribute's keyword name (e.g. :quantization) and its transform function.

   The transform function is a higher-order function which takes a new,
   user-friendly value (e.g. 100 instead of 1.0) and returns the function to
   apply to an instrument's existing value to update it to the new value.
   (See alda.lisp.attributes for examples of transform functions.)

   Throws an exception if the argument supplied is not a valid keyword name
   or alias for an existing attribute."
  [kw]
  (if-let [attr (*attribute-table* kw)]
    attr
    (throw (Exception. (str kw " is not a valid attribute.")))))

(defn get-kw-name
  "Given an attr (e.g. :tempo), which could be an alias (e.g. :quant for
   :quantization), returns the correct keyword name of the attribute to which
   it refers.

   Throws an exception if the argument supplied is not a valid keyword name or
   alias for an existing attribute."
  [attr]
  (:kw-name (get-attr attr)))

(defn get-val-fn
  "Given an attr (e.g. :tempo) and a user-friendly val (e.g. 100), returns the
   function to apply to an instrument's existing value to update it to the new
   value.

   Throws an exception if the argument supplied is not a valid keyword name or
   alias for an existing attribute."
  [attr val]
  (let [{:keys [transform-fn]} (get-attr attr)]
    (transform-fn val)))

(defn log-attribute-change
  [{:keys [id] :as inst} attr old-val new-val]
  (log/debugf "%s %s changed from %s to %s." id (str attr) old-val new-val))

(defn record-attribute-change
  "Records an attribute change for an instrument.

   Returns the updated instrument.

   The attribute change will take place subsequently when attributes are
   applied."
  [score {:keys [current-offset] :as inst} attr val]
  (let [kw-name (get-kw-name attr)]
    (update-in inst
               [:attributes (absolute-offset current-offset score) kw-name]
               (fnil conj [])
               val)))

(defn apply-attribute
  "Given an instrument map, a keyword representing an attribute, and a value,
   returns the updated instrument with that attribute update applied."
  [{:keys [beats-tally] :as score} inst attr val]
  (let [{:keys [transform-fn kw-name]} (*attribute-table* attr)
        old-val (kw-name inst)
        new-val ((transform-fn val) old-val)]
    (when (and (not= old-val new-val) (not beats-tally))
      (log-attribute-change inst kw-name old-val new-val))
    (assoc inst kw-name new-val)))

(defmethod update-score :attribute-change
  [{:keys [current-instruments] :as score}
   {:keys [attr val] :as attr-change}]
  (update-instruments score
    (fn [{:keys [id] :as inst}]
      (if (contains? current-instruments id)
        (record-attribute-change score inst attr val)
        inst))))

(defn set-attribute
  "Public fn for setting attributes in a score.
   e.g. (set-attribute :tempo 100)"
  [attr val]
  {:event-type :attribute-change
   :attr       (get-kw-name attr)
   :val        val})

(defn set-attributes
  "Convenience fn for setting multiple attributes at once.
   e.g. (set-attributes :tempo 100 :volume 50)"
  [& attrs]
  (for [[attr val] (partition 2 attrs)]
    (set-attribute attr val)))

(defn attribute-changes
  "Determines the attribute changes to apply to an instrument, based on the
   attribute changes established in the score (global attributes) and the
   instrument, as well as the instrument's :last- and :current-offset.

   Returns a map of updated attributes to their new values.

   Each 'value' is actually a vector of values, the length of which depends on
   the number of times this attribute was changed at that point in time. For
   example, if the octave is incremented a bunch of times, then the 'value'
   here will be something like [:up :up :up], whereas a single octave change
   will just be [:up]. `update-score` below will apply the updates
   sequentially."
  [{:keys [global-attributes] :as score}
   {:keys [attributes current-offset last-offset] :as inst}]
  (let [[last current] (map #(absolute-offset % score)
                            [last-offset current-offset])
        [start end]    (if (<= last current) [last current] [0 current])]
    (->> (merge-with (partial merge-with concat) global-attributes attributes)
         (drop-while #(<= (key %) start))
         (take-while #(<= (key %) end))
         (map val)
         (apply merge))))

(defmethod update-score :apply-attributes
  [{:keys [global-attributes instruments current-instruments]
    :as score} _]
  (update-instruments score
    (fn [{:keys [id current-offset last-offset] :as inst}]
      (if (contains? current-instruments id)
        (let [attr-changes (attribute-changes score inst)]
          (reduce (fn [inst [attr vals]]
                    (reduce #(apply-attribute score %1 attr %2)
                            inst
                            vals))
                  inst
                  attr-changes))
        inst))))

(defn apply-attributes
  "For each instrument in :current-instruments, looks between the instrument's
   :last-offset and :current-offset and applies any attribute changes occurring
   within that window.

   Both global and per-instrument attributes are applied; in the case that a
   per-instrument attribute is applied at the exact same time as a global
   attribute, the per-instrument attribute takes precedence for that instrument."
  []
  {:event-type :apply-attributes})

