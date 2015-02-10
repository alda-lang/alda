(ns alda.lisp-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]
            [instaparse.core :as insta]))

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
    (tempo 120)
    (quant 100)
    (is (== (:duration (note (pitch "c") (duration (note-length 4)))) 500))
    (quant 90)
    (is (== (:duration (note (pitch "c") (duration (note-length 4)))) 450))))

(deftest lisp-test
  (testing "instrument part consolidation"
    (testing "all watched over by machines of loving grace"
      (let [result (parse-input (slurp "test/examples/awobmolg.alda"))]
        (pprint (eval result))))
    (testing "debussy string quartet"
      (let [result (parse-input (slurp "test/examples/debussy_quartet.alda"))]
        (pprint (eval result))))))
