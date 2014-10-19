# yggdrasil

A music programming language for musicians, implemented in Clojure.

Inspired by other music/audio programming languages such as [PPMCK][ppmck], 
[LilyPond][lilypond] and [ChucK][chuck], the goal of Yggdrasil is to provide a 
powerful and extremely flexible programming language for the musician who wants 
to easily compose and generate music on the fly using only a text editor and a 
compiler. Yggdrasil is designed in a way that equally favors aesthetics, 
flexibility and ease of use, with (eventual) support for the text-based creation 
of all manner of music: classical, popular, chiptune, electroacoustic, and more!

[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu

## Usage

TBA.

## Syntax example

    piano: o3 
    g8 a b > c d e f+ g | a b > c d e f+ g4
    g8 f+ e d c < b a g | f+ e d c < b a g4 
    << g1/>g/>g/b/>d/g

## License

Copyright © 2012-2014 Dave Yarwood

Distributed under the Eclipse Public License either version 1.0 or (at
your option) any later version.
