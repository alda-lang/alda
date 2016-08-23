(ns alda.server
  (:require [alda.util                 :as util]
            [cheshire.core             :as json]
            [me.raynes.conch.low-level :as sh]
            [taoensso.timbre           :as log]
            [zeromq.device             :as zmqd]
            [zeromq.zmq                :as zmq])
  (:import [java.net ServerSocket]
           [java.util.concurrent ConcurrentLinkedQueue]
           [org.zeromq ZFrame ZMQException ZMQ$Error ZMsg]))

; the number of ms between heartbeats
(def ^:const HEARTBEAT-INTERVAL 1000)
; the amount of missed heartbeats before a worker is pronounced dead
(def ^:const WORKER-LIVES          3)

(defn worker-expiration-date []
  (+ (System/currentTimeMillis) (* HEARTBEAT-INTERVAL WORKER-LIVES)))

(defn worker [address]
  {:address address
   :expiry  (worker-expiration-date)})

(def available-workers (util/queue))
(def busy-workers      (ref #{}))
(defn all-workers []   (concat @available-workers (for [address @busy-workers]
                                                    {:address address})))

; (doseq [[x y] [[:available available-workers]
;                [:busy busy-workers]]]
;   (add-watch y :key (fn [_ _ old new]
;                       (when (not= old new)
;                         (prn x new)))))

(def no-workers-available-response
  (json/generate-string
    {:success false
     :body "No workers processes are ready yet. Please wait a minute."}))

(def all-workers-are-busy-response
  (json/generate-string
    {:success false
     :body "All worker processes are currently busy. Please wait until playback is complete and re-submit your request."}))

(defn add-or-requeue-worker
  [address]
  (dosync
    (alter busy-workers disj address)
    (if (util/check-queue available-workers #(= address (:address %)))
      (util/re-queue available-workers
                     #(= address (:address %))
                     #(assoc % :expiry (worker-expiration-date)))
      (util/push-queue available-workers (worker address)))))

(defn note-that-worker-is-busy
  [address]
  (dosync
    (util/remove-from-queue available-workers #(= address (:address %)))
    (alter busy-workers conj address)))

(defn fire-lazy-workers
  []
  (dosync
    (util/remove-from-queue available-workers
                            #(< (:expiry %) (System/currentTimeMillis)))))

(defn- find-open-port
  []
  (let [tmp-socket (ServerSocket. 0)
        port       (.getLocalPort tmp-socket)]
    (.close tmp-socket)
    port))

(defn start-workers! [workers port]
  (let [program-path (util/program-path)
        cmd (if (re-find #"clojure.*jar$" program-path)
              ; this means we are running the `boot dev` task, and the "program
              ; path" ends up being clojure-<version>.jar instead of alda; in
              ; this scenario, we can use the `boot dev` task to start each
              ; worker
              ["boot" "dev" "--alda-fingerprint" "-a" "worker" "-p" (str port)]
              ; otherwise, use the same program that was used to start the
              ; server (e.g. /usr/local/bin/alda)
              [program-path "-p" (str port) "--alda-fingerprint" "worker"])]
    (dotimes [_ workers]
      (apply sh/proc cmd))))

(defn supervise-workers
  "Ensures that there are at least `desired` number of workers available by
   counting how many we have and starting more if needed."
  [port desired]
  (let [current (count (all-workers))
        needed  (- desired current)]
    (when (pos? needed)
      (log/infof "Starting %s more workers..." needed)
      (start-workers! needed port))))

(def ^:const WORKER-CHECK-INTERVAL 30000)

(def supervising? (atom true))

(defn start-supervisor!
  [desired-workers worker-port]
  (future
    (while @supervising?
      (try
        (Thread/sleep WORKER-CHECK-INTERVAL)
        (log/debugf "WORKERS: %s" (count (all-workers)))
        (supervise-workers worker-port desired-workers)
        (catch Throwable e
          (log/error e e))))))

(defn shut-down!
  [backend]
  (log/info "Murdering supervisor...")
  (reset! supervising? false)

  (log/info "Murdering workers...")
  (doseq [{:keys [address]} (all-workers)]
    (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
    (.send (ZFrame. "KILL") backend 0)))

(defn start-server!
  ([workers frontend-port]
   (start-server! workers frontend-port (find-open-port)))
  ([workers frontend-port backend-port]
   (let [zmq-ctx        (zmq/zcontext)
         poller         (zmq/poller zmq-ctx 2)
         last-heartbeat (atom 0)]
     (log/infof "Binding frontend socket on port %s..." frontend-port)
     (log/infof "Binding backend socket on port %s..." backend-port)
     (with-open [frontend (doto (zmq/socket zmq-ctx :router)
                            (zmq/bind (str "tcp://*:" frontend-port)))
                 backend  (doto (zmq/socket zmq-ctx :router)
                            (zmq/bind (str "tcp://*:" backend-port)))]
       (zmq/register poller frontend :pollin)
       (zmq/register poller backend :pollin)
       (log/infof "Spawning %s workers..." workers)
       (start-workers! workers backend-port)
       (log/infof "Starting supervisor thread to check on workers every %s ms..."
                  WORKER-CHECK-INTERVAL)
       (start-supervisor! workers backend-port)
       (.addShutdownHook (Runtime/getRuntime)
         (Thread. (fn []
                    (log/info "Interrupt (e.g. Ctrl-C) received.")
                    (shut-down! backend)
                    (try
                      ((.interrupt (. Thread currentThread))
                       (.join (. Thread currentThread)))
                      (catch InterruptedException e)))))
       (try
         (while true
           (zmq/poll poller HEARTBEAT-INTERVAL)
           (when (zmq/check-poller poller 1 :pollin) ; backend
             (when-let [msg (ZMsg/recvMsg backend)]
               (let [address (.unwrap msg)]
                 (if (= 1 (.size msg))
                   (let [frame   (.getFirst msg)
                         data    (-> frame .getData (String.))]
                     (case data
                       "BUSY"      (note-that-worker-is-busy address)
                       "AVAILABLE" (add-or-requeue-worker address)
                       "READY"     (add-or-requeue-worker address)))
                   (do
                     (log/debug "Forwarding backend response to frontend...")
                     (.send msg frontend))))))
           (when (zmq/check-poller poller 0 :pollin) ; frontend
             (when-let [msg (ZMsg/recvMsg frontend)]
               (cond
                 (not (empty? @available-workers))
                 (do
                   (log/debug "Receiving message from frontend...")
                   (let [{:keys [address]}
                         (dosync (util/pop-queue available-workers))]
                     (log/debugf "Forwarding message to worker %s..." address)
                     (.push msg address)
                     (.send msg backend)))

                 (not (empty? @busy-workers))
                 (do
                   (log/debug (str "All workers are currently busy. "
                                   "Letting the client know..."))
                   (let [envelope (.unwrap msg)
                         msg      (doto (ZMsg/newStringMsg
                                          (into-array String
                                                      [all-workers-are-busy-response]))
                                    (.wrap envelope))]
                     (.send msg frontend)))

                 :else
                 (do
                   (log/debug (str "Workers not ready yet. "
                                   "Letting the client know..."))
                   (let [envelope (.unwrap msg)
                         msg      (doto (ZMsg/newStringMsg
                                          (into-array String
                                                      [no-workers-available-response]))
                                    (.wrap envelope))]
                     (.send msg frontend))))))
           (when (> (System/currentTimeMillis)
                    (+ @last-heartbeat HEARTBEAT-INTERVAL))
             (reset! last-heartbeat (System/currentTimeMillis))
             (doseq [{:keys [address]} (all-workers)]
               (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
               (.send (ZFrame. "HEARTBEAT") backend 0)))
           (fire-lazy-workers))
         (catch ZMQException e
           (when (= (.getErrorCode e) (.. ZMQ$Error ETERM getCode))
             (.. Thread currentThread interrupt)))
         (finally
           (log/info "Destroying zmq context...")
           (zmq/destroy zmq-ctx)

           (log/info "Exiting.")
           (System/exit 0)))))))

