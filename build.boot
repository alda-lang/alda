(set-env!
  :source-paths #{"client/src"}
  :resource-paths #{"server/src" "server/test"
                    "server/grammar" "examples" "resources"}
  :dependencies '[
                  ; dev
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-jar2bin   "1.1.0"  :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]

                  ; server
                  [org.clojure/clojure    "1.8.0"]
                  [instaparse             "1.4.1"]
                  [io.aviso/pretty        "0.1.20"]
                  [com.taoensso/timbre    "4.1.1"]
                  [clj-http               "2.0.0"]
                  [ring                   "1.4.0"]
                  [ring/ring-defaults     "0.1.5"]
                  [compojure              "1.4.0"]
                  [djy                    "0.1.4"]
                  [str-to-argv            "0.1.0"]
                  [jline                  "2.12.1"]
                  [org.clojars.sidec/jsyn "16.7.3"]

                  ; client
                  [org.apache.commons/commons-lang3     "3.4"]
                  [org.apache.httpcomponents/httpclient "4.5.1"]
                  [com.google.code.gson/gson            "2.6.1"]
                  [com.beust/jcommander                 "1.48"]
                  [org.fusesource.jansi/jansi           "1.11"]
                  [net.jodah/recurrent                  "0.4.0"]
                  [us.bpsm/edn-java                     "0.4.6"]
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

(defn- exe-version
  "Convert non-exe-friendly version numbers like 1.0.0-rc1 to four-number
   version numbers like 1.0.0.1 that launch4j can use to make exe files."
  [version]
  (if-let [[_ n rc] (re-matches #"(.*)-rc(\d+)" version)]
    (format "%s.%s" n rc)
    (if-let [[_ n _] (re-matches #"(.*)-SNAPSHOT" version)]
      (format "%s.999" n)
      version)))

(def jvm-opts #{"-Dclojure.compiler.direct-linking=true"})

(task-options!
  javac   {:options ["-source" "1.7"
                     "-target" "1.7"
                     "-bootclasspath" (System/getenv "JDK7_BOOTCLASSPATH")]}

  pom     {:project 'alda
           :version alda.version/-version-
           :description "A music programming language for musicians"
           :url "https://github.com/alda-lang/alda"
           :scm {:url "https://github.com/alda-lang/alda"}
           :license {"name" "Eclipse Public License"
                     "url" "http://www.eclipse.org/legal/epl-v10.html"}}

  install {:pom "alda/alda"}

  jar     {:file "alda.jar"
           :main 'alda.Client}

  bin     {:jvm-opt jvm-opts}

  exe     {:name      'alda
           :main      'alda.Client
           :version   (exe-version alda.version/-version-)
           :desc      "A music programming language for musicians"
           :copyright "2016 Dave Yarwood et al"
           :jvm-opt   jvm-opts}

  target  {:dir #{"target"}}

  test    {:namespaces '#{
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

(deftask assert-jdk7-bootclasspath
  "Ensures that the JDK7_BOOTCLASSPATH environment variable is set, as required
   to build the uberjar with JDK7 support."
  []
  (with-pre-wrap fileset
    (assert (not (empty? (System/getenv "JDK7_BOOTCLASSPATH")))
            (str "Alda requires JDK7 in order to build its uberjar, in order "
                 "to provide out-of-the-box support for users who may have "
                 "older versions of Java. Please install JDK7 and set the "
                 "environment variable JDK7_BOOTCLASSPATH to the path to your "
                 "JDK7 classpath jar, e.g. (OS X example) "
                 "/Library/Java/JavaVirtualMachines/jdk1.7.0_71.jdk/Contents/"
                 "Home/jre/lib/rt.jar"))
    fileset))

(deftask dev
  "Runs the Alda server for development.

   There is a middleware that reloads all the server namespaces before each
   request, so that the server does not need to be restarted after making
   changes.

   The -F/--alda-fingerprint option technically does nothing, but including it
   as a long-style option when running this task from the command line* allows
   the Alda client to identify the dev server process as an Alda server and
   include it in the list of running servers.

   * e.g.: boot dev --port 27713 --alda-fingerprint

   Take care to include the --port long option as well, so the client knows
   the port on which the dev server is running."
  [p port             PORT int  "The port on which to start the server."
   F alda-fingerprint      bool "Allow the Alda client to identify this as an Alda server."]
  (comp
    (with-pre-wrap fs
      (let [direct-linking
            (System/getProperty "clojure.compiler.direct-linking")]
        (if-not (= direct-linking "true")
          (println "WARNING: You should include the JVM option"
                   "-Dclojure.compiler.direct-linking=true, as this option is"
                   "included in the binary. This will help catch potential"
                   "bugs caused by defining dynamic things without declaring"
                   "them ^:dynamic or ^:redef.")))
      (require 'alda.server)
      (require 'alda.util)
      ((resolve 'alda.util/set-timbre-level!) :debug)
      ((resolve 'alda.server/start-server!) (or port 27713))
      fs)
    (wait)))

(deftask package
  "Builds an uberjar."
  []
  (comp (assert-jdk7-bootclasspath)
        (javac)
        (pom)
        (uber)
        (jar)))

(deftask build
  "Builds an uberjar and executable binaries for Unix/Linux and Windows."
  [f file       PATH file "The path to an already-built uberjar."
   o output-dir PATH str  "The directory in which to places the binaries."]
  (comp
    (if-not file (package) identity)
    (bin :file file :output-dir output-dir)
    (exe :file file :output-dir output-dir)))

(deftask deploy
  "Builds uberjar, installs it to local Maven repo, and deploys it to Clojars."
  []
  (comp (build-jar) (push-release)))
