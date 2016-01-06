(ns alda.lisp.markers-test
  (:require [clojure.test            :refer :all]
            [clojure.pprint          :refer :all]
            [alda.lisp               :refer :all]
            [alda.lisp.score.context :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest marker-tests
  (testing "a marker:"
    (let [current-marker ($current-marker)
          moment ($current-offset)
          test-marker (marker "test-marker")]
      (testing "returns a Marker record"
        (is (= test-marker
               (alda.lisp.model.records.Marker. "test-marker"
                 (alda.lisp.model.records.AbsoluteOffset.
                   (absolute-offset moment))))))
      (testing "it should create an entry for the marker in *events*"
        (is (contains? *events* "test-marker")))
      (testing "its offset should be correct"
        (is (offset= moment (get-in *events* ["test-marker" :offset]))))
      (testing "placing a marker doesn't change the current marker"
        (is (= current-marker ($current-marker))))))
  (testing "at-marker:"
    (pause (duration (note-length 1) (note-length 1)))
    (at-marker "test-marker")
    (testing "new events should be placed from the marker / offset 0"
      (is (= "test-marker" ($current-marker)))
      (is (zero? (count (get-in *events* ["test-marker" :events]))))
      (let [first-note (first (note (pitch :d)))]
        (is (= 1 (count (get-in *events* ["test-marker" :events]))))
        (is (offset= (alda.lisp.model.records.RelativeOffset. "test-marker" 0)
                     (:offset first-note))))))
  (testing "using at-marker before marker:"
    (testing "at-marker still works;"
      (at-marker "test-marker-2")
      (testing "it sets *current-marker*"
        (is (= "test-marker-2" ($current-marker))))
      (testing "new events are placed from the marker / offset 0"
        (is (zero? (count (get-in *events* ["test-marker-2" :events]))))
        (let [first-note (first (note (pitch :e)))]
          (is (= 1 (count (get-in *events* ["test-marker-2" :events]))))
          (is (offset= (alda.lisp.model.records.RelativeOffset. "test-marker-2" 0)
                       (:offset first-note))))))
    (testing "marker adds an offset to the marker"
      (at-marker "test-marker")
      (pause (duration (note-length 1)
                       (note-length 1)
                       (note-length 1)
                       (note-length 1)))
      (marker "test-marker-2")
      (is (offset= ($current-offset)
                   (get-in *events* ["test-marker-2" :offset]))))))

