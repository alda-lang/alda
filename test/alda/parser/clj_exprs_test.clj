(ns alda.parser.clj-exprs-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest attribute-tests
  (testing "volume change"
    (is (= (test-parse :clj-expr "(volume 50)") '(volume 50))))
  (testing "tempo change"
    (is (= (test-parse :clj-expr "(tempo 100)") '(tempo 100))))
  (testing "quantization change"
    (is (= (test-parse :clj-expr "(quant 75)") '(quant 75))))
  (testing "panning change"
    (is (= (test-parse :clj-expr "(panning 0)") '(panning 0)))))

(deftest multiple-attribute-change-tests
  (testing "attribute changes"
    (is (= (test-parse :clj-expr "(do (vol 50) (tempo 100))")
           '(do (vol 50) (tempo 100))))
    (is (= (test-parse :clj-expr "(do (quant! 50) (tempo 90))")
           '(do (quant! 50) (tempo 90)))))
  (testing "global attribute changes"
    (is (= (test-parse :clj-expr "(tempo! 126)")
           '(tempo! 126)))
    (is (= (test-parse :clj-expr "(do (tempo! 130) (quant! 80))")
           '(do (tempo! 130) (quant! 80))))))

(deftest comma-and-semicolon-tests
  (testing "commas/semicolons can exist in strings"
    (is (= (test-parse :clj-expr "(println \"hi; hi, hi\")")
           '(println "hi; hi, hi"))))
  (testing "commas inside [brackets] and {braces} won't break things"
    (is (= (test-parse :clj-expr "(prn [1,2,3])")
           '(prn [1 2 3])))
    (is (= (test-parse :clj-expr "(prn {:a 1, :b 2})")
           '(prn {:a 1 :b 2}))))
  (testing "comma/semicolon character literals are OK too"
    (is (= (test-parse :clj-expr "(println \\, \\;)")
           '(println \, \;)))))

(deftest paren-tests
  (testing "parens inside of a string are NOT a clj-expr"
    (is (= (test-parse :clj-expr "(prn \"a string (with parens)\")")
           '(prn "a string (with parens)")))
    (is (= (test-parse :clj-expr "(prn \"a string with just a closing paren)\")")
           '(prn "a string with just a closing paren)"))))
  (testing "paren character literals don't break things"
    (is (= (test-parse :clj-expr "(prn \\()")
           '(prn \()))
    (is (= (test-parse :clj-expr "(prn \\))")
           '(prn \))))
    (is (= (test-parse :clj-expr "(prn \\( (+ 1 1) \\))")
           '(prn \( (+ 1 1) \))))))

(deftest vector-tests
  (testing "vectors are a thing"
    (is (= (test-parse :clj-expr "(prn [1 2 3 \\a :b \"c\"])")
           '(prn [1 2 3 \a :b "c"]))))
  (testing "vectors can have commas in them"
    (is (= (test-parse :clj-expr "(prn [1, 2, 3])")
           '(prn [1 2 3])))))

(deftest map-tests
  (testing "maps are a thing"
    (is (= (test-parse :clj-expr "(prn {:a 1 :b 2 :c 3})")
           '(prn {:a 1 :b 2 :c 3}))))
  (testing "maps can have commas in them"
    (is (= (test-parse :clj-expr "(prn {:a 1, :b 2, :c 3})")
           '(prn {:a 1 :b 2 :c 3})))))

(deftest set-tests
  (testing "sets are a thing"
    (is (= (test-parse :clj-expr "(prn #{1 2 3})")
           '(prn #{1 2 3}))))
  (testing "sets can have commas in them"
    (is (= (test-parse :clj-expr "(prn #{1, 2, 3})")
           '(prn #{1 2 3})))))

(deftest nesting-things
  (testing "things can be nested and it won't break shit"
    (is (= (test-parse :clj-expr "(prn [1 2 [3 4] 5])")
           '(prn [1 2 [3 4] 5])))
    (is (= (test-parse :clj-expr "(prn #{1 2 #{3 4} 5})")
           '(prn #{1 2 #{3 4} 5})))
    (is (= (test-parse :clj-expr "(prn (+ 1 [2 {3 #{4 5}}]))")
           '(prn (+ 1 [2 {3 #{4 5}}]))))))

