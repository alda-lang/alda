(ns alda.lisp.model.instrument)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.instrument...")

(declare ^:dynamic *initial-attr-values*)

(def ^:dynamic *stock-instruments* {})
(def ^:dynamic *instruments* {})
(def ^:dynamic *current-instruments* #{})

(defmacro definstrument
  "Defines a stock instrument."
  [inst-name & things]
  (let [{:keys [aliases initial-vals type] :as opts}
        (if (string? (first things)) (rest things) things)
        inst-aliases (vec (cons (str inst-name) (or aliases [])))
        initial-vals (or initial-vals {})]
    `(doseq [alias# ~inst-aliases]
       (alter-var-root (var *stock-instruments*)
         assoc alias# {:type ~type
                       :initial-vals (assoc ~initial-vals
                                       :stock ~(str inst-name))}))))
