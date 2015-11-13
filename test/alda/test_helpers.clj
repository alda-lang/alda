(ns alda.test-helpers
  (:require [alda.parser :refer :all]
            [alda.lisp :refer :all]
            [instaparse.core :as insta]
            [clojure.java.io :as io]))

(defn get-instrument
  "Returns the first instrument in *instruments* whose id starts with inst-name."
  [inst-name]
  (first (for [[id instrument] *instruments*
               :when (.startsWith id (str inst-name \-))]
           instrument)))

