(ns alda.lisp.model.event)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.event...")

(declare ^:dynamic *events*)

(defn add-event
  [instrument event]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil conj []) event)))

(defn add-events
  [instrument events]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil into []) events)))
