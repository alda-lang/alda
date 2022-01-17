package io.alda.player

import com.illposed.osc.OSCMessage
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.CompletableFuture
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.thread
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

val playerQueue = LinkedBlockingQueue<List<OSCMessage>>()

private var _midi : MidiEngine? = null

fun midi() : MidiEngine {
  if (_midi == null) {
    _midi = MidiEngine()
  }

  return _midi!!
}

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

  private fun withMidiChannel(f : (Int) -> Unit) {
    midiChannel()?.also { channel ->
      f(channel)
    } ?: run {
      log.warn { "No MIDI channel available for track ${trackNumber}." }
    }
  }

  fun useMidiPercussionChannel() { _midiChannel = 9 }

  val eventBufferQueue = LinkedBlockingQueue<List<Event>>()

  // A count of tasks (List<Event>) that have been taken off of the
  // `eventBufferQueue` and are currently being processed.
  //
  // The count is incremented before a task is placed on the queue and
  // decremented once the task is complete (i.e. events have been scheduled).
  val activeTasks = AtomicInteger(0)

  // The set of patterns that are currently looping (whether that be a finite
  // number of times or indefinitely).
  val activePatterns = mutableSetOf<String>()

  // A monotonically increasing integer representing all of the events we are
  // going to schedule until the track is cleared. Event scheduling is done on
  // multiple threads, so incrementing `era` is a way to signal to the various
  // threads that the track has been cleared, i.e. don't proceed to schedule
  // events.
  var era = 0

  // The base offset that is added to upcoming notes to be scheduled. As notes
  // are scheduled, this base offset is updated to reflect the offset at which
  // the last note to be scheduled will end, so that subsequent notes will line
  // up in time right after the last note.
  var startOffset = 0

  fun clear() {
    synchronized(era) {
      era++
      startOffset = 0
      eventBufferQueue.clear()
      activeTasks.set(0)
      activePatterns.clear()
      withMidiChannel { midi().clearChannel(it) }
    }
  }

  fun mute() {
    withMidiChannel { midi().muteChannel(it) }
  }

  fun unmute() {
    withMidiChannel { midi().unmuteChannel(it) }
  }

  fun schedule(event : Schedulable) {
    withMidiChannel { channel -> event.schedule(channel) }
  }

  /**
   * Schedules the notes of a pattern, blocking until all iterations of the
   * pattern have been scheduled.
   *
   * Patterns can be looped, and the pattern can be changed while it's looping.
   * When this happens, the change is picked up upon the next iteration of the
   * loop. We accomplish this by scheduling each iteration in a "just in time"
   * manner, i.e. shortly before it is due to be played.
   *
   * @param patternEvent An event that specifies a pattern, a relative offset
   * where it should begin, and a number of times to play it.
   * @param _startOffset The absolute offset to which the relative offset is
   * added.
   * @return The list of scheduled events across all iterations of the pattern.
   */
  fun schedulePattern(patternEvent : PatternEventBase, _startOffset : Int)
  : List<Schedulable> {
    var startOffset = _startOffset + patternEvent.offset
    val patternEvents = mutableListOf<Schedulable>()

    // A loop can be stopped externally by removing the pattern from
    // `activePatterns`. If this happens, we stop looping.
    activePatterns.add(patternEvent.patternName)

    try {
      var iteration = 1

      while (!patternEvent.isDone(iteration) &&
             activePatterns.contains(patternEvent.patternName)) {
        log.debug {
          "scheduling iteration $iteration; startOffset: $startOffset; " +
          "patternEvent.offset: ${patternEvent.offset}"
        }

        // This value is the point in time where we schedule the metamessage
        // that signals the lookup and scheduling of the pattern's events.
        //
        // This scheduling happens shortly before the pattern is to be played.
        val patternSchedule = adjustStartOffset(startOffset)

        log.debug { "patternSchedule: $patternSchedule" }

        // This returns a CountDownLatch that starts at 1 and counts down to 0
        // when the `patternSchedule` offset is reached in the sequence.
        val latch = midi().scheduleEvent(
          patternSchedule, patternEvent.patternName
        )

        // Wait until it's time to look up the pattern's current value and
        // schedule the events.
        //
        // We check the `era` now and then again when it's time to schedule, and
        // see if the track has been cleared in the meantime. We only proceed if
        // the track hasn't been cleared (i.e. if the `era` hasn't changed).
        val eraBefore = synchronized(era) { era }
        log.debug { "awaiting latch" }
        latch.await()
        val eraAfter = synchronized(era) { era }
        log.debug { "eraBefore: $eraBefore; eraAfter: $eraAfter" }
        if (eraBefore != eraAfter) break

        log.debug { "scheduling pattern ${patternEvent.patternName}" }

        val pattern = pattern(patternEvent.patternName)

        // It's safe to filter a List<Event> down to just the ones that are
        // Schedulable and then cast it to a List<Schedulable>.
        @Suppress("UNCHECKED_CAST")
        val events : MutableList<Schedulable> =
          (pattern.events.map { it.addOffset(startOffset) }
                         .filter { it is Schedulable }
                         as List<Schedulable>)
            .toMutableList()

        events.forEach { schedule(it) }

        // Now that we've scheduled at least one iteration, we can start
        // playing. (Unless we've already started playing, in which case this is
        // a no-op.)
        synchronized(midi().isPlaying) {
          if (midi().isPlaying) midi().startSequencer()
        }

        // It's safe to filter a List<Event> down to just the ones that are
        // PatternEvents and then cast it to a List<PatternEvent>.
        @Suppress("UNCHECKED_CAST")
        // Here, we handle the case where the pattern's events include further
        // pattern events, i.e. the pattern references another pattern.
        //
        // NB: Because of the "just in time" semantics of scheduling patterns,
        // this means we block here until the subpattern is about due to be
        // played.
        (pattern.events.filter { it is PatternEvent } as List<PatternEvent>)
          .forEach { events.addAll(schedulePattern(it, startOffset)) }

        if (!events.isEmpty())
          startOffset = events.map { (it as Event).endOffset() }.maxOrNull()!!

        patternEvents.addAll(events)

        iteration++
      }
    } finally {
      activePatterns.remove(patternEvent.patternName)
    }

    return patternEvents
  }

  private fun adjustStartOffset(_startOffset : Int) : Int {
    var startOffset = _startOffset

    val now = Math.round(midi().currentOffset()).toInt()

    // If we're not scheduling into the future, then whatever we're supposed to
    // be scheduling should happen ASAP.
    if (startOffset < now) startOffset = now

    // Ensure that there is time to schedule the events before they're due to
    // come up in the sequence.
    if (midi().isPlaying && (startOffset - now < SCHEDULE_BUFFER_TIME_MS)) {
      log.trace { "The note would be due in ${startOffset - now} ms, so " +
                  "adding ${SCHEDULE_BUFFER_TIME_MS} to the scheduled offset" }
      startOffset += SCHEDULE_BUFFER_TIME_MS
    }

    return startOffset
  }

  fun scheduleEvents(events : List<Event>, _startOffset : Int) : Int {
    val startOffset = adjustStartOffset(_startOffset)

    events.filter { it is MidiPatchEvent }.forEach {
      schedule((it as MidiPatchEvent).addOffset(startOffset))
    }

    events.filter { it is MidiPercussionEvent }.forEach {
      val event = it as MidiPercussionEvent

      if (event.offset == 0) {
        midi().percussionImmediate(trackNumber)
      } else {
        midi().percussionScheduled(trackNumber, startOffset + event.offset)
      }
    }

    val scheduledEvents = mutableListOf<Schedulable>()

    // It's safe to filter a List<Event> down to just the ones that are
    // Schedulable and then cast it to a List<Schedulable>.
    @Suppress("UNCHECKED_CAST")
    val immediateEvents : List<Schedulable> =
      events.map { it.addOffset(startOffset) }
            .filter { it is Schedulable }
            as List<Schedulable>

    immediateEvents.forEach { schedule(it) }

    scheduledEvents.addAll(immediateEvents)

    // For each pattern event, we...
    // * wait until right before the pattern is supposed to be played
    // * look up the pattern
    // * schedule the pattern's events
    // * add the pattern's events to `scheduledEvents`
    //
    // Patterns can include other patterns, and to support dynamically changing
    // pattern contents on the fly, we look up each pattern's contents shortly
    // before it is scheduled to play.
    //
    // Patterns can also loop indefinitely.
    //
    // To support multiple patterns looping concurrently on the same track, we
    // do the on-the-fly scheduling of each pattern on a separate thread and
    // collect the results in a CompletableFuture.
    events.filter { it is PatternEventBase }.map {
      val event = it as PatternEventBase
      val future = CompletableFuture<List<Schedulable>>()

      thread {
        future.complete(schedulePattern(event, startOffset))
      }

      future
    }.forEach { future ->
      scheduledEvents.addAll(future.get())
    }

    // Now that all the notes have been scheduled, we can start the sequencer
    // (assuming it hasn't been started already, in which case this is a no-op).
    synchronized(midi().isPlaying) {
      if (midi().isPlaying) midi().startSequencer()
    }

    // At this point, `noteEvents` should contain all of the notes we've
    // scheduled, including the values of patterns at the moment right before
    // they were scheduled.
    //
    // We can now calculate the latest note end offset, which shall be our new
    // `startOffset`.

    if (scheduledEvents.isEmpty())
      return _startOffset

    return scheduledEvents.map { (it as Event).endOffset() }.maxOrNull()!!
  }

  init {
    // This thread schedules events on this track.
    thread {
      // When new events come in on the `eventBufferQueue`, it may be the case
      // that previous events are still lined up to be scheduled (e.g. a pattern
      // is looping). When this is the case, the new events wait in line until
      // the previous scheduling has completed and the offset where the next
      // events should start is updated.
      val scheduling = ReentrantLock(true) // fairness enabled

      while (!Thread.currentThread().isInterrupted()) {
        try {
          val events = eventBufferQueue.take()

          events.filter { it is FinishLoopEvent }.forEach {
            thread {
              val event = it as FinishLoopEvent
              val offset = adjustStartOffset(startOffset) + event.offset
              val latch = midi().scheduleEvent(offset, "FinishLoop")
              latch.await()
              log.debug { "clearing active patterns" }
              activePatterns.clear()
            }
          }

          // We start a new thread here so that we can wait for the opportunity
          // to schedule new events, while the parent thread continues to
          // receive new events on the queue.
          thread {
            // Wait for the previous scheduling of events to finish.
            //
            // We check the `era` now and then again when it's time to schedule,
            // and see if the track has been cleared in the meantime. We only
            // proceed if the track hasn't been cleared (i.e. if the `era`
            // hasn't changed).
            val eraBefore = synchronized(era) { era }
            scheduling.lock()
            val eraAfter = synchronized(era) { era }

            log.debug { "TRACK ${trackNumber}: startOffset is ${startOffset}" }

            try {
              log.debug { "eraBefore: $eraBefore; eraAfter: $eraAfter" }
              if (eraBefore == eraAfter) {
                // Schedule events and update `startOffset` to be the offset at
                // which the next events should start (after the ones we're
                // scheduling here).
                startOffset = scheduleEvents(events, startOffset)
              }
            } finally {
              activeTasks.updateAndGet { n -> if (n == 0) 0 else n - 1 }
              scheduling.unlock()
            }
          }
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

// Wait until any events that are about to be or are actively being scheduled
// are finished being scheduled.
fun awaitActiveTasks() {
  tracks.forEach { (_, track) ->
    while (track.activeTasks.get() > 0) {
      Thread.sleep(100)
    }
  }
}

private fun applyUpdates(updates : Updates) {
  log.trace { "----" }
  log.trace { updates.systemActions }
  log.trace { updates.trackActions }
  log.trace { updates.systemEvents }
  log.trace { updates.trackEvents }
  log.trace { updates.patternActions }
  log.trace { updates.patternEvents }
  log.trace { "----" }

  // PHASE 1: shutdown/stop/offset/mute/clear

  if (updates.systemActions.contains(SystemAction.SHUTDOWN))
    isRunning = false

  if (updates.systemActions.contains(SystemAction.STOP))
    midi().stopSequencer()

  if (updates.systemActions.contains(SystemAction.CLEAR)) {
    tracks.forEach { _, track -> track.clear() }
  }

  updates.systemEvents.filter { it is SetOffsetEvent }.forEach {
    val setOffsetEvent = it as SetOffsetEvent
    awaitActiveTasks()
    midi().setSequencerOffset(setOffsetEvent.offset)
  }

  updates.trackActions.forEach { (trackNumber, actions) ->
    if (actions.contains(TrackAction.MUTE)) {
      track(trackNumber).mute()
    }

    if (actions.contains(TrackAction.CLEAR)) {
      track(trackNumber).clear()
    }
  }

  updates.patternActions.forEach { (patternName, actions) ->
    if (actions.contains(PatternAction.CLEAR)) {
      pattern(patternName).events.clear()
    }
  }

  // PHASE 2: update tempo and patterns

  updates.systemEvents.filter { it is TempoEvent }.forEach {
    val tempoEvent = it as TempoEvent
    midi().setTempo(tempoEvent.offset, tempoEvent.bpm)
  }

  updates.patternEvents.forEach { (patternName, events) ->
    pattern(patternName).events.addAll(events)
  }

  // PHASE 3: update tracks

  updates.trackEvents.forEach { (trackNumber, events) ->
    val track = track(trackNumber)
    track.activeTasks.incrementAndGet()
    track.eventBufferQueue.put(events)
  }

  // PHASE 4: export

  updates.systemEvents.filter { it is MidiExportEvent }.forEach {
    awaitActiveTasks()
    midi().export((it as MidiExportEvent).filepath)
  }

  // PHASE 5: unmute/play

  updates.trackActions.forEach { (trackNumber, actions) ->
    if (actions.contains(TrackAction.UNMUTE)) {
      track(trackNumber).unmute()
    }
  }

  // PHASE 6: Scheduled shutdown
  // (It's important that we do this sometime _after_ handling tempo events.
  // Otherwise, the timing of the shutdown can be off. The scheduling of the
  // shutdown needs to be done with an awareness of all of the tempo changes
  // that will occur in the score.)
  updates.systemEvents.filter { it is ShutdownEvent }.forEach {
    val shutdownEvent = it as ShutdownEvent
    midi().scheduleShutdown(shutdownEvent.offset)
  }

  // NB: We don't actually start the sequencer here; that action needs to be
  // deferred until after a track thread finishes scheduling a buffer of events.
  if (updates.systemActions.contains(SystemAction.PLAY)) {
    awaitActiveTasks()
    midi().isPlaying = true
  }
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

