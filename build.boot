(set-env!
  :source-paths #{"src" "test"}
  :resource-paths #{"grammar" "examples"}
  :dependencies '[[org.clojure/clojure   "1.7.0"]
                  [org.clojure/tools.cli "0.3.1"]
                  [instaparse            "1.4.1"]
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]
                  [com.taoensso/timbre   "4.1.1"]
                  [clj-http              "2.0.0"]
                  [djy                   "0.1.4"]
                  [str-to-argv           "0.1.0"]
                  [overtone/at-at        "1.2.0"]
                  [midi.soundfont        "0.1.0"]
                  [jline                 "2.12.1"]])

(require '[adzerk.bootlaces :refer :all]
         '[adzerk.boot-test :refer :all]
         '[alda.version]
         '[alda.cli]
         '[alda.lisp        :refer :all]
         '[str-to-argv      :refer (split-args)])

; version number is stored in alda.version
(bootlaces! alda.version/-version-)

(task-options!
  aot {:namespace '#{alda.cli}}
  pom {:project 'alda
       :version alda.version/-version-
       :description "A music programming language for musicians"
       :url "https://github.com/alda-lang/alda"
       :scm {:url "https://github.com/alda-lang/alda"}
       :license {"name" "Eclipse Public License"
                 "url" "http://www.eclipse.org/legal/epl-v10.html"}}
  jar {:main 'alda.cli}
  test {:namespaces '#{
                       ; general tests
                       alda.parser.barlines-test
                       alda.parser.clj-exprs-test
                       alda.parser.event-sequences-test
                       alda.parser.comments-test
                       alda.parser.duration-test
                       alda.parser.events-test
                       alda.parser.octaves-test
                       alda.parser.repeats-test
                       alda.parser.score-test
                       alda.lisp.attributes-test
                       alda.lisp.cram-test
                       alda.lisp.chords-test
                       alda.lisp.duration-test
                       alda.lisp.global-attributes-test
                       alda.lisp.markers-test
                       alda.lisp.notes-test
                       alda.lisp.parts-test
                       alda.lisp.pitch-test
                       alda.lisp.score-test
                       alda.lisp.voices-test
                       alda.util-test

                       ; benchmarks / smoke tests
                       alda.parser.examples-test
                       }})

(deftask alda
  "Run Alda CLI tasks.

   Whereas running `bin/alda <cmd> <args>` will use the latest deployed
   version of Alda, running this task (`boot alda -x '<cmd> <args>'`)
   will use the current (local) version of this repo."
  [x execute ARGS str "The Alda CLI task and args as a single string."]
  (fn [next-task]
    (fn [fileset]
      (require '[alda.version] :reload-all)
      (require '[alda.repl] :reload-all)
      ;(require '[alda.cli] :reload-all)
      (when execute
        (let [cli-args (split-args execute)]
          (apply (resolve 'alda.cli/-main) cli-args)))
      (next-task fileset))))
