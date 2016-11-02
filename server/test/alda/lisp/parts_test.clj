(ns alda.lisp.parts-test
  (:require [clojure.test      :refer :all]
            [alda.test-helpers :refer (get-instrument)]
            [alda.lisp         :refer :all]))

(deftest part-tests
  (testing "a part:"
    (let [s       (score (part "piano/trumpet 'trumpiano'"))
          piano   (get-instrument s "piano")
          trumpet (get-instrument s "trumpet")]
      (testing "starts at offset 0"
        (is (zero? (:offset (:current-offset piano))))
        (is (zero? (:offset (:current-offset trumpet)))))
      (testing "starts at the :start marker"
        (is (= :start (:current-marker piano)))
        (is (= :start (:current-marker trumpet))))
      (testing "has the instruments that it has"
        (let [{:keys [current-instruments]} s]
          (is (= 2 (count current-instruments)))
          (is (some #(re-find #"^piano-" %) current-instruments))
          (is (some #(re-find #"^trumpet-" %) current-instruments))))
      (testing "sets a nickname if applicable"
        (is (contains? (:nicknames s) "trumpiano"))
        (let [trumpiano (get (:nicknames s) "trumpiano")]
          (is (= (count trumpiano) 2))
          (is (some #(re-find #"^piano-" %) trumpiano))
          (is (some #(re-find #"^trumpet-" %) trumpiano))))
      (let [s              (continue s
                             (note (pitch :d)
                                   (duration (note-length 2 {:dots 1}))))
            piano          (get-instrument s "piano")
            trumpet        (get-instrument s "trumpet")
            piano-offset   (:current-offset piano)
            trumpet-offset (:current-offset trumpet)]
        (testing "instruments from a group can be accessed using the dot
                  operator"
          (let [s     (continue s
                        (part "trumpiano.piano"))
                piano (get-instrument s "piano")]
            (is (= 1 (count (:current-instruments s))))
            (is (re-find #"^piano-" (first (:current-instruments s))))
            (is (= piano-offset (:current-offset piano)))
            (let [s       (continue s
                            (part "trumpiano.trumpet"))
                  trumpet (get-instrument s "trumpet")]
              (is (= 1 (count (:current-instruments s))))
              (is (re-find #"^trumpet-" (first (:current-instruments s))))
              (is (= trumpet-offset (:current-offset trumpet)))))))
      (testing "referencing a nickname"
        (let [{:keys [current-instruments] :as s}
              (continue s
                (part "bassoon"
                  (note (pitch :c))
                  (note (pitch :d))
                  (note (pitch :e))
                  (note (pitch :f))
                  (note (pitch :g)))
                (part "trumpiano"))]
          (is (= 2 (count current-instruments)))
          (is (some #(re-find #"^piano-" %) current-instruments))
          (is (some #(re-find #"^trumpet-" %) current-instruments)))))))

(deftest instance-assignment-tests
  (testing "one name, without nickname:"
    (testing "if the name refers to a stock instrument,"
      (testing "and we don't have that instrument yet in the score,"
        (let [s (score
                  (part "piano"))]
          (testing "a new instance of that stock instrument is created"
            (is (= 1 (count (:current-instruments s))))
            (let [inst (first (:current-instruments s))
                  stock-instrument (-> s :instruments (get inst) :stock)]
              (is (= "midi-acoustic-grand-piano" stock-instrument))
              (testing "and subsequent calls to that name will refer to that instance"
                (let [s (continue s
                          (part "piano"))]
                  (is (= inst (first (:current-instruments s))))))))))
      (testing "and we already have one of that instrument"
        (testing "and it's a named instance,"
          (let [s (score
                    (part "piano 'foo'"))]
            (testing "an ambiguity error is thrown"
              (is (thrown-with-msg?
                    Exception
                    #"Ambiguous instrument reference"
                    (continue s
                      (part "piano")))))))
        (testing "and it doesn't have a name,"
          (let [s     (score
                        (part "piano"))
                piano (first (:current-instruments s))]
            (testing "the name refers to the existing instance of that instrument"
              (let [s (continue s
                        (part "piano"))]
                (is (= piano (first (:current-instruments s))))))))))
    (testing "if the name refers to a previously named instance,"
      (let [s       (score
                      (part "piano 'piano-1'"))
            piano-1 (first (:current-instruments s))
            s       (continue s
                      (part "piano-1"))]
        (testing "the nickname then refers to that instrument"
          (is (= 1 (count (:current-instruments s))))
          (is (= piano-1 (first (:current-instruments s)))))))
    (testing "if the name refers to a previously named group,"
      (let [s       (score
                      (part "piano/guitar 'foos'"))
            foos    (:current-instruments s)
            s       (continue s
                      (part "foos"))]
        (testing "the nickname then refers to that group"
          (is (= 2 (count (:current-instruments s))))
          (is (= foos (:current-instruments s))))))
    (testing "if the name does not refer to a stock instrument or an existing
              instance or group,"
      (testing "an unrecognized instrument error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Unrecognized instrument"
              (score (part "quizzledyblarf")))))))
  (testing "one name + nickname:"
    (testing "if the name is a named instance,"
      (testing "an informative error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Can't assign alias \"bar\" to existing instance \"foo\"."
              (score
                (part "piano 'foo'")
                (part "foo 'bar'"))))))
    (testing "if the name is otherwise not a stock instrument,"
      (testing "an unrecognized instrument error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Unrecognized instrument"
              (score (part "quizzledyblarf 'norman'"))))))
    (testing "if the nickname was already assigned to another instance,"
      (testing "an informative error is thrown"
        (is (thrown-with-msg?
              Exception
              #"The alias \"foo\" has already been assigned to another instrument/group."
              (score
                (part "piano 'foo'")
                (part "clarinet 'foo'"))))))
    (testing "if the nickname was already assigned to a group,"
      (testing "an informative error is thrown"
        (is (thrown-with-msg?
              Exception
              #"The alias \"foo\" has already been assigned to another instrument/group."
              (score
                (part "piano/accordion 'foo'")
                (part "clarinet 'foo'"))))))
    (testing "if there is already an unnamed instance of that instrument,"
      (testing "an ambiguity error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Ambiguous instrument reference"
              (score
                (part "piano")
                (part "piano 'foo'"))))))
    (testing "creates a new named instrument instance"
      (let [s        (score
                       (part "piano 'foo'"))
            piano-id (-> s :current-instruments first)
            piano    (-> s :instruments (get piano-id))]
        (testing "with the correct stock instrument"
          (is (= "midi-acoustic-grand-piano" (:stock piano))))
        (testing "and subsequent use of the nickname refers to that instance"
          (let [s (continue s
                    (part "foo"))]
            (is (= piano-id (-> s :current-instruments first))))))))
  (testing "multiple names, without nickname:"
    (testing "if all the names refer to the same named instance,"
      (let [s (score
                (part "piano 'foo'"))]
        (testing "a grouping error is thrown" ; because it makes no sense
          (is (thrown-with-msg?
                Exception
                #"Invalid instrument grouping"
                (continue s
                  (part "foo/foo")))))))
    (testing "if all the names are the same and they refer to a stock
              instrument,"
      ; the next time that instrument is called, it would be an ambiguous
      ; reference
      (testing "a grouping error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Invalid instrument grouping"
              (score
                (part "piano/piano"))))))
    (testing "if all the names refer to previously named instruments,"
      (let [s   (score
                  (part "piano 'foo'"))
            foo (-> s :current-instruments first)
            s   (continue s
                  (part "clarinet 'bar'"))
            bar (-> s :current-instruments first)]
        (testing "it refers to those instruments as a group"
          (let [s (continue s
                    (part "foo/bar"))]
            (is (= #{foo bar} (:current-instruments s)))))))
    (testing "if all the names refer to stock instruments,"
      (testing "it refers to those instruments as a group, creating new
                instances for any instruments that don't exist yet in the
                score"
        (let [s (score
                  (part "piano/clarinet"))]
          (is (= 2 (count (:current-instruments s))))
          (is (some (fn [[_ {:keys [stock]}]]
                      (= "midi-acoustic-grand-piano" stock))
                    (:instruments s)))
          (is (some (fn [[_ {:keys [stock]}]]
                      (= "midi-clarinet" stock))
                    (:instruments s))))))
    (testing "if the names are a mix of stock instruments and named instances,"
      (let [s (score
                (part "piano 'foo'"))]
        ; nicknames should be used for creating new instances or grouping
        ; existing ones, not both
        (testing "a grouping error is thrown"
          (is (thrown-with-msg?
                Exception
                #"Invalid instrument grouping"
                (continue s
                  (part "foo/trumpet"))))))))
  (testing "multiple names + nickname:"
    (testing "if all the names refer to the same named instance,"
      (let [s (score
                (part "piano 'foo'"))]
        (testing "a grouping error is thrown" ; because it makes no sense
          (is (thrown-with-msg?
                Exception
                #"Invalid instrument grouping"
                (continue s
                          (part "foo/foo 'bar'")))))))
    (testing "if all the names are the same and they refer to a stock
              instrument,"
      ; if you want to call pianos.piano subsequently to refer to one of the
      ; pianos, it won't be clear which one you mean
      ;
      ; the moral of the story is, if you want a group containing two of the
      ; same instrument, you have to create the two named instances first and
      ; then group them
      (testing "a grouping error is thrown"
        (is (thrown-with-msg?
              Exception
              #"Invalid instrument grouping"
              (score
                (part "piano/piano 'pianos'"))))))
    (testing "if all the names refer to previously named instruments,"
      (let [s   (score
                  (part "piano 'foo'"))
            foo (-> s :current-instruments first)
            s   (continue s
                  (part "clarinet 'bar'"))
            bar (-> s :current-instruments first)]
        (testing "it refers to those instruments as a group"
          (let [s (continue s
                    (part "foo/bar 'baz'"))]
            (is (= #{foo bar} (:current-instruments s)))
            (testing "and you can now use the nickname to refer to that group"
              (let [s (continue s
                        (part "baz"))]
                (is (= #{foo bar} (:current-instruments s)))))
            (testing "and you can continue to use the individual names to refer
                      to each instance"
              (let [s (continue s
                        (part "foo"))]
                (is (= #{foo} (:current-instruments s))))
              (let [s (continue s
                        (part "bar"))]
                (is (= #{bar} (:current-instruments s)))))
            (testing "and you can now use the group-member operator to refer to
                      each instance individually"
              (let [s (continue s
                        (part "baz.foo"))]
                (is (= #{foo} (:current-instruments s))))
              (let [s (continue s
                        (part "baz.bar"))]
                (is (= #{bar} (:current-instruments s)))))))))
    (testing "if all the names refer to stock instruments,"
      ; regardless of whether there are existing instances of those
      ; instruments
      (let [s     (score
                    (part "piano"))
            s     (continue s
                    (part "banjo"))
            s     (continue s
                    (part "piano/banjo/tuba 'floop'"))
            piano (->> s
                       :current-instruments
                       (filter #(.startsWith % "piano-"))
                       first)
            banjo (->> s
                       :current-instruments
                       (filter #(.startsWith % "banjo-"))
                       first)
            tuba  (->> s
                       :current-instruments
                       (filter #(.startsWith % "tuba-"))
                       first)]
        (testing "it creates new instances for each group member"
          (is (= 5 (count (:instruments s))))
          (is (some (fn [[_ {:keys [stock]}]]
                      (= "midi-acoustic-grand-piano" stock))
                    (:instruments s)))
          (is (some (fn [[_ {:keys [stock]}]]
                      (= "midi-banjo" stock))
                    (:instruments s)))
          (is (some (fn [[_ {:keys [stock]}]]
                      (= "midi-tuba" stock))
                    (:instruments s))))
        (testing "the group name now refers to the instances as a group"
          (is (= 3 (count (:current-instruments s)))))
        (testing "you can now use the group-member operator to refer to each
                  instance individually"
          (let [s (continue s
                    (part "floop.piano"))]
            (is (= #{piano} (:current-instruments s))))
          (let [s (continue s
                    (part "floop.banjo"))]
            (is (= #{banjo} (:current-instruments s))))
          (let [s (continue s
                    (part "floop.tuba"))]
            (is (= #{tuba} (:current-instruments s)))))))
    (testing "if the names are a mix of stock instruments and named instances,"
      (let [s (score
                (part "piano 'foo'"))]
        ; nicknames should be used for creating new instances or grouping
        ; existing ones, not both
        (testing "a grouping error is thrown"
          (is (thrown-with-msg?
                Exception
                #"Invalid instrument grouping"
                (continue s
                  (part "foo/trumpet 'engelberthumperdinck'"))))))))
  (testing "groups within groups:"
    (testing "a group consisting of two groups"
      (let [s        (score
                       (part "clarinet/flute 'woodwinds'")
                       (part "trumpet/trombone 'brass'"))
            clarinet (get-instrument s "clarinet")
            flute    (get-instrument s "flute")
            trumpet  (get-instrument s "trumpet")
            trombone (get-instrument s "trombone")
            s        (continue s
                       (part "woodwinds/brass 'wwab'"))]
        (testing "refers to the set of instruments in both groups"
          (is (= 4 (count (:current-instruments s))))
          (doseq [expected ["midi-clarinet" "midi-flute"
                            "midi-trumpet" "midi-trombone"]]
            (is (some (fn [[_ {:keys [stock]}]]
                        (= expected stock))
                      (:instruments s))))
          (is (= (set (map :id #{clarinet flute trumpet trombone}))
                 (:current-instruments s))))))
    (testing "a group consisting of two overlapping groups"
      (let [s (score
                (part "clarinet 'foo'")
                (part "flute 'bar'")
                (part "trumpet 'baz'")
                (part "foo/bar 'group1'")
                (part "foo/baz 'group2'")
                (part "group1/group2 'groups1and2'"))]
        (testing "refers to the set of instruments in both groups"
          (is (= 3 (count (:current-instruments s))))
          (doseq [expected ["midi-clarinet" "midi-flute" "midi-trumpet"]]
            (some (fn [[_ {:keys [stock]}]]
                    (= expected stock))
                  (:instruments s)))))))
  (testing "a group containing the member of another group"
    (let [s (score
              (part "piano/guitar 'foo'")
              (part "clarinet 'bob'")
              (part "bob/foo.piano 'bar'"))]
      (testing "contains the correct instance"
        (is (= 3 (count (:instruments s))))
        (is (= 2 (count (:current-instruments s))))
        (is (some #(.startsWith % "clarinet-") (:current-instruments s)))
        (is (some #(.startsWith % "piano-") (:current-instruments s))))
      (testing "creates the right aliases"
        (let [s (continue s
                          (part "foo.piano"))]
          (is (= 1 (count (:current-instruments s))))
          (is (some #(.startsWith % "piano-") (:current-instruments s))))
        (let [s (continue s
                  (part "bar.foo.piano"))]
          (is (= 1 (count (:current-instruments s))))
          (is (some #(.startsWith % "piano-") (:current-instruments s))))))))
