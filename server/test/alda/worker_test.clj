(ns alda.worker-test
  (:require [clojure.test  :refer :all]
            [cheshire.core :as    json]
            [alda.worker   :as    worker]
            [alda.util     :as    util]
            [alda.version  :refer (-version-)])
  (:import [org.zeromq ZMQ ZContext ZMsg]))

(def ^:dynamic *zmq-context* nil)
(def ^:dynamic *port*        nil)
(def ^:dynamic *socket*      nil)

(defn- init!
  [var val]
  (alter-var-root var (constantly val)))

(defn receive-all
  "Receives one frame, then keeps checking to see if there is more until there
   are no more frames to receive. Returns the result as a vector of strings."
  [socket]
  (loop [frames [(.recvStr socket)]
         more? (.hasReceiveMore socket)]
    (if more?
      (recur (conj frames (.recvStr socket))
             (.hasReceiveMore socket))
      frames)))

(defn response-for-zmsg
  [zmsg]
  (.send zmsg *socket*)
  (loop [res (receive-all *socket*)]
    (if (#{"READY" "AVAILABLE" "BUSY"} (first res))
      (recur (receive-all *socket*))
      res)))

(defn response-for
  [{:keys [command] :as req}]
  (let [zmsg (doto (ZMsg.)
               (.add (.getBytes "fakeclientaddress"))
               (.addString (json/generate-string req))
               (.addString command))]
    (response-for-zmsg zmsg)))

(use-fixtures :once
  (fn [run-tests]
    (init! #'*zmq-context* (ZContext. 1))
    (init! #'*port*        (util/find-open-port))
    (init! #'*socket*      (doto (.createSocket *zmq-context* ZMQ/DEALER)
                             (.bind (format "tcp://*:%s" *port*))))
    (future
      (Thread/sleep 1000) ; to make sure we don't miss the READY signal
      (binding [worker/*no-system-exit* true]
        (worker/start-worker! *port* true)))
    (run-tests)
    (.destroy *zmq-context*)))

(deftest worker-tests
  (testing "a worker process"
    (testing "should send a READY signal when it starts up"
      (let [msg (receive-all *socket*)]
        (is (= "READY" (first msg)))))
    (testing "should send AVAILABLE heartbeats while waiting for work"
      (dotimes [_ 5]
        (let [msg (receive-all *socket*)]
          (is (= "AVAILABLE" (first msg))))))
    ; NOTE: the expected responses are tested in alda.server-test
    (testing "should successfully respond to"
      (testing "a 'parse' command"
        (let [req {:command "parse"
                   :body "piano: c8 d e f g2"
                   :options {:as "lisp"}}
              [_ _ json] (response-for req)
              {:keys [success body]} (json/parse-string json true)]
          (is success)))
      (testing "a 'play' command"
        (let [req {:command "play" :body "piano: (vol 0) c2"}
              [_ _ json] (response-for req)
              {:keys [success body]} (json/parse-string json true)]
          (is success)))
      (testing "a 'play-status' command"
        (let [req {:command "play-status"}
              [_ _ json] (response-for req)
              {:keys [success body] :as res} (json/parse-string json true)]
          (is success))))
    (testing "should accept signals from the server"
      (doseq [signal ["HEARTBEAT" "KILL"]]
        (let [msg (doto (ZMsg.) (.addString signal))]
          (.send msg *socket*))))))
