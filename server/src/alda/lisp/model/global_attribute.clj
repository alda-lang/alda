(ns alda.lisp.model.global-attribute
  (:require [alda.lisp.model.attribute :refer (set-attribute
                                               apply-attribute)]
            [alda.lisp.model.event     :refer (update-score)]
            [alda.lisp.model.offset    :refer (absolute-offset
                                               instruments-all-at-same-offset)]
            [alda.lisp.score.util      :refer (update-instruments)]
            [taoensso.timbre           :as    log]))

(defmethod update-score :global-attribute-change
  [score {:keys [attr val] :as event}]
  (if-let [offset (instruments-all-at-same-offset score)]
    (let [abs-offset (absolute-offset offset score)]
      (log/debugf "Set global attribute %s %s at offset %d."
                  attr val (int abs-offset))
      (-> score
          (update-in [:global-attributes abs-offset attr] (fnil conj []) val)
          (update-score (set-attribute attr val))))
    (throw (Exception.
             (str "Can't set global attribute " attr " to " val " - offset "
                  "unclear. There are multiple instruments active with "
                  "different time offsets.")))))

(defn- global-attribute-changes
  "Determines the attribute changes to apply to an instrument, based on the
   attribute changes established in the score (global attributes) and the
   instrument's :last- and :current-offset.

   Returns a map of updated attributes to their new values.

   Each 'value' is actually a vector of values, the length of which depends on
   the number of times this attribute was changed at that point in time. For
   example, if the octave is incremented a bunch of times, then the 'value'
   here will be something like [:up :up :up], whereas a single octave change
   will just be [:up]. `update-score` below will apply the updates
   sequentially."
  [{:keys [global-attributes] :as score}
   {:keys [current-offset last-offset] :as inst}]
  (let [[last current] (map #(absolute-offset % score)
                            [last-offset current-offset])
        [start end]    (if (<= last current) [last current] [0 current])]
    (->> global-attributes
         (drop-while #(<= (key %) start))
         (take-while #(<= (key %) end))
         (map val)
         (apply merge))))

(defmethod update-score :apply-global-attributes
  [{:keys [global-attributes instruments current-instruments]
    :as score} _]
  (update-instruments score
    (fn [{:keys [id current-offset last-offset] :as inst}]
      (if (contains? current-instruments id)
        (let [attr-changes (global-attribute-changes score inst)]
          (reduce (fn [inst [attr vals]]
                    (reduce #(apply-attribute score %1 attr %2)
                            inst
                            vals))
                  inst
                  attr-changes))
        inst))))

