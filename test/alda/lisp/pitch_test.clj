(ns alda.lisp.pitch-test
  (:require [clojure.test :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest pitch-tests
  (testing "pitch converts a note in a given octave to frequency in Hz"
    (is (== 440 ((pitch :a) 4)))
    (is (== 880 ((pitch :a) 5)))
    (is (< 261 ((pitch :c) 4) 262)))
  (testing "flats and sharps"
    (is (>  ((pitch :c :sharp) 4) ((pitch :c) 4)))
    (is (>  ((pitch :c) 5) ((pitch :c :sharp) 4)))
    (is (<  ((pitch :b :flat) 4)  ((pitch :b) 4)))
    (is (== ((pitch :c :sharp) 4) ((pitch :d :flat) 4)))
    (is (== ((pitch :c :sharp :sharp) 4) ((pitch :d) 4)))
    (is (== ((pitch :f :flat) 4) ((pitch :e) 4)))
    (is (== ((pitch :a :flat :flat) 4) ((pitch :g) 4)))
    (is (== ((pitch :c :sharp :flat :flat :sharp) 4) ((pitch :c) 4)))))
