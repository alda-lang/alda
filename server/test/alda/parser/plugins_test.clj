(ns alda.parser.plugins-test
  (:require [clojure.test :refer :all]
            [alda.parser :refer (parse-input)]))

(deftest chord-plugin-tests
  (testing "shorthand chords are replaced with notes"
    (is (= (parse-input "piano: Em")
           '(alda.lisp/score
             (alda.lisp/part
              {:names ["piano"]}
              (alda.lisp/octave 2)
              (alda.lisp/chord
               (alda.lisp/note (alda.lisp/pitch :e))
               (alda.lisp/octave :up)
               (alda.lisp/note (alda.lisp/pitch :b))
               (alda.lisp/note (alda.lisp/pitch :e))
               (alda.lisp/note (alda.lisp/pitch :g))
               (alda.lisp/note (alda.lisp/pitch :b))
               (alda.lisp/octave :up)
               (alda.lisp/note (alda.lisp/pitch :e)))
              (alda.lisp/octave :down)
              (alda.lisp/octave :down)))))))
