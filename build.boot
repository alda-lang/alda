#!/usr/bin/env boot

(set-env!
 :source-paths #{"src" "test"}
 :resource-paths #{"grammar"}
 :dependencies '[[org.clojure/clojure "1.6.0"]
                 [org.clojure/tools.cli "0.3.1"]
                 [instaparse "1.3.5"]
                 [adzerk/bootlaces "0.1.9" :scope "test"]
                 [adzerk/boot-test "1.0.3" :scope "test"]
                 [com.taoensso/timbre "3.4.0"]
                 [djy "0.1.3"]])

(require '[adzerk.bootlaces :refer :all]
         '[adzerk.boot-test :refer :all]
         '[alda.core]
         '[alda.parser :refer (parse-input)])

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
  test {:namespaces '#{alda.parser-test
                       alda.lisp-attribute-test
                       alda.lisp-event-test
                       alda.lisp-score-test}})

(deftask build
  "Builds uberjar.
   TODO: be able to build an executable à la lein bin"
  []
  (comp (aot) (pom) (uber) (jar)))

(deftask parse
  "Parse an input alda file and print the results to the console."
  [f file FILE str  "The path to a file containing alda code."
   c code CODE str  "A string of Alda code."
   l lisp      bool "Parse into alda-lisp code."
   m map       bool "Evaluate the score and show the resulting instruments/events map."]
  (let [alda-lisp-code (parse-input (if code code (slurp file)))]
    (when lisp
      (prn alda-lisp-code)
    (when map
      (require 'alda.lisp)
      (println)
      (prn (eval alda-lisp-code))))))

(defn -main [& args]
  (apply alda.core/-main args))
