(set-env!
  :source-paths #{"client/src"}
  :resource-paths #{"server/src" "server/test"
                    "server/grammar" "examples"}
  :dependencies '[
                  ; server
                  [org.clojure/clojure   "1.7.0"]
                  [org.clojure/tools.cli "0.3.1"]
                  [boot/core             "2.4.2"]
                  [boot/base             "2.4.2"]
                  [instaparse            "1.4.1"]
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-jar2bin   "1.0.0"  :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]
                  [com.taoensso/timbre   "4.1.1"]
                  [clj-http              "2.0.0"]
                  [ring                  "1.4.0"]
                  [ring/ring-defaults    "0.1.5"]
                  [compojure             "1.4.0"]
                  [djy                   "0.1.4"]
                  [str-to-argv           "0.1.0"]
                  [overtone/at-at        "1.2.0"]
                  [midi.soundfont        "0.1.1"]
                  [jline                 "2.12.1"]

                  ; client
                  [com.beust/jcommander                 "1.48"]
                  [org.fusesource.jansi/jansi           "1.11"]
                  [org.apache.httpcomponents/httpclient "4.5.1"]
                  [net.jodah/recurrent                  "0.4.0"]
                  ])

(require '[adzerk.bootlaces    :refer :all]
         '[adzerk.boot-jar2bin :refer :all]
         '[adzerk.boot-test    :refer :all]
         '[alda.util]
         '[alda.version])

; sets log level to TIMBRE_LEVEL (if set) or :warn
(alda.util/set-timbre-level!)

; version number is stored in alda.version
(bootlaces! alda.version/-version-)

(task-options!
  pom    {:project 'alda
          :version alda.version/-version-
          :description "A music programming language for musicians"
          :url "https://github.com/alda-lang/alda"
          :scm {:url "https://github.com/alda-lang/alda"}
          :license {"name" "Eclipse Public License"
                    "url" "http://www.eclipse.org/legal/epl-v10.html"}}
  jar    {:file    "alda.jar"
          :main    'alda.Client}
  exe    {:name    'alda
          :main    'alda.Client
          :version alda.version/-version-
          :desc    "A music programming language for musicians"
          :copyright "2015 Dave Yarwood et al"}
  target {:dir #{"target"}}
  test   {:namespaces '#{
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

(deftask package
  "Builds an uberjar."
  []
  (comp (javac) (pom) (uber) (jar)))

(deftask build
  "Builds an uberjar and executable binaries for Unix/Linux and Windows."
  [f file       PATH file "The path to an already-built uberjar."
   o output-dir PATH str  "The directory in which to places the binaries."]
  (comp
    (if-not file (package) identity)
    (bin :file file :output-dir output-dir)
    (exe :file file :output-dir output-dir)
    (target)))

(deftask deploy
  "Builds uberjar, installs it to local Maven repo, and deploys it to Clojars."
  []
  (comp (package) (install) (push-release)))

