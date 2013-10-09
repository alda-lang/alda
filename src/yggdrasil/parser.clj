(ns yggdrasil.parser
	(:require [instaparse.core :as insta]))

(comment 
	"Rough structural sketch. The idea is that an entire .yg file will be first
	 fed into the strip-comments parser, which does the obvious, then that is 
	 fed to the separate-instruments parser, then the resulting parse tree will
	 undergo a transformation (via insta/transform) that consolidates repeated
	 calls to the same instrument into a single instrument with one unbroken
	 string of music data. 

	 At that point we'll have a parse tree with one or more instrument nodes,
	 one for each instrument in the score, and each of those nodes will contain
	 an un-parsed string of music data. The music data for each instrument can
	 then be separately fed to the ygg-parse parser, which will turn the music
	 data into a detailed parse tree of chords, notes, attribute changes, etc. 

	 I am undecided at this point which way to proceed. I have two options:

		1) Consider instaparse done with its job, and hand off the final parse
		   tree to yggdrasil.generator, which will employ Overtone to turn the
		   parse tree into music.

		2) Make instaparse do more of the work via insta/transform, 
		   transforming the completed parse tree into a string of Overtone
		   code. Then all yggdrasil.generator has to do is run the code
		   within the parameters selected by the user at runtime (i.e. play
		   the score vs. save it as a wav file, optional start/end points, 
		   etc.)

	 I'll make a decision on this when I know more about what Instaparse and Overtone
	 can and can't do.")

(def strip-comments
	"Strips comments from a yggdrasil score."
	(insta/parser "grammar/strip-comments.txt"))

(def separate-instruments
	"Takes a complete yggdrasil score and returns a simple parse tree consisting
	 of instrument-calls that include their respective music data as an un-parsed
	string.

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
	"Takes a string of music-data (for one instrument) and returns a parse tree of
	 music events, including attribute changes and notes."
	(insta/parser "grammar/parse-ygg-code.txt"))