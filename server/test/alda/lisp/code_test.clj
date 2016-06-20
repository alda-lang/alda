(ns alda.lisp.code-test
  (:require [clojure.test :refer :all]
            [alda.lisp    :refer :all]))

(deftest code-tests
  (testing "the alda-code function:"
    (let [s      (score
                   (part "piano"
                     (alda-code "c d e")))
          events (:events s)]
      (testing "should parse events out of a string of alda code and splice
                them into the score"
        (is (= 3 (count events)))
        (is (= #{60 62 64} (set (map :midi-note events))))))
    (let [s      (score
                   (alda-code "piano:    c d e
                               clarinet: f g a"))
          events      (:events s)
          instruments (:instruments s)]
      (testing "should parse events out of a string of alda code and splice
                them into the score"
        (is (= 2 (count instruments)))
        (is (= 6 (count events)))
        (is (= #{60 62 64 65 67 69} (set (map :midi-note events))))))))
