(ns alda.parser-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.parser :refer :all]
            [instaparse.core :as insta]))

(deftest parser-test
  (testing "parsing valid input"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/examples/awobmolg.alda"))]
        (is (not (insta/failure? result)))
        (pprint result)))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/examples/debussy_quartet.alda"))]
        (is (not (insta/failure? result)))
        (pprint result)))))
