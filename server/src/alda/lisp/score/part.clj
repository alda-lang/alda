(ns alda.lisp.score.part
  (:require [djy.char                         :refer (char-range)]
            [instaparse.core                  :as    insta]
            [clojure.java.io                  :as    io]
            [clojure.string                   :as    str]
            [alda.lisp.model.global-attribute :refer (apply-global-attributes)]
            [alda.lisp.model.records          :refer (->AbsoluteOffset)]
            [alda.lisp.score.context          :refer (*current-instruments*
                                                      *initial-attr-values*
                                                      *instruments*
                                                      *nicknames*
                                                      *stock-instruments*)]
            [alda.parser-util                 :refer (parse-with-context)]
            [taoensso.timbre                  :as    log]))

(defn generate-id
  [name]
  (let [rand-char (fn [] (rand-nth (concat (char-range \0 \9)
                                           (char-range \a \z)
                                           (char-range \A \Z))))
        id (apply str (take 5 (repeatedly #(rand-char))))]
    (str name \- id)))

(defn init-instrument
  "Initializes a stock instrument instance with values for tempo,
   current-offset, volume, octave, etc. Adds it to *instruments* and also
   returns it.

   Logs an error if the stock instrument hasn't been defined."
  [stock-inst & attrs]
  (when-let [instrument (if-let [{:keys [initial-vals]}
                               (*stock-instruments* stock-inst)]
                          (merge *initial-attr-values*
                                 {:id (generate-id stock-inst)}
                                 initial-vals
                                 (apply hash-map attrs))
                          (log/error "Stock instrument"
                                     (str \" stock-inst \")
                                     "not defined."))]
    (alter-var-root #'*instruments*
                    assoc (:id instrument) instrument)
    instrument))

(defn determine-instances
  "Given an instrument call (as a map with names and nickname keys), determines
   the instrument instances that will become the *current-instruments*.
   Initializes instrument instances / updates nicknames when appropriate."
  [{:keys [names nickname]}]
  (let [instances
        (remove nil?
          (flatten
            (for [name names]
              (if (contains? *nicknames* name)
                (*nicknames* name)
                (if nickname
                  (:id (init-instrument name))
                  (if-let [existing-inst
                           (first
                             (for [[id attrs] *instruments*
                                   :when (str/starts-with? id (str name \-))]
                               (:id attrs)))]
                    existing-inst
                    (:id (init-instrument name))))))))]
    (when nickname
      (alter-var-root #'*nicknames* assoc nickname instances))
    (set instances)))

(defn parse-instrument-call [s]
  (parse-with-context :calls (-> s
                                 (str/replace #":$" "")
                                 (str/replace #"'" "\"")
                                 (str \:))))

(defmulti part* type)

(defmethod part* clojure.lang.PersistentArrayMap
  [instrument-call]
  (alter-var-root (var *current-instruments*)
                  (constantly (determine-instances instrument-call)))
  (doseq [instrument *current-instruments*]
    (apply-global-attributes instrument (->AbsoluteOffset 0))))

(defmethod part* String
  [instrument-call]
  (part* (parse-instrument-call instrument-call)))

(defmacro part
  "Determines the current instrument(s) and executes the events."
  [instrument-call & events]
  `(do
     (part* ~instrument-call)
     ~@events))
