(ns alda.lisp.model.event)
(in-ns 'alda.lisp)

(declare ^:dynamic *events*)

(defn add-event
  [instrument event]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil conj []) event)
    event))

(defn add-events
  [instrument events]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil into []) events)
    events))
