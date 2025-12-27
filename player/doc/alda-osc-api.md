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

> The parameters here are, in this order:
>
> * Offset
> * Note number
> * Duration
> * Audible duration
> * Velocity
>
> For example, the first three messages below specify that MIDI notes 60, 64,
> and 67 should each be played at offset 0 with a velocity of 100/127, and each
> note should stop sounding after 450 ms.
>
> (If you're wondering what the duration of 500 ms is for, keep reading!)

```
/track/1/midi/note 0    60 500 450 100
/track/1/midi/note 0    64 500 450 100
/track/1/midi/note 0    67 500 450 100
/track/1/midi/note 500  69 500 450 100
/track/1/midi/note 1000 72 500 450 100
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
/track/1/midi/note 0    74 500 450 100
/track/1/midi/note 500  76 500 450 100
/track/1/midi/note 1000 77 500 450 100
```

The first of these notes starts at "offset 0," but this is a new bundle, so the
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

So that we can have our cake and eat it too, we keep a record of all the tempo
changes in a score, and we use this "tempo itinerary" to help us convert
absolute offsets in milliseconds to PPQ offsets expressed in ticks, where the
physical duration depends on the tempo.

This conversion all happens under the hood. From the client's perspective,
offsets are expressed in milliseconds, and tempo can be set at any point in the
score for compatibility with any music software that needs to know about tempo.

Strictly speaking, it isn't necessary to set the tempo, but it can be done, and
you should do it if you want to export the score to a MIDI file that is usable
with other music software (e.g. sheet music notation programs).

## Bundle vs. message

* The top-level can be either a message or a bundle containing 1+ messages.
  Bundles cannot be nested.

* When a bundle is received, the player will synchronize the timing of all of
  the messages in the bundle. This will allow, for example, chords to be
  constructed by sending multiple `/track/1/midi/note` messages, each of which
  has the same offset.

* When an individual message is received, it is treated as if it's a bundle
  containing one message.

## API

<table>
  <thead>
    <tr>
      <th>Address pattern</th>
      <th>Arguments</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>/system/shutdown</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
        </ul>
      </td>
      <td>
        <p>Shut down the player process.</p>
        <p>
          If the provided offset is 0, the player process will shut down
          immediately.
        </p>
        <p>
          Otherwise, the player process will be scheduled to be shut down at the
          specified offset. For example, if you want to schedule three notes
          that are 500 ms long (500 * 3 = 1500 ms total length) and then have
          the player shut down shortly after the last note ends, you can send a
          bundle consisting of three <code>/track/{number}/midi/note</code>
          messages and a <code>/system/shutdown</code> message with an offset of
          something like 2000.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/system/play</code></td>
      <td></td>
      <td>Start playing all tracks at the current system offset.</td>
    </tr>
    <tr>
      <td><code>/system/playback-finished</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Signal that playback is finished. This is used by the Alda client in
          conjunction with the <code>--wait</code> flag to block the
          <code>alda play</code> command until playback is complete.
        </p>
        <p>
          If the provided offset is 0, the player's state will transition to
          "finished" immediately.
        </p>
        <p>
          Otherwise, the player will be scheduled to transition to the
          "finished" state at the specified offset.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/system/stop</code></td>
      <td></td>
      <td>
        <p>Stop playing all tracks.</p>
        <p>
          This puts the player into a "paused" state, such that when a
          <code>/system/play</code> message is received again, playback will
          continue where it left off.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/system/offset</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
        </ul>
      </td>
      <td>
        Set the sequence position to the provided offset, which is expressed in
        milliseconds since the beginning of the score.
      </td>
    </tr>
    <tr>
      <td><code>/system/clear</code></td>
      <td></td>
      <td>Clear all tracks of upcoming events from looping patterns. All notes already scheduled will proceed to play.</td>
    </tr>
    <tr>
      <td><code>/system/tempo</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
          <li>BPM (float)</li>
        </ul>
      </td>
      <td>Sets the tempo in BPM.</td>
    </tr>
    <tr>
      <td><code>/system/midi/export</code></td>
      <td>
        <ul>
          <li>File path (string)</li>
        </ul>
      </td>
      <td>Writes the current state of the sequence to a MIDI file.</td>
    </tr>
    <tr>
      <td><code>/track/{number}/clear</code></td>
      <td></td>
      <td>Clear this track of upcoming events from looping patterns. All notes already scheduled will proceed to play.</td>
    </tr>
    <tr>
      <td><code>/track/{number}/midi/patch</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>Patch number (integer)</li>
        </ul>
      </td>
      <td>Set the MIDI patch number on the specified channel.</td>
    </tr>
    <tr>
      <td><code>/track/{number}/midi/note</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>MIDI note number (integer)</li>
          <li>Duration (integer)</li>
          <li>Audible duration (integer)</li>
          <li>Velocity (integer)</li>
        </ul>
      </td>
      <td>
        <p>Schedule a MIDI note on ... note off on the specified channel.</p>
        <p>Velocity is expected to be an integer in the range 0-127.</p>
      </td>
    </tr>
    <tr>
      <td><code>/track/{number}/midi/volume</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>Volume (integer)</li>
        </ul>
      </td>
      <td>
        <p>Schedule a MIDI expression (11) control change event.</p>
        <p>Volume is expected to be an integer in the range 0-127.</p>
        <p>
          "Volume" here refers to the overall volume of the channel, not
          per-note velocity.
        </p>
        <p>
          You can use expression control messages to set the volume of each
          track to mix their levels relative to one another, and then set the
          velocity on each note (as a parameter of
          <code>/track/{number}/midi/note</code>) for finer-grained control over
          volume from note to note.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/track/{number}/midi/panning</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>Panning (integer)</li>
        </ul>
      </td>
      <td>
        <p>Schedule a MIDI panning (10) control change event.</p>
        <p>Panning is expected to be an integer in the range 0-127.</p>
      </td>
    </tr>
    <tr>
      <td><code>/track/{number}/pattern</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>Pattern name (string)</li>
          <li>Times (integer)</li>
        </ul>
      </td>
      <td>
        <p>Schedule an instance of a pattern on the specified channel.</p>
        <p>
          The "times" argument allows for convenient finite loops without having
          to send numerous <code>/track/{number}/pattern</code> messages. You
          can, for example, send a single message that says to play a pattern 16
          times.
        </p>
        <p>
          If the pattern is mutated during playback, the new version will be
          picked up upon the next iteration through the pattern.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/track/{number}/pattern-loop</code></td>
      <td>
        <ul>
          <li>Channel number (integer)</li>
          <li>Offset (integer)</li>
          <li>Pattern name (string)</li>
        </ul>
      </td>
      <td>
        <p>
          Loop a pattern indefinitely on the specified channel until
          <code>stop</code> or <code>clear</code> occurs.
        </p>
        <p>
          This is scheduled with an offset, exactly like
          <code>/track/{number}/midi/note</code> and
          <code>/track/{number}/pattern</code>.
        </p>
        <p>
          Any subsequent notes will be placed "on hold" until the loop is
          terminated via <code>/track/{number}/finish-loop</code>.
        </p>
        <p>
          If the pattern is mutated during playback, the new version will be
          picked up upon the next iteration through the pattern.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/track/{number}/finish-loop</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Finish the current iteration of the pattern being looped and stop
          looping.
        </p>
        <p>
          After the final iteration of the loop, any notes scheduled after the
          loop will play in time.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/pattern/{name}/clear</code></td>
      <td></td>
      <td>Clear this pattern's contents.</td>
    </tr>
    <tr>
      <td><code>/pattern/{name}/midi/note</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
          <li>MIDI note number (integer)</li>
          <li>Duration (integer)</li>
          <li>Audible duration (integer)</li>
          <li>Velocity (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Append a MIDI note on ... note off to this pattern's contents.
        </p>
        <p>
          To <em>replace</em> a pattern's contents, send a bundle that starts
          with <code>/pattern/{name}/clear</code> and is followed by a number of
          <code>/pattern/{name}/midi/note</code> messages.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/pattern/{name}/midi/volume</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
          <li>Volume (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Append a MIDI expression (11) control change message to the pattern's
          contents.
        </p>
        <p>
          See <code>/track/{number}/midi/volume</code>.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/pattern/{name}/midi/panning</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
          <li>Panning (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Append a MIDI panning (10) control change message to the pattern's
          contents.
        </p>
        <p>
          See <code>/track/{number}/midi/panning</code>.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>/pattern/{name}/pattern</code></td>
      <td>
        <ul>
          <li>Offset (integer)</li>
          <li>Pattern name (string)</li>
          <li>Times (integer)</li>
        </ul>
      </td>
      <td>
        <p>
          Append a reference to another pattern to this pattern's contents.
        </p>
        <p>
          If the referenced pattern is mutated during looped playback of this
          pattern, the new version will be picked up upon the next iteration
          through this pattern.
        </p>
      </td>
    </tr>
  </tbody>
</table>

## Examples

* Load instrument 37 (Slap Bass 1) on MIDI channel 0:

  ```
  /track/1/midi/patch 0 0 37
  ```

  The second 0 parameter is the offset, just like for note messages. This allows
  one to schedule patch changes in relation to notes.

* Play MIDI note 64 for 5000 ms at velocity 127:

  ```
  /track/1/midi/note 0 0 64 5000 5000 127
  ```

* Start playback (all tracks):

  ```
  /system/play
  ```

* Bundle containing a patch change, multiple notes placed 500 ms apart lasting
  almost that long (450 ms) for a little bit of space between notes:

  ```
  /track/1/midi/patch   0 0    37
  /track/1/midi/note    0 0    64 500 450 127
  /track/1/midi/note    0 500  64 500 450 127
  /track/1/midi/note    0 1000 64 500 450 127
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

  Note that the `/pattern/*` endpoints do NOT have an initial channel number
  parameter. This is because patterns can be shared between tracks, and thus,
  the notes and other events (e.g. volume and panning changes) in a pattern, do
  not have a channel number. Instead, a channel number is provided when a
  pattern is invoked, and that channel is used for all of the events in the
  pattern.

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

  ```alda
  (tempo! 120)
  bar = c8/e/g
  foo = bar g8 g+ a
  ```

  (...but replace `=` mentally with a theoretical `+=` operator)

* Fetch the value of `foo` and play it (one time) on channel 0:

  ```
  /track/1/pattern 0 0 foo 1
  ```

* Play `foo` twice:

  ```
  /track/1/pattern 0 0 foo 2
  ```

* Play `foo` 32 times:

  ```
  /track/1/pattern 0 0 foo 32
  ```

* Loop `foo` indefinitely:

  ```
  /track/1/pattern-loop 0 0 foo
  ```

* Stop looping `foo` after the current iteration:

  ```
  /track/1/finish-loop 0
  ```

* After the current iteration, play `foo` 4 more times (and stop looping it):

  ```
  /track/1/finish-loop 0
  /track/1/pattern     0 0 foo 4
  ```
