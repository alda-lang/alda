(ns alda.parser.duration-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest duration-tests
  (testing "duration"
    (is (= (test-parse :duration "2")
           '(alda.lisp/duration (alda.lisp/note-length 2))))))

(deftest dot-tests
  (testing "dots"
    (is (= (test-parse :duration "2..")
           '(alda.lisp/duration (alda.lisp/note-length 2 {:dots 2}))))))

(deftest tie-and-slur-tests
  (testing "ties"
    (testing "ties"
      (is (= (test-parse :duration "1~2~4")
             '(alda.lisp/duration (alda.lisp/note-length 1)
                                  (alda.lisp/note-length 2)
                                  (alda.lisp/note-length 4)))))
    (testing "slurs"
      (is (= (test-parse :duration "4~")
             '(alda.lisp/duration (alda.lisp/note-length 4) :slur))))))

