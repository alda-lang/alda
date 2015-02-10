(ns alda.lisp
  "alda.parser transforms Alda code into Clojure code, which can then be
   evaluated with the help of this namespace.")

;;; utils ;;;

(def ^:private intervals
  {"c" 0, "d" 2, "e" 4, "f" 5, "g" 7, "a" 9, "b" 11})

(defn- midi-note
  "Given a letter and an octave, returns the MIDI note number.
   e.g. 'c', 4  =>  60"
  [letter octave]
  (+ (intervals letter) (* octave 12) 12))

(defn- midi->hz
  "Converts a MIDI note number to the note's frequency in Hz."
  [note]
  (* 440.0 (Math/pow 2.0 (/ (- note 69.0) 12.0))))

;;; TODO: make this all happen encapsulated in a pod ;;;

(def ^:dynamic *instrument* nil)
(def ^:dynamic *tempo* 120)
(def ^:dynamic *duration* 1) ; default note length in beats
(def ^:dynamic *octave* 4)
(def ^:dynamic *quant* 0.9)
(def ^:dynamic *volume* 1.0)
(def ^:dynamic *panning* 0.5)
(def ^:dynamic *global-attributes* {})
(def ^:dynamic *events* {:start {:offset 0, :events []}})
(def ^:dynamic *current-marker* :start)
(def ^:dynamic *last-offset* 0)
(def ^:dynamic *current-offset* 0)

(defn octave
  "Sets the current octave."
  [arg]
  (alter-var-root #'*octave* (case arg
                               "<" dec
                               ">" inc
                               (constantly arg))))

(defn tempo
  "Sets the current tempo."
  [bpm]
  (alter-var-root #'*tempo* (constantly bpm)))

(defn volume
  "Sets the current volume (0-100)."
  [vol]
  (alter-var-root #'*volume* (constantly (/ vol 100.0))))

(defn panning
  "Sets the current panning (L 0 <- C 50 -> R 100)."
  [pan]
  (alter-var-root #'*panning* (constantly (/ pan 100.0))))

(defn quant
  "Sets the current quantization (0-100)."
  [q]
  (alter-var-root #'*quant* (constantly (/ q 100.0))))

(defn set-attribute
  "Top-level fn for setting attributes."
  [attr num]
  (let [attrs {"volume"       volume
               "panning"      panning
               "pan"          panning
               "quantization" quant
               "quantize"     quant
               "quant"        quant
               "tempo"        tempo
               "octave"       octave}
        f (attrs attr)]
    (f num)))

(defn note-length
  "Converts a number, representing a note type, e.g. 4 = quarter, 8 = eighth,
   into a number of beats. Handles dots if present."
  ([number]
    (/ 4 number))
  ([number {:keys [dots]}]
    (let [value (/ 4 number)]
      (loop [total value, factor 1/2, dots dots]
        (if (pos? dots)
          (recur (+ total (* value factor)) (* factor 1/2) (dec dots))
          total)))))

(defn duration
  "Combines a variable number of tied note-lengths into one.

   A slur may appear as the final argument of a duration, making the current
   note legato (effectively slurring it into the next).

   Returns a map containing the duration in ms (within the context of the
   current tempo) and whether or not the note is slurred."
  [& components]
  (let [[note-lengths slurred] (if (= (last components) :slur)
                                 (conj [(drop-last components)] true)
                                 (conj [components] false))
        beats (apply + note-lengths)]
    (alter-var-root #'*duration* (constantly beats))
    {:duration (* beats (/ 60000 *tempo*))
     :slurred slurred}))

(defn pitch
  "Determines the frequency in Hz, within the context of the current
   octave."
  [letter & accidentals]
  (let [midi-note (reduce (fn [number accidental]
                            (case accidental
                              :flat  (dec number)
                              :sharp (inc number)))
                          (midi-note letter *octave*)
                          accidentals)]
    (midi->hz midi-note)))

(defn note
  ([pitch]
   {:pre [(number? pitch)]}
    (note pitch (duration *duration*) false))
  ([pitch arg2] ; arg2 could be a duration or :slur
    (cond
      (map? arg2)    (note pitch arg2 false)
      (= :slur arg2) (note pitch (duration *duration*) true)))
  ([pitch {:keys [duration slurred]} slur?]
    (binding [*quant* (if (or slur? slurred)
                        1.0
                        *quant*)]
      (let [event {:offset *current-offset*
                   :instrument *instrument*
                   :volume *volume*
                   :pitch pitch
                   :duration (* duration *quant*)}]
        (alter-var-root #'*events* update-in [*current-marker* :events] conj event)
        (alter-var-root #'*last-offset* (constantly *current-offset*))
        (alter-var-root #'*current-offset* (partial + duration))
        event))))

(defn pause
  ([]
    (pause (duration *duration*)))
  ([{:keys [duration]}]
    (alter-var-root #'*last-offset* (constantly *current-offset*))
    (alter-var-root #'*current-offset* (partial + duration))))

(defmacro chord
  "Chords contain notes/rests that all start at the same time/offset.
   The resulting *current-offset* is that of the shortest note/rest in
   the chord."
  [& events]
  (let [start *current-offset*
        num-of-events (count (filter #(= (first %) 'note) events))
        note-durations (gensym)]
    (list* 'do
           `(def ~note-durations (atom []))
           (concat
             (interleave
               (repeat `(alter-var-root (var *current-offset*) (constantly ~start)))
               events
               (repeat `(swap! ~note-durations conj *current-offset*)))
             [`(alter-var-root (var *current-offset*)
                 (constantly (apply min
                               (remove #(= % ~start)
                                 (deref ~note-durations)))))]
             [`(take-last ~num-of-events
                          (get-in *events* [*current-marker* :events]))]))))

;; everything below this line is old and overly complicated --
;; TODO: rewrite using the dynamic var system instead of callbacks

;; (for global-attribute, I'm thinking, make it add a time marking -> events
;; key/value pair to the *global-attributes* map, then maybe write a macro
;; called "obey-global-attributes" that executes the events in its body by
;; checking if any global-attributes have occurred at a given time marking
;; and applying them, before each event. The function below kind of does this
;; by way of modifying an add-event function that would be called when adding
;; any event, but it seems like doing it that way would have to involve callbacks,
;; which is what we're moving away from (not to mention the fact that it would make
;; the behavior of add-event unpredictable).

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
