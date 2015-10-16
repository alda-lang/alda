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

(deftest millisecond-duration-tests
  (testing "duration in milliseconds"
    (is (= (test-parse :duration "400ms")
           '(alda.lisp/duration (alda.lisp/ms 400))))
    (is (= (test-parse :note "c450ms")
           '(alda.lisp/note (alda.lisp/pitch :c)
                            (alda.lisp/duration (alda.lisp/ms 450)))))))

(deftest second-duration-tests
  (testing "duration in seconds"
    (is (= (test-parse :duration "5s")
           '(alda.lisp/duration (alda.lisp/ms 5000))))
    (is (= (test-parse :note "c2s")
           '(alda.lisp/note (alda.lisp/pitch :c)
                            (alda.lisp/duration (alda.lisp/ms 2000)))))))

(deftest tie-and-slur-tests
  (testing "ties"
    (testing "ties"
      (is (= (test-parse :duration "1~2~4")
             '(alda.lisp/duration (alda.lisp/note-length 1)
                                  (alda.lisp/note-length 2)
                                  (alda.lisp/note-length 4))))
      (is (= (test-parse :duration "500ms~350ms")
             '(alda.lisp/duration (alda.lisp/ms 500)
                                  (alda.lisp/ms 350))))
      (is (= (test-parse :duration "5s~4~350ms")
             '(alda.lisp/duration (alda.lisp/ms 5000)
                                  (alda.lisp/note-length 4)
                                  (alda.lisp/ms 350)))))
    (testing "slurs"
      (is (= (test-parse :duration "4~")
             '(alda.lisp/duration (alda.lisp/note-length 4) :slur)))
      (is (= (test-parse :duration "420ms~")
             '(alda.lisp/duration (alda.lisp/ms 420) :slur))))))

