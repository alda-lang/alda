(ns alda.parser.duration-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-with-context)]))

(deftest duration-tests
  (testing "duration"
    (is (= (parse-with-context :music-data "c2")
           '((alda.lisp/note
               (alda.lisp/pitch :c)
               (alda.lisp/duration (alda.lisp/note-length 2))))))))

(deftest dot-tests
  (testing "dots"
    (is (= (parse-with-context :music-data "c2..")
           '((alda.lisp/note
               (alda.lisp/pitch :c)
               (alda.lisp/duration (alda.lisp/note-length 2 {:dots 2}))))))))

(deftest millisecond-duration-tests
  (testing "duration in milliseconds"
    (is (= (parse-with-context :music-data "c450ms")
           '((alda.lisp/note
               (alda.lisp/pitch :c)
               (alda.lisp/duration (alda.lisp/ms 450))))))))

(deftest second-duration-tests
  (testing "duration in seconds"
    (is (= (parse-with-context :music-data "c2s")
           '((alda.lisp/note
               (alda.lisp/pitch :c)
               (alda.lisp/duration (alda.lisp/ms 2000))))))))

(deftest tie-and-slur-tests
  (testing "ties"
    (testing "ties"
      (is (= (parse-with-context :music-data "c1~2~4")
             '((alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/note-length 1)
                                     (alda.lisp/note-length 2)
                                     (alda.lisp/note-length 4))))))
      (is (= (parse-with-context :music-data "c500ms~350ms")
             '((alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/ms 500)
                                     (alda.lisp/ms 350))))))
      (is (= (parse-with-context :music-data "c5s~4~350ms")
             '((alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/ms 5000)
                                     (alda.lisp/note-length 4)
                                     (alda.lisp/ms 350)))))))
    (testing "slurs"
      (is (= (parse-with-context :music-data "c4~")
             '((alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/note-length 4))
                 :slur))))
      (is (= (parse-with-context :music-data "c420ms~")
             '((alda.lisp/note
                 (alda.lisp/pitch :c)
                 (alda.lisp/duration (alda.lisp/ms 420))
                 :slur)))))))

