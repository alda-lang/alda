(ns alda.repl.core
  (:require [instaparse.core :as    insta]
            [taoensso.timbre :as    log]
            [clojure.java.io :as    io]
            [clojure.string  :as    str]
            [alda.parser     :refer (parse-input)]
            [alda.lisp       :refer :all]
            [alda.now        :as    now]))

(declare ^:dynamic *parsing-context*)
(declare ^:dynamic *repl-reader*)

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
   REPL, then parse it within that context."
  [alda-code]
  (letfn [(try-ctxs [[ctx & ctxs]]
            (if ctx
              (let [parsed (parse-with-start-rule ctx alda-code)]
                (if (insta/failure? parsed)
                  (try-ctxs ctxs)
                  (do 
                    (alter-var-root #'*parsing-context* (constantly ctx)) 
                    parsed)))
              (log/error "Invalid Alda syntax.")))]
    (try-ctxs [:music-data :part :score])))

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
    (now/play! (eval (case *parsing-context*
                       :music-data (cons 'do parsed)
                       :score (cons 'do (rest parsed))
                       parsed)))
    (boolean parsed)))
