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
          :octave-change     #(list 'alda.lisp/octave (case %
                                                         "<" :down
                                                         ">" :up
                                                         %))
          :attribute-change  #(list 'alda.lisp/set-attribute (keyword %1) %2)
          :global-attribute-change
                             #(list 'alda-lisp/global-attribute (keyword %1) %2)
          :voice             #(list* 'alda.lisp/voice %&)
          :voices            #(list* 'alda.lisp/voices %&)
          :marker            #(list 'alda.lisp/marker (:name %))
          :at-marker         #(list 'alda.lisp/at-marker (:name %))
          :calls             (fn [& calls]
                               (let [names    (keep :name calls)
                                     nickname (some :nickname calls)]
                                 {:names names, :nickname nickname}))
          :music-data        #(list* 'alda.lisp/music-data %&)
          :part              #(list* 'alda.lisp/part %&)
          :score             #(list* 'alda.lisp/score %&)})))

(comment
  "To do:

     - Implement a way to embed time markers of some sort among the musical
       events, so that they end up synchronized between the musical instruments.
       Each instrument by default starts its musical events at '0', i.e. the
       beginning, but the composer will be able to specify where a musical event
       will fall by using markers.

         Special cases:
           * Chords: each note in the chord starts at the same time mark. The
                     next event after the chord will be assigned that time mark
                     + the duration of the shortest note/rest in the chord.
           * Voices: voices work just like chords. Each voice in the voices
                     grouping starts at the same time mark. The next event after
                     the voices grouping (whether via explicit V0: or switching
                     back from another instrument) will start whenever the last
                     voice has finished its music data... i.e. the 'largest' time
                     marking out of all the voices.

     -  At this point, will probably hand off the final parse trees (one per
        instrument) to alda.sound_generator, which will hopefully be able to
        create audio segments of each instrument at each time marking, and then
        use the time markings to layer all the different audio segments together
        to create the final audio file.")
