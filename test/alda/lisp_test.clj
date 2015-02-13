(ns alda.lisp-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]))

(deftest attribute-tests
  (testing "octaves"
    (octave 4)
    (is (= *octave* 4))
    (octave 2)
    (is (= *octave* 2))
    (octave "<")
    (is (= *octave* 1))
    (octave ">")
    (is (= *octave* 2))
    (set-attribute "octave" 5)
    (is (= *octave* 5)))
  (testing "volume"
    (volume 50)
    (is (== *volume* 0.5))
    (volume 75)
    (is (== *volume* 0.75))
    (set-attribute "volume" 100)
    (is (== *volume* 1.0)))
  (testing "panning"
    (panning 25)
    (is (== *panning* 0.25))
    (panning 75)
    (is (== *panning* 0.75))
    (set-attribute "panning" 75))
  (testing "quantization"
    (quant 50)
    (is (== *quant* 0.5))
    (quant 100)
    (is (== *quant* 1.0))
    (set-attribute "quant" 90)
    (is (== *quant* 0.9))))

(deftest duration-tests
  (testing "note-length converts note length to number of beats"
    (is (== 1 (note-length 4)))
    (is (== 1.5 (note-length 4 {:dots 1})))
    (is (== 4 (note-length 1)))
    (is (== 6 (note-length 1 {:dots 1})))
    (is (== 7 (note-length 1 {:dots 2}))))
  (testing "duration converts beats to ms"
    (tempo 60)
    (is (= {:duration 1000 :slurred true} (duration (note-length 4) :slur)))
    (is (= {:duration 7500 :slurred true}
           (duration (note-length 2)
                     (note-length 2)
                     (note-length 2 {:dots 2}) :slur)))
    (tempo 120)
    (is (= {:duration 500 :slurred false} (duration (note-length 4))))
    (is (= {:duration 750 :slurred false} (duration (note-length 4 {:dots 1})))))
  (testing "quantization quantizes note durations"
    (set-attributes "tempo" 120 "quant" 100)
    (is (== (:duration (note (pitch "c") (duration (note-length 4)))) 500))
    (quant 0)
    (is (== (:duration (note (pitch "c") (duration (note-length 4)))) 0))
    (quant 90)
    (is (== (:duration (note (pitch "c") (duration (note-length 4)))) 450))))

(deftest pitch-tests
  (testing "pitch converts a note in a given octave to frequency in Hz"
    (octave 4)
    (is (== 440 (pitch "a")))
    (octave 5)
    (is (== 880 (pitch "a")))
    (octave 4)
    (is (< 261 (pitch "c") 262)))
  (testing "flats and sharps"
    (is (>  (pitch "c" :sharp) (pitch "c")))
    (is (<  (pitch "b" :flat)  (pitch "b")))
    (is (== (pitch "c" :sharp) (pitch "d" :flat)))
    (is (== (pitch "c" :sharp :sharp) (pitch "d")))
    (is (== (pitch "f" :flat)  (pitch "e")))
    (is (== (pitch "a" :flat :flat) (pitch "g")))
    (is (== (pitch "c" :sharp :flat :flat :sharp) (pitch "c")))))

(deftest note-tests
  (testing "*duration* is used if a duration is not provided"
    (is (== (:duration (note (pitch "c") :slur)) (:duration (duration *duration*)))))
  (testing "a note is placed at the current offset"
    (let [offset *current-offset*]
      (is (== offset (:offset (note (pitch "c")))))))
  (testing "a note event bumps *current-offset* forward by its duration"
    (let [offset *current-offset*
          note-duration (:duration (note (pitch "c") (duration (note-length 4) :slur)))]
        (is (== (+ offset note-duration) *current-offset*))))
  (testing "a note event has the pitch it is provided"
    (let [g-sharp (pitch "g" :sharp)]
      (is (== g-sharp (:pitch (note (pitch "g" :sharp))))))))

(deftest chord-tests
  (testing "the notes/rests in a chord should all start at the same time"
    (alter-var-root #'*current-offset* (constantly 0))
    (let [start *current-offset*
          new-notes (:events (chord (note (pitch "c"))
                                    (note (pitch "c"))
                                    (pause)
                                    (note (pitch "c"))))]
      (is (every? #(= start (:offset %)) new-notes))))
  (testing "the shortest note/rest duration in the chord should bump
            *current-offset* forward by that duration"
    (alter-var-root #'*current-offset* (constantly 0))
    (let [start *current-offset*]
      (chord (note (pitch "c") (duration (note-length 1)))
             (note (pitch "e") (duration (note-length 4)))
             (pause (duration (note-length 8)))
             (note (pitch "g") (duration (note-length 2))))
      (is (= *current-offset* (+ start (:duration (duration (note-length 8)))))))))

(deftest voice-tests
  (testing "a voice returns as many notes as it has"
    (is (= 5
           (count (voice 42
                    (note (pitch "c")) (note (pitch "d")) (note (pitch "e"))
                    (note (pitch "f")) (note (pitch "g")))))))
  (testing "a voice has the notes that it has"
    (let [a-note (note (pitch "a"))
          b-note (note (pitch "b"))
          c-note (note (pitch "c"))
          the-voice (voice 1 a-note b-note c-note)]
      (is (contains? (set the-voice) a-note))
      (is (contains? (set the-voice) b-note))
      (is (contains? (set the-voice) c-note))))
  (testing "the first note of each voice should all start at the same time"
    (alter-var-root #'*current-offset* (constantly 0))
    (let [start *current-offset*
          {:keys [v1 v2 v3]} (voices
                               (voice 1 (note (pitch "g")) (note (pitch "b")))
                               (voice 2 (note (pitch "b")) (note (pitch "d")))
                               (voice 3 (note (pitch "d")) (note (pitch "f"))))]
      (is (every? #(= start (:offset %)) (map first [v1 v2 v3])))))
  (testing "the voice lasting the longest should bump *current-offset* forward
            by however long it takes to finish"
    (alter-var-root #'*current-offset* (constantly 0))
    (let [start *current-offset*
          voice-group (voices
                        (voice 1
                          (note (pitch "g") (duration (note-length 1)))
                          (note (pitch "b") (duration (note-length 2))))
                        (voice 2
                          (note (pitch "b") (duration (note-length 1)))
                          (note (pitch "d") (duration (note-length 1))))
                        (voice 3
                          (note (pitch "d") (duration (note-length 1)))
                          (note (pitch "f") (duration (note-length 4)))))]
      (is (= *current-offset* (+ start
                                 (:duration (duration (note-length 1)
                                                      (note-length 1)))))))))


(deftest lisp-test
  (testing "instrument part consolidation"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/examples/awobmolg.alda"))]
        (pprint (eval result))))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/examples/debussy_quartet.alda"))]
        (pprint (eval result))))))