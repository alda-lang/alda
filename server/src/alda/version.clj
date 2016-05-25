(ns alda.version
  (:require [clojure.java.io :as io]
            [clojure.string  :as str]
            [manifest.core   :refer (manifest)]))

(def ^:const -version-
  (:alda-version (manifest "alda.Main")))
