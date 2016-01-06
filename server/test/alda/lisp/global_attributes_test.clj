(ns alda.lisp.global-attributes-test
  (:require [clojure.test            :refer :all]
            [clojure.pprint          :refer :all]
            [clojure.string          :as    str]
            [alda.lisp               :refer :all]
            [alda.lisp.score.context :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest global-attribute-tests
  (let [piano (first (for [[id instrument] *instruments*
                           :when (str/starts-with? id "piano-")]
                       id))]
    (testing "a global tempo change:"
      (set-current-offset piano (alda.lisp.model.records.AbsoluteOffset. 0))
      (at-marker :start)
      (tempo 120)
      (pause (duration (note-length 1)))
      (global-attribute :tempo 60) ; 2000 ms from :start
      (testing "it should change the tempo"
        (is (= ($tempo) 60)))
      (pause (duration (note-length 1) (note-length 1) (note-length 1)))
      (marker "test-marker-3") ; for later test
      (testing "when another part starts,"
        (set-current-offset piano (alda.lisp.model.records.AbsoluteOffset. 0))
        (tempo 120)
        (testing "the tempo should change once it encounters the global attribute"
          (is (= ($tempo) 120)) ; not yet...
          (pause (duration (note-length 2 {:dots 1})))
          (is (= ($tempo) 120)) ; not yet...
          (pause)
          (is (= ($tempo) 60)))) ; now!
      (testing "it should use absolute offset, not relative to marker"
        (at-marker "test-marker-3")
        (is (offset= ($current-offset)
                     (alda.lisp.model.records.RelativeOffset. "test-marker-3" 0)))
        (is (= ($current-marker) "test-marker-3"))
        (tempo 120)
        (is (= ($tempo) 120))
        (pause (duration (note-length 2 {:dots 1})))
        (is (= ($tempo) 120))
        (pause)
        (is (= ($tempo) 120)))))) ; tempo should still be 120,
                                  ; despite having passed 2000 ms

(alter-var-root #'*global-attributes* (constantly {}))
