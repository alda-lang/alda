(ns alda.lisp.model.offset)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.offset...")

(declare ^:dynamic *events*
         ^:dynamic *current-instruments*
         apply-global-attributes)

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
    (apply == (map :offset offsets))
    (apply == (map absolute-offset offsets))))

;;;

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
   that offset. Returns nil otherwise.

   (Returns 0 if there are no instruments defined yet, e.g. when placing a
    marker or a global attribute at the beginning of a score.)"
  []
  (if (empty? *current-instruments*)
    (AbsoluteOffset. 0)
    (let [offsets (for [instrument *current-instruments*
                        :let [get-attr #(-> (*instruments* instrument) %)
                              current-offset (get-attr :current-offset)]]
                    (absolute-offset current-offset))]
      (when (apply == offsets)
        (AbsoluteOffset. (first offsets))))))
