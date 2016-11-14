(set-env!
  :source-paths #{"src"}
  :resource-paths #{"../server/src" "../server/test"
                    "../server/grammar" "../examples"}
  :dependencies '[
                  ; dev
                  [adzerk/bootlaces      "0.1.12" :scope "test"]
                  [adzerk/boot-test      "1.0.4"  :scope "test"]
                  [str-to-argv           "0.1.0"  :score "test"]

                  ; silence slf4j logging dammit
                  [org.slf4j/slf4j-nop        "1.7.21"]

                  ; shared
                  [org.zeromq/jeromq "0.3.5"]

                  ; server / worker
                  [org.clojure/clojure    "1.8.0"]
                  [instaparse             "1.4.1"]
                  [io.aviso/pretty        "0.1.20"]
                  [com.taoensso/timbre    "4.1.1"]
                  [cheshire               "5.6.3"]
                  [djy                    "0.1.4"]
                  [jline                  "2.12.1"]
                  [org.clojars.sidec/jsyn "16.7.3"]
                  [potemkin               "0.4.1"]
                  [org.zeromq/cljzmq      "0.1.4" :exclusions (org.zeromq/jzmq)]
                  [me.raynes/conch        "0.8.0"]
                  [clj_manifest           "0.2.0"]

                  ; client
                  [com.beust/jcommander                 "1.48"]
                  [commons-io/commons-io                "2.5"]
                  [org.apache.commons/commons-lang3     "3.4"]
                  [org.apache.httpcomponents/httpclient "4.5.1"]
                  [com.google.code.gson/gson            "2.6.1"]
                  [org.fusesource.jansi/jansi           "1.11"]
                  [us.bpsm/edn-java                     "0.4.6"]
                  [com.jcabi/jcabi-manifests            "1.1"]])

(require '[adzerk.bootlaces :refer :all])

(def ^:const +version+ "0.0.1")

(bootlaces! +version+)

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
  javac   {:options (let [jdk7-bootclasspath (System/getenv "JDK7_BOOTCLASSPATH")]
                      (concat
                        []
                        (when-not (empty? jdk7-bootclasspath)
                          ["-source" "1.7"
                           "-target" "1.7"
                           "-bootclasspath" jdk7-bootclasspath])))}

  pom     {:project 'alda/client-java
           :version +version+
           :description "A music programming language for musicians"
           :url "https://github.com/alda-lang/alda"
           :scm {:url "https://github.com/alda-lang/alda"}
           :license {"name" "Eclipse Public License"
                     "url" "http://www.eclipse.org/legal/epl-v10.html"}}

  jar     {:file "alda-client.jar"
           :main 'alda.Main}

  install {:pom "alda/client-java"}

  target  {:dir #{"target"}})

; (deftask heading
;   [t text TEXT str "The text to display."]
;   (with-pass-thru fs
;     (println)
;     (println "---" text "---")))

; (deftask unit-tests
;   []
;   (comp (heading :text "UNIT TESTS")
;         (adzerk.boot-test/test
;           :namespaces '#{; general tests
;                          alda.parser.barlines-test
;                          alda.parser.clj-exprs-test
;                          alda.parser.event-sequences-test
;                          alda.parser.comments-test
;                          alda.parser.duration-test
;                          alda.parser.events-test
;                          alda.parser.octaves-test
;                          alda.parser.repeats-test
;                          alda.parser.score-test
;                          alda.parser.variables-test
;                          alda.lisp.attributes-test
;                          alda.lisp.cram-test
;                          alda.lisp.chords-test
;                          alda.lisp.code-test
;                          alda.lisp.duration-test
;                          alda.lisp.global-attributes-test
;                          alda.lisp.markers-test
;                          alda.lisp.notes-test
;                          alda.lisp.parts-test
;                          alda.lisp.pitch-test
;                          alda.lisp.score-test
;                          alda.lisp.variables-test
;                          alda.lisp.voices-test
;                          alda.util-test

;                          ; benchmarks / smoke tests
;                          alda.examples-test})))

; (deftask integration-tests
;   []
;   (comp (heading :text "INTEGRATION TESTS")
;         (adzerk.boot-test/test
;           :namespaces '#{alda.server-test
;                          alda.worker-test})))

; (deftask test
;   [i integration bool "Run only integration tests."
;    a all         bool "Run all tests."]
;   (cond
;     all         (comp (unit-tests)
;                       (integration-tests))
;     integration (integration-tests)
;     :default    (unit-tests)))

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

   *** WORKER ***

   Starts a worker process that will receive requests from the socket on the
   port specified by the -p/--port option. This option is required."
  [a app              APP     str  "The Alda application to run (server, repl or client)."
   x args             ARGS    str  "The string of CLI args to pass to the client."
   p port             PORT    int  "The port on which to start the server/worker."
   w workers          WORKERS int  "The number of workers for the server to start."
   F alda-fingerprint         bool "Allow the Alda client to identify this as an Alda process."]
  (comp
    (javac)
    (with-pre-wrap fs
      (let [direct-linking (System/getProperty "clojure.compiler.direct-linking")
            start-server!  (fn []
                             (require 'alda.server)
                             (require 'alda.util)
                             ((resolve 'alda.server/start-server!)
                                (or workers 2)
                                (or port 27713)
                                true))
            start-worker!  (fn []
                             (assert port
                               "The --port option is mandatory for workers.")
                             (require 'alda.worker)
                             (require 'alda.util)
                             ((resolve 'alda.worker/start-worker!) port true))
            start-repl!    (fn []
                             (require 'alda.repl)
                             ((resolve 'alda.repl/start-repl!)))
            run-client!    (fn []
                             (require '[str-to-argv])
                             (import 'alda.Main)
                             (eval `(alda.Main/main
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
          "worker" (start-worker!)
          "repl"   (start-repl!)
          "client" (run-client!)
          (do
            (println "ERROR: -a/--app must be server, repl or client")
            (System/exit 1))))
      fs)
    (wait)))

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

(deftask package
  "Builds an uberjar."
  []
  (comp (assert-jdk7-bootclasspath)
        (javac)
        (pom)
        (uber)
        (jar)))

(deftask deploy
  "Builds uberjar, installs it to local Maven repo, and deploys it to Clojars."
  []
  (comp (build-jar) (push-release)))

