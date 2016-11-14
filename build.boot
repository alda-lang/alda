(set-env!
  :dependencies '[; build
                  [adzerk/boot-jar2bin "1.1.0" :scope "test"]
                  [alda/client-java    "0.0.1"]

                  ; silence slf4j logging dammit
                  [org.slf4j/slf4j-nop "1.7.21"]])

(require '[adzerk.boot-jar2bin :refer :all])

(def ^:const +version+ "1.0.0-rc99")

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

(deftask build
  "Builds an uberjar and executable binaries for Unix/Linux and Windows."
  [f file       PATH file "The path to an already-built uberjar."
   o output-dir PATH str  "The directory in which to place the binaries."]
  (comp
    (if-not file (package) identity)
    (bin :file file :output-dir output-dir)
    (exe :file file :output-dir output-dir)))

