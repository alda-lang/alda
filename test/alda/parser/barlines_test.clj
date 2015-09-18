(ns alda.parser.barlines-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(def alda-code
  "violin: c d | e f | g a")

(def parse-tree
  [:score
   [:part
    [:calls [:name "violin"]]
    [:note [:pitch "c"]]
    [:note [:pitch "d"]]
    [:barline]
    [:note [:pitch "e"]]
    [:note [:pitch "f"]]
    [:barline]
    [:note [:pitch "g"]]
    [:note [:pitch "a"]]]])

(def alda-lisp-code
  '(alda.lisp/score
     (alda.lisp/part {:names ["violin"]}
       (alda.lisp/note (alda.lisp/pitch :c))
       (alda.lisp/note (alda.lisp/pitch :d))
       (alda.lisp/barline)
       (alda.lisp/note (alda.lisp/pitch :e))
       (alda.lisp/note (alda.lisp/pitch :f))
       (alda.lisp/barline)
       (alda.lisp/note (alda.lisp/pitch :g))
       (alda.lisp/note (alda.lisp/pitch :a)))))

(deftest barline-tests
  (testing "bar-lines are included in the parse tree"
    (is (= [:barline] (test-parse :barline "|" {:tree true})))
    (is (= parse-tree (test-parse :score alda-code {:tree true}))))
  (testing "bar-lines are included in alda.lisp code (even though they don't do anything)"
    (is (= alda-lisp-code (test-parse :score alda-code)))))

