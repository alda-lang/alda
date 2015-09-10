(ns alda.test.lisp.pitch
  (:require [clojure.test :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(defn- =% [x y]
  (< x y (+ x 0.01)))

(deftest pitch-tests
  (testing "pitch converts a note in a given octave to frequency in Hz"
    (is (== 440 ((pitch :a) 4)))
    (is (== 880 ((pitch :a) 5)))
    (is (=% 261.62 ((pitch :c) 4))))
  (testing "pitch converts a note in a given octave to a MIDI note"
    (is (= 69 ((pitch :a) 4 :default :midi true)))
    (is (= 81 ((pitch :a) 5 :default :midi true)))
    (is (= 60 ((pitch :c) 4 :default :midi true))))
  (testing "pitch converts using a well-tempered tuning"
    (is (=% 437.02 ((pitch :a) 4 :well)))
    (is (=% 874.05 ((pitch :a) 5 :well)))
    (is (=% 261.62 ((pitch :c) 4 :well))))
  (testing "pitch converts well-tempered tuning notes to MIDI also"
    (is (=% 68.88 ((pitch :a) 4 :well :midi true)))
    (is (=% 80.88 ((pitch :a) 5 :well :midi true)))
    (is (= 60.0 ((pitch :c) 4 :well :midi true))))
  (testing "flats and sharps"
    (is (>  ((pitch :c :sharp) 4) ((pitch :c) 4)))
    (is (>  ((pitch :c) 5) ((pitch :c :sharp) 4)))
    (is (<  ((pitch :b :flat) 4)  ((pitch :b) 4)))
    (is (== ((pitch :c :sharp) 4) ((pitch :d :flat) 4)))
    (is (== ((pitch :c :sharp :sharp) 4) ((pitch :d) 4)))
    (is (== ((pitch :f :flat) 4) ((pitch :e) 4)))
    (is (== ((pitch :a :flat :flat) 4) ((pitch :g) 4)))
    (is (== ((pitch :c :sharp :flat :flat :sharp) 4) ((pitch :c) 4)))))
