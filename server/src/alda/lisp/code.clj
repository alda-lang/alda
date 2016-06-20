(ns alda.lisp.code
  (:require [alda.parser-util :refer (parse-to-events-with-context)]))

(defn alda-code
  "Attempts to parse a string of text within the context of the current score;
   if the code parses successfully, the result is one or more events that are
   spliced into the score."
  [code]
  (let [[context parse-result] (parse-to-events-with-context code)]
    (if (= context :parse-failure)
      (throw (Exception. (str "Invalid Alda code: " code)))
      parse-result)))
