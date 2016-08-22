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

(defn start-alda-environment!
  []
  (midi/fill-midi-synth-pool!)
  (while (not (midi/midi-synth-available?))
    (Thread/sleep 250)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- success-response
  [body]
  {:success true
   :body    (if (map? body)
              (json/generate-string body)
              body)})

(defn- error-response
  [e]
  {:success false
   :body    (if (string? e)
              e
              (.getMessage e))})

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn handle-code
  [code]
  (try
    (log/debug "Requiring alda.lisp...")
    (require '[alda.lisp :refer :all])
    (if-let [score (try
                     (log/debug "Parsing input...")
                     (parse-input code :map)
                     (catch Throwable e
                       (log/error e e)
                       nil))]
      (do
        (log/debug "Playing score...")
        (now/play-score! score)
        (success-response "Playing..."))
      (error-response "Invalid Alda syntax."))
    (catch Throwable e
      (log/error e e)
      (error-response e))))

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

(defmethod process "parse"
  [{:keys [body options]}]
  (let [{:keys [as]} options]
    (case as
      "lisp" (handle-code-parse body :mode :lisp)
      "map"  (handle-code-parse body :mode :map)
      nil    (error-response "Missing option: as")
      (error-response (format "Invalid format: %s" as)))))

(defmethod process "ping"
  [_]
  (success-response "OK"))

(defmethod process "play"
  [{:keys [body options]}]
  (let [{:keys [from to]} options]
    (binding [*play-opts* (assoc *play-opts*
                                 :from     from
                                 :to       to
                                 :one-off? true)]
      (handle-code body))))

(defmethod process "stop-server"
  [_]
  (stop-server!)
  (success-response "Shutting down."))

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
        (log/debug "Receiving request...")
        (let [req (zmq/recv socket)]
          (log/debug "Request received.")
          (try
            (let [msg (json/parse-string (String. req) true)
                  _   (log/debug "Processing message...")
                  res (process msg)]
              (log/debug "Sending response...")
              (zmq/send socket (json/generate-string res))
              (log/debug "Response sent."))
            (catch Throwable e
              (log/error e e)
              (zmq/send socket (json/generate-string (error-response e)))))))))
  (System/exit 0))

