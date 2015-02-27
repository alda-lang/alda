(ns alda.lisp.model.event)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.event...")

(def ^:dynamic *events* {:start {:offset (AbsoluteOffset. 0), :events []}})

(defn add-event
  [instrument event]
  (let [marker (-> (*instruments* instrument) :current-marker)]
    (alter-var-root #'*events* update-in [marker :events] (fnil conj []) event)))
