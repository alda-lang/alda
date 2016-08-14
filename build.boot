(set-env!
  :source-paths #{"client/src"}
  :resource-paths #{"server/src" "server/test"
                    "server/grammar" "examples" "resources"}
  :dependencies '[
                  ; dev
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-jar2bin   "1.1.0"  :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]
                  [str-to-argv           "0.1.0"  :score "test"]

                  ; server
                  [org.clojure/clojure    "1.8.0"]
                  [instaparse             "1.4.1"]
                  [io.aviso/pretty        "0.1.20"]
                  [com.taoensso/timbre    "4.1.1"]
                  [cheshire               "5.6.3"]
                  [djy                    "0.1.4"]
                  [jline                  "2.12.1"]
                  [org.clojars.sidec/jsyn "16.7.3"]
                  [potemkin               "0.4.1"]
                  [cc.qbits/jilch         "0.3.0"]
                  [str-to-argv            "0.1.0"]

                  ; client
                  [com.beust/jcommander                 "1.48"]
                  [net.jodah/recurrent                  "0.4.0"]
                  [org.apache.commons/commons-lang3     "3.4"]
                  [org.apache.httpcomponents/httpclient "4.5.1"]
                  [com.google.code.gson/gson            "2.6.1"]
                  [org.fusesource.jansi/jansi           "1.11"]
                  [us.bpsm/edn-java                     "0.4.6"]
                  ])

(require '[adzerk.bootlaces    :refer :all]
         '[adzerk.boot-jar2bin :refer :all]
         '[adzerk.boot-test    :refer :all]
         '[alda.version])

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
                          alda.parser.variables-test
                          alda.lisp.attributes-test
                          alda.lisp.cram-test
                          alda.lisp.chords-test
                          alda.lisp.code-test
                          alda.lisp.duration-test
                          alda.lisp.global-attributes-test
                          alda.lisp.markers-test
                          alda.lisp.notes-test
                          alda.lisp.parts-test
                          alda.lisp.pitch-test
                          alda.lisp.score-test
                          alda.lisp.variables-test
                          alda.lisp.voices-test
                          alda.util-test

                          ; benchmarks / smoke tests
                          alda.examples-test
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
  "Runs the Alda server (default), REPL, or client for development.

   *** REPL ***

   Simply run `boot dev -a repl` and you're in!

   *** CLIENT ***

   To test changes to the Alda client, run `boot dev -a client -x \"args here\"`.

   For example:

      boot dev -a client -x \"play --file /path/to/file.alda\"

   The arguments must be a single command-line string to be passed to the
   command-line client as if entering them on the command line. The example
   above is equivalent to running `alda play --file /path/to/file.alda` on the
   command line.

   One caveat to running the client this way (as opposed to building it and
   running the resulting executable) is that the client does not have the
   necessary permissions to start a new process, e.g. to start an Alda server
   via the client. If you'd like to test local changes to the server code,
   you'll need to run the server instead (see SERVER below).

   *** SERVER ***

   The -F/--alda-fingerprint option technically does nothing, but including it
   as a long-style option when running this task from the command line* allows
   the Alda client to identify the dev server process as an Alda server and
   include it in the list of running servers.

   For example:

      boot dev -a server --port 27713 --alda-fingerprint

   Take care to include the --port long option as well, so the client knows
   the port on which the dev server is running.

   There is a middleware that reloads all the server namespaces before each
   request, so that the server does not need to be restarted after making
   changes."
  [a app              APP  str  "The Alda application to run (server, repl or client)."
   x args             ARGS str  "The string of CLI args to pass to the client."
   p port             PORT int  "The port on which to start the server."
   F alda-fingerprint      bool "Allow the Alda client to identify this as an Alda server."]
  (comp
    (if (= app "client") (javac) identity)
    (with-pre-wrap fs
      (let [direct-linking (System/getProperty "clojure.compiler.direct-linking")
            start-server!  (fn []
                             (require 'alda.server)
                             (require 'alda.util)
                             ((resolve 'alda.util/set-timbre-level!) :debug)
                             ((resolve 'alda.server/start-server!) (or port 27713)))
            start-repl!    (fn []
                             (require 'alda.repl)
                             ((resolve 'alda.repl/start-repl!)))
            run-client!    (fn []
                             (require '[str-to-argv])
                             (import 'alda.Client)
                             (eval `(alda.Client/main
                                      (into-array String
                                        (str-to-argv/split-args (or ~args ""))))))]
        (when-not (= direct-linking "true")
          (println "WARNING: You should include the JVM option"
                   "-Dclojure.compiler.direct-linking=true, as this option is"
                   "included in the binary. This will help catch potential bugs"
                   "caused by defining dynamic things without declaring them"
                   "^:dynamic or ^:redef."))
        (case app
          nil      (start-server!)
          "server" (start-server!)
          "repl"   (start-repl!)
          "client" (run-client!)
          (do
            (println "ERROR: -a/--app must be server, repl or client")
            (System/exit 1))))
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

;; misc tasks ;;

(deftask generate-completions
  "Generates the `completions.cson` file used for autocompletions in the Atom
   Alda language plugin. This file contains instrument and attribute names.

   https://github.com/MadcapJake/language-alda/blob/master/completions.cson"
  []
  (require '[alda.lisp.model.instrument :as instrument]
           '[alda.lisp.model.attribute  :as attribute]
           '[alda.lisp.instruments.midi]
           '[alda.lisp.attributes]
           '[clojure.string             :as str])
  (let [cson-format (fn [xs]
                      (->> (sort xs)
                           (map #(str "  '" % \'))
                           ((resolve 'str/join) \newline)))
        instruments (->> (resolve 'instrument/*stock-instruments*)
                         var-get
                         keys
                         cson-format)
        attributes  (->> (resolve 'attribute/*attribute-table*)
                         var-get
                         keys
                         (map name)
                         cson-format)]
    (println "'instruments': [")
    (println instruments)
    (println \])
    (println "'attributes': [")
    (println attributes)
    (println \])))

