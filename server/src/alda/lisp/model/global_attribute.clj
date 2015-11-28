(ns alda.lisp.model.global-attribute)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.model.global-attribute...")

(comment
  "*global-attributes* is a map of offsets to the global attribute changes that
   occur (for all instruments) at each offset.

   apply-global-attributes is a fn that looks at the window between an
   instrument's :last-offset and :current-offset and sees if any global attribute
   changes have occurred in that window, and if so, applies them for the
   instrument. This fn is called by set-current-offset each time an instrument's
   :current-offset is changed, so that global attributes can be applied at the
   new offset.

   Note: global attributes only work moving forward. That is to say, a global
   attribute change will only affect the current part and any others that
   follow it in the score.")

(declare ^:dynamic *global-attributes*)

(defrecord GlobalAttribute [offset attr val])

(defn global-attribute
  [attr val]
  (set-attribute attr val)
  (if-let [offset (instruments-all-at-same-offset)]
    (do
      (alter-var-root #'*global-attributes* update-in [(absolute-offset offset)]
                                            (fnil conj []) [attr val])
      (log/debug "Set global attribute" attr val "at offset"
                 (str (int (absolute-offset offset)) \.))
      (GlobalAttribute. offset attr val))
    (throw (Exception. (str "Can't set global attribute " attr " to " val
                            " - offset unclear. There are multiple instruments "
                            "active with different time offsets.")))))

(defn global-attributes
  "Convenience fn for setting multiple global attributes at once."
  [& attrs]
  (doall
    (for [[attr val] (partition 2 attrs)]
      (global-attribute attr val))))

(defn apply-global-attributes
  [instrument now]
  (let [now            (absolute-offset now)
        last-offset    (absolute-offset ($last-offset instrument))
        new-global-attrs (->> *global-attributes*
                              (filter (fn [[offset attrs]]
                                        (<= last-offset
                                            offset
                                            now)))
                              (mapcat (fn [[offset attrs]] attrs)))]
    (binding [*current-instruments* #{instrument}]
      (doall
        (for [[attr val] new-global-attrs]
          (set-attribute attr val))))))
