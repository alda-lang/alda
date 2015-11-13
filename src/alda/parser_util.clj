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

(defn parse-with-context
  "Parse a string of Alda code within a particular context, e.g. to parse
   additional music data for an already existing part, or to parse a single
   instrument part in an already existing score.

   If `ctx` is provided, will attempt to parse the code within that context,
   which will throw an error if parsing fails.

   If `ctx` is NOT provided, will try to parse the code in increasingly broad
   contexts until it parses successfully. When it does parse successfully,
   returns a vector containing the context and the parse result. If parsing
   fails in all contexts, returns a vector containing `:parse-failure` and
   the error that was thrown at the broadest context level."
  ([code]
   (letfn [(try-ctxs [[ctx & ctxs]]
            (try
              (let [parsed (parse-with-context ctx code)]
                [ctx parsed])
              (catch Exception e
                (if ctxs
                  (try-ctxs ctxs)
                  [:parse-failure e]))))]
    (try-ctxs [:music-data :part :score])))
  ([ctx code]
    (case ctx
      :music-data (test-parse-music-data code)
      :part       (test-parse-part code)
      :calls      (test-parse-calls code)
      :score      (parse-input code))))

