(ns alda.util
  (:require [clojure.string  :as str]
            [taoensso.timbre :as timbre])
  (:import (java.io File)))

(defn set-timbre-level!
  []
  (timbre/set-level! (if-let [level (System/getenv "TIMBRE_LEVEL")]
                       (keyword (str/replace level #":" ""))
                       :warn)))

(defn check-for
  "Checks to see if a given file already exists. If it does, prompts the user
   whether he/she wants to overwrite the file. If he/she doesn't, then prompts
   the user to choose a new filename and calls itself to check the new file, etc.
   Returns a filename that does not exist, or does exist and the user is OK with
   overwriting it."
  [filename]
  (letfn [(prompt []
            (print "> ") (flush) (read-line))
          (overwrite-dialog []
            (println
              (format "File \"%s\" already exists. Overwrite? (y/n)" filename))
            (let [response (prompt)]
              (cond
                (re-find #"(?i)y(es)?" response)
                filename

                (re-find #"(?i)no?" response)
                (do
                  (println "Please specify a different filename.")
                  (check-for (prompt)))

               :else
               (do
                 (println "Answer the question, sir.")
                 (recur)))))]
    (cond
      (.isFile (File. filename))
      (overwrite-dialog)

      (.isDirectory (File. filename))
      (do
        (printf "\"%s\" is a directory. Please specify a filename.\n" filename)
        (recur (prompt)))

      :else
      filename)))
