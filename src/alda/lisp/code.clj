(ns alda.lisp.code)
(in-ns 'alda.lisp)

(require '[alda.parser-util :refer (parse-with-context)])

(defn code-block
  "Represents a literal string of Alda code in alda.lisp.
   
   When evaluated, simply returns the string of code."
  [code]
  code)

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
