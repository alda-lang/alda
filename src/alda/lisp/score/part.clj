(ns alda.lisp.score.part)
(in-ns 'alda.lisp)

(require '[djy.char :refer (char-range)])

(log/debug "Loading alda.lisp.score.part...")

(declare ^:dynamic *nicknames*)

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
  (let [attr-map (apply hash-map attrs)
        id (generate-id stock-inst)
        instrument (merge *initial-attr-values*
                          {:id id}
                          (if-let [{:keys [initial-vals]}
                                   (*stock-instruments* stock-inst)]
                            initial-vals
                            (log/error "Stock instrument"
                                       (str \" stock-inst \")
                                       "not defined."))
                          attr-map)]
    (alter-var-root #'*instruments* assoc-in [id] instrument)
    instrument))

(defn determine-instances
  "Given an instrument call (as a map with names and nickname keys), determines
   the instrument instances that will become the *current-instruments*.
   Initializes instrument instances / updates nicknames when appropriate."
  [{:keys [names nickname]}]
  (let [instances
        (flatten
          (for [name names]
            (if (contains? *nicknames* name)
              (*nicknames* name)
              (if nickname
                (:id (init-instrument name))
                (if-let [existing-inst (first
                                        (for [[id attrs] *instruments*
                                              :when (.startsWith id (str name \-))]
                                          (:id attrs)))]
                  existing-inst
                  (:id (init-instrument name)))))))]
    (when nickname
      (alter-var-root #'*nicknames* assoc nickname instances))
    (set instances)))

(defn part*
  [instrument-call]
  (alter-var-root (var *current-instruments*)
                  (constantly (determine-instances instrument-call)))
  (doseq [instrument# *current-instruments*]
    (apply-global-attributes instrument# (AbsoluteOffset. 0))))

(defmacro part
  "Determines the current instrument(s) and executes the events."
  [instrument-call & events]
  `(do
     (part* ~instrument-call)
     ~@events))
