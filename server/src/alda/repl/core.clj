(ns alda.repl.core
  (:require [instaparse.core :as    insta]
            [taoensso.timbre :as    log]
            [clojure.java.io :as    io]
            [clojure.string  :as    str]
            [alda.parser-util]
            [alda.lisp       :refer :all]
            [alda.now        :as    now]))

(declare ^:dynamic *parsing-context*)
(declare ^:dynamic *repl-reader*)
(declare ^:dynamic *score-text*)

(defn score-text<< [s]
  (if (empty? *score-text*)
    (alter-var-root #'*score-text* str s)
    (alter-var-root #'*score-text* str \newline s)))

(defn parse-with-context
  "Determine the appropriate context to parse a line of code from the Alda
   REPL, then parse it within that context.

   Sets `*parsing-context*` or logs an error, depending on the outcome of the
   parse attempt."
  [alda-code]
  (let [[context parse-result] (alda.parser-util/parse-with-context alda-code)]
    (if (= context :parse-failure)
      (log/error "Invalid Alda syntax.")
      (do
        (alter-var-root #'*parsing-context* (constantly context))
        parse-result))))

(defn set-prompt!
  "Sets the REPL prompt to give the user clues about the current context."
  []
  (let [abbrevs (for [inst *current-instruments*]
                  (if-let [nickname (first (for [[k v] *nicknames*
                                                 :when (= v (list inst))]
                                             k))]
                    (->> (re-seq #"(\w)\w*" nickname)
                         (map second)
                         (apply str))
                    (->> (re-seq #"(\w)\w*-" inst)
                         (map second)
                         (apply str))))
        prompt  (str (str/join "/" abbrevs) "> ")]
    (.setPrompt *repl-reader* prompt)))

(defn interpret!
  "Parse and playback alda-code in the current context. Return true iff code was parsable"
  [alda-code]
  (log/debug "Parsing code...")
  (let [parsed (parse-with-context alda-code)]
    (log/debug "Done parsing code.")
    (require '[alda.lisp :refer :all])
    (now/play! (eval (case *parsing-context*
                       :music-data (cons 'do parsed)
                       :score (cons 'do (rest parsed))
                       parsed)))
    (boolean parsed)))
