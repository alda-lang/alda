(ns alda.lisp.markers-test
  (:require [clojure.test            :refer :all]
            [alda.test-helpers       :refer (get-instrument)]
            [alda.lisp               :refer :all]
            [alda.lisp.model.records :refer (->RelativeOffset)]))

(deftest marker-tests
  (testing "a marker:"
    (let [s             (score (part "piano"))
          piano         (get-instrument s "piano")
          offset-before (:current-offset piano)
          marker-before (:current-marker piano)
          s             (continue s (marker "test-marker"))
          piano         (get-instrument s "piano")
          offset-after  (:current-offset piano)
          marker-after  (:current-marker piano)]
      (testing "placing a marker doesn't change the current offset"
        (is (= offset-before offset-after)))
      (testing "placing a marker doesn't change the current marker"
        (is (= marker-before marker-after)))
      (let [test-marker (-> s :events (get "test-marker"))]
        (testing "it should create an entry for the marker in :events"
          (is (not (nil? test-marker))))
        (testing "its offset should be correct"
          (is (offset= offset-before (:offset test-marker)))))))
  (testing "at-marker:"
    (let [s     (score
                  (part "piano"
                    (marker "test-marker")
                    (pause (duration (note-length 1) (note-length 1)))
                    (at-marker "test-marker")))
          piano (get-instrument s "piano")]
      (testing "new events should be placed from the marker / offset 0"
        (is (= "test-marker" (:current-marker piano)))
        (is (zero? (count (get-in s [:events "test-marker" :events]))))
        (let [s               (continue s
                                (note (pitch :d)))
              notes-at-marker (get-in s [:events "test-marker" :events])]
          (is (= 1 (count notes-at-marker)))
          (is (offset= (->RelativeOffset "test-marker" 0)
                       (:offset (first notes-at-marker))))))
      (testing "using at-marker before marker:"
        (let [s               (continue s
                                (at-marker "test-marker-2"))
              piano           (get-instrument s "piano")
              notes-at-marker (get-in s [:events "test-marker-2" :events])]
          (testing "at-marker still works;"
            (testing "it sets :current-marker"
              (is (= "test-marker-2" (:current-marker piano))))
            (testing "new events are placed from the marker / offset 0"
              (is (zero? (count notes-at-marker)))
              (let [s               (continue s
                                      (note (pitch :e)))
                    notes-at-marker (get-in s [:events
                                               "test-marker-2"
                                               :events])]
                (is (= 1 (count notes-at-marker)))
                (is (offset= (->RelativeOffset "test-marker-2" 0)
                             (:offset (first notes-at-marker)))))))
          (testing "marker adds an offset to the marker"
            (let [s             (continue s
                                  (note (pitch :e))
                                  (at-marker "test-marker")
                                  (pause (duration (note-length 1)
                                                   (note-length 1)
                                                   (note-length 1)
                                                   (note-length 1)))
                                  (marker "test-marker-2"))
                  piano         (get-instrument s "piano")
                  marker-offset (get-in s [:events "test-marker-2" :offset])]
              (is (offset= (:current-offset piano) marker-offset)))))))))

