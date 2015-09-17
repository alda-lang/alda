(ns alda.repl.commands
  (:require [alda.lisp               :refer :all]
            [alda.parser             :refer (parse-input)]
            [alda.repl.core          :as    repl :refer (*repl-reader*
                                                         *parsing-context*)]
            [alda.now                :as    now]
            [boot.from.io.aviso.ansi :refer (bold)]
            [clojure.string          :as    str]
            [clojure.pprint          :refer (pprint)]
            [instaparse.core         :as    insta]))

(defn huh? []
  (println "Sorry, what? I don't understand that command."))

(defn dirty?
  "Returns whether the current score has any unsaved changes.
   
   Note: right now this is just checking to see if the score has ANY changes. 

   TODO: 
   - implement :save command
   - check whether there is any difference between the score and the last-saved version of the score."
  []
  (not (empty? *score-text*)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti repl-command (fn [command rest-of-line] command))

(defmethod repl-command :default [_ _] 
  (huh?))

(defmacro defcommand [cmd-name & things]
  (let [[args & body]  (if (string? (first things)) (rest things) things)
        [rest-of-line] args]
    `(defmethod repl-command ~(str cmd-name) 
       [_# ~rest-of-line] 
       ~@body)))

(defcommand new
  "Create a new score or part."
  [rest-of-line]
  (cond
    (contains? #{"" "score"} rest-of-line)
    (do
      (score*)
      (println "New score initialized."))

    (.startsWith rest-of-line "part ")
    :TODO

    :else
    (huh?)))

(defcommand score
  "Prints the score (as Alda code)."
  [_]
  (println *score-text*))

(defcommand map
  "Prints the data representation of the score in progress."
  [_]
  (pprint (score-map)))

(defcommand play
  "Plays the current score.
   
   TODO: support `from` and `to` arguments (as markers or minute/second marks)"
  [rest-of-line]
  (if (empty? *score-text*)
    (println "You must first create or :load a score.")  
    (do
      (now/refresh!)
      (repl/interpret! *score-text*))))

(defcommand load
  "Load an Alda score into the current REPL session."
  [filename]
  (letfn [(confirm-load []
            (println "Are you sure you want to load" (str filename \?))
            (let [yes-no-prompt (str "(" (bold "y") "es/" (bold "n") "o) > ")
                  response (str/trim (.readLine *repl-reader* yes-no-prompt))]
              (cond
                (contains? #{"y" "yes"} response) true
                (contains? #{"n" "no"} response) false
                :else (confirm-load))))
          (load-score [score-text]
            (let [code (parse-input score-text)]
              (if (insta/failure? code) 
                (do
                  (println)
                  (println code)
                  (println "File load aborted."))
                (do
                  (score*)
                  (eval code)
                  (alter-var-root #'*score-text* (constantly score-text))
                  (println "Score loaded.")))))
          (confirm-and-load-score [score-text]
            (if (confirm-load)
              (load-score score-text)
              (println "File load aborted.")))]
    (if-let [score-text (try
                          (slurp filename)
                          (catch java.io.FileNotFoundException e nil))]
      (if (dirty?)
        (do 
          (println "You have made changes to the current score that will be"
                   "lost if you load" (str filename "."))
          (confirm-and-load-score score-text))
        (load-score score-text))
      (println "File not found:" filename))))

