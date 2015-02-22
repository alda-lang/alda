(ns alda.lisp-score-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]))

#_(deftest part-tests
  (testing "a part:"
    (part {:names ["piano" "trumpet"] :nickname "trumpiano"}
      (testing "starts at offset 0"
        (is (zero? (:current-offset (*instruments* "piano"))))
        (is (zero? (:current-offset (*instruments* "trumpet")))))
      (testing "starts at the :start marker"
        (is (= :start (:current-marker (*instruments* "piano"))))
        (is (= :start (:current-marker (*instruments* "trumpet")))))
      (testing "has the instruments that it has"
        (is (= #{"piano" "trumpet"} *current-instruments*)))
      (testing "sets a nickname if applicable"
        (is (contains? *nicknames* "trumpiano"))
        (is (= #{"piano" "trumpet"} (*nicknames* "trumpiano"))))
      (note (pitch :d) (duration (note-length 2 {:dots 1}))))
    (def piano-offset (-> (*instruments* "piano") :current-offset))
    (def trumpet-offset (-> (*instruments* "trumpet") :current-offset))
    (testing "instruments from a group can be separated at will"
      (part {:names ["piano"]}
        (is (= *current-instruments* #{"piano"}))
        (is (= piano-offset (-> (*instruments* "piano") :current-offset)))
        (chord (note (pitch :a))
               (note (pitch :c :sharp))
               (note (pitch :e))))
      (alter-var-root #'piano-offset
                      (constantly (-> (*instruments* "piano") :current-offset)))
      (part {:names ["trumpet"]}
        (is (= *current-instruments* #{"trumpet"}))
        (is (= trumpet-offset (-> (*instruments* "trumpet") :current-offset)))
        (note (pitch :d))
        (note (pitch :e))
        (note (pitch :f :sharp)))
      (alter-var-root #'trumpet-offset
                      (constantly (-> (*instruments* "trumpet") :current-offset)))
      (is (= piano-offset (-> (*instruments* "piano") :current-offset))))
    (testing "referencing a nickname"
      (part {:names ["trumpiano"]}
        (is (= *current-instruments* #{"piano" "trumpet"}))))))

#_(deftest lisp-test
  (testing "instrument part consolidation"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/examples/awobmolg.alda"))]
        (pprint (eval result))))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/examples/debussy_quartet.alda"))]
        (pprint (eval result))))))
