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
         '[adzerk.boot-test :refer :all])

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

(defn -main [& args]
  (require 'alda.core)
  (apply (resolve 'alda.core/-main) args))
