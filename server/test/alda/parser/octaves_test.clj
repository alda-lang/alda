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
           (parse-to-lisp-with-context :music-data "o5")))))

