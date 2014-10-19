# yggdrasil

A music programming language for musicians, implemented in Clojure.

Inspired by other music/audio programming languages such as [PPMCK][ppmck], 
[LilyPond][lilypond] and [ChucK][chuck], Yggdrasil aims to provide a 
powerful and flexible programming language for the musician who wants to easily
compose and generate music on the fly using only a text editor and a compiler. 
Yggdrasil is designed in a way that equally favors aesthetics, flexibility and 
ease of use, with (eventual) support for the text-based creation of all manner 
of music: classical, popular, chiptune, electroacoustic, and more!

## Etymology

In Norse mythology, [Yggdrasil](http://en.wikipedia.org/wiki/Yggdrasil) is the 
central, venerated, immense tree around which the universe revolves and on which 
the mythical "nine worlds" exist. Using this as a metaphor, you could say that 
*sound* or *music* is an immense, holy tree that supports numerous different 
"worlds" (different genres, tonalities, paradigms, etc.) that are all part of 
the same tree. 

There is a plethora of different music software out there, but most of these 
programs tend to specialize or "reside" in at most one or two different "worlds" 
-- [FamiTracker][famitracker] and MCK are specifically for the creation of NES 
music; [puredata][pd], [Csound][csound] and ChucK are mostly useful for 
experimental electronic music; Lilypond, [Rosegarden][rosegarden], and 
[MuseScore][musescore] can be used for more than just classical music, but 
their standard notation interface suggests a preference for classical music; 
[Guitar Pro][guitarpro] is specific to the creation of guitar-based music. Why 
not have one piece of software that can serve as the Great Tree that supports 
all of these existing worlds? 

[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu
[famitracker]: http://famitracker.com 
[pd]: http://puredata.info
[csound]: http://www.csounds.com
[rosegarden]: http://www.rosegardenmusic.com
[musescore]: http://musescore.org
[guitarpro]: http://www.guitar-pro.com

## Syntax example

    piano: o3 
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4 
    << g1/>g/>g/b/>d/g

## License

Copyright © 2012-2014 Dave Yarwood

Distributed under the Eclipse Public License version 1.0.
