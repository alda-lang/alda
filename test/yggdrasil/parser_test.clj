(ns yggdrasil.parser-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [yggdrasil.parser :refer :all]))

(deftest parser-test
  (testing "parsing of valid input"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/yggdrasil/awobmolg.yd"))]
        (is (not (instance? instaparse.gll.Failure result)))
        (pprint result)))))
