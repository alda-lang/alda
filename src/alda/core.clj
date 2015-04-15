(ns alda.core
  (:require [clojure.tools.cli :refer (parse-opts)]
            [alda.sound :as sound])
  (:gen-class))

(def cli-options
  [["-h" "--help" "Display help text."]
   ["-s" "--start START"
    "Start reading the score at a specific minute/second mark or score marker."
    :default "0:00"]
    ; TODO: write :parse-fn and :validate functions (use for both start & end)
   ["-e" "--end END"
    "Stop playing the score at a specific minute/second mark or score marker."]])

(defn dispatcher
  "If no output file is specified, plays the file using the specified options.
   If an output file is specified, creates a .wav file using the specified options."
  ([input-file opts]
    (sound/play! input-file opts))
  ([input-file output-file opts]
    (sound/make-wav! input-file output-file opts)))

(defn -main
  [& args]
  (let [{:keys [options arguments errors summary]} (parse-opts args cli-options)]
    (cond
      (:help options)
      (do
        (println summary)
        (System/exit 0))
      (zero? (count args))
      (do
        (println "Please specify an input file containing some alda code."
                 "\n\nexample:    alda chorale.alda")
        (System/exit 1))
      (> (count args) 2)
      (do
        (println "Invalid number of arguments. You must specify only one input file and"
                 "(optionally) one output file."
                 "\n\nexample:    alda chorale.alda chorale.wav")
        (System/exit 1))
      :else
      (try
        (apply dispatcher (conj arguments options))
        (catch java.io.FileNotFoundException e
          (let [bad-filename (first arguments)]
            (println (format "Input file \"%s\" not found." bad-filename))))))))
