(ns alda.lisp-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]
            [instaparse.core :as insta]))

(deftest duration-tests
  (testing "note-length converts note length to number of beats"
    (is (== 1 (note-length 4)))
    (is (== 1.5 (note-length 4 {:dots 1})))
    (is (== 4 (note-length 1)))
    (is (== 6 (note-length 1 {:dots 1})))
    (is (== 7 (note-length 1 {:dots 2}))))
;; alda.lisp has moved to using dynamic vars.
;; TODO: rewrite using *tempo* instead of providing it as an arg
    (testing "duration converts beats to ms, given tempo"
    (is (= {:duration 1000 :slurred true}
            ((duration (note-length 4) :slur) 60)))
    (is (= {:duration 500 :slurred false}
            ((duration (note-length 4)) 120)))
    (is (= {:duration 750 :slurred false}
            ((duration (note-length 4 {:dots 1})) 120)))
    (is (= {:duration 7500 :slurred true}
           ((duration (note-length 2)
                      (note-length 2)
                      (note-length 2 {:dots 2}) :slur) 60)))))

(deftest lisp-test
  (testing "instrument part consolidation"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/examples/awobmolg.alda"))]
        (pprint (eval result))))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/examples/debussy_quartet.alda"))]
        (pprint (eval result))))))
