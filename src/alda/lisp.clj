(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace."
  (:require [alda.lisp.attributes]
            [alda.lisp.events]))

;;; TODO: make this all happen encapsulated in a pod ;;;

(defn global-attribute
  "TODO: (tentative idea) create an agent and add a watch for any changes to
   *current-offset* -- on every change, check *last-offset* and *current-offset*
   and if the global attribute change happens in that range, make it happen.

   Note: this only works moving forward. That is to say, a global attribute
         change event will only affect the current part and any others that
         follow it in the score."
  [attr val])

;; everything below this line is old and overly complicated -- TODO: rewrite

(comment

  (defn global-attribute
    "Stores a global attribute change event in *global-attribute-events*.
    Upon evaluation of the score (after all instrument instances are recorded),
    the attribute is changed for every instrument at that time marking."
    [attribute value]
    (alter-var-root #'*add-event*
                    (fn [f]
                      (fn [{:keys [last-offset current-offset instrument] :as context}
                           & event-map]
                        (let [context
                              (if (<= last-offset *current-offset* current-offset)
                                (attribute-change context instrument attribute value)
                                context)]
                          (apply f context event-map))))))

;;; score-builder utils ;;;

(defn- add-globals
  "If initial global attributes are set, add them to the first instrument's
   music-data."
  [global-attrs instruments]
  (letfn [(with-global-attrs [[tag & names-and-data :as instrument]]
            (let [[data & names] (reverse names-and-data)]
              `(~tag
                ~@names
                (music-data ~global-attrs ~@(rest data)))))]
    (if global-attrs
      (cons (with-global-attrs (first instruments)) (rest instruments))
      instruments)))

(defn part [& args]
  (identity args))

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
