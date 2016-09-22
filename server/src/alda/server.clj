(ns alda.server
  (:require [alda.util                 :as    util]
            [alda.version              :refer (-version-)]
            [cheshire.core             :as    json]
            [me.raynes.conch.low-level :as    sh]
            [taoensso.timbre           :as    log]
            [zeromq.device             :as    zmqd]
            [zeromq.zmq                :as    zmq])
  (:import [java.net ServerSocket]
           [java.util.concurrent ConcurrentLinkedQueue]
           [org.zeromq ZFrame ZMQException ZMQ$Error ZMsg]))

; the number of ms between heartbeats
(def ^:const HEARTBEAT-INTERVAL 1000)
; the amount of missed heartbeats before a worker is pronounced dead
(def ^:const WORKER-LIVES          3)

; the number of ms between checking on the number of workers
(def ^:const WORKER-CHECK-INTERVAL 60000)

; the number of ms to wait between starting workers
; (starting them all at the same time causes heavy CPU usage)
(def ^:const WORKER-START-INTERVAL 10000)

; the number of ms since last server heartbeat before we conclude the process has
; been suspended (e.g. laptop lid closed)
(def ^:const SUSPENDED-INTERVAL 10000)

(defn worker-expiration-date []
  (+ (System/currentTimeMillis) (* HEARTBEAT-INTERVAL WORKER-LIVES)))

(defn worker [address]
  {:address address
   :expiry  (worker-expiration-date)})

(def available-workers (util/queue))
(def busy-workers      (ref #{}))
(def worker-blacklist  (ref #{}))
(defn all-workers []   (->> @available-workers
                            (concat (for [address @busy-workers]
                                      {:address address}))
                            (remove #(contains? @worker-blacklist (:address %)))
                            doall))

(defn- friendly-id
  [address]
  (->> address
       .getData
       (map int)
       (apply str)))

; (doseq [[x y] [[:available available-workers]
;                [:busy      busy-workers]
;                [:blacklist worker-blacklist]]]
;   (add-watch y :key (fn [_ _ old new]
;                       (when (not= old new)
;                         (prn x (for [worker new]
;                                  (conj []
;                                        (if-let [address (:address worker)]
;                                          (friendly-id address)
;                                          (friendly-id worker))
;                                        (when-let [expiry (:expiry worker)]
;                                          expiry))))))))

(defn- json-response
  [success?]
  (fn [body]
    (json/generate-string {:success success? :body body})))

(def successful-response   (json-response true))
(def unsuccessful-response (json-response false))

(def pong-response (successful-response "PONG"))

(def shutting-down-response (successful-response "Shutting down..."))

(defn status-response
  [available total]
  (successful-response (format "Server up (%s/%s workers available)"
                               available
                               total)))

(def version-response (successful-response -version-))

(def no-workers-available-response
  (unsuccessful-response
    "No workers processes are ready yet. Please wait a minute."))

(def all-workers-are-busy-response
  (unsuccessful-response
    (str "All worker processes are currently busy. Please wait until playback "
         "is complete and try again.")))

(defn add-or-requeue-worker
  [address]
  (dosync
    (alter busy-workers disj address)
    (util/remove-from-queue available-workers #(= address (:address %)))
    (util/push-queue available-workers (worker address))))

(defn note-that-worker-is-busy
  [address]
  (dosync
    (util/remove-from-queue available-workers #(= address (:address %)))
    (alter busy-workers conj address)))

(defn fire-lazy-workers!
  []
  (dosync
    (util/remove-from-queue available-workers
                            #(< (:expiry %) (System/currentTimeMillis)))))

(defn blacklist-worker!
  [address]
  (dosync (alter worker-blacklist conj address))
  (future
    (Thread/sleep 30000)
    (dosync (alter worker-blacklist disj address))))

(defn lay-off-worker!
  "Lays off the last worker in the queue. Bad luck on its part.

   To 'lay off' a worker, we cannot simply send a KILL message, otherwise the
   worker may die without finishing its workload (e.g. if it's in the middle of
   playing a score). We want the worker to finish what it's doing and then shut
   down.

   To do this, we...
   - remove it from the queue
   - temporarily 'blacklist' its address so it won't be re-queued the next time
     it sends a heartbeat
   - in a separate thread, wait a safe amount of time and then remove the
     address from the blacklist; this is to keep the blacklist from growing too
     big over time

   Once the worker is out of the queue and we aren't letting it back in, the
   worker will stop getting heartbeats from the server and shut itself down."
  []
  (dosync
    (let [{:keys [address]} (util/reverse-pop-queue available-workers)]
      (blacklist-worker! address))))

(defn- find-open-port
  []
  (let [tmp-socket (ServerSocket. 0)
        port       (.getLocalPort tmp-socket)]
    (.close tmp-socket)
    port))

(defn start-workers!
  [workers port]
  (let [program-path (util/program-path)
        cmd (if (re-find #"clojure.*jar$" program-path)
              ; this means we are running the `boot dev` task, and the "program
              ; path" ends up being clojure-<version>.jar instead of alda; in
              ; this scenario, we can use the `boot dev` task to start each
              ; worker
              ["boot" "dev" "--alda-fingerprint" "-a" "worker"
                                                 "--port" (str port)]
              ; otherwise, use the same program that was used to start the
              ; server (e.g. /usr/local/bin/alda)
              [program-path "--port" (str port) "--alda-fingerprint" "worker"])]
    (future
      (dotimes [_ workers]
        (let [{:keys [in out err]} (apply sh/proc cmd)]
          (.close in)
          (.close out)
          (.close err))
        (Thread/sleep WORKER-START-INTERVAL)))))

(defn supervise-workers!
  "Ensures that there are at least `desired` number of workers available by
   counting how many we have and starting more if needed."
  [port desired]
  (let [current (count (all-workers))
        needed  (- desired current)]
    (cond
      (pos? needed)
      (do
        (log/info "Supervisor says we need more workers.")
        (log/infof "Starting %s more worker(s)..." needed)
        (start-workers! needed port))

      (neg? needed)
      (do
        (log/info "Supervisor says there are too many workers.")
        (log/infof "Laying off %s worker(s)..." (- needed))
        (dotimes [_ (- needed)]
          (lay-off-worker!)
          (Thread/sleep 100)))

      :else
      (log/debug "Supervisor approves of the current number of workers."))))

(defn murder-workers!
  [backend]
  (doseq [{:keys [address]} (all-workers)]
    (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
    (.send (ZFrame. "KILL") backend 0)
    (blacklist-worker! address)))

(defn cycle-workers!
  [backend port workers]
  ; kill workers (this might only get the busy ones)
  (murder-workers! backend)
  ; wait for any stray zombie workers to wander in
  (Thread/sleep 500)
  ; clear out the worker queues
  (dosync
    (alter available-workers empty)
    (alter busy-workers empty))
  ; start new workers
  (start-workers! workers port))

(def running? (atom true))

(defn shut-down!
  [backend]
  (log/info "Murdering workers...")
  (murder-workers! backend)
  (reset! running? false))

(defn start-server!
  ([workers frontend-port]
   (start-server! workers frontend-port (find-open-port)))
  ([workers frontend-port backend-port]
   (let [zmq-ctx         (zmq/zcontext)
         poller          (zmq/poller zmq-ctx 2)
         last-heartbeat  (atom (System/currentTimeMillis))
         last-supervised (atom (System/currentTimeMillis))]
     (log/infof "Binding frontend socket on port %s..." frontend-port)
     (log/infof "Binding backend socket on port %s..." backend-port)
     (with-open [frontend (try
                            (doto (zmq/socket zmq-ctx :router)
                              (zmq/bind (str "tcp://*:" frontend-port)))
                            (catch ZMQException e
                              (if (= 48 (.getErrorCode e))
                                (log/error
                                  (str "There is already an Alda server "
                                       "running on this port."))
                                (throw e))
                              (System/exit 1)))
                 backend  (doto (zmq/socket zmq-ctx :router)
                            (zmq/bind (str "tcp://*:" backend-port)))]
       (zmq/register poller frontend :pollin)
       (zmq/register poller backend :pollin)
       (log/infof "Spawning %s workers..." workers)
       (start-workers! workers backend-port)
       (.addShutdownHook (Runtime/getRuntime)
         (Thread. (fn []
                    (when @running?
                      (log/info "Interrupt (e.g. Ctrl-C) received.")
                      (shut-down! backend)))))
       (while @running?
         (zmq/poll poller HEARTBEAT-INTERVAL)
         (when (zmq/check-poller poller 1 :pollin) ; backend
           (when-let [msg (ZMsg/recvMsg backend)]
             (let [address (.unwrap msg)]
               (if (= 1 (.size msg))
                 (let [frame   (.getFirst msg)
                       data    (-> frame .getData (String.))]
                   (when-not (contains? @worker-blacklist address)
                     (case data
                       "BUSY"      (note-that-worker-is-busy address)
                       "AVAILABLE" (add-or-requeue-worker address)
                       "READY"     (add-or-requeue-worker address)
                       (log/errorf "Invalid message: %s" data))))
                 (do
                   (log/debug "Forwarding backend response to frontend...")
                   (.add msg address)
                   (.send msg frontend))))))
         (when (zmq/check-poller poller 0 :pollin) ; frontend
           (when-let [msg (ZMsg/recvMsg frontend)]
             (let [cmd (-> msg .getLast .getData (String.))]
               (case cmd
                 ; the server responds directly to certain commands
                 "ping"
                 (util/respond-to msg frontend pong-response)

                 "play-status"
                 (do
                   (let [client-address (.pop msg)
                         request        (.pop msg)
                         address        (.pop msg)]
                     (log/debugf "Forwarding message to worker %s..." address)
                     (.push msg request)
                     (.push msg client-address)
                     (.push msg address)
                     (.send msg backend)))

                 "status"
                 (util/respond-to msg frontend
                                  (status-response (count @available-workers)
                                                   workers))

                 "stop-server"
                 (do
                   (util/respond-to msg frontend shutting-down-response)
                   (shut-down! backend))

                 "version"
                 (util/respond-to msg frontend version-response)

                 ; any other message is forwarded to the next available
                 ; worker
                 (cond
                   (not (empty? @available-workers))
                   (do
                     (log/debug "Receiving message from frontend...")
                     (let [{:keys [address]}
                           (dosync (util/pop-queue available-workers))]
                       (log/debugf "Forwarding message to worker %s..." address)
                       (.push msg address)
                       (.send msg backend)))

                   ; if no workers are available, respond immediately so the
                   ; client isn't left waiting
                   (not (empty? @busy-workers))
                   (do
                     (log/debug (str "All workers are currently busy. "
                                     "Letting the client know..."))
                     (util/respond-to msg frontend
                                      all-workers-are-busy-response))

                   :else
                   (do
                     (log/debug (str "Workers not ready yet. "
                                     "Letting the client know..."))
                     (util/respond-to msg frontend
                                      no-workers-available-response)))))))

         ; purge workers we haven't heard from in too long
         (fire-lazy-workers!)

         ; detect when the system has been suspended and cycle workers
         ; (fixes a bug where MIDI audio is delayed)
         (when (> (System/currentTimeMillis)
                  (+ @last-heartbeat SUSPENDED-INTERVAL))
           (log/info "Process suspension detected. Cycling workers...")
           (cycle-workers! backend backend-port workers)
           (reset! last-heartbeat  (System/currentTimeMillis))
           (reset! last-supervised (System/currentTimeMillis)))

         ; make sure we still have the desired number of workers
         (when (> (System/currentTimeMillis)
                  (+ @last-supervised WORKER-CHECK-INTERVAL))
           (reset! last-supervised (System/currentTimeMillis))
           (when-not (System/getenv "ALDA_DISABLE_SUPERVISOR")
             (supervise-workers! backend-port workers)))

         ; send a heartbeat to all current workers
         (when (> (System/currentTimeMillis)
                  (+ @last-heartbeat HEARTBEAT-INTERVAL))
           (reset! last-heartbeat (System/currentTimeMillis))
           (doseq [{:keys [address]} (all-workers)]
             (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
             (.send (ZFrame. "HEARTBEAT") backend 0))))

       (log/info "Destroying zmq context...")
       (zmq/destroy zmq-ctx)

       (log/info "Exiting.")
       (System/exit 0)))))

