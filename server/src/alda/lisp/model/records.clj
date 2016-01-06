(ns alda.lisp.model.records)

; attributes
(defrecord AttributeChange [inst attr from to])
(defrecord GlobalAttribute [offset attr val])

; events
(defrecord Note [offset instrument volume track-volume panning midi-note pitch duration])
(defrecord Rest [offset instrument duration])
(defrecord Chord [events])
(defrecord Function [offset function])

; markers
(defrecord Marker [name offset])

; offset
(defrecord AbsoluteOffset [offset])
(defrecord RelativeOffset [marker offset])

