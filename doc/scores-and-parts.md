# Scores and Parts

The top level of a piece of music written in Alda is the **score**. A score consists of any number of instrument **parts**, each of which have their own [note](notes.md) events, which occur simultaneously. 

Alda is designed to be flexible about how a score is organized. For the same piece of music, a composer can choose to write each instrument part's notes from beginning to end before moving on to the next instrument part (something like ex. 1), or alternate between the instrument parts, organizing the score by section rather than by part (ex. 2).

**Ex. 1**

    trumpet: 
    o4 c d e f g a b > c d e f g a b > c

    trombone:
    o3 e f g a b > c d e f g a b > c d e

**Ex. 2**

    trumpet: o4 c d e f g a b > c
    trombone: o3 e f g a b > c d e

    trumpet: d e f g a b > c
    trombone: f g a b > c d e

Under the hood, Alda processes a score sequentially, keeping track of information about each instrument, including the instrument's volume, tempo, duration, offset, and octave. The nice thing about this is that, when switching to another instrument and then switching back, you don't have to worry about changing the volume, tempo, octave, etc. back to what they were when you were last using the instrument - Alda keeps track of that for you. 

#### Instrument groups

It's possible in Alda to use the same note events for multiple instruments at once by grouping them, e.g.:

    trumpet/trombone: c d e f g f e d c

Keep in mind that Alda is still keeping track of each instrument's volume, tempo, octave, offset, etc. separately, which means it is up to the composer to ensure that the instruments are playing in sync, if that's what the composer wants. In ex. 3, the trumpet plays some repeated D notes at the start of the score, then an ascending D minor scale; the trombone also plays the D minor scale, however it starts at the beginning of the score, so it beats the trumpet to the punch. Ex. 4 shows a way to remedy this situation, in cases where the really want both instruments playing in unison. Ex. 5 shows another way to achieve the same effect using [markers](markers.md). 

**Ex. 3**

    trumpet: d d d d d d d d
    
    # not in sync, trombone starts earlier
    trumpet/trombone: d e f g a b- > c d

**Ex. 4**

    trumpet: d d d d d d d d
    trombone: r1~1 # (rest for 8 beats) 

    # in sync 
    trumpet/trombone: d e f g a b- > c d

**Ex. 5**

    trumpet: 
    d d d d d d d d %scaleTime

    trumpet/trombone: 
    @scaleTime d e f g a b- > c d

Alda chooses not to force instrument parts to sync up when used as a group in order to allow composers the freedom to experiment with multiple instruments playing the same notes in different ways. For example, you could give the instruments different tempos and/or note durations and have them play the same notes:

**Ex. 6**

    violin: (tempo 100)
    viola: (tempo 112)
    cello: (tempo 124)

    violin/viola/cello: e f g e f g e f g e f g e f g
