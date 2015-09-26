# Chords

A **chord** is a collection of [notes](notes.md) which all start at the same [[offset]], i.e. they all start at the exact same time. In Alda, a chord is expressed as notes with slashes in between them: `c/e/g`

It's acceptable to have octave changes in between the notes of a chord, which allows for chords spanning multiple octaves: `c/g/>c/e/g`

The notes in a chord can all be different lengths, in which case, the next note event after the chord will happen **after the shortest note in the chord**. This makes it easy to have chords with shifting tones, e.g.: `c1~1/>c/<e4 f g f e1` (also, note that, just like with sequential notes, each note duration becomes the default for all notes that follow - both C notes in this chord are 2 whole notes long). 

Alda also allows you to use [rests](rests.md) in a chord. Because the next note event after a chord will start after the shortest note/rest in the chord, this can be useful for writing melodies entwined with chords, e.g. `c1/e/g/r4 b e g`
