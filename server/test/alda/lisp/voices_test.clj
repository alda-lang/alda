(ns alda.lisp.voices-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest voice-tests
  (testing "a voice returns as many notes as it has"
    (is (= 5 (count (voice 42
                      (note (pitch :c)) (note (pitch :d)) (note (pitch :e))
                      (note (pitch :f)) (note (pitch :g)))))))
  (testing "a voice has the notes that it has"
    (let [a-note (note (pitch :a))
          b-note (note (pitch :b))
          c-note (note (pitch :c))
          the-voice (voice 1 a-note b-note c-note)
          has-note? (fn [voice note]
                      ; (a note is technically a list of Note records,
                      ; one for each instrument in *current-instruments*)
                      (every? #(contains? (set voice) %) note))]
      (is (has-note? the-voice a-note))
      (is (has-note? the-voice b-note))
      (is (has-note? the-voice c-note))
      (is (= 3 (count the-voice)))))
  (testing "a voice group:"
    (let [start ($current-offset)
          {:keys [v1 v2 v3]} (voices
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
                                 (note (pitch :g))))]
      (testing "the first note of each voice should all start at the same time"
        (is (every? #(= start (:offset %)) (map first [v1 v2 v3]))))
      (testing "repeated calls to the same voice should append events"
        (is (= 6 (count v2))))
      (testing "the voice lasting the longest should bump :current-offset
        forward by however long it takes to finish"
        (let [dur-fn (:duration-fn (duration (note-length 1)
                                             (note-length 1)
                                             (note-length 1)
                                             (note-length 1)))
              bump (dur-fn ($tempo))]
          (is (offset= ($current-offset) (offset+ start bump)))))
      (testing ":last-offset should be updated to the :last-offset as of the
        point where the longest voice finishes"
        (let [dur-fn (:duration-fn (duration (note-length 1)
                                             (note-length 1)
                                             (note-length 1)))
              bump (dur-fn ($tempo))]
          (is (offset= ($last-offset) (offset+ start bump))))))))

