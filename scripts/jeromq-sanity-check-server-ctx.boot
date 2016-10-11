#!/usr/bin/env boot

(set-env! :dependencies '[[org.zeromq/jeromq "0.3.5"]])

(import '[org.zeromq ZMQ ZContext])

(defn print-usage
  []
  (println "Example usage:")
  (println "  $ boot scripts/jeromq-sanity-check-server.boot 12345")
  (println)
  (println "  # in another terminal:")
  (println "  $ boot scripts/jeromq-sanity-check-client.boot 12345"))

(defn hello-world-test
  [socket]
  (println "Hello world client/server example:")
  (dotimes [n 5]
    (let [req (.recv socket)
          res (format "Hello #%s from server" (inc n))]
      (println "Received request:" (String. req))
      (Thread/sleep 1000) ; simulate doing work
      (println "Sending response:" res)
      (.send socket res))))

(defn run-tests
  [port ctx-desc ctx]
  (println)
  (println (format "Testing with %s..." ctx-desc))
  (println)

  (with-open [socket (doto (.socket ctx ZMQ/REP)
                       (.bind (format "tcp://*:%s" port)))]
    (hello-world-test socket))

  (println)
  (print "Destroying context... ")
  (flush)
  (.term ctx)
  (println "done."))

(defn -main
  ([]
   (println "No port specified.")
   (println)
   (print-usage)
   (System/exit 1))
  ([port]
   (println (format "Using port %s for tests." port))
   (println
     (format "(In another terminal, run the client script on port %s.)" port))

   (run-tests port "Ctx (ZMQ.createContext(1))" (ZMQ/context 1))

   (println)
   (println "Done.")))
