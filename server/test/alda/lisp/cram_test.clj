(ns alda.lisp.cram-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]
            [alda.util         :refer (=%)]))

(deftest cram-tests-1
  (testing "a cram event with equally distributed notes:"
    (let [s      (score (part "piano"))
          piano  (get-instrument s "piano")
          start  (:current-offset piano)
          s      (continue s
                   (cram (note (pitch :c) :slur)
                         (note (pitch :d) :slur)
                         (note (pitch :e) :slur)
                         ; a half note at 120 bpm = 1000 ms
                         (duration (note-length 2))))
          piano  (get-instrument s "piano")
          end    (:current-offset piano)
          events (:events s)]
      (testing "the first note should be placed at the current offset"
        (let [earliest-note (apply min-key :offset events)
              offset        (:offset earliest-note)]
          (is (offset= s start offset))))
      (testing "should bump :current-offset forward by its duration"
        (is (offset= s (offset+ start 1000) end)))
      (testing "the notes in a cram should be divided evenly across its duration"
        (is (every? (fn [{:keys [duration]}] (=% duration 333.3333)) events))))))

(deftest cram-tests-2
  (testing "a cram event with no duration provided:"
    (let [s      (score (part "piano"))
          piano  (get-instrument s "piano")
          start  (:current-offset piano)
          s      (continue s
                   (set-duration 4) ; a whole note at 120 bpm = 2000 ms
                   (cram (note (pitch :c) :slur)
                         (note (pitch :g) :slur)))
          piano  (get-instrument s "piano")
          end    (:current-offset piano)
          events (:events s)]
      (testing "should use the instrument's duration attribute value"
        (is (offset= s (offset+ start 2000) end))
        (is (every? (fn [{:keys [duration]}] (=% duration 1000.0)) events))))))

(deftest cram-tests-3
  (testing "a cram event with a variety of note lengths:"
    (let [s       (score (part "piano"))
          piano   (get-instrument s "piano")
          start   (:current-offset piano)
          s       (continue s
                    (cram (note (pitch :c) :slur) ; 250 ms
                          (note (pitch :d)
                                (duration (note-length 2)) :slur) ; 500 ms
                          (note (pitch :e)
                                (duration (note-length 4)) :slur) ; 250 ms
                          (duration (note-length 2)))) ; total duration = 1000 ms
          piano   (get-instrument s "piano")
          end     (:current-offset piano)
          events  (:events s)
          offsets (->> events
                       (sort-by :offset)
                       (map :duration))]
      (testing "notes in a cram scale in proportion to one another"
        (is (= [250.0 500.0 250.0] offsets))))))

(deftest cram-tests-4
  (testing "nested cram events:"
    (let [s       (score (part "piano"))
          piano   (get-instrument s "piano")
          start   (:current-offset piano)
          s       (continue s
                    (cram (note (pitch :c) :slur) ; 500 ms
                          (cram (note (pitch :e) :slur)  ; 250 ms
                                (note (pitch :g) :slur)) ; 250 ms
                          (duration (note-length 2)))) ; total duration = 1000 ms
          piano   (get-instrument s "piano")
          end     (:current-offset piano)
          events  (:events s)
          offsets (->> events
                       (sort-by :offset)
                       (map :duration))]
      (testing "offset is bumped forward by the duration of the outermost cram"
        (is (offset= s (offset+ start 1000) end)))
      (testing "note durations are divided up correctly"
        (is (= [500.0 250.0 250.0] offsets)))))
  (testing "repeated nested cram events:"
    (let [s       (score (part "piano"))
          piano   (get-instrument s "piano")
          start   (:current-offset piano)
          s       (continue s
                    (times 2
                      (cram (note (pitch :c) :slur) ; 500 ms
                            (cram (note (pitch :e) :slur)  ; 250 ms
                                  (note (pitch :g) :slur)) ; 250 ms
                            (duration (note-length 2))))) ; total dur = 1000 ms
          piano   (get-instrument s "piano")
          end     (:current-offset piano)
          events  (:events s)
          offsets (->> events
                      (sort-by :offset)
                      (map :duration))]
      (testing "offset is bumped forward by the duration of the outermost cram,
                twice"
        (is (offset= s (offset+ start 2000) end)))
      (testing "note durations are divided up correctly"
        (is (= [500.0 250.0 250.0 500.0 250.0 250.0] offsets))))))

