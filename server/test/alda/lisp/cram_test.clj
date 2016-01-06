(ns alda.lisp.cram-test
  (:require [clojure.test            :refer :all]
            [alda.lisp               :refer :all]
            [alda.lisp.score.context :refer :all]
            [alda.util               :refer (=%)]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest cram-tests-1
  (testing "a cram event with equally distributed notes:"
    (let [start  ($current-offset)
          _      (cram (note (pitch :c) :slur)
                       (note (pitch :d) :slur)
                       (note (pitch :e) :slur)
                       ; a half note at 120 bpm = 1000 ms
                       (duration (note-length 2)))
          events (get-in *events* [($current-marker) :events])]
      (testing "the first note should be placed at the current offset"
        (let [earliest-note (apply min-key #(-> % :offset :offset) events)
              offset        (:offset earliest-note)]
          (is (offset= start offset))))
      (testing "should bump :current-offset forward by its duration"
        (is (offset= (offset+ start 1000) ($current-offset))))
      (testing "the notes in a cram should be divided evenly across its duration"
        (every? (fn [{:keys [duration]}] (=% duration 166.6666)) events)))))

(deftest cram-tests-2
  (testing "a cram event with no duration provided:"
    (set-duration 4) ; a whole note at 120 bpm = 2000 ms
    (let [start  ($current-offset)
          _      (cram (note (pitch :c) :slur)
                       (note (pitch :g) :slur))
          events (get-in *events* [($current-marker) :events])]
      (testing "should use the instrument's duration attribute value"
        (is (offset= (offset+ start 2000) ($current-offset)))
        (is (every? (fn [{:keys [duration]}] (=% duration 1000.0)) events))))))

(deftest cram-tests-3
  (testing "a cram event with a variety of note lengths:"
    (cram (note (pitch :c) :slur) ; 250 ms
          (note (pitch :d) (duration (note-length 2)) :slur) ; 500 ms
          (note (pitch :e) (duration (note-length 4)) :slur) ; 250 ms
          (duration (note-length 2))) ; total duration = 1000 ms
    (let [events  (get-in *events* [($current-marker) :events])
          offsets (map :duration events)]
      (testing "notes in a cram scale in proportion to one another"
        (is (= [250.0 500.0 250.0] offsets))))))

(deftest cram-tests-4
  (testing "nested cram events:"
    (let [start   ($current-offset)
          _       (cram (note (pitch :c) :slur) ; 500 ms
                       (cram (note (pitch :e) :slur)  ; 250 ms
                             (note (pitch :g) :slur)) ; 250 ms
                       (duration (note-length 2))) ; total duration = 1000 ms
          events  (get-in *events* [($current-marker) :events])
          offsets (map :duration events)]
      (testing "offset is bumped forward by the duration of the outermost cram"
        (is (offset= (offset+ start 1000) ($current-offset))))
      (testing "note durations are divided up correctly"
        (is (= [500.0 250.0 250.0] offsets))))))
