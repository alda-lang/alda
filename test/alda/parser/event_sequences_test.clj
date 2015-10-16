(ns alda.parser.event-sequences-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest event-sequence-tests
  (testing "event sequences"
    (is (= (test-parse :event-sequence "[ c d e f c/e/g ]")
           '(do
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/note (alda.lisp/pitch :f))
              (alda.lisp/chord
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :g))))))
    (is (= (test-parse :event-sequence "[c d [e f] g]")
           '(do
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (do
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :f)))
              (alda.lisp/note (alda.lisp/pitch :g)))))))
