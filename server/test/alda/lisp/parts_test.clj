(ns alda.lisp.parts-test
  (:require [clojure.test :refer :all]
            [clojure.pprint :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp :refer :all]
            [alda.parser :refer :all]))

(deftest part-tests
  (score*)
  (testing "a part:"
    (part* "piano/trumpet 'trumpiano'")
    (testing "starts at offset 0"
      (is (zero? (:offset (:current-offset (get-instrument "piano")))))
      (is (zero? (:offset (:current-offset (get-instrument "trumpet"))))))
    (testing "starts at the :start marker"
      (is (= :start (:current-marker (get-instrument "piano"))))
      (is (= :start (:current-marker (get-instrument "trumpet")))))
    (testing "has the instruments that it has"
      (is (= 2 (count *current-instruments*)))
      (is (some #(re-find #"^piano-" %) *current-instruments*))
      (is (some #(re-find #"^trumpet-" %) *current-instruments*)))
    (testing "sets a nickname if applicable"
      (is (contains? *nicknames* "trumpiano"))
      (let [trumpiano (*nicknames* "trumpiano")]
        (is (= (count trumpiano) 2))
        (is (some #(re-find #"^piano-" %) trumpiano))
        (is (some #(re-find #"^trumpet-" %) trumpiano))))
    (note (pitch :d) (duration (note-length 2 {:dots 1})))
    (def piano-offset (-> (get-instrument "piano") :current-offset))
    (def trumpet-offset (-> (get-instrument "trumpet") :current-offset))
    (testing "instruments from a group can be separated at will"
      (part* "piano")
      (is (= 1 (count *current-instruments*)))
      (is (re-find #"^piano-" (first *current-instruments*)))
      (is (= piano-offset (-> (get-instrument "piano") :current-offset)))
      (chord (note (pitch :a))
             (note (pitch :c :sharp))
             (note (pitch :e)))
      (alter-var-root #'piano-offset
                      (constantly (-> (get-instrument "piano") :current-offset)))

      (part* "trumpet")
      (is (= 1 (count *current-instruments*)))
      (is (re-find #"^trumpet-" (first *current-instruments*)))
      (is (= trumpet-offset (-> (get-instrument "trumpet") :current-offset)))
      (note (pitch :d))
      (note (pitch :e))
      (note (pitch :f :sharp))
      (alter-var-root #'trumpet-offset
                      (constantly (-> (get-instrument "trumpet") :current-offset)))
      (is (= piano-offset (-> (get-instrument "piano") :current-offset))))
    (testing "referencing a nickname"
      (part* "trumpiano")
      (is (= 2 (count *current-instruments*)))
      (is (some #(re-find #"^piano-" %) *current-instruments*))
      (is (some #(re-find #"^trumpet-" %) *current-instruments*)))))

