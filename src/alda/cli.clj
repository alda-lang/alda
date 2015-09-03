(ns alda.cli
  (:require [boot.cli     :refer (defclifn)]
            [boot.core    :refer (merge-env!)]
            [alda.parser  :refer (parse-input)]
            [alda.repl]))

(defn fluid-r3! 
  "Fetches FluidR3 dependency and returns the input stream handle."
  []
  (eval
    '(do (merge-env!
           :dependencies '[[org.bitbucket.daveyarwood/fluid-r3 "0.1.1"]])
         (require '[midi.soundfont.fluid-r3 :as fluid-r3])
         fluid-r3/sf2)))

(defclifn parse
  "Parse some Alda code and print the results to the console."
  [f file FILE str  "The path to a file containing Alda code."
   c code CODE str  "A string of Alda code."
   l lisp      bool "Parse into alda.lisp code."
   m map       bool "Evaluate the score and show the resulting instruments/events map."]
  (let [alda-lisp-code (parse-input (if code code (slurp file)))]
    (when lisp
      (prn alda-lisp-code))
    (when map
      (require 'alda.lisp)
      (println)
      (prn (eval alda-lisp-code)))))

(defclifn play
  "Parse some Alda code and play the resulting score."
  [f file        FILE str "The path to a file containing Alda code."
   c code        CODE str "A string of Alda code."
   ; TODO: implement smart buffering and remove the buffer options
   p pre-buffer  MS  int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS  int  "The number of milliseconds to keep the synth open after the score ends. (default: 1000)"
   s stock           bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (require '[alda.lisp]
           '[alda.sound]
           '[instaparse.core])
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))
            alda.sound/*play-opts* {:pre-buffer  (or pre-buffer 0)
                                    :post-buffer (or post-buffer 1000)
                                    :one-off?    true}]
    (let [parsed (parse-input (if code code (slurp file)))]
      (if (instaparse.core/failure? parsed)
        (prn parsed) 
        (alda.sound/play! (eval parsed)))
      identity)))

(defclifn alda-repl
  "Starts an Alda Read-Evaluate-Play-Loop."
  [p pre-buffer  MS int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS int  "The number of milliseconds to wait after the score ends. (default: 0)"
   s stock          bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))
            alda.sound/*play-opts* {:pre-buffer  pre-buffer
                                    :post-buffer post-buffer
                                    :async?      true}]
    (alda.repl/start-repl!)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(def help-text
  "alda help text - TODO")

(defn- delegate
  [cmd args]
  (if (empty? args)
    (cmd "--help")
    (apply cmd args)))

(defn -main [& [cmd & args]]
  (case cmd
    nil      (println help-text)
    "help"   (println help-text)
    "--help" (println help-text)
    "parse"  (delegate parse args)
    "play"   (delegate play args)
    "repl"   (delegate alda-repl args)
    (printf "[alda] Invalid command '%s'.\n\n%s\n" cmd)))
