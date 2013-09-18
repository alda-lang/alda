(ns yggdrasil.core
	(:use [clojure.tools.cli :only (cli)])
	(:gen-class :main true))

(defn -main
	"Parses command line arguments and calls the appropriate functions."
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
	   (cond 
	   	   (zero? (count args)) (comment 
	   	   	   "Error message about needing to specify a .yg file, and exit.")
	   	   (> 2 (count args)) (comment
	   	   	   "Error message about > 2 arguments. Remind user of syntax and exit."))	   
	   (let [wav-data (comment 
	   		   "Code for parsing the .yg file, which should be (first args).
	   		   Should first check that the file exists and return an error message if
	   		   the file is not found.
	   		   This should return a working wav-data object of some sort?")]
	      (if (= 2 (count args)) 
	      	  (comment 
	      	  	  "Code that checks (last args), which should be the destination .wav file name,
	      	  	  and sees if that file already exists in the current working directory. 
	      	  	  If so, prompts the user if he/she wants to overwrite the file.
	      	  	  Once a good file name is established, writes wav-data to the file, prints 
	      	  	  something letting the user know it has done so, then exits happily.")
	      	  (comment
	      	  	  "Code that plays the wav-data. Short term goal is just to get this working.
	      	  	  I'd eventually like to implement a slick CLI music player display that shows
	      	  	  how long the music is in minutes/seconds, how far into it you are with a
	      	  	  progress bar, and so on. Maybe even listen for keyboard input and allow
	      	  	  the user to pause, stop, restart, jump to a specific time marking or named
	      	  	  marker, etc. Could eventually expand this into a full-blown GUI.")))))
