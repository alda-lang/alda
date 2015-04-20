(ns alda.repl
  (:require [alda.parser :refer (parse-input)]
            [alda.lisp]
            [alda.sound :refer (play!)]
            [instaparse.core :as insta]
            [boot.from.io.aviso.ansi :refer :all]))

(def ascii-art "
 █████╗ ██╗     ██████╗  █████╗
██╔══██╗██║     ██╔══██╗██╔══██╗
███████║██║     ██║  ██║███████║
██╔══██║██║     ██║  ██║██╔══██║
██║  ██║███████╗██████╔╝██║  ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝
")

(defn banner [version]
  (format (str (blue ascii-art) (cyan "
            v%s
         repl session
")) version))

(defn start-repl
  [version]
  (println (banner version))
  (loop [score ""]
    (print "> ")
    (flush)
    (let [alda-code (read-line)]
      (if (contains? #{"quit" "exit" "bye"} alda-code)
        (println "score:" \newline score)
        (let [parsed (parse-input alda-code)]
          (if (insta/failure? parsed)
            (do
              (println "Invalid Alda syntax.")
              (recur score))
            (do
              (play! (eval parsed))
              (recur (str score \newline alda-code)))))))))
