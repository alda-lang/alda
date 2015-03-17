(ns alda.parser-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.parser :refer :all]
            [instaparse.core :as insta]
            [clojure.java.io :as io]))

(defn test-parse
  "Uses instaparse's partial parse mode to parse individual pieces of a score."
  [start input]
  (with-redefs [alda.parser/alda-parser
                #((insta/parser (io/resource "alda.bnf")) % :start start)]
    (parse-input input)))

(deftest duration-tests
  (testing "duration"
    (is (= (test-parse :duration "2")
           '(alda.lisp/duration (alda.lisp/note-length 2))))
    (testing "dots"
      (is (= (test-parse :duration "2..")
             '(alda.lisp/duration (alda.lisp/note-length 2 {:dots 2})))))
    (testing "ties"
      (is (= (test-parse :duration "1~2~4")
             '(alda.lisp/duration (alda.lisp/note-length 1)
                                  (alda.lisp/note-length 2)
                                  (alda.lisp/note-length 4)))))
    (testing "slurs"
      (is (= (test-parse :duration "4~")
             '(alda.lisp/duration (alda.lisp/note-length 4) :slur))))))

(deftest attribute-tests
  (testing "octave"
    (is (= (test-parse :octave-change ">") '(alda.lisp/octave :up)))
    (is (= (test-parse :octave-change "<") '(alda.lisp/octave :down)))
    (is (= (test-parse :octave-change "o5") '(alda.lisp/octave 5))))
  (testing "volume, etc."
    (is (= (test-parse :attribute-change "volume 50")
           '(alda.lisp/set-attribute :volume 50)))
    (is (= (test-parse :attribute-change "tempo 100")
           '(alda.lisp/set-attribute :tempo 100)))
    (is (= (test-parse :attribute-change "quant 75")
           '(alda.lisp/set-attribute :quant 75)))
    (is (= (test-parse :attribute-change "panning 0")
           '(alda.lisp/set-attribute :panning 0)))
    (is (= (test-parse :attribute-change "note-length 1")
           '(alda.lisp/set-attribute :note-length
                                     (alda.lisp/duration
                                       (alda.lisp/note-length 1)))))
    (is (= (test-parse :attribute-change "note-length 2..")
           '(alda.lisp/set-attribute :note-length
                                     (alda.lisp/duration
                                       (alda.lisp/note-length 2 {:dots 2}))))))
  (testing "attribute changes"
    (is (= (test-parse :attribute-changes "(volume 50, tempo 100)")
           '((alda.lisp/set-attribute :volume 50)
             (alda.lisp/set-attribute :tempo 100))))
    (is (= (test-parse :attribute-changes "(quant! 50, tempo 90)")
           '((alda.lisp/global-attribute :quant 50)
             (alda.lisp/set-attribute :tempo 90)))))
  (testing "global attribute changes"
    (is (= (test-parse :global-attribute-change "tempo! 126")
           '(alda.lisp/global-attribute :tempo 126)))
    (is (= (test-parse :global-attributes "(tempo! 130, quant! 80)")
           '((alda.lisp/global-attribute :tempo 130)
             (alda.lisp/global-attribute :quant 80))))))

(deftest event-tests
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
           '(alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 1))))))
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
              (alda.lisp/pause (alda.lisp/duration (alda.lisp/note-length 8)))))))
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
  (is (= (test-parse :marker "%chorus") '(alda.lisp/marker "chorus")))
  (is (= (test-parse :at-marker "@verse-1") '(alda.lisp/at-marker "verse-1"))))

(deftest score-tests
  (is (= (parse-input "theremin: c d e")
         '(alda.lisp/score
            (alda.lisp/part {:names ["theremin"]}
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))))))
  (is (= (parse-input "trumpet/trombone/tuba \"brass\": f+1")
         '(alda.lisp/score
            (alda.lisp/part {:names ["trumpet" "trombone" "tuba"]
                             :nickname "brass"}
              (alda.lisp/note (alda.lisp/pitch :f :sharp)
                              (alda.lisp/duration (alda.lisp/note-length 1)))))))
  (is (= (parse-input "guitar: e
                       bass: e")
         '(alda.lisp/score
            (alda.lisp/part {:names ["guitar"]}
              (alda.lisp/note (alda.lisp/pitch :e)))
            (alda.lisp/part {:names ["bass"]}
              (alda.lisp/note (alda.lisp/pitch :e)))))))
