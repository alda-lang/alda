(ns alda.lisp.attributes-test
  (:require [clojure.test :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest octave-tests
  (testing "octaves"
    (octave 4)
    (is (= ($octave) 4))
    (octave 2)
    (is (= ($octave) 2))
    (octave :down)
    (is (= ($octave) 1))
    (octave :up)
    (is (= ($octave) 2))
    (set-attribute :octave 5)
    (is (= ($octave) 5))))

(deftest volume-tests
  (testing "volume"
    (volume 50)
    (is (== ($volume) 0.5))
    (volume 75)
    (is (== ($volume) 0.75))
    (set-attribute :volume 100)
    (is (== ($volume) 1.0))))

(deftest panning-tests
  (testing "panning"
    (panning 25)
    (is (== ($panning) 0.25))
    (panning 75)
    (is (== ($panning) 0.75))
    (set-attribute :panning 50)
    (is (== ($panning) 0.5))))

(deftest quantization-tests
  (testing "quantization"
    (quant 50)
    (is (== ($quantization) 0.5))
    (quant 100)
    (is (== ($quantization) 1.0))
    (quant 9001)
    (is (== ($quantization) 90.01))
    (set-attribute :quant 90)
    (is (== ($quantization) 0.9))))

(deftest note-length-tests
  (testing "note-length"
    (set-attribute :note-length (duration (note-length 2 {:dots 2})))
    (is (== ($duration) 3.5))
    (set-attribute :note-length (duration (note-length 1) (note-length 1)))
    (is (== ($duration) 8))))
