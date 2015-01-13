(ns code-generator)

(def letters (seq "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
(def numbers (seq "1234567890"))
(def possible-characters (concat letters numbers))

(def master-code-list (ref []))

(defn generate-code []
  (let
    [new-combo (fn [] (take 3 (repeatedly #(rand-nth possible-characters))))
     type-count (fn [type]
                  (fn [code]
                    (->> code
                         (filter (fn [ch] (some #{ch} type)))
                         (count))))
     letter-count (type-count letters)
     number-count (type-count numbers)]
    (loop [code (new-combo)]
      (if (and
            (not-any? #{code} @master-code-list)
            (< (letter-count code) 3)
            (< (number-count code) 3))
          code
          (recur (new-combo))))))

(dosync
  (dotimes [n 20]
    (alter master-code-list conj (generate-code))))

(prn @master-code-list)
