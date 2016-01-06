(ns alda.lisp.score.context
  (:require [alda.lisp.model.records :refer (->AbsoluteOffset)]))

; score
(def ^:dynamic *events* {:start {:offset (->AbsoluteOffset 0), :events []}})
(def ^:dynamic *score-text* "")
(def ^:dynamic *time-scaling* 1)
(def ^:dynamic *beats-tally* nil)

; instruments
(def ^:dynamic *current-instruments* #{})
(def ^:dynamic *instruments* {})
(def ^:dynamic *nicknames* {})
(def ^:dynamic *stock-instruments* {})

; attributes
(def ^:dynamic *global-attributes* {})

(def ^:dynamic *initial-attr-values* {:current-offset (->AbsoluteOffset 0)
                                      :last-offset (->AbsoluteOffset 0)
                                      :current-marker :start})

