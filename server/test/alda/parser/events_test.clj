(ns alda.parser.events-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-with-context)]))

(deftest note-tests
  (testing "notes"
    (is (= (parse-with-context :music-data "c")
           '((alda.lisp/note (alda.lisp/pitch :c)))))
    (is (= (parse-with-context :music-data "c4")
           '((alda.lisp/note (alda.lisp/pitch :c)
                             (alda.lisp/duration (alda.lisp/note-length 4))))))
    (is (= (parse-with-context :music-data "c+")
           '((alda.lisp/note (alda.lisp/pitch :c :sharp)))))
    (is (= (parse-with-context :music-data "b-")
           '((alda.lisp/note (alda.lisp/pitch :b :flat))))))
  (testing "rests"
    (is (= (parse-with-context :music-data "r")
           '((alda.lisp/pause)))
        (= (parse-with-context :music-data "r1")
           '((alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 1))))))))

(deftest chord-tests
  (testing "chords"
    (is (= (parse-with-context :music-data "c/e/g")
           '((alda.lisp/chord
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :e))
              (alda.lisp/note (alda.lisp/pitch :g))))))
    (is (= (parse-with-context :music-data "c1/>e2/g4/r8")
           '((alda.lisp/chord
               (alda.lisp/note (alda.lisp/pitch :c)
                               (alda.lisp/duration (alda.lisp/note-length 1)))
               (alda.lisp/octave :up)
               (alda.lisp/note (alda.lisp/pitch :e)
                               (alda.lisp/duration (alda.lisp/note-length 2)))
               (alda.lisp/note (alda.lisp/pitch :g)
                               (alda.lisp/duration (alda.lisp/note-length 4)))
               (alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 8)))))))
    (is (= (parse-with-context :music-data "b>/d/f2.")
           '((alda.lisp/chord
               (alda.lisp/note (alda.lisp/pitch :b))
               (alda.lisp/octave :up)
               (alda.lisp/note (alda.lisp/pitch :d))
               (alda.lisp/note (alda.lisp/pitch :f)
                               (alda.lisp/duration (alda.lisp/note-length 2 {:dots 1})))))))))

(deftest voice-tests
  (testing "voices"
    (is (= (parse-with-context :part "piano: V1: a b c")
           '(alda.lisp/part {:names ["piano"]}
              (alda.lisp/voices
                (alda.lisp/voice 1
                  (alda.lisp/note (alda.lisp/pitch :a))
                  (alda.lisp/note (alda.lisp/pitch :b))
                  (alda.lisp/note (alda.lisp/pitch :c)))))))
    (is (= (parse-with-context :part "piano:
                                        V1: a b c
                                        V2: d e f")
           '(alda.lisp/part {:names ["piano"]}
              (alda.lisp/voices
                (alda.lisp/voice 1
                  (alda.lisp/note (alda.lisp/pitch :a))
                  (alda.lisp/note (alda.lisp/pitch :b))
                  (alda.lisp/note (alda.lisp/pitch :c)))
                (alda.lisp/voice 2
                  (alda.lisp/note (alda.lisp/pitch :d))
                  (alda.lisp/note (alda.lisp/pitch :e))
                  (alda.lisp/note (alda.lisp/pitch :f)))))))
    (is (= (parse-with-context :part "piano:
                                        V1: a b c | V2: d e f")
           '(alda.lisp/part {:names ["piano"]}
              (alda.lisp/voices
                (alda.lisp/voice 1
                  (alda.lisp/note (alda.lisp/pitch :a))
                  (alda.lisp/note (alda.lisp/pitch :b))
                  (alda.lisp/note (alda.lisp/pitch :c))
                  (alda.lisp/barline))
                (alda.lisp/voice 2
                  (alda.lisp/note (alda.lisp/pitch :d))
                  (alda.lisp/note (alda.lisp/pitch :e))
                  (alda.lisp/note (alda.lisp/pitch :f)))))))
    (is (= (parse-with-context :part "piano:
                                        V1: [a b c] *8
                                        V2: [d e f] *8")
           '(alda.lisp/part {:names ["piano"]}
              (alda.lisp/voices
                (alda.lisp/voice 1
                  (alda.lisp/times 8
                    [(alda.lisp/note (alda.lisp/pitch :a))
                     (alda.lisp/note (alda.lisp/pitch :b))
                     (alda.lisp/note (alda.lisp/pitch :c))]))
                (alda.lisp/voice 2
                  (alda.lisp/times 8
                    [(alda.lisp/note (alda.lisp/pitch :d))
                     (alda.lisp/note (alda.lisp/pitch :e))
                     (alda.lisp/note (alda.lisp/pitch :f))]))))))))

(deftest marker-tests
  (testing "markers"
    (is (= (parse-with-context :music-data "%chorus")
           '((alda.lisp/marker "chorus"))))
    (is (= (parse-with-context :music-data "@verse-1")
           '((alda.lisp/at-marker "verse-1"))))))

(deftest cram-tests
  (testing "crams"
    (is (= (parse-with-context :music-data "{c d e}")
           '((alda.lisp/cram
               (alda.lisp/note (alda.lisp/pitch :c))
               (alda.lisp/note (alda.lisp/pitch :d))
               (alda.lisp/note (alda.lisp/pitch :e))))))
    (is (= (parse-with-context :music-data "{c d e}2")
           '((alda.lisp/cram
               (alda.lisp/note (alda.lisp/pitch :c))
               (alda.lisp/note (alda.lisp/pitch :d))
               (alda.lisp/note (alda.lisp/pitch :e))
               (alda.lisp/duration (alda.lisp/note-length 2))))))))
