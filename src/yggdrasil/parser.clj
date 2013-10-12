(ns yggdrasil.parser
  (:require [instaparse.core :as insta]))

(comment
  "Rough structural sketch.

   The entire .yg file is fed to several parsers in succession:

     --> strip-comments: removes comments/barlines

     --> separate-instruments: separates the score into separate instrument-calls
         with their respective music-data.

       -> before going to the next parser, the output of this is transformed by a
          function that consolidates repeated calls to the same instrument into a
          single instrument node with one unbroken string of music data

     --> parse-ygg-code (each instrument's music-data goes through this parser
         separately): parses a string of music data and returns a parse tree of 
         music events, in order. Music events are optionally grouped by 'marker'
         nodes. Simultaneous events can be grouped (within a list of music events)
         into 'chord' nodes.

   At this point, will probably hand off the final parse trees (one per instrument
   or instrument group) to yggdrasil.generator, which will hopefully be able to 
   create audio segments of each instrument at each marker, assigning time 
   markings where each one starts (or perhaps for every single music event), and
   then use the time markings to layer all the different audio segments together to
   create the final audio file.")

(def strip-comments
  "Strips comments from a yggdrasil score."
  (insta/parser "grammar/strip-comments.txt"))

(def separate-instruments
  "Takes a complete yggdrasil score and returns a simple parse tree consisting of 
   instrument-calls that include their respective music data as an un-parsed string.

   e.g. 
   [:score
     [:instrument
       [:instrument-call 
         [:name 'cello']]
       [:music-data 'string of ygg code for the cello']]
     [:instrument
       [:instrument-call
         [:name 'trumpet']
         [:name 'trombone']
         [:name 'tuba']
         [:nickname 'brass']]
       [:music-data 'string of ygg code for the brass group']]]"

  (insta/parser "grammar/separate-instruments.txt"))

(def parse-ygg-code
  "Takes a string of music-data (for one instrument) and returns a parse tree of music 
   events, including attribute changes and notes."
  (insta/parser "grammar/parse-ygg-code.txt"))