(ns alda.lisp-attribute-test
  (:require [clojure.test :refer :all]
            [alda.lisp :refer :all]))

(deftest attribute-tests
  (part {:names ["piano"]}
    (let [piano (fn [] (*instruments* "piano"))
          current-octave  (fn [] (:octave (piano)))
          current-volume  (fn [] (:volume (piano)))
          current-panning (fn [] (:panning (piano)))
          current-quant   (fn [] (:quantization (piano)))]
      (testing "octaves"
        (octave 4)
        (is (= (current-octave) 4))
        (octave 2)
        (is (= (current-octave) 2))
        (octave :down)
        (is (= (current-octave) 1))
        (octave :up)
        (is (= (current-octave) 2))
        (set-attribute :octave 5)
        (is (= (current-octave) 5)))
      (testing "volume"
        (volume 50)
        (is (== (current-volume) 0.5))
        (volume 75)
        (is (== (current-volume) 0.75))
        (set-attribute :volume 100)
        (is (== (current-volume) 1.0)))
      (testing "panning"
        (panning 25)
        (is (== (current-panning) 0.25))
        (panning 75)
        (is (== (current-panning) 0.75))
        (set-attribute :panning 50)
        (is (== (current-panning) 0.5)))
      (testing "quantization"
        (quant 50)
        (is (== (current-quant) 0.5))
        (quant 100)
        (is (== (current-quant) 1.0))
        (set-attribute :quant 90)
        (is (== (current-quant) 0.9))))))

(deftest duration-tests
  (part {:names ["piano"]}
    (testing "note-length converts note length to number of beats"
      (is (== 1 (note-length 4)))
      (is (== 1.5 (note-length 4 {:dots 1})))
      (is (== 4 (note-length 1)))
      (is (== 6 (note-length 1 {:dots 1})))
      (is (== 7 (note-length 1 {:dots 2}))))
    (testing "duration converts beats to ms"
      (let [{:keys [duration-fn]} (duration (note-length 4) :slur)]
        (is (== 1000 (duration-fn 60))))
      (let [{:keys [duration-fn]} (duration (note-length 2)
                                            (note-length 2)
                                            (note-length 2 {:dots 2}) :slur)]
        (is (= 7500 (duration-fn 60))))
      (let [{:keys [duration-fn]} (duration (note-length 4))]
        (is (= 500 (duration-fn 120))))
      (let [{:keys [duration-fn]} (duration (note-length 4 {:dots 1}))]
        (is (= 750 (duration-fn 120)))))
    (testing "quantization quantizes note durations"
      (set-attributes :tempo 120 :quant 100)
      (is (== 500
              (:duration (first
                           (note (pitch :c) (duration (note-length 4)))))))
      (quant 0)
      (is (== 0
              (:duration (first
                           (note (pitch :c) (duration (note-length 4)))))))
      (quant 90)
      (is (== 450
              (:duration (first
                           (note (pitch :c) (duration (note-length 4)))))))
      (testing "slurred notes ignore quantization"
        (quant 90)
        (is (== 500
                (:duration (first
                             (note (pitch :c)
                                   (duration (note-length 4) :slur))))))
        (is (== 1000
                (:duration (first
                             (note (pitch :c)
                                   (duration (note-length 2))
                                   :slur)))))))))

(deftest pitch-tests
  (part {:names ["piano"]}
    (testing "pitch converts a note in a given octave to frequency in Hz"
      (is (== 440 ((pitch :a) 4)))
      (is (== 880 ((pitch :a) 5)))
      (is (< 261 ((pitch :c) 4) 262)))
    (testing "flats and sharps"
      (is (>  ((pitch :c :sharp) 4) ((pitch :c) 4)))
      (is (>  ((pitch :c) 5) ((pitch :c :sharp) 4)))
      (is (<  ((pitch :b :flat) 4)  ((pitch :b) 4)))
      (is (== ((pitch :c :sharp) 4) ((pitch :d :flat) 4)))
      (is (== ((pitch :c :sharp :sharp) 4) ((pitch :d) 4)))
      (is (== ((pitch :f :flat) 4) ((pitch :e) 4)))
      (is (== ((pitch :a :flat :flat) 4) ((pitch :g) 4)))
      (is (== ((pitch :c :sharp :flat :flat :sharp) 4) ((pitch :c) 4))))))