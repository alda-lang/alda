(ns alda.lisp.code
  (:require [alda.parser-util :refer (parse-to-lisp-with-context)]))

(defn alda-code
  "Attempts to parse a string of text within the context of the current score;
   if the code parses successfully, the result is one or more events that are
   spliced into the score."
  [code]
  (let [[context parse-result] (parse-to-lisp-with-context code)]
    (eval (case context
            :music-data    (cons 'vector parse-result)
            :score         (cons 'vector (rest parse-result))
            :parse-failure (throw (Exception. (str "Invalid Alda code: " code)))
            parse-result))))

