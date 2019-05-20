# Alda OSC API

This document describes the OSC API for the Alda player process. The Alda client
communicates with the player using this API. Alternate (non-Alda) clients could
potentially be implemented that use the Alda player as a backend, using the API
below to communicate with the player via OSC.

## Duration

There are two aspects of the duration of a note:

1. How long to wait before starting the next note. (**duration**)

2. How long the note is audible. (**audible duration**)

Alda's OSC API separates these explicitly into separate fields that must _both_
be present when specifying the duration of a note.

In the simplest case, these values are the same. This has the effect of legato
playing, where every note is held out for its full value, with no gaps between
the notes.

When the audible duration is _less than_ the duration, the effect is that there
is some amount of gap between this note and the next, which typically sounds
like a crisper attack, staccato, etc.

When the audible duration is _greater than_ the duration, the notes end up
overlapping, which can be used to emulate reverb, or to represent overlapping
notes played on an instrument capable of playing multiple notes at once, e.g. a
piano with the sustain pedal held down.

## Timing

The player is responsible for keeping track of absolute timing of events.

The client expresses timing to the player in terms of relative **offsets**.

For example, a client might ask the player to play a three-note chord followed
by two other notes by sending a bundle including the following messages:

> NB: This isn't the correct order of parameters; I'm still sketching at this
point. For the purposes of this example, the parameters are offset, note number,
duration, audible duration.

```
/track/1/midi/note 0    60 500 500
/track/1/midi/note 0    64 500 500
/track/1/midi/note 0    67 500 500
/track/1/midi/note 500  69 500 450
/track/1/midi/note 1000 72 500 450
```

In Alda notation, this might look like `(tempo! 120) c8/e/g a > c`

This offset is relative to the player's idea of the current "absolute offset"
for that track. From the client's perspective, when sending a bundle, the offset
0 represents the point at which the previous bundle's last note has ended.

For example, in the case of the bundle above, the last note to end happens to be
the last message, MIDI note 72 (C5), which starts at offset 1000 and ends 500 ms
later (1000 + 500 = 1500). A subsequent bundle could then be sent to play
another sequence of notes after that:

```
/track/1/midi/note 0    74 500 450
/track/1/midi/note 500  76 500 450
/track/1/midi/note 1000 77 500 450
```

Note that these notes start from "offset 0," but this is a new bundle, so the
notes will be placed in time _after_ the notes sent in the previous bundle. This
means that "offset 0" in the second bundle is conceptually "offset 1500" in the
first bundle.

The idea here is that clients do not need to be concerned about what's been
scheduled previously, or have any notion of absolute time (that's the player's
concern). A client need only be aware that some other events may have been
scheduled already, and any notes being scheduled now will start once those other
events have ended, and that point in time is "offset 0" from the client's
perspective.

## Tempo

Note that offset is expressed in terms of milliseconds. We prefer to work with
milliseconds instead of standard note lengths (quarter, eighth, etc.) because
milliseconds are more precise, not coupled to tempo, and allow the composition
of music with less rhythmic limits.

Nonetheless, most music software (notation software, etc.) works under the
assumption that music is measured in terms of standard musical notation and is
necessarily coupled to tempo. We want to be able to export MIDI files that work
in this way, which means the division type must be pulses per quarter note
(PPQ).

So that we can have our cake and eat it too, we keep track of the history of
tempo changes during a score, and use this "tempo itinerary" to help us convert
absolute offsets in milliseconds to PPQ offsets expressed in ticks, where the
physical duration depends on the tempo.

This conversion all happens under the hood. From the client's perspective,
offsets are expressed in milliseconds, and tempo can be set at any point in the
score for compatibility with any music software that needs to know about tempo.

Strictly speaking, it isn't necessary to set the tempo, but it can be done, and
it should be done if one wants to export the score to a MIDI file.

## Bundle vs. message

* The top-level can be either a message or a bundle containing 1+ messages.
  * Bundles cannot be nested, as it isn't really clear to me right now what that
    would imply.

* When a bundle is received, the player will synchronize the timing of all of
  the messages in the bundle. This will allow, for example, chords to be
  constructed by sending multiple `/track/1/midi/note` messages, each of which
  has the same offset.

* When an individual message is received, it is treated as if it's a bundle
  containing one message.

## Address patterns

* `/system`
  * `/play` - start playing all tracks at the current system offset
  * `/stop` - stop playing all tracks
    * Remembers the state of the tracks, such that when a `/system/play` is
      received again, playback will continue where it left off.
  * `/clear` - clear all tracks of upcoming events
  * `/tempo` - sets the tempo in BPM

* `/track/1`
  * `/mute` - mute this track
  * `/unmute` - unmute this track
  * `/clear` - clear this track of upcoming events
  * `/midi/patch` - set the MIDI patch number for this track
  * `/midi/percussion` - designate this track to use the MIDI percussion channel
  * `/midi/note` - schedule a MIDI note on ... note off on this track
  * `/pattern` - schedule an instance of a pattern on this track
    * Includes a "times" parameter, for convenient finite loops without having
      to send numerous `/track/1/pattern` messages. You can, for example, send a
      single message that says to play a pattern 16 times.
    * If the pattern is mutated before its scheduled time, the mutated version
      is picked up when the pattern is dereferenced.
  * `/pattern-loop` - loop a pattern indefinitely until stop or clear occurs
    * This is scheduled with an offset, exactly like `/track/1/midi/note` and
      `/track/1/pattern`.
    * If the pattern is mutated during playback, the new version will be picked
      up upon the next iteration through the pattern.
    * Any subsequent notes will be placed "on hold" until the loop is terminated
      via `/track/1/finish-loop`.
  * `/finish-loop` - finish the current iteration of the pattern being looped
    and stop looping.
    * After the final iteration of the loop, any notes scheduled after the loop
      will play in time.

* `/pattern/foo`
  * `/clear` - clear this pattern's contents
  * `/midi/note` - append a note on ... note off to this pattern's contents
    * To _replace_ a pattern's contents, send a bundle that starts with
      `/pattern/foo/clear` and is followed by a number of
      `/pattern/foo/midi/note` messages.
  * `/pattern` - appends a reference to another pattern with the given ID to
    this pattern's contents
    * If the referenced pattern is mutated during looped playback of this
      pattern, the new version will be picked up upon the next iteration through
      this pattern.

## Examples

> This is only a sketch. Parameter order almost guaranteed to be wrong!

* Make track 1 a MIDI channel with instrument 37 (Slap Bass 1) loaded:


  ```
  /track/1/midi/patch 0 37
  ```

  The 0 parameter is offset, just like for note messages. This allows one to
  schedule patch changes in relation to notes.

* On track 1, play MIDI note 64 for 5000 ms at velocity 127:

  ```
  /track/1/midi/note 0 64 5000 5000 127
  ```

* Start playback (all tracks):

  ```
  /system/play
  ```

* Bundle containing a patch change, multiple notes placed 500 ms apart lasting
  almost that long (450 ms) for a little bit of space between notes:

  ```
  /track/1/midi/patch   0    37
  /track/1/midi/note    0    64 500 450 127
  /track/1/midi/note    500  64 500 450 127
  /track/1/midi/note    1000 64 500 450 127
  ```

* Define a pattern `foo`:

  ```
  /pattern/foo/midi/note   0    64 500 450 127
  /pattern/foo/midi/note   500  64 500 450 127
  /pattern/foo/midi/note   1000 64 500 450 127
  ```

  Additional notes can be appended to the pattern by sending subsequent messages
  to that same address. If this is done in a separate bundle, the offset starts
  over at 0, i.e. the next note's offset would be 0, not 1500.

* Redefine `foo`, changing the notes:

  ```
  /pattern/foo/clear
  /pattern/foo/midi/note   0    67 500 450 127
  /pattern/foo/midi/note   500  68 500 450 127
  /pattern/foo/midi/note   1000 69 500 450 127
  ```

* Define a second pattern, `bar`, and have `foo` refer to it:

  ```
  /pattern/bar/midi/note   0    60  500 500 127
  /pattern/bar/midi/note   0    64  500 500 127
  /pattern/bar/midi/note   0    67  500 500 127

  /pattern/foo/pattern     0    bar 1
  /pattern/foo/midi/note   500  67  500 450 127
  /pattern/foo/midi/note   1000 68  500 450 127
  /pattern/foo/midi/note   1500 69  500 450 127
  ```

  In Alda notation, this might look roughly like:

  ```
  (tempo! 120)
  bar = c8/e/g
  foo = bar g8 g+ a
  ```

  (...but replace `=` mentally with a theoretical `+=` operator)

* Fetch the value of `foo` and play it (one time) on track 1:

  ```
  /track/1/pattern 0 foo 1
  ```

* Play `foo` on track 1 twice:

  ```
  /track/1/pattern 0 foo 2
  ```

* Play `foo` on track 1 32 times:

  ```
  /track/1/pattern 0 foo 32
  ```

* Loop `foo` indefinitely:

  ```
  /track/1/pattern-loop 0 foo
  ```

* Stop looping `foo` after the current iteration:

  ```
  /track/1/finish-loop 0
  ```

* After the current iteration, play `foo` 4 more times (and stop looping it):

  ```
  /track/1/finish-loop 0
  /track/1/pattern     0 foo 4
  ```

* Mute track 1 immediately:

  ```
  /track/1/mute
  ```

* Unmute track 1 immediately:

  ```
  /track/1/unmute
  ```

