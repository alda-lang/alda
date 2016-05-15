(ns alda.lisp.model.marker
  (:require [alda.lisp.model.event   :refer (update-score)]
            [alda.lisp.model.offset  :refer (absolute-offset
                                             instruments-all-at-same-offset)]
            [alda.lisp.model.records :refer (->AbsoluteOffset
                                             ->RelativeOffset)]
            [alda.lisp.score.util    :refer (update-instruments)]
            [taoensso.timbre         :as    log])
  (:import  [alda.lisp.model.records RelativeOffset]))

(defmethod update-score :marker
  [score {:keys [name] :as marker}]
  (if-let [abs-offset (instruments-all-at-same-offset score)]
    (let [offset (:offset abs-offset)]
      (log/debug "Set marker" (str \" name \") "at offset"
                 (str (int offset) \.))
      (assoc-in score [:markers name] offset))
    (throw (Exception. (str "Can't place marker" (str \" name \")
                            "- offset unclear.")))))

(defmethod update-score :at-marker
  [{:keys [current-instruments markers] :as score}
   {:keys [name] :as at-marker}]
  (when-not (contains? markers name)
    (throw (Exception.
             (format (str "Can't set current marker to \"%s\"; "
                          "marker does not exist.")
                     name))))
  (doseq [inst current-instruments]
    (log/debug inst "is now at marker" (str name \.)))
  (update-instruments score
    (fn [{:keys [id current-offset] :as inst}]
      (if (contains? current-instruments id)
        (-> inst
            (assoc :current-marker name)
            (assoc :current-offset (->RelativeOffset name 0)))
        inst))))

(defn marker
  "Places a marker at the current absolute offset. Throws an exception if there
   are multiple instruments active at different offsets."
  [name]
  {:event-type :marker
   :name       name})

(defn at-marker
  "Set the marker at which events will be added."
  [name]
  {:event-type :at-marker
   :name       name})

