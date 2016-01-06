(ns alda.lisp.model.marker
  (:require [alda.lisp.model.event   :refer (set-current-offset)]
            [alda.lisp.model.offset  :refer (absolute-offset
                                             instruments-all-at-same-offset)]
            [alda.lisp.model.records :refer (->AttributeChange
                                             ->Marker
                                             ->RelativeOffset)]
            [alda.lisp.score.context :refer (*current-instruments*
                                             *events*
                                             *instruments*)]
            [taoensso.timbre         :as    log]))

(defn $current-marker
  "Get the :current-marker of an instrument."
  ([] ($current-marker (first *current-instruments*)))
  ([instrument] (-> (*instruments* instrument) :current-marker)))

(defn marker
  "Places a marker at the current absolute offset. Logs an error if there are
   multiple instruments active at different offsets."
  [name]
  (if-let [offset (instruments-all-at-same-offset)]
    (do
      (alter-var-root #'*events* assoc-in [name :offset] offset)
      (log/debug "Set marker" (str \" name \") "at offset"
                 (str (int (absolute-offset offset)) \.))
      (->Marker name offset))
    (log/error "Can't place marker" (str \" name \") "- offset unclear.")))

(defn at-marker
  "Set the marker that events will be added to."
  [marker]
  (doall
    (for [instrument *current-instruments*]
      (let [old-marker ($current-marker instrument)]
        (set-current-offset instrument (->RelativeOffset marker 0))
        (alter-var-root #'*instruments* assoc-in [instrument :current-marker]
                                                 marker)
        (log/debug instrument "is now at marker" (str marker \.))
        (->AttributeChange instrument :current-marker old-marker marker)))))
