(ns alda.parser.repeats-test
  (:require [clojure.test     :refer :all]
            [alda.parser-util :refer (parse-to-lisp-with-context)]))

(deftest repeat-tests
  (testing "repeated events"
    (is (= (parse-to-lisp-with-context :music-data "[c d e] *4")
           '((alda.lisp/times 4
               [(alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/note (alda.lisp/pitch :d))
                (alda.lisp/note (alda.lisp/pitch :e))]))))
    (is (= (parse-to-lisp-with-context :music-data "[ c > ]*5")
           '((alda.lisp/times 5
               [(alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/octave :up)]))))
    (is (= (parse-to-lisp-with-context :music-data "[ c > ] * 5")
           '((alda.lisp/times 5
               [(alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/octave :up)]))))
    (is (= (parse-to-lisp-with-context :music-data "c8*7")
           '((alda.lisp/times 7
               (alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/note-length 8)))))))
    (is (= (parse-to-lisp-with-context :music-data "c8 *7")
           '((alda.lisp/times 7
               (alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/note-length 8)))))))
    (is (= (parse-to-lisp-with-context :music-data "c8 * 7")
           '((alda.lisp/times 7
               (alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/note-length 8)))))))))

