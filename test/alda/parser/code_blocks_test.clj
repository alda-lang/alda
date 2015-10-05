(ns alda.parser.code-blocks-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest code-block-tests
  (testing "anything between [square brackets] is a code block"
    (is (= (test-parse :code-block "[ c d e f c/e/g ]")
           '(alda.lisp/code-block " c d e f c/e/g ")))
    (is (= (test-parse :code-block "[aoeuaoeuaoeuaoe]")
           '(alda.lisp/code-block "aoeuaoeuaoeuaoe")))
    (is (= (test-parse :code-block "[c d [e f] g]")
           '(alda.lisp/code-block "c d [e f] g")))))
