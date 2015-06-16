(ns alda.repl
  (:require [alda.parser                  :refer (parse-input)]
            [alda.lisp                    :refer :all]
            [alda.sound                   :refer (set-up! tear-down! play!)]
            [alda.sound.midi              :as    midi]
            [alda.repl.commands           :refer (repl-command)]
            [instaparse.core              :as    insta]
            [boot.from.io.aviso.ansi      :refer :all]
            [boot.from.io.aviso.exception :as    pretty]
            [boot.util                    :refer (while-let)]
            [taoensso.timbre              :as    log]
            [clojure.java.io              :as    io]
            [clojure.set                  :as    set])
  (:import  [jline.console ConsoleReader]
            [jline.console.completer Completer]
            [javax.sound.midi MidiSystem]))

(def ascii-art
  (str " █████╗ ██╗     ██████╗  █████╗ " \newline
       "██╔══██╗██║     ██╔══██╗██╔══██╗" \newline
       "███████║██║     ██║  ██║███████║" \newline
       "██╔══██║██║     ██║  ██║██╔══██║" \newline
       "██║  ██║███████╗██████╔╝██║  ██║" \newline
       "╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝"))

(defn text-below-ascii-art [version]
  (str "            v" version \newline
       "         repl session"))

(defn banner [version]
  (str (blue ascii-art)
       \newline
       \newline
       (cyan (text-below-ascii-art version))))

(defn- parse-with-context
  "Parse a string of Alda code starting from a particular level of the tree.
   (e.g. starting from :part, will parse the string as a part, not a whole score)

   With each line of Alda code entered into the REPL, we determine the context of
   evaluation (Are we starting a new part? Are we appending events to an existing
   part? Are we continuing a previous voice? Starting a new voice?) and parse the
   code accordingly."
  [start code]
  (with-redefs [alda.parser/alda-parser
                #((insta/parser (io/resource "alda.bnf")) % :start start)]
    (parse-input code)))

(defn- asap
  "Tranforms a set of events by making the first one start at offset 0,
   maintaining the intervals between the offsets of all the events."
  [events]
  (let [earliest (apply min (map :offset events))]
    (into #{}
      (map #(update-in % [:offset] - earliest) events))))

(defn play-with-context!
  "Plays a partial Alda score, within the context of the current score."
  [context x & [opts]]
  (let [y (case context
            :score x
            :part  (assoc (score-map)
                     :events (asap (event-set
                                     {:start {:offset (->AbsoluteOffset 0)
                                              :events x}}))))]
    (play! y opts)))

(defn start-repl!
  [version & [opts]]
  (println)
  (println (banner version) \newline)
  (let [done?   (atom false)
        context (atom :part)
        reader  (doto (ConsoleReader.) (.setPrompt "> "))]
    (print "Loading MIDI synth... ")
    (set-up! :midi {})
    (println "done." \newline)
    (score*) ; initialize score
    (binding [*out* (.getOutput reader)]
      (while-let [alda-code (when-not @done? (.readLine reader))]
        (try
          (cond
            (re-find #"^\s*$" alda-code)
            :do-nothing

            (re-find #"^:" alda-code)
            (let [[_ cmd rest-of-line] (re-matches #":(\S+)\s*(.*)" alda-code)]
              (repl-command cmd rest-of-line))

            (re-find #"^(quit|exit|bye)" alda-code)
            (do
              (println)
              (println (str "score:" \newline *score-text*))
              (tear-down! :midi {})
              (reset! done? true))

            :else
            (do
              (log/debug "Parsing code...")
              (case @context
                :part
                (let [parsed (parse-with-context :part alda-code)]
                  (log/debug "Done parsing code.")
                  (if (insta/failure? parsed)
                    (log/error "Invalid Alda syntax.")
                    (let [old-score  (score-map)
                          new-score  (do (eval parsed) (score-map))
                          new-events (set/difference
                                       (:events new-score)
                                       (:events old-score))]
                      (midi/load-instruments! new-score)
                      (play-with-context! @context new-events opts)
                      (score-text<< alda-code)))))))
          (catch Throwable e
            (pretty/write-exception *err* e)))))))
