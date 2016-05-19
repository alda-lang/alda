(ns alda.lisp.voices-test
  (:require [clojure.test            :refer :all]
            [alda.test-helpers       :refer (get-instrument dur->ms)]
            [alda.lisp               :refer :all]
            [alda.lisp.model.records :refer (->AbsoluteOffset)]))

(deftest voice-tests
  (testing "a voice group:"
    (testing "the first note of each voice should all start at the same time"
      (let [s      (score
                     (part "piano"
                       (voices
                         (voice 1
                           (note (pitch :g) (duration (note-length 1))))
                         (voice 2
                           (note (pitch :b) (duration (note-length 1))))
                         (voice 3
                           (note (pitch :d) (duration (note-length 1)))))))
            events (-> s :events)]
        (is (= 1 (count (distinct (map :offset events)))))))
    (let [s            (score
                         (part "piano"
                           (voices
                             (voice 1
                               (note (pitch :g) (duration (note-length 1)))
                               (note (pitch :b) (duration (note-length 2))))
                             (voice 2
                               (note (pitch :b) (duration (note-length 1)))
                               (note (pitch :d) (duration (note-length 1))))
                             (voice 3
                               (note (pitch :d) (duration (note-length 1)))
                               (note (pitch :f) (duration (note-length 4))))
                             (voice 2
                               (octave :up)
                               (octave :down)
                               (note (pitch :g))
                               (note (pitch :g))))
                           (end-voices)))
          piano        (get-instrument s "piano")
          events       (-> s :events)
          voice-events (group-by :voice events)]
      (testing "repeated calls to the same voice should append events"
        (is (= 4 (count (get voice-events 2)))))
      (let [bump  (dur->ms (duration (note-length 1)
                                     (note-length 1)
                                     (note-length 1)
                                     (note-length 1))
                           (:tempo piano))
            bump2 (dur->ms (duration (note-length 1)
                                     (note-length 1)
                                     (note-length 1))
                           (:tempo piano))]
        (testing "the voice lasting the longest should bump :current-offset
                  forward by however long it takes to finish"
          (is (offset= s
                       (:current-offset piano)
                       (offset+ (->AbsoluteOffset 0) bump))))
        (testing ":last-offset should be updated to the :last-offset as of the
                  point where the longest voice finishes"
          (is (offset= s
                       (:last-offset piano)
                       (offset+ (->AbsoluteOffset 0) bump2)))))))
  (testing "a voice containing a cram expression"
    (testing "should not throw an exception"
      (is (score
            (part "piano"
              (voices
                (voice 1
                  (cram
                    (note (pitch :c))
                    (octave :down)
                    (note (pitch :b))
                    (note (pitch :a))
                    (note (pitch :g)))))))))))

