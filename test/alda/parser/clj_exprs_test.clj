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
  (testing "commas inside [brackets] and {braces} won't break things"
    (is (= (test-parse :clj-expr "(prn [1,2,3])")
           '(clojure.core/prn [1 2 3])))
    (is (= (test-parse :clj-expr "(prn {:a 1, :b 2})")
           '(clojure.core/prn {:a 1 :b 2}))))
  (testing "comma/semicolon character literals are OK too"
    (is (= (test-parse :clj-expr "(println \\, \\;)")
           '(clojure.core/println \, \;)))))

(deftest paren-tests
  (testing "parens inside of a string are NOT a clj-expr"
    (is (= (test-parse :clj-expr "(prn \"a string (with parens)\")")
           '(clojure.core/prn "a string (with parens)")))
    (is (= (test-parse :clj-expr "(prn \"a string with just a closing paren)\")")
           '(clojure.core/prn "a string with just a closing paren)"))))
  (testing "paren character literals don't break things"
    (is (= (test-parse :clj-expr "(prn \\()")
           '(clojure.core/prn \()))
    (is (= (test-parse :clj-expr "(prn \\))")
           '(clojure.core/prn \))))
    (is (= (test-parse :clj-expr "(prn \\( (+ 1 1) \\))")
           '(clojure.core/prn \( (clojure.core/+ 1 1) \))))))

(deftest vector-tests
  (testing "vectors are a thing"
    (is (= (test-parse :clj-expr "(prn [1 2 3 \\a :b \"c\"])")
           '(clojure.core/prn [1 2 3 \a :b "c"]))))
  (testing "vectors can have commas in them"
    (is (= (test-parse :clj-expr "(prn [1, 2, 3])")
           '(clojure.core/prn [1 2 3])))))

(deftest map-tests
  (testing "maps are a thing"
    (is (= (test-parse :clj-expr "(prn {:a 1 :b 2 :c 3})")
           '(clojure.core/prn {:a 1 :b 2 :c 3}))))
  (testing "maps can have commas in them"
    (is (= (test-parse :clj-expr "(prn {:a 1, :b 2, :c 3})")
           '(clojure.core/prn {:a 1 :b 2 :c 3})))))

(deftest set-tests
  (testing "sets are a thing"
    (is (= (test-parse :clj-expr "(prn #{1 2 3})")
           '(clojure.core/prn #{1 2 3}))))
  (testing "sets can have commas in them"
    (is (= (test-parse :clj-expr "(prn #{1, 2, 3})")
           '(clojure.core/prn #{1 2 3})))))

(deftest nesting-things
  (testing "things can be nested and it won't break shit"
    (is (= (test-parse :clj-expr "(prn [1 2 [3 4] 5])")
           '(clojure.core/prn [1 2 [3 4] 5])))
    (is (= (test-parse :clj-expr "(prn #{1 2 #{3 4} 5})")
           '(clojure.core/prn #{1 2 #{3 4} 5})))
    (is (= (test-parse :clj-expr "(prn (+ 1 [2 {3 #{4 5}}]))")
           '(clojure.core/prn (clojure.core/+ 1 [2 {3 #{4 5}}]))))))

(deftest quirky-tests
  (testing "parameter order in subsequent expressions gets weird if there are data structures"
    (is (= (test-parse :clj-expr "(prn [1, 2], prn [3, 4])")
           '(do (clojure.core/prn [1 2])
                (clojure.core/prn [3 4])))))
  (testing "commas / semis should not get expanded in outer expressions"
    (is (= (test-parse :clj-expr "(prn [1 2], prn (+ 1, 2))")
           '(do (clojure.core/prn [1 2])
                (clojure.core/prn
                 (clojure.core/+ 1 2)))))))
