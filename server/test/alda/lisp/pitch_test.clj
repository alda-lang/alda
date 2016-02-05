(ns alda.lisp.pitch-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]))

(deftest pitch-tests
  (testing "pitch converts a note in a given octave to frequency in Hz"
    (is (== 440 ((pitch :a) 4 {})))
    (is (== 880 ((pitch :a) 5 {})))
    (is (< 261 ((pitch :c) 4 {}) 262)))
  (testing "flats and sharps"
    (is (> ((pitch :c :sharp) 4 {})
           ((pitch :c) 4 {})))
    (is (> ((pitch :c) 5 {})
           ((pitch :c :sharp) 4 {})))
    (is (< ((pitch :b :flat) 4 {})
           ((pitch :b) 4 {})))
    (is (== ((pitch :c :sharp) 4 {})
            ((pitch :d :flat) 4 {})))
    (is (== ((pitch :c :sharp :sharp) 4 {})
            ((pitch :d) 4 {})))
    (is (== ((pitch :f :flat) 4 {})
            ((pitch :e) 4 {})))
    (is (== ((pitch :a :flat :flat) 4 {})
            ((pitch :g) 4 {})))
    (is (== ((pitch :c :sharp :flat :flat :sharp) 4 {})
            ((pitch :c) 4 {})))))

(deftest key-tests
  (testing "you can set and get a key signature"
    (let [s     (score
                  (part "piano"
                    (key-signature {:b [:flat] :e [:flat]})))
          piano (get-instrument s "piano")]
      (is (= {:b [:flat] :e [:flat]}
             (:key-signature piano))))
    (let [s     (score
                  (part "piano"
                    (key-sig "f+ c+ g+")))
          piano (get-instrument s "piano")]
      (is (= {:f [:sharp] :c [:sharp] :g [:sharp]}
             (:key-signature piano))))
    (let [s     (score
                  (part "piano"
                    (key-sig [:a :flat :major])))
          piano (get-instrument s "piano")]
      (is (= {:b [:flat] :e [:flat] :a [:flat] :d [:flat]}
             (:key-signature piano))))
    (let [s     (score
                  (part "piano"
                    (key-sig [:e :minor])))
          piano (get-instrument s "piano")]
      (is (= {:f [:sharp]}
             (:key-signature piano)))))
  (testing "the pitch of a note is affected by the key signature"
    (is (= ((pitch :b) 4 {:b [:flat]})
           ((pitch :b :flat) 4 {})))
    (is (= ((pitch :b :natural) 4 {:b [:flat]})
           ((pitch :b) 4 {})))
    (let [s         (score
                      (part "piano"
                            (key-signature "f+")))
          piano     (get-instrument s "piano")
          f-sharp-4 ((pitch :f) 4 (:key-signature piano))]
      (is (= f-sharp-4 ((pitch :f :sharp) 4 {}))))))
