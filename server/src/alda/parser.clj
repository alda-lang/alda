(ns alda.parser
  (:require [instaparse.core          :as insta]
            [clojure.string           :as str]
            [clojure.java.io          :as io]
            [taoensso.timbre          :as log]
            [alda.lisp.attributes     :as attrs]
            [alda.lisp.events         :as evts]
            [alda.lisp.model.duration :as dur]
            [alda.lisp.model.pitch    :as pitch]))

(defn- parser-from-grammars
  "Builds a parser from any number of BNF grammars, concatenated together."
  [& grammars]
  (insta/parser (str/join \newline
                          (map #(slurp (io/resource (str % ".bnf"))) grammars))))

(def comment-parser (parser-from-grammars "comments"
                                          "clojure"))

(def score-parser   (parser-from-grammars "score"
                                          "names"
                                          "ows"))

(def header-parser  (parser-from-grammars "header"
                                          "clojure-cached"
                                          "ows"))

(def part-parser    (parser-from-grammars "events"
                                          "clojure-cached"
                                          "voices"
                                          "event-sequence"
                                          "cram"
                                          "duration"
                                          "barline"
                                          "names"
                                          "numbers"
                                          "ows"))

(defn- read-clj-expr
  "Reads an inline Clojure expression within Alda code.

   This expression will be evaluated within the `boot.user` context, which has
   the vars in `alda.lisp` referred in.

   Returns ready-to-evaluate Clojure code."
  [expr]
  (read-string (str \( (apply str expr) \))))

(def number-transforms
  {:positive-number #(Integer/parseInt %)
   :negative-number #(Integer/parseInt %)
   :voice-number    #(Integer/parseInt %)})

(def name-transforms
  {:name     #(hash-map :name %)
   :nickname #(hash-map :nickname %)})

(def clj-expr-transforms
  {:clj-character #(str \\ %)
   :clj-string    #(str \" (apply str %&) \")
   :clj-expr      #(read-clj-expr %&)})

(defn check-for-failure
  "Determines whether its input is an Instaparse failure, throwing an exception
   if it is. If it isn't, passes it through so we can continue parsing."
  [x]
  (if (insta/failure? x)
    (throw (Exception. (pr-str x)))
    x))

(defn- store-in-cache!
  "Parsing an Alda score is a multi-step process that sometimes has to involve
   looking at the same entity multiple times, but in different contexts as we
   parse the score from inside out. To avoid having to parse the same entity
   more than once, we can cache it the first time, storing it back in the text
   in a uniquely parseable form, a generated id (gensym) surrounded by ⚙
   (Unicode code point 2699), so we can retrieve it from the cache later."
  [cache prefix x]
  (let [id (gensym prefix)]
    (swap! cache assoc id x)
    (str \u2699 id \u2699)))

(defn- get-from-cache
  [cache id]
  (get @cache (symbol id)))

(defn remove-comments
  "Strips comments from a string of Alda code.

   We have to also parse Clojure expressions at this stage in order to avoid
   ambiguity between Alda comments and portions of Clojure expressions. But we
   don't want to have to parse the Clojure expressions again later, so we cache
   them and return them along with the code."
  [[input cache]]
  (let [code (->> input
                  comment-parser
                  check-for-failure
                  (insta/transform
                    (merge clj-expr-transforms
                           {:score
                            #(reduce (fn [acc x]
                                       (if (string? x)
                                         (str acc x)
                                         (str acc (store-in-cache!
                                                    cache "clj-expr" x))))
                                     ""
                                     %&)})))]
    [code cache]))

(defn separate-parts
  "Separates out instrument parts (including subsequent calls to existing
   parts)."
  [[input cache]]
  (->> input
       score-parser
       check-for-failure
       (insta/transform
         (merge name-transforms
                {:calls (fn [& calls]
                          (let [names    (vec (keep :name calls))
                                nickname (some :nickname calls)]
                            (if nickname
                              {:names names, :nickname nickname}
                              {:names names})))}))))

(defn parse-header
  "Parses the (optional) string of non-instrument-specific events that may
   occur at the beginning of an Alda score (e.g. setting variables, global
   attributes, inline Clojure code)."
  [cache header]
  (->> header
       header-parser
       check-for-failure
       (insta/transform
         {:header #(list* %&)
          :clj-expr-cached #(eval (get-from-cache cache %))})))

(defn parse-part
  "Parses the events of a single call to an instrument part."
  [cache part]
  (->> part
       part-parser
       check-for-failure
       (insta/transform
         (merge name-transforms
                number-transforms
                {:events          vector
                 :repeat          (fn [event n]
                                    (evts/times n event))
                 :event-sequence  vector
                 :cram            #(apply evts/cram %&)
                 :voices          #(apply evts/voices %&)
                 :voice           (fn [voice-number & events]
                                    (apply evts/voice
                                           voice-number
                                           events))
                 :voice-zero      #(evts/voice 0 (evts/end-voices))
                 :tie             (constantly :tie)
                 :slur            (constantly :slur)
                 :flat            (constantly :flat)
                 :sharp           (constantly :sharp)
                 :natural         (constantly :natural)
                 :dots            #(hash-map :dots (count %))
                 :note-length     #(apply dur/note-length %&)
                 :milliseconds    #(dur/ms %)
                 :seconds         #(dur/ms (* % 1000))
                 :duration        #(apply dur/duration %&)
                 :pitch           (fn [letter & accidentals]
                                    (apply pitch/pitch
                                           (keyword letter)
                                           accidentals))
                 :note            #(apply evts/note %&)
                 :rest            #(apply evts/pause %&)
                 :chord           #(apply evts/chord %&)
                 :octave-set      #(attrs/octave %)
                 :octave-up       #(attrs/octave :up)
                 :octave-down     #(attrs/octave :down)
                 :marker          #(evts/marker (:name %))
                 :at-marker       #(evts/at-marker (:name %))
                 :barline         #(evts/barline)
                 :clj-expr-cached #(eval (get-from-cache cache %))}))))

(defn parse-input
  "Parses a string of Alda code and turns it into a list of events, which can
   be used to generate a score."
  [alda-code]
  (let [cache (atom {})]
    (->> [alda-code cache]
         remove-comments
         separate-parts
         (insta/transform
           {:score  #(apply vector %&)
            :header #(parse-header cache (apply str %&))
            :part   (fn [names & music-data]
                      (apply evts/part
                             names
                             (parse-part cache (apply str music-data))))}))))

