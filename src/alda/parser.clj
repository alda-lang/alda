(ns alda.parser
  (:require [instaparse.core :as insta]
            [clojure.string :as str]
            [clojure.java.io :as io]))

(def ^:private alda-parser
  (insta/parser (io/resource "alda.bnf")))

(defn parse-input
  "Parses a string of Alda code and turns it into Clojure code."
  [alda-code]
  (->> alda-code
       alda-parser
       (insta/transform
         {:name              #(hash-map :name %)
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
          :attribute-change  #(list 'alda.lisp/set-attribute (keyword %1) %2)
          :global-attribute-change
                             #(list 'alda.lisp/global-attribute (keyword %1) %2)
          :voice             #(list* 'alda.lisp/voice %&)
          :voices            #(list* 'alda.lisp/voices %&)
          :marker            #(list 'alda.lisp/marker (:name %))
          :at-marker         #(list 'alda.lisp/at-marker (:name %))
          :calls             (fn [& calls]
                               (let [names    (vec (keep :name calls))
                                     nickname (some :nickname calls)]
                                 (if nickname
                                   {:names names, :nickname nickname}
                                   {:names names})))
          :part              #(list* 'alda.lisp/part %&)
          :score             #(list* 'alda.lisp/score %&)})))
