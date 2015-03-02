(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace."
  (:require [taoensso.timbre :as log]
            [alda.lisp.model]
            [alda.lisp.attributes]
            [alda.lisp.events]
            [alda.lisp.instruments]
            [alda.lisp.score]))

; test setup
(let [{:keys [id]} (init-instrument "piano")]
  (alter-var-root (var *current-instruments*) conj id))
