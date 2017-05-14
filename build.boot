(set-env!
  :dependencies '[; build
                  [adzerk/boot-jar2bin   "1.1.0" :scope "test"]
                  [org.clojure/clojure   "1.8.0"]
                  [alda/client-java      "0.1.1"]
                  [alda/server-clj       "0.1.4"]
                  [alda/core             "0.1.2"]
                  [alda/sound-engine-clj "0.1.2"]
                  [alda/repl-clj         "0.1.0"]

                  ; silence slf4j logging dammit
                  [org.slf4j/slf4j-nop "1.7.21"]])

(require '[adzerk.boot-jar2bin :refer :all])

(def ^:const +version+ "1.0.0-rc56")

(defn- exe-version
  "Convert non-exe-friendly version numbers like 1.0.0-rc1 to four-number
   version numbers like 1.0.0.1 that launch4j can use to make exe files."
  [version]
  (if-let [[_ n rc] (re-matches #"(.*)-rc(\d+)" version)]
    (format "%s.%s" n rc)
    (if-let [[_ n _] (re-matches #"(.*)-SNAPSHOT" version)]
      (format "%s.999" n)
      version)))

(def jvm-opts #{"-XX:+UseG1GC"
                "-XX:MaxGCPauseMillis=100"
                "-Xms256m" "-Xmx1024m"
                "-Dclojure.compiler.direct-linking=true"})

(task-options!
  pom     {:project 'alda
           :version +version+
           :description "A music programming language for musicians"
           :url "https://github.com/alda-lang/alda"
           :scm {:url "https://github.com/alda-lang/alda"}
           :license {"name" "Eclipse Public License"
                     "url" "http://www.eclipse.org/legal/epl-v10.html"}}

  install {:pom "alda/alda"}

  jar     {:file     "alda.jar"
           :manifest {"alda-version" +version+}
           :main     'alda.Main}

  bin     {:jvm-opt jvm-opts}

  exe     {:name      'alda
           :main      'alda.Main
           :version   (exe-version +version+)
           :desc      "A music programming language for musicians"
           :copyright "2016 Dave Yarwood et al"
           :jvm-opt   jvm-opts}

  target  {:dir #{"target"}})

(deftask package
  "Builds an uberjar."
  []
  (comp (pom)
        (uber)
        (jar)))

(deftask build
  "Builds an uberjar and executable binaries for Unix/Linux and Windows."
  [f file       PATH file "The path to an already-built uberjar."
   o output-dir PATH str  "The directory in which to place the binaries."]
  (comp
    (if-not file (package) identity)
    (bin :file file :output-dir output-dir)
    (exe :file file :output-dir output-dir)))

