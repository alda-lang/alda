(ns alda.parser
  (:require [instaparse.core :as insta]
            [clojure.string  :as str]
            [clojure.java.io :as io]
            [alda.util       :as util]
            [backtick        :refer (defquote)]))

; sets log level to TIMBRE_LEVEL (if set) or :warn
(util/set-timbre-level!)

(declare alda-parser parse-input)
(require '[alda.lisp :as lisp])

(def ^:private alda-parser
  (insta/parser (io/resource "alda.bnf")))

(defn parse-tree
  "Returns the intermediate parse tree resulting from parsing a string of Alda
   code."
  [alda-code]
  (alda-parser alda-code))

(defquote alda-lisp-quote
  #(if-let [{:keys [ns name]} (meta (ns-resolve 'alda.lisp %))]
     (symbol (str ns "/" name))
     %))

(defn read-to-alda-lisp
  [code]
  (load-string (format "(alda.parser/alda-lisp-quote %s)" code)))

; wrapping :clj-string nodes in this record so that :clj-expr nodes can
; differentiate between strings containing commas/semicolons and whitespace 
; commas/semicolons
(defrecord StringHack [value])

(defn- read-clj-expr
  "Reads an inline Clojure expression within Alda code.
   
   Special rules:
     - each comma or semicolon will split an S-expression, e.g.:
         (volume 50, quant 50) => (do (volume 50) (quant 50))
       - unless:
         - it's a character literal (prefixed by a backslash)
         - it's inside of a string
         - it's inside of [square brackets]
         - it's inside of {curly braces}
     - symbols will first try to be resolved within the context of alda.lisp,
       then if that fails, the current run-time namespace

   Returns ready-to-evaluate Clojure code."
  [exprs]
  (let [add-if-appropriate (fn [results i current-str]
                             (if (empty? current-str)
                               [results i current-str]
                               [(update results i conj current-str) i ""]))
        conj* (fn [xs x] (concat xs (list x)))
        exprs (loop [results [[]]
                     i 0
                     current-str ""
                     [expr & more] exprs]
                (let [[results i current-str]
                      (condp = (type expr)
                        String (if (#{"," ";"} expr)
                                 [(update results i conj current-str)
                                  (inc i)
                                  ""]
                                 [results i (str current-str expr)])
                        StringHack (update-in (add-if-appropriate results
                                                                  i
                                                                  current-str)
                                              [0 i]
                                              conj* (:value expr))
                        (update-in (add-if-appropriate results
                                                       i
                                                       current-str)
                                   [0 i]
                                   conj* expr))]
                  (if (seq more)
                    (recur results i current-str more)
                    (first (add-if-appropriate results i current-str)))))
        exprs (map #(str \( (apply str %) \)) exprs)]
    (if (> (count exprs) 1)
      (cons 'do (map read-to-alda-lisp exprs))
      (read-to-alda-lisp (first exprs)))))

(defn- read-stringhacks
  [coll]
  (map #(if (= StringHack (type %))
          (:value %)
          %)
       coll))

(defn- read-clj-coll
  [coll format-str]
  (->> (read-stringhacks coll)
       (apply str)
       (format format-str)
       read-string))

(defn parse-input
  "Parses a string of Alda code and turns it into Clojure code."
  [alda-code]
  (->> alda-code
       alda-parser
       (insta/transform
         {:clj-character     #(StringHack. (str \\ %))
          :clj-string        #(StringHack. (str \" (apply str %&) \"))
          :clj-list          #(read-clj-coll %& "(%s)")
          :clj-vector        #(read-clj-coll %& "[%s]")
          :clj-map           #(read-clj-coll %& "{%s}")
          :clj-set           #(read-clj-coll %& "#{%s}")
          :clj-expr          #(read-clj-expr %&)
          :name              #(hash-map :name %)
          :nickname          #(hash-map :nickname %)
          :number            #(Integer/parseInt %)
          :voice-number      #(Integer/parseInt %)
          :tie               (constantly :tie)
          :slur              (constantly :slur)
          :flat              (constantly :flat)
          :sharp             (constantly :sharp)
          :dots              #(hash-map :dots (count %))
          :note-length       #(list* 'alda.lisp/note-length %&)
          :duration          #(list* 'alda.lisp/duration %&)
          :pitch             (fn [s]
                               (list* 'alda.lisp/pitch
                                      (keyword (str (first s)))
                                      (map #(case %
                                              \- :flat
                                              \+ :sharp)
                                           (rest s))))
          :note              #(list* 'alda.lisp/note %&)
          :rest              #(list* 'alda.lisp/pause %&)
          :chord             #(list* 'alda.lisp/chord %&)
          :octave-set        #(list 'alda.lisp/octave %)
          :octave-up         #(list 'alda.lisp/octave :up)
          :octave-down       #(list 'alda.lisp/octave :down)
          :voice             #(list* 'alda.lisp/voice %&)
          :voices            #(list* 'alda.lisp/voices %&)
          :marker            #(list 'alda.lisp/marker (:name %))
          :at-marker         #(list 'alda.lisp/at-marker (:name %))
          :barline           #(list 'alda.lisp/barline)
          :calls             (fn [& calls]
                               (let [names    (vec (keep :name calls))
                                     nickname (some :nickname calls)]
                                 (if nickname
                                   {:names names, :nickname nickname}
                                   {:names names})))
          :part              #(list* 'alda.lisp/part %&)
          :score             #(list* 'alda.lisp/score %&)})))
