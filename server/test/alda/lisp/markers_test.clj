(ns alda.lisp.markers-test
  (:require [clojure.test            :refer :all]
            [alda.test-helpers       :refer (get-instrument
                                             dur->ms)]
            [alda.lisp               :refer :all]
            [alda.lisp.model.records :refer (->AbsoluteOffset
                                             ->RelativeOffset)]))

(deftest marker-tests
  (testing "a marker:"
    (let [s                    (score (part "piano"))
          events-before        (:events s)
          marker-events-before (:marker-events s)
          piano                (get-instrument s "piano")
          offset-before        (:current-offset piano)
          marker-before        (:current-marker piano)
          s                    (continue s
                                 (marker "test-marker"))
          piano                (get-instrument s "piano")
          offset-after         (:current-offset piano)
          marker-after         (:current-marker piano)]
      (testing "placing a marker doesn't change the current offset"
        (is (= offset-before offset-after)))
      (testing "placing a marker doesn't change the current marker"
        (is (= marker-before marker-after)))
      (let [test-marker-offset (-> s :markers (get "test-marker"))]
        (testing "it should record the offset of the marker in :markers"
          (is (not (nil? test-marker-offset)))
          (is (number? test-marker-offset)))
        (testing "its offset should be correct"
          (is (offset= s offset-before (->AbsoluteOffset test-marker-offset)))))
      (testing "events should continue to go in :events, not :marker-events"
        (let [s                   (continue s
                                    (note (pitch :c)))
              events-after        (:events s)
              marker-events-after (:marker-events s)]
          (is (empty? marker-events-before))
          (is (empty? marker-events-after))
          (is (empty? events-before))
          (is (not (empty? events-after)))))))
  (testing "at-marker:"
    (let [s     (score
                  (part "piano"
                    (marker "test-marker")
                    (pause (duration (note-length 1) (note-length 1)))
                    (at-marker "test-marker")))
          piano (get-instrument s "piano")]
      (testing "it should change the current marker of the instrument"
        (is (= "test-marker" (:current-marker piano))))
      (let [s      (continue s
                     (note (pitch :d)))
            events (:events s)]
        (testing "events should go in :events if the marker has been placed"
          (is (= 1 (count events))))
        (testing "events should be placed starting at the marker's offset + 0 ms"
          (is (offset= s
                       (->RelativeOffset "test-marker" 0)
                       (->AbsoluteOffset (:offset (first events)))))))
      (testing "using at-marker without first placing the marker results in an
                exception"
        (is (thrown-with-msg? Exception
                              #"marker does not exist"
                              (continue s
                                (at-marker "nonexistent"))))))))

