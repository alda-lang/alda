(ns alda.lisp.score-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]))

(deftest score-tests
  (testing "a score:"
    (score*)
    (part* "piano/violin/cello")
    (note (pitch :c))
    (note (pitch :d))
    (note (pitch :e))
    (note (pitch :f))
    (note (pitch :g))
    (note (pitch :a))
    (note (pitch :b))
    (octave :up)
    (note (pitch :c))
    (let [score (score-map)]
      (testing "it has the right number of instruments"
        (is (= 3 (count (:instruments score)))))
      (testing "it has the right number of events"
        (is (= (* 3 8) (count (:events score))))))))
