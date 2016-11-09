(ns alda.parser.comments-test
  (:require [clojure.test :refer :all]
            [alda.parser  :refer (parse-input)]))

(def expected
  '(alda.lisp/score
     (alda.lisp/part {:names ["piano"]}
       (alda.lisp/note (alda.lisp/pitch :c))
       (alda.lisp/note (alda.lisp/pitch :e)))))

(deftest short-comment-tests
  (testing "a short comment"
    (is (= expected (parse-input "piano: c
                                  # d
                                  e")))
    (testing "at the end of a line"
      (is (= expected (parse-input "piano: c # d
                                    e")))
      (testing "without a leading space"
        (is (= expected (parse-input "piano: c#d\ne")))))))

