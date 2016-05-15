(ns alda.lisp.global-attributes-test
  (:require [clojure.test            :refer :all]
            [alda.test-helpers       :refer (get-instrument)]
            [alda.lisp               :refer :all]
            [alda.lisp.model.records :refer (->RelativeOffset)]))

(deftest global-attribute-tests
  (testing "a global tempo change:"
    (let [s     (score
                  (part "piano"
                    (pause (duration (note-length 1))) ; 2000 ms from start
                    (global-attribute :tempo 60)))
          piano (get-instrument s "piano")]
      (testing "it should exist at the right point in the score"
        (is (= (get (:global-attributes s) 2000.0) {:tempo [60]})))
      (testing "it should change the tempo"
        (let [s     (continue s)
              piano (get-instrument s "piano")]
          (is (= (:tempo piano) 60))))
      (testing "when another part starts,"
        (let [s       (continue s
                        (pause (duration (note-length 1)
                                         (note-length 1)
                                         (note-length 1)))
                        (marker "test-marker-3") ; for later test
                        (part "viola"))
              tempo-1 (:tempo (get-instrument s "viola"))
              s       (continue s
                        (pause (duration (note-length 2 {:dots 1}))))
              tempo-2 (:tempo (get-instrument s "viola"))
              s       (continue s
                        (pause))
              tempo-3 (:tempo (get-instrument s "viola"))]
          (testing "the tempo should change once it encounters the global attribute"
            (is (= tempo-1 120)) ; not yet...
            (is (= tempo-2 120)) ; not yet...
            (is (= tempo-3 60))) ; now!
          (testing "it should use absolute offset, not relative to marker"
            (let [s       (continue s
                            (at-marker "test-marker-3"))
                  viola   (get-instrument s "viola")
                  marker  (:current-marker viola)
                  offset  (:current-offset viola)
                  s       (continue s
                            (tempo 120))
                  tempo-1 (:tempo (get-instrument s "viola"))
                  s       (continue s
                            (pause (duration (note-length 2 {:dots 1}))))
                  tempo-2 (:tempo (get-instrument s "viola"))
                  s       (continue s
                            (pause))
                  tempo-3 (:tempo (get-instrument s "viola"))]
              (is (= marker "test-marker-3"))
              (is (offset= s offset (->RelativeOffset "test-marker-3" 0)))
              (is (= tempo-1 120))
              (is (= tempo-2 120))
              ; tempo should still be 120, despite having passed 2000 ms
              (is (= tempo-3 120)))))))))

