(ns alda.cli
  (:require [clojure.pprint  :refer (pprint)]
            [clojure.string  :as    str]
            [clojure.tools.cli :refer (parse-opts summarize)]
            [alda.parser     :refer (parse-input parse-tree)]
            [alda.repl       :refer (start-repl!)]
            [alda.version    :refer (-version-)]
            [alda.sound      :refer (play!)]
            [alda.util       :as    util]
            [instaparse.core :as    insta])
  (:gen-class))

(def cli-options
  [["-h" "--help" "Display this help text."]
   ["-v" "--version" "Display Alda version number."]])

(def help-option ["-h" "--help" "Display the help text for this task."])
(def file-option ["-f" "--file FILE" "The path to a file containing Alda code to play."])
(def code-option ["-c" "--code CODE" "The string of Alda code to play."])

(def pre-buffer-option
  ["-p" "--pre-buffer N" "The number of milliseconds of lead time for buffering."
   :default 0])

(def post-buffer-option ["-P" "--post-buffer N" "The number of milliseconds to keep the synth open after the score ends."])

(def play-options
  [help-option
   file-option
   code-option
   pre-buffer-option
   (conj post-buffer-option :default 1000)
   ["-F" "--from N" "Position to start playback from."]
   ["-T" "--to N" "Position to end playback at."]])

(defn ^:alda-task play
  "Parse some Alda code and play the resulting score."
  [file code pre-buffer post-buffer from to]
  (binding [alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :from        from
                                    :to          to
                                    :one-off?    true}]
    (let [parsed (parse-input (if code code (slurp file)))]
      (if (insta/failure? parsed)
        (do (prn parsed) (System/exit 1))
        (play! (eval parsed)))
      identity)))

(def parse-options
  [help-option
   file-option
   code-option
   ["-t" "--tree" "Show the intermediate parse tree."]
   ["-l" "--lisp" "Parse into alda.lisp code."]
   ["-m" "--map" "Evaluate the score and show the resulting instruments/events map."]])

(defn ^:alda-task parse
  "Parse some Alda code and print the results to the console."
  [file code tree lisp]
  (let [input (if code code (slurp file))
        alda-lisp-code (parse-input input)]
    (when (insta/failure? alda-lisp-code)
      (pprint alda-lisp-code)
      (System/exit 1))
    (when tree
      (pprint (parse-tree input))
      (println))
    (when lisp
      (pprint alda-lisp-code)
      (println))
    (when map
      (require 'alda.lisp)
      (pprint (eval alda-lisp-code))
      (println))))

(def repl-options
  [help-option
   file-option
   code-option
   pre-buffer-option
   (conj post-buffer-option :default 0)])

(defn ^:alda-task repl
  "Start an Alda Read-Evaluate-Play-Loop."
  [pre-buffer post-buffer]
  (binding [alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :async?      true}]
    (start-repl!)))

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

(defmacro do-and-exit-with
  [code b]
  `(do
     ~b
     (System/exit ~code)))

(defn- do-and-exit-with-success [b] (do-and-exit-with 0 b))
(defn- do-and-exit-with-error [b] (do-and-exit-with 1 b))

(defn command-options
  [args opts]
  (let [{:keys [options summary]} (parse-opts (drop 1 args) opts :no-defaults true)]
    (cond
      (:help options) (do-and-exit-with-success (println summary))
      :else [options summary])))

(defn -main [& args]
  (util/set-timbre-level!)
  (let [{:keys [options arguments summary]} (parse-opts args cli-options :no-defaults true)]
    (cond
      (and (contains? options :help) (empty? arguments)) ; Generic help, as opposed to per task help
      (do-and-exit-with-error (println help-text))
      (contains? options :version) (do-and-exit-with-success (println (format "alda v%s" -version-)))
      :else
      (case (first arguments)
        "help"
        (do-and-exit-with-error (println help-text))
        "version"
        (do-and-exit-with-error (println (format "alda v%s" -version-)))
        "parse"
        (let [[{:keys [file code tree lisp]} summary] (command-options args parse-options)]
          (if (or file code)
            (parse file code tree lisp)
            (do-and-exit-with-error (println summary))))
        "play"
        (let [[{:keys [file code pre-buffer post-buffer from to]} summary] (command-options args play-options)]
          (if (or file code)
            (play file code pre-buffer post-buffer from to)
            (do-and-exit-with-error (println summary))))
        "repl"
        (let [options (command-options args repl-options)]
          (repl (:pre-buffer options) (:post-buffer options)))
        (do-and-exit-with-error (println (format "[alda] Invalid command '%s'.\n\n%s\n" (first arguments) summary)))))))