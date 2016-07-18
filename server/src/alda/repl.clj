(ns alda.repl
  (:require [alda.version       :refer (-version-)]
            [alda.lisp          :refer :all]
            [alda.sound         :as    sound]
            [alda.sound.midi    :as    midi]
            [alda.repl.core     :as    repl :refer (*repl-reader*
                                                    *current-score*
                                                    score-text<<)]
            [alda.repl.commands :refer (repl-command)]
            [alda.util          :refer (while-let)]
            [io.aviso.ansi      :refer :all]
            [io.aviso.exception :as    pretty]
            [clojure.string     :as    str]
            [taoensso.timbre    :as    log])
  (:import  [jline.console ConsoleReader]
            [jline.console.completer Completer]))

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

(defn print-banner!
  []
  (println)
  (println banner))

(defn prepare-midi-system!
  []
  (print "Preparing MIDI system... ") (flush)
  (midi/fill-midi-synth-pool!)
  (while (not (midi/midi-synth-available?))
    (Thread/sleep 250))
  (println "done.")
  (Thread/sleep 400))

(defn start-repl! [& {:keys [pre-buffer post-buffer]}]
  (prepare-midi-system!)
  (print-banner!)
  (let [done? (atom false)]
    (binding [*out*             (.getOutput ^ConsoleReader *repl-reader*)
              sound/*play-opts* {:pre-buffer  pre-buffer
                                 :post-buffer post-buffer
                                 :async?      true}]
      (repl/refresh-prompt!)
      (require '[alda.lisp :refer :all])
      (while-let [alda-code (when-not @done?
                              (println)
                              (.readLine ^ConsoleReader *repl-reader*))]
        (try
          (cond
            (re-find #"^\s*$" alda-code)
            :do-nothing

            (re-find #"^:?(quit|exit|bye)\s*$" alda-code)
            (do
              (sound/tear-down! @*current-score*)
              (reset! done? true))

            (re-find #"^:" alda-code)
            (let [[_ cmd rest-of-line] (re-matches #":(\S+)\s*(.*)" alda-code)]
              (repl-command cmd (str/trim rest-of-line)))

            :else
            (when (repl/interpret! alda-code)
              (score-text<< alda-code)))
          (repl/refresh-prompt!)
          (catch Throwable e
            (pretty/write-exception *err* e))))))
  (System/exit 0))

