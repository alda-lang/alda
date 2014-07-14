(ns yggdrasil.parser
  (:require [instaparse.core :as insta]
            [clojure.string :as str]))

(comment
  "STEP ONE:
   The .yg file is fed through strip-comments to remove comments and barlines,
   then the result of that is fed to separate instruments, which returns a simple
   parse-tree with this kind of structure:

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
       [:music-data 'string of ygg code for the brass group']]]    ")

(defn- strip-comments [score]
  "Takes a string of ygg code, returns a string with comments/barlines removed."
  (->> score
    (insta/parse (insta/parser "grammar/strip-comments.txt"))
    (insta/transform {:score str})))

(defn- separate-instruments [score]
  "Takes a string of de-commented ygg code, returns a parse tree."
  (->> score
    (insta/parse (insta/parser "grammar/separate-instruments.txt"))))

(comment
  "STEP TWO:
   Before going to the next parser, the result of Step 1 is transformed by two
   functions. The first function, assign-instances, goes through the score linearly
   and deduces what specific instrument 'instances' are referred to by each
   instrument call. The resulting parse-tree is different in that each :instrument
   node has a :tracks node instead of an :instrument-call node. The :tracks node
   indicates which instrument instances are being called,
   e.g. [:tracks {'guitar' 1} {'bass' 1}].

   (NOTE: A new numbered instance of an instrument is created whenever a stock
          instrument is called with a nickname, either as part of a group or not.

          If there is already a clarinet(1) and there is already a cello(1)
          nicknamed 'thor', then calling -- thor/clarinet 'band': -- will refer to
          the same instance of cello (1, thor), but a *new* instance of clarinet(2)
          because a nickname, 'band', is being given to the group, and 'clarinet'
          refers to the stock instrument, not any particular named instance of
          clarinet.

          On the other hand, calling -- thor/clarinet: -- in the same scenario will
          refer to cello(1) and clarinet(1), the same instances that were already
          'in play,' so to speak.)

   The second function consolidates repeated calls to the same instrument instance
   and returns a simple map of each instance to its consolidated music data.")

(defn- assign-instances [[_ & instrument-nodes :as parse-tree]]
  (loop [table {}, ; a map of the "names" to the instrument instance(s)
                   ; e.g. {"bill" [{"cello" 1}], "bob" [{"cello" 2}], "trumpet" [{"trumpet" 1}],
                   ;       "brass" [{"trumpet" 1} {"trombone" 1} {"tuba" 1}]}

         nicknames {}, ; exactly the same as table, but only the nicknames are added

         score [:score] ; re-constructing the parse-tree node by node

         nodes instrument-nodes]
    (if-not (seq nodes)
      score
      (let [[_ [_ & name-nodes] music-data-node :as instrument-node] (first nodes)

            nickname
            ; returns the nickname of the node or nil
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

(defn- consolidate-instruments [[_ & instrument-nodes :as parse-tree]]
  "Returns a map of instrument instances to their consolidated music-data."
  	(letfn [(add-music-data
  			      [score ; map
  			      [_ [_ & instances] [_ music-data]]] ; instrument-node
  			      (reduce (fn [m instance]
  			      		      (merge-with #(str % " " %2) m {instance music-data}))
  			               score
  			               instances))]
  		(reduce add-music-data {} instrument-nodes)))

(comment
  "STEP THREE:
  Each instance's music-data goes through the parse-ygg-code parser separately.
  The music-data is a string of ygg-code representing all of what that instrument
  will do for the duration of the score. The parser parses the music-data and
  returns a parse-tree of music events, in linear order. Music events are
  optionally grouped by 'marker' nodes. Simultaneous events can be grouped (within
  a list of linear music events) into 'chord' nodes.

   At this point, will probably hand off the final parse trees (one per instrument
   or instrument group) to yggdrasil.generator, which will hopefully be able to
   create audio segments of each instrument at each marker, assigning time
   markings where each one starts (or perhaps for every single music event), and
   then use the time markings to layer all the different audio segments together to
   create the final audio file.")

(def parse-ygg-code
  (insta/parser "grammar/parse-ygg-code.txt"))


; example -- it's working so far!
(->> (slurp "test/yggdrasil/awobmolg.yg")
     (strip-comments)
     (separate-instruments)
     (assign-instances)
     (consolidate-instruments))



