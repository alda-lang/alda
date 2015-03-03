(ns alda.lisp.model.instrument)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.instrument...")

(declare ^:dynamic *initial-attr-values*)

(def ^:dynamic *stock-instruments* {})
(declare ^:dynamic *instruments*)
(declare ^:dynamic *current-instruments*)

(defmacro definstrument
  "Defines a stock instrument."
  [inst-name & things]
  (let [{:keys [aliases initial-vals config] :as opts}
        (if (string? (first things)) (rest things) things)
        inst-aliases (vec (cons (str inst-name) (or aliases [])))
        initial-vals (or initial-vals {})]
    `(doseq [alias# ~inst-aliases]
       (alter-var-root (var *stock-instruments*)
         assoc alias# {:initial-vals (merge ~initial-vals
                                       {:stock ~(str inst-name)
                                        :config ~config})}))))
