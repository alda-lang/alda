(ns alda.lisp.variables-test
  (:require [clojure.test             :refer :all]
            [alda.test-helpers        :refer (get-instrument)]
            [alda.lisp                :refer :all]))

(deftest variable-tests
  (testing "setting a variable"
    (let [s1 (score
               (set-variable :foo
                 (note (pitch :c))
                 (note (pitch :d))
                 (note (pitch :e))))
          s2 (score
               (part "piano"
                 (set-variable :foo
                   (note (pitch :c))
                   (note (pitch :d))
                   (note (pitch :e)))))]
      (testing "should update the score correctly"
        (doseq [s [s1 s2]]
          (is (contains? (-> s :variables) :foo))
          (is (contains? (-> s :variables :foo) :env))
          (is (contains? (-> s :variables :foo) :events))
          (is (= {} (-> s :variables :foo :env)))
          (is (= 3 (count (-> s :variables :foo :events))))))))
  (testing "getting a variable that has NOT been set"
    (testing "should result in an exception"
      (is (thrown-with-msg? Exception
                            #"Undefined variable"
                            (score
                              (get-variable :lolbadvariable)))))
    (testing "should result in an exception"
      (is (thrown-with-msg? Exception
                            #"Undefined variable"
                            (score
                              (part "piano"
                                (get-variable :lolbadvariable)))))))
  (testing "getting a variable that HAS been set"
    (let [s (score
              (set-variable :foo
                (note (pitch :c))
                (note (pitch :d))
                (note (pitch :e)))
              (part "piano"
                (get-variable :foo)))]
      (testing "should add events to the score"
        (is (= 3 (count (:events s)))))))
  (testing "reusing a variable in its own redefinition"
    (let [s (score
              (set-variable :foo
                (note (pitch :c))
                (note (pitch :d))
                (note (pitch :e)))
              (set-variable :foo
                (get-variable :foo)
                (note (pitch :f))
                (note (pitch :g))))]
      (testing "should result in a nested variable definition"
        (is (not (nil? (-> s :variables :foo :env))))
        (is (contains? (-> s :variables :foo :env) :foo))
        (is (= {} (-> s :variables :foo :env :foo :env)))
        (is (not (nil? (-> s :variables :foo :env :foo :events)))))
      (testing "should correctly use the old value in the new definition"
        (let [s (continue s
                  (part "piano"
                    (get-variable :foo)))]
          (is (= 5 (count (:events s))))))))
  (testing "defining a variable that uses another variable"
    (let [s (score
              (set-variable :foo
                (note (pitch :c))
                (note (pitch :d))
                (note (pitch :e)))
              (set-variable :bar
                (get-variable :foo)
                (note (pitch :f))
                (note (pitch :g))))]
      (testing "should result in the latter variable including the events of the former"
        (let [s (continue s
                  (part "piano"
                    (get-variable :bar)))]
          (is (= 5 (count (:events s))))))
      (testing "and then redefining the first variable"
        (let [s (continue s
                  (set-variable :foo
                    (note (pitch :c))))]
          (testing "should not change the value of the second variable"
            (let [s (continue s
                      (part "piano"
                        (get-variable :bar)))]
              (is (= 5 (count (:events s))))))))))
  (testing "defining a variable that uses an undefined variable"
    (testing "should throw an undefined variable exception"
      (is (thrown-with-msg? Exception
                            #"Undefined variable: bar"
                            (score
                              (set-variable :foo
                                (get-variable :bar))
                              (part "piano"
                                (get-variable :foo))))))
    (testing "should throw an undefined variable exception every time"
      (is (thrown-with-msg? Exception
                            #"Undefined variable: bar"
                            (score
                              (set-variable :foo
                                (get-variable :bar))
                              (part "piano"
                                (get-variable :foo)
                                (set-variable :baz
                                  (get-variable :quux))
                                (get-variable :foo)))))
      (is (thrown-with-msg? Exception
                            #"Undefined variable: quux"
                            (score
                              (set-variable :foo
                                (get-variable :bar))
                              (set-variable :baz
                                (get-variable :quux))
                              (part "piano"
                                (get-variable :baz))))))))
