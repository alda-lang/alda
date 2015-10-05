(ns alda.parser-util
  (:require [alda.parser     :refer (parse-input)]
            [instaparse.core :as    insta]
            [clojure.java.io :as    io]))

(defn parse-with-start-rule
  "Parse a string of Alda code starting from a particular level of the tree.
   (e.g. starting from :part, will parse the string as a part, not a whole score)

   With each line of Alda code entered into the REPL, we determine the context of
   evaluation (Are we starting a new part? Are we appending events to an existing
   part? Are we continuing a previous voice? Starting a new voice?) and parse the
   code accordingly."
  [start code]
  (with-redefs [alda.parser/alda-parser
                #((insta/parser (io/resource "alda.bnf")) % :start start)]
    (parse-input code)))

(defn parse-with-context
  "Determine the appropriate context to parse a line of code from the Alda
   REPL, then parse it within that context.
   
   Returns both the context (the name of a parse tree node) and the resulting
   parse tree.
   
   If parsing fails, returns `:parse-failure` and the Instaparse failure object."
  [alda-code]
  (letfn [(try-ctxs [[ctx & ctxs]]
            (let [parsed (parse-with-start-rule ctx alda-code)]
              (if (insta/failure? parsed)
                (if ctxs
                  (try-ctxs ctxs)
                  [:parse-failure parsed])
                [ctx parsed])))]
    (try-ctxs [:music-data :part :score])))

