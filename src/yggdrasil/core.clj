(ns yggdrasil.core
	(:use [clojure.tools.cli :only (cli)])
	(:import (java.io File))
	(:gen-class :main true))

(defn -main
	"Parses command line arguments and dispatches the appropriate functions."
	[& args]
	(let [[opts args banner]
		(cli args
			["-h" "--help" 
				"Display help text." 
				:flag true, :default false]
			["-s" "--start" 
				"Start playing the score at a specific minute/second mark." 
				:default "0:00"]
			["-e" "--end" 
				"Stop playing at a specific minute/second mark in the score."] ;; optional
			)]
	   (when (:help opts)
	   	   (println banner)
	   	   (System/exit 0))
	   (try
		   (apply parse-and-dispatch (conj args opts))
		   (catch ArityException e
			   (cond 
				   (zero? (count args)) 
				   (println "Please specify an input file containing some yggdrasil code."
					    "\n\n" 
					    "example:    ygg chorale.yg")
				   (> (count args) 2) 
				   (println "Invalid number of arguments. You must specify only one input"
					    " file and (optionally) one output file."
					    "\n\n"
					    "example:    ygg chorale.yg chorale.wav")))
		   (catch FileNotFoundException e
			   (let [bad-filename (first args)]
			      (println "Input file \"" bad-filename "\" not found."))))))

(defn parse-and-dispatch
	([input-file {:keys [start end]}]
		(let [wav-data (ygg-to-wav input-file)]
			(comment
			"Code that plays the wav-data. Short term goal is just to get this working.
			I'd eventually like to implement a slick CLI music player display that shows
			how long the music is in minutes/seconds, how far into it you are with a
			progress bar, and so on. Maybe even listen for keyboard input and allow
			the user to pause, stop, restart, jump to a specific time marking or named
			marker, etc. Could eventually expand this into a full-blown GUI."))))
	([input-file output-file {:keys [start end]}]
		(let [wav-data (ygg-to-wav input-file)
		      target-file (check-for output-file)]
			(comment
			"Code that writes wav-data to target-file, prints something letting the user 
			know it has done so, then exits happily."))))

(defn check-for
	"Checks to see if a given file already exists. If it does, prompts the user whether 
	he/she wants to overwrite the file. If he/she doesn't, then prompts the user to 
	choose a new filename and calls itself to check the new file, etc. Returns a filename 
	that either does not exist, or does exist and the user is OK with overwriting it."
	[filename]
	(letfn [(prompt [] 
			(print "> ") (read-line))
		(overwrite-dialog []
			(println "File \"" filename "\" already exists. Overwrite? (y/n)")
				(let [response (prompt)]
					(cond
						(some #{response} ["y" "yes" "Y" "YES"])
						filename

						(some #{response} ["n" "no" "N" "NO"])
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
				(println "\"" filename "\" is a directory. Please specify a filename.")
				(recur (prompt)))

			:else filename)))

(defn ygg-to-wav
	"Parses ygg code and returns wav data."
	[ygg-code]
	(comment "To do. 
	          Consider putting this in a separate namespace?"))