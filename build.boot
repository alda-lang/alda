#!/usr/bin/env boot

(set-env!
 :source-paths #{"src" "test"}
 :resource-paths #{"grammar"}
 :dependencies '[[org.clojure/clojure   "1.7.0"]
                 [org.clojure/tools.cli "0.3.1"]
                 [instaparse            "1.4.1"]
                 [adzerk/bootlaces      "0.1.11" :scope "test"]
                 [adzerk/boot-test      "1.0.4"  :scope "test"]
                 [com.taoensso/timbre   "3.4.0"]
                 [djy                   "0.1.4"]
                 [overtone              "0.9.1"]
                 [midi.soundfont        "0.1.0"]
                 [reply                 "0.3.7"]
                 [defun                 "0.2.0-RC"]])

(require '[adzerk.bootlaces :refer :all]
         '[adzerk.boot-test :refer :all]
         '[alda.core]
         '[alda.parser :refer (parse-input)]
         '[alda.repl])

(def +version+ "0.1.0")
(bootlaces! +version+)

(task-options!
  aot {:namespace '#{alda.core}}
  pom {:project 'alda
       :version +version+
       :description "A music programming language for musicians"
       :url "https://github.com/alda-lang/alda"
       :scm {:url "https://github.com/alda-lang/alda"}
       :license {"name" "Eclipse Public License"
                 "url" "http://www.eclipse.org/legal/epl-v10.html"}}
  jar {:main 'alda.core}
  test {:namespaces '#{alda.test.parser.attributes
                       alda.test.parser.duration
                       alda.test.parser.events
                       alda.test.parser.score
                       alda.test.lisp.attributes
                       alda.test.lisp.chords
                       alda.test.lisp.duration
                       alda.test.lisp.global-attributes
                       alda.test.lisp.markers
                       alda.test.lisp.notes
                       alda.test.lisp.parts
                       alda.test.lisp.pitch
                       alda.test.lisp.score
                       alda.test.lisp.voices}})

(deftask build
  "Builds uberjar.
   TODO: be able to build an executable à la lein bin"
  []
  (comp (aot) (pom) (uber) (jar)))

(deftask parse
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

(defn fluid-r3! 
  "Fetches FluidR3 dependency and returns the input stream handle."
  []
  (clojure.core/eval
    '(do (merge-env!
           :dependencies '[[org.bitbucket.daveyarwood/fluid-r3 "0.1.1"]])
         (require '[midi.soundfont.fluid-r3 :as fluid-r3])
         fluid-r3/sf2)))

(deftask play
  "Parse some Alda code and play the resulting score."
  [f file        FILE str "The path to a file containing Alda code."
   c code        CODE str "A string of Alda code."
   ; TODO: implement smart buffering and remove the buffer options
   p pre-buffer  MS  int  "The number of milliseconds of lead time for buffering. (default: 4000)"
   P post-buffer MS  int  "The number of milliseconds to keep the synth open after the score ends. (default: 4000)"
   s stock           bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (require '[alda.lisp]
           '[alda.sound])
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))]
    (alda.sound/play! (eval (parse-input (if code code (slurp file))))
                      {:pre-buffer  (or pre-buffer  4000)
                       :post-buffer (or post-buffer 4000)
                       :one-off?    true})))

(deftask alda-repl
  "Starts an Alda Read-Evaluate-Play-Loop."
  [p pre-buffer  MS int  "The number of milliseconds of lead time for buffering. (default: 0)"
   P post-buffer MS int  "The number of milliseconds to wait after the score ends. (default: 0)"
   s stock          bool "Use the default MIDI soundfont of your JVM, instead of FluidR3."]
  (binding [alda.sound.midi/*midi-soundfont* (when-not stock (fluid-r3!))]
    (alda.repl/start-repl! +version+ {:pre-buffer  pre-buffer
                                      :post-buffer post-buffer})))

(defn -main [& args]
  (apply alda.core/-main args))
