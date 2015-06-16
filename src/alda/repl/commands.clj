(ns alda.repl.commands
  (:require [alda.lisp      :refer :all]
            [clojure.string :as    str]))

(defn huh? []
  (println "Sorry, what? I don't understand that command."))

(defmulti repl-command (fn [cmd rest-of-line] cmd))

(defmethod repl-command :default [_ _] 
  (huh?))

(defmethod repl-command "new" [_ rest-of-line]
  (let [trimmed (str/trim rest-of-line)]
    (cond
      (contains? #{"" "score"} trimmed)
      (do
        (score*)
        (println "New score initialized."))

      (.startsWith trimmed "part ")
      :TODO

      :else
      (huh?))))
