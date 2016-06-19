(ns alda.repl.core
  (:require [instaparse.core  :as    insta]
            [taoensso.timbre  :as    log]
            [clojure.java.io  :as    io]
            [clojure.string   :as    str]
            [alda.parser      :refer (parse-input)]
            [alda.parser-util :as    p-util]
            [alda.lisp        :refer :all]
            [alda.now         :as    now])
  (:import [jline.console ConsoleReader]))

(def ^:dynamic *repl-reader* (doto (ConsoleReader.)
                               (.setExpandEvents false)
                               (.setPrompt "> ")))

(defn new-repl-score
  [& [score]]
  (doto (if score
          (now/new-score score)
          (now/new-score))
    (swap! assoc :parsing-context :part
                 :score-text      "")))

(def ^:dynamic *current-score* (new-repl-score))

(defn score-text<<
  [txt]
  (swap! *current-score*
         update :score-text
         #(str % (when-not (empty? %) \newline) txt)))

(defn close-score!
  []
  (now/tear-down! *current-score*))

(defn new-score!
  []
  (alter-var-root #'*current-score* (constantly (new-repl-score))))

(defn load-score!
  [score-text]
  (let [loaded-score (-> score-text (parse-input :map) new-repl-score)]
    (alter-var-root #'*current-score* (constantly loaded-score))
    (swap! *current-score* assoc :score-text score-text)))

(defn parse-with-context!
  "Determine the appropriate context to parse a line of code from the Alda
   REPL, then parse it within that context.

   Sets the :parsing-context of the score or logs an error, depending on the
   outcome of the parse attempt."
  [alda-code]
  (let [[context parse-result] (p-util/parse-to-lisp-with-context alda-code)]
    (if (= context :parse-failure)
      (log/error "Invalid Alda syntax.")
      (do
        (swap! *current-score* assoc :parsing-context context)
        parse-result))))

(defn refresh-prompt!
  "Sets the REPL prompt to give the user clues about the current context."
  []
  (let [{:keys [current-instruments nicknames]} @*current-score*
        abbrevs (for [inst current-instruments]
                  (if-let [nickname (first (for [[k v] nicknames
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
  "Parse and playback alda-code in the current context.

   Returns true iff code was parsable."
  [alda-code]
  (log/debug "Parsing code...")
  (let [parsed (parse-with-context! alda-code)]
    (log/debug "Done parsing code.")
    (require '[alda.lisp :refer :all])
    (now/with-score *current-score*
      (now/play! (eval (case (:parsing-context @*current-score*)
                         :music-data (vec parsed)
                         :part       parsed
                         :score      (vec (rest parsed))
                         parsed))))
    (boolean parsed)))

(defn play-score!
  "Plays back the current score."
  []
  (now/play-score! *current-score*))
