(set-env!
  :dependencies '[; build / release
                  [adzerk/boot-jar2bin   "1.1.1" :scope "test"]
                  [io.djy/boot-github    "0.1.4" :scope "test"]
                  [org.clojure/clojure   "1.8.0"]
                  [alda/client-java      "0.6.0"]
                  [alda/server-clj       "0.4.1"]
                  [alda/core             "0.5.0"]
                  [alda/sound-engine-clj "1.0.0"]

                  ; silence slf4j logging dammit
                  [org.slf4j/slf4j-nop "1.7.25"]])

(require '[adzerk.boot-jar2bin :refer :all]
         '[io.djy.boot-github  :refer (push-version-tag create-release
                                       changelog-for)]
         '[adzerk.env      :as env]
         '[boot.util       :as util]
         '[cheshire.core   :as json]
         '[clj-http.client :as http])

(def ^:const +version+ "1.1.0")

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
           :desc      "Alda"
           :copyright "2012-2018 Dave Yarwood et al"
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

(deftask announce-release
  "Announces the release (with version and changelog) in the #general Alda Slack
   channel."
  [v version VERSION str "The version being released."]
  (with-pass-thru _
    (util/info "Announcing release on Slack...\n")
    (env/def ALDA_SLACK_WEBHOOK_URL :required)
    (let [text     (format "Alda version %s released!\n\n%s"
                           +version+
                           (changelog-for +version+))
          payload  (json/generate-string {:text text :channel "#general"})
          response (http/post ALDA_SLACK_WEBHOOK_URL
                              {:form-params {:payload payload}})]
      (if (= 200 (:status response))
        (util/info "Announcement successful.\n")
        (do
          (util/fail "Announcement failed:\n")
          (println response))))))

(deftask release
  "* Builds to `output-dir`.
   * Pushes a new git version tag.
   * Creates a new release via the GitHub API.
   * Generates a release description from the CHANGELOG.
   * Uploads the executables to the release.
   * Posts to Alda Slack #general about the new release."
  []
  (env/def GITHUB_TOKEN :required)
  (let [tmpdir (System/getProperty "java.io.tmpdir")
        assets (into #{} (map #(str tmpdir "/" %) ["alda" "alda.exe"]))]
    (comp
      (build :output-dir tmpdir)
      (push-version-tag :version +version+)
      (create-release :version      +version+
                      :changelog    true
                      :assets       assets
                      :github-token GITHUB_TOKEN)
      (announce-release :version +version+))))

