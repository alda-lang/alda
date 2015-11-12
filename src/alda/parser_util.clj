(ns alda.parser-util
  (:require [alda.parser     :refer :all]
            [instaparse.core :as    insta]
            [clojure.java.io :as    io]))

(defn- test-parse-music-data
  [alda-code]
  (let [cache (atom {})]
    (->> [alda-code cache]
         remove-comments
         ((fn [[alda-code cache]]
            (parse-part cache alda-code))))))

(defn- test-parse-part
  [alda-code]
  (let [cache (atom {})]
    (->> [alda-code cache]
         remove-comments
         separate-parts
         (insta/transform
           {:score  #(if (> (count %&) 1)
                       (throw (Exception. "This is more than one part."))
                       (first %&))
            :header #(parse-header cache (apply str %&))
            :part   (fn [names & music-data]
                      (list* 'alda.lisp/part
                             names
                             (parse-part cache (apply str music-data))))}))))

(defn- test-parse-calls
  [alda-code]
  (->> alda-code
       score-parser
       check-for-failure
       (insta/transform
         (merge name-transforms
                {:score  #(if (> (count %&) 1)
                           (throw (Exception. "More than one group of calls."))
                           (first %&))
                 :header #(if (pos? (count %&))
                            (throw (Exception. "Not an instrument call.")))
                 :part   #(if (> (count %&) 1)
                            (throw (Exception. "Not an instrument call."))
                            (first %&))
                 :calls  (fn [& calls]
                           (let [names    (vec (keep :name calls))
                                 nickname (some :nickname calls)]
                             (if nickname
                               {:names names, :nickname nickname}
                               {:names names})))}))))

(defn parse-with-start-rule
  "Parse a string of Alda code starting from a particular level of the tree.
   (e.g. starting from :part, will parse the string as a part, not a whole score)

   With each line of Alda code entered into the REPL, we determine the context of
   evaluation (Are we starting a new part? Are we appending events to an existing
   part? Are we continuing a previous voice? Starting a new voice?) and parse the
   code accordingly."
  [start code]
  (case start
    :music-data (test-parse-music-data code)
    :part       (test-parse-part code)
    :calls      (test-parse-calls code)
    :score      (parse-input code)))

(defn parse-with-context
  "Determine the appropriate context to parse a line of code from the Alda
   REPL, then parse it within that context.

   Returns both the context (the name of a parse tree node) and the resulting
   parse tree.

   If parsing fails, returns `:parse-failure` and the error that was thrown."
  [alda-code]
  (letfn [(try-ctxs [[ctx & ctxs]]
            (try
              (let [parsed (parse-with-start-rule ctx alda-code)]
                [ctx parsed])
              (catch Exception e
                (if ctxs
                  (try-ctxs ctxs)
                  [:parse-failure e]))))]
    (try-ctxs [:music-data :part :score])))

