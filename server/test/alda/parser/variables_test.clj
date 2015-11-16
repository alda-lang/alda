(ns alda.parser.variables-test
  (:require [clojure.test      :refer :all]
            [alda.parser-util  :refer (parse-to-lisp-with-context)]
            [instaparse.core   :refer (failure?)]))

(deftest variable-name-tests
  (testing "variable names"
    (testing "must start with two letters"
      (is (= '((alda.lisp/get-variable :aa))
             (parse-to-lisp-with-context :music-data "aa")))
      (is (= '((alda.lisp/get-variable :aaa))
             (parse-to-lisp-with-context :music-data "aaa")))
      (is (= '((alda.lisp/get-variable :HI))
             (parse-to-lisp-with-context :music-data "HI")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "x")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "y2")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "1234kittens")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "r2d2")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "i_like_underscores"))))
    (testing "can't contain pluses or minuses"
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "jar-jar-binks")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "han+leia")))
      (is (thrown? Exception (parse-to-lisp-with-context :music-data "ionlyprograminc++"))))
    (testing "can contain digits"
      (is (= '((alda.lisp/get-variable :celloPart2))
             (parse-to-lisp-with-context :music-data "celloPart2")))
      (is (= '((alda.lisp/get-variable :xy42))
             (parse-to-lisp-with-context :music-data "xy42")))
      (is (= '((alda.lisp/get-variable :my20cats))
             (parse-to-lisp-with-context :music-data "my20cats"))))
    (testing "can contain underscores"
      (is (= '((alda.lisp/get-variable :apple_cider))
             (parse-to-lisp-with-context :music-data "apple_cider")))
      (is (= '((alda.lisp/get-variable :underscores__are___great____))
             (parse-to-lisp-with-context :music-data "underscores__are___great____"))))))

(deftest variable-get-tests
  (testing "variable getting"
    (is (= '(alda.lisp/score
              (alda.lisp/part {:names ["flute"]}
                (alda.lisp/note (alda.lisp/pitch :c))
                (alda.lisp/get-variable :flan)
                (alda.lisp/note (alda.lisp/pitch :f))))
           (parse-to-lisp-with-context :score "flute: c flan f")))
    (is (= '(alda.lisp/score
              (alda.lisp/part {:names ["clarinet"]}
                (alda.lisp/get-variable :pudding123)))
           (parse-to-lisp-with-context :score "clarinet: pudding123")))))

(deftest variable-set-tests
  (testing "variable setting"
    (testing "within an instrument part"
      (is (= '(alda.lisp/score
                (alda.lisp/part {:names ["harpsichord"]}
                  (alda.lisp/set-variable :custard_
                    (alda.lisp/note (alda.lisp/pitch :c))
                    (alda.lisp/note (alda.lisp/pitch :d))
                    (alda.lisp/chord
                      (alda.lisp/note (alda.lisp/pitch :e))
                      (alda.lisp/note (alda.lisp/pitch :g))))))
             (parse-to-lisp-with-context :score "harpsichord:\n\ncustard_ = c d e/g")))
      (is (= '(alda.lisp/score
                (alda.lisp/part {:names ["glockenspiel"]}
                  (alda.lisp/set-variable :sorbet
                    (alda.lisp/note (alda.lisp/pitch :c))
                    (alda.lisp/note (alda.lisp/pitch :d))
                    (alda.lisp/chord
                      (alda.lisp/note (alda.lisp/pitch :e))
                      (alda.lisp/note (alda.lisp/pitch :g))))
                  (alda.lisp/note (alda.lisp/pitch :c))))
             (parse-to-lisp-with-context :score "glockenspiel:\n\nsorbet=c d e/g\nc"))))
    (testing "at the top of a score"
      (is (= '(alda.lisp/score
                (alda.lisp/set-variable :GELATO
                  (alda.lisp/note (alda.lisp/pitch :d))
                  (alda.lisp/note (alda.lisp/pitch :e)))
                (alda.lisp/part {:names ["clavinet"]}
                  (alda.lisp/chord
                    (alda.lisp/note (alda.lisp/pitch :c))
                    (alda.lisp/note (alda.lisp/pitch :f)))))
             (parse-to-lisp-with-context :score "GELATO=d e\n\nclavinet: c/f"))))))
