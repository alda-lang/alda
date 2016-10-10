#!/usr/bin/env boot

(set-env! :dependencies '[[cheshire          "5.6.3"]
                          [org.zeromq/jeromq "0.3.5"]])

(require '[cheshire.core :as json])
(import '[org.zeromq ZMQ ZContext ZMsg ZMQException])

(def frontend-msgs
  (let [jsons [{:command "parse" :body "piano: c8 d e f g2" :options {:as "lisp"}}
               {:command "play" :body "piano: c8 d e f g2"}]]
    (for [{:keys [command] :as json} jsons]
      (doto (ZMsg.)
        (.add (.getBytes "fakeclientaddress"))
        (.addString (json/generate-string json))
        (.addString command)))))

(def server-signals
  (for [signal ["HEARTBEAT" "KILL"]]
    (doto (ZMsg.) (.addString signal))))

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

(defn print-usage
  []
  (println "Example usage:")
  (println "  $ boot scripts/test-worker.boot 12345")
  (println)
  (println "  # then start a worker in another terminal:")
  (println "  $ alda -v -p 12345 worker"))

(defn test-msg
  [msg socket expect-response?]
  (println)
  (println "Sending msg:" \newline msg)
  (.send msg socket)
  (when expect-response?
    (loop [response (receive-all socket)]
      (if (#{"READY" "AVAILABLE" "BUSY"} (first response))
        (recur (receive-all socket))
        (do
          (println "Response received:" \newline response)
          response)))))

(defn expect-signals
  [socket & signals]
  ; blocks until a message is received from worker
  (let [msg    (receive-all socket)
        signal (first msg)]
    (when-not ((set signals) signal)
      (println "Unexpected message:" msg)
      (System/exit 1))))

(defn -main
  ([]
   (println "No port specified.")
   (println)
   (print-usage)
   (System/exit 1))
  ([port]
   (let [ctx      (ZContext. 1)]
     (with-open [socket (try
                          (doto (.createSocket ctx ZMQ/DEALER)
                            (.bind (format "tcp://*:%s" port)))
                          (catch ZMQException e
                            (if (= 48 (.getErrorCode e))
                              (do
                                (println
                                  (format "Port %s is already in use." port)
                                  "Please choose a different port number.")
                                (System/exit 1))
                              (throw e))))]
       (println
         (format "Please start an Alda worker process listening on port %s."
                 port))
       (println)
       (println "Waiting for READY signal from worker...")
       (expect-signals socket "READY")
       (println "READY signal received.")
       (Thread/sleep 500)
       (println)
       (println "Measuring worker's heartbeat...")
       (dotimes [_ 5]
         (expect-signals socket "READY" "AVAILABLE" "BUSY")
         (.send (doto (ZMsg.) (.addString "HEARTBEAT")) socket)
         (println "th-thump"))
       (Thread/sleep 500)
       (doseq [msg (concat frontend-msgs)]
         (test-msg msg socket true)
         (Thread/sleep 500))
       (doseq [msg (concat server-signals)]
         (test-msg msg socket false)
         (Thread/sleep 500))))
   (println)
   (println "Done.")))
