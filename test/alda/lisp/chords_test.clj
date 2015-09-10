(ns alda.lisp.chords-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest chord-tests
  (testing "a chord event:"
    (let [start ($current-offset)
          {:keys [events]}
          (first
           (chord (note (pitch :c) (duration (note-length 1)))
                  (note (pitch :e) (duration (note-length 4)))
                  (pause (duration (note-length 8)))
                  (note (pitch :g) (duration (note-length 2)))))]
      (testing "the notes should all start at the same time"
        (is (every? #(= start (:offset %)) events)))
      (testing ":current-offset should be bumped forward by the shortest
        note/rest duration"
        (let [dur-fn (:duration-fn (duration (note-length 8)))
              bump   (dur-fn ($tempo))]
          (is (= ($current-offset) (offset+ start bump)))))
      (testing ":last-offset should be updated correctly"
        (is (offset= ($last-offset) start))))))

