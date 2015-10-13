(ns alda.util
  (:require [clojure.string  :as str]
            [taoensso.timbre :as timbre])
  (:import (java.io File)))

(defmacro pdoseq
  "A fairly efficient hybrid of `doseq` and `pmap`"
  [binding & body]
  `(doseq ~binding (future @body)))

(defmacro pdoseq-block
  "A fairly efficient hybrid of `doseq` and `pmap`, that blocks."
  [binding & body]
  `(let [latch# (atom (count ~(second binding)))
         done# (promise)]
     (doseq ~binding
       (future
         ~@body
         (when (zero? (swap! latch# dec))
           (deliver done# true))))
     ;; don't block if unless loop will run and check latch
     (when (seq ~(second binding))
       @done#)))

(defmacro resetting [vars & body]
  (if (seq vars)
    (let [[x & xs] vars]
      `(let [before# ~x
             result# (resetting ~xs ~@body)]
         (alter-var-root (var ~x) (constantly before#))
         result#))
    `(do ~@body)))

(defn strip-nil-values
  "Strip `nil` values from a map."
  [hsh]
  (into (empty hsh) (remove (comp nil? last)) hsh))

(defn parse-str-opts
  "Transform string based keyword arguments into a regular map, eg.
   IN:  \"from 0:20 to :third-movement some-junk-at-end\"
   OUT: {:from  \"0:20\"
         :to \":third-movement\"}"
  [opts-str]
  (let [pairs (partition 2 (str/split opts-str #"\s"))]
    (into {} (map (fn [[k v]] [(keyword k) v])) pairs)))

(defn parse-time
  "Convert a human readable duration into milliseconds, eg. \"02:31\" => 151 000"
  [time-str]
  (let [[s m h] (as-> (str/split time-str #":") x
                      (reverse x)
                      (map #(Double/parseDouble %) x)
                      (concat x [0 0 0]))]
    (* (+ (* 60 (+ (* 60 h) m)) s) 1000)))

(def ^:private duration-regex
  #"^(\d+(\.\d+)?)(:\d+(\.\d+)?)*$")

(defn parse-position
  "Convert a string denoting a position in a song into the appropriate type.
   For explicit timepoints this is a double denoting milliseconds, and for
   markers this is a keyword."
  [position-str]
  (when position-str
    (if (re-find duration-regex position-str)
      (parse-time position-str)
      (if (.startsWith position-str ":")
        (keyword (subs position-str 1))
        (keyword position-str)))))

(defn =%
  "Returns true if all arguments are within 0.01 of each other."
  [& xs]
  (let [[x & xs] (sort xs)]
    (apply <= x (conj (vec xs) (+ x 0.01)))))

(defn set-timbre-level!
  []
  (timbre/set-level! (if-let [level (System/getenv "TIMBRE_LEVEL")]
                       (keyword (str/replace level #":" ""))
                       :warn)))

(defn check-for
  "Checks to see if a given file already exists. If it does, prompts the user
   whether he/she wants to overwrite the file. If he/she doesn't, then prompts
   the user to choose a new filename and calls itself to check the new file, etc.
   Returns a filename that does not exist, or does exist and the user is OK with
   overwriting it."
  [filename]
  (letfn [(prompt []
            (print "> ") (flush) (read-line))
          (overwrite-dialog []
            (println
              (format "File \"%s\" already exists. Overwrite? (y/n)" filename))
            (let [response (prompt)]
              (cond
                (re-find #"(?i)y(es)?" response)
                filename

                (re-find #"(?i)no?" response)
                (do
                  (println "Please specify a different filename.")
                  (check-for (prompt)))

               :else
               (do
                 (println "Answer the question, sir.")
                 (recur)))))]
    (cond
      (.isFile (File. filename))
      (overwrite-dialog)

      (.isDirectory (File. filename))
      (do
        (printf "\"%s\" is a directory. Please specify a filename.\n" filename)
        (recur (prompt)))

      :else
      filename)))
