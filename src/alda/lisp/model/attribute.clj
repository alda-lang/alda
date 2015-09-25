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
        aliases      (or aliases [])
        attr-aliases (vec (cons (keyword attr-name) aliases))
        kw-name      (or kw (keyword attr-name))
        transform-fn (or transform #(constantly %))
        getter-fn    (symbol (str \$ attr-name))
        fn-name      (or fn-name attr-name)
        fn-names     (vec (cons fn-name (map (comp symbol name) aliases)))
        global-fns   (vec (map (comp symbol #(str % \!)) fn-names))]
    (list* 'do
      `(alter-var-root (var *initial-attr-values*) assoc ~kw-name ~initial-val)
      `(doseq [alias# ~attr-aliases]
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
      `(defn ~getter-fn
         ([] (~getter-fn (first *current-instruments*)))
         ([instrument#] (-> (*instruments* instrument#) ~kw-name)))
       (concat 
         (for [fn-name fn-names]
           `(defn ~fn-name [x#]
              (set-attribute ~(keyword attr-name) x#)))
         (for [global-fn-name global-fns]
           `(defn ~global-fn-name [x#]
              (global-attribute ~(keyword attr-name) x#)))))))

(defn snapshot
  [instrument]
  (*instruments* instrument))

(defn load-snapshot
  [instrument snapshot]
  (alter-var-root #'*instruments* assoc instrument snapshot))
