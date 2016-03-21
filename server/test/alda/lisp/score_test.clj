(ns alda.lisp.score-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]))

(deftest score-tests
  (testing "a score:"
    (let [s (score
              (part "piano/violin/cello"
                (note (pitch :c))
                (note (pitch :d))
                (note (pitch :e))
                (note (pitch :f))
                (note (pitch :g))
                (note (pitch :a))
                (note (pitch :b))
                (octave :up)
                (note (pitch :c))))]
      (testing "it has the right number of instruments"
        (is (= 3 (count (:instruments s)))))
      (testing "it has the right number of events"
        (is (= (* 3 8) (count (:events s))))))))
