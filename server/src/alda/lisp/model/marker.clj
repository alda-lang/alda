(ns alda.lisp.model.marker
  (:require [alda.lisp.model.event   :refer (update-score)]
            [alda.lisp.model.offset  :refer (absolute-offset
                                             instruments-all-at-same-offset)]
            [alda.lisp.model.records :refer (->RelativeOffset)]
            [taoensso.timbre         :as    log]))

(defmethod update-score :marker
  [score {:keys [name] :as marker}]
  (if-let [offset (instruments-all-at-same-offset score)]
    (do
      (log/debug "Set marker" (str \" name \") "at offset"
                 (str (int (absolute-offset offset score)) \.))
      (assoc-in score [:events name :offset] offset))
    (throw (Exception. (str "Can't place marker" (str \" name \")
                            "- offset unclear.")))))

(defmethod update-score :at-marker
  [{:keys [current-instruments instruments current-voice] :as score}
   {:keys [name] :as at-marker}]
  (doseq [inst current-instruments]
    (log/debug inst "is now at marker" (str name \.)))
  (let [instruments-key (if current-voice
                          [:voice-instruments current-voice]
                          [:instruments])]
    (update-in score instruments-key
               #(into {}
                  (map (fn [[id inst]]
                         (if (contains? current-instruments id)
                           [id (assoc inst :current-marker name)]
                           [id inst]))
                       %)))))

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

