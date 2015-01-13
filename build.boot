(set-env!
 :source-paths #{"src" "test"}
 :resource-paths #{"resources"}
 :dependencies '[[org.clojure/clojure "1.6.0"]
                 [org.clojure/tools.cli "0.2.4"]
                 [instaparse "1.3.3"]
                 [adzerk/bootlaces "0.1.8" :scope "test"]
                 [adzerk/boot-test "1.0.3" :scope "test"]])

(require '[adzerk.bootlaces :refer :all]
         '[adzerk.boot-test :refer :all])

(def +version+ "0.1.0")
(bootlaces! +version+)

(task-options!
  pom {:project 'alda
       :version +version+
       :description "A music programming language for musicians"
       :url "https://github.com/daveyarwood/alda"
       :scm {:url "https://github.com/daveyarwood/alda"}
       :license {:name "Eclipse Public License"
                 :url "http://www.eclipse.org/legal/epl-v10.html"}}
  test {:namespaces '#{alda.parser-test}})

(defn -main [& args]
  "To do: refer to boot-uberjar example")
