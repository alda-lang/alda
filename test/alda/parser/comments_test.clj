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
                                  e")))))

