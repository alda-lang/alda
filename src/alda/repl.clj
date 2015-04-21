(ns alda.repl
  (:require [alda.parser                  :refer (parse-input)]
            [alda.lisp]
            [alda.sound                   :refer (play!)]
            [instaparse.core              :as    insta]
            [boot.from.io.aviso.ansi      :refer :all]
            [boot.from.io.aviso.exception :as    pretty]
            [boot.util                    :refer (while-let)])
  (:import  [jline.console ConsoleReader]
            [jline.console.completer Completer]))

(def ascii-art
  (str " █████╗ ██╗     ██████╗  █████╗ " \newline
       "██╔══██╗██║     ██╔══██╗██╔══██╗" \newline
       "███████║██║     ██║  ██║███████║" \newline
       "██╔══██║██║     ██║  ██║██╔══██║" \newline
       "██║  ██║███████╗██████╔╝██║  ██║" \newline
       "╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝"))

(defn text-below-ascii-art [version]
  (str "            v" version \newline
       "         repl session"))

(defn banner [version]
  (str (blue ascii-art)
       \newline
       \newline
       (cyan (text-below-ascii-art version))))

(defn start-repl
  [version]
  (println)
  (println (banner version) \newline)
  (let [done?  (atom false)
        score  (atom "")
        reader (doto (ConsoleReader.) (.setPrompt "> "))]
    (binding [*out* (.getOutput reader)]
      (while-let [alda-code (when-not @done? (.readLine reader))]
        (try
          (cond
            (re-find #"^\s*$" alda-code)
              :do-nothing
            (re-find #"^(quit|exit|bye)" alda-code)
              (do
                (println)
                (println "score:" \newline @score)
                (reset! done? true))
            :else
              (let [parsed (parse-input alda-code)]
                (if (insta/failure? parsed)
                  (println "Invalid Alda syntax.")
                  (do
                    (play! (eval parsed))
                    (swap! score str \newline alda-code)))))
          (catch Throwable e
            (pretty/write-exception *err* e)))))))
