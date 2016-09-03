(ns alda.util
  (:require [clojure.string                              :as str]
            [taoensso.timbre                             :as timbre]
            [taoensso.timbre.appenders.core              :as appenders]
            [taoensso.timbre.appenders.3rd-party.rolling :as rolling])
  (:import (java.io File)
           (java.nio.file Paths)
           (org.zeromq ZMsg)))

 (defmacro while-let
  "Repeatedly executes body while test expression is true. Test
  expression is bound to binding.

  (copied from boot.util)"
  [[binding test] & body]
  `(loop [~binding ~test]
     (when ~binding ~@body (recur ~test))))

(defmacro pdoseq
  "A fairly efficient hybrid of `doseq` and `pmap`"
  [binding & body]
  `(doseq ~binding (future @body)))

(defmacro pdoseq-block
  "A fairly efficient hybrid of `doseq` and `pmap`, that blocks.

   If an error occurs on an async thread, it is rethrown on the main thread."
  [binding & body]
  `(let [remaining# (atom (count ~(second binding)))
         error#     (atom nil)]
     (doseq ~binding
       (future
         (try
           ~@body
           (swap! remaining# dec)
           (catch Throwable e#
             (reset! error# e#)))))
     (when (seq ~(second binding))
       (loop []
         (cond
           @error#             (throw @error#)
           (zero? @remaining#) nil
           :else               (recur))))))

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
      (if (str/starts-with? position-str ":")
        (keyword (subs position-str 1))
        (keyword position-str)))))

(defn =%
  "Returns true if all arguments are within 0.01 of each other."
  [& xs]
  (let [[x & xs] (sort xs)]
    (apply <= x (conj (vec xs) (+ x 0.01)))))

(defn set-log-level!
  ([]
    (timbre/set-level! (if-let [level (System/getenv "TIMBRE_LEVEL")]
                         (keyword (str/replace level #":" ""))
                         :warn)))
  ([level]
    (timbre/set-level! level)))

(defn log-to-file!
  [filename]
  (timbre/merge-config!
    {:appenders {:spit (appenders/spit-appender {:fname filename})}
     :output-fn (partial timbre/default-output-fn {:stacktrace-fonts {}})}))

(defn rolling-log!
  [filename]
  (timbre/merge-config!
    {:appenders {:spit (rolling/rolling-appender {:path    filename
                                                  :pattern :weekly})}
     :output-fn (partial timbre/default-output-fn {:stacktrace-fonts {}})}))

(defn program-path
  "utility function to get the filename of jar in which this function is invoked
   (source: http://stackoverflow.com/a/13276993/2338327)"
  [& [ns]]
  (-> (or ns (class *ns*))
      .getProtectionDomain .getCodeSource .getLocation .getPath))

(defn alda-home-path
  "Returns the path to a folder/file inside the Alda home directory, or the
   directory itself if no arguments are provided.

   e.g. on a Unix system:
   (alda-home-path) => ~/.alda
   (alda-home-path \"logs\" \"error.log\") => ~/.alda/logs/error.log

   e.g. on a Windows system:

   (alda-home-path) => C:\\dave\\.alda
   (alda-home-path \"logs\" \"error.log\") => C:\\dave\\.alda\\logs\\error.log"
  [& segments]
  (->> (cons ".alda" segments)
       (into-array String)
       (Paths/get (System/getProperty "user.home"))
       str))

(defn queue
  "A janky custom data structure that acts like a queue and is a ref.

   It's really just a vector wrapped in a ref.

   Items are popped from the left and pushed onto the right."
  ([]
   (ref []))
  ([init]
   (ref (into [] init))))

(defn push-queue
  [q x]
  (alter q #(conj % x)))

(defn pop-queue
  "Pops off the first item in the queue and returns it."
  [q]
  (let [x (first @q)]
    (alter q #(vec (drop 1 %)))
    x))

(defn reverse-pop-queue
  "Pops off the LAST item in the queue and returns it."
  [q]
  (let [x (last @q)]
    (alter q #(vec (drop-last 1 %)))
    x))

(defn check-queue
  "Returns true if there is at least one item in the queue that satisfies the
   predicate."
  [q pred]
  (boolean (some pred @q)))

(defn re-queue
  "Finds all items in the queue that satisfy the predicate, and re-queues them
   onto the end of the queue.

   When a second function `f` is provided, it is called on each re-queued
   element. This can be used e.g. to update the timestamps of queued items."
  [q pred & [f]]
  (alter q #(let [yes (for [x % :when (pred x)]
                        ((or f identity) x))
                  no  (filter (complement pred) %)]
              (vec (concat no yes)))))

(defn remove-from-queue
  "Removes all items from the queue that satisfy the predicate."
  [q pred]
  (alter q #(vec (filter (complement pred) %))))

(defn respond-to
  [msg socket response & [envelope]]
  (let [envelope (or envelope (.unwrap msg))
        msg      (doto (ZMsg/newStringMsg (into-array String [response]))
                   (.wrap envelope))]
    (.send msg socket)))

