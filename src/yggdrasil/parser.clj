(ns yggdrasil.parser
  (:require [instaparse.core :as insta]
            [clojure.string :as str]
            [clojure.java.io :as io]))

(declare assign-instances consolidate-instruments)

(def ygg-parser
  (insta/parser (io/resource "grammar/yggdrasil.bnf")))

(defn parse-input
  "Parses a string of ygg code, determines which instrument instances are assigned to
   each instrument call, and returns a map of each instrument instance to its own parse
   tree of music data."
  [ygg-code]
  (->> ygg-code
       ygg-parser
       (insta/transform {:score assign-instances})
       (insta/transform {:score consolidate-instruments})))

(defn- assign-instances
  "Reconstructs the parse tree by going through the score linearly and deducing what
   specific instrument 'instances' are referred to by each instrument call. In the
   resulting parse tree, each :instrument node has a :tracks node that lists these
   instrument instances, e.g. [:tracks {'guitar' 1} {'bass' 1}], rather than an
   :instrument-call node.

   A new numbered instance of an instrument is created whenever a stock instrument is
   called with a nickname, either as part of a group or not.

   e.g. If there is already a clarinet 1 and there is already a cello 1 nicknamed
   'thor', then this -- thor/clarinet 'band': -- will refer to the same instance of
   cello (cello 1, 'thor'), but a new clarinet instance (clarinet 2) because a nickname,
   'band' is being given to this group, and 'clarinet' refers to the stock instrument,
   not any particular named instance of clarinet.

   On the other hand, -- thor/clarinet: -- in the same scenario would refer to cello 1
   and clarinet 1, the same instances that were already in use."
  [& instrument-nodes]
  (loop [table {}, ; a map of the "names" to the instrument instance(s)
                   ; e.g. {"bill" [{"cello" 1}], "bob" [{"cello" 2}], "trumpet" [{"trumpet" 1}],
                   ;       "brass" [{"trumpet" 1} {"trombone" 1} {"tuba" 1}]}

         nicknames {}, ; exactly the same as table, but only the nicknames are added

         score [:score] ; re-constructing the parse-tree node by node

         nodes instrument-nodes]
    (if-not (seq nodes)
      score
      (let [[_ [_ & name-nodes] music-data-node] (first nodes)

            nickname ; returns the nickname of the node or nil
            (last (last (filter #(= (first %) :nickname) name-nodes)))

            assign
            (fn [[tag name :as name-node]]
              "Assigns an instance to each name-node. Returns nil for nickname nodes."
              (when (= tag :name)
                (if-not nickname
                  (table name [{name 1}])
                  (nicknames name [{name
                                    (let [instances (flatten (vals table))]
                                      (if (some #(% name) instances)
                                        (->> (filter #(% name) instances)
                                             (map #(% name))
                                             (apply max)
                                             (inc))
                                        1))}]))))

            instances
            (vec (remove nil? (map assign name-nodes)))

            names
            (map second name-nodes)

            updated-table
            (merge table
                   (zipmap names
                           (conj instances
                                 (vec (flatten instances)))))

            updated-nicknames
            (if nickname
              (assoc nicknames nickname (vec (flatten instances)))
              nicknames)

            updated-score
            (conj score
                  [:instrument
                    (apply vector :tracks (flatten instances))
                    music-data-node])]

        (recur updated-table updated-nicknames updated-score (rest nodes))))))

(defn- consolidate-instruments
  "Returns a map of instrument instances to their consolidated music data."
  [& instrument-nodes]
  	(letfn [(add-music-data [score ; map
  			                    [_ [_ & instances] [_ & events]]] ; instrument node
  			      (reduce (fn [m instance] (merge-with concat m {instance events}))
  			               score
  			               instances))]
  		(reduce add-music-data {} instrument-nodes)))

(comment
  "Each instrument now has its own vector of music events, representing everything that instrument
   will do for the duration of the score.

   To do:

     - Implement a way to embed time markers of some sort among the musical events,
       so that they end up synchronized between the musical instruments. Each instrument
       by default starts its musical events at '0', i.e. the beginning, but the composer
       will be able to specify where a musical event will fall by using markers.

         Special cases:
           * Chords: each note in the chord starts at the same time mark. The next event
                     after the chord will be assigned that time mark + the duration of the
                     longest note/rest in the chord.
           * Voices: voices work just like chords. Each voice in the voices grouping
                     starts at the same time mark. The next event after the voices grouping
                     (whether via explicit V0: or switching back from another instrument)
                     will start whenever the last voice has finished its music data...
                     i.e. the 'largest' time marking out of all the voices.

                     Implications: the composer will need to either make use of named markers
                     or make sure, when switching from instrument to instrument, that he be
                     aware that the end of the longest held voice is where the next event
                     will come in after switching back to that instrument. I don't really
                     foresee this as being an inconvience. Markers will be easy to use.

     -  At this point, will probably hand off the final parse trees (one per instrument)
        to yggdrasil.sound_generator, which will hopefully be able to create audio segments
        of each instrument at each time marking, and then use the time markings to layer all
        the different audio segments together to create the final audio file.")

(parse-input (slurp "test/yggdrasil/awobmolg.yg"))
