(ns alda.version
  (:require [clojure.java.io :as io]
            [clojure.string  :as str]))

(def ^:const -version-
  (str/trim (slurp (io/resource "version.txt"))))
