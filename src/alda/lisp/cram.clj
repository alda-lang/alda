(ns alda.lisp.cram)

(in-ns 'alda.lisp)

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
      ~@body
      (- (:offset ($current-offset)) start#))))

(defmacro cram [& body]
  (let [lst (last body)
        dur (and (list? lst)
                 (= 'alda.lisp/duration (first lst))
                 lst)
        body (if dur (butlast body) body)]
    `(let [b#       (or ~dur ($duration))
           d#       (calc-duration ~@body)
           t#       ((:duration-fn (duration b#)) ($tempo))
           scaling# (* (/ t# d#) ~'alda.lisp/*time-scaling*)]
       (binding [~'alda.lisp/*time-scaling* scaling#]
         ~@body
         (set-duration b#)))))
