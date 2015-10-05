(ns alda.parser.barlines-test
  (:require [clojure.test :refer :all]
            [alda.test-helpers :refer (test-parse)]))

(def alda-code-1
  "violin: c d | e f | g a")

(def parse-tree-1
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

(def alda-lisp-code-1
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

(def alda-code-2
  "marimba: c1|~1|~1~|1|~1~|2.")

(def parse-tree-2
  [:score
    [:part
      [:calls [:name "marimba"]]
      [:note [:pitch "c"]
             [:duration
               [:note-length [:positive-number "1"]]
               [:barline]
               [:note-length [:positive-number "1"]]
               [:barline]
               [:note-length [:positive-number "1"]]
               [:barline]
               [:note-length [:positive-number "1"]]
               [:barline]
               [:note-length [:positive-number "1"]]
               [:barline]
               [:note-length [:positive-number "2"] [:dots "."]]]]]])

(def alda-lisp-code-2
  '(alda.lisp/score
     (alda.lisp/part {:names ["marimba"]}
       (alda.lisp/note 
         (alda.lisp/pitch :c)
         (alda.lisp/duration
           (alda.lisp/note-length 1)
           (alda.lisp/barline)
           (alda.lisp/note-length 1)
           (alda.lisp/barline)
           (alda.lisp/note-length 1)
           (alda.lisp/barline)
           (alda.lisp/note-length 1)
           (alda.lisp/barline)
           (alda.lisp/note-length 1)
           (alda.lisp/barline)
           (alda.lisp/note-length 2 {:dots 1}))))))

(deftest barline-tests
  (testing "barlines are included in the parse tree"
    (is (= [:barline] (test-parse :barline "|" {:tree true})))
    (is (= parse-tree-1 (test-parse :score alda-code-1 {:tree true}))))
  (testing "barlines are included in alda.lisp code (even though they don't do anything)"
    (is (= alda-lisp-code-1 (test-parse :score alda-code-1))))
  (testing "notes can be tied over barlines"
    (is (= parse-tree-2 (test-parse :score alda-code-2 {:tree true})))
    (is (= alda-lisp-code-2 (test-parse :score alda-code-2)))))

