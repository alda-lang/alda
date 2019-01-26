# Attributes

An **attribute** defines some quality of how an [instrument](scores-and-parts.md) (or multiple instruments) plays its [notes](notes.md).

Under the hood, attributes are implemented as [inline clojure code](inline-clojure-code.md).

## Setting the Value of an Attribute

Just like setting [octaves](notes.md#octave), setting an attribute will take effect for all of an instrument's upcoming notes, until that attribute is changed again. (In fact, `octave` is also a settable attribute.)

Different attributes take different kinds of values. A lot of the time, the value is a number between 0 and 100, but this is not always the case.

### Examples

```
(volume 50)
```

```
(quant 85)
```

```
(octave :up)
```

```
(tempo 240)
```

```
(key-signature "f+ c+ g+")
```

See [below](#list-of-attributes) for more information about the different kinds of attributes that are available to you when writing a score.

## Per-Instrument vs. Global

By default, an attribute change event is only applied to the instrument(s) that you're currently working with. For instance, in a score with four instruments:

```
violin "violin-1":
  o4 f2   g4 a   b-2   a

violin "violin-2":
  o4 c2   e4 f   f2    f

viola:
  o3 a2 > c4 c   d2    c

cello:
  o3 f2   c4 f < b-2 > f
```

Changing an attribute will only affect the instrument(s) whose part you are currently editing:

```
violin "violin-1":
  o4 f2   g4 a   b-2   a

violin "violin-2":
  o4 c2   e4 f   f2    f

viola:
  o3 a2 > c4 c   d2    c

cello:
  (volume 75)
  o3 f2   c4 f < b-2 > f
```

To change an attribute **globally** (i.e. for every instrument in the score), add an exclamation mark (`!`) after the name of the attribute:

```
violin "violin-1":
  (tempo! 80)
  o4 f2   g4 a   b-2   a

violin "violin-2":
  o4 c2   e4 f   f2    f

viola:
  o3 a2 > c4 c   d2    c

cello:
  o3 f2   c4 f < b-2 > f
```

Attributes can also be set globally at the beginning of a score, before you start writing out any instrument parts. The attributes will still be set for every instrument at the beginning of the score.

```
(tempo! 80)

violin "violin-1":
  o4 f2   g4 a   b-2   a

violin "violin-2":
  o4 c2   e4 f   f2    f

viola:
  o3 a2 > c4 c   d2    c

cello:
  o3 f2   c4 f < b-2 > f
```

## List of Attributes

### `duration`

* **Abbreviations:** (none)

* **Description:** The length that a note will have, if not specified. For example, `c4` is explicitly a quarter note; `c` will have a note-length equal to the value of the instrument's `duration` attribute.

  Note that this attribute is more of an implementation detail, as an instrument's default note length is set implicitly whenever you specify a note-length on a note. You will probably never need to use this attribute directly.

  Worth mentioning: if you do ever find yourself needing to set the duration attribute directly, the function you use is `set-duration`, *not* `duration`.

* **Value:** either:
  - a number of beats (e.g. 2.5, which represents a dotted half note), or
  - a number of milliseconds

  If the value you provide is a number, it will be intepreted as a number of beats.

  Note that this value is a different number than the note value. For example, to set the duration to a quarter note, the value is `1`, not `4`, because a quarter note is 1 beat long:

  ```
  (set-duration 1)
  ```

  You can use the `note-length` function to translate a note value (e.g. `4` for a quarter note) into the right number of beats:

  ```
  (set-duration (note-length 4))
  ```

  To specify the duration as a number of milliseconds, use the `ms` function:

  ```
  (set-duration (ms 2000))
  ```

* **Initial Value:** `(note-length 4)` (i.e. a quarter note, or 1 beat)

### `key-signature`

* **Abbreviations:** `key-sig`

* **Description:** [a set of sharp or flat symbols](https://en.wikipedia.org/wiki/Key_signature) to be applied to certain notes by default when the note doesn't include accidentals. For example, if the key signature contains G-sharp, then a note `g` will become G-sharp by default, unless an accidental is placed after the note, i.e. `g-` (G-flat) or `g_` (G natural).

* **Value:** either:
  * a map of letters (as keywords) to lists of accidentals for that letter, e.g. `{:f [:sharp] :c [:sharp] :g [:sharp]}`,
  * a string like `"f+ c+ g+"`, or
  * a vector like `[:a :major]` or `[:e :flat :minor]`
    * supported scales/modes:
      * `:ionian` (`:major`)
      * `:dorian`
      * `:phrygian`
      * `:lydian`
      * `:mixolydian`
      * `:aeolian` (`:minor`)
      * `:locrian`

* **Initial Value:** `{}` (an empty map, signifying no flats/sharps will be applied for any letter)

### `octave`

* **Abbreviations:** (none)

* **Description:** The octave of a note. (Note that Alda also has built-in syntax for setting specific octaves, e.g. `o5`, and moving up `>` or down `<` octaves.)

* **Value:** either a number representing an octave in [scientific pitch notation](https://en.wikipedia.org/wiki/Scientific_pitch_notation), or the keyword `:up` or `:down` to move up or down by one from the current octave.

* **Initial Value:** 4

### `panning`

* **Abbreviations:** `pan`

* **Description:** How far left/right the note is panned in your speakers.

* **Value:** a number from 0-100 representing the panning from hard left (0) to hard right (100). 50 is center.

* **Initial Value:** 50

### `quantization`

* **Abbreviations:** `quant`, `quantize`

* **Description:** The percentage of a note's full duration that is heard. Setting lower `quantization` values translates into putting more space between notes, making them sound more *staccato*, whereas setting higher values translates into putting *less* space between notes, making them sound more *legato*.

* **Value:** a number between 0 and 100

* **Initial Value:** 90

### `tempo`

* **Abbreviations:** (none)

* **Description:** How fast or slow notes are played. This value is used in combination with the length of a note to determine how long to play it in milliseconds.

* **Value:** a number representing a [tempo](https://en.wikipedia.org/wiki/Tempo) in beats per minute (BPM)

* **Initial Value:** 120

> By default, `(tempo 100)` will set the tempo to 100 beats per minute, where
> each beat takes up the length of a quarter note.
>
> Alda also offers additional ways to express tempo. See: [tempo](tempo.md).

### `track-volume`

* **Abbreviations:** `track-vol`

* **Description:** The overall volume of an instrument. For MIDI instruments, this corresponds to track volume, as opposed to velocity. Typically, you would set `track-volume` once (if at all) at the beginning of the score, and then use `volume` for finer-grained control over volume between notes. (When in doubt, just use `volume`!)

* **Value:** a number between 0 and 100

* **Initial Value:** 78.7 (this number comes from 100/127, which is the default track volume for MIDI, at least in the Java VM)

### `transposition`

* **Abbreviations:** `transpose`

* **Description:** Moves all notes up or down by a desired number of semitones.
  This attribute can be used to make writing parts for [transposing
  instruments](https://en.wikipedia.org/wiki/Transposing_instrument) more
  convenient.

* **Value:** a positive or negative integer representing a number of semitones
  (half-steps) to move each note up or down.

* **Initial Value:** 0

### `volume`

* **Abbreviations:** `vol`

* **Description:** How loud or soft a note is. For MIDI instruments, this corresponds to the *velocity* of each note, which has to do with not only how loud the note is, but also how *strongly* the note is played. The details of this vary from instrument to instrument; often, it has an effect on the sharpness of the attack at the beginning of the note.

* **Value:** a number between 0 and 100

* **Initial Value:** 100
