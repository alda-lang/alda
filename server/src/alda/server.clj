(ns alda.server
  (:require [alda.lisp        :refer :all]
            [alda.now         :as    now]
            [alda.parser      :refer (parse-input)]
            [alda.parser-util :refer (parse-to-events-with-context)]
            [alda.sound       :refer (*play-opts*)]
            [alda.sound.midi  :as    midi]
            [alda.util]
            [alda.version     :refer (-version-)]
            [qbits.jilch.mq   :as    zmq]
            [taoensso.timbre  :as    log]
            [cheshire.core    :as    json]
            [clojure.edn      :as    edn]
            [clojure.java.io  :as    io]
            [clojure.pprint   :refer (pprint)]
            [clojure.string   :as    str]))

(defn new-server-score
  [& [score]]
  (doto (if score
          (now/new-score score)
          (now/new-score))
    (swap! assoc :score-text ""
                 :filename   nil)))

(def ^:dynamic *current-score* (new-server-score))

(defn score-text<<
  [txt]
  (swap! *current-score*
         update :score-text
         #(str % (when-not (empty? %) \newline) txt)))

(defn close-score!
  []
  (now/tear-down! *current-score*))

(defn new-score!
  []
  (alter-var-root #'*current-score* (constantly (new-server-score))))

(defn load-score!
  [score-text]
  (let [loaded-score (-> score-text parse-input eval new-server-score)]
    (alter-var-root #'*current-score* (constantly loaded-score))
    (swap! *current-score* assoc :score-text score-text)))

(defn start-alda-environment!
  []
  (midi/fill-midi-synth-pool!)
  (while (not (midi/midi-synth-available?))
    (Thread/sleep 250)))

(defn modified?
  []
  (let [{:keys [filename score-text]} @*current-score*]
    (if filename
      (try
        (let [file-contents (slurp (io/file filename))]
          (not= score-text file-contents))
        (catch java.io.FileNotFoundException e
          ; if the file no longer exists (e.g. has been deleted since opening),
          ; consider the score to have been modified -- this should prompt the
          ; user to save changes, which will re-create the file
          true))
      (not (empty? score-text)))))

(defn confirming?
  [{:keys [headers] :as request}]
  (let [confirm-header (get headers "x-alda-confirm" "false")]
    (not (contains? #{"" "no" "false"} (.toLowerCase confirm-header)))))

(defn score-info
  [& {:keys [formatted]}]
  (let [{:keys [filename score-text instruments]} @*current-score*
        info {:status      "up"
              :version     -version-
              :filename    filename
              :modified?   (modified?)
              :line-count  (count (str/split score-text #"\n\r|\r\n|\n|\r"))
              :char-count  (count score-text)
              :instruments (for [[k v] instruments] {:name  k
                                                     :stock (:stock v)})}]
    (if formatted
      (let [{:keys [status version filename modified?
                    line-count char-count instruments]} info]
        (format (str "Server status: %s\n"
                     "Server version: %s\n"
                     "Filename: %s\n"
                     "Modified: %s\n"
                     "Line count: %d\n"
                     "Character count: %d\n"
                     "Instruments:%s")
                status
                version
                (or filename "(new score)")
                (if modified? "yes" "no")
                line-count
                char-count
                (if (empty? instruments)
                  " (none)"
                  (apply str
                         (for [{:keys [name stock]} instruments]
                           (format "\n  • %s (%s)" name stock))))))
      info)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- success-response
  [body]
  {:success true
   :body    (if (map? body)
              (json/generate-string body)
              body)})

(defn- need-confirmation-response
  [prompt]
  {:success false
   :body (str prompt \newline
              \newline
              "If so, please re-submit your request and include the parameter "
              "\"confirming\": true")})

(def ^:private unsaved-changes-response
  (-> (need-confirmation-response
        (str "Warning: the score has unsaved changes. Are you sure you want to "
             "do this?"))
      (assoc :signal "unsaved-changes")))

(def ^:private existing-file-response
  (-> (need-confirmation-response
        (str "Warning: there is an existing file with the filename you "
             "specified. Saving the score to this file will erase whatever is "
             "already there. Are you sure you want to do this?"))
      (assoc :signal "existing-file")))

(defn- error-response
  [e]
  {:success false
   :body    (if (string? e)
              e
              (.getMessage e))})

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn handle-code
  [code & {:keys [; replace the current score
                  replace-score?
                  ; add to/replace the current score only, don't play
                  silent?
                  ; don't replace or add to the current score, just play the code
                  one-off?
                  ]}]
  (let [msg (if silent? "Loading..." "Playing...")
        save-state @*current-score*]
    (try
      (require '[alda.lisp :refer :all])
      (if one-off?
        (if-let [score (try
                         (parse-input code :map)
                         (catch Throwable e
                           (log/error e e)
                           nil))]
          (do
            (now/play-score! score)
            (success-response msg))
          (error-response "Invalid Alda syntax."))
        (let [[context events] (parse-to-events-with-context code)]
          (if (= context :parse-failure)
            (error-response "Invalid Alda syntax.")
            (do
              (when replace-score?
                (close-score!)
                (new-score!))
              (score-text<< code)
              (if silent?
                (swap! *current-score* continue events)
                (now/with-score *current-score*
                  (now/play! events)))
              (success-response msg)))))
      (catch Throwable e
        (log/error e e)
        (reset! *current-score* save-state)
        (error-response e)))))

(defn handle-code-parse
  [code & {:keys [mode] :or {mode :lisp}}]
  (try
    (require '[alda.lisp :refer :all])
    (success-response (case mode
                        :lisp (let [result (parse-input code mode)]
                                (with-out-str (pprint result)))
                        :map  (parse-input code mode)))
    (catch Throwable e
      (log/error e e)
      (error-response "Invalid Alda syntax."))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(def ^:private running? (atom true))

(defn stop-server!
  []
  (log/info "Received request to stop. Shutting down...")
  (reset! running? false))

(defmulti process :command)

(defmethod process :default
  [{:keys [command]}]
  (error-response (format "Unrecognized command: %s" command)))

(defmethod process nil
  [_]
  (error-response "Missing command"))

(defmethod process "append"
  [{:keys [body]}]
  (handle-code body :silent? true))

(defmethod process "current-score"
  [{:keys [options]}]
  (let [{:keys [as]} options]
    (case as
      "lisp" (handle-code-parse (:score-text @*current-score*) :mode :lisp)
      "map"  (success-response (dissoc @*current-score* :audio-context))
      "text" (success-response (:score-text @*current-score*))
      nil    (error-response "Missing option: as")
      (error-response (format "Invalid score format: \"%s\"" as)))))

(defmethod process "filename"
  [_]
  (success-response (:filename (score-info))))

(defmethod process "info"
  [_]
  (success-response (score-info :formatted true)))

(defmethod process "load"
  [{:keys [body options confirming]}]
  (if (and (modified?) (not confirming))
    unsaved-changes-response
    (let [{:keys [filename]} options
          result (handle-code body :silent? true :replace-score? true)]
      (swap! *current-score* assoc :filename filename)
      result)))

(defmethod process "modified?"
  [_]
  (success-response (str (modified?))))

(defmethod process "new-score"
  [{:keys [confirming]}]
  (if (and (modified?) (not confirming))
    unsaved-changes-response
    (do
      (close-score!)
      (new-score!)
      (success-response "New score initialized."))))

(defmethod process "parse"
  [{:keys [body options]}]
  (let [{:keys [as]} options]
    (case as
      "lisp" (handle-code-parse body :mode :lisp)
      "map"  (handle-code-parse body :mode :map)
      nil    (error-response "Missing option: as")
      (error-response (format "Invalid format: %s" as)))))

(defmethod process "play"
  [{:keys [body options]}]
  (let [{:keys [from to append]} options]
    (if append
      (handle-code body)
      (binding [*play-opts* (assoc *play-opts*
                                   :from     from
                                   :to       to
                                   :one-off? true)]
        (handle-code body :one-off? true)))))

(defmethod process "play-score"
  [{:keys [body options]}]
  (let [{:keys [from to]} options]
    (binding [*play-opts* (assoc *play-opts* :from from :to to)]
      (now/play-score! *current-score*)
      (success-response "Playing..."))))

(defmethod process "save"
  [{:keys [options confirming]}]
  (let [{:keys [score-text]} @*current-score*]
    (if-let [filename (:filename options)]
      (if (and (.exists (io/as-file filename)) (not confirming))
        existing-file-response
        (do
          (spit filename score-text)
          (swap! *current-score* assoc :filename filename)
          (success-response (format "File saved: %s" filename))))
      (if-let [filename (:filename @*current-score*)]
        (do
          (spit filename score-text)
          (success-response (format "File saved: %s" filename)))
        (error-response "You must supply a filename.")))))

(defmethod process "stop-server"
  [{:keys [confirming]}]
  (if (and (modified?) (not confirming))
    unsaved-changes-response
    (do
      (close-score!)
      (stop-server!)
      (success-response "Shutting down."))))

(defmethod process "version"
  [_]
  (success-response -version-))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn start-server!
  [port]
  (log/info "Loading Alda environment...")
  (start-alda-environment!)

  (log/infof "Starting Alda zmq server on port %s..." port)
  (zmq/with-context zmq-ctx 2
    (with-open [socket (-> zmq-ctx
                           (zmq/socket zmq/rep)
                           (zmq/bind (str "tcp://*:" port)))]
      (while (and (not (.. Thread currentThread isInterrupted)) @running?)
        (let [req (zmq/recv socket)]
          (try
            (let [msg (json/parse-string (String. req) true)
                  res (process msg)]
              (zmq/send socket (json/generate-string res)))
            (catch Throwable e
              (log/error e e)
              (zmq/send socket (json/generate-string (error-response e)))))))))
  (System/exit 0))

