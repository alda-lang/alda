# CRAM

**Crazy Rhythm Alda Magic (C.R.A.M.)** is an alternative way of representing rhythms in Alda. It can be useful for n-tuplets and polyrhythms.

The idea is that you're "cramming" a bunch of [notes](notes.md) into a single note [duration](notes.md#duration). For example, you may want 5 notes "crammed" into the duration of a half note:

```
{c d e f g}2
```

As with notes in general, leaving the duration off of the end of a cram will use the last-used note duration. In the example below, for instance, the notes between the brackets are crammed into the duration of a whole note, since that was the last-specified duration.

```
c1 e {g a b} > c
```

You can also include note-lengths on the notes *inside* of a cram, which will have the effect of giving the longer notes more time, relative to the time allotted for the entire cram. The duration of the entire cram does not change.

```
{c d e}2 {c2 d4 e} {c1 d4 e}
```

> By default, the first note of each cram expression is a quarter note.

Crams can be nested. Each internal cram will take up the appropriate amount of space within the cram containing it.

```
{c e {g a b}}1 c
```
