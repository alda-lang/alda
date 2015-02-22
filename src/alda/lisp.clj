(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace."
  (:require [alda.lisp.attributes]
            [alda.lisp.events]))

;;; TODO: make this all happen encapsulated in a pod ;;;

(defn init-instrument
  "Initializes an instrument instance with values for tempo, current-offset,
   volume, octave, etc. Adds it to *instruments* and also returns it."
  ([name & attrs]
    (let [attr-map (apply hash-map attrs)
          instrument (merge *initial-attr-values* {:name name} attr-map)]
      (alter-var-root #'*instruments* assoc-in [name] instrument)
      instrument)))

(defmacro part
  "Determines the current instrument(s) and executes the events."
  [{:keys [names nickname]} & events]
  `(do ~@events))

; test setup
(init-instrument "piano")
(alter-var-root (var *current-instruments*) conj "piano")


;; everything below this line is old and overly complicated -- TODO: rewrite

(comment

;;; score-builder utils ;;;

(defn build-parts
  "Walks through a variable number of instrument calls, building a score
   from scratch. Handles initial global attributes, if present."
  [components]
  (let [[global-attrs & instrument-calls] (if (= (ffirst components)
                                                 'alda.lisp/global-attributes)
                                            components
                                            (cons nil components))
        instrument-calls (add-globals global-attrs instrument-calls)]
    `(for [[[name# number#] music-data#] (-> {:parts {} :name-table {} :nickname-table {}}
                                             ((apply comp ~instrument-calls))
                                             :parts)]
       (part name# number# music-data#))))

(defn- assign-new
  "Assigns a new instance of x, given the table of existing instances."
  [x name-table]
  (let [existing-numbers (for [[name num] (apply concat (vals name-table))
                               :when (= name x)]
                           num)]
    (if (seq existing-numbers)
      [[x (inc (apply max existing-numbers))]]
      [[x 1]])))

;;; score-builder ;;;

(defmacro alda-score
  "Returns a new version of the code involving consolidated instrument parts
   instead of overlapping instrument calls."
  [& components]
  (let [parts (build-parts components)]
  `(score ~parts)))

(defn instrument-call
  "Returns a function which, given the context of the score-builder in
   progress, adds the music data to the appropriate instrument part(s)."
  [& components]
  (let [[[_ & music-data] & names-and-nicks] (reverse components)
        names (for [{:keys [name]} names-and-nicks :when name] name)
        nickname (some :nickname names-and-nicks)]
    (fn update-data [working-data]
      (reduce (fn [{:keys [parts name-table nickname-table]} name]
                (let [name-table (or (and (map? name-table) name-table) {})
                      instance
                      (if nickname
                        (nickname-table name (assign-new name name-table))
                        (name-table name [[name 1]]))]
                  {:parts (merge-with concat parts {instance music-data})
                   :name-table (assoc name-table name instance)
                   :nickname-table (if nickname
                                     (merge-with concat nickname-table
                                                        {nickname instance})
                                     nickname-table)}))
              working-data
              names))))

)
