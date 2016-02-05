(ns alda.lisp.model.records)

; attributes
(defrecord Attribute [kw-name transform-fn])

; events
(defrecord Note [offset instrument volume track-volume panning midi-note pitch duration voice])
(defrecord Function [offset instrument function])

; offset
(defrecord AbsoluteOffset [offset])
(defrecord RelativeOffset [marker offset])

