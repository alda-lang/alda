# Tempo

The **tempo** of a score describes how fast or slow notes are played. This
value, typically expressed in beats per minute (BPM), is used in combination
with the length of a note (e.g. quarter, half, whole) to determine how long the
note should last in milliseconds.

## The `tempo` attribute

In Alda, the simplest way to specify a tempo is via the `tempo` attribute, which
allows you to specify the tempo in beats per minute, where each beat takes up
the length of a quarter note.

For example, to specify a tempo of 180 BPM (♩ = 180):

```alda
(tempo! 180)
```

In traditional music notation, it is common to see a tempo described in terms of
beats per minute, where the note length that takes up a beat is something other
than a quarter note. For example, the tempo might be expressed in terms of "half
notes per minute," e.g. 𝅗𝅥 = 100. This is still "beats per minute," it's just
that the notes are played at a speed where a half note lasts for one beat.

Alda's `tempo` function is flexible in that it allows you to specify the note
length that gets the beat:

```alda
(tempo! 2 100)
```

In triple meters, a dotted note value typically gets the beat. In 6/8, for
example, each measure is typically expressed as 2 beats, where each beat is one
dotted quarter note (♩.) long. In scenarios like this, it is convenient to
express the tempo in terms of the note value that takes the beat.

The note value argument to the `tempo` attribute must be a valid number or
string. Because e.g. `4.` is not a valid number, you must use a string to
represent a dotted note:

```alda
# ♩. = 100
(tempo! "4." 100)
```

## Metric modulation

You can also express tempo in terms of [metric
modulation](https://en.wikipedia.org/wiki/Metric_modulation), i.e. shifting from
one meter to another.

Say, for example, that you're writing a score that starts in 9/8 -- 3 beats per
measure, where each beat is a dotted quarter note.

At a certain point in the piece, you want to transition into a 3/2 section --
still 3 beats per measure, but now each beat is a half note. You want the
"pulse" to stay the same, but now each beat is subdivided into 4 eighth notes
instead of 3. How do you do it?

In traditional notation, it is common to see annotations like "♩. = 𝅗𝅥 " at the
moment in the score where the time signature changes. This signifies that at
that moment, the pulse stays the same, but the amount of time that used to
represent a dotted quarter note now represents a half note. When the orchestra
arrives at that point in the score, the conductor continues to conduct at the
same "speed," but each musician mentally adjusts his/her perception of how to
read his/her part, mentally subdividing each beat into 4 eighth notes instead
of 3 eighth notes.

In Alda, you can express a metric modulation like "♩. = 𝅗𝅥 " as:

```alda
(metric-modulation! "4." 2)
```
