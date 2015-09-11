(ns alda.lisp.notes-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest note-tests
  (testing "a note event:"
    (let [start ($current-offset)
          c ((pitch :c) ($octave))
          {:keys [duration offset pitch]}
          (first (note (pitch :c) (duration (note-length 4) :slur)))]
      (testing "should be placed at the current offset"
        (is (offset= start offset)))
      (testing "should bump :current-offset forward by its duration"
        (is (offset= (offset+ start duration) ($current-offset))))
      (testing "should update :last-offset"
        (is (offset= start ($last-offset))))
      (testing "should have the pitch it was given"
        (is (== pitch c)))))
  (testing "a note event with no duration provided:"
    (let [default-dur-fn     (:duration-fn (duration ($duration)))
          default-duration   (default-dur-fn ($tempo))
          {:keys [duration]} (first (note (pitch :c) :slur))]
      (testing "the default duration (:duration) should be used"
        (is (== duration default-duration))))))

