(ns alda.cli
  (:require [boot.cli        :refer (defclifn)]
            [boot.core       :refer (merge-env!)]
            [clojure.string  :as    str]
            [clojure.pprint  :refer (pprint)]
            [clj-http.client :as    client]
            [alda.parser     :refer (parse-input)]
            [alda.version    :refer (-version-)]
            [alda.sound]
            [alda.util       :as    util])
  (:gen-class))

(defclifn ^:alda-task parse
  "Parse some Alda code and print the results to the console."
  [f file FILE str  "The path to a file containing Alda code to parse."
   c code CODE str  "The string of Alda code to parse."
   l lisp      bool "Parse into alda.lisp code."
   m map       bool "Evaluate the score and show the resulting instruments/events map."]
  (if-not (or file code)
    (parse "--help")
    (try
      (let [input (if code code (slurp file))
            alda-lisp-code (parse-input input)]
        (when lisp
          (pprint alda-lisp-code)
          (println))
        (when map
          (require 'alda.lisp)
          (pprint (eval alda-lisp-code))
          (println)))
      (catch Exception e
        (println (.getMessage e))
        (System/exit 1)))))

(defclifn ^:alda-task play
  "Parse some Alda code and play the resulting score."
  [f file        FILE str "The path to a file containing Alda code to play."
   c code        CODE str "The string of Alda code to play."
   ; TODO: implement smart buffering and remove the buffer options
   p pre-buffer  MS  int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS  int  "The number of milliseconds to keep the synth open after the score ends. (default: 1000)"
   F from        POS str  "Position to start playback from"
   T to          POS str  "Position to end playback at"]
  (require '[alda.lisp]
           '[instaparse.core])
  (binding [alda.sound/*play-opts* {:pre-buffer  (or pre-buffer 0)
                                    :post-buffer (or post-buffer 1000)
                                    :from        from
                                    :to          to
                                    :one-off?    true}]
    (if-not (or file code)
      (play "--help")
      (try
        (let [parsed (parse-input (if code code (slurp file)))]
          (alda.sound/play! (eval parsed))
          (System/exit 0))
        (catch Exception e
          (println (.getMessage e))
          (System/exit 1))))))

(defclifn ^:alda-task repl
  "Start an Alda Read-Evaluate-Play-Loop."
  [p pre-buffer  MS int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS int  "The number of milliseconds to wait after the score ends. (default: 0)"]
  (binding [alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :async?      true}]
    (eval
      '(do
         (require '[alda.repl])
         (alda.repl/start-repl!)))))

(defclifn ^:alda-task script
  "Print the latest `alda` start script to STDOUT."
  []
  (let [script-url "https://raw.githubusercontent.com/alda-lang/alda/master/bin/alda"
        script (:body (client/get script-url))]
    (println script)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(def alda-tasks
  (into {'help    "Display this help text."
         'version "Display Alda version number."}
    (for [[sym var] (ns-publics *ns*)
          :when (:alda-task (meta var))
          :let [doc (:doc (meta var))
                help-blurb (apply str (take-while (partial not= \newline) doc))]]
      [sym help-blurb])))

(def help-text
  (format (str "alda v%s\n\nUsage:\n\n    alda <task> <options>\n\n"
               "To see options for a task:\n\n    alda <task> --help\n\n"
               "Tasks:\n\n%s")
          -version-
          (str/join \newline
                    (for [[task blurb] alda-tasks]
                      (str "    " task \tab blurb)))))

(defn- delegate
  [cmd args]
  (if (empty? args)
    (cmd "")
    (apply cmd args)))

(defn -main [& [cmd & args]]
  (util/set-timbre-level!)
  (case cmd
    nil         (println help-text)
    "help"      (println help-text)
    "--help"    (println help-text)
    "-h"        (println help-text)
    "version"   (printf "alda v%s\n" -version-)
    "--version" (printf "alda v%s\n" -version-)
    "-v"        (printf "alda v%s\n" -version-)
    "parse"     (delegate parse args)
    "play"      (delegate play args)
    "repl"      (delegate repl args)
    "script"    (delegate script args)
    (do
      (printf "[alda] Invalid command '%s'.\n" cmd)
      (System/exit 1)))
  (System/exit 0))
