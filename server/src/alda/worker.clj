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

(def playing? (atom false))

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
        (future
          (reset! playing? true)
          (now/play-score! score {:async? false :one-off? false})
          (log/debug "Done playing score.")
          (reset! playing? false))
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

(defmethod process "version"
  [_]
  (success-response -version-))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(def ^:const HEARTBEAT-INTERVAL 1000)
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
        last-heartbeat (atom 0)]
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
              (if @playing?
                (log/debug "Ignoring message. Busy playing.")
                (let [envelope (.unwrap msg)
                      body     (-> msg .pop .getData (String.))]
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
            (when (and (<= @lives 0) (not @playing?))
              (log/error "Unable to reach the server.")
              (reset! running? false))))
        (when (> (System/currentTimeMillis)
                 (+ @last-heartbeat HEARTBEAT-INTERVAL))
          (.send (ZFrame. (if @playing? "BUSY" "AVAILABLE")) socket 0))))
    (log/info "Shutting down.")
    (System/exit 0)))

