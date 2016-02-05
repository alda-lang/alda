(ns alda.lisp.model.global-attribute
  (:require [alda.lisp.model.attribute :refer (get-attr)]
            [alda.lisp.model.event     :refer (update-score)]
            [alda.lisp.model.offset    :refer (absolute-offset
                                               instruments-all-at-same-offset)]
            [taoensso.timbre           :as    log]))

(defmethod update-score :global-attribute-change
  [score {:keys [attr val] :as event}]
  (if-let [offset (instruments-all-at-same-offset score)]
    (let [abs-offset (absolute-offset offset score)]
      (log/debugf "Set global attribute %s %s at offset %d."
                  attr val (int abs-offset))
      (update-in score [:global-attributes abs-offset]
                 (fnil assoc {}) attr val))
    (throw (Exception.
             (str "Can't set global attribute " attr " to " val " - offset "
                  "unclear. There are multiple instruments active with "
                  "different time offsets.")))))

(defn global-attribute
  "Public fn for setting global attributes in a score.
   e.g. (global-attribute :tempo 100)"
  [attr val]
  {:event-type :global-attribute-change
   :attr       (:kw-name (get-attr attr))
   :val        val})

(defn global-attributes
  "Convenience fn for setting multiple global attributes at once.
   e.g. (global-attributes :tempo 100 :volume 50)"
  [& attrs]
  (for [[attr val] (partition 2 attrs)]
    (global-attribute attr val)))

