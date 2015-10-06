(ns alda.lisp.events.cram)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.cram...")

(def no-events
  {:start {:offset (->AbsoluteOffset 0), :events []}})

(defmacro resetting [vars & body]
  (if (seq vars)
    (let [[x & xs] vars]
      `(let [before# ~x
             result# (resetting ~xs ~@body)]
         (alter-var-root (var ~x) (constantly before#))
         result#))
    `(do ~@body)))

(defmacro calc-duration [& body]
  `(resetting
    [~'alda.lisp/*events*
     ~'alda.lisp/*instruments*
     ~'alda.lisp/*current-instruments*
     ~'alda.lisp/*global-attributes*]
    (let [start#  (:offset ($current-offset))]
      (set-duration 1)
      ~@body
      (- (:offset ($current-offset)) start#))))

(defmacro cram [& body]
  (let [lst (last body)
        dur (and (or (coll? lst))
                 (= 'alda.lisp/duration (first lst))
                 lst)
        body (if dur (butlast body) body)]
    `(let [dur#   (:beats ~dur)
           d#     (calc-duration ~@body)
           is#    alda.lisp/*current-instruments*
           ts#    alda.lisp/*time-scaling*
           beats# (zipmap is# (for [i# is#] (or dur# ($duration i#))))
           durs#  (zipmap is# (for [i# is#]
                                (let [f# (:duration-fn
                                           (duration (beats# i#)))]
                                  (f# ($tempo i#)))))
           scale# (merge
                    ts#
                    (zipmap is#
                            (map #(* (/ (durs# %) d#) (ts# % 1)) is#)))]
       (binding [~'alda.lisp/*time-scaling* scale#]
         (set-duration 1)
         ~@body
         (doseq [i# is#]
           (binding [alda.lisp/*current-instruments* #{i#}]
             (set-duration (beats# i#))))))))
