#!/usr/bin/env boot

(set-env! :dependencies '[[cheshire          "5.6.3"]
                          [org.zeromq/jeromq "0.3.5"]])

(require '[cheshire.core :as json])
(import '[org.zeromq ZMQ ZContext ZMsg])

(def frontend-msgs
  (let [jsons [{:command "parse" :body "piano: c8 d e f g2" :options {:as "lisp"}}
               {:command "ping"}
               {:command "status"}
               {:command "version"}]]
    (for [{:keys [command] :as json} jsons]
      (doto (ZMsg.)
        (.addString (json/generate-string json))
        (.addString command)))))

(def frontend-play-msg
  (let [{:keys [command] :as json} {:command "play" :body "piano: c8 d e f g2"}]
    (doto (ZMsg.)
      (.addString (json/generate-string json))
      (.addString command))))

(defn frontend-play-status-msg
  [worker-address]
  (let [{:keys [command] :as json} {:command "play-status"}]
    (doto (ZMsg.)
      (.addString (json/generate-string json))
      (.add worker-address)
      (.addString command))))

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

(defn print-usage
  []
  (println "Example usage:")
  (println "  $ alda -v -p 23232 server")
  (println "  # take note of the frontend and backend ports")
  (println "  # let's say for example the backend port it uses is 56667")
  (println)
  (println "  # in another terminal:")
  (println "  $ boot scripts/test-server.boot 23232 56667"))

(defn test-msg
  [msg socket]
  (println)
  (println "Sending msg:" \newline msg)
  (.send msg socket)
  (let [response (receive-all socket)]
    (println "Response received:" \newline (map #(String. %) response))
    response))

(defn -main
  ([]
   (println "No ports specified.")
   (println)
   (print-usage)
   (System/exit 1))
  ([port]
   (println "Only one port specified.")
   (println)
   (print-usage)
   (System/exit 1))
  ([port1 port2]
   (let [ctx      (ZContext. 1)
         frontend (doto (.createSocket ctx ZMQ/DEALER)
                    (.connect (format "tcp://*:%s" port1)))
         backend  (doto (.createSocket ctx ZMQ/DEALER)
                    (.connect (format "tcp://*:%s" port2)))]
     (println
       (format "Testing server with frontend port %s, backend port %s..."
               port1 port2))
     (println)
     (println "Testing frontend port...")
     (doseq [msg frontend-msgs]
       (test-msg msg frontend)
       (Thread/sleep 500))
     ; doing play msg separately so we can note the worker address...
     (let [response (test-msg frontend-play-msg frontend)
           ; ...and use it to send the play-status msg
           msg      (frontend-play-status-msg (last response))]
       (test-msg msg frontend)
       (Thread/sleep 500))
     (println)
     (println "Testing backend port...")
     (println)

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

        Testing the second type of message here. As a worker, if we send an
        AVAILABLE heartbeat to the server, it should start sending us HEARTBEAT
        signals.")

     (let [msg (doto (ZMsg.) (.addString "READY"))]
       (println "Sending msg:" \newline msg)
       (.send msg backend)
       (loop [n 5]
         (when (pos? n)
           (let [msg (doto (ZMsg.) (.addString "AVAILABLE"))]
             (println)
             (println "Sending msg:" \newline msg)
             (.send msg backend)
             (println)
             (let [msg (receive-all backend)]
               (println "Received msg:" \newline (map #(String. %) msg))))
           (recur (dec n))))))
   (println)
   (println "Done.")))
