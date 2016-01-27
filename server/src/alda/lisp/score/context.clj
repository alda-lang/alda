(ns alda.lisp.score.context
  (:require [alda.lisp.model.records :refer (->AbsoluteOffset)]))

; score
(def ^:dynamic *events* {:start {:offset (->AbsoluteOffset 0), :events []}})
(def ^:dynamic *score-text* "")
(def ^:dynamic *time-scaling* 1)
(def ^:dynamic *beats-tally* nil)
(def ^:dynamic *global-attributes* {})

; instruments
(def ^:dynamic *current-instruments* #{})
(def ^:dynamic *instruments* {})
(def ^:dynamic *nicknames* {})

(declare new-score-context load-score-context)

(defn score-context
  "Captures the current state of all the vars above, which represent the score
   evaluation context."
  []
  (into {}
    (for [[sym var] (ns-publics *ns*)
          :when (not (contains? #{'score-context
                                  'new-score-context
                                  'load-score-context}
                                sym))]
      [sym (var-get var)])))

; initial values for a new score
(def new-score-context
  (score-context))

(defn load-score-context
  "Set the values of all the score evaluation context vars to those stored in
   `ctx`."
  [ctx]
  (doseq [[sym val] ctx]
    (alter-var-root (ns-resolve 'alda.lisp.score.context sym) (constantly val))))
