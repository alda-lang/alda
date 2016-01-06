(ns alda.lisp.model.attribute
  (:require [taoensso.timbre         :as    log]
            [alda.lisp.model.records :refer (->AttributeChange)]
            [alda.lisp.score.context :refer (*current-instruments*
                                             *initial-attr-values*
                                             *instruments*)]))

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

(defn snapshot
  [instrument]
  (*instruments* instrument))

(defn load-snapshot
  [instrument snapshot]
  (alter-var-root #'*instruments* assoc instrument snapshot))
