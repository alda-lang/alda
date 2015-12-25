(ns alda.lisp.duration-test
  (:require [clojure.test :refer :all]
            [alda.lisp :refer :all]))

(use-fixtures :each
  (fn [run-tests]
    (score*)
    (part* "piano")
    (run-tests)))

(deftest duration-tests
  (testing "note-length converts note length to number of beats"
    (is (== 1 (:value (note-length 4))))
    (is (== 1.5 (:value (note-length 4 {:dots 1}))))
    (is (== 4 (:value (note-length 1))))
    (is (== 6 (:value (note-length 1 {:dots 1}))))
    (is (== 7 (:value (note-length 1 {:dots 2})))))
  (testing "duration converts beats to ms"
    (let [{:keys [duration-fn]} (duration (note-length 4) :slur)]
      (is (== 1000 (duration-fn 60))))
    (let [{:keys [duration-fn]} (duration (note-length 2)
                                          (note-length 2)
                                          (note-length 2 {:dots 2}) :slur)]
      (is (== 7500 (duration-fn 60))))
    (let [{:keys [duration-fn]} (duration (note-length 4))]
      (is (== 500 (duration-fn 120))))
    (let [{:keys [duration-fn]} (duration (note-length 4 {:dots 1}))]
      (is (== 750 (duration-fn 120)))))
  (testing "duration can be described in milliseconds"
    (let [{:keys [duration-fn]} (duration (ms 1000) :slur)]
      (is (== 1000 (duration-fn 42))))
    (let [{:keys [duration-fn]} (duration (ms 2000)
                                          (ms 2000)
                                          (ms 3500) :slur)]
      (is (== 7500 (duration-fn 123)))))
  (testing "note-lengths and millisecond values can be combined"
    (let [{:keys [duration-fn]} (duration (ms 2000)
                                          (note-length 2)
                                          (ms 45) :slur)]
      (is (== 4045 (duration-fn 60))))
    (let [{:keys [duration-fn]} (duration (note-length 1 {:dots 1})
                                          (ms 333) :slur)]
      (is (== 3333 (duration-fn 120)))))
  (testing "barlines don't break duration"
    (let [{:keys [duration-fn]} (duration (note-length 4)
                                          (barline)
                                          (note-length 4) :slur)]
      (is (== 2000 (duration-fn 60)))))
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
                                :slur))))))))

