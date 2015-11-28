(ns alda.server
  (:require [alda.now                 :as    now]
            [alda.lisp                :refer :all]
            [alda.parser              :refer (parse-input)]
            [alda.parser-util         :refer (parse-with-context)]
            [alda.sound               :refer (*play-opts*)]
            [alda.version             :refer (-version-)]
            [ring.middleware.defaults :refer (wrap-defaults api-defaults)]
            [ring.adapter.jetty       :refer (run-jetty)]
            [compojure.core           :refer :all]
            [compojure.route          :refer (not-found)]
            [taoensso.timbre          :as    log]
            [clojure.pprint           :refer (pprint)]))

(defn start-alda-environment!
  []
  ; set up audio generators
  (now/set-up! :midi)
  ; initialize a new score
  (score*))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- response
  [code]
  (fn [body]
    {:status  code
     :headers {"Content-Type"   "text/html"
               "X-Alda-Version" -version-}
     :body    body}))

(def ^:private success      (response 200))
(def ^:private user-error   (response 400))
(def ^:private server-error (response 500))

(defn- edn-response
  [x]
  (-> (success (with-out-str (pprint x)))
      (update :headers
              assoc "Content-Type" "application/edn")))

(defn handle-code
  [code]
  (try
    (require '[alda.lisp :refer :all])
    (let [[context parse-result] (parse-with-context code)]
      (if (= context :parse-failure)
        (user-error "Invalid Alda syntax.")
        (do
          (now/play! (eval (case context
                           :music-data (cons 'do parse-result)
                           :score (cons 'do (rest parse-result))
                           parse-result)))
          (success "OK"))))
    (catch Throwable e
      (server-error (str "ERROR: " (.getMessage e))))))

(defn handle-code-parse
  [code & {:keys [mode] :or {mode :lisp}}]
  (try
    (require '[alda.lisp :refer :all])
    (let [parse-result (parse-input code)]
      (edn-response (case mode
                      :lisp parse-result
                      :map (eval parse-result))))
    (catch Throwable e
      (server-error (str "ERROR: " (.getMessage e))))))

(defroutes server-routes
  ; get the current score-map
  (GET "/" []
    (edn-response (score-map)))
  ; delete the current score and start a new one
  (DELETE "/" []
    (score*)
    (success "New score initialized."))
  ; evaluate code within the context of the current score
  (POST "/" {:keys [play-opts body]:as request}
    (let [code (slurp body)]
      (binding [*play-opts* play-opts]
        (handle-code code))))
  ; overwrite the current score
  (PUT "/" {:keys [play-opts body] :as request}
    (let [code (slurp body)]
      (score*)
      (binding [*play-opts* play-opts]
        (handle-code code))))

  (POST "/parse" {:keys [body] :as request}
    (let [code (slurp body)]
      (handle-code-parse code)))

  (POST "/parse/lisp" {:keys [body] :as request}
    (let [code (slurp body)]
      (handle-code-parse code :mode :lisp)))

  (POST "/parse/map" {:keys [body] :as request}
    (let [code (slurp body)]
      (handle-code-parse code :mode :map)))

  (GET "/version" []
    (success (str "alda v" -version-)))

  (GET "/stop" []
    (log/info "Received request to stop. Shutting down...")
    (future
      (Thread/sleep 1000)
      (System/exit 0))
    (success "Shutting down."))

  (not-found "Invalid route."))

(defn wrap-play-opts
  [next-handler play-opts]
  (fn [request]
    (-> (assoc request :play-opts play-opts)
        next-handler)))

(def app
  (-> (wrap-defaults server-routes api-defaults)
      (wrap-play-opts *play-opts*)))

(defn start-server!
  [port]
  (log/info "Loading Alda environment...")
  (start-alda-environment!)
  (log/infof "Starting Alda server on port %s..." port)
  (run-jetty app {:port port}))

