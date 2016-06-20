(ns alda.examples-test
  (:require [clojure.test    :refer :all]
            [clojure.java.io :as    io]
            [alda.parser     :refer (parse-input)]
            [instaparse.core :as    insta]
            [io.aviso.ansi   :refer :all]))

(def example-scores
  ; Ideally, we would be able to dynamically test all .alda files in the
  ; examples/ resource directory, but Clojure resources are all thrown into a
  ; bucket and referenced by filename, and as far as I can tell we can't just
  ; filter "all resource files" by whether or not they end in ".alda", so I
  ; guess we'll just have to manually list them all out here.
  ;
  ; On the plus side, this does allow us to easily test only certain scores by
  ; commenting out the ones we don't want to test.
  '[
   across_the_sea
   awobmolg
   bach_cello_suite_no_1
   debussy_quartet
   entropy
   gau
   hello_world
   key_signature
   multi-poly
   nesting
   overriding-a-global-attribute
   panning
   percussion
   phase
   poly
   printing
   ])

(def longest-score-name-length
  (apply max (map (comp count str) example-scores)))

(defn- spacing
  [score]
  (let [name-length (count (str score))
        spaces      (- longest-score-name-length name-length)]
    (apply str (repeat spaces \space))))

(defmacro time+
  "A modified version of clojure.core/time which measures the time it takes to
   evaluate an expression, returning both the result and the number of
   milliseconds that it took."
  {:added "1.0"}
  [expr]
  `(let [start# (. System (nanoTime))
         ret#   ~expr
         time#  (Math/round (/ (double (- (. System (nanoTime)) start#))
                               1000000.0))]
     [ret# time#]))

(deftest examples-test
  (require '[alda.lisp :refer :all])
  (testing "example scores:"
    (doseq [score example-scores]
      (let [score-text (-> (str score ".alda")
                           io/resource
                           io/file
                           slurp)]
        (testing (format "parsing (as code) %s.alda" score)
          (println \newline (str score ".alda"))
          (printf "   Parsing as code...        ")
          (flush)
          (is
            (try
              (let [[result time-ms] (time+ (parse-input score-text :lisp))]
                (println (green "OK") (format "(%s ms)" time-ms))
                true)
              (catch Exception e
                (println (red "FAIL"))
                (throw e)))))
        (let [parsed-score (atom nil)]
          (testing (format "parsing (as score) %s.alda" score)
            (printf "   Parsing as score...       ")
            (flush)
            (is
              (try
                (let [[result time-ms] (time+ (parse-input score-text :events))]
                  (println (green "OK") (format "(%s ms)" time-ms))
                  (reset! parsed-score result)
                  true)
                (catch Exception e
                  (println (red "FAIL"))
                  (throw e)))))
          (testing (format "realizing parsed score %s.alda" score)
            (printf "   Realizing parsed score... ")
            (flush)
            (is
              (try
                (let [[result time-ms]
                      (time+ (apply (resolve 'alda.lisp/score) @parsed-score))]
                  (println (green "OK") (format "(%s ms)" time-ms))
                  true)
                (catch Exception e
                  (println (red "FAIL"))
                  (throw e))))))))))
