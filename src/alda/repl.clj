(ns alda.repl
  (:require [alda.version                 :refer (-version-)]
            [alda.lisp                    :refer :all]
            [alda.sound                   :refer (set-up! tear-down!)]
            [alda.repl.core               :as    repl :refer (*repl-reader* 
                                                              *parsing-context*)]
            [alda.repl.commands           :refer (repl-command)]
            [io.aviso.ansi                :refer :all]
            [io.aviso.exception           :as    pretty]
            [clojure.string               :as    str])
  (:import  [jline.console ConsoleReader]))

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
  (str (blue ascii-art)
       \newline
       \newline
       (cyan text-below-ascii-art)))

(defmacro while-let
  "Repeatedly executes body while test expression is true. Test
  expression is bound to binding.
  Copied from boot.util."
  [[binding test] & body]
  `(loop [~binding ~test]
     (when ~binding ~@body (recur ~test))))

(defn start-repl! []
  (println)
  (println banner \newline)
  (alter-var-root #'*parsing-context* (constantly :part))
  (alter-var-root #'*repl-reader* (constantly (doto (ConsoleReader.) 
                                                (.setPrompt "> "))))
  (let [done? (atom false)]
    (print "Loading MIDI synth... ")
    (set-up! :midi)
    (println "done." \newline)
    (score*) ; initialize a new score
    (binding [*out* (.getOutput *repl-reader*)]
      (repl/set-prompt!)
      (while-let [alda-code (when-not @done? (.readLine *repl-reader*))]
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
            (pretty/write-exception *err* e)))))))
