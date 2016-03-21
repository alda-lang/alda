(ns alda.lisp.chords-test
  (:require [clojure.test      :refer :all]
            [clojure.pprint    :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]))

(deftest chord-tests
  (testing "a chord event:"
    (let [s      (score (part "piano"))
          piano  (get-instrument s "piano")
          start  (:current-offset piano)
          s      (continue s
                   (chord (note (pitch :c) (duration (note-length 1)))
                          (note (pitch :e) (duration (note-length 4)))
                          (pause (duration (note-length 8)))
                          (note (pitch :g) (duration (note-length 2)))))
          piano  (get-instrument s "piano")
          end    (:current-offset piano)
          prev   (:last-offset piano)
          events (get-in s [:events (:current-marker piano) :events])]
      (testing "the notes should all start at the same time"
        (is (every? #(= start (:offset %)) events)))
      (testing ":current-offset should be bumped forward by the shortest
        note/rest duration"
        (is (= end (offset+ start 250.0))))
      (testing ":last-offset should be updated correctly"
        (is (offset= s prev start))))))

