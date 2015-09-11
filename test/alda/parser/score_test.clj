(ns alda.parser.score-test
  (:require [clojure.test :refer :all]
            [alda.parser :refer (parse-input)]
            [alda.test-helpers :refer (test-parse)]))

(deftest score-tests
  (is (= (parse-input "theremin: c d e")
         '(alda.lisp/score
            (alda.lisp/part {:names ["theremin"]}
              (alda.lisp/note (alda.lisp/pitch :c))
              (alda.lisp/note (alda.lisp/pitch :d))
              (alda.lisp/note (alda.lisp/pitch :e))))))
  (is (= (parse-input "trumpet/trombone/tuba \"brass\": f+1")
         '(alda.lisp/score
            (alda.lisp/part {:names ["trumpet" "trombone" "tuba"]
                             :nickname "brass"}
              (alda.lisp/note (alda.lisp/pitch :f :sharp)
                              (alda.lisp/duration (alda.lisp/note-length 1)))))))
  (is (= (parse-input "guitar: e
                       bass: e")
         '(alda.lisp/score
            (alda.lisp/part {:names ["guitar"]}
              (alda.lisp/note (alda.lisp/pitch :e)))
            (alda.lisp/part {:names ["bass"]}
              (alda.lisp/note (alda.lisp/pitch :e)))))))
