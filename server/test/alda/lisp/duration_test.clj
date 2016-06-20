(ns alda.lisp.duration-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument dur->ms)]
            [alda.lisp         :refer :all]))

(deftest duration-tests
  (testing "note-length converts note length to number of beats"
    (is (== 1 (:value (note-length 4))))
    (is (== 1.5 (:value (note-length 4 {:dots 1}))))
    (is (== 4 (:value (note-length 1))))
    (is (== 6 (:value (note-length 1 {:dots 1}))))
    (is (== 7 (:value (note-length 1 {:dots 2})))))
  (testing "duration converts beats to ms"
    (is (== 1000 (dur->ms (duration (note-length 4)) 60)))
    (is (== 7500 (dur->ms (duration (note-length 2)
                                    (note-length 2)
                                    (note-length 2 {:dots 2}))
                          60)))
    (is (== 500 (dur->ms (duration (note-length 4)) 120)))
    (is (== 750 (dur->ms (duration (note-length 4 {:dots 1})) 120))))
  (testing "duration can be described in milliseconds"
    (is (== 1000 (dur->ms (duration (ms 1000)) 42)))
    (is (== 7500 (dur->ms (duration (ms 2000)
                                    (ms 2000)
                                    (ms 3500))
                          123))))
  (testing "note-lengths and millisecond values can be combined"
    (is (== 4045 (dur->ms (duration (ms 2000)
                                    (note-length 2)
                                    (ms 45))
                          60)))
    (is (== 3333 (dur->ms (duration (note-length 1 {:dots 1})
                                    (ms 333))
                          120))))
  (testing "barlines don't break duration"
    (is (== 2000 (dur->ms (duration (note-length 4)
                                    (barline)
                                    (note-length 4))
                          60))))
  (testing "quantization quantizes note durations"
    (let [s      (score
                   (part "piano" (tempo 120) (quant 100)
                     (note (pitch :c) (duration (note-length 4)))))
          piano  (get-instrument s "piano")
          events (:events s)]
      (is (== 500 (:duration (first events)))))
    (let [s      (score
                   (part "piano" (tempo 120) (quant 0)
                     (note (pitch :c) (duration (note-length 4)))))
          piano  (get-instrument s "piano")
          events (:events s)]
      (is (empty? events)))
    (let [s      (score
                   (part "piano" (tempo 120) (quant 90)
                     (note (pitch :c) (duration (note-length 4)))))
          piano  (get-instrument s "piano")
          events (:events s)]
      (is (== 450 (:duration (first events))))))
  (testing "slurred notes ignore quantization"
    (let [s      (score
                   (part "piano" (tempo 120) (quant 90)
                     (note (pitch :c) (duration (note-length 4)) :slur)))
          piano  (get-instrument s "piano")
          events (:events s)]
      (is (== 500 (:duration (first events)))))
    (let [s      (score
                   (part "piano" (tempo 120) (quant 90)
                     (note (pitch :c) (duration (note-length 2)) :slur)))
          piano  (get-instrument s "piano")
          events (:events s)]
      (is (== 1000 (:duration (first events)))))))

