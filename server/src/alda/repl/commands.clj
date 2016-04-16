(ns alda.repl.commands
  (:require [alda.lisp       :refer :all]
            [alda.now        :as    now]
            [alda.parser     :refer (parse-input)]
            [alda.repl.core  :as    repl :refer (*repl-reader*
                                                 *current-score*)]
            [alda.sound      :as    sound]
            [alda.util       :as    util]
            [io.aviso.ansi   :refer (bold)]
            [clojure.pprint  :refer (pprint)]
            [clojure.string  :as    str]
            [instaparse.core :as    insta]))

(defn huh? []
  (println "Sorry, what? I don't understand that command."))

(defn dirty?
  "Returns whether `score` has any unsaved changes.

   Note: right now this is just checking to see if the score has ANY changes.

   TODO:
   - implement :save command
   - check whether there is any difference between the score and the last-saved version of the score."
  [{:keys [score-text] :as score}]
  (not (empty? score-text)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defmulti repl-command (fn [command rest-of-line] command))

(defmethod repl-command :default [_ _]
  (huh?))

(def repl-commands
  (atom {"quit" "Exits the Alda REPL session."}))

(defmacro defcommand [cmd-name & things]
  (let [[doc args & body]  (if (string? (first things))
                             things
                             (cons "" things))
        [rest-of-line] args]
    `(do
       (defmethod repl-command ~(str cmd-name)
         [_# ~rest-of-line]
         ~@body)
       (swap! repl-commands assoc ~(str cmd-name) ~doc))))

; FIXME
(defcommand new
  ; TODO: implement ":new part" and then update the docstring
  "Creates a new score."
  [rest-of-line]
  (cond
    (contains? #{"" "score"} rest-of-line)
    (do
      ; (score*)
      (println "New score initialized."))

    (str/starts-with? rest-of-line "part ")
    :TODO

    :else
    (huh?)))

(defcommand score
  "Prints the score (as Alda code)."
  [_]
  (println (:score-text @*current-score*)))

(defcommand map
  "Prints the data representation of the score in progress."
  [_]
  (pprint @*current-score*))

(defcommand play
  "Plays the current score.

   Can take optional `from` and `to` arguments, in the form of markers or mm:ss times.

   Without arguments, will play the entire score from beginning to end.

   Example usage:

     :play
     :play from 0:05
     :play to 0:10
     :play from 0:05 to 0:10
     :play from guitarIn
     :play to verse
     :play from verse to bridge"
  [rest-of-line]
  (if (empty? (:score-text @*current-score*))
    (println "You must first create or :load a score.")
    (let [{:keys [from to]} (util/parse-str-opts rest-of-line)]
      (sound/with-play-opts (util/strip-nil-values {:from from, :to to})
        (repl/interpret! (:score-text @*current-score*))))))

; FIXME
(defcommand load
  "Loads an Alda score into the current REPL session.

   Usage:

     :load test/examples/bach_cello_suite_no_1.alda
     :load /Users/rick/Scores/love_is_alright_tonite.alda"
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
            (try
              (let [code (parse-input score-text)]
                ; (score*)
                ; (eval code)
                ; (alter-var-root #'*score-text* (constantly score-text))
                (println "Score loaded."))
              (catch Exception e
                (println)
                (println (.getMessage e))
                (println "File load aborted."))))
          (confirm-and-load-score [score-text]
            (if (confirm-load)
              (load-score score-text)
              (println "File load aborted.")))]
    (if (empty? filename)
      (println "Load what?")
      (if-let [score-text (try
                            (slurp filename)
                            (catch java.io.FileNotFoundException e nil))]
        (if (dirty? @*current-score*)
          (do
            (println "You have made changes to the current score that will be"
                     "lost if you load" (str filename "."))
            (confirm-and-load-score score-text))
          (load-score score-text))
        (println "File not found:" filename)))))

(defn- parse-docstring
  "Parses the docstring of a REPL command defined in this namespace into two
   things -- a brief description of what it does (the first line of the
   docstring) and a more detailed description (any subsequent lines)."
  [docstring]
  (let [[description & details] (str/split docstring #"\n")]
    [description (when details (str/join \newline details))]))

(defn- generate-help-text
  ([]
    (str "For commands marked with (*), more detailed information about the "
         "command is available via the :help command.\n\ne.g. :help play\n\n"
         "Available commands:\n\n"
         (str/join \newline
                   (for [[cmd docstring] @repl-commands
                         :let [[desc details] (parse-docstring docstring)]]
                     (str "    :" cmd \tab (str desc (when details " (*)")))))))
  ([subject]
    (if-let [docstring (get @repl-commands subject)]
      (let [[description details] (parse-docstring docstring)]
        (str \: subject \newline
             \newline
             description
             (when details (str \newline details))))
      (format "Help is not available on '%s'." subject))))

(defcommand help
  "Display this help text."
  [subject]
  (if (empty? subject)
    (println (generate-help-text))
    (println (generate-help-text subject))))
