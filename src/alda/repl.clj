(ns alda.repl
  (:require [alda.parser                  :refer (parse-input)]
            [alda.lisp                    :refer :all]
            [alda.sound                   :refer (set-up! tear-down! play!)]
            [alda.sound.midi              :as    midi]
            [alda.repl.commands           :refer (repl-command)]
            [alda.now                     :as    now]
            [instaparse.core              :as    insta]
            [boot.from.io.aviso.ansi      :refer :all]
            [boot.from.io.aviso.exception :as    pretty]
            [boot.util                    :refer (while-let)]
            [taoensso.timbre              :as    log]
            [clojure.java.io              :as    io]
            [clojure.string               :as    str]
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

(defn parse-with-start-rule
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

(defn parse-with-context
  "Determine the appropriate context to parse a line of code from the Alda 
   REPL, then parse it within that context.
   
   context is an atom representing the existing context -- this function 
   will reset its value to the determined context."
  [context alda-code]
  (letfn [(try-ctxs [[ctx & ctxs]]
            (if ctx
              (let [parsed (parse-with-start-rule ctx alda-code)]
                (if (insta/failure? parsed)
                  (try-ctxs ctxs)
                  (do (reset! context ctx) parsed)))
              (log/error "Invalid Alda syntax.")))]
    (try-ctxs [:music-data :part :score])))

(defn set-repl-prompt!
  "Sets the REPL prompt to give the user clues about the current context."
  [rdr]
  (let [abbrevs (for [inst *current-instruments*]
                  (if-let [nickname (first (for [[k v] *nicknames*
                                                 :when (= v (list inst))]
                                             k))]
                    (->> (re-seq #"(\w)\w*" nickname)
                         (map second)
                         (apply str))
                    (->> (re-seq #"(\w)\w*-" inst)
                         (map second)
                         (apply str))))
        prompt  (str (str/join "/" abbrevs) "> ")]
    (.setPrompt rdr prompt)))

(defn start-repl! [version]
  (println)
  (println (banner version) \newline)
  (let [done?   (atom false)
        context (atom :part)
        reader  (doto (ConsoleReader.) (.setPrompt "> "))]
    (print "Loading MIDI synth... ")
    (set-up! :midi)
    (println "done." \newline)
    (score*) ; initialize a new score
    (binding [*out* (.getOutput reader)]
      (set-repl-prompt! reader)
      (while-let [alda-code (when-not @done? (.readLine reader))]
        (try
          (cond
            (re-find #"^\s*$" alda-code)
            :do-nothing

            (re-find #"^:?(quit|exit|bye)" alda-code)
            (do
              (println)
              (println (str "score:" \newline *score-text*))
              (tear-down! :midi)
              (reset! done? true))

            (re-find #"^:" alda-code)
            (let [[_ cmd rest-of-line] (re-matches #":(\S+)\s*(.*)" alda-code)]
              (repl-command cmd rest-of-line))

            :else
            (let [_          (log/debug "Parsing code...")
                  parsed     (parse-with-context context alda-code)
                  _          (log/debug "Done parsing code.")]
              (now/play! (eval (case @context
                                 :music-data (cons 'do parsed)
                                 parsed)))
              (when parsed (score-text<< alda-code))))
          (set-repl-prompt! reader)
          (catch Throwable e
            (pretty/write-exception *err* e)))))))
