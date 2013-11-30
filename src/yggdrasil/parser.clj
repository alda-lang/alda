(ns yggdrasil.parser
  (:require [instaparse.core :as insta]
  	        [clojure.string :as str]))

(comment
  "Rough structural sketch.

   The entire .yg file is fed to several parsers in succession:

     --> strip-comments: removes comments/barlines

     --> separate-instruments: separates the score into separate instrument-calls
         with their respective music-data.

       -> before going to the next parser, the output of this is transformed by a
          function that consolidates repeated calls to the same instrument into a
          single instrument node with one unbroken string of music data

     --> parse-ygg-code (each instrument's music-data goes through this parser
         separately): parses a string of music data and returns a parse tree of
         music events, in order. Music events are optionally grouped by 'marker'
         nodes. Simultaneous events can be grouped (within a list of music events)
         into 'chord' nodes.

   At this point, will probably hand off the final parse trees (one per instrument
   or instrument group) to yggdrasil.generator, which will hopefully be able to
   create audio segments of each instrument at each marker, assigning time
   markings where each one starts (or perhaps for every single music event), and
   then use the time markings to layer all the different audio segments together to
   create the final audio file.")

(defn strip-comments
  "Strips comments from a yggdrasil score. Returns a new score (in string form),
   devoid of comments."
   [score]
  (->> score
    (insta/parse (insta/parser "grammar/strip-comments.txt"))
    (insta/transform {:score str})))

(defn separate-instruments
  "Takes a complete yggdrasil score and returns a simple parse tree consisting of
   instrument-calls that include their respective music data as an un-parsed string.

   e.g.
   [:score
     [:instrument
       [:instrument-call
         [:name 'cello']]
       [:music-data 'string of ygg code for the cello']]
     [:instrument
       [:instrument-call
         [:name 'trumpet']
         [:name 'trombone']
         [:name 'tuba']
         [:nickname 'brass']]
       [:music-data 'string of ygg code for the brass group']]]"
  [score]
  (->> score
    (insta/parse (insta/parser "grammar/separate-instruments.txt"))
    (insta/transform {:music-data (fn [& chars]
    		                            [:music-data (str/join chars)])})))

(defn assign-instances
  "Takes the output of the separate-instruments parser and transforms it by
   expanding nicknamed groups/instruments and assigning numbered 'instances'
   of each stock instrument, so that it is clear what specific instrument or
   or combination of instruments is assigned to each block of music-data.

  A new instance of an instrument is created whenever a stock instrument is
  called with a nickname, either as part of a group or not.

  If there is already a clarinet, and there is already a cello nicknamed
  'thor', then the following call --

    thor/clarinet 'band':   (with double quotes, though)

  -- will be the same instance of the cello nicknamed Thor, but a NEW instance
  of a clarinet, because the nickname 'band' is being given to this group, and
  the clarinet is a stock instrument.

  By contrast, this call --

    thor/clarinet:

  -- will call the same instances of both thor and the clarinet as before."

  [[_ & instrument-nodes :as parse-tree]]
  (loop [table {}, ; a map of the "names" to the instrument instance(s)
                   ; e.g. {"bill" [{"cello" 1}], "bob" [{"cello" 2}], "trumpet" [{"trumpet" 1}],
                   ;       "brass" [{"trumpet" 1} {"trombone" 1} {"tuba" 1}]}

         nicknames {}, ; exactly the same as table, but only the nicknames are added

         assigned [:score] ; re-constructing the parse-tree node by node

         nodes instrument-nodes]
    (if-not (seq nodes)
      assigned
      (let [[_ [_ & name-nodes] music-data-node :as instrument-node] (first nodes)
            nickname (last (last (filter #(= (first %) :nickname) name-nodes))) ; returns the nickname or nil
            assign (fn [[tag name :as name-node]]
                     (when (= tag :name) ; returns nil for nickname nodes
                       (if-not nickname
                         (table name [{name 1}])
                         (nicknames name [{name
                                           (let [instances (flatten (vals table))]
                                             (if-not (some #(% name) instances)
                                               1
                                               (->> instances
                                                    (filter (% name))
                                                    (map (% name))
                                                    (apply max)
                                                    (inc))))}]))))
            updated-table  ; to do...



(defn nicknames
  "Takes the output of the separate-instruments parser and returns a map of
   nicknames to the instrument(s) to which they refer."
   [parse-tree]
   (letfn [(has-nickname? [instrument]
              (->> (second instrument)
                   (rest)
                   (some #(= (first %) :nickname))))]
     (->> (rest parse-tree)
          (filter has-nickname?)
          (map (fn [instrument]
                 (let [nickname    (->> (second instrument)
                                        (rest)
                                        (last)
                                        (last))
                       instruments (->> (second instrument)
                                        (rest)
                                        (drop-last)
                                        (map last))]
                   {nickname instruments})))
          (apply merge))))

(defn expand-nicknames
  "Takes the output of the separate-instruments parser and transforms it by
   expanding nicknamed groups/instruments, returning a more organized parse tree
   consisting of each instrument-call and its respective music data.
   (This does not consolidate repeated instrument calls.)"
  [parse-tree]
  (let [nicknames (nicknames parse-tree)

        expand (fn expand [x]
                 (cond
                   (not (coll? (first x)))  ; x is a single node
                   (let [[type name] x]
                     (cond
                       (= type :nickname)
                       nil

                       (contains? nicknames name)
                       (expand (map #(vector :name %) (nicknames name)))

                       :else
                       x))

                   :else  ; x is a seq of nodes
                   (->> (map expand x)
                        (remove nil?))))

        expand-and-tag (fn [& name-nodes]
                         (->> (expand name-nodes)
                              (mapcat #(if (coll? (first %)) % [%]))
                              (cons :instrument-call)
                              (apply vector)))]

    (insta/transform {:instrument-call expand-and-tag} parse-tree)))

(defn consolidate-instruments
  "Takes the output of separate-instruments parser -> expand-nicknames and
   consolidates repeated calls to the same instrument, returning a map of
   individual instruments to their respective music data."

  [[_ & instrument-nodes :as parse-tree]]
  	(letfn [(add-music-data
  			      [score ; map
  			      [_ [_ & name-nodes] [_ music-data]]] ; instrument-node
  			      (reduce (fn [m [_ name :as name-node]]
  			      		      (merge-with str m {name (str " " music-data)}))
  			               score
  			               name-nodes))]
  		(reduce add-music-data {} instrument-nodes)))

(def parse-ygg-code
  "Takes a string of music-data (for one instrument) and returns a parse tree of music
   events, including attribute changes and notes."
  (insta/parser "grammar/parse-ygg-code.txt"))






