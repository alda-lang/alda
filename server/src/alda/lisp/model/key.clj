(ns alda.lisp.model.key)

(defn- remove-first [x coll]
  (concat (take-while (partial not= x) coll)
          (rest (drop-while (partial not= x) coll))))

(defn- shift-key
  "Raises or lowers a key signature by one semitone, depending on the values of
   `this` and `that`. (Implementation for sharpen-key and flatten-key.)"
  [this that]
  (fn [key-sig]
    (apply conj {}
              (map (fn [letter]
                     (let [accidentals (key-sig letter)]
                       (cond
                         (nil? accidentals)
                         [letter [this]]

                         (= accidentals [that])
                         nil

                         (contains? (set accidentals) that)
                         [letter (remove-first that accidentals)]
                         
                         :else
                         [letter (conj accidentals this)])))
                   (map (comp keyword str) "abcdefg")))))

(def ^:private sharpen-key
  "Raises a key signature by one semitone.
   
   All notes in the key signature that were flat become natural, notes that
   were natural become sharp, notes that were sharp become double-sharp, etc.
   
   e.g. Db major -> D major
   {:b [:flat] :e [:flat] :a [:flat] :d [:flat] :g [:flat]}
   becomes
   {:f [:sharp] :c [:sharp]}"
  (shift-key :sharp :flat))

(def ^:private flatten-key
  "Lowers a key signature by one semitone.
   
   All notes in the key signature that were sharp become natural, notes that
   were natural become flat, notes that were flat become double-flat, etc.
   
   e.g. F# major to F major
   {:f [:sharp] :c [:sharp] :g [:sharp] :d [:sharp] :a [:sharp] :e [:sharp]}
   becomes
   {:b [:flat]}"
  (shift-key :flat :sharp))

(defn partial-circle-of-fifths
  [scale-type]
  (zipmap (map (comp keyword str) "fcgdaeb")
          (case scale-type
            :major (range -1 6)
            :minor (range -4 3))))

(defn get-key-signature
  ([scale-type letter]
    (into {}
      (let [n          (get (partial-circle-of-fifths scale-type) letter)
            letters    (take (Math/abs n) 
                             (map (comp keyword str) 
                                  (if (pos? n) "fcgdaeb" "beadgcf")))
            accidental (if (pos? n) :sharp :flat)]
        (map (fn [ltr] [ltr [accidental]]) letters))))
  ([scale-type letter accidentals]
    (reduce (fn [key-sig accidental]
              (case accidental
                :flat (flatten-key key-sig)
                :sharp (sharpen-key key-sig)))
            (get-key-signature scale-type letter)
            accidentals)))
