(ns alda.worker
  (:require [alda.now        :as    now]
            [alda.parser     :refer (parse-input)]
            [alda.sound      :as    sound :refer (*play-opts*)]
            [alda.sound.midi :as    midi]
            [alda.util       :as    util]
            [alda.version    :refer (-version-)]
            [cheshire.core   :as    json]
            [clojure.pprint  :refer (pprint)]
            [taoensso.timbre :as    log]
            [zeromq.zmq      :as    zmq])
  (:import [org.zeromq ZFrame ZMQ ZMsg]))

(defn start-alda-environment!
  []
  (sound/start-synthesis-engine!)
  (midi/open-midi-synth!))

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

(def current-status (atom :available))
(def current-error (atom nil))

(defn handle-code-play
  [code]
  (future
    (reset! current-status :parsing)
    (log/debug "Requiring alda.lisp...")
    (require '[alda.lisp :refer :all])
    (let [score (try
                  (log/debug "Parsing input...")
                  (parse-input code :map)
                  (catch Throwable e
                    {:error e}))]
      (if-let [error (:error score)]
        (do
          (log/error error error)
          (reset! current-status :error)
          (reset! current-error error))
        (try
          (log/debug "Playing score...")
          (reset! current-status :playing)
          (now/play-score! score {:async? false :one-off? false})
          (log/debug "Done playing score.")
          (reset! current-status :available)
          (catch Throwable e
            (log/error e e)
            (reset! current-status :error)
            (reset! current-error e))))))
  (success-response "Request received."))

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
      (error-response e))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

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
      (handle-code-play body))))

(defmethod process "play-status"
  [_]
  (if (= :error @current-status)
    (let [error @current-error]
      (reset! current-status :available)
      (reset! current-error nil)
      (error-response error))
    (-> (success-response (name @current-status))
        (assoc :pending (not (#{:available :playing} @current-status))))))

(defmethod process "version"
  [_]
  (success-response -version-))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(def ^:const HEARTBEAT-INTERVAL 1000)
(def ^:const SUSPENDED-INTERVAL 10000)
(def ^:const MAX-LIVES          10)

(def lives (atom MAX-LIVES))

(defn start-worker!
  [port]
  (when-not (System/getenv "ALDA_DEBUG")
    (let [log-path (util/alda-home-path "logs" "error.log")]
      (log/infof "Logging errors to %s" log-path)
      (util/set-log-level! :error)
      (util/rolling-log! log-path)))
  (log/info "Loading Alda environment...")
  (start-alda-environment!)
  (log/info "Worker reporting for duty!")
  (log/infof "Connecting to socket on port %d..." port)
  (let [zmq-ctx        (zmq/zcontext)
        poller         (zmq/poller zmq-ctx 1)
        running?       (atom true)
        last-heartbeat (atom (System/currentTimeMillis))]
    (with-open [socket (doto (zmq/socket zmq-ctx :dealer)
                         (zmq/connect (str "tcp://*:" port))) ]
      (log/info "Sending READY signal.")
      (.send (ZFrame. "READY") socket 0)

      (zmq/register poller socket :pollin)

      (while (and @running? (not (.. Thread currentThread isInterrupted)))
        (zmq/poll poller HEARTBEAT-INTERVAL)
        (if (zmq/check-poller poller 0 :pollin)
          (when-let [msg (ZMsg/recvMsg socket ZMQ/DONTWAIT)]
            (cond
              ; the server sends 1-frame messages as signals
              (= 1 (.size msg))
              (let [frame (.getFirst msg)
                    data  (String. (.getData frame))]
                (case data
                  "KILL"      (do
                                (log/info "Received KILL signal from server.")
                                (reset! running? false))
                  "HEARTBEAT" (do
                                (log/debug "Got HEARTBEAT from server.")
                                (reset! lives MAX-LIVES))
                  (log/errorf "Invalid message: %s" data)))

              ; the server also forwards 3-frame messages from the client
              ; Frames:
              ;   1) the return address of the client
              ;   2) a JSON string representing the client's request
              ;   3) the command as a string (for use by the server)
              (= 3 (.size msg))
              (let [envelope (.unwrap msg)
                    command  (-> msg .getLast .getData (String.))
                    body     (-> msg .pop .getData (String.))]
                (if (and (not= :available @current-status)
                         (not= "play-status" command))
                  (log/debug "Ignoring message. I'm busy.")
                  (try
                    (log/debug "Processing message...")
                    (let [req (json/parse-string body true)
                          res (json/generate-string (process req))]
                      (log/debug "Sending response...")
                      (util/respond-to msg socket res envelope)
                      (log/debug "Response sent."))
                    (catch Throwable e
                      (log/error e e)
                      (log/info "Sending error response...")
                      (let [err (json/generate-string (error-response e))]
                        (util/respond-to msg socket err envelope))
                      (log/info "Error response sent.")))))

              :else
              (log/errorf "Invalid message: %s" msg)))
          (do
            (swap! lives dec)
            (when (and (<= @lives 0) (= :available @current-status))
              (log/error "Unable to reach the server.")
              (reset! running? false))))

        ; detect when the system has been suspended and stop working so the
        ; server can replace it with a fresh worker (this fixes a bug where
        ; MIDI audio is delayed)
        (when (> (System/currentTimeMillis)
                 (+ @last-heartbeat SUSPENDED-INTERVAL))
          (log/info "Process suspension detected. Shutting down...")
          (reset! last-heartbeat (System/currentTimeMillis))
          (reset! running? false))

        (when (> (System/currentTimeMillis)
                 (+ @last-heartbeat HEARTBEAT-INTERVAL))
          (.send (ZFrame. (if (= :available @current-status)
                            "AVAILABLE"
                            "BUSY"))
                 socket
                 0)
          (reset! last-heartbeat (System/currentTimeMillis)))))
    (log/info "Shutting down.")
    (System/exit 0)))

