(ns alda.lisp.model.attribute)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.attribute...")

(declare ^:dynamic *instruments*)

(defmulti set-attribute
  "Top level fn for setting attributes."
  (fn [attr val] attr))

(defmethod set-attribute :default [attr val]
  (log/error (str attr " is not a valid attribute.")))

(defn set-attributes
  "Convenience fn for setting multiple attributes at once.
   e.g. (set-attributes :tempo 100 :volume 50)"
  [& attrs]
  (doall
    (for [[attr num] (partition 2 attrs)]
      (set-attribute attr num))))

(defrecord AttributeChange [inst attr from to])

(defmacro defattribute
  "Convenience macro for setting up attributes."
  [attr-name & things]
  (let [{:keys [aliases kw initial-val fn-name transform] :as opts}
        (if (string? (first things)) (rest things) things)
        kw-name      (or kw (keyword attr-name))
        fn-name      (or fn-name attr-name)
        getter-fn    (symbol (str \$ attr-name))
        attr-aliases (vec (cons (keyword attr-name) (or aliases [])))
        transform-fn (or transform #(constantly %))]
    `(do
       (alter-var-root (var *initial-attr-values*) assoc ~kw-name ~initial-val)
       (doseq [alias# ~attr-aliases]
         (defmethod set-attribute alias# [attr# val#]
           (doall
             (for [instrument# *current-instruments*]
               (let [old-val# (-> (*instruments* instrument#) ~kw-name)
                     new-val# ((~transform-fn val#) old-val#)]
                 (alter-var-root (var *instruments*) assoc-in
                                                     [instrument# ~kw-name]
                                                     new-val#)
                 (if (not= old-val# new-val#)
                   (log/debug (format "%s %s changed from %s to %s."
                                      instrument#
                                      ~(str attr-name)
                                      old-val#
                                      new-val#)))
                 (AttributeChange. instrument# ~(keyword attr-name)
                                   old-val# new-val#))))))
       (defn ~fn-name [x#]
         (set-attribute ~(keyword attr-name) x#))
       (defn ~getter-fn
         ([] (~getter-fn (first *current-instruments*)))
         ([instrument#] (-> (*instruments* instrument#) ~kw-name))))))

(defn snapshot
  [instrument]
  (*instruments* instrument))

(defn load-snapshot
  [instrument snapshot]
  (alter-var-root #'*instruments* assoc instrument snapshot))
