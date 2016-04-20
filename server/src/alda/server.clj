(ns alda.server
  (:require [alda.lisp                        :refer :all]
            [alda.now                         :as    now]
            [alda.parser                      :refer (parse-input)]
            [alda.parser-util                 :refer (parse-with-context)]
            [alda.sound                       :refer (*play-opts*)]
            [alda.sound.midi                  :as    midi]
            [alda.util]
            [alda.version                     :refer (-version-)]
            [ring.middleware.defaults         :refer (wrap-defaults api-defaults)]
            [ring.middleware.multipart-params :refer (wrap-multipart-params)]
            [ring.adapter.jetty               :refer (run-jetty)]
            [compojure.core                   :refer :all]
            [compojure.route                  :refer (not-found)]
            [taoensso.timbre                  :as    log]
            [clojure.java.io                  :as    io]
            [clojure.pprint                   :refer (pprint)]
            [clojure.string                   :as    str]))

(defn new-server-score
  [& [score]]
  (doto (if score
          (now/new-score score)
          (now/new-score))
    (swap! assoc :score-text ""
                 :filename   nil)))

(def ^:dynamic *current-score* (new-server-score))

(defn score-text<<
  [txt]
  (swap! *current-score*
         update :score-text
         #(str % (when-not (empty? %) \newline) txt)))

(defn close-score!
  []
  (now/tear-down! *current-score*))

(defn new-score!
  []
  (alter-var-root #'*current-score* (constantly (new-server-score))))

(defn load-score!
  [score-text]
  (let [loaded-score (-> score-text parse-input eval new-server-score)]
    (alter-var-root #'*current-score* (constantly loaded-score))
    (swap! *current-score* assoc :score-text score-text)))

(defn start-alda-environment!
  []
  (midi/fill-midi-synth-pool!)
  (while (not (midi/midi-synth-available?))
    (Thread/sleep 250)))

(defn modified?
  []
  (let [{:keys [filename score-text]} @*current-score*]
    (if filename
      (try
        (let [file-contents (slurp (io/file filename))]
          (not= score-text file-contents))
        (catch java.io.FileNotFoundException e
          ; if the file no longer exists (e.g. has been deleted since opening),
          ; consider the score to have been modified -- this should prompt the
          ; user to save changes, which will re-create the file
          true))
      (not (empty? score-text)))))

(defn confirming?
  [{:keys [headers] :as request}]
  (let [confirm-header (get headers "x-alda-confirm" "false")]
    (not (contains? #{"" "no" "false"} (.toLowerCase confirm-header)))))

(defn score-info
  []
  (let [{:keys [filename score-text instruments]} @*current-score*]
    {:status      "up"
     :version     -version-
     :filename    (or filename "(unsaved)")
     :modified?   (modified?)
     :line-count  (count (str/split score-text #"\n\r|\r\n|\n|\r"))
     :char-count  (count score-text)
     :instruments (for [[k v] instruments] {:name  k
                                            :stock (:stock v)})}))

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

(def unsaved-changes-response
  ((response 409) (str "Warning: the score has unsaved changes. Are you sure "
                       "you want to do this?\n\n"
                       "If so, please re-submit your request and include the "
                       "header X-Alda-Confirm:yes.")))

(def existing-file-response
  ((response 409) (str "Warning: there is an existing file with the filename "
                       "you specified. Saving the score to this file will "
                       "erase whatever is already there. Are you sure you "
                       "want to do this?\n\n"
                       "If so, please re-submit your request and include the "
                       "header X-Alda-Confirm:yes.")))

(defn- edn-response
  [x]
  (-> (success (with-out-str (pprint x)))
      (assoc-in [:headers "Content-Type"] "application/edn")))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defn handle-code
  [code & {:keys [replace-score? ; replace the existing score
                  silent?        ; add to/replace the score only, don't play
                  ]}]
  (try
    (require '[alda.lisp :refer :all])
    (let [[context parse-result] (parse-with-context code)]
      (cond
        (and replace-score? (not
                              (or (= context :score)
                                  (= context :part))))
        (user-error "Invalid Alda syntax.")

        (= context :parse-failure)
        (user-error "Invalid Alda syntax.")

        :else
        (do
          (when replace-score?
            (close-score!)
            (new-score!))
          (score-text<< code)
          (let [events (-> (case context
                             :music-data (vec parse-result)
                             :part       parse-result
                             :score      (vec (rest parse-result))
                             parse-result)
                           eval)]
            (if silent?
              (swap! *current-score* continue events)
              (now/with-score *current-score*
                (now/play! events))))
          (success "OK"))))
    (catch Throwable e
      (server-error (.getMessage e)))))

(defn handle-code-parse
  [code & {:keys [mode] :or {mode :lisp}}]
  (try
    (require '[alda.lisp :refer :all])
    (let [parse-result (parse-input code)]
      (edn-response (case mode
                      :lisp parse-result
                      :map (eval parse-result))))
    (catch Throwable e
      (server-error (.getMessage e)))))

(defn stop-server!
  []
  (log/info "Received request to stop. Shutting down...")
  (try
    (future
      (Thread/sleep 300)
      (System/exit 0))
    (catch Throwable e
      (server-error (.getMessage e))))
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

  ; stop the server and exits (alias: GET "/stop")
  (DELETE "/" {:as request}
    (if (and (modified?) (not (confirming? request)))
      unsaved-changes-response
      (do
        (close-score!)
        (stop-server!))))

  ; add to the current score without playing anything
  (POST "/add" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (handle-code code :silent? true)))

  ; (implementation detail, for use by alda client)
  ; sets the filename of the score
  (PUT "/filename" {:keys [params body] :as request}
    (let [filename (get-input params body)]
      (swap! *current-score* assoc :filename filename)
      (success "OK")))

  ; replace the current score with code from a file or string
  (PUT "/load" {:keys [params body] :as request}
    (let [code (get-input params body)]
      (if (and (modified?) (not (confirming? request)))
        unsaved-changes-response
        (do
          (swap! *current-score* assoc :filename nil)
          (handle-code code :silent? true
                            :replace-score? true)))))

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
      (if (and (modified?) (not (confirming? request)))
        unsaved-changes-response
        (handle-code-parse code :mode :map))))

  ; play the full score (default), or from `from` to `to` params
  (GET "/play" {:keys [play-opts params] :as request}
    (let [{:keys [from to]} params]
      (binding [*play-opts* (assoc play-opts :from from :to to)]
        (now/play-score! *current-score*)
        (success "OK"))))

  ; evaluate/play code within the context of the current score
  (POST "/play" {:keys [play-opts params body]:as request}
    (let [code (get-input params body)]
      (binding [*play-opts* play-opts]
        (handle-code code))))

  ; overwrite the current score and play it
  (PUT "/play" {:keys [play-opts params body] :as request}
    (if (and (modified?) (not (confirming? request)))
      unsaved-changes-response
      (let [code (get-input params body)]
        (swap! *current-score* assoc :filename nil)
        (binding [*play-opts* play-opts]
          (handle-code code :replace-score? true)))))

  ; save changes to a score's file
  (GET "/save" []
    (let [{:keys [filename score-text]} @*current-score*]
      (if filename
        (do
          (spit filename score-text)
          (success (str "File saved: " filename)))
        (user-error "You must supply a filename."))))

  ; save changes to a file
  (PUT "/save" {:keys [params body] :as request}
    (let [{:keys [score-text]} @*current-score*
          filename             (str/trim (get-input params body))]
      (if (and (.exists (io/as-file filename)) (not (confirming? request)))
        existing-file-response
        (do
          (spit filename score-text)
          (swap! *current-score* assoc :filename filename)
          (success (str "File saved: " filename))))))

  ; get the current score text
  (GET "/score" []
    (success (:score-text @*current-score*)))
  (GET "/score/text" []
    (success (:score-text @*current-score*)))

  ; get the current score, as alda.lisp code
  (GET "/score/lisp" []
    (handle-code-parse (:score-text @*current-score*) :mode :lisp))

  ; get the current score-map
  (GET "/score/map" []
    (edn-response @*current-score*))

  ; delete the current score and start a new one
  (DELETE "/score" {:as request}
    (if (and (modified?) (not (confirming? request)))
      unsaved-changes-response
      (do
        (close-score!)
        (new-score!)
        (success "New score initialized."))))

  ; stop the server (alias for DELETE "/")
  (GET "/stop" {:as request}
    (if (and (modified?) (not (confirming? request)))
      unsaved-changes-response
      (do
        (close-score!)
        (stop-server!))))

  (GET "/version" []
    (success -version-))

  (not-found (user-error "Invalid route.")))

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

