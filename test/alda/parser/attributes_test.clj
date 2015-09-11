(ns alda.parser.attributes-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(deftest octave-tests
  (testing "octave change"
    (is (= '(alda.lisp/octave :up)   (test-parse :octave-up ">")))
    (is (= '(alda.lisp/octave :down) (test-parse :octave-down "<")))
    (is (= '(alda.lisp/octave 5)     (test-parse :octave-set "o5")))))

(deftest misc-attribute-tests
  (testing "volume change"
    (is (= (test-parse :attribute-change "volume 50")
           '(alda.lisp/set-attribute :volume 50))))
  (testing "tempo change"
    (is (= (test-parse :attribute-change "tempo 100")
           '(alda.lisp/set-attribute :tempo 100))))
  (testing "quantization change"
    (is (= (test-parse :attribute-change "quant 75")
           '(alda.lisp/set-attribute :quant 75))))
  (testing "panning change"
    (is (= (test-parse :attribute-change "panning 0")
           '(alda.lisp/set-attribute :panning 0))))
  (testing "note-length change"
    (is (= (test-parse :attribute-change "note-length 1")
           '(alda.lisp/set-attribute :note-length
                                     (alda.lisp/duration
                                       (alda.lisp/note-length 1)))))
    (is (= (test-parse :attribute-change "note-length 2..")
           '(alda.lisp/set-attribute :note-length
                                     (alda.lisp/duration
                                       (alda.lisp/note-length 2 {:dots 2})))))))

(deftest multiple-attribute-change-tests
  (testing "attribute changes"
    (is (= (test-parse :attribute-changes "(volume 50, tempo 100)")
           '((alda.lisp/set-attribute :volume 50)
             (alda.lisp/set-attribute :tempo 100))))
    (is (= (test-parse :attribute-changes "(quant! 50, tempo 90)")
           '((alda.lisp/global-attribute :quant 50)
             (alda.lisp/set-attribute :tempo 90)))))
  (testing "global attribute changes"
    (is (= (test-parse :global-attribute-change "tempo! 126")
           '(alda.lisp/global-attribute :tempo 126)))
    (is (= (test-parse :global-attributes "(tempo! 130, quant! 80)")
           '((alda.lisp/global-attribute :tempo 130)
             (alda.lisp/global-attribute :quant 80))))))
