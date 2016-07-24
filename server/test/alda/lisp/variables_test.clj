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
          (is (contains? (-> s :env) :foo))
          (is (not (nil? (-> s :env :foo))))
          (is (= 3 (count (-> s :env :foo))))))))
  (testing "getting a variable that has NOT been set"
    (testing "should result in an exception"
      (is (thrown-with-msg? Exception
                            #"Undefined variable: lolbadvariable"
                            (score
                              (get-variable :lolbadvariable)))))
    (testing "should result in an exception"
      (is (thrown-with-msg? Exception
                            #"Undefined variable: lolbadvariable"
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
                (note (pitch :f))))]
      (testing "should result in a nested variable definition"
        (is (contains? (-> s :env) :foo))
        (is (not (nil? (-> s :env :foo))))
        ; one event seq, one F note
        (is (= 2 (count (-> s :env :foo))))
        ; the event seq should have 3 events (the notes C, D, and E)
        (is (= 3 (count (-> s :env :foo first)))))
      (testing "should correctly use the old value in the new definition"
        (let [s (continue s
                  (part "piano"
                    (get-variable :foo)))]
          (is (= 4 (count (:events s))))))))
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
                                (get-variable :bar))))))))

