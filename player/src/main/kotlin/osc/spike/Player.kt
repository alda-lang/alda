package osc.spike

import com.illposed.osc.OSCMessage
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.Phaser
import kotlin.concurrent.thread

val playerQueue = LinkedBlockingQueue<List<OSCMessage>>()

val midi = MidiEngine()

val availableChannels = ((0..15).toSet() - 9).toMutableSet()

class Track(val trackNumber : Int) {
  private var _midiChannel : Int? = null
  fun midiChannel() : Int? {
    synchronized(availableChannels) {
      if (_midiChannel == null && !availableChannels.isEmpty()) {
        val channel = availableChannels.first()
        availableChannels.remove(channel)
        _midiChannel = channel
      }
    }

    return _midiChannel
  }

  fun useMidiPercussionChannel() { _midiChannel = 9 }

  val eventBufferQueue = LinkedBlockingQueue<List<Event>>()

  fun scheduleMidiPatch (event : MidiPatchEvent, startOffset : Int) {
    midiChannel()?.also { channel ->
      // debug
      println("track ${trackNumber} is channel ${channel}")
      midi.patch(startOffset + event.offset, channel, event.patch)
    } ?: run {
      println("WARN: No MIDI channel available for track ${trackNumber}.")
    }
  }

  fun scheduleMidiNote(event : MidiNoteEvent, startOffset : Int) {
    midiChannel()?.also { channel ->
      val noteStart = startOffset + event.offset
      val noteEnd = noteStart + event.audibleDuration
      midi.note(noteStart, noteEnd, channel, event.noteNumber, event.velocity)
    } ?: run {
      println("WARN: No MIDI channel available for track ${trackNumber}.")
    }
  }

  fun schedulePattern(
    // An event that specifies a pattern, a relative offset where it should
    // begin, and a number of times to play it.
    event : PatternEvent,
    // The absolute offset to which the relative offset is added.
    startOffset : Int,
    // A phaser used to coordinate scheduling this pattern along with others.
    phaser : Phaser,
    // After we schedule all the notes in the pattern, we add them to this list
    // as an output mechanism, and then deregister with the phaser.
    noteEvents : MutableList<MidiNoteEvent>
  ) {
    phaser.register()

    val patternStart = startOffset + event.offset

    // This value is the point in time where we schedule the metamessage that
    // signals the lookup and scheduling of the pattern's events.  This
    // scheduling happens shortly before the pattern is to be played.
    val patternSchedule = Math.max(
      startOffset, patternStart - SCHEDULE_BUFFER_TIME_MS
    )

    // This returns a CountDownLatch that starts at 1 and counts down to 0 when
    // the pattern metamessage is reached in the sequence.
    val latch = midi.pattern(patternSchedule, event.patternName)

    thread {
      // Wait until it's time to look up the pattern's current value and
      // schedule the events.
      println("awaiting latch")
      latch.await()

      println("scheduling pattern ${event.patternName}")

      val pattern = pattern(event.patternName)

      val patternNoteEvents : MutableList<MidiNoteEvent> =
        pattern.events.filter { it is MidiNoteEvent }
        as MutableList<MidiNoteEvent>

      patternNoteEvents.forEach { scheduleMidiNote(it, patternStart) }

      val patternEvents =
        pattern.events.filter { it is PatternEvent } as List<PatternEvent>

      // Here, we handle the case where the pattern's events include further
      // pattern events, i.e. the pattern references another pattern.
      //
      // When the inner pattern event's events are scheduled, they are added to
      // this pattern's note events (`patternNoteEvents`).
      val innerPhaser = Phaser()
      innerPhaser.register()

      // FIXME: we need to preserve the original start offset somehow, otherwise
      // the startOffset won't be correctly updated

      patternEvents.forEach { event ->
        schedulePattern(
          event as PatternEvent, patternStart, innerPhaser, patternNoteEvents
        )
      }

      innerPhaser.arriveAndDeregister()
      innerPhaser.awaitAdvance(0)

      // If the pattern event is to played more than once, then we enter a
      // second phase where we schedule the next iteration of the pattern.
      if (event.times > 1) {
        innerPhaser.register()

        val nextStartOffset =
          patternNoteEvents.map { it.offset + it.duration }.max()!!

        schedulePattern(
          PatternEvent(nextStartOffset, event.patternName, event.times - 1),
          patternStart,
          innerPhaser,
          patternNoteEvents
        )

        println("awaiting schedule of next iteration")
        innerPhaser.arriveAndDeregister()
        innerPhaser.awaitAdvance(1)
      }

      println("adding patternNoteEvents to noteEvents")
      noteEvents.addAll(patternNoteEvents)

      println("arriveAndDeregister")
      phaser.arriveAndDeregister()
    }
  }

  init {
    // This thread schedules events on this track.
    thread {
      // Before we can schedule these events, we need to know the start offset.
      // This can change dynamically, e.g. if a pattern is changed on-the-fly
      // during playback, so we defer scheduling the next buffer of events as
      // long as we can.
      var startOffset = 0

      while (!Thread.currentThread().isInterrupted()) {
        try {
          println("TRACK ${trackNumber}: startOffset is ${startOffset}")

          val events = eventBufferQueue.take()

          val now = Math.round(midi.currentOffset()).toInt()

          // If we're not scheduling into the future, then the notes should be
          // played ASAP.
          if (startOffset < now) startOffset = now

          // Ensure that there is time to schedule the notes before it's time to
          // play them.
          if (midi.isPlaying && (startOffset - now < SCHEDULE_BUFFER_TIME_MS))
            startOffset += SCHEDULE_BUFFER_TIME_MS

          events.filter { it is MidiPatchEvent }.forEach {
            scheduleMidiPatch(it as MidiPatchEvent, startOffset)
          }

          events.filter { it is MidiPercussionEvent }.forEach {
            midi.percussion(
              startOffset + (it as MidiPercussionEvent).offset, trackNumber
            )
          }

          val noteEvents =
            events.filter { it is MidiNoteEvent } as MutableList<MidiNoteEvent>

          val patternEvents =
            events.filter { it is PatternEvent } as MutableList<PatternEvent>

          events.forEach { event ->
            when (event) {
              is PatternLoopEvent -> {
                // TODO
              }

              is FinishLoopEvent -> {
                // TODO
              }
            }
          }

          noteEvents.forEach { scheduleMidiNote(it, startOffset) }

          // Patterns can include other patterns, and to support dynamically
          // changing pattern contents on the fly, we look up each pattern's
          // contents shortly before it is scheduled to play. This means that
          // the total number of patterns can change at a moment's notice.
          //
          // We're using a Phaser here as a more flexible version of the
          // CountDownLatch that can also count up as needed, i.e. whenever a
          // pattern is discovered within another pattern's contents, and we
          // must now wait on the new pattern to be scheduled as well.
          val phaser = Phaser()

          // Register the main thread with the phaser. This is necessary because
          // unless there is at least one registered party, the call to
          // `phaser.awaitAdvance` below will block forever.
          phaser.register()

          // For each pattern event, register with the phaser and then start a
          // thread that:
          //
          // * waits until right before the pattern is supposed to be played
          // * looks up the pattern
          // * schedules the pattern's events
          // * adds the pattern's events to `noteEvents`
          // * deregisters with the phaser
          patternEvents.forEach { event ->
            schedulePattern(event, startOffset, phaser, noteEvents)
          }

          // Deregister the main thread. At this point, we're waiting for any
          // pattern-scheduling threads that have registered with the phaser to
          // finish what they're doing and deregister themselves.
          phaser.arriveAndDeregister()

          println("awaiting phaser")

          // Block until all patterns' events have been scheduled.
          phaser.awaitAdvance(0)

          println("phaser done")

          // Once we've finished awaiting the phaser, `noteEvents` should
          // contain all of the notes we've scheduled, including the values of
          // patterns at the moment right before they were scheduled.
          //
          // At that point, we will be able to calculate the latest note end
          // offset, which shall be our new `startOffset`.

          synchronized(midi.isPlaying) {
            if (midi.isPlaying) midi.startSequencer()
          }

          if (!noteEvents.isEmpty())
            startOffset =
              noteEvents.map { startOffset + it.offset + it.duration }.max()!!
        } catch (iex : InterruptedException) {
          Thread.currentThread().interrupt()
        }
      }
    }
  }
}

val tracks = mutableMapOf<Int, Track>()

fun track(trackNumber: Int): Track {
  if (!tracks.containsKey(trackNumber))
    tracks.put(trackNumber, Track(trackNumber))

  return tracks.get(trackNumber)!!
}

class Pattern() {
  val events = mutableListOf<Event>()
}

val patterns = mutableMapOf<String, Pattern>()

fun pattern(patternName: String): Pattern {
  if (!patterns.containsKey(patternName))
    patterns.put(patternName, Pattern())

  return patterns.get(patternName)!!
}

private fun applyUpdates(updates : Updates) {
  // debug
  println("----")
  println(updates.systemActions)
  println(updates.trackActions)
  println(updates.trackEvents)
  println(updates.patternActions)
  println(updates.patternEvents)
  println("----")

  // PHASE 1: stop/mute/clear

  if (updates.systemActions.contains(SystemAction.STOP))
    midi.stopSequencer()

  if (updates.systemActions.contains(SystemAction.CLEAR)) {
    // TODO
  }

  updates.trackActions.forEach { (trackNumber, actions) ->
    if (actions.contains(TrackAction.MUTE)) {
      // TODO
    }

    if (actions.contains(TrackAction.CLEAR)) {
      // TODO
    }
  }

  updates.patternActions.forEach { (patternName, actions) ->
    if (actions.contains(PatternAction.CLEAR)) {
      pattern(patternName).events.clear()
    }
  }

  // PHASE 2: update patterns

  updates.patternEvents.forEach { (patternName, events) ->
    pattern(patternName).events.addAll(events)
  }

  // PHASE 3: update tracks

  updates.trackEvents.forEach { (trackNumber, events) ->
    track(trackNumber).eventBufferQueue.put(events)
  }

  // PHASE 4: unmute/play

  updates.trackActions.forEach { (trackNumber, actions) ->
    if (actions.contains(TrackAction.UNMUTE)) {
      // TODO
    }
  }

  // NB: We don't actually start the sequencer here; that action needs to be
  // deferred until after a track thread finishes scheduling a buffer of events.
  if (updates.systemActions.contains(SystemAction.PLAY))
    midi.isPlaying = true
}

fun player() : Thread {
  return thread(start = false) {
    while (!Thread.currentThread().isInterrupted()) {
      try {
        val instructions = playerQueue.take()
        val updates = parseUpdates(instructions)
        applyUpdates(updates)
      } catch (iex : InterruptedException) {
        Thread.currentThread().interrupt()
      }
    }
  }
}

