(ns alda.parser.repeats-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest repeat-tests
  (testing "repeated events"
    (is (= (test-parse :repeat "[c d e] *4")
           '(alda.lisp/times 4
              (do
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/note (alda.lisp/pitch :d))
                (alda.lisp/note (alda.lisp/pitch :e))))))
    (is (= (test-parse :repeat "[ c > ]*5")
           '(alda.lisp/times 5
              (do
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/octave :up)))))
    (is (= (test-parse :repeat "c8*7")
           '(alda.lisp/times 7
              (alda.lisp/note
                (alda.lisp/pitch :c)
                (alda.lisp/duration (alda.lisp/note-length 8))))))))
