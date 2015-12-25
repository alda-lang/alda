(ns alda.repl
  (:require [alda.version                 :refer (-version-)]
            [alda.lisp                    :refer :all]
            [alda.sound                   :refer (set-up! tear-down! *play-opts*)]
            [alda.repl.core               :as    repl :refer (*repl-reader*
                                                              *parsing-context*)]
            [alda.repl.commands           :refer (repl-command)]
            [alda.util]
            [boot.from.io.aviso.ansi      :refer :all]
            [boot.from.io.aviso.exception :as    pretty]
            [boot.util                    :refer (while-let)]
            [clojure.string               :as    str]
            [taoensso.timbre              :as    log])
  (:import  [jline.console ConsoleReader]
            [jline.console.completer Completer]))

; sets log level to TIMBRE_LEVEL (if set) or :warn
(alda.util/set-timbre-level!)

(def ascii-art
  (str " █████╗ ██╗     ██████╗  █████╗ " \newline
       "██╔══██╗██║     ██╔══██╗██╔══██╗" \newline
       "███████║██║     ██║  ██║███████║" \newline
       "██╔══██║██║     ██║  ██║██╔══██║" \newline
       "██║  ██║███████╗██████╔╝██║  ██║" \newline
       "╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝"))

(def text-below-ascii-art
  (str "            v" -version- \newline
       "         repl session"))

(def banner
  (str (blue ascii-art) \newline
       \newline
       (cyan text-below-ascii-art) \newline
       \newline
       (bold-white "Type :help for a list of available commands.")))

(defn start-repl! [& {:keys [pre-buffer post-buffer]}]
  (println)
  (println banner \newline)
  (alter-var-root #'*parsing-context* (constantly :part))
  (alter-var-root #'*repl-reader* (constantly (doto (ConsoleReader.)
                                                (.setExpandEvents false)
                                                (.setPrompt "> "))))
  (let [done? (atom false)]
    (print "Loading MIDI synth... ") (flush)
    (set-up! :midi)
    (println "done.")
    (score*) ; initialize a new score
    (binding [*out*       (.getOutput *repl-reader*)
              *play-opts* {:pre-buffer  pre-buffer
                           :post-buffer post-buffer
                           :async?      true}]
      (repl/set-prompt!)
      (require '[alda.lisp :refer :all])
      (while-let [alda-code (when-not @done?
                              (println)
                              (.readLine *repl-reader*))]
        (try
          (cond
            (re-find #"^\s*$" alda-code)
            :do-nothing

            (re-find #"^:?(quit|exit|bye)" alda-code)
            (do
              (tear-down! :midi)
              (reset! done? true))

            (re-find #"^:" alda-code)
            (let [[_ cmd rest-of-line] (re-matches #":(\S+)\s*(.*)" alda-code)]
              (repl-command cmd (str/trim rest-of-line)))

            :else
            (when (repl/interpret! alda-code) (score-text<< alda-code)))
          (repl/set-prompt!)
          (catch Throwable e
            (pretty/write-exception *err* e))))))
  (System/exit 0))

