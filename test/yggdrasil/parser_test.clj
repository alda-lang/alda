(ns yggdrasil.parser-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [yggdrasil.parser :refer :all]))

(deftest parser-test
  (testing "parsing valid input"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/yggdrasil/awobmolg.yd"))]
        (is (not (instance? instaparse.gll.Failure result)))
        (pprint result)))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/yggdrasil/debussy_quartet.yd"))]
        (is (not (instance? instaparse.gll.Failure result)))
        (pprint result)))))
