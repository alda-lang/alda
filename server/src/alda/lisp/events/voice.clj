(ns alda.lisp.events.voice)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.events.voice...")

(defn voice
  "Returns a list of the events, executing them in the process."
  [& events]
  (remove #(not (contains? #{alda.lisp.Note alda.lisp.Chord} (type %)))
          (flatten events)))

(defmacro voices*
  "Voices are chronological sequences of events that each start at the same
   time. The resulting :current-offset is at the end of the voice that finishes
   last."
  [instrument & voices]
  (let [voice-snapshots (gensym "voice-snapshots")
        voice-events    (gensym "voice-events")]
    (list* 'let ['start-snapshot `(snapshot ~instrument)
                 voice-snapshots (list 'atom {})
                 voice-events    (list 'atom {})]
           (concat
             (for [[_ num# & events# :as voice#] voices]
               (list 'let ['voice-name (keyword (str \v num#))]
                     `(load-snapshot ~instrument
                                     (get (deref ~voice-snapshots)
                                          ~'voice-name ~'start-snapshot))
                      (list 'swap! voice-events
                        (list 'fn ['m]
                          (list 'merge-with 'concat 'm
                                {'voice-name
                                 (list 'vec
                                       (list 'map 'first (vec events#)))})))
                     `(swap! ~voice-snapshots assoc ~'voice-name
                                              (snapshot ~instrument))))
             [`(let [last-voice#
                     (apply (partial max-key #(-> % :current-offset :offset))
                            (vals (deref ~voice-snapshots)))]
                 (load-snapshot ~instrument last-voice#))
              `(deref ~voice-events)]))))

(defmacro voices
  [& args]
  `(let [voices-per-instrument#
         (for [instrument# *current-instruments*]
           (binding [*current-instruments* #{instrument#}]
             (voices* instrument# ~@args)))]
     (apply merge-with (comp vec concat) voices-per-instrument#)))
