# Alda

## A music programming language for musicians

Inspired by other music/audio programming languages such as [PPMCK][ppmck],
[LilyPond][lilypond] and [ChucK][chuck], Alda aims to be a
powerful and flexible programming language for the musician who wants to easily
compose and generate music on the fly, using naught but a text editor.
Alda is designed in a way that equally favors aesthetics, flexibility and
ease of use, with (eventual) support for the text-based creation of all manner
of music: classical, popular, chiptune, electroacoustic, and more!

## Etymology

Alda was originally named after [Yggdrasil][yggdrasil], the venerated tree of Norse legend which held aloft the mythical nine worlds. I thought it to be a fitting name, imagining the realm of sound/music to be an immense tree bearing numerous branches which could represent genres, tonalities, paradigms, etc.

By incredible coincidence, [the company I work for][adzerk] uses Norse mythology as a theme for naming our software projects, and there will soon be a major one in production called Yggdrasil. To be completely honest, I was never totally happy with Yggdrasil as the name for my music programming language (it's a mouthful), so I took this as an opportunity to rename it. *Alda* is [Quenya][quenya] for "tree."

There is a plethora of music software out there, but most of these
programs tend to specialize or "reside" in at most one or two different realms
-- [FamiTracker][famitracker] and MCK are specifically for the creation of NES
music; [puredata][pd], [Csound][csound] and ChucK are mostly useful for
experimental electronic music; Lilypond, [Rosegarden][rosegarden], and
[MuseScore][musescore] can be used for more than just classical music, but
their standard notation interface suggests a preference for classical music;
[Guitar Pro][guitarpro] is targeted toward the creation of guitar-based music. Why
not have one piece of software that can serve as the Great Tree that supports
all of these existing worlds?

[ppmck]: http://ppmck.wikidot.com/what-is-ppmck
[lilypond]: http://www.lilypond.org
[chuck]: http://chuck.cs.princeton.edu
[yggdrasil]: http://en.wikipedia.org/wiki/Yggdrasil
[adzerk]: http://www.adzerk.com
[quenya]: http://en.wikipedia.org/wiki/Quenya
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

Copyright © 2012-2015 Dave Yarwood

Distributed under the Eclipse Public License version 1.0.
