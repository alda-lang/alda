(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace."
  (:require [taoensso.timbre :as log]
            [alda.lisp.model]
            [alda.lisp.attributes]
            [alda.lisp.events]
            [alda.lisp.instruments]
            [alda.lisp.score]
            [alda.lisp.code]))

; example: creating a score interactively at the repl

; (in-ns 'alda.lisp)
; (score*)
; (part* "piano")
; (note (pitch :c) (duration (note-length 8)))
; (note (pitch :d))
; (note (pitch :e))
; (chord (note (pitch :b :flat) (duration (note-length 2)))
;        (octave :up)
;        (note (pitch :d)))
; (score-map)
