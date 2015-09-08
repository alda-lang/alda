(set-env!
  :source-paths #{"src" "test"}
  :resource-paths #{"grammar"}
  :dependencies '[[org.clojure/clojure   "1.7.0"]
                  [org.clojure/tools.cli "0.3.1"]
                  [instaparse            "1.4.1"]
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]
                  [com.taoensso/timbre   "4.1.1"]
                  [djy                   "0.1.4"]
                  [str-to-argv           "0.1.0"]
                  [overtone              "0.9.1"]
                  [midi.soundfont        "0.1.0"]
                  [reply                 "0.3.7"]])

(require '[adzerk.bootlaces :refer :all]
         '[adzerk.boot-test :refer :all]
         '[alda.version]
         '[alda.cli]
         '[str-to-argv :refer (split-args)])

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
  test {:namespaces '#{alda.test.parser.attributes
                       alda.test.parser.comments
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

(deftask alda
  "Run Alda CLI tasks.
   
   Whereas running `bin/alda <cmd> <args>` will use the latest deployed 
   version of Alda, running this task (`boot alda -x '<cmd> <args>'`)
   will use the current (local) version of this repo."
  [x execute ARGS str "The Alda CLI task and args as a single string."]
  (when execute
    (let [cli-args (split-args execute)]
      (apply (resolve 'alda.cli/-main) cli-args))))
