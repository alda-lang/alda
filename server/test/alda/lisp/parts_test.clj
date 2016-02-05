(ns alda.lisp.parts-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]))

(deftest part-tests
  (testing "a part:"
    (let [s       (score (part "piano/trumpet 'trumpiano'"))
          piano   (get-instrument s "piano")
          trumpet (get-instrument s "trumpet")]
      (testing "starts at offset 0"
        (is (zero? (:offset (:current-offset piano))))
        (is (zero? (:offset (:current-offset trumpet)))))
      (testing "starts at the :start marker"
        (is (= :start (:current-marker piano)))
        (is (= :start (:current-marker trumpet))))
      (testing "has the instruments that it has"
        (let [{:keys [current-instruments]} s]
          (is (= 2 (count current-instruments)))
          (is (some #(re-find #"^piano-" %) current-instruments))
          (is (some #(re-find #"^trumpet-" %) current-instruments))))
      (testing "sets a nickname if applicable"
        (is (contains? (:nicknames s) "trumpiano"))
        (let [trumpiano (get (:nicknames s) "trumpiano")]
          (is (= (count trumpiano) 2))
          (is (some #(re-find #"^piano-" %) trumpiano))
          (is (some #(re-find #"^trumpet-" %) trumpiano))))
      (let [s              (continue s
                             (note (pitch :d)
                                   (duration (note-length 2 {:dots 1}))))
            piano          (get-instrument s "piano")
            trumpet        (get-instrument s "trumpet")
            piano-offset   (:current-offset piano)
            trumpet-offset (:current-offset trumpet)]
        (testing "instruments from a group can be separated at will"
          (let [s     (continue s
                        (part "piano"))
                piano (get-instrument s "piano")]
            (is (= 1 (count (:current-instruments s))))
            (is (re-find #"^piano-" (first (:current-instruments s))))
            (is (= piano-offset (:current-offset piano)))
            (let [s       (continue s
                            (part "trumpet"))
                  trumpet (get-instrument s "trumpet")]
              (is (= 1 (count (:current-instruments s))))
              (is (re-find #"^trumpet-" (first (:current-instruments s))))
              (is (= trumpet-offset (:current-offset trumpet)))))))
      (testing "referencing a nickname"
        (let [{:keys [current-instruments] :as s}
              (continue s
                (part "bassoon"
                  (note (pitch :c))
                  (note (pitch :d))
                  (note (pitch :e))
                  (note (pitch :f))
                  (note (pitch :g)))
                (part "trumpiano"))]
          (is (= 2 (count current-instruments)))
          (is (some #(re-find #"^piano-" %) current-instruments))
          (is (some #(re-find #"^trumpet-" %) current-instruments)))))))

