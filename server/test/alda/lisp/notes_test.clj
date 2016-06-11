(ns alda.lisp.notes-test
  (:require [clojure.test             :refer :all]
            [alda.test-helpers        :refer (get-instrument dur->ms)]
            [alda.lisp                :refer :all]))

(deftest note-tests
  (testing "a note event:"
    (let [s     (score (part "piano"))
          piano (get-instrument s "piano")
          start (:current-offset piano)
          c     ((pitch :c) (:octave piano) (:key-signature piano))
          s     (continue s
                  (note (pitch :c) (duration (note-length 4) :slur)))
          piano (get-instrument s "piano")
          {:keys [duration offset pitch] :as note} (first (:events s))]
      (testing "should be placed at the current offset"
        (is (offset= s start offset)))
      (testing "should bump :current-offset forward by its duration"
        (is (offset= s (offset+ start duration) (:current-offset piano))))
      (testing "should update :last-offset"
        (is (offset= s start (:last-offset piano))))
      (testing "should have the pitch it was given"
        (is (== pitch c)))))
  (testing "a note event with no duration provided:"
    (let [s                  (score (part "piano"))
          piano              (get-instrument s "piano")
          default-duration   (dur->ms {:beats (:duration piano)} (:tempo piano))
          s                  (continue s
                               (note (pitch :c) :slur))
          piano              (get-instrument s "piano")
          {:keys [duration]} (first (:events s))]
      (testing "the default :duration of the instrument should be used"
        (is (== duration default-duration))))))

