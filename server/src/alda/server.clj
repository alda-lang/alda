(ns alda.server
  (:require [alda.now                         :as    now]
            [alda.lisp                        :refer :all]
            [alda.parser                      :refer (parse-input)]
            [alda.parser-util                 :refer (parse-with-context)]
            [alda.sound                       :refer (*play-opts*)]
            [alda.util]
            [alda.version                     :refer (-version-)]
            [ring.middleware.defaults         :refer (wrap-defaults api-defaults)]
            [ring.middleware.multipart-params :refer (wrap-multipart-params)]
            [ring.adapter.jetty               :refer (run-jetty)]
            [compojure.core                   :refer :all]
            [compojure.route                  :refer (not-found)]
            [taoensso.timbre                  :as    log]
            [clojure.pprint                   :refer (pprint)]
            [clojure.string                   :as    str]))

; sets log level to TIMBRE_LEVEL (if set) or :warn
(alda.util/set-timbre-level!)

(defn start-alda-environment!
  []
  ; set up audio generators
  (now/set-up! :midi)
  ; initialize a new score
  (score*))

; TODO:
;
; - represent scores as self-contained maps, instead of a combination of things
;   defined at the top-level
;
; - do one of two things:
;
;   a) include the filename as a key in the map;
;      manage the whole thing as an atom
;
;   b) manage the filename and the current score as refs

(def filename (atom nil))

(defn score-info
  []
  {:status      "up"
   :version     -version-
   :filename    @filename
   :line-count  (count (str/split *score-text* #"[\n\r]+"))
   :char-count  (count *score-text*)
   :instruments (for [[k v] (:instruments (score-map))]
                  {:name  k
                   :stock (:stock v)})})

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
      (assoc-in [:headers "Content-Type"] "application/edn")))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn handle-code
  [code & {:keys [append-only?]}]
  (try
    (require '[alda.lisp :refer :all])
    (let [[context parse-result] (parse-with-context code)]
      (if (= context :parse-failure)
        (user-error "Invalid Alda syntax.")
        (do
          (score-text<< code)
          (let [clj-code (case context
                           :music-data (cons 'do parse-result)
                           :score (cons 'do (rest parse-result))
                           parse-result)]
            (if append-only?
              (eval clj-code)
              (now/play! (eval clj-code))))
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

(defn stop-server!
  []
  (log/info "Received request to stop. Shutting down...")
  (future
    (Thread/sleep 300)
    (System/exit 0))
  (success "Shutting down."))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn- get-input [params body]
  (slurp (if-let [file (:file params)]
           (:tempfile file)
           body)))

(defroutes server-routes
  ; ping for server status
  (GET "/" []
    (edn-response (score-info)))

  ; stop the server and exits
  (DELETE "/" []
    (stop-server!))

  ; TODO: make the /parse/map endpoint not overwrite the current score
  ; (will require refactoring alda.lisp to not use top-level vars)

  ; parse code and return the resulting alda.lisp code
  (POST "/parse" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (handle-code-parse code)))
  (POST "/parse/lisp" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (handle-code-parse code :mode :lisp)))

  ; parse + evaluate code and return the resulting score map
  (POST "/parse/map" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (handle-code-parse code :mode :map)))

  ; play the full score (default), or from `from` to `to` params
  (GET "/play" {:keys [play-opts params] :as request}
    (let [{:keys [from to]} params]
      (binding [*play-opts* (assoc play-opts :from from :to to)]
        (let [score-text *score-text*]
          (score*)
          (handle-code score-text)))))

  ; evaluate/play code within the context of the current score
  (POST "/play" {:keys [play-opts params body]:as request}
    (let [code (get-input params body)]
      (binding [*play-opts* play-opts]
        (handle-code code))))

  ; overwrite the current score and play it
  (PUT "/play" {:keys [play-opts params body] :as request}
    (let [code (get-input params body)]
      (score*)
      (binding [*play-opts* play-opts]
        (handle-code code))))

  ; add to the current score without playing anything
  (POST "/add" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (handle-code code :append-only? true)))

  ; get the current score text
  (GET "/score" []
    (success *score-text*))
  (GET "/score/text" []
    (success *score-text*))

  ; get the current score, as alda.lisp code
  (GET "/score/lisp" []
    (handle-code-parse *score-text* :mode :lisp))

  ; get the current score-map
  (GET "/score/map" []
    (edn-response (score-map)))

  ; delete the current score and start a new one
  (DELETE "/score" []
    (score*)
    (success "New score initialized."))

  ; stop the server (alias for DELETE "/")
  (GET "/stop" []
    (stop-server!))

  (GET "/version" []
    (success (str "alda v" -version-)))

  (not-found "Invalid route."))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn wrap-print-request
  "For debug purposes."
  [next-handler]
  (fn [request]
    (prn request)
    (next-handler request)))

(defn wrap-play-opts
  [next-handler play-opts]
  (fn [request]
    (-> (assoc request :play-opts play-opts)
        next-handler)))

(defn app
  [& {:keys [play-opts]}]
  (-> (wrap-defaults server-routes api-defaults)
      (wrap-multipart-params)
      (wrap-play-opts (or play-opts *play-opts*))
      #_(wrap-print-request)))

(defn start-server!
  [port & {:keys [pre-buffer post-buffer]}]
  (log/info "Loading Alda environment...")
  (start-alda-environment!)

  (log/infof "Starting Alda server on port %s..." port)
  (run-jetty (app :play-opts {:pre-buffer  pre-buffer
                              :post-buffer post-buffer
                              :async?      true})
             {:port  port
              :join? false}))

