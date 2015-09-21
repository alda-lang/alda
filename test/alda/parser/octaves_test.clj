(ns alda.parser.octaves-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest octave-tests
  (testing "octave change"
    (is (= '(alda.lisp/octave :up)   (test-parse :octave-up ">")))
    (is (= '(alda.lisp/octave :down) (test-parse :octave-down "<")))
    (is (= '(alda.lisp/octave 5)     (test-parse :octave-set "o5")))))

