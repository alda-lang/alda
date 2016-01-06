(ns alda.lisp.model.event
  (:require [alda.lisp.model.global-attribute :refer (apply-global-attributes)]
            [alda.lisp.model.offset           :refer ($current-offset $last-offset)]
            [alda.lisp.model.records          :refer (->AttributeChange)]
            [alda.lisp.score.context          :refer (*events* *instruments*)]))

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

; stashing these here to avoid circular dependencies between namespaces
; ¯\_(ツ)_/¯

(defn set-current-offset
  "Set the offset, in ms, where the next event will occur."
  [instrument offset]
  (let [old-offset ($current-offset instrument)]
    (alter-var-root #'*instruments* assoc-in [instrument :current-offset] offset)
    (apply-global-attributes instrument offset)
    (->AttributeChange instrument :current-offset old-offset offset)))

(defn set-last-offset
  "Set the :last-offset; this value will generally be the value of
   :current-offset before it was last changed. This value is used in
   conjunction with :current-offset to determine whether an event
   occurred within a given window."
  [instrument offset]
  (let [old-offset ($last-offset instrument)]
    (alter-var-root #'*instruments* assoc-in [instrument :last-offset] offset)
    (->AttributeChange instrument :last-offset old-offset offset)))

