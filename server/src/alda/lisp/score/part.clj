(ns alda.lisp.score.part
  (:require [djy.char                   :refer (char-range)]
            [clojure.string             :as    str]
            [alda.lisp.attributes       :refer (*initial-attr-vals*)]
            [alda.lisp.events           :refer (apply-global-attributes)]
            [alda.lisp.events.voice     :refer (end-voice-group)]
            [alda.lisp.model.event      :refer (update-score)]
            [alda.lisp.model.instrument :refer (*stock-instruments*)]
            [alda.parser-util           :refer (parse-to-lisp-with-context)]))

(defn- generate-id
  [name]
  (let [rand-char (fn [] (rand-nth (concat (char-range \0 \9)
                                           (char-range \a \z)
                                           (char-range \A \Z))))
        id (apply str (take 5 (repeatedly rand-char)))]
    (str name \- id)))

(defn- new-part
  "Returns a new instance of a stock instrument identified by `stock-inst`,
   with initial values for tempo, current-offset, volume, octave, etc. as
   specified in *initial-attr-vals*.

   Attribute values can be manually overridden via the rest arg `attrs`,
   e.g.: (new-part 'piano' :volume 0.75)

   Throws an error if `stock-inst` is not a valid identifier for a stock Alda
   instrument (see alda.lisp.instrument.*)."
  [stock-inst & attrs]
  (if-let [{:keys [initial-vals]} (*stock-instruments* stock-inst)]
    (merge *initial-attr-vals*
           {:id (generate-id stock-inst)}
           initial-vals
           (apply hash-map attrs))
    (throw (Exception.
             (format "Stock instrument \"%s\" not defined." stock-inst)))))

(defn- add-part
  "Adds a new instrument instance to `score`.

   `stock-inst` must be a valid identifier for a stock Alda instrument.

   When present, `attrs` may override the part's initial attribute values.
   e.g.: (add-part 'bassoon' :tempo 150)"
  [score stock-inst & attrs]
  (let [{:keys [id] :as inst} (apply new-part stock-inst attrs)]
    (assoc-in score [:instruments id] inst)))

(defn- look-up
  "Looks for an instrument named i.e. 'name-XXXXX' in `instruments`.

   Returns either the ID of the first instrument found or nil."
  [instruments name]
  (first (for [[id inst] instruments
               :when (str/starts-with? id (str name \-))]
           (:id inst))))

(defn- determine-current-instruments
  "Given a score and an instrument call (a map with names and nickname keys),
   determines the instrument instances that will become the :current-instruments
   of the score.

   Returns the updated score. In addition to updating :current-instruments, new
   :instruments and :nicknames may be added."
  [{:keys [nicknames instruments] :as score}
   {:keys [names nickname]}]
  (let [new-instruments
        (into {}
          (for [name names
                :when (and (not (contains? nicknames name))
                           (or nickname
                               (not (look-up instruments name))))]
            (let [{:keys [id] :as inst} (new-part name)]
              [id inst])))

        instruments
        (merge instruments new-instruments)

        instances
        (->> (for [name names]
               (get nicknames name (if nickname
                                     (look-up new-instruments name)
                                     (look-up instruments name))))
             flatten
             (remove nil?))

        nicknames
        (if nickname
          (assoc nicknames nickname instances)
          nicknames)]
    (assoc score :nicknames nicknames
                 :instruments instruments
                 :current-instruments (set instances))))

(defn- parse-instrument-call [s]
  (parse-to-lisp-with-context :calls (-> s
                                         (str/replace #":$" "")
                                         (str/replace #"'" "\"")
                                         (str \:))))

(defmethod update-score :part
  [score {:keys [instrument-call events] :as part}]
  (let [instrument-call (cond
                          (map? instrument-call)
                          instrument-call

                          (string? instrument-call)
                          (parse-instrument-call instrument-call)

                          :else
                          (throw (Exception. (str "Invalid instrument call:"
                                                  (pr-str instrument-call)))))
        score (-> score
                  end-voice-group
                  (determine-current-instruments instrument-call))]
    (reduce update-score
            score
            (cons (apply-global-attributes) events))))

