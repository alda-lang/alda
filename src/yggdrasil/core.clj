(ns yggdrasil.core
  (:use [clojure.tools.cli :only (cli)])
  (:require [yggdrasil.sound-generator :as soundgen])
  (:gen-class :main true))

(defn dispatcher
  "If no output file is specified, plays the file using the specified options. 
   If an output file is specified, creates a .wav file using the specified options."
  ([input-file opts]
    (soundgen/play input-file opts))
  ([input-file output-file opts]
    (soundgen/make-wav input-file output-file opts)))

(defn -main
  "Parses command line arguments and dispatches the appropriate functions."
  [& args]
  (let [[opts args banner]
        (cli args
         ["-h" "--help" 
               "Display help text." 
               :flag true, :default false]
         ["-s" "--start" 
               "Start playing the score at a specific minute/second mark or score marker." 
               :default "0:00"]
         ["-e" "--end" 
               "Stop playing at a specific minute/second mark or score marker."] ;; optional
         )]
    (when (:help opts)
      (println banner)
      (System/exit 0))
    (try
      (apply dispatcher (conj args opts))
      (catch clojure.lang.ArityException e
        (cond 
          (zero? (count args)) 
          (println "Please specify an input file containing some yggdrasil code."
                   "\n\n" 
                   "example:    ygg chorale.yd")
          (> (count args) 2) 
          (println "Invalid number of arguments. You must specify only one input file and"
                   "(optionally) one output file."
                   "\n\n"
                   "example:    ygg chorale.yd chorale.wav")))
      (catch java.io.FileNotFoundException e
        (let [bad-filename (first args)]
          (println (format "Input file \"%s\" not found." bad-filename)))))))
