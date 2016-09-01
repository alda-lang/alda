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

; the number of ms between checking on the number of workers
(def ^:const WORKER-CHECK-INTERVAL 30000)


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

; (doseq [[x y] [[:available available-workers]
;                [:busy      busy-workers]
;                [:blacklist worker-blacklist]]]
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
      (alter worker-blacklist conj address)
      (future
        (Thread/sleep 30000)
        (dosync (alter worker-blacklist disj address))))))

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
              ["boot" "dev" "--alda-fingerprint" "-a" "worker" "-p" (str port)]
              ; otherwise, use the same program that was used to start the
              ; server (e.g. /usr/local/bin/alda)
              [program-path "-p" (str port) "--alda-fingerprint" "worker"])]
    (dotimes [_ workers]
      (apply sh/proc cmd))))

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

(defn shut-down!
  [backend]
  (log/info "Murdering workers...")
  (doseq [{:keys [address]} (all-workers)]
    (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
    (.send (ZFrame. "KILL") backend 0)))

(defn start-server!
  ([workers frontend-port]
   (start-server! workers frontend-port (find-open-port)))
  ([workers frontend-port backend-port]
   (let [zmq-ctx         (zmq/zcontext)
         poller          (zmq/poller zmq-ctx 2)
         last-heartbeat  (atom 0)
         last-supervised (atom (System/currentTimeMillis))]
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

           ; purge workers we haven't heard from in too long
           (fire-lazy-workers!)

           ; make sure we still have the desired number of workers
           (when (> (System/currentTimeMillis)
                    (+ @last-supervised WORKER-CHECK-INTERVAL))
             (reset! last-supervised (System/currentTimeMillis))
             (supervise-workers! backend-port workers))

           ; send a heartbeat to all current workers
           (when (> (System/currentTimeMillis)
                    (+ @last-heartbeat HEARTBEAT-INTERVAL))
             (reset! last-heartbeat (System/currentTimeMillis))
             (doseq [{:keys [address]} (all-workers)]
               (.send address backend (+ ZFrame/REUSE ZFrame/MORE))
               (.send (ZFrame. "HEARTBEAT") backend 0))))

         (catch ZMQException e
           (when (= (.getErrorCode e) (.. ZMQ$Error ETERM getCode))
             (.. Thread currentThread interrupt)))

         (finally
           (log/info "Destroying zmq context...")
           (zmq/destroy zmq-ctx)

           (log/info "Exiting.")
           (System/exit 0)))))))

