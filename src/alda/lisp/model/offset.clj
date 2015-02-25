(ns alda.lisp.model.offset)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.offset...")

(declare ^:dynamic *events*)

(defprotocol Offset
  (absolute-offset [this] "Returns the offset in ms from the start of the score.")
  (offset+ [this bump] "Returns a new offset bump ms later."))

(defrecord AbsoluteOffset [offset]
  Offset
  (absolute-offset [this]
    offset)
  (offset+ [this bump]
    (AbsoluteOffset. (+ offset bump))))

(defrecord RelativeOffset [marker offset]
  Offset
  (absolute-offset [this]
    (if-let [marker-offset (-> (*events* marker) :offset)]
      (+ (absolute-offset marker-offset) offset)
      (log/warn "Can't calculate offset - marker" (str \" marker \") "does not"
                "have a defined offset.")))
  (offset+ [this bump]
    (RelativeOffset. marker (+ offset bump))))

(defn offset=
  "Convenience fn for comparing absolute/relative offsets."
  [& offsets]
  (if (and (every? #(instance? alda.lisp.RelativeOffset %) offsets)
           (apply = (map :marker offsets)))
    (apply = (map :offset offsets))
    (apply = (map absolute-offset offsets))))
