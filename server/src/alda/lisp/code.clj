(ns alda.lisp.code
  (:require [alda.parser-util :refer (parse-with-context)]
            [taoensso.timbre  :as    log]))

(defn alda-code
  "Attempts to parse a string of text within the context of the current score,
   then evaluates the result."
  [code]
  (let [[context parse-result] (parse-with-context code)]
    (eval (case context
            :music-data    (cons 'do parse-result)
            :score         (cons 'do (rest parse-result))
            :parse-failure (log/error (pr-str parse-result))
            parse-result))))

(defmacro times
  "Evaluates an Alda event (or sequence of events) `n` times."
  [n event]
  `(dotimes [_# ~n] ~event))
