(ns alda.parser.barlines-test
  (:require [clojure.test :refer :all]
            [alda.parser-util :refer (parse-to-lisp-with-context)]))

(def alda-code-1
  "violin: c d | e f | g a")

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
  (testing "barlines are included in alda.lisp code (even though they don't do anything)"
    (is (= alda-lisp-code-1 (parse-to-lisp-with-context :score alda-code-1))))
  (testing "notes can be tied over barlines"
    (is (= alda-lisp-code-2 (parse-to-lisp-with-context :score alda-code-2)))))

