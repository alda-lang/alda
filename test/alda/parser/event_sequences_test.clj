(ns alda.parser.event-sequences-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-with-context)]))

(deftest event-sequence-tests
  (testing "event sequences"
    (is (= (parse-with-context :music-data "[ c d e f c/e/g ]")
           '((do
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/note (alda.lisp/pitch :f))
              (alda.lisp/chord
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :g)))))))
    (is (= (parse-with-context :music-data "[c d [e f] g]")
           '((do
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (do
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :f)))
              (alda.lisp/note (alda.lisp/pitch :g))))))))
