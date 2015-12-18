(ns alda.cli
  (:require [boot.cli        :refer (defclifn)]
            [boot.core       :refer (merge-env!)]
            [clojure.string  :as    str]
            [clojure.pprint  :refer (pprint)]
            [clj-http.client :as    client]
            [alda.parser     :refer (parse-input)]
            [alda.version    :refer (-version-)]
            [alda.sound]
            [alda.util       :as    util]))

(defn fluid-r3!
  "Fetches FluidR3 dependency and returns the input stream handle."
  []
  (eval
    '(do (merge-env!
           :dependencies '[[org.bitbucket.daveyarwood/fluid-r3 "0.1.1"]])
         (require '[midi.soundfont.fluid-r3 :as fluid-r3])
         fluid-r3/sf2)))

(defclifn ^:alda-task repl
  "Start an Alda Read-Evaluate-Play-Loop."
  [p pre-buffer  MS int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS int  "The number of milliseconds to wait after the score ends. (default: 0)"
   s stock          bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))
            alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :async?      true}]
    (eval
      '(do
         (require '[alda.repl])
         (alda.repl/start-repl!)))))

(defclifn ^:alda-task server
  "Start an Alda server."
  [p port        PORT int  "The port on which to run the Alda server."
   b pre-buffer  MS   int  "The number of milliseconds of lead time for buffering. (default: 0)"
   B post-buffer MS   int  "The number of milliseconds to wait after the score ends. (default: 0)"
   s stock            bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))
            alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :async?      true}]
    (require 'alda.server)
    ((resolve 'alda.server/start-server!) (or port 27713))))

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

(defn main [[cmd & args]]
  (util/set-timbre-level!)
  (case cmd
    nil         (println help-text)
    "help"      (println help-text)
    "--help"    (println help-text)
    "-h"        (println help-text)
    "version"   (printf "alda v%s\n" -version-)
    "--version" (printf "alda v%s\n" -version-)
    "-v"        (printf "alda v%s\n" -version-)
    "repl"      (delegate repl args)
    "server"    (delegate server args)
    (println (format "[alda] Invalid command '%s'.\n" cmd))))
