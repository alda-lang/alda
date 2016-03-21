(ns alda.lisp.model.offset
  (:require [alda.lisp.model.records]
            [alda.lisp.score.util :refer (get-current-instruments)]
            [alda.util            :refer (=%)]
            [taoensso.timbre      :as    log])
  (:import [alda.lisp.model.records AbsoluteOffset RelativeOffset]))

(defprotocol Offset
  (absolute-offset [this score] "Returns the offset in ms from the start of the score.")
  (offset+ [this bump] "Returns a new offset bump ms later."))

(extend-protocol Offset
  Number
  (absolute-offset [x _] x)
  (offset+ [x bump] (+ x bump))

  AbsoluteOffset
  (absolute-offset [this _]
    (:offset this))
  (offset+ [this bump]
    (AbsoluteOffset. (+ (:offset this) bump)))

  RelativeOffset
  (absolute-offset [this {:keys [events markers] :as score}]
    (if-let [marker-offset (get markers (:marker this))]
      (+ marker-offset (:offset this))
      (log/error "Can't calculate offset - marker" (str \" (:marker this) \")
                 "does not have a defined offset.")))
  (offset+ [this bump]
    (RelativeOffset. (:marker this) (+ (:offset this) bump))))

(defn offset=
  "Convenience fn for comparing absolute/relative offsets."
  [score & offsets]
  (if (and (every? #(instance? RelativeOffset %) offsets)
           (apply = (map :marker offsets)))
    (apply =% (map :offset offsets))
    (apply =% (map #(absolute-offset % score) offsets))))

(defn offset<=
  "Convenience fn for determining if offsets are in order."
  [score & offsets]
  (if (and (every? #(instance? RelativeOffset %) offsets)
           (apply = (map :marker offsets)))
    (apply <= (map :offset offsets))
    (apply <= (map #(absolute-offset % score) offsets))))

(defn instruments-all-at-same-offset
  "If all of the :current-instruments are at the same absolute offset, returns
   that offset. Returns nil otherwise.

   (Returns 0 if there are no instruments defined yet, e.g. when placing a
    marker or a global attribute at the beginning of a score.)"
  [{:keys [current-instruments instruments] :as score}]
  (if (empty? current-instruments)
    (AbsoluteOffset. 0)
    (let [offsets (for [{:keys [current-offset]} (get-current-instruments score)]
                    (absolute-offset current-offset score))]
      (when (apply == offsets)
        (AbsoluteOffset. (first offsets))))))
