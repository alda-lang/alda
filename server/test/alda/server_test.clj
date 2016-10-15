(ns alda.server-test
  (:require [clojure.test  :refer :all]
            [cheshire.core :as    json]
            [alda.server   :as    server]
            [alda.util     :as    util]
            [alda.version  :refer (-version-)])
  (:import [org.zeromq ZMQ ZContext ZMsg]))

(def ^:dynamic *zmq-context*     nil)
(def ^:dynamic *frontend-port*   nil)
(def ^:dynamic *backend-port*    nil)
(def ^:dynamic *frontend-socket* nil)
(def ^:dynamic *backend-socket*  nil)

(defn- init!
  [var val]
  (alter-var-root var (constantly val)))

(defn start-server!
  []
  (future (server/start-server! 2 *frontend-port*)))

(defn receive-all
  "Receives one frame, then keeps checking to see if there is more until there
   are no more frames to receive. Returns the result as a vector of strings."
  [socket]
  (loop [frames [(.recv socket)]
         more? (.hasReceiveMore socket)]
    (if more?
      (recur (conj frames (.recv socket))
             (.hasReceiveMore socket))
      frames)))

(defn play-status-request
  [worker-address]
  (doto (ZMsg.)
    (.addString (json/generate-string {:command "play-status"}))
    (.add worker-address)
    (.addString "play-status")))

(defn response-for-zmsg
  [zmsg]
  (.send zmsg *frontend-socket*)
  (-> (receive-all *frontend-socket*)
      second
      (String.)
      (json/parse-string true)))

(defn response-for
  [{:keys [command] :as req}]
  (let [zmsg (doto (ZMsg.)
               (.addString (json/generate-string req))
               (.addString command))]
    (response-for-zmsg zmsg)))

(defn complete-response-for
  [{:keys [command] :as req}]
  (let [msg (doto (ZMsg.)
              (.addString (json/generate-string req))
              (.addString command))]
    (.send msg *frontend-socket*))
  (receive-all *frontend-socket*))

(defn get-backend-port
  []
  (let [req {:command "status"}
        res (response-for req)]
    (->> (:body res)
         (re-find #"backend port: (\d+)")
         second)))

(defn wait-for-a-worker
  []
  (let [workers (->> (response-for {:command "status"})
                     :body
                     (re-find #"(\d+)/\d+ workers available")
                     second
                     Integer/parseInt)]
    (when (zero? workers) (Thread/sleep 100) (recur))))

(use-fixtures :once
  (fn [run-tests]
    (init! #'*zmq-context*     (ZContext. 1))
    (init! #'*frontend-port*   (util/find-open-port))
    (init! #'*frontend-socket* (doto (.createSocket *zmq-context* ZMQ/DEALER)
                                 (.connect (format "tcp://*:%s" *frontend-port*))))
    (start-server!)
    (init! #'*backend-port*   (get-backend-port))
    (init! #'*backend-socket* (doto (.createSocket *zmq-context* ZMQ/DEALER)
                                 (.connect (format "tcp://*:%s" *backend-port*))))
    (run-tests)
    (.destroy *zmq-context*)))

(deftest frontend-tests
  (testing "the 'ping' command"
    (let [req {:command "ping"}]
      (testing "gets a successful response"
        (is (:success (response-for req))))))
  (testing "the 'status' command"
    (let [req {:command "status"}
          res (-> (response-for req) :body)]
      (testing "says the server is up"
        (is (re-find #"Server up" res)))
      (testing "has a response that includes"
        (testing "the number of available workers"
          (is (re-find #"\d+/\d+ workers available" res)))
        (testing "the backend port number"
          (is (re-find #"backend port: \d+" res))))))
  (testing "the 'version' command"
    (let [req {:command "version"}
          res (-> (response-for req) :body)]
      (testing "reports the current version"
        (is (re-find (re-pattern -version-) res)))))
  (testing "until the first worker process is available,"
    (testing "the 'parse' command"
      (let [req {:command "parse" :body "it doesn't matter what i put here"}]
        (testing "gets a 'no workers available yet' response"
          (let [{:keys [success body]} (response-for req)]
            (is (not success))
            (is (re-find #"No worker processes are ready yet" body)))))))
  (testing "once there is a worker available,"
    (println "Waiting for a worker process to become available...")
    (wait-for-a-worker)
    (println "Worker ready.")
    (testing "the 'parse' command"
      (testing "with the :as 'lisp' option"
        (let [req {:command "parse" :body "piano: c" :options {:as "lisp"}}
              {:keys [success body]} (response-for req)]
          (testing "should get a successful response containing parsed code"
            (is success)
            (is (= body "(alda.lisp/score\n (alda.lisp/part\n  {:names [\"piano\"]}\n  (alda.lisp/note (alda.lisp/pitch :c))))\n")))))
      (wait-for-a-worker)
      (testing "with the :as 'map' option"
        (let [req {:command "parse" :body "piano: c" :options {:as "map"}}
              {:keys [success body]} (response-for req)]
          (testing "should get a successful response containing a score map"
            (is success)
            (let [score-map (json/parse-string body true)]
              (is (contains? score-map :events))
              (is (contains? score-map :instruments)))))))
    (wait-for-a-worker)
    ; forcing parsing to take at least 2 seconds for play-status test below
    (let [req {:command "play" :body "piano: (Thread/sleep 2000) (vol 0) c2"}
          [_ json worker-address] (complete-response-for req)
          {:keys [success body]}  (json/parse-string (String. json) true)]
      (testing "the play command"
        (testing "should get a successful response"
          (is success)
          (testing "that includes the address of the worker playing the score"
            (not (nil? worker-address)))))
      (testing "the play-status command"
        (let [req (play-status-request worker-address)
              {:keys [success pending body]} (response-for-zmsg req)]
          (testing "should get a successful response"
            (is success))
          (testing "should say the status is 'parsing' while the worker is parsing"
            (is (= body "parsing"))
            (testing "and 'pending' should be true"
              (is pending))))))))

(deftest backend-tests
  (comment
    "In these tests, we are acting like a worker process interacting with
     the server we are testing.

     The backend receives two different types of messages from workers:
     - responses to forward to the client that made the request
     - heartbeats

     It is difficult to test the first type of message, because the server
     takes it and forwards it to the client, and does not send anything back
     to the worker to let us know that this happened.

     The frontend tests above actually confirm that this works properly; if
     the server has at least one worker running, then if you send a message
     to the frontend and the server sends a response from the worker, then
     we know that the first type of message works.

     Testing the second type of message here. As a worker, if we send a READY
     or AVAILABLE heartbeat to the server, it should start sending us HEARTBEAT
     signals.")
  (testing "the backend socket"
    (testing "can send and receive heartbeats"
      (let [ready-signal (doto (ZMsg.) (.addString "READY"))]
        (.send ready-signal *backend-socket*))
      (let [msg (receive-all *backend-socket*)]
        (is (= "HEARTBEAT" (-> msg first String.))))
      (dotimes [_ 5]
        (let [heartbeat (doto (ZMsg.) (.addString "AVAILABLE"))]
          (.send heartbeat *backend-socket*))
        (let [msg (receive-all *backend-socket*)]
          (is (= "HEARTBEAT" (-> msg first String.))))))
    (testing "can send DONE signal"
      (let [done-signal (doto (ZMsg.) (.addString "DONE"))]
        (.send done-signal *backend-socket*)))))
