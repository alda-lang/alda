(ns alda.parser.event-sequences-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-to-lisp-with-context)]))

(deftest event-sequence-tests
  (testing "event sequences"
    (is (= (parse-to-lisp-with-context :music-data "[ c d e f c/e/g ]")
           '([(alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/note (alda.lisp/pitch :f))
              (alda.lisp/chord
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :g)))])))
    (is (= (parse-to-lisp-with-context :music-data "[c d [e f] g]")
           '([(alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              [(alda.lisp/note (alda.lisp/pitch :e))
               (alda.lisp/note (alda.lisp/pitch :f))]
              (alda.lisp/note (alda.lisp/pitch :g))])))))
