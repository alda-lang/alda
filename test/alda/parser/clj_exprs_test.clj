(ns alda.parser.clj-exprs-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest attribute-tests
  (testing "volume change"
    (is (= (test-parse :clj-expr "(volume 50)") '(alda.lisp/volume 50))))
  (testing "tempo change"
    (is (= (test-parse :clj-expr "(tempo 100)") '(alda.lisp/tempo 100))))
  (testing "quantization change"
    (is (= (test-parse :clj-expr "(quant 75)") '(alda.lisp/quant 75))))
  (testing "panning change"
    (is (= (test-parse :clj-expr "(panning 0)") '(alda.lisp/panning 0)))))

(deftest multiple-attribute-change-tests
  (testing "attribute changes"
    (is (= (test-parse :clj-expr "(vol 50, tempo 100)")
           '(do (alda.lisp/vol 50) (alda.lisp/tempo 100))))
    (is (= (test-parse :clj-expr "(quant! 50; tempo 90)")
           '(do (alda.lisp/quant! 50) (alda.lisp/tempo 90)))))
  (testing "global attribute changes"
    (is (= (test-parse :clj-expr "(tempo! 126)")
           '(alda.lisp/tempo! 126)))
    (is (= (test-parse :clj-expr "(tempo! 130, quant! 80)")
           '(do (alda.lisp/tempo! 130) (alda.lisp/quant! 80))))))

(deftest more-comma-and-semicolon-tests
  (testing "commas/semicolons can exist in strings"
    (is (= (test-parse :clj-expr "(println \"hi; hi, hi\")")
           '(clojure.core/println "hi; hi, hi"))))
  (testing "comma/semicolon character literals are OK too"
    (is (= (test-parse :clj-expr "(println \\, \\;)")
           '(clojure.core/println \, \;)))))

