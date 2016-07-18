(ns alda.parser.octaves-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-to-lisp-with-context)]))

(deftest octave-tests
  (testing "octave change"
    (is (= '((alda.lisp/octave :up))
           (parse-to-lisp-with-context :music-data ">")))
    (is (= '((alda.lisp/octave :down))
           (parse-to-lisp-with-context :music-data "<")))
    (is (= '((alda.lisp/octave 5))
           (parse-to-lisp-with-context :music-data "o5"))))
  (testing "multiple octave changes back to back without spaces"
    (is (= '((alda.lisp/octave :up)
             (alda.lisp/octave :up)
             (alda.lisp/octave :up))
           (parse-to-lisp-with-context :music-data ">>>")))
    (is (= '((alda.lisp/octave :down)
             (alda.lisp/octave :down)
             (alda.lisp/octave :down))
           (parse-to-lisp-with-context :music-data "<<<")))
    (is (= '((alda.lisp/octave :up)
             (alda.lisp/octave :down)
             (alda.lisp/octave :up))
           (parse-to-lisp-with-context :music-data "><>"))))
  (testing "octave changes back to back with notes"
    (is (= '((alda.lisp/octave :up)
             (alda.lisp/note (alda.lisp/pitch :c)))
           (parse-to-lisp-with-context :music-data ">c")))
    (is (= '((alda.lisp/note (alda.lisp/pitch :c))
             (alda.lisp/octave :down))
           (parse-to-lisp-with-context :music-data "c<")))
    (is (= '((alda.lisp/octave :up)
             (alda.lisp/note (alda.lisp/pitch :c))
             (alda.lisp/octave :down))
           (parse-to-lisp-with-context :music-data ">c<")))))

