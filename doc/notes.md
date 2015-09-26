# Notes

Alda's syntax for **notes** is heavily inspired by [MML](http://www.nullsleep.com/treasure/mck_guide). 

## Components

### Octave

Western music theory divides pitches into repeating groups of 12 notes, e.g. (ascending) `c c# d d# e f f# g g# a a# b (next octave) c c# d`, etc. The combination of the letter pitch (e.g. C#) and the octave determines the frequency of the note in Hz. Octave is expressed as a number, typically between 1 and 7, corresponding to [scientific pitch notation](http://en.wikipedia.org/wiki/Scientific_pitch_notation). For example, middle C and A440 are both in octave 4, which is the default octave in Alda. Just like in MML, the octave is set separately from the notes themselves - i.e. it's not "attached to" or "part of" the note, rather, each note looks at the current octave in order to determine its pitch. 

You can set the octave two ways:

`o5` sets the octave to octave 5. Any integer can follow o. 

`<` decreases current octave by 1. `>` increases current octave by 1.

### Duration 

Duration in Alda (as in MML) is expressed in note lengths from standard music notation, in number form. 4 is a quarter note, 2 is a half note, 1 is a whole note, etc. 

Any number of dots can be added to a note duration, which has the same effect as in standard music notation - it essentially adds half of the note duration to the total duration of the note. 

e.g.

    2 = half note, 2 beats
    2. = dotted half note, 3 beats (2 + 1)
    2.. = double-dotted half note, 3-1/2 beats (2 + 1 + 1/2)

Note durations can also be added together using the tie syntax, `~`. (`4~4` = two quarter notes tied together, 2 beats total.)

A special feature of Alda is that you can use non-standard numbers as note durations. For example, 6 is a note that lasts 1/6 of a measure in 4/4 time. In standard notation, there is no such thing as a "sixth note," but this note length would be expressed as one note in a quarter note triplet; in Alda, a "6th note" doesn't necessarily need to be part of a triplet, however, which raises interesting rhythmic possibilities. 

Alda keeps track of both the current octave and the current default note duration as notes are processed sequentially in a score. Each time a note duration is specified, that duration becomes the new default note duration. Each note that follows, when no note duration is specified, will have the default note duration. At the beginning of each instrument part, the default octave is 4 and the default note duration is 4 (i.e. a quarter note, 1 beat). 

### Letter pitch 

A note in Alda is expressed as a letter from a-g, any number of accidentals (optional), and a note duration (also optional).

Flats and sharps will decrease/increase the pitch by one half step, e.g. C + 1/2 step = C#. Flats and sharps are expressed in Alda as - and +, and you can have multiple sharps or multiple flats, or even combine them, if you'd like. e.g. `c++` = C double-sharp = D. 

## Example 

The following is a 1-octave B major scale, ascending and descending, starting in octave 4:

    o4 b4 > c+8 d+ e f+ g+ a+ b4
    a+ g+ f+ e d+ c+ < b2.
