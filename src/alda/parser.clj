(ns alda.parser
  (:require [instaparse.core :as insta]
            [clojure.string  :as str]
            [clojure.java.io :as io]
            [alda.util       :as util]))

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

(defn- read-clj-expr
  "Reads an inline Clojure expression within Alda code.

   This expression will be evaluated within the `boot.user` context, which has
   the vars in `alda.lisp` referred in.

   Returns ready-to-evaluate Clojure code."
  [expr]
  (read-string (str \( (apply str expr) \))))

(defn parse-input
  "Parses a string of Alda code and turns it into Clojure code."
  [alda-code]
  (->> alda-code
       alda-parser
       (insta/transform
         {:clj-character     #(str \\ %)
          :clj-string        #(str \" (apply str %&) \")
          :clj-expr          #(read-clj-expr %&)
          :code-block        #(list 'alda.lisp/code-block (second %))
          :code-block-no-ws  #(list :code-block-no-ws
                                    (apply str 
                                           (map (fn [x]
                                                  (if (list? x)
                                                    (str \[ (second x) \])
                                                    x))
                                                %&)))
          :cram              #(list* 'alda.lisp/cram %&)
          :name              #(hash-map :name %)
          :nickname          #(hash-map :nickname %)
          :number            identity
          :positive-number   #(Integer/parseInt %)
          :negative-number   #(Integer/parseInt %)
          :voice-number      #(Integer/parseInt %)
          :tie               (constantly :tie)
          :slur              (constantly :slur)
          :flat              (constantly :flat)
          :sharp             (constantly :sharp)
          :natural           (constantly :natural)
          :dots              #(hash-map :dots (count %))
          :note-length       #(list* 'alda.lisp/note-length %&)
          :duration          #(list* 'alda.lisp/duration %&)
          :pitch             (fn [letter & accidentals] 
                               (list* 'alda.lisp/pitch (keyword letter) accidentals))
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

