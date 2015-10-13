(ns alda.parser.events-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest note-tests
  (testing "notes"
    (is (= (test-parse :note "c")
           '(alda.lisp/note (alda.lisp/pitch :c))))
    (is (= (test-parse :note "c4")
           '(alda.lisp/note (alda.lisp/pitch :c)
                            (alda.lisp/duration (alda.lisp/note-length 4)))))
    (is (= (test-parse :note "c+")
           '(alda.lisp/note (alda.lisp/pitch :c :sharp))))
    (is (= (test-parse :note "b-")
           '(alda.lisp/note (alda.lisp/pitch :b :flat)))))
  (testing "rests"
    (is (= (test-parse :rest "r")
           '(alda.lisp/pause))
        (= (test-parse :rest "r1")
           '(alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 1)))))))

(deftest chord-tests
  (testing "chords"
    (is (= (test-parse :chord "c/e/g")
           '(alda.lisp/chord
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/note (alda.lisp/pitch :g)))))
    (is (= (test-parse :chord "c1/>e2/g4/r8")
           '(alda.lisp/chord
              (alda.lisp/note (alda.lisp/pitch :c)
                              (alda.lisp/duration (alda.lisp/note-length 1)))
              (alda.lisp/octave :up)
              (alda.lisp/note (alda.lisp/pitch :e)
                              (alda.lisp/duration (alda.lisp/note-length 2)))
              (alda.lisp/note (alda.lisp/pitch :g)
                              (alda.lisp/duration (alda.lisp/note-length 4)))
              (alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 8))))))))

(deftest voice-tests
  (testing "voices"
    (is (= (test-parse :voice "V1: a b c")
           '(alda.lisp/voice 1
              (alda.lisp/note (alda.lisp/pitch :a))
              (alda.lisp/note (alda.lisp/pitch :b))
              (alda.lisp/note (alda.lisp/pitch :c)))))
    (is (= (test-parse :voices "V1: a b c
                                V2: d e f")
           '(alda.lisp/voices
              (alda.lisp/voice 1
                (alda.lisp/note (alda.lisp/pitch :a))
                (alda.lisp/note (alda.lisp/pitch :b))
                (alda.lisp/note (alda.lisp/pitch :c)))
              (alda.lisp/voice 2
                (alda.lisp/note (alda.lisp/pitch :d))
                (alda.lisp/note (alda.lisp/pitch :e))
                (alda.lisp/note (alda.lisp/pitch :f))))))))

(deftest marker-tests
  (testing "markers"
    (is (= (test-parse :marker "%chorus") '(alda.lisp/marker "chorus")))
    (is (= (test-parse :at-marker "@verse-1") '(alda.lisp/at-marker "verse-1")))))

(deftest cram-tests
  (testing "crams"
    (is (= (test-parse :cram "{c d e}")
           '(alda.lisp/cram
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e)))))
    (is (= (test-parse :cram "{c d e}2")
           '(alda.lisp/cram
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/duration (alda.lisp/note-length 2)))))))
